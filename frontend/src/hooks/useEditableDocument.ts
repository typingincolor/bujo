import { useState, useEffect, useCallback, useRef } from 'react'
import {
  GetEditableDocument,
  ValidateEditableDocument,
  ApplyEditableDocument,
} from '../wailsjs/go/wails/App'

export interface ValidationError {
  lineNumber: number
  message: string
  quickFixes?: string[]
}

export interface DeletedEntry {
  entityId: string
  content: string
}

export interface ApplyResult {
  inserted: number
  updated: number
  deleted: number
  migrated: number
}

export interface SaveResult {
  success: boolean
  error?: string
  result?: ApplyResult
}

export interface EditableDocumentState {
  document: string
  setDocument: (doc: string) => void
  isLoading: boolean
  error: string | null
  isDirty: boolean
  validationErrors: ValidationError[]
  deletedEntries: DeletedEntry[]
  trackDeletion: (entityId: string, content: string) => void
  restoreDeletion: (entityId: string) => void
  save: () => Promise<SaveResult>
  discardChanges: () => void
  lastSaved: Date | null
  hasDraft: boolean
  restoreDraft: () => void
  discardDraft: () => void
}

const DEBOUNCE_MS = 500

function formatDateKey(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function getDraftKey(date: Date): string {
  return `bujo.draft.${formatDateKey(date)}`
}

interface Draft {
  document: string
  deletedIds: string[]
  timestamp: number
}

export function useEditableDocument(date: Date): EditableDocumentState {
  const [document, setDocumentState] = useState('')
  const [originalDocument, setOriginalDocument] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [validationErrors, setValidationErrors] = useState<ValidationError[]>([])
  const [deletedEntries, setDeletedEntries] = useState<DeletedEntry[]>([])
  const [lastSaved, setLastSaved] = useState<Date | null>(null)
  const [hasDraft, setHasDraft] = useState(false)

  const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const draftKey = getDraftKey(date)

  const checkForDraft = useCallback(() => {
    const stored = localStorage.getItem(draftKey)
    if (stored) {
      setHasDraft(true)
    }
  }, [draftKey])

  useEffect(() => {
    let cancelled = false

    async function loadDocument() {
      setIsLoading(true)
      setError(null)

      try {
        const doc = await GetEditableDocument(date)
        if (!cancelled) {
          setDocumentState(doc)
          setOriginalDocument(doc)
          setIsLoading(false)
          checkForDraft()
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : String(err))
          setIsLoading(false)
        }
      }
    }

    loadDocument()

    return () => {
      cancelled = true
    }
  }, [date, checkForDraft])

  const saveDraft = useCallback(
    (doc: string, deletedIds: string[]) => {
      const draft: Draft = {
        document: doc,
        deletedIds,
        timestamp: Date.now(),
      }
      localStorage.setItem(draftKey, JSON.stringify(draft))
    },
    [draftKey]
  )

  const clearDraft = useCallback(() => {
    localStorage.removeItem(draftKey)
    setHasDraft(false)
  }, [draftKey])

  const validateDocument = useCallback(async (doc: string) => {
    try {
      const result = await ValidateEditableDocument(doc)
      setValidationErrors(result.errors || [])
    } catch {
      // Validation failures are not critical errors
    }
  }, [])

  const setDocument = useCallback(
    (newDoc: string) => {
      setDocumentState(newDoc)

      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }

      debounceTimerRef.current = setTimeout(() => {
        validateDocument(newDoc)
        const deletedIds = deletedEntries.map((e) => e.entityId)
        saveDraft(newDoc, deletedIds)
      }, DEBOUNCE_MS)
    },
    [validateDocument, saveDraft, deletedEntries]
  )

  const isDirty = document !== originalDocument || deletedEntries.length > 0

  const trackDeletion = useCallback((entityId: string, content: string) => {
    setDeletedEntries((prev) => [...prev, { entityId, content }])
  }, [])

  const restoreDeletion = useCallback((entityId: string) => {
    setDeletedEntries((prev) => prev.filter((e) => e.entityId !== entityId))
  }, [])

  const save = useCallback(async (): Promise<SaveResult> => {
    try {
      const validation = await ValidateEditableDocument(document)
      if (!validation.isValid) {
        setValidationErrors(validation.errors || [])
        return { success: false, error: 'Validation failed' }
      }

      const deletedIds = deletedEntries.map((e) => e.entityId)
      const result = await ApplyEditableDocument(document, date, deletedIds)

      setOriginalDocument(document)
      setDeletedEntries([])
      setLastSaved(new Date())
      clearDraft()

      return { success: true, result }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err)
      return { success: false, error: errorMessage }
    }
  }, [document, date, deletedEntries, clearDraft])

  const discardChanges = useCallback(() => {
    setDocumentState(originalDocument)
    setDeletedEntries([])
    setValidationErrors([])
  }, [originalDocument])

  const restoreDraft = useCallback(() => {
    const stored = localStorage.getItem(draftKey)
    if (stored) {
      const draft: Draft = JSON.parse(stored)
      setDocumentState(draft.document)
      setHasDraft(false)
    }
  }, [draftKey])

  const discardDraft = useCallback(() => {
    clearDraft()
  }, [clearDraft])

  useEffect(() => {
    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }
    }
  }, [])

  return {
    document,
    setDocument,
    isLoading,
    error,
    isDirty,
    validationErrors,
    deletedEntries,
    trackDeletion,
    restoreDeletion,
    save,
    discardChanges,
    lastSaved,
    hasDraft,
    restoreDraft,
    discardDraft,
  }
}

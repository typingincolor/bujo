import { useState, useEffect, useCallback, useRef } from 'react'
import {
  GetEditableDocument,
  ValidateEditableDocument,
  ApplyEditableDocument,
  ApplyEditableDocumentWithActions,
} from '../wailsjs/go/wails/App'
import { toWailsTime } from '@/lib/wailsTime'

export interface ValidationError {
  lineNumber: number
  message: string
  quickFixes?: string[]
}

export interface ApplyResult {
  inserted: number
  deleted: number
}

export interface SaveResult {
  success: boolean
  error?: string
  result?: ApplyResult
}

export interface SaveActions {
  migrateDate?: Date
  listId?: number
}

export interface EditableDocumentState {
  document: string
  originalDocument: string
  setDocument: (doc: string) => void
  isLoading: boolean
  error: string | null
  isDirty: boolean
  validationErrors: ValidationError[]
  save: () => Promise<SaveResult>
  saveWithActions: (actions: SaveActions) => Promise<SaveResult>
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

export function useEditableDocument(date: Date): EditableDocumentState {
  const [document, setDocumentState] = useState('')
  const [originalDocument, setOriginalDocument] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [validationErrors, setValidationErrors] = useState<ValidationError[]>([])
  const [lastSaved, setLastSaved] = useState<Date | null>(null)
  const [hasDraft, setHasDraft] = useState(false)

  const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const originalDocumentRef = useRef(originalDocument)
  const draftKey = getDraftKey(date)

  useEffect(() => {
    originalDocumentRef.current = originalDocument
  }, [originalDocument])

  const checkForDraft = useCallback((loadedDocument: string) => {
    const stored = localStorage.getItem(draftKey)
    if (stored) {
      try {
        const draft = JSON.parse(stored)
        if (draft.document !== loadedDocument) {
          setHasDraft(true)
        } else {
          localStorage.removeItem(draftKey)
          setHasDraft(false)
        }
      } catch {
        localStorage.removeItem(draftKey)
        setHasDraft(false)
      }
    } else {
      setHasDraft(false)
    }
  }, [draftKey])

  useEffect(() => {
    let cancelled = false

    async function loadDocument() {
      setIsLoading(true)
      setError(null)
      setValidationErrors([])

      try {
        const result = await GetEditableDocument(toWailsTime(date))
        if (!cancelled) {
          setDocumentState(result)
          setOriginalDocument(result)
          setIsLoading(false)
          checkForDraft(result)
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
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
        debounceTimerRef.current = null
      }
    }
  }, [date, checkForDraft])

  const saveDraft = useCallback(
    (doc: string) => {
      localStorage.setItem(draftKey, JSON.stringify({ document: doc, timestamp: Date.now() }))
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
        if (newDoc !== originalDocumentRef.current) {
          saveDraft(newDoc)
        }
      }, DEBOUNCE_MS)
    },
    [validateDocument, saveDraft]
  )

  const isDirty = document !== originalDocument

  const save = useCallback(async (): Promise<SaveResult> => {
    try {
      const validation = await ValidateEditableDocument(document)
      if (!validation.isValid) {
        setValidationErrors(validation.errors || [])
        return { success: false, error: 'Validation failed' }
      }

      const result = await ApplyEditableDocument(document, toWailsTime(date))

      const reloaded = await GetEditableDocument(toWailsTime(date))
      setDocumentState(reloaded)
      setOriginalDocument(reloaded)
      setLastSaved(new Date())
      clearDraft()

      return { success: true, result }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err)
      return { success: false, error: errorMessage }
    }
  }, [document, date, clearDraft])

  const saveWithActions = useCallback(async (actions: SaveActions): Promise<SaveResult> => {
    try {
      const validation = await ValidateEditableDocument(document)
      if (!validation.isValid) {
        setValidationErrors(validation.errors || [])
        return { success: false, error: 'Validation failed' }
      }

      const migrateDate = actions.migrateDate ? toWailsTime(actions.migrateDate) : null
      const listId = actions.listId ?? null

      const result = await ApplyEditableDocumentWithActions(
        document,
        toWailsTime(date),
        migrateDate as any, // eslint-disable-line @typescript-eslint/no-explicit-any -- Wails codegen types arg3 as time.Time
        listId as any // eslint-disable-line @typescript-eslint/no-explicit-any -- Wails codegen types arg4 as any
      )

      const reloaded = await GetEditableDocument(toWailsTime(date))
      setDocumentState(reloaded)
      setOriginalDocument(reloaded)
      setLastSaved(new Date())
      clearDraft()

      return { success: true, result }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err)
      return { success: false, error: errorMessage }
    }
  }, [document, date, clearDraft])

  const discardChanges = useCallback(() => {
    setDocumentState(originalDocument)
    setValidationErrors([])
  }, [originalDocument])

  const restoreDraft = useCallback(() => {
    const stored = localStorage.getItem(draftKey)
    if (stored) {
      const draft = JSON.parse(stored)
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
    originalDocument,
    setDocument,
    isLoading,
    error,
    isDirty,
    validationErrors,
    save,
    saveWithActions,
    discardChanges,
    lastSaved,
    hasDraft,
    restoreDraft,
    discardDraft,
  }
}

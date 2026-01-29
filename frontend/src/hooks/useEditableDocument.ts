import { useState, useEffect, useCallback, useRef } from 'react'
import {
  GetEditableDocumentWithEntries,
  ValidateEditableDocument,
  ApplyEditableDocument,
} from '../wailsjs/go/wails/App'
import { toWailsTime } from '@/lib/wailsTime'

interface EntryMapping {
  entityId: string
  content: string
  fullLine: string
}

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
const MIN_ENTRY_CONTENT_LENGTH = 5

const ENTRY_SYMBOLS = ['.', '-', 'o', 'x', '>']

function allEntriesHaveMinContent(doc: string): boolean {
  const lines = doc.split('\n')
  for (const line of lines) {
    const trimmed = line.trim()
    if (trimmed.length === 0) continue

    const firstChar = trimmed[0]
    if (ENTRY_SYMBOLS.includes(firstChar)) {
      const content = trimmed.slice(1).trim()
      if (content.length < MIN_ENTRY_CONTENT_LENGTH) {
        return false
      }
    }
    // Lines with invalid symbols validate immediately (they're definitely wrong)
  }
  return true
}

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
  const [entryMappings, setEntryMappings] = useState<EntryMapping[]>([])

  const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const draftKey = getDraftKey(date)

  const checkForDraft = useCallback((loadedDocument: string) => {
    const stored = localStorage.getItem(draftKey)
    if (stored) {
      try {
        const draft: Draft = JSON.parse(stored)
        // Only show draft banner if the draft content differs from loaded document
        if (draft.document !== loadedDocument || draft.deletedIds.length > 0) {
          setHasDraft(true)
        } else {
          // Draft matches loaded document, clear it
          localStorage.removeItem(draftKey)
        }
      } catch {
        // Invalid draft JSON, remove it
        localStorage.removeItem(draftKey)
      }
    }
  }, [draftKey])

  useEffect(() => {
    let cancelled = false

    async function loadDocument() {
      setIsLoading(true)
      setError(null)

      try {
        const result = await GetEditableDocumentWithEntries(toWailsTime(date))
        if (!cancelled) {
          setDocumentState(result.document)
          setOriginalDocument(result.document)
          const mappings: EntryMapping[] = (result.entries || []).map(
            (e: { entityId: string; content: string }) => ({
              entityId: e.entityId,
              content: e.content,
              fullLine: result.document
                .split('\n')
                .find((line: string) => line.includes(e.content)) || '',
            })
          )
          setEntryMappings(mappings)
          setIsLoading(false)
          checkForDraft(result.document)
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
      setDocumentState((prevDoc) => {
        // Count non-empty lines to detect actual line deletions
        // This handles the case where the only line is deleted (empty string)
        const prevNonEmptyCount = prevDoc.split('\n').filter((l) => l.trim()).length
        const newNonEmptyCount = newDoc.split('\n').filter((l) => l.trim()).length

        // Only detect deletions when non-empty line count decreases (e.g., Cmd+Shift+k delete)
        // This avoids false positives when editing existing lines
        if (newNonEmptyCount < prevNonEmptyCount) {
          const prevLines = new Set(prevDoc.split('\n'))
          const newLines = new Set(newDoc.split('\n'))

          entryMappings.forEach((mapping) => {
            const wasPresent = prevLines.has(mapping.fullLine)
            const isPresent = newLines.has(mapping.fullLine)

            if (wasPresent && !isPresent) {
              setDeletedEntries((prev) => {
                if (prev.some((e) => e.entityId === mapping.entityId)) {
                  return prev
                }
                return [...prev, { entityId: mapping.entityId, content: mapping.fullLine }]
              })
            }
          })
        }

        return newDoc
      })

      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }

      debounceTimerRef.current = setTimeout(() => {
        if (allEntriesHaveMinContent(newDoc)) {
          validateDocument(newDoc)
        }
        const deletedIds = deletedEntries.map((e) => e.entityId)
        saveDraft(newDoc, deletedIds)
      }, DEBOUNCE_MS)
    },
    [validateDocument, saveDraft, deletedEntries, entryMappings]
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
      const result = await ApplyEditableDocument(document, toWailsTime(date), deletedIds)

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

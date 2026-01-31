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
  debugLog: string[]
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
  const [debugLog, setDebugLog] = useState<string[]>([])

  const addDebug = useCallback((msg: string) => {
    const ts = new Date().toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
    setDebugLog((prev) => [...prev.slice(-29), `[${ts}] ${msg}`])
  }, [])

  const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const draftKey = getDraftKey(date)

  const checkForDraft = useCallback((loadedDocument: string) => {
    const stored = localStorage.getItem(draftKey)
    addDebug(`checkForDraft key=${draftKey} found=${!!stored}`)
    if (stored) {
      try {
        const draft: Draft = JSON.parse(stored)
        const docMatch = draft.document === loadedDocument
        addDebug(`draft.deletedIds=${JSON.stringify(draft.deletedIds)} docMatch=${docMatch}`)
        if (!docMatch) {
          addDebug(`MISMATCH: draft.doc.len=${draft.document.length} loaded.len=${loadedDocument.length}`)
          addDebug(`draft first 80: ${JSON.stringify(draft.document.slice(0, 80))}`)
          addDebug(`loaded first 80: ${JSON.stringify(loadedDocument.slice(0, 80))}`)
        }
        if (draft.document !== loadedDocument || draft.deletedIds.length > 0) {
          addDebug('=> setHasDraft(true)')
          setHasDraft(true)
        } else {
          addDebug('=> draft matches, removing from localStorage')
          localStorage.removeItem(draftKey)
          setHasDraft(false)
        }
      } catch {
        localStorage.removeItem(draftKey)
        setHasDraft(false)
      }
    } else {
      addDebug('=> no draft in localStorage')
      setHasDraft(false)
    }
  }, [draftKey, addDebug])

  useEffect(() => {
    let cancelled = false

    async function loadDocument() {
      setIsLoading(true)
      setError(null)
      setDeletedEntries([])
      setValidationErrors([])

      try {
        const result = await GetEditableDocumentWithEntries(toWailsTime(date))
        addDebug(`loadDocument for ${formatDateKey(date)}, doc.len=${result.document.length}`)
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
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
        debounceTimerRef.current = null
      }
    }
  }, [date, checkForDraft])

  const saveDraft = useCallback(
    (doc: string, deletedIds: string[]) => {
      addDebug(`saveDraft key=${draftKey} doc.len=${doc.length} deletedIds=${JSON.stringify(deletedIds)}`)
      const draft: Draft = {
        document: doc,
        deletedIds,
        timestamp: Date.now(),
      }
      localStorage.setItem(draftKey, JSON.stringify(draft))
    },
    [draftKey, addDebug]
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

  const originalDocumentRef = useRef(originalDocument)
  useEffect(() => {
    originalDocumentRef.current = originalDocument
  }, [originalDocument])

  const setDocument = useCallback(
    (newDoc: string) => {
      addDebug(`setDocument called, newDoc.len=${newDoc.length}`)
      setDocumentState((prevDoc) => {
        const prevNonEmptyCount = prevDoc.split('\n').filter((l) => l.trim()).length
        const newNonEmptyCount = newDoc.split('\n').filter((l) => l.trim()).length

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

        setDeletedEntries((prev) => {
          if (prev.length === 0) return prev
          const filtered = prev.filter(
            (entry) => !newDoc.includes(`[${entry.entityId}]`)
          )
          return filtered.length === prev.length ? prev : filtered
        })

        return newDoc
      })

      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }

      debounceTimerRef.current = setTimeout(() => {
        if (allEntriesHaveMinContent(newDoc)) {
          validateDocument(newDoc)
        }
        const matchesOriginal = newDoc === originalDocumentRef.current
        addDebug(`debounce: matchesOriginal=${matchesOriginal} newDoc.len=${newDoc.length} orig.len=${originalDocumentRef.current.length}`)
        if (!matchesOriginal) {
          const deletedIds = deletedEntries.map((e) => e.entityId)
          saveDraft(newDoc, deletedIds)
        }
      }, DEBOUNCE_MS)
    },
    [validateDocument, saveDraft, deletedEntries, entryMappings, addDebug]
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
    debugLog,
  }
}

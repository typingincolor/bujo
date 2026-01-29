import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { useEditableDocument } from './useEditableDocument'

const mockGetEditableDocument = vi.fn()
const mockGetEditableDocumentWithEntries = vi.fn()
const mockValidateEditableDocument = vi.fn()
const mockApplyEditableDocument = vi.fn()

vi.mock('../wailsjs/go/wails/App', () => ({
  GetEditableDocument: (...args: unknown[]) => mockGetEditableDocument(...args),
  GetEditableDocumentWithEntries: (...args: unknown[]) => mockGetEditableDocumentWithEntries(...args),
  ValidateEditableDocument: (...args: unknown[]) => mockValidateEditableDocument(...args),
  ApplyEditableDocument: (...args: unknown[]) => mockApplyEditableDocument(...args),
}))

const mockLocalStorage: Record<string, string> = {}
const localStorageMock = {
  getItem: vi.fn((key: string) => mockLocalStorage[key] || null),
  setItem: vi.fn((key: string, value: string) => {
    mockLocalStorage[key] = value
  }),
  removeItem: vi.fn((key: string) => {
    delete mockLocalStorage[key]
  }),
  clear: vi.fn(() => {
    Object.keys(mockLocalStorage).forEach((key) => delete mockLocalStorage[key])
  }),
}

Object.defineProperty(window, 'localStorage', { value: localStorageMock })

describe('useEditableDocument', () => {
  const testDate = new Date(2026, 0, 27) // Jan 27, 2026

  beforeEach(() => {
    vi.clearAllMocks()
    localStorageMock.clear()
    mockGetEditableDocument.mockResolvedValue('. Buy groceries\n- Meeting notes')
    mockGetEditableDocumentWithEntries.mockResolvedValue({
      document: '. Buy groceries\n- Meeting notes',
      entries: [
        { entityId: 'entity-grocery', content: 'Buy groceries' },
        { entityId: 'entity-meeting', content: 'Meeting notes' },
      ],
    })
    mockValidateEditableDocument.mockResolvedValue({
      isValid: true,
      errors: [],
    })
    mockApplyEditableDocument.mockResolvedValue({
      inserted: 0,
      updated: 1,
      deleted: 0,
      migrated: 0,
    })
  })

  describe('initial loading', () => {
    it('loads document on mount', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      expect(result.current.isLoading).toBe(true)

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.document).toBe('. Buy groceries\n- Meeting notes')
      expect(mockGetEditableDocumentWithEntries).toHaveBeenCalled()
    })

    it('sets error state on load failure', async () => {
      mockGetEditableDocumentWithEntries.mockRejectedValue(new Error('Network error'))

      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.error).toBe('Network error')
    })
  })

  describe('dirty state tracking', () => {
    it('starts not dirty after load', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.isDirty).toBe(false)
    })

    it('becomes dirty when document changes', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.setDocument('. Modified content')
      })

      expect(result.current.isDirty).toBe(true)
    })

    it('resets dirty state after save', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.setDocument('. Modified content')
      })

      expect(result.current.isDirty).toBe(true)

      await act(async () => {
        await result.current.save()
      })

      expect(result.current.isDirty).toBe(false)
    })
  })

  describe('validation', () => {
    it('does not validate when any entry has fewer than 5 characters of content', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      vi.useFakeTimers()

      // Entry content "Hi" is only 2 characters (symbol doesn't count)
      act(() => {
        result.current.setDocument('. Hi')
      })

      await act(async () => {
        await vi.advanceTimersByTimeAsync(500)
      })

      expect(mockValidateEditableDocument).not.toHaveBeenCalled()

      vi.useRealTimers()
    })

    it('validates when all entries have 5 or more characters of content', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      vi.useFakeTimers()

      // Entry content "Hello" is 5 characters
      act(() => {
        result.current.setDocument('. Hello')
      })

      await act(async () => {
        await vi.advanceTimersByTimeAsync(500)
      })

      expect(mockValidateEditableDocument).toHaveBeenCalledWith('. Hello')

      vi.useRealTimers()
    })

    it('does not validate when one entry is complete but another is short', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      vi.useFakeTimers()

      // First entry is complete, second entry "Hi" is only 2 characters
      act(() => {
        result.current.setDocument('. Complete task\n. Hi')
      })

      await act(async () => {
        await vi.advanceTimersByTimeAsync(500)
      })

      expect(mockValidateEditableDocument).not.toHaveBeenCalled()

      vi.useRealTimers()
    })


    it('validates document after change with debounce', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      vi.useFakeTimers()

      act(() => {
        result.current.setDocument('. Changed content')
      })

      expect(mockValidateEditableDocument).not.toHaveBeenCalled()

      await act(async () => {
        await vi.advanceTimersByTimeAsync(500)
      })

      expect(mockValidateEditableDocument).toHaveBeenCalledWith('. Changed content')

      vi.useRealTimers()
    })

    it('exposes validation errors', async () => {
      mockValidateEditableDocument.mockResolvedValue({
        isValid: false,
        errors: [{ lineNumber: 1, message: 'Unknown entry type' }],
      })

      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      vi.useFakeTimers()

      act(() => {
        result.current.setDocument('^ Invalid line')
      })

      await act(async () => {
        await vi.advanceTimersByTimeAsync(500)
      })

      vi.useRealTimers()

      await waitFor(() => {
        expect(result.current.validationErrors).toHaveLength(1)
      })

      expect(result.current.validationErrors[0].lineNumber).toBe(1)
    })
  })

  describe('deletion tracking', () => {
    it('starts with no deletions', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.deletedEntries).toHaveLength(0)
    })

    it('auto-detects deletion when line is removed from document', async () => {
      // Setup: return document with entries mapping
      mockGetEditableDocumentWithEntries.mockResolvedValue({
        document: '. Buy groceries\n- Meeting notes',
        entries: [
          { entityId: 'entity-grocery', content: 'Buy groceries' },
          { entityId: 'entity-meeting', content: 'Meeting notes' },
        ],
      })

      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      // User removes the first line
      act(() => {
        result.current.setDocument('- Meeting notes')
      })

      // Deletion should be auto-tracked
      expect(result.current.deletedEntries).toHaveLength(1)
      expect(result.current.deletedEntries[0].entityId).toBe('entity-grocery')
      expect(result.current.deletedEntries[0].content).toBe('. Buy groceries')
    })

    it('tracks deleted line with entity ID', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.trackDeletion('entity-123', '. Deleted task')
      })

      expect(result.current.deletedEntries).toHaveLength(1)
      expect(result.current.deletedEntries[0].entityId).toBe('entity-123')
      expect(result.current.deletedEntries[0].content).toBe('. Deleted task')
    })

    it('can restore deleted entry', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.trackDeletion('entity-123', '. Deleted task')
      })

      expect(result.current.deletedEntries).toHaveLength(1)

      act(() => {
        result.current.restoreDeletion('entity-123')
      })

      expect(result.current.deletedEntries).toHaveLength(0)
    })

    it('clears deletions after save', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.trackDeletion('entity-123', '. Deleted task')
      })

      await act(async () => {
        await result.current.save()
      })

      expect(result.current.deletedEntries).toHaveLength(0)
    })

    it('detects all entries as deleted when document is cleared to empty in one step', async () => {
      mockGetEditableDocumentWithEntries.mockResolvedValue({
        document: '. Buy groceries\n- Meeting notes',
        entries: [
          { entityId: 'entity-grocery', content: 'Buy groceries' },
          { entityId: 'entity-meeting', content: 'Meeting notes' },
        ],
      })

      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      // Verify initial state has entries
      expect(result.current.document).toBe('. Buy groceries\n- Meeting notes')
      expect(result.current.deletedEntries).toHaveLength(0)

      // Simulate clearing document in one step (e.g., Cmd+A, Delete)
      act(() => {
        result.current.setDocument('')
      })

      // Verify document is now empty
      expect(result.current.document).toBe('')

      // Both entries should be marked for deletion
      expect(result.current.deletedEntries).toHaveLength(2)
      expect(result.current.deletedEntries.map(e => e.entityId)).toContain('entity-grocery')
      expect(result.current.deletedEntries.map(e => e.entityId)).toContain('entity-meeting')
    })

    it('un-tracks deletion when entity ID reappears after paste', async () => {
      mockGetEditableDocumentWithEntries.mockResolvedValue({
        document: '[entity-event] o Event 2x\n[entity-note] - Note 5',
        entries: [
          { entityId: 'entity-event', content: 'Event 2x' },
          { entityId: 'entity-note', content: 'Note 5' },
        ],
      })

      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      // Step 1: Cut — removes event line, triggers deletion detection
      act(() => {
        result.current.setDocument('[entity-note] - Note 5')
      })

      // After cut, entity-event is detected as deleted
      expect(result.current.deletedEntries).toHaveLength(1)
      expect(result.current.deletedEntries[0].entityId).toBe('entity-event')

      // Step 2: Paste — entity ID reappears beneath note (with indentation)
      act(() => {
        result.current.setDocument('[entity-note] - Note 5\n  [entity-event] o Event 2x')
      })

      // Entity ID is back in the document — should be un-tracked
      expect(result.current.deletedEntries).toHaveLength(0)
    })

    it('still detects deletion when entity ID is truly removed from document', async () => {
      mockGetEditableDocumentWithEntries.mockResolvedValue({
        document: '[entity-event] o Event 2x\n[entity-note] - Note 5',
        entries: [
          { entityId: 'entity-event', content: 'Event 2x' },
          { entityId: 'entity-note', content: 'Note 5' },
        ],
      })

      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      // User deletes the event line entirely (no paste back)
      act(() => {
        result.current.setDocument('[entity-note] - Note 5')
      })

      // Entity ID is gone — should be detected as deleted
      expect(result.current.deletedEntries).toHaveLength(1)
      expect(result.current.deletedEntries[0].entityId).toBe('entity-event')
    })

    it('sends detected deletions to backend when document cleared in one step', async () => {
      // When the document is cleared in one step, the frontend detects the deletions
      // The backend also auto-detects missing entries as a safety net
      mockGetEditableDocumentWithEntries.mockResolvedValue({
        document: '. Task\n- Note',
        entries: [
          { entityId: 'entity-task', content: 'Task' },
          { entityId: 'entity-note', content: 'Note' },
        ],
      })

      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      // Clear document in one step - frontend detects both entries as deleted
      act(() => result.current.setDocument(''))

      // Save the empty document with detected deletions
      await act(async () => {
        await result.current.save()
      })

      // Frontend detected the deletions and sends them to backend
      expect(mockApplyEditableDocument).toHaveBeenCalledWith(
        '',
        expect.any(String),
        expect.arrayContaining(['entity-task', 'entity-note'])
      )
    })
  })

  describe('saving', () => {
    it('calls API with document and deletions', async () => {
      mockGetEditableDocumentWithEntries.mockResolvedValue({
        document: '. Original task',
        entries: [],
      })

      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.setDocument('. Updated task')
        result.current.trackDeletion('entity-456', '. Old task')
      })

      await act(async () => {
        await result.current.save()
      })

      expect(mockApplyEditableDocument).toHaveBeenCalledWith(
        '. Updated task',
        expect.any(String),
        ['entity-456']
      )
    })

    it('validates before saving', async () => {
      mockValidateEditableDocument.mockResolvedValue({
        isValid: false,
        errors: [{ lineNumber: 1, message: 'Invalid' }],
      })

      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.setDocument('^ Invalid')
      })

      await act(async () => {
        const saveResult = await result.current.save()
        expect(saveResult.success).toBe(false)
      })

      expect(mockApplyEditableDocument).not.toHaveBeenCalled()
    })

    it('returns save result with counts', async () => {
      mockApplyEditableDocument.mockResolvedValue({
        inserted: 2,
        updated: 1,
        deleted: 1,
        migrated: 0,
      })

      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.setDocument('. New content')
      })

      let saveResult: Awaited<ReturnType<typeof result.current.save>>
      await act(async () => {
        saveResult = await result.current.save()
      })

      expect(saveResult!.success).toBe(true)
      expect(saveResult!.result?.inserted).toBe(2)
    })

    it('updates lastSaved timestamp after successful save', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.lastSaved).toBeNull()

      await act(async () => {
        await result.current.save()
      })

      expect(result.current.lastSaved).toBeInstanceOf(Date)
    })
  })

  describe('discard changes', () => {
    it('restores original document', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.setDocument('. Modified')
        result.current.trackDeletion('entity-1', '. Old')
      })

      expect(result.current.isDirty).toBe(true)

      act(() => {
        result.current.discardChanges()
      })

      expect(result.current.document).toBe('. Buy groceries\n- Meeting notes')
      expect(result.current.isDirty).toBe(false)
      expect(result.current.deletedEntries).toHaveLength(0)
    })
  })

  describe('crash recovery (localStorage)', () => {
    it('saves draft to localStorage on change', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      vi.useFakeTimers()

      act(() => {
        result.current.setDocument('. Unsaved content')
      })

      await act(async () => {
        await vi.advanceTimersByTimeAsync(500)
      })

      vi.useRealTimers()

      expect(localStorageMock.setItem).toHaveBeenCalled()
      const savedDraft = JSON.parse(mockLocalStorage['bujo.draft.2026-01-27'])
      expect(savedDraft.document).toBe('. Unsaved content')
    })

    it('detects existing draft on load', async () => {
      mockLocalStorage['bujo.draft.2026-01-27'] = JSON.stringify({
        document: '. Recovered content',
        deletedIds: [],
        timestamp: Date.now(),
      })

      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      expect(result.current.hasDraft).toBe(true)
    })

    it('restores draft when requested', async () => {
      mockLocalStorage['bujo.draft.2026-01-27'] = JSON.stringify({
        document: '. Recovered content',
        deletedIds: ['entity-old'],
        timestamp: Date.now(),
      })

      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.restoreDraft()
      })

      expect(result.current.document).toBe('. Recovered content')
      expect(result.current.isDirty).toBe(true)
    })

    it('discards draft when requested', async () => {
      mockLocalStorage['bujo.draft.2026-01-27'] = JSON.stringify({
        document: '. Old draft',
        deletedIds: [],
        timestamp: Date.now(),
      })

      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.discardDraft()
      })

      expect(result.current.hasDraft).toBe(false)
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('bujo.draft.2026-01-27')
    })

    it('clears draft after successful save', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      vi.useFakeTimers()

      act(() => {
        result.current.setDocument('. Changed')
      })

      await act(async () => {
        await vi.advanceTimersByTimeAsync(500)
      })

      vi.useRealTimers()

      localStorageMock.removeItem.mockClear()

      await act(async () => {
        await result.current.save()
      })

      expect(localStorageMock.removeItem).toHaveBeenCalledWith('bujo.draft.2026-01-27')
    })
  })
})

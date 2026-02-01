import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { useEditableDocument } from './useEditableDocument'

const mockGetEditableDocument = vi.fn()
const mockValidateEditableDocument = vi.fn()
const mockApplyEditableDocument = vi.fn()

vi.mock('../wailsjs/go/wails/App', () => ({
  GetEditableDocument: (...args: unknown[]) => mockGetEditableDocument(...args),
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
    mockValidateEditableDocument.mockResolvedValue({
      isValid: true,
      errors: [],
    })
    mockApplyEditableDocument.mockResolvedValue({
      inserted: 0,
      deleted: 0,
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
      expect(mockGetEditableDocument).toHaveBeenCalled()
    })

    it('sets error state on load failure', async () => {
      mockGetEditableDocument.mockRejectedValue(new Error('Network error'))

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

  describe('saving', () => {
    it('calls API with document and date', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      act(() => {
        result.current.setDocument('. Updated task')
      })

      await act(async () => {
        await result.current.save()
      })

      expect(mockApplyEditableDocument).toHaveBeenCalledWith(
        '. Updated task',
        expect.any(String)
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
        deleted: 1,
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
      })

      expect(result.current.isDirty).toBe(true)

      act(() => {
        result.current.discardChanges()
      })

      expect(result.current.document).toBe('. Buy groceries\n- Meeting notes')
      expect(result.current.isDirty).toBe(false)
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

    it('does not save draft when document matches original', async () => {
      const { result } = renderHook(() => useEditableDocument(testDate))

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      vi.useFakeTimers()

      // "Edit" the document to the exact same content (e.g., CodeMirror re-fires onChange)
      act(() => {
        result.current.setDocument('. Buy groceries\n- Meeting notes')
      })

      await act(async () => {
        await vi.advanceTimersByTimeAsync(500)
      })

      vi.useRealTimers()

      // No draft should be saved since document is unchanged
      expect(mockLocalStorage['bujo.draft.2026-01-27']).toBeUndefined()
    })

    it('resets hasDraft to false when navigating to a day with no draft', async () => {
      const friday = new Date(2026, 0, 30)
      const thursday = new Date(2026, 0, 29)

      // Friday has a draft in localStorage
      mockLocalStorage['bujo.draft.2026-01-30'] = JSON.stringify({
        document: '. Unsaved friday edit',
        timestamp: Date.now(),
      })

      // Thursday has no draft
      // (no entry in mockLocalStorage for bujo.draft.2026-01-29)

      mockGetEditableDocument.mockResolvedValue('. Some content')

      const { result, rerender } = renderHook(
        ({ date }) => useEditableDocument(date),
        { initialProps: { date: friday } }
      )

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      // Friday should show draft banner since draft differs from loaded doc
      expect(result.current.hasDraft).toBe(true)

      // Navigate to Thursday
      rerender({ date: thursday })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false)
      })

      // Thursday has no draft — banner should NOT appear
      expect(result.current.hasDraft).toBe(false)
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

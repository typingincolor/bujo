import { renderHook, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useAttentionScores } from './useAttentionScores'

vi.mock('@/wailsjs/go/wails/App', () => ({
  GetAttentionScores: vi.fn(),
}))

import { GetAttentionScores } from '@/wailsjs/go/wails/App'

const mockGetAttentionScores = vi.mocked(GetAttentionScores)

describe('useAttentionScores', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns empty map when no entry IDs provided', async () => {
    const { result } = renderHook(() => useAttentionScores([]))
    expect(result.current.scores).toEqual({})
    expect(mockGetAttentionScores).not.toHaveBeenCalled()
  })

  it('fetches scores from backend and returns camelCase results', async () => {
    mockGetAttentionScores.mockResolvedValue({
      1: { Score: 75, Indicators: ['overdue', 'priority'], DaysOld: 5 },
      2: { Score: 30, Indicators: ['aging'], DaysOld: 8 },
    })

    const { result } = renderHook(() => useAttentionScores([1, 2]))

    await waitFor(() => {
      expect(result.current.scores[1]).toBeDefined()
    })

    expect(result.current.scores[1]).toEqual({
      score: 75,
      indicators: ['overdue', 'priority'],
      daysOld: 5,
    })
    expect(result.current.scores[2]).toEqual({
      score: 30,
      indicators: ['aging'],
      daysOld: 8,
    })
  })

  it('calls GetAttentionScores with entry IDs', async () => {
    mockGetAttentionScores.mockResolvedValue({})
    renderHook(() => useAttentionScores([10, 20, 30]))

    await waitFor(() => {
      expect(mockGetAttentionScores).toHaveBeenCalledWith([10, 20, 30])
    })
  })

  it('indicates loading state while fetching', async () => {
    let resolvePromise: (value: Record<number, unknown>) => void
    mockGetAttentionScores.mockReturnValue(
      new Promise((resolve) => { resolvePromise = resolve })
    )

    const { result } = renderHook(() => useAttentionScores([1]))

    expect(result.current.loading).toBe(true)

    resolvePromise!({ 1: { Score: 10, Indicators: [], DaysOld: 1 } })

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })
  })

  it('refetches when entry IDs change', async () => {
    mockGetAttentionScores.mockResolvedValue({
      1: { Score: 50, Indicators: [], DaysOld: 0 },
    })

    const { result, rerender } = renderHook(
      ({ ids }) => useAttentionScores(ids),
      { initialProps: { ids: [1] } }
    )

    await waitFor(() => {
      expect(result.current.scores[1]).toBeDefined()
    })

    mockGetAttentionScores.mockResolvedValue({
      2: { Score: 60, Indicators: ['priority'], DaysOld: 3 },
    })

    rerender({ ids: [2] })

    await waitFor(() => {
      expect(result.current.scores[2]).toBeDefined()
    })

    expect(mockGetAttentionScores).toHaveBeenCalledTimes(2)
  })

  it('returns default score for entries not in backend response', async () => {
    mockGetAttentionScores.mockResolvedValue({
      1: { Score: 50, Indicators: ['overdue'], DaysOld: 3 },
    })

    const { result } = renderHook(() => useAttentionScores([1, 999]))

    await waitFor(() => {
      expect(result.current.scores[1]).toBeDefined()
    })

    expect(result.current.scores[999]).toBeUndefined()
  })
})

import { renderHook, act } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { useNavigationHistory } from '../useNavigationHistory'

describe('useNavigationHistory', () => {
  it('starts with no history', () => {
    const { result } = renderHook(() => useNavigationHistory())

    expect(result.current.canGoBack).toBe(false)
    expect(result.current.history).toBeNull()
  })

  it('stores navigation state when pushing', () => {
    const { result } = renderHook(() => useNavigationHistory())

    act(() => {
      result.current.pushHistory({ view: 'week', scrollPosition: 100, entryId: 42 })
    })

    expect(result.current.canGoBack).toBe(true)
    expect(result.current.history).toEqual({ view: 'week', scrollPosition: 100, entryId: 42 })
  })

  it('clears history on goBack', () => {
    const { result } = renderHook(() => useNavigationHistory())

    act(() => {
      result.current.pushHistory({ view: 'week', scrollPosition: 100, entryId: 42 })
    })

    const returned = act(() => result.current.goBack())

    expect(result.current.canGoBack).toBe(false)
    expect(result.current.history).toBeNull()
  })

  it('returns history state on goBack', () => {
    const { result } = renderHook(() => useNavigationHistory())

    act(() => {
      result.current.pushHistory({ view: 'week', scrollPosition: 100, entryId: 42 })
    })

    let returnedHistory: ReturnType<typeof result.current.goBack>
    act(() => {
      returnedHistory = result.current.goBack()
    })

    expect(returnedHistory!).toEqual({ view: 'week', scrollPosition: 100, entryId: 42 })
  })

  it('clearHistory removes history without returning it', () => {
    const { result } = renderHook(() => useNavigationHistory())

    act(() => {
      result.current.pushHistory({ view: 'week', scrollPosition: 100, entryId: 42 })
    })

    act(() => {
      result.current.clearHistory()
    })

    expect(result.current.canGoBack).toBe(false)
  })
})

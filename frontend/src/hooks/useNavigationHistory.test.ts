import { renderHook, act } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { useNavigationHistory } from './useNavigationHistory'

describe('useNavigationHistory', () => {
  it('starts with no history and canGoBack false', () => {
    const { result } = renderHook(() => useNavigationHistory())
    expect(result.current.canGoBack).toBe(false)
    expect(result.current.history).toBeNull()
  })

  it('pushHistory makes canGoBack true', () => {
    const { result } = renderHook(() => useNavigationHistory())
    act(() => {
      result.current.pushHistory({ view: 'week', scrollPosition: 0 })
    })
    expect(result.current.canGoBack).toBe(true)
  })

  it('goBack returns the last pushed state', () => {
    const { result } = renderHook(() => useNavigationHistory())
    act(() => {
      result.current.pushHistory({ view: 'week', scrollPosition: 100 })
    })
    let state: ReturnType<typeof result.current.goBack>
    act(() => {
      state = result.current.goBack()
    })
    expect(state!).toEqual({ view: 'week', scrollPosition: 100 })
    expect(result.current.canGoBack).toBe(false)
  })

  it('supports multiple pushes and pops in stack order', () => {
    const { result } = renderHook(() => useNavigationHistory())
    act(() => {
      result.current.pushHistory({ view: 'pending', scrollPosition: 50 })
      result.current.pushHistory({ view: 'week', scrollPosition: 200 })
    })
    let state: ReturnType<typeof result.current.goBack>
    act(() => {
      state = result.current.goBack()
    })
    expect(state!).toEqual({ view: 'week', scrollPosition: 200 })
    expect(result.current.canGoBack).toBe(true)

    act(() => {
      state = result.current.goBack()
    })
    expect(state!).toEqual({ view: 'pending', scrollPosition: 50 })
    expect(result.current.canGoBack).toBe(false)
  })

  it('clearHistory removes all entries', () => {
    const { result } = renderHook(() => useNavigationHistory())
    act(() => {
      result.current.pushHistory({ view: 'week', scrollPosition: 0 })
      result.current.pushHistory({ view: 'pending', scrollPosition: 0 })
    })
    act(() => {
      result.current.clearHistory()
    })
    expect(result.current.canGoBack).toBe(false)
  })
})

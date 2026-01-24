import { useState, useCallback } from 'react'
import { ViewType } from '@/components/bujo/Sidebar'

export interface NavigationState {
  view: ViewType
  scrollPosition: number
  entryId?: number
}

const MAX_HISTORY_DEPTH = 50

export function useNavigationHistory() {
  const [history, setHistory] = useState<NavigationState[]>([])

  const pushHistory = useCallback((state: NavigationState) => {
    setHistory((prev) => {
      const updated = [...prev, state]
      return updated.length > MAX_HISTORY_DEPTH
        ? updated.slice(-MAX_HISTORY_DEPTH)
        : updated
    })
  }, [])

  const goBack = useCallback(() => {
    const current = history[history.length - 1] || null
    setHistory((prev) => prev.slice(0, -1))
    return current
  }, [history])

  const clearHistory = useCallback(() => {
    setHistory([])
  }, [])

  return {
    history: history.length > 0 ? history[history.length - 1] : null,
    canGoBack: history.length > 0,
    pushHistory,
    goBack,
    clearHistory,
  }
}

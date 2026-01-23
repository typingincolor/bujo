import { useState, useCallback } from 'react'
import { ViewType } from '@/components/bujo/Sidebar'

export interface NavigationState {
  view: ViewType
  scrollPosition: number
  entryId?: number
}

export function useNavigationHistory() {
  const [history, setHistory] = useState<NavigationState | null>(null)

  const pushHistory = useCallback((state: NavigationState) => {
    setHistory(state)
  }, [])

  const goBack = useCallback(() => {
    const current = history
    setHistory(null)
    return current
  }, [history])

  const clearHistory = useCallback(() => {
    setHistory(null)
  }, [])

  return {
    history,
    canGoBack: history !== null,
    pushHistory,
    goBack,
    clearHistory,
  }
}

import { useState, useEffect, useRef } from 'react'
import { GetAttentionScores } from '@/wailsjs/go/wails/App'

export interface AttentionScore {
  score: number
  indicators: string[]
  daysOld: number
}

function mapScores(result: Record<string, { Score: number; Indicators: string[]; DaysOld: number }>): Record<number, AttentionScore> {
  const mapped: Record<number, AttentionScore> = {}
  for (const [id, raw] of Object.entries(result)) {
    mapped[Number(id)] = {
      score: raw.Score,
      indicators: raw.Indicators,
      daysOld: raw.DaysOld,
    }
  }
  return mapped
}

export function useAttentionScores(entryIds: number[]) {
  const [scores, setScores] = useState<Record<number, AttentionScore>>({})
  const prevIdsRef = useRef<string>('')

  useEffect(() => {
    const idsKey = JSON.stringify(entryIds)
    if (idsKey === prevIdsRef.current) return
    prevIdsRef.current = idsKey

    if (entryIds.length === 0) {
      return
    }

    let cancelled = false
    GetAttentionScores(entryIds).then((result) => {
      if (!cancelled) {
        setScores(mapScores(result))
      }
    })
    return () => { cancelled = true }
  }, [entryIds])

  return { scores }
}

import { useState, useEffect, useRef } from 'react'
import { GetAttentionScores } from '@/wailsjs/go/wails/App'

export interface AttentionScore {
  score: number
  indicators: string[]
  daysOld: number
}

export function useAttentionScores(entryIds: number[]) {
  const [scores, setScores] = useState<Record<number, AttentionScore>>({})
  const [loading, setLoading] = useState(false)
  const prevIdsRef = useRef<string>('')

  useEffect(() => {
    const idsKey = JSON.stringify(entryIds)
    if (idsKey === prevIdsRef.current) return
    prevIdsRef.current = idsKey

    if (entryIds.length === 0) {
      setScores({})
      setLoading(false)
      return
    }

    setLoading(true)
    GetAttentionScores(entryIds).then((result) => {
      const mapped: Record<number, AttentionScore> = {}
      for (const [id, raw] of Object.entries(result)) {
        mapped[Number(id)] = {
          score: raw.Score,
          indicators: raw.Indicators,
          daysOld: raw.DaysOld,
        }
      }
      setScores(mapped)
      setLoading(false)
    })
  }, [entryIds])

  return { scores, loading }
}

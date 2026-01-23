import { Entry } from '@/types/bujo'

export type AttentionIndicator = 'overdue' | 'priority' | 'aging' | 'migrated'

export interface AttentionResult {
  score: number
  indicators: AttentionIndicator[]
  migrationCount?: number
  daysOld?: number
}

interface ExtendedEntry extends Entry {
  scheduledDate?: string
  migrationCount?: number
  entityId?: string
  depth?: number
}

const URGENT_KEYWORDS = ['urgent', 'asap', 'blocker', 'waiting', 'blocked']

export function calculateAttentionScore(
  entry: ExtendedEntry,
  now: Date,
  parentType?: string
): AttentionResult {
  let score = 0
  const indicators: AttentionIndicator[] = []

  // Past scheduled date: +50
  if (entry.scheduledDate) {
    const scheduled = new Date(entry.scheduledDate)
    if (scheduled < now) {
      score += 50
      indicators.push('overdue')
    }
  }

  // Priority set: +30, high/urgent: additional +20
  if (entry.priority && entry.priority !== '' && entry.priority !== 'none') {
    score += 30
    indicators.push('priority')
    if (entry.priority === 'high') {
      score += 20
    }
  }

  // Age calculations
  const loggedDate = new Date(entry.loggedDate)
  const daysOld = Math.floor((now.getTime() - loggedDate.getTime()) / (1000 * 60 * 60 * 24))

  if (daysOld > 7) {
    score += 25
    indicators.push('aging')
  } else if (daysOld > 3) {
    score += 15
    indicators.push('aging')
  }

  // Migration count: +15 per migration
  if (entry.migrationCount && entry.migrationCount > 0) {
    score += entry.migrationCount * 15
    indicators.push('migrated')
  }

  // Urgent keywords: +20
  const contentLower = entry.content.toLowerCase()
  if (URGENT_KEYWORDS.some(keyword => contentLower.includes(keyword))) {
    score += 20
  }

  // Questions: +10
  if (entry.type === 'question') {
    score += 10
  }

  // Parent is event: +5
  if (entry.parentId && parentType === 'event') {
    score += 5
  }

  return {
    score,
    indicators,
    migrationCount: entry.migrationCount,
    daysOld,
  }
}

export function sortByAttentionScore(
  entries: ExtendedEntry[],
  now: Date,
  parentTypes?: Map<number, string>
): ExtendedEntry[] {
  return [...entries].sort((a, b) => {
    const scoreA = calculateAttentionScore(a, now, parentTypes?.get(a.parentId ?? 0))
    const scoreB = calculateAttentionScore(b, now, parentTypes?.get(b.parentId ?? 0))
    return scoreB.score - scoreA.score
  })
}

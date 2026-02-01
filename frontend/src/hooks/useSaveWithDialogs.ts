export interface SpecialEntries {
  migratedEntries: string[]
  movedToListEntries: string[]
  hasSpecialEntries: boolean
}

export function scanForSpecialEntries(doc: string): SpecialEntries {
  const lines = doc.split('\n')
  const migratedEntries: string[] = []
  const movedToListEntries: string[] = []

  for (const line of lines) {
    const match = line.match(/^\s*([>^])\s+(.+)/)
    if (match) {
      const [, symbol, content] = match
      if (symbol === '>') migratedEntries.push(content.trim())
      if (symbol === '^') movedToListEntries.push(content.trim())
    }
  }

  return {
    migratedEntries,
    movedToListEntries,
    hasSpecialEntries: migratedEntries.length > 0 || movedToListEntries.length > 0,
  }
}

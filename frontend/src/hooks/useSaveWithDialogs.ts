export interface SpecialEntries {
  migratedEntries: string[]
  movedToListEntries: string[]
  hasSpecialEntries: boolean
}

export function scanForNewSpecialEntries(currentDoc: string, originalDoc: string): SpecialEntries {
  const current = scanForSpecialEntries(currentDoc)
  const original = scanForSpecialEntries(originalDoc)

  const remainingMigrated = [...original.migratedEntries]
  const newMigrated = current.migratedEntries.filter(entry => {
    const idx = remainingMigrated.indexOf(entry)
    if (idx !== -1) {
      remainingMigrated.splice(idx, 1)
      return false
    }
    return true
  })

  const remainingMoved = [...original.movedToListEntries]
  const newMoved = current.movedToListEntries.filter(entry => {
    const idx = remainingMoved.indexOf(entry)
    if (idx !== -1) {
      remainingMoved.splice(idx, 1)
      return false
    }
    return true
  })

  return {
    migratedEntries: newMigrated,
    movedToListEntries: newMoved,
    hasSpecialEntries: newMigrated.length > 0 || newMoved.length > 0,
  }
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

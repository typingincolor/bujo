const ENTRY_SYMBOL_PATTERN = /^(\s*)[.\-oxX>?~]([!*^]?)(.+)$/

export interface LineMatch {
  line: number
  from: number
  to: number
}

export function findEntryLine(doc: string, searchText: string): LineMatch | null {
  if (!doc || !searchText) return null

  const lines = doc.split('\n')
  let offset = 0

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]
    const match = line.match(ENTRY_SYMBOL_PATTERN)
    if (match) {
      const content = match[3].trim()
      if (content === searchText || content.includes(searchText)) {
        return {
          line: i + 1,
          from: offset,
          to: offset + line.length,
        }
      }
    }
    offset += line.length + 1
  }

  return null
}

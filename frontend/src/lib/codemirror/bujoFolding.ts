export interface FoldRange {
  from: number
  to: number
}

const INDENT_SIZE = 2

function getIndentDepth(line: string): number {
  const leadingSpaces = line.match(/^(\s*)/)?.[1].length ?? 0
  return Math.floor(leadingSpaces / INDENT_SIZE)
}

export function getFoldRange(lines: string[], lineIndex: number): FoldRange | null {
  const currentLine = lines[lineIndex]
  if (!currentLine || currentLine.trim() === '') return null
  if (lineIndex >= lines.length - 1) return null

  const currentDepth = getIndentDepth(currentLine)
  let lastChildIndex = lineIndex

  for (let i = lineIndex + 1; i < lines.length; i++) {
    const line = lines[i]
    if (line.trim() === '') continue

    const depth = getIndentDepth(line)
    if (depth <= currentDepth) break

    lastChildIndex = i
  }

  if (lastChildIndex === lineIndex) return null

  return { from: lineIndex, to: lastChildIndex }
}

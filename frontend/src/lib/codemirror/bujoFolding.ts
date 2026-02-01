import { EditorState, Extension } from '@codemirror/state'
import { foldService, foldedRanges } from '@codemirror/language'
import { codeFolding, foldGutter, foldKeymap } from '@codemirror/language'
import { keymap } from '@codemirror/view'

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

function bujoFoldServiceFn(state: EditorState, lineStart: number, _lineEnd: number): { from: number; to: number } | null {
  const doc = state.doc
  const lines: string[] = []
  for (let i = 1; i <= doc.lines; i++) {
    lines.push(doc.line(i).text)
  }

  const lineObj = doc.lineAt(lineStart)
  const lineIndex = lineObj.number - 1

  const range = getFoldRange(lines, lineIndex)
  if (!range) return null

  return {
    from: lineObj.to,
    to: doc.line(range.to + 1).to,
  }
}

export function expandRangeForFolds(state: EditorState, from: number, to: number): { from: number; to: number } {
  const folded = foldedRanges(state)
  if (folded.size === 0) return { from, to }

  let expandedFrom = from
  let expandedTo = to

  const cursor = folded.iter()
  while (cursor.value) {
    if (cursor.from >= expandedFrom && cursor.from <= expandedTo) {
      expandedTo = Math.max(expandedTo, cursor.to)
    }
    if (cursor.to >= expandedFrom && cursor.to <= expandedTo) {
      expandedFrom = Math.min(expandedFrom, cursor.from)
    }
    cursor.next()
  }

  return { from: expandedFrom, to: expandedTo }
}

export function bujoFoldExtension(): Extension {
  return [
    foldService.of(bujoFoldServiceFn),
    codeFolding(),
    foldGutter(),
    keymap.of(foldKeymap),
  ]
}

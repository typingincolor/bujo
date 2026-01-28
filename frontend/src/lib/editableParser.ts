export type EntrySymbol = '.' | '-' | 'o' | 'x' | '~' | '?' | '>'

export type Priority = 0 | 1 | 2 | 3

export interface ParsedLine {
  lineNumber: number
  raw: string
  depth: number
  symbol: EntrySymbol | null
  priority: Priority
  content: string
  migrationTarget: string | null
  isValid: boolean
  isEmpty: boolean
  isHeader: boolean
  errorMessage: string | null
}

export interface DocumentError {
  lineNumber: number
  message: string
}

export interface ParsedDocument {
  lines: ParsedLine[]
  isValid: boolean
  errors: DocumentError[]
}

const VALID_SYMBOLS: Set<string> = new Set(['.', '-', 'o', 'x', '~', '?', '>'])

const INDENT_SIZE = 2

export function parseLine(line: string, lineNumber: number): ParsedLine {
  const result: ParsedLine = {
    lineNumber,
    raw: line,
    depth: 0,
    symbol: null,
    priority: 0,
    content: '',
    migrationTarget: null,
    isValid: true,
    isEmpty: false,
    isHeader: false,
    errorMessage: null,
  }

  if (!line || line.trim() === '') {
    result.isEmpty = true
    return result
  }

  if (line.includes('──')) {
    result.isHeader = true
    return result
  }

  const normalizedLine = line.replace(/\t/g, '  ')

  const leadingSpaces = normalizedLine.length - normalizedLine.trimStart().length
  result.depth = Math.round(leadingSpaces / INDENT_SIZE)

  const trimmed = normalizedLine.trimStart()

  if (trimmed.length === 0) {
    result.isEmpty = true
    return result
  }

  const firstChar = trimmed[0]

  if (!VALID_SYMBOLS.has(firstChar)) {
    result.isValid = false
    result.errorMessage = 'Unknown entry type'
    return result
  }

  result.symbol = firstChar as EntrySymbol

  let remainder = trimmed.slice(1).trimStart()

  if (result.symbol === '>' && remainder.startsWith('[')) {
    const closeBracket = remainder.indexOf(']')
    if (closeBracket > 0) {
      result.migrationTarget = remainder.slice(1, closeBracket)
      remainder = remainder.slice(closeBracket + 1).trimStart()
    }
  }

  const priorityMatch = remainder.match(/^(!!!|!!|!)\s*/)
  if (priorityMatch) {
    const markers = priorityMatch[1]
    if (markers === '!!!') {
      result.priority = 1
    } else if (markers === '!!') {
      result.priority = 2
    } else if (markers === '!') {
      result.priority = 3
    }
    remainder = remainder.slice(priorityMatch[0].length)
  }

  result.content = remainder

  if (!result.content || result.content.trim() === '') {
    result.isValid = false
    result.errorMessage = 'Entry content required'
  }

  return result
}

export function parseDocument(document: string): ParsedDocument {
  if (!document) {
    return {
      lines: [],
      isValid: true,
      errors: [],
    }
  }

  const rawLines = document.split('\n')
  const lines: ParsedLine[] = []
  const errors: DocumentError[] = []

  for (let i = 0; i < rawLines.length; i++) {
    const lineNumber = i + 1
    const parsed = parseLine(rawLines[i], lineNumber)
    lines.push(parsed)

    if (!parsed.isValid && parsed.errorMessage) {
      errors.push({
        lineNumber,
        message: parsed.errorMessage,
      })
    }
  }

  return {
    lines,
    isValid: errors.length === 0,
    errors,
  }
}

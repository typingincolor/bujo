import { StateField, StateEffect } from '@codemirror/state'
import { ViewPlugin, Decoration, DecorationSet, EditorView } from '@codemirror/view'
import type { DocumentError } from '../editableParser'

export interface Diagnostic {
  from: number
  to: number
  severity: 'error' | 'warning' | 'info'
  message: string
}

export function linesToDiagnostics(errors: DocumentError[], doc: string): Diagnostic[] {
  if (errors.length === 0) return []

  const lines = doc.split('\n')
  const diagnostics: Diagnostic[] = []

  for (const error of errors) {
    const lineIndex = error.lineNumber - 1
    if (lineIndex < 0 || lineIndex >= lines.length) continue

    let from = 0
    for (let i = 0; i < lineIndex; i++) {
      from += lines[i].length + 1
    }

    const to = from + lines[lineIndex].length

    diagnostics.push({
      from,
      to,
      severity: 'error',
      message: error.message
    })
  }

  return diagnostics
}

export const setErrors = StateEffect.define<DocumentError[]>()

const errorsField = StateField.define<DocumentError[]>({
  create() {
    return []
  },
  update(errors, tr) {
    for (const effect of tr.effects) {
      if (effect.is(setErrors)) {
        return effect.value
      }
    }
    return errors
  },
})

const errorLineMark = Decoration.line({ class: 'error-line' })

function buildDecorations(view: EditorView): DecorationSet {
  const errors = view.state.field(errorsField)
  if (errors.length === 0) return Decoration.none

  const decorations: { pos: number }[] = []

  for (const error of errors) {
    const lineIndex = error.lineNumber - 1
    if (lineIndex < 0 || lineIndex >= view.state.doc.lines) continue

    const line = view.state.doc.line(error.lineNumber)
    decorations.push({ pos: line.from })
  }

  return Decoration.set(
    decorations.map((d) => errorLineMark.range(d.pos))
  )
}

export function errorHighlightExtension() {
  return [
    errorsField,
    ViewPlugin.fromClass(
      class {
        decorations: DecorationSet

        constructor(view: EditorView) {
          this.decorations = buildDecorations(view)
        }

        update(update: { view: EditorView; docChanged: boolean; transactions: readonly { effects: readonly StateEffect<unknown>[] }[] }) {
          const hasErrorChange = update.transactions.some((tr) =>
            tr.effects.some((e) => e.is(setErrors))
          )
          if (update.docChanged || hasErrorChange) {
            this.decorations = buildDecorations(update.view)
          }
        }
      },
      {
        decorations: (v) => v.decorations,
      }
    ),
  ]
}

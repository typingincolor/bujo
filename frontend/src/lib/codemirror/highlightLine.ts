import { StateField, StateEffect } from '@codemirror/state'
import { ViewPlugin, Decoration, DecorationSet, EditorView } from '@codemirror/view'

export interface HighlightRange {
  from: number
  to: number
}

export const setHighlight = StateEffect.define<HighlightRange | null>()

const highlightField = StateField.define<HighlightRange | null>({
  create() {
    return null
  },
  update(highlight, tr) {
    for (const effect of tr.effects) {
      if (effect.is(setHighlight)) {
        return effect.value
      }
    }
    return highlight
  },
})

const highlightLineMark = Decoration.line({ class: 'highlight-line' })

function buildDecorations(view: EditorView): DecorationSet {
  const highlight = view.state.field(highlightField)
  if (!highlight) return Decoration.none

  const line = view.state.doc.lineAt(highlight.from)
  return Decoration.set([highlightLineMark.range(line.from)])
}

export function highlightLineExtension() {
  return [
    highlightField,
    ViewPlugin.fromClass(
      class {
        decorations: DecorationSet

        constructor(view: EditorView) {
          this.decorations = buildDecorations(view)
        }

        update(update: { view: EditorView; docChanged: boolean; transactions: readonly { effects: readonly StateEffect<unknown>[] }[] }) {
          const hasHighlightChange = update.transactions.some((tr) =>
            tr.effects.some((e) => e.is(setHighlight))
          )
          if (update.docChanged || hasHighlightChange) {
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

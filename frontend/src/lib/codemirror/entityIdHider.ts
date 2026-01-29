import { ViewPlugin, Decoration, DecorationSet, EditorView } from '@codemirror/view'
import { RangeSet, RangeValue } from '@codemirror/state'

class AtomicRange extends RangeValue {}
const atomicMark = new AtomicRange()

export interface EntityIdRange {
  start: number
  end: number
}

export function findEntityIdPrefix(line: string): EntityIdRange | null {
  const match = line.match(/^(\s*)\[[^\]\n]+\] /)
  if (!match) return null

  const start = match[1].length
  const end = match[0].length

  return { start, end }
}

const hideDecoration = Decoration.replace({})

function buildDecorations(view: EditorView): DecorationSet {
  const ranges: { from: number; to: number }[] = []

  for (let i = 1; i <= view.state.doc.lines; i++) {
    const line = view.state.doc.line(i)
    const prefix = findEntityIdPrefix(line.text)

    if (prefix) {
      ranges.push({
        from: line.from + prefix.start,
        to: line.from + prefix.end,
      })
    }
  }

  return Decoration.set(
    ranges.map((r) => hideDecoration.range(r.from, r.to))
  )
}

function buildAtomicRanges(view: EditorView): RangeSet<AtomicRange> {
  const ranges: { from: number; to: number }[] = []

  for (let i = 1; i <= view.state.doc.lines; i++) {
    const line = view.state.doc.line(i)
    const prefix = findEntityIdPrefix(line.text)

    if (prefix) {
      ranges.push({
        from: line.from + prefix.start,
        to: line.from + prefix.end,
      })
    }
  }

  return RangeSet.of(
    ranges.map((r) => atomicMark.range(r.from, r.to))
  )
}

export function entityIdAtomicRanges() {
  return EditorView.atomicRanges.of((view) => buildAtomicRanges(view))
}

export function entityIdHiderExtension() {
  return ViewPlugin.fromClass(
    class {
      decorations: DecorationSet

      constructor(view: EditorView) {
        this.decorations = buildDecorations(view)
      }

      update(update: { view: EditorView; docChanged: boolean }) {
        if (update.docChanged) {
          this.decorations = buildDecorations(update.view)
        }
      }
    },
    {
      decorations: (v) => v.decorations,
    }
  )
}

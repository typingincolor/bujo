import { ViewPlugin, Decoration, DecorationSet, EditorView } from '@codemirror/view'

export interface PriorityMarker {
  start: number
  end: number
  priority: 1 | 2 | 3
}

export function findPriorityMarkers(line: string): PriorityMarker[] {
  const match = line.match(/^(\s*)([.\-ox>])\s/)
  if (!match) return []

  const symbolEnd = match[1].length + match[2].length + 1
  const afterSymbol = line.slice(symbolEnd)

  if (afterSymbol.startsWith('!!! ')) {
    return [{ start: symbolEnd, end: symbolEnd + 3, priority: 1 }]
  }
  if (afterSymbol.startsWith('!! ')) {
    return [{ start: symbolEnd, end: symbolEnd + 2, priority: 2 }]
  }
  if (afterSymbol.startsWith('! ')) {
    return [{ start: symbolEnd, end: symbolEnd + 1, priority: 3 }]
  }

  return []
}

const priorityMarkDecorations = {
  1: Decoration.mark({ class: 'priority-badge priority-badge-1' }),
  2: Decoration.mark({ class: 'priority-badge priority-badge-2' }),
  3: Decoration.mark({ class: 'priority-badge priority-badge-3' }),
} as const

function buildDecorations(view: EditorView): DecorationSet {
  const ranges: { from: number; to: number; priority: 1 | 2 | 3 }[] = []

  for (let i = 1; i <= view.state.doc.lines; i++) {
    const line = view.state.doc.line(i)
    const markers = findPriorityMarkers(line.text)

    for (const marker of markers) {
      ranges.push({
        from: line.from + marker.start,
        to: line.from + marker.end,
        priority: marker.priority,
      })
    }
  }

  return Decoration.set(
    ranges.map((r) => priorityMarkDecorations[r.priority].range(r.from, r.to))
  )
}

export function priorityBadgeExtension() {
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

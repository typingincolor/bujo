import { ViewPlugin, Decoration, DecorationSet, EditorView, WidgetType } from '@codemirror/view'

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

class PriorityBadgeWidget extends WidgetType {
  constructor(readonly priority: 1 | 2 | 3) {
    super()
  }

  toDOM(): HTMLElement {
    const span = document.createElement('span')
    span.className = `priority-badge priority-badge-${this.priority}`
    span.textContent = String(this.priority)
    return span
  }

  eq(other: PriorityBadgeWidget): boolean {
    return this.priority === other.priority
  }
}

function buildDecorations(view: EditorView): DecorationSet {
  const decorations: { from: number; to: number; widget: PriorityBadgeWidget }[] = []

  for (let i = 1; i <= view.state.doc.lines; i++) {
    const line = view.state.doc.line(i)
    const markers = findPriorityMarkers(line.text)

    for (const marker of markers) {
      decorations.push({
        from: line.from + marker.start,
        to: line.from + marker.end,
        widget: new PriorityBadgeWidget(marker.priority),
      })
    }
  }

  return Decoration.set(
    decorations.map((d) =>
      Decoration.replace({ widget: d.widget }).range(d.from, d.to)
    )
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

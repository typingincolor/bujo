import { ViewPlugin, Decoration, DecorationSet, EditorView, WidgetType } from '@codemirror/view'

const INDENT_SIZE = 2

export function getIndentDepth(line: string): number {
  const leadingSpaces = line.match(/^(\s*)/)?.[1].length ?? 0
  return Math.floor(leadingSpaces / INDENT_SIZE)
}

class IndentGuideWidget extends WidgetType {
  constructor(readonly depth: number) {
    super()
  }

  toDOM(): HTMLElement {
    const container = document.createElement('span')
    for (let i = 1; i <= this.depth; i++) {
      const guide = document.createElement('span')
      guide.className = `indent-guide indent-guide-${i}`
      container.appendChild(guide)
    }
    return container
  }

  eq(other: IndentGuideWidget): boolean {
    return this.depth === other.depth
  }
}

function buildDecorations(view: EditorView): DecorationSet {
  const decorations: { from: number; widget: IndentGuideWidget }[] = []

  for (let i = 1; i <= view.state.doc.lines; i++) {
    const line = view.state.doc.line(i)
    const depth = getIndentDepth(line.text)

    if (depth > 0) {
      decorations.push({
        from: line.from,
        widget: new IndentGuideWidget(depth),
      })
    }
  }

  return Decoration.set(
    decorations.map((d) =>
      Decoration.widget({ widget: d.widget, side: -1 }).range(d.from)
    )
  )
}

export function indentGuidesExtension() {
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

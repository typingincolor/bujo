import { ViewPlugin, Decoration, DecorationSet, EditorView } from '@codemirror/view'

type EntryStyleType = 'task' | 'note' | 'event' | 'done' | 'migrated' | 'cancelled' | 'question'

const symbolToType: Record<string, EntryStyleType> = {
  '.': 'task',
  '-': 'note',
  'o': 'event',
  'x': 'done',
  '>': 'migrated',
  '~': 'cancelled',
  '?': 'question',
}

const lineDecorations: Record<EntryStyleType, Decoration> = {
  task: Decoration.line({ class: 'cm-entry-task' }),
  note: Decoration.line({ class: 'cm-entry-note' }),
  event: Decoration.line({ class: 'cm-entry-event' }),
  done: Decoration.line({ class: 'cm-entry-done' }),
  migrated: Decoration.line({ class: 'cm-entry-migrated' }),
  cancelled: Decoration.line({ class: 'cm-entry-cancelled' }),
  question: Decoration.line({ class: 'cm-entry-question' }),
}

function getEntryType(lineText: string): EntryStyleType | null {
  const match = lineText.match(/^\s*([.\-ox>~?])\s/)
  if (!match) return null
  return symbolToType[match[1]] ?? null
}

function buildDecorations(view: EditorView): DecorationSet {
  const decorations: { from: number; deco: Decoration }[] = []

  for (let i = 1; i <= view.state.doc.lines; i++) {
    const line = view.state.doc.line(i)
    const type = getEntryType(line.text)
    if (type) {
      decorations.push({ from: line.from, deco: lineDecorations[type] })
    }
  }

  return Decoration.set(decorations.map((d) => d.deco.range(d.from)))
}

export function entryTypeStyleExtension() {
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

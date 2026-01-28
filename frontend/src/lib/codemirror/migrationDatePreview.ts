import { StateEffect, StateField, Extension } from '@codemirror/state'
import {
  Decoration,
  DecorationSet,
  EditorView,
  ViewPlugin,
  WidgetType,
} from '@codemirror/view'

export interface MigrationDate {
  from: number
  to: number
  dateString: string
}

export interface ResolvedDateInfo {
  dateString: string
  iso: string | null
  display: string | null
  error?: string
}

const MIGRATION_PATTERN = />\[([^\]\n]+)\]/g

export function findMigrationDates(text: string): MigrationDate[] {
  const results: MigrationDate[] = []
  let match

  while ((match = MIGRATION_PATTERN.exec(text)) !== null) {
    results.push({
      from: match.index + 1,
      to: match.index + match[0].length,
      dateString: match[1],
    })
  }

  MIGRATION_PATTERN.lastIndex = 0

  return results
}

export const setResolvedDates = StateEffect.define<ResolvedDateInfo[]>()

const resolvedDatesField = StateField.define<ResolvedDateInfo[]>({
  create() {
    return []
  },
  update(dates, tr) {
    for (const effect of tr.effects) {
      if (effect.is(setResolvedDates)) {
        return effect.value
      }
    }
    return dates
  },
})

class MigrationDateWidget extends WidgetType {
  constructor(
    readonly info: ResolvedDateInfo
  ) {
    super()
  }

  toDOM(): HTMLElement {
    const span = document.createElement('span')
    span.className = 'migration-date-preview'
    if (this.info.error || !this.info.display) {
      span.classList.add('error')
      span.textContent = this.info.error || 'Invalid date'
    } else {
      span.textContent = this.info.display
    }
    return span
  }

  eq(other: MigrationDateWidget): boolean {
    return (
      this.info.dateString === other.info.dateString &&
      this.info.iso === other.info.iso &&
      this.info.display === other.info.display &&
      this.info.error === other.info.error
    )
  }
}

function buildDecorations(view: EditorView): DecorationSet {
  const resolvedDates = view.state.field(resolvedDatesField)
  if (resolvedDates.length === 0) return Decoration.none

  const migrationDates = findMigrationDates(view.state.doc.toString())
  const decorations: { pos: number; widget: MigrationDateWidget }[] = []

  for (const migration of migrationDates) {
    const resolved = resolvedDates.find(
      (r) => r.dateString === migration.dateString
    )
    if (resolved) {
      decorations.push({
        pos: migration.to,
        widget: new MigrationDateWidget(resolved),
      })
    }
  }

  return Decoration.set(
    decorations.map((d) =>
      Decoration.widget({ widget: d.widget, side: 1 }).range(d.pos)
    )
  )
}

export function migrationDatePreviewExtension(): Extension {
  return [
    resolvedDatesField,
    ViewPlugin.fromClass(
      class {
        decorations: DecorationSet

        constructor(view: EditorView) {
          this.decorations = buildDecorations(view)
        }

        update(update: {
          view: EditorView
          docChanged: boolean
          transactions: readonly { effects: readonly StateEffect<unknown>[] }[]
        }) {
          const hasResolvedChange = update.transactions.some((tr) =>
            tr.effects.some((e) => e.is(setResolvedDates))
          )
          if (update.docChanged || hasResolvedChange) {
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

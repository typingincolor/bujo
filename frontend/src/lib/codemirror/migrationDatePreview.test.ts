import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import {
  findMigrationDates,
  migrationDatePreviewExtension,
  setResolvedDates,
  ResolvedDateInfo,
} from './migrationDatePreview'

describe('findMigrationDates', () => {
  it('finds migration date pattern at start of line', () => {
    const text = '>[tomorrow] Call dentist'
    const results = findMigrationDates(text)

    expect(results).toHaveLength(1)
    expect(results[0]).toEqual({
      from: 1,
      to: 11,
      dateString: 'tomorrow',
    })
  })

  it('finds ISO date pattern', () => {
    const text = '>[2026-01-29] Submit report'
    const results = findMigrationDates(text)

    expect(results).toHaveLength(1)
    expect(results[0].dateString).toBe('2026-01-29')
  })

  it('finds natural language date pattern', () => {
    const text = '>[next monday] Review PR'
    const results = findMigrationDates(text)

    expect(results).toHaveLength(1)
    expect(results[0].dateString).toBe('next monday')
  })

  it('finds multiple migration dates in multiline text', () => {
    const text = '>[tomorrow] Task 1\n. Regular task\n>[next week] Task 2'
    const results = findMigrationDates(text)

    expect(results).toHaveLength(2)
    expect(results[0].dateString).toBe('tomorrow')
    expect(results[1].dateString).toBe('next week')
  })

  it('handles indented migration entries', () => {
    const text = '  >[tomorrow] Indented migration'
    const results = findMigrationDates(text)

    expect(results).toHaveLength(1)
    expect(results[0].dateString).toBe('tomorrow')
  })

  it('returns empty array for non-migration lines', () => {
    const text = '. Regular task\n- Note\no Event'
    const results = findMigrationDates(text)

    expect(results).toHaveLength(0)
  })

  it('does not match incomplete patterns', () => {
    const text = '>[incomplete\n> [space before bracket]'
    const results = findMigrationDates(text)

    expect(results).toHaveLength(0)
  })

  it('captures position correctly for display widget', () => {
    const text = '>[tomorrow] Call dentist'
    const results = findMigrationDates(text)

    // from is position after '>' (where '[' starts)
    // to is position after ']'
    expect(results[0].from).toBe(1)
    expect(results[0].to).toBe(11)
  })
})

describe('migrationDatePreviewExtension', () => {
  let container: HTMLElement
  let view: EditorView

  beforeEach(() => {
    container = document.createElement('div')
    document.body.appendChild(container)
  })

  afterEach(() => {
    view?.destroy()
    container?.remove()
  })

  it('creates an editor without errors', () => {
    const state = EditorState.create({
      doc: '>[tomorrow] Call dentist',
      extensions: [migrationDatePreviewExtension()],
    })

    view = new EditorView({
      state,
      parent: container,
    })

    expect(view.state.doc.toString()).toBe('>[tomorrow] Call dentist')
  })

  it('displays resolved date when setResolvedDates effect is dispatched', () => {
    const state = EditorState.create({
      doc: '>[tomorrow] Call dentist',
      extensions: [migrationDatePreviewExtension()],
    })

    view = new EditorView({
      state,
      parent: container,
    })

    const resolvedDates: ResolvedDateInfo[] = [
      { dateString: 'tomorrow', iso: '2026-01-29', display: 'Wed, Jan 29' },
    ]

    view.dispatch({
      effects: setResolvedDates.of(resolvedDates),
    })

    const widget = container.querySelector('.migration-date-preview')
    expect(widget).not.toBeNull()
    expect(widget?.textContent).toContain('Wed, Jan 29')
  })

  it('shows multiple resolved dates for multiple migration entries', () => {
    const state = EditorState.create({
      doc: '>[tomorrow] Task 1\n>[next week] Task 2',
      extensions: [migrationDatePreviewExtension()],
    })

    view = new EditorView({
      state,
      parent: container,
    })

    const resolvedDates: ResolvedDateInfo[] = [
      { dateString: 'tomorrow', iso: '2026-01-29', display: 'Wed, Jan 29' },
      { dateString: 'next week', iso: '2026-02-04', display: 'Wed, Feb 4' },
    ]

    view.dispatch({
      effects: setResolvedDates.of(resolvedDates),
    })

    const widgets = container.querySelectorAll('.migration-date-preview')
    expect(widgets).toHaveLength(2)
  })

  it('shows error indicator for invalid dates', () => {
    const state = EditorState.create({
      doc: '>[invalid-date] Task',
      extensions: [migrationDatePreviewExtension()],
    })

    view = new EditorView({
      state,
      parent: container,
    })

    const resolvedDates: ResolvedDateInfo[] = [
      { dateString: 'invalid-date', iso: null, display: null, error: 'Invalid date' },
    ]

    view.dispatch({
      effects: setResolvedDates.of(resolvedDates),
    })

    const widget = container.querySelector('.migration-date-preview')
    expect(widget).not.toBeNull()
    expect(widget?.classList.contains('error')).toBe(true)
  })

  it('clears resolved dates when empty array is dispatched', () => {
    const state = EditorState.create({
      doc: '>[tomorrow] Task',
      extensions: [migrationDatePreviewExtension()],
    })

    view = new EditorView({
      state,
      parent: container,
    })

    view.dispatch({
      effects: setResolvedDates.of([
        { dateString: 'tomorrow', iso: '2026-01-29', display: 'Wed, Jan 29' },
      ]),
    })

    expect(container.querySelector('.migration-date-preview')).not.toBeNull()

    view.dispatch({
      effects: setResolvedDates.of([]),
    })

    expect(container.querySelector('.migration-date-preview')).toBeNull()
  })
})

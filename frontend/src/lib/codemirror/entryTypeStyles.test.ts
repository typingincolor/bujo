import { describe, it, expect } from 'vitest'
import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { getEntryType, buildDecorations, entryTypeStyleExtension } from './entryTypeStyles'

describe('getEntryType', () => {
  it('returns task for . symbol', () => {
    expect(getEntryType('. Buy milk')).toBe('task')
  })

  it('returns note for - symbol', () => {
    expect(getEntryType('- A note')).toBe('note')
  })

  it('returns event for o symbol', () => {
    expect(getEntryType('o Meeting at 3pm')).toBe('event')
  })

  it('returns done for x symbol', () => {
    expect(getEntryType('x Completed task')).toBe('done')
  })

  it('returns migrated for > symbol', () => {
    expect(getEntryType('> Moved to tomorrow')).toBe('migrated')
  })

  it('returns cancelled for ~ symbol', () => {
    expect(getEntryType('~ No longer needed')).toBe('cancelled')
  })

  it('returns question for ? symbol', () => {
    expect(getEntryType('? What about this')).toBe('question')
  })

  it('returns answered for * symbol', () => {
    expect(getEntryType('* The answer is 42')).toBe('answered')
  })

  it('returns movedToList for ^ symbol', () => {
    expect(getEntryType('^ Moved to shopping list')).toBe('movedToList')
  })

  it('returns null for invalid symbol', () => {
    expect(getEntryType('# Not a valid entry')).toBeNull()
  })

  it('returns null for empty line', () => {
    expect(getEntryType('')).toBeNull()
  })

  it('handles indented entries', () => {
    expect(getEntryType('  . Nested task')).toBe('task')
    expect(getEntryType('    - Deep note')).toBe('note')
  })

  it('returns null when symbol lacks trailing space', () => {
    expect(getEntryType('.NoSpace')).toBeNull()
  })
})

describe('buildDecorations', () => {
  function createView(doc: string): EditorView {
    const state = EditorState.create({
      doc,
      extensions: [entryTypeStyleExtension()],
    })
    return new EditorView({ state })
  }

  it('creates decorations for entry lines', () => {
    const view = createView('. Task one\n- A note')
    const decos = buildDecorations(view)
    expect(decos.size).toBe(2)
    view.destroy()
  })

  it('skips empty lines', () => {
    const view = createView('. Task\n\n- Note')
    const decos = buildDecorations(view)
    expect(decos.size).toBe(2)
    view.destroy()
  })

  it('returns empty set for blank document', () => {
    const view = createView('')
    const decos = buildDecorations(view)
    expect(decos.size).toBe(0)
    view.destroy()
  })
})

describe('entryTypeStyleExtension', () => {
  function createView(doc: string): EditorView {
    const state = EditorState.create({
      doc,
      extensions: [entryTypeStyleExtension()],
    })
    return new EditorView({ state })
  }

  it('returns a valid CodeMirror extension', () => {
    const extension = entryTypeStyleExtension()
    expect(extension).toBeDefined()
  })

  it('can be used to create an EditorState', () => {
    const view = createView('. Task')
    expect(view.state.doc.toString()).toBe('. Task')
    view.destroy()
  })

  it('applies line decoration class for task', () => {
    const view = createView('. A task')
    const line = view.dom.querySelector('.cm-entry-task')
    expect(line).not.toBeNull()
    view.destroy()
  })

  it('applies line decoration class for note', () => {
    const view = createView('- A note')
    const line = view.dom.querySelector('.cm-entry-note')
    expect(line).not.toBeNull()
    view.destroy()
  })

  it('applies line decoration class for done', () => {
    const view = createView('x Done item')
    const line = view.dom.querySelector('.cm-entry-done')
    expect(line).not.toBeNull()
    view.destroy()
  })
})

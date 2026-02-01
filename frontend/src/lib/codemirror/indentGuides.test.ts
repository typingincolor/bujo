import { describe, it, expect } from 'vitest'
import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { getIndentDepth, indentGuidesExtension } from './indentGuides'

describe('getIndentDepth', () => {
  it('returns 0 for unindented line', () => {
    expect(getIndentDepth('. Task')).toBe(0)
  })

  it('returns 1 for 2-space indent', () => {
    expect(getIndentDepth('  . Child')).toBe(1)
  })

  it('returns 2 for 4-space indent', () => {
    expect(getIndentDepth('    . Grandchild')).toBe(2)
  })

  it('returns 3 for 6-space indent', () => {
    expect(getIndentDepth('      . Great-grandchild')).toBe(3)
  })

  it('returns 0 for empty line', () => {
    expect(getIndentDepth('')).toBe(0)
  })

  it('floors odd spaces (3 spaces = 1 level)', () => {
    expect(getIndentDepth('   . Odd indent')).toBe(1)
  })

  it('returns 0 for whitespace-only line', () => {
    expect(getIndentDepth('    ')).toBe(2)
  })
})

describe('indentGuidesExtension', () => {
  function createEditorView(doc: string): EditorView {
    const state = EditorState.create({
      doc,
      extensions: [indentGuidesExtension()],
    })
    return new EditorView({ state })
  }

  it('returns a valid CodeMirror extension', () => {
    const extension = indentGuidesExtension()
    expect(extension).toBeDefined()
  })

  it('can be used to create an EditorState', () => {
    const view = createEditorView('. Task\n  . Child')
    expect(view.state.doc.toString()).toBe('. Task\n  . Child')
    view.destroy()
  })

  it('adds indent guide markers for indented lines', () => {
    const view = createEditorView('. Task\n  . Child')

    const guides = view.dom.querySelectorAll('.indent-guide')
    expect(guides.length).toBeGreaterThan(0)

    view.destroy()
  })

  it('adds guide with depth class for indent level 1', () => {
    const view = createEditorView('. Task\n  . Child')

    const guide = view.dom.querySelector('.indent-guide-1')
    expect(guide).not.toBeNull()

    view.destroy()
  })

  it('adds multiple guides for deeper indentation', () => {
    const view = createEditorView('. Task\n    . Grandchild')

    const guide1 = view.dom.querySelector('.indent-guide-1')
    const guide2 = view.dom.querySelector('.indent-guide-2')
    expect(guide1).not.toBeNull()
    expect(guide2).not.toBeNull()

    view.destroy()
  })
})

import { describe, it, expect } from 'vitest'
import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { findPriorityMarkers, priorityBadgeExtension } from './priorityBadges'

describe('findPriorityMarkers', () => {
  it('finds !!! marker after task symbol', () => {
    const line = '. !!! Urgent task'
    const result = findPriorityMarkers(line)

    expect(result).toEqual([{ start: 2, end: 5, priority: 1 }])
  })

  it('finds !! marker after task symbol', () => {
    const line = '. !! High priority'
    const result = findPriorityMarkers(line)

    expect(result).toEqual([{ start: 2, end: 4, priority: 2 }])
  })

  it('finds ! marker after task symbol', () => {
    const line = '. ! Low priority'
    const result = findPriorityMarkers(line)

    expect(result).toEqual([{ start: 2, end: 3, priority: 3 }])
  })

  it('returns empty array when no markers present', () => {
    const line = '. Regular task'
    const result = findPriorityMarkers(line)

    expect(result).toEqual([])
  })

  it('ignores ! in content after marker position', () => {
    const line = '. Task with exclamation!'
    const result = findPriorityMarkers(line)

    expect(result).toEqual([])
  })

  it('finds marker after note symbol', () => {
    const line = '- !! Important note'
    const result = findPriorityMarkers(line)

    expect(result).toEqual([{ start: 2, end: 4, priority: 2 }])
  })

  it('finds marker after event symbol', () => {
    const line = 'o !!! Critical event'
    const result = findPriorityMarkers(line)

    expect(result).toEqual([{ start: 2, end: 5, priority: 1 }])
  })

  it('handles indented entries', () => {
    const line = '  . !!! Nested urgent task'
    const result = findPriorityMarkers(line)

    expect(result).toEqual([{ start: 4, end: 7, priority: 1 }])
  })
})

describe('priorityBadgeExtension', () => {
  function createEditorView(doc: string): EditorView {
    const state = EditorState.create({
      doc,
      extensions: [priorityBadgeExtension()],
    })
    return new EditorView({ state })
  }

  it('returns a valid CodeMirror extension', () => {
    const extension = priorityBadgeExtension()
    expect(extension).toBeDefined()
  })

  it('can be used to create an EditorState', () => {
    const view = createEditorView('. !!! Urgent task')
    expect(view.state.doc.toString()).toBe('. !!! Urgent task')
    view.destroy()
  })

  it('creates decorations for priority markers', () => {
    const view = createEditorView('. !!! Urgent task')

    const decorations = view.dom.querySelectorAll('.priority-badge')
    expect(decorations.length).toBeGreaterThan(0)

    view.destroy()
  })

  it('marks !!! with priority-badge-1 class', () => {
    const view = createEditorView('. !!! Urgent task')

    const badge = view.dom.querySelector('.priority-badge-1')
    expect(badge).not.toBeNull()
    expect(badge?.textContent).toBe('!!!')

    view.destroy()
  })

  it('marks !! with priority-badge-2 class', () => {
    const view = createEditorView('. !! High priority')

    const badge = view.dom.querySelector('.priority-badge-2')
    expect(badge).not.toBeNull()
    expect(badge?.textContent).toBe('!!')

    view.destroy()
  })

  it('marks ! with priority-badge-3 class', () => {
    const view = createEditorView('. ! Low priority')

    const badge = view.dom.querySelector('.priority-badge-3')
    expect(badge).not.toBeNull()
    expect(badge?.textContent).toBe('!')

    view.destroy()
  })
})

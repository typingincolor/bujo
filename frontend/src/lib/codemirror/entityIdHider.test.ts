import { describe, it, expect } from 'vitest'
import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { findEntityIdPrefix, entityIdHiderExtension, entityIdAtomicRanges } from './entityIdHider'

describe('findEntityIdPrefix', () => {
  it('finds entity ID prefix at start of line', () => {
    const result = findEntityIdPrefix('[abc-123] . Buy groceries')

    expect(result).toEqual({ start: 0, end: 10 })
  })

  it('finds entity ID prefix after indentation', () => {
    const result = findEntityIdPrefix('  [abc-123] - Child note')

    expect(result).toEqual({ start: 2, end: 12 })
  })

  it('returns null when no entity ID prefix', () => {
    const result = findEntityIdPrefix('. Regular task')

    expect(result).toBeNull()
  })

  it('finds UUID-style entity ID', () => {
    const result = findEntityIdPrefix('[019c0b3e-6382-7bdb-8a4a-853b49ce110c] . Task')

    expect(result).toEqual({ start: 0, end: 39 })
  })

  it('ignores brackets not at line start or after indentation', () => {
    const result = findEntityIdPrefix('. Task with [brackets]')

    expect(result).toBeNull()
  })
})

describe('entityIdHiderExtension', () => {
  function createEditorView(doc: string): EditorView {
    const state = EditorState.create({
      doc,
      extensions: [entityIdHiderExtension()],
    })
    return new EditorView({ state })
  }

  it('returns a valid CodeMirror extension', () => {
    const extension = entityIdHiderExtension()
    expect(extension).toBeDefined()
  })

  it('hides entity ID prefix from rendered output', () => {
    const view = createEditorView('[abc-123] . Task')

    const text = view.dom.textContent
    expect(text).not.toContain('[abc-123]')
    expect(text).toContain('. Task')

    view.destroy()
  })

  it('hides entity ID prefix on indented lines', () => {
    const view = createEditorView('  [abc-123] - Note')

    const text = view.dom.textContent
    expect(text).not.toContain('[abc-123]')
    expect(text).toContain('- Note')

    view.destroy()
  })

  it('preserves entity ID in document model', () => {
    const view = createEditorView('[abc-123] . Task')

    expect(view.state.doc.toString()).toBe('[abc-123] . Task')

    view.destroy()
  })
})

describe('entityIdAtomicRanges', () => {
  it('returns a valid CodeMirror extension', () => {
    const extension = entityIdAtomicRanges()
    expect(extension).toBeDefined()
  })

  it('provides atomic ranges covering entity ID prefixes', () => {
    const state = EditorState.create({
      doc: '[abc-123] . Task',
      extensions: [entityIdHiderExtension(), entityIdAtomicRanges()],
    })
    const view = new EditorView({ state })

    const atomicFacet = view.state.facet(EditorView.atomicRanges)
    expect(atomicFacet.length).toBeGreaterThan(0)

    const rangeFn = atomicFacet[0]
    const ranges = rangeFn(view)
    let count = 0
    const iter = ranges.iter()
    while (iter.value) {
      count++
      expect(iter.from).toBe(0)
      expect(iter.to).toBe(10)
      iter.next()
    }
    expect(count).toBe(1)

    view.destroy()
  })
})

import { describe, it, expect } from 'vitest'
import { EditorState } from '@codemirror/state'
import { CompletionContext } from '@codemirror/autocomplete'
import { tagCompletionSource, tagAutocomplete } from './tagAutocomplete'

describe('tagCompletionSource', () => {
  function createContext(doc: string, pos: number): CompletionContext {
    const state = EditorState.create({ doc })
    return new CompletionContext(state, pos, false)
  }

  it('returns null when cursor is not after a tag', () => {
    const source = tagCompletionSource(['work', 'personal'])
    const result = source(createContext('hello world', 5))
    expect(result).toBeNull()
  })

  it('returns null when only # is typed without letters', () => {
    const source = tagCompletionSource(['work', 'personal'])
    const result = source(createContext('hello #', 7))
    expect(result).toBeNull()
  })

  it('returns matching tags when # followed by text', () => {
    const source = tagCompletionSource(['work', 'weekend', 'personal'])
    const result = source(createContext('task #we', 8))
    expect(result).not.toBeNull()
    expect(result!.from).toBe(5)
    expect(result!.options).toHaveLength(1)
    expect(result!.options[0].label).toBe('#weekend')
  })

  it('returns multiple matching tags for shared prefix', () => {
    const source = tagCompletionSource(['work', 'workout', 'personal'])
    const result = source(createContext('task #wo', 8))
    expect(result).not.toBeNull()
    expect(result!.options).toHaveLength(2)
    expect(result!.options.map(o => o.label)).toContain('#work')
    expect(result!.options.map(o => o.label)).toContain('#workout')
  })

  it('returns all tags when single letter matches all', () => {
    const source = tagCompletionSource(['work', 'weekend'])
    const result = source(createContext('#w', 2))
    expect(result!.options).toHaveLength(2)
  })

  it('returns null when no tags match', () => {
    const source = tagCompletionSource(['work', 'personal'])
    const result = source(createContext('#z', 2))
    expect(result).toBeNull()
  })

  it('matches case-insensitively', () => {
    const source = tagCompletionSource(['Work', 'personal'])
    const result = source(createContext('#wo', 3))
    expect(result!.options).toHaveLength(1)
    expect(result!.options[0].label).toBe('#Work')
  })

  it('returns null for empty tag list', () => {
    const source = tagCompletionSource([])
    const result = source(createContext('#w', 2))
    expect(result).toBeNull()
  })

  it('handles tag with hyphens', () => {
    const source = tagCompletionSource(['work-life', 'work'])
    const result = source(createContext('#work-l', 7))
    expect(result!.options).toHaveLength(1)
    expect(result!.options[0].label).toBe('#work-life')
  })
})

describe('tagAutocomplete', () => {
  it('returns a valid CodeMirror extension', () => {
    const extension = tagAutocomplete(['work'])
    expect(extension).toBeDefined()
  })

  it('can be used to create an EditorState', () => {
    const state = EditorState.create({
      doc: 'test #w',
      extensions: [tagAutocomplete(['work'])],
    })
    expect(state.doc.toString()).toBe('test #w')
  })
})

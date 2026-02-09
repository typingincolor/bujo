import { describe, it, expect } from 'vitest'
import { EditorState } from '@codemirror/state'
import { foldService, foldEffect } from '@codemirror/language'
import { getFoldRange, bujoFoldExtension, expandRangeForFolds, computeFoldAllEffects } from './bujoFolding'

describe('getFoldRange', () => {
  it('returns null for entry without children', () => {
    const lines = ['. Task']
    const result = getFoldRange(lines, 0)

    expect(result).toBeNull()
  })

  it('returns range covering direct children', () => {
    const lines = [
      '. Parent',
      '  . Child 1',
      '  . Child 2',
    ]
    const result = getFoldRange(lines, 0)

    expect(result).toEqual({ from: 0, to: 2 })
  })

  it('returns range covering nested children', () => {
    const lines = [
      '. Parent',
      '  . Child',
      '    . Grandchild',
    ]
    const result = getFoldRange(lines, 0)

    expect(result).toEqual({ from: 0, to: 2 })
  })

  it('stops at sibling entry', () => {
    const lines = [
      '. Parent',
      '  . Child',
      '. Sibling',
    ]
    const result = getFoldRange(lines, 0)

    expect(result).toEqual({ from: 0, to: 1 })
  })

  it('returns null for last line', () => {
    const lines = ['. First', '. Last']
    const result = getFoldRange(lines, 1)

    expect(result).toBeNull()
  })

  it('handles middle entry with children', () => {
    const lines = [
      '. First',
      '. Middle',
      '  . Middle child',
      '. Last',
    ]
    const result = getFoldRange(lines, 1)

    expect(result).toEqual({ from: 1, to: 2 })
  })

  it('handles deeply nested structures', () => {
    const lines = [
      '. Root',
      '  . Level 1',
      '    . Level 2',
      '      . Level 3',
      '. Next root',
    ]
    const result = getFoldRange(lines, 0)

    expect(result).toEqual({ from: 0, to: 3 })
  })

  it('returns null for blank lines', () => {
    const lines = ['. Task', '', '. Another']
    const result = getFoldRange(lines, 1)

    expect(result).toBeNull()
  })
})

function createState(doc: string): EditorState {
  return EditorState.create({ doc, extensions: [bujoFoldExtension()] })
}

function getFoldAt(state: EditorState, lineNumber: number): { from: number; to: number } | null {
  const line = state.doc.line(lineNumber)
  const services = state.facet(foldService)
  for (const service of services) {
    const range = service(state, line.from, line.to)
    if (range) return range
  }
  return null
}

describe('bujoFoldService', () => {
  it('returns fold range for parent with children', () => {
    const state = createState('. Parent\n  . Child 1\n  . Child 2')
    const range = getFoldAt(state, 1)

    expect(range).not.toBeNull()
    expect(range!.from).toBe(state.doc.line(1).to)
    expect(range!.to).toBe(state.doc.line(3).to)
  })

  it('returns null for line without children', () => {
    const state = createState('. Task 1\n. Task 2')
    const range = getFoldAt(state, 1)

    expect(range).toBeNull()
  })

  it('folds nested hierarchies from mid-level parent', () => {
    const state = createState('. Root\n  . Mid\n    . Leaf\n. Next')
    const range = getFoldAt(state, 2)

    expect(range).not.toBeNull()
    expect(range!.from).toBe(state.doc.line(2).to)
    expect(range!.to).toBe(state.doc.line(3).to)
  })

  it('recomputes after document changes (cut/paste simulation)', () => {
    const state = createState('. A\n  . A1\n. B\n  . B1')

    const rangeA = getFoldAt(state, 1)
    expect(rangeA).not.toBeNull()

    const newState = state.update({
      changes: { from: 0, to: state.doc.line(2).to + 1, insert: '' },
    }).state

    const rangeBAfterCut = getFoldAt(newState, 1)
    expect(rangeBAfterCut).not.toBeNull()
    expect(rangeBAfterCut!.from).toBe(newState.doc.line(1).to)
    expect(rangeBAfterCut!.to).toBe(newState.doc.line(2).to)
  })
})

describe('expandRangeForFolds', () => {
  it('returns original range when no folds exist', () => {
    const state = createState('. Parent\n  . Child 1\n  . Child 2')
    const line1 = state.doc.line(1)
    const result = expandRangeForFolds(state, line1.from, line1.to)

    expect(result.from).toBe(line1.from)
    expect(result.to).toBe(line1.to)
  })

  it('expands range to include folded children when parent is folded', () => {
    const state = createState('. Parent\n  . Child 1\n  . Child 2')
    const foldRange = getFoldAt(state, 1)!
    const foldedState = state.update({
      effects: foldEffect.of({ from: foldRange.from, to: foldRange.to }),
    }).state

    const line1 = foldedState.doc.line(1)
    const result = expandRangeForFolds(foldedState, line1.from, line1.to)

    expect(result.from).toBe(line1.from)
    expect(result.to).toBe(foldedState.doc.line(3).to)
  })

  it('expands range when selection spans multiple folded parents', () => {
    const state = createState('. A\n  . A1\n. B\n  . B1')
    const foldA = getFoldAt(state, 1)!
    const foldB = getFoldAt(state, 3)!
    const foldedState = state.update({
      effects: [
        foldEffect.of({ from: foldA.from, to: foldA.to }),
        foldEffect.of({ from: foldB.from, to: foldB.to }),
      ],
    }).state

    const line1 = foldedState.doc.line(1)
    const line3 = foldedState.doc.line(3)
    const result = expandRangeForFolds(foldedState, line1.from, line3.to)

    expect(result.from).toBe(line1.from)
    expect(result.to).toBe(foldedState.doc.line(4).to)
  })
})

describe('computeFoldAllEffects', () => {
  it('returns fold effects for all foldable lines', () => {
    const state = createState('. Parent\n  . Child 1\n  . Child 2')
    const effects = computeFoldAllEffects(state)

    expect(effects).toHaveLength(1)
  })

  it('returns empty array when no lines are foldable', () => {
    const state = createState('. Task 1\n. Task 2\n. Task 3')
    const effects = computeFoldAllEffects(state)

    expect(effects).toHaveLength(0)
  })

  it('returns effects for multiple foldable parents', () => {
    const state = createState('. A\n  . A1\n. B\n  . B1')
    const effects = computeFoldAllEffects(state)

    expect(effects).toHaveLength(2)
  })

  it('returns effects for nested foldable lines', () => {
    const state = createState('. Root\n  . Mid\n    . Leaf')
    const effects = computeFoldAllEffects(state)

    // Both Root and Mid are foldable
    expect(effects).toHaveLength(2)
  })

  it('effects can be applied to fold the document', () => {
    const state = createState('. Parent\n  . Child 1\n  . Child 2')
    const effects = computeFoldAllEffects(state)

    const foldedState = state.update({ effects }).state

    // After folding, the fold range should exist
    const expandedRange = expandRangeForFolds(foldedState, 0, state.doc.line(1).to)
    expect(expandedRange.to).toBe(state.doc.line(3).to)
  })
})
import { describe, it, expect } from 'vitest'
import { getFoldRange } from './bujoFolding'

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

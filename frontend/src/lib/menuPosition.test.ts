import { describe, it, expect } from 'vitest'
import { calculateMenuPosition } from './menuPosition'

describe('calculateMenuPosition', () => {
  const menuWidth = 150
  const menuHeight = 200
  const viewportWidth = 1000
  const viewportHeight = 800

  it('positions menu at click coordinates when there is enough space', () => {
    const result = calculateMenuPosition(
      { x: 100, y: 100 },
      { width: menuWidth, height: menuHeight },
      { width: viewportWidth, height: viewportHeight }
    )

    expect(result).toEqual({ x: 100, y: 100 })
  })

  it('flips menu upward when not enough space below', () => {
    const result = calculateMenuPosition(
      { x: 100, y: 700 },
      { width: menuWidth, height: menuHeight },
      { width: viewportWidth, height: viewportHeight }
    )

    expect(result.y).toBe(700 - menuHeight)
  })

  it('shifts menu left when not enough space on right', () => {
    const result = calculateMenuPosition(
      { x: 900, y: 100 },
      { width: menuWidth, height: menuHeight },
      { width: viewportWidth, height: viewportHeight }
    )

    expect(result.x).toBe(viewportWidth - menuWidth)
  })

  it('handles corner case when menu would overflow both right and bottom', () => {
    const result = calculateMenuPosition(
      { x: 900, y: 700 },
      { width: menuWidth, height: menuHeight },
      { width: viewportWidth, height: viewportHeight }
    )

    expect(result.x).toBe(viewportWidth - menuWidth)
    expect(result.y).toBe(700 - menuHeight)
  })

  it('clamps to viewport edges if menu is larger than available space', () => {
    const result = calculateMenuPosition(
      { x: 50, y: 50 },
      { width: menuWidth, height: menuHeight },
      { width: viewportWidth, height: viewportHeight }
    )

    expect(result.x).toBeGreaterThanOrEqual(0)
    expect(result.y).toBeGreaterThanOrEqual(0)
  })
})

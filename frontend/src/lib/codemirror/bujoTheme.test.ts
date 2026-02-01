import { describe, it, expect } from 'vitest'
import { bujoTheme } from './bujoTheme'

describe('bujoTheme', () => {
  it('exports a valid CodeMirror extension', () => {
    expect(bujoTheme).toBeDefined()
    expect(Array.isArray(bujoTheme) || typeof bujoTheme === 'object').toBe(true)
  })
})

import { describe, it, expect } from 'vitest'
import { bujoTheme, bujoThemeStyles } from './bujoTheme'

describe('bujoTheme', () => {
  it('exports a valid CodeMirror extension', () => {
    expect(bujoTheme).toBeDefined()
    expect(Array.isArray(bujoTheme) || typeof bujoTheme === 'object').toBe(true)
  })

  it('search buttons use primary styling for readability', () => {
    const buttonStyles = bujoThemeStyles['& .cm-search button']
    expect(buttonStyles.backgroundColor).toBe('hsl(var(--primary))')
    expect(buttonStyles.color).toBe('hsl(var(--primary-foreground))')
  })

  it('search buttons have adequate padding for readability', () => {
    const buttonStyles = bujoThemeStyles['& .cm-search button']
    expect(buttonStyles.padding).toBe('4px 12px')
  })
})

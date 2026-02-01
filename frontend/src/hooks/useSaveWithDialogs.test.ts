import { describe, it, expect } from 'vitest'
import { scanForSpecialEntries, scanForNewSpecialEntries } from './useSaveWithDialogs'

describe('scanForSpecialEntries', () => {
  it('detects migrated entries', () => {
    const doc = '. Buy milk\n> Call dentist\n- A note'
    const result = scanForSpecialEntries(doc)
    expect(result.migratedEntries).toEqual(['Call dentist'])
    expect(result.movedToListEntries).toEqual([])
  })

  it('detects moved-to-list entries', () => {
    const doc = '. Buy milk\n^ Fix bike\n- A note'
    const result = scanForSpecialEntries(doc)
    expect(result.migratedEntries).toEqual([])
    expect(result.movedToListEntries).toEqual(['Fix bike'])
  })

  it('detects both types', () => {
    const doc = '> Call dentist\n^ Fix bike'
    const result = scanForSpecialEntries(doc)
    expect(result.migratedEntries).toEqual(['Call dentist'])
    expect(result.movedToListEntries).toEqual(['Fix bike'])
  })

  it('handles indented entries', () => {
    const doc = '  > Indented migrate\n  ^ Indented list'
    const result = scanForSpecialEntries(doc)
    expect(result.migratedEntries).toEqual(['Indented migrate'])
    expect(result.movedToListEntries).toEqual(['Indented list'])
  })

  it('returns empty when no special entries', () => {
    const doc = '. Task\n- Note\no Event'
    const result = scanForSpecialEntries(doc)
    expect(result.migratedEntries).toEqual([])
    expect(result.movedToListEntries).toEqual([])
  })

  it('handles priority markers in special entries', () => {
    const doc = '> !!! Important migrated task\n^ !! Medium list task'
    const result = scanForSpecialEntries(doc)
    expect(result.migratedEntries).toEqual(['!!! Important migrated task'])
    expect(result.movedToListEntries).toEqual(['!! Medium list task'])
  })

  it('returns empty for empty document', () => {
    const result = scanForSpecialEntries('')
    expect(result.migratedEntries).toEqual([])
    expect(result.movedToListEntries).toEqual([])
  })

  it('returns hasSpecialEntries flag', () => {
    const withSpecial = scanForSpecialEntries('> Migrate this')
    expect(withSpecial.hasSpecialEntries).toBe(true)

    const withoutSpecial = scanForSpecialEntries('. Normal task')
    expect(withoutSpecial.hasSpecialEntries).toBe(false)
  })

  it('does not count children of special entries as special', () => {
    const doc = '> Migrate parent\n  . Child task\n  - Child note'
    const result = scanForSpecialEntries(doc)
    expect(result.migratedEntries).toEqual(['Migrate parent'])
    expect(result.movedToListEntries).toEqual([])
  })

  it('handles multiple migrated entries', () => {
    const doc = '> First migrate\n. Normal task\n> Second migrate\n> Third migrate'
    const result = scanForSpecialEntries(doc)
    expect(result.migratedEntries).toEqual(['First migrate', 'Second migrate', 'Third migrate'])
    expect(result.movedToListEntries).toEqual([])
  })

  it('handles multiple moved-to-list entries', () => {
    const doc = '^ First move\n- Note\n^ Second move'
    const result = scanForSpecialEntries(doc)
    expect(result.movedToListEntries).toEqual(['First move', 'Second move'])
    expect(result.migratedEntries).toEqual([])
  })

  it('ignores lines with > or ^ not at start position', () => {
    const doc = '. Task with > arrow\n- Note about ^ caret'
    const result = scanForSpecialEntries(doc)
    expect(result.hasSpecialEntries).toBe(false)
  })

  it('ignores blank lines between entries', () => {
    const doc = '> Migrate this\n\n\n^ Move this'
    const result = scanForSpecialEntries(doc)
    expect(result.migratedEntries).toEqual(['Migrate this'])
    expect(result.movedToListEntries).toEqual(['Move this'])
  })

  it('handles whitespace-only document', () => {
    const result = scanForSpecialEntries('   \n  \n')
    expect(result.hasSpecialEntries).toBe(false)
  })
})

describe('scanForNewSpecialEntries', () => {
  it('returns no special entries when document unchanged', () => {
    const doc = '> Already migrated\n. Task'
    const result = scanForNewSpecialEntries(doc, doc)
    expect(result.hasSpecialEntries).toBe(false)
    expect(result.migratedEntries).toEqual([])
  })

  it('detects new migrated entry added to document with existing ones', () => {
    const original = '> Already migrated\n. Task'
    const current = '> Already migrated\n. Task\n> New migrate'
    const result = scanForNewSpecialEntries(current, original)
    expect(result.migratedEntries).toEqual(['New migrate'])
    expect(result.hasSpecialEntries).toBe(true)
  })

  it('detects new moved-to-list entry', () => {
    const original = '. Task'
    const current = '. Task\n^ Move this'
    const result = scanForNewSpecialEntries(current, original)
    expect(result.movedToListEntries).toEqual(['Move this'])
    expect(result.hasSpecialEntries).toBe(true)
  })

  it('ignores existing migrated entries after save-reload', () => {
    const original = '> Call dentist\n> Book flight\n. Buy milk'
    const current = '> Call dentist\n> Book flight\n. Buy milk\n- New note'
    const result = scanForNewSpecialEntries(current, original)
    expect(result.hasSpecialEntries).toBe(false)
  })

  it('detects new entries when original already has some', () => {
    const original = '> Existing migrate\n^ Existing move'
    const current = '> Existing migrate\n^ Existing move\n> New migrate\n^ New move'
    const result = scanForNewSpecialEntries(current, original)
    expect(result.migratedEntries).toEqual(['New migrate'])
    expect(result.movedToListEntries).toEqual(['New move'])
  })

  it('handles duplicate content correctly', () => {
    const original = '> Same task'
    const current = '> Same task\n> Same task'
    const result = scanForNewSpecialEntries(current, original)
    expect(result.migratedEntries).toEqual(['Same task'])
    expect(result.hasSpecialEntries).toBe(true)
  })

  it('treats empty original as all entries being new', () => {
    const current = '> Migrate this\n^ Move this'
    const result = scanForNewSpecialEntries(current, '')
    expect(result.migratedEntries).toEqual(['Migrate this'])
    expect(result.movedToListEntries).toEqual(['Move this'])
  })
})

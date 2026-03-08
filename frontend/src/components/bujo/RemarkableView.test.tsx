import { describe, it, expect } from 'vitest'
import { isFolder, isImportable, formatDate, sortDocuments } from './remarkableUtils'
import { remarkable } from '../../wailsjs/go/models'

function makeDoc(overrides: Partial<remarkable.Document> = {}): remarkable.Document {
  return {
    ID: 'doc-1',
    Hash: '',
    VisibleName: 'Test',
    LastModified: '1700000000000',
    Parent: '',
    FileType: 'notebook',
    ...overrides,
  } as remarkable.Document
}

describe('isFolder', () => {
  it('returns true for empty FileType', () => {
    expect(isFolder(makeDoc({ FileType: '' }))).toBe(true)
  })

  it('returns false for notebook', () => {
    expect(isFolder(makeDoc({ FileType: 'notebook' }))).toBe(false)
  })

  it('returns false for pdf', () => {
    expect(isFolder(makeDoc({ FileType: 'pdf' }))).toBe(false)
  })
})

describe('isImportable', () => {
  it('returns true for notebook', () => {
    expect(isImportable(makeDoc({ FileType: 'notebook' }))).toBe(true)
  })

  it('returns true for pdf', () => {
    expect(isImportable(makeDoc({ FileType: 'pdf' }))).toBe(true)
  })

  it('returns false for epub', () => {
    expect(isImportable(makeDoc({ FileType: 'epub' }))).toBe(false)
  })

  it('returns false for folders', () => {
    expect(isImportable(makeDoc({ FileType: '' }))).toBe(false)
  })
})

describe('formatDate', () => {
  it('formats millisecond timestamp as readable date', () => {
    const result = formatDate('1700000000000')
    expect(result).toMatch(/Nov/)
    expect(result).toMatch(/2023/)
  })

  it('returns non-numeric strings as-is', () => {
    expect(formatDate('not-a-number')).toBe('not-a-number')
  })
})

describe('sortDocuments', () => {
  const folder1 = makeDoc({ ID: 'f1', VisibleName: 'Zebra Folder', FileType: '', Parent: '' })
  const folder2 = makeDoc({ ID: 'f2', VisibleName: 'Alpha Folder', FileType: '', Parent: '' })
  const file1 = makeDoc({ ID: 'd1', VisibleName: 'Zebra Notes', FileType: 'notebook', Parent: '' })
  const file2 = makeDoc({ ID: 'd2', VisibleName: 'Alpha Notes', FileType: 'notebook', Parent: '' })
  const nested = makeDoc({ ID: 'n1', VisibleName: 'Nested', FileType: 'pdf', Parent: 'f1' })

  const allDocs = [file1, folder1, nested, file2, folder2]

  it('filters by parent folder id', () => {
    const result = sortDocuments(allDocs, '')
    expect(result.map(d => d.ID)).not.toContain('n1')
  })

  it('places folders before files', () => {
    const result = sortDocuments(allDocs, '')
    expect(result[0].FileType).toBe('')
    expect(result[1].FileType).toBe('')
    expect(result[2].FileType).toBe('notebook')
  })

  it('sorts folders alphabetically', () => {
    const result = sortDocuments(allDocs, '')
    expect(result[0].VisibleName).toBe('Alpha Folder')
    expect(result[1].VisibleName).toBe('Zebra Folder')
  })

  it('sorts files alphabetically', () => {
    const result = sortDocuments(allDocs, '')
    expect(result[2].VisibleName).toBe('Alpha Notes')
    expect(result[3].VisibleName).toBe('Zebra Notes')
  })

  it('returns nested documents when filtering by parent', () => {
    const result = sortDocuments(allDocs, 'f1')
    expect(result).toHaveLength(1)
    expect(result[0].ID).toBe('n1')
  })

  it('returns empty array for folder with no children', () => {
    const result = sortDocuments(allDocs, 'nonexistent')
    expect(result).toHaveLength(0)
  })
})

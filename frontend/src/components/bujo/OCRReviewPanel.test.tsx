import { describe, it, expect, vi } from 'vitest'
import { render } from '@testing-library/react'
import { OCRReviewPanel } from './OCRReviewPanel'

vi.mock('../../wailsjs/go/wails/App', () => ({
  ImportEntries: vi.fn(),
}))

vi.mock('../../wailsjs/go/models', () => ({
  wails: {
    ImportedPage: class {},
  },
  remarkable: {
    OCRResult: class {},
  },
}))

const makePage = (overrides: Record<string, unknown> = {}) => ({
  pageID: 'p1',
  png: 'base64png',
  ocrResults: [],
  text: '. line one\n. line two\n. line three',
  lowConfidenceCount: 1,
  lowConfidenceLines: [1],
  ...overrides,
})

describe('OCRReviewPanel', () => {
  it('renders amber dots for low-confidence lines', () => {
    const { container } = render(
      <OCRReviewPanel
        pages={[makePage() as any]}
        documentName="Test"
        onDone={() => {}}
        onBack={() => {}}
      />
    )
    const dots = container.querySelectorAll('.bg-amber-500')
    expect(dots).toHaveLength(1)
  })

  it('renders no dots when all lines are high confidence', () => {
    const { container } = render(
      <OCRReviewPanel
        pages={[makePage({ lowConfidenceLines: [], lowConfidenceCount: 0 }) as any]}
        documentName="Test"
        onDone={() => {}}
        onBack={() => {}}
      />
    )
    const dots = container.querySelectorAll('.bg-amber-500')
    expect(dots).toHaveLength(0)
  })

  it('renders multiple dots for multiple low-confidence lines', () => {
    const { container } = render(
      <OCRReviewPanel
        pages={[makePage({ lowConfidenceLines: [0, 2], lowConfidenceCount: 2 }) as any]}
        documentName="Test"
        onDone={() => {}}
        onBack={() => {}}
      />
    )
    const dots = container.querySelectorAll('.bg-amber-500')
    expect(dots).toHaveLength(2)
  })
})

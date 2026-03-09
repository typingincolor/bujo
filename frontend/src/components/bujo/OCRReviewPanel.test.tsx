import { describe, it, expect, vi } from 'vitest'
import { render } from '@testing-library/react'
import { OCRReviewPanel } from './OCRReviewPanel'
import { wails } from '../../wailsjs/go/models'

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

const makePage = (overrides: Partial<wails.ImportedPage> = {}): wails.ImportedPage => ({
  pageID: 'p1',
  png: 'base64png',
  ocrResults: [],
  text: '. line one\n. line two\n. line three',
  lowConfidenceCount: 1,
  lowConfidenceLines: [1],
  uncertainLines: [],
  error: '',
  ...overrides,
} as wails.ImportedPage)

describe('OCRReviewPanel', () => {
  it('renders confidence bars for low-confidence lines', () => {
    const { container } = render(
      <OCRReviewPanel
        pages={[makePage()]}
        documentName="Test"
        onDone={() => {}}
        onBack={() => {}}
      />
    )
    const bars = container.querySelectorAll('.bg-amber-500')
    expect(bars).toHaveLength(1)
    const bar = bars[0] as HTMLElement
    expect(bar.className).toContain('w-1')
    expect(bar.className).toContain('h-full')
    expect(bar.className).toContain('rounded-sm')
  })

  it('renders no bars when all lines are high confidence', () => {
    const { container } = render(
      <OCRReviewPanel
        pages={[makePage({ lowConfidenceLines: [], lowConfidenceCount: 0 })]}
        documentName="Test"
        onDone={() => {}}
        onBack={() => {}}
      />
    )
    const bars = container.querySelectorAll('.bg-amber-500')
    expect(bars).toHaveLength(0)
  })

  it('renders multiple bars for multiple low-confidence lines', () => {
    const { container } = render(
      <OCRReviewPanel
        pages={[makePage({ lowConfidenceLines: [0, 2], lowConfidenceCount: 2 })]}
        documentName="Test"
        onDone={() => {}}
        onBack={() => {}}
      />
    )
    const bars = container.querySelectorAll('.bg-amber-500')
    expect(bars).toHaveLength(2)
  })

  it('renders blue bars for uncertain lines', () => {
    const { container } = render(
      <OCRReviewPanel
        pages={[makePage({ uncertainLines: [0, 2], lowConfidenceLines: [], lowConfidenceCount: 0 })]}
        documentName="Test"
        onDone={() => {}}
        onBack={() => {}}
      />
    )
    const bars = container.querySelectorAll('.bg-blue-500')
    expect(bars).toHaveLength(2)
  })

  it('renders both amber and blue bars on lines that are both low-confidence and uncertain', () => {
    const { container } = render(
      <OCRReviewPanel
        pages={[makePage({ lowConfidenceLines: [1], lowConfidenceCount: 1, uncertainLines: [1] })]}
        documentName="Test"
        onDone={() => {}}
        onBack={() => {}}
      />
    )
    const amberBars = container.querySelectorAll('.bg-amber-500')
    const blueBars = container.querySelectorAll('.bg-blue-500')
    expect(amberBars).toHaveLength(1)
    expect(blueBars).toHaveLength(1)
  })

  it('renders confidence warning as a sibling below the scrollable editor, not inside it', () => {
    const { container } = render(
      <OCRReviewPanel
        pages={[makePage()]}
        documentName="Test"
        onDone={() => {}}
        onBack={() => {}}
      />
    )
    const warning = container.querySelector('p.text-amber-500')
    expect(warning).toBeInTheDocument()
    const warningParent = warning?.parentElement
    expect(warningParent?.className).toContain('flex-col')
    const scrollableEditor = warningParent?.querySelector('.flex-1.min-h-0.overflow-auto')
    expect(scrollableEditor).toBeTruthy()
    expect(scrollableEditor?.contains(warning!)).toBe(false)
  })
})

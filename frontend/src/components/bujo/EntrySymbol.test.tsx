import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { EntrySymbol } from './EntrySymbol'

describe('EntrySymbol', () => {
  describe('basic entry types with unicode symbols', () => {
    it('renders task symbol as bullet point', () => {
      render(<EntrySymbol type="task" />)
      expect(screen.getByText('•')).toBeInTheDocument()
    })

    it('renders note symbol as en dash', () => {
      render(<EntrySymbol type="note" />)
      expect(screen.getByText('–')).toBeInTheDocument()
    })

    it('renders event symbol as circle', () => {
      render(<EntrySymbol type="event" />)
      expect(screen.getByText('⚬')).toBeInTheDocument()
    })

    it('renders done symbol as check mark', () => {
      render(<EntrySymbol type="done" />)
      expect(screen.getByText('✓')).toBeInTheDocument()
    })

    it('renders migrated symbol as arrow', () => {
      render(<EntrySymbol type="migrated" />)
      expect(screen.getByText('→')).toBeInTheDocument()
    })

    it('renders cancelled symbol as ballot X', () => {
      render(<EntrySymbol type="cancelled" />)
      expect(screen.getByText('✗')).toBeInTheDocument()
    })
  })

  describe('question/answer entry types', () => {
    it('renders question symbol', () => {
      render(<EntrySymbol type="question" />)
      expect(screen.getByText('?')).toBeInTheDocument()
    })

    it('renders answered symbol', () => {
      render(<EntrySymbol type="answered" />)
      expect(screen.getByText('★')).toBeInTheDocument()
    })

    it('renders answer symbol', () => {
      render(<EntrySymbol type="answer" />)
      expect(screen.getByText('↳')).toBeInTheDocument()
    })
  })

  describe('alignment', () => {
    it('has fixed width for consistent alignment', () => {
      const { container } = render(<EntrySymbol type="task" />)
      const symbolElement = container.querySelector('span span')
      expect(symbolElement).toHaveClass('w-5')
    })
  })

  describe('priority symbols', () => {
    it('shows low priority symbol', () => {
      render(<EntrySymbol type="task" priority="low" />)
      expect(screen.getByText('!')).toBeInTheDocument()
    })

    it('shows medium priority symbol', () => {
      render(<EntrySymbol type="task" priority="medium" />)
      expect(screen.getByText('!!')).toBeInTheDocument()
    })

    it('shows high priority symbol', () => {
      render(<EntrySymbol type="task" priority="high" />)
      expect(screen.getByText('!!!')).toBeInTheDocument()
    })

    it('hides priority when none', () => {
      render(<EntrySymbol type="task" priority="none" />)
      expect(screen.queryByText('!')).not.toBeInTheDocument()
    })
  })
})

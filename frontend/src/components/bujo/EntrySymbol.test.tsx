import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { EntrySymbol } from './EntrySymbol'

describe('EntrySymbol', () => {
  describe('basic entry types', () => {
    it('renders task symbol', () => {
      render(<EntrySymbol type="task" />)
      expect(screen.getByText('.')).toBeInTheDocument()
    })

    it('renders note symbol', () => {
      render(<EntrySymbol type="note" />)
      expect(screen.getByText('-')).toBeInTheDocument()
    })

    it('renders event symbol', () => {
      render(<EntrySymbol type="event" />)
      expect(screen.getByText('o')).toBeInTheDocument()
    })

    it('renders done symbol', () => {
      render(<EntrySymbol type="done" />)
      expect(screen.getByText('x')).toBeInTheDocument()
    })

    it('renders migrated symbol', () => {
      render(<EntrySymbol type="migrated" />)
      expect(screen.getByText('>')).toBeInTheDocument()
    })

    it('renders cancelled symbol', () => {
      render(<EntrySymbol type="cancelled" />)
      expect(screen.getByText('X')).toBeInTheDocument()
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

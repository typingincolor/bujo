import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CaptureBar } from './CaptureBar'
import { Entry } from '@/types/bujo'

describe('CaptureBar', () => {
  const defaultProps = {
    onSubmit: vi.fn(),
    onSubmitChild: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  describe('rendering', () => {
    it('does NOT render type selector buttons (prefix characters are sufficient)', () => {
      render(<CaptureBar {...defaultProps} />)

      // Type buttons should NOT exist - design decision: prefix characters are sufficient
      expect(screen.queryByRole('button', { name: /task/i })).not.toBeInTheDocument()
      expect(screen.queryByRole('button', { name: /note/i })).not.toBeInTheDocument()
      expect(screen.queryByRole('button', { name: /event/i })).not.toBeInTheDocument()
      expect(screen.queryByRole('button', { name: /question/i })).not.toBeInTheDocument()
    })

    it('renders input with placeholder', () => {
      render(<CaptureBar {...defaultProps} />)

      expect(screen.getByPlaceholderText(/capture a thought/i)).toBeInTheDocument()
    })

  })

  describe('type selection via prefix only', () => {
    it('Tab blurs the input (no type cycling)', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.click(input)
      await user.keyboard('{Tab}')

      // Tab should blur the input, not cycle type
      expect(input).not.toHaveFocus()
    })
  })

  describe('prefix detection', () => {
    it('typing ". " keeps prefix in input', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, '. ')

      expect(input).toHaveValue('. ') // Prefix kept in input
    })

    it('typing "- " keeps prefix in input', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, '- ')

      expect(input).toHaveValue('- ') // Prefix kept in input
    })

    it('typing "o " keeps prefix in input', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, 'o ')

      expect(input).toHaveValue('o ') // Prefix kept in input
    })

    it('typing "? " keeps prefix in input', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, '? ')

      expect(input).toHaveValue('? ') // Prefix kept in input
    })
  })

  describe('prefix detection edge cases', () => {
    it('keeps full content when pasting dash-prefixed text', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)
      const textarea = screen.getByTestId('capture-bar-input')

      await user.click(textarea)
      await user.paste('- hello')

      expect(textarea).toHaveValue('- hello') // Full content including prefix
    })

    it('keeps full content when pasting task-prefixed text', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)
      const textarea = screen.getByTestId('capture-bar-input')

      await user.click(textarea)
      await user.paste('. buy milk')

      expect(textarea).toHaveValue('. buy milk') // Full content including prefix
    })

    it('preserves content when prefix appears mid-content', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)
      const textarea = screen.getByTestId('capture-bar-input')

      await user.type(textarea, 'buy - groceries')

      expect(textarea).toHaveValue('buy - groceries')
    })
  })

  describe('submission', () => {
    it('Enter submits content exactly as typed', async () => {
      const user = userEvent.setup()
      const onSubmit = vi.fn()
      render(<CaptureBar {...defaultProps} onSubmit={onSubmit} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, '. Buy groceries{Enter}')

      expect(onSubmit).toHaveBeenCalledWith('. Buy groceries')
    })

    it('clears input after submission', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, '. Buy groceries{Enter}')

      expect(input).toHaveValue('')
    })

    it('does not submit empty input', async () => {
      const user = userEvent.setup()
      const onSubmit = vi.fn()
      render(<CaptureBar {...defaultProps} onSubmit={onSubmit} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.click(input)
      await user.keyboard('{Enter}')

      expect(onSubmit).not.toHaveBeenCalled()
    })

    it('keeps focus after submission', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, '. Buy groceries{Enter}')

      expect(input).toHaveFocus()
    })
  })

  describe('parent context', () => {
    it('shows parent context when parentEntry provided', () => {
      render(
        <CaptureBar
          {...defaultProps}
          parentEntry={{ id: 1, content: 'Team standup', type: 'event', priority: 'none', parentId: null, loggedDate: '' } as Entry}
        />
      )

      expect(screen.getByText(/adding to:/i)).toBeInTheDocument()
      expect(screen.getByText('Team standup')).toBeInTheDocument()
    })

    it('shows clear button when parent is set', () => {
      render(
        <CaptureBar
          {...defaultProps}
          parentEntry={{ id: 1, content: 'Team standup', type: 'event', priority: 'none', parentId: null, loggedDate: '' } as Entry}
        />
      )

      expect(screen.getByRole('button', { name: /clear parent/i })).toBeInTheDocument()
    })

    it('calls onSubmitChild when parent is set', async () => {
      const user = userEvent.setup()
      const onSubmitChild = vi.fn()
      render(
        <CaptureBar
          {...defaultProps}
          onSubmitChild={onSubmitChild}
          parentEntry={{ id: 1, content: 'Team standup', type: 'event', priority: 'none', parentId: null, loggedDate: '' } as Entry}
        />
      )

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, '. Action item{Enter}')

      expect(onSubmitChild).toHaveBeenCalledWith(1, '. Action item')
    })

    it('calls onClearParent when X clicked', async () => {
      const user = userEvent.setup()
      const onClearParent = vi.fn()
      render(
        <CaptureBar
          {...defaultProps}
          parentEntry={{ id: 1, content: 'Team standup', type: 'event', priority: 'none', parentId: null, loggedDate: '' } as Entry}
          onClearParent={onClearParent}
        />
      )

      await user.click(screen.getByRole('button', { name: /clear parent/i }))

      expect(onClearParent).toHaveBeenCalled()
    })
  })

  describe('draft persistence', () => {
    it('saves draft to localStorage', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, 'Draft text')

      expect(localStorage.getItem('bujo-capture-bar-draft')).toBe('Draft text')
    })

    it('restores draft on mount', () => {
      localStorage.setItem('bujo-capture-bar-draft', 'Restored draft')

      render(<CaptureBar {...defaultProps} />)

      expect(screen.getByDisplayValue('Restored draft')).toBeInTheDocument()
    })

    it('clears draft after submission', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, '. Draft text{Enter}')

      expect(localStorage.getItem('bujo-capture-bar-draft')).toBeNull()
    })
  })

  describe('escape handling', () => {
    it('Escape clears input', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, 'Some text')
      await user.keyboard('{Escape}')

      expect(input).toHaveValue('')
    })

    it('Escape on empty input blurs', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.click(input)
      await user.keyboard('{Escape}')

      expect(input).not.toHaveFocus()
    })
  })

  describe('multiline', () => {
    it('Shift+Enter adds newline', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, 'Line 1{Shift>}{Enter}{/Shift}Line 2')

      expect(input).toHaveValue('Line 1\nLine 2')
    })
  })

  describe('textarea auto-grow', () => {
    it('expands textarea height for multiline content', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)
      const textarea = screen.getByTestId('capture-bar-input') as HTMLTextAreaElement

      await user.type(textarea, 'Line 1{Shift>}{Enter}{/Shift}Line 2{Shift>}{Enter}{/Shift}Line 3')

      expect(textarea.style.height).not.toBe('')
      expect(textarea.style.height).not.toBe('auto')
    })
  })

  describe('prefix kept in content', () => {
    it('keeps prefix visible in input field when typing prefix', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, '. Buy groceries')

      // The prefix should remain in the input
      expect(input).toHaveValue('. Buy groceries')
    })

    it('submits content exactly as typed including prefix', async () => {
      const user = userEvent.setup()
      const onSubmit = vi.fn()
      render(<CaptureBar {...defaultProps} onSubmit={onSubmit} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, '. Test task')
      await user.keyboard('{Enter}')

      // Should submit exactly what was typed
      expect(onSubmit).toHaveBeenCalledWith('. Test task')
    })

    it('does not double-prepend prefix when content already has prefix', async () => {
      const user = userEvent.setup()
      const onSubmit = vi.fn()
      render(<CaptureBar {...defaultProps} onSubmit={onSubmit} />)

      const input = screen.getByPlaceholderText(/capture a thought/i)
      await user.type(input, '. Buy groceries')
      await user.keyboard('{Enter}')

      // Should NOT submit ". . Buy groceries" - just the typed content
      expect(onSubmit).toHaveBeenCalledWith('. Buy groceries')
    })
  })

  describe('fixed positioning', () => {
    it('has fixed bottom positioning classes', () => {
      render(<CaptureBar {...defaultProps} />)
      const container = screen.getByTestId('capture-bar')
      expect(container).toHaveClass('fixed', 'bottom-3')
    })
  })
})

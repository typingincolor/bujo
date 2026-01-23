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
    it('renders type buttons', () => {
      render(<CaptureBar {...defaultProps} />)

      expect(screen.getByRole('button', { name: /task/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /note/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /event/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /question/i })).toBeInTheDocument()
    })

    it('displays bullet journal symbols instead of word labels', () => {
      render(<CaptureBar {...defaultProps} />)

      // Buttons should show symbols, not words
      expect(screen.getByRole('button', { name: /task/i })).toHaveTextContent('.')
      expect(screen.getByRole('button', { name: /note/i })).toHaveTextContent('-')
      expect(screen.getByRole('button', { name: /event/i })).toHaveTextContent('o')
      expect(screen.getByRole('button', { name: /question/i })).toHaveTextContent('?')
    })

    it('renders input with placeholder', () => {
      render(<CaptureBar {...defaultProps} />)

      expect(screen.getByPlaceholderText(/add a task/i)).toBeInTheDocument()
    })

    it('renders file import button', () => {
      render(<CaptureBar {...defaultProps} />)

      expect(screen.getByRole('button', { name: /import file/i })).toBeInTheDocument()
    })

    it('has task selected by default', () => {
      render(<CaptureBar {...defaultProps} />)

      const taskButton = screen.getByRole('button', { name: /task/i })
      expect(taskButton).toHaveAttribute('aria-pressed', 'true')
    })
  })

  describe('type selection', () => {
    it('clicking type button changes selection', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const noteButton = screen.getByRole('button', { name: /note/i })
      await user.click(noteButton)

      expect(noteButton).toHaveAttribute('aria-pressed', 'true')
      expect(screen.getByRole('button', { name: /task/i })).toHaveAttribute('aria-pressed', 'false')
    })

    it('updates placeholder when type changes', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /note/i }))

      expect(screen.getByPlaceholderText(/add a note/i)).toBeInTheDocument()
    })

    it('Tab cycles types when input is empty', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.click(input)
      await user.keyboard('{Tab}')

      expect(screen.getByRole('button', { name: /note/i })).toHaveAttribute('aria-pressed', 'true')
    })

    it('Tab does not cycle when input has content', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Some content')
      await user.keyboard('{Tab}')

      // Task should still be selected
      expect(screen.getByRole('button', { name: /task/i })).toHaveAttribute('aria-pressed', 'true')
    })
  })

  describe('prefix detection', () => {
    it('typing ". " sets type to task', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      // Start with note selected
      await user.click(screen.getByRole('button', { name: /note/i }))

      const input = screen.getByPlaceholderText(/add a note/i)
      await user.type(input, '. ')

      expect(screen.getByRole('button', { name: /task/i })).toHaveAttribute('aria-pressed', 'true')
      expect(input).toHaveValue('') // Prefix consumed
    })

    it('typing "- " sets type to note', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, '- ')

      expect(screen.getByRole('button', { name: /note/i })).toHaveAttribute('aria-pressed', 'true')
    })

    it('typing "o " sets type to event', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'o ')

      expect(screen.getByRole('button', { name: /event/i })).toHaveAttribute('aria-pressed', 'true')
    })

    it('typing "? " sets type to question', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, '? ')

      expect(screen.getByRole('button', { name: /question/i })).toHaveAttribute('aria-pressed', 'true')
    })
  })

  describe('prefix detection edge cases', () => {
    it('preserves content when pasting dash-prefixed text', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)
      const textarea = screen.getByTestId('capture-bar-input')

      // Simulate paste of "- hello" - should switch to note and keep "hello"
      await user.click(textarea)
      await user.paste('- hello')

      expect(screen.getByRole('button', { name: /note/i })).toHaveAttribute('aria-pressed', 'true')
      expect(textarea).toHaveValue('hello')
    })

    it('preserves content when pasting task-prefixed text', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)
      const textarea = screen.getByTestId('capture-bar-input')

      await user.click(textarea)
      await user.paste('. buy milk')

      expect(screen.getByRole('button', { name: /task/i })).toHaveAttribute('aria-pressed', 'true')
      expect(textarea).toHaveValue('buy milk')
    })

    it('does not switch type when prefix appears mid-content', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)
      const textarea = screen.getByTestId('capture-bar-input')

      // Start with non-prefix text, then add text containing a prefix
      await user.type(textarea, 'buy - groceries')

      // Should remain as task (default) and preserve full content
      expect(screen.getByRole('button', { name: /task/i })).toHaveAttribute('aria-pressed', 'true')
      expect(textarea).toHaveValue('buy - groceries')
    })
  })

  describe('submission', () => {
    it('Enter submits with type prefix', async () => {
      const user = userEvent.setup()
      const onSubmit = vi.fn()
      render(<CaptureBar {...defaultProps} onSubmit={onSubmit} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Buy groceries{Enter}')

      expect(onSubmit).toHaveBeenCalledWith('. Buy groceries')
    })

    it('clears input after submission', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Buy groceries{Enter}')

      expect(input).toHaveValue('')
    })

    it('does not submit empty input', async () => {
      const user = userEvent.setup()
      const onSubmit = vi.fn()
      render(<CaptureBar {...defaultProps} onSubmit={onSubmit} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.click(input)
      await user.keyboard('{Enter}')

      expect(onSubmit).not.toHaveBeenCalled()
    })

    it('keeps focus after submission', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Buy groceries{Enter}')

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

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Action item{Enter}')

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

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Draft text')

      expect(localStorage.getItem('bujo-capture-bar-draft')).toBe('Draft text')
    })

    it('restores draft on mount', () => {
      localStorage.setItem('bujo-capture-bar-draft', 'Restored draft')
      localStorage.setItem('bujo-capture-bar-type', 'note')

      render(<CaptureBar {...defaultProps} />)

      expect(screen.getByDisplayValue('Restored draft')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /note/i })).toHaveAttribute('aria-pressed', 'true')
    })

    it('clears draft after submission', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Draft text{Enter}')

      expect(localStorage.getItem('bujo-capture-bar-draft')).toBeNull()
    })
  })

  describe('escape handling', () => {
    it('Escape clears input', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.type(input, 'Some text')
      await user.keyboard('{Escape}')

      expect(input).toHaveValue('')
    })

    it('Escape on empty input blurs', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
      await user.click(input)
      await user.keyboard('{Escape}')

      expect(input).not.toHaveFocus()
    })
  })

  describe('file import', () => {
    it('calls onFileImport when import button clicked', async () => {
      const user = userEvent.setup()
      const onFileImport = vi.fn()
      render(<CaptureBar {...defaultProps} onFileImport={onFileImport} />)

      await user.click(screen.getByRole('button', { name: /import file/i }))

      expect(onFileImport).toHaveBeenCalled()
    })
  })

  describe('multiline', () => {
    it('Shift+Enter adds newline', async () => {
      const user = userEvent.setup()
      render(<CaptureBar {...defaultProps} />)

      const input = screen.getByPlaceholderText(/add a task/i)
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
})

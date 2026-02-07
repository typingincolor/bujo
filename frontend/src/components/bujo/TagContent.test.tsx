import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TagContent } from './TagContent'

describe('TagContent', () => {
  describe('rendering plain content', () => {
    it('renders content without tags as plain text', () => {
      render(<TagContent content="Buy groceries" />)
      expect(screen.getByText('Buy groceries')).toBeInTheDocument()
    })

    it('renders empty content without crashing', () => {
      const { container } = render(<TagContent content="" />)
      expect(container.querySelector('span')).toBeInTheDocument()
    })
  })

  describe('rendering tags', () => {
    it('renders tag with # prefix as styled span', () => {
      render(<TagContent content="Buy groceries #shopping" />)
      expect(screen.getByText('#shopping')).toBeInTheDocument()
      expect(screen.getByText('Buy groceries')).toBeInTheDocument()
    })

    it('renders multiple tags', () => {
      render(<TagContent content="Task #work #urgent" />)
      expect(screen.getByText('#work')).toBeInTheDocument()
      expect(screen.getByText('#urgent')).toBeInTheDocument()
    })

    it('applies tag styling class to tag spans', () => {
      render(<TagContent content="Task #work" />)
      const tag = screen.getByText('#work')
      expect(tag.tagName).toBe('SPAN')
      expect(tag).toHaveClass('tag')
    })

    it('handles tag with hyphens', () => {
      render(<TagContent content="Note #my-project" />)
      expect(screen.getByText('#my-project')).toBeInTheDocument()
    })

    it('does not match hash followed by number', () => {
      render(<TagContent content="Issue #123" />)
      expect(screen.getByText('Issue #123')).toBeInTheDocument()
    })

    it('handles tag at start of content', () => {
      render(<TagContent content="#important Buy milk" />)
      expect(screen.getByText('#important')).toBeInTheDocument()
      expect(screen.getByText('Buy milk')).toBeInTheDocument()
    })

    it('handles content that is only a tag', () => {
      render(<TagContent content="#solo" />)
      expect(screen.getByText('#solo')).toBeInTheDocument()
    })
  })

  describe('tag click handling', () => {
    it('calls onTagClick with tag name when tag is clicked', async () => {
      const user = userEvent.setup()
      const onTagClick = vi.fn()
      render(<TagContent content="Task #work" onTagClick={onTagClick} />)

      await user.click(screen.getByText('#work'))

      expect(onTagClick).toHaveBeenCalledWith('work')
      expect(onTagClick).toHaveBeenCalledTimes(1)
    })

    it('passes correct tag name for each tag clicked', async () => {
      const user = userEvent.setup()
      const onTagClick = vi.fn()
      render(<TagContent content="Task #work #urgent" onTagClick={onTagClick} />)

      await user.click(screen.getByText('#urgent'))

      expect(onTagClick).toHaveBeenCalledWith('urgent')
    })

    it('does not crash when onTagClick is not provided', async () => {
      const user = userEvent.setup()
      render(<TagContent content="Task #work" />)

      await user.click(screen.getByText('#work'))
      // Should not throw
    })

    it('applies cursor-pointer class when onTagClick is provided', () => {
      const onTagClick = vi.fn()
      render(<TagContent content="Task #work" onTagClick={onTagClick} />)
      const tag = screen.getByText('#work')
      expect(tag).toHaveClass('cursor-pointer')
    })
  })
})

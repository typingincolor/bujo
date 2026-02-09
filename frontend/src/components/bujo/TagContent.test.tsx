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

  describe('rendering mentions', () => {
    it('renders @mention as styled span', () => {
      render(<TagContent content="Met with @john today" />)
      expect(screen.getByText('@john')).toBeInTheDocument()
      expect(screen.getByText('@john')).toHaveClass('mention')
    })

    it('renders multiple mentions', () => {
      render(<TagContent content="Call @alice and @bob" />)
      expect(screen.getByText('@alice')).toBeInTheDocument()
      expect(screen.getByText('@bob')).toBeInTheDocument()
    })

    it('handles mentions with dots', () => {
      render(<TagContent content="Email @alice.smith" />)
      expect(screen.getByText('@alice.smith')).toBeInTheDocument()
    })

    it('does not match @ followed by nothing', () => {
      render(<TagContent content="Send email @ noon" />)
      expect(screen.getByText('Send email @ noon')).toBeInTheDocument()
    })

    it('handles mixed tags and mentions', () => {
      render(<TagContent content="Meeting with @john about #project" />)
      expect(screen.getByText('@john')).toBeInTheDocument()
      expect(screen.getByText('#project')).toBeInTheDocument()
    })
  })

  describe('mention click handling', () => {
    it('calls onMentionClick with mention name on click', async () => {
      const user = userEvent.setup()
      const onMentionClick = vi.fn()
      render(<TagContent content="Met @john" onMentionClick={onMentionClick} />)

      await user.click(screen.getByText('@john'))

      expect(onMentionClick).toHaveBeenCalledWith('john')
      expect(onMentionClick).toHaveBeenCalledTimes(1)
    })

    it('applies cursor-pointer class when onMentionClick is provided', () => {
      const onMentionClick = vi.fn()
      render(<TagContent content="Met @john" onMentionClick={onMentionClick} />)
      const mention = screen.getByText('@john')
      expect(mention).toHaveClass('cursor-pointer')
    })

    it('does not apply cursor-pointer when onMentionClick is not provided', () => {
      render(<TagContent content="Met @john" />)
      const mention = screen.getByText('@john')
      expect(mention).not.toHaveClass('cursor-pointer')
    })
  })
})

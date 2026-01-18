import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AnswerQuestionModal } from './AnswerQuestionModal'

vi.mock('@/wailsjs/go/wails/App', () => ({
  AnswerQuestion: vi.fn().mockResolvedValue(undefined),
}))

import { AnswerQuestion } from '@/wailsjs/go/wails/App'

describe('AnswerQuestionModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders nothing when not open', () => {
    const { container } = render(
      <AnswerQuestionModal
        isOpen={false}
        questionId={1}
        questionContent="What is TDD?"
        onClose={() => {}}
        onAnswered={() => {}}
      />
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders modal when open', () => {
    render(
      <AnswerQuestionModal
        isOpen={true}
        questionId={1}
        questionContent="What is TDD?"
        onClose={() => {}}
        onAnswered={() => {}}
      />
    )
    expect(screen.getByText('Answer Question')).toBeInTheDocument()
  })

  it('displays the question content', () => {
    render(
      <AnswerQuestionModal
        isOpen={true}
        questionId={1}
        questionContent="What is TDD?"
        onClose={() => {}}
        onAnswered={() => {}}
      />
    )
    expect(screen.getByText('What is TDD?')).toBeInTheDocument()
  })

  it('renders answer input', () => {
    render(
      <AnswerQuestionModal
        isOpen={true}
        questionId={1}
        questionContent="What is TDD?"
        onClose={() => {}}
        onAnswered={() => {}}
      />
    )
    expect(screen.getByPlaceholderText(/your answer/i)).toBeInTheDocument()
  })

  it('calls AnswerQuestion binding when submitting', async () => {
    const onAnswered = vi.fn()
    const user = userEvent.setup()

    render(
      <AnswerQuestionModal
        isOpen={true}
        questionId={42}
        questionContent="What is TDD?"
        onClose={() => {}}
        onAnswered={onAnswered}
      />
    )

    const input = screen.getByPlaceholderText(/your answer/i)
    await user.type(input, 'Test-Driven Development')
    await user.click(screen.getByText('Submit Answer'))

    await waitFor(() => {
      expect(AnswerQuestion).toHaveBeenCalledWith(42, 'Test-Driven Development')
    })
  })

  it('calls onAnswered callback after successful submission', async () => {
    const onAnswered = vi.fn()
    const user = userEvent.setup()

    render(
      <AnswerQuestionModal
        isOpen={true}
        questionId={42}
        questionContent="What is TDD?"
        onClose={() => {}}
        onAnswered={onAnswered}
      />
    )

    const input = screen.getByPlaceholderText(/your answer/i)
    await user.type(input, 'Test-Driven Development')
    await user.click(screen.getByText('Submit Answer'))

    await waitFor(() => {
      expect(onAnswered).toHaveBeenCalled()
    })
  })

  it('calls onClose when cancel button is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()

    render(
      <AnswerQuestionModal
        isOpen={true}
        questionId={1}
        questionContent="What is TDD?"
        onClose={onClose}
        onAnswered={() => {}}
      />
    )

    await user.click(screen.getByText('Cancel'))
    expect(onClose).toHaveBeenCalled()
  })

  it('disables submit when answer is empty', () => {
    render(
      <AnswerQuestionModal
        isOpen={true}
        questionId={1}
        questionContent="What is TDD?"
        onClose={() => {}}
        onAnswered={() => {}}
      />
    )

    const submitButton = screen.getByText('Submit Answer')
    expect(submitButton).toBeDisabled()
  })

  it('clears input after successful submission', async () => {
    const user = userEvent.setup()

    render(
      <AnswerQuestionModal
        isOpen={true}
        questionId={42}
        questionContent="What is TDD?"
        onClose={() => {}}
        onAnswered={() => {}}
      />
    )

    const input = screen.getByPlaceholderText(/your answer/i)
    await user.type(input, 'Test-Driven Development')
    await user.click(screen.getByText('Submit Answer'))

    await waitFor(() => {
      expect(input).toHaveValue('')
    })
  })
})

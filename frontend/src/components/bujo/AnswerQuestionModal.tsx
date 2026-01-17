import { useState } from 'react'
import { cn } from '@/lib/utils'
import { AnswerQuestion } from '@/wailsjs/go/wails/App'

interface AnswerQuestionModalProps {
  isOpen: boolean
  questionId: number
  questionContent: string
  onClose: () => void
  onAnswered: () => void
}

export function AnswerQuestionModal({
  isOpen,
  questionId,
  questionContent,
  onClose,
  onAnswered,
}: AnswerQuestionModalProps) {
  const [answer, setAnswer] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  if (!isOpen) return null

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!answer.trim()) return

    setIsSubmitting(true)
    try {
      await AnswerQuestion(questionId, answer.trim())
      setAnswer('')
      onAnswered()
    } catch (err) {
      console.error('Failed to answer question:', err)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-background/80 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="relative z-10 w-full max-w-md bg-card border rounded-lg shadow-lg p-6 mx-4">
        <h2 className="text-lg font-display font-semibold mb-4">Answer Question</h2>

        {/* Question display */}
        <div className="mb-4 p-3 bg-secondary/50 rounded-md">
          <span className="text-bujo-question font-medium mr-2">?</span>
          <span className="text-sm">{questionContent}</span>
        </div>

        <form onSubmit={handleSubmit}>
          <textarea
            value={answer}
            onChange={(e) => setAnswer(e.target.value)}
            placeholder="Enter your answer..."
            rows={4}
            className={cn(
              'w-full px-3 py-2 rounded-md border bg-background',
              'focus:outline-none focus:ring-2 focus:ring-primary',
              'placeholder:text-muted-foreground text-sm resize-none'
            )}
            autoFocus
          />

          <div className="flex justify-end gap-3 mt-4">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm rounded-md border hover:bg-secondary transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!answer.trim() || isSubmitting}
              className={cn(
                'px-4 py-2 text-sm rounded-md transition-colors',
                answer.trim() && !isSubmitting
                  ? 'bg-primary text-primary-foreground hover:bg-primary/90'
                  : 'bg-muted text-muted-foreground cursor-not-allowed'
              )}
            >
              {isSubmitting ? 'Submitting...' : 'Submit Answer'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

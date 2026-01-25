import { useState, useRef, useEffect, useCallback, forwardRef, useImperativeHandle } from 'react'
import { Entry } from '@/types/bujo'
import { NAV_SIDEBAR_LEFT_CLASS, JOURNAL_SIDEBAR_RIGHT_CLASS } from '@/lib/layoutConstants'

interface CaptureBarProps {
  onSubmit: (content: string) => void
  onSubmitChild?: (parentId: number, content: string) => void
  onClearParent?: () => void
  onFileImport?: () => void
  parentEntry?: Entry | null
}

const DRAFT_KEY = 'bujo-capture-bar-draft'

export const CaptureBar = forwardRef<HTMLTextAreaElement, CaptureBarProps>(function CaptureBar(
  {
    onSubmit,
    onSubmitChild,
    onClearParent,
    onFileImport,
    parentEntry,
  },
  ref
) {
  const [content, setContent] = useState(() => {
    try {
      return localStorage.getItem(DRAFT_KEY) || ''
    } catch {
      return ''
    }
  })
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useImperativeHandle(ref, () => textareaRef.current as HTMLTextAreaElement)

  useEffect(() => {
    try {
      if (content) {
        localStorage.setItem(DRAFT_KEY, content)
      } else {
        localStorage.removeItem(DRAFT_KEY)
      }
    } catch {
      // Ignore localStorage errors (e.g., incognito mode)
    }
  }, [content])

  useEffect(() => {
    const textarea = textareaRef.current
    if (textarea) {
      textarea.style.height = 'auto'
      textarea.style.height = `${textarea.scrollHeight}px`
    }
  }, [content])

  const handleSubmit = useCallback(() => {
    if (!content.trim()) return

    // Submit content exactly as typed - user types their own prefix
    if (parentEntry && onSubmitChild) {
      onSubmitChild(parentEntry.id, content)
    } else {
      onSubmit(content)
    }

    setContent('')
    try {
      localStorage.removeItem(DRAFT_KEY)
    } catch {
      // Ignore localStorage errors (e.g., incognito mode)
    }
    textareaRef.current?.focus()
  }, [content, parentEntry, onSubmitChild, onSubmit])

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setContent(e.target.value)
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
    } else if (e.key === 'Escape') {
      e.preventDefault()
      if (content) {
        setContent('')
        try {
          localStorage.removeItem(DRAFT_KEY)
        } catch {
          // Ignore localStorage errors (e.g., incognito mode)
        }
      } else {
        textareaRef.current?.blur()
      }
    }
  }

  return (
    <div data-testid="capture-bar" className={`fixed bottom-0 ${NAV_SIDEBAR_LEFT_CLASS} ${JOURNAL_SIDEBAR_RIGHT_CLASS} flex flex-col gap-2 p-3 bg-card border rounded-lg`}>
      {parentEntry && (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <span>Adding to:</span>
          <span className="font-medium text-foreground">{parentEntry.content}</span>
          <button
            type="button"
            onClick={onClearParent}
            aria-label="Clear parent"
            className="ml-auto text-muted-foreground hover:text-foreground"
          >
            &times;
          </button>
        </div>
      )}
      <div className="flex items-center gap-2">
        <textarea
          ref={textareaRef}
          data-testid="capture-bar-input"
          value={content}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          placeholder="Capture a thought..."
          rows={1}
          style={{ fontFamily: 'monospace' }}
          className="flex-1 bg-transparent border-none outline-none text-sm placeholder:text-muted-foreground resize-none font-mono"
        />
        <button
          type="button"
          onClick={onFileImport}
          aria-label="Import file"
          className="text-muted-foreground hover:text-foreground"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="17 8 12 3 7 8" />
            <line x1="12" y1="3" x2="12" y2="15" />
          </svg>
        </button>
      </div>
    </div>
  )
})

import { useState, useRef, useEffect, useCallback, forwardRef, useImperativeHandle } from 'react'
import { Entry } from '@/types/bujo'

interface CaptureBarProps {
  onSubmit: (content: string) => void
  onSubmitChild?: (parentId: number, content: string) => void
  onClearParent?: () => void
  parentEntry?: Entry | null
  sidebarWidth?: number
  isSidebarCollapsed?: boolean
}

const DRAFT_KEY = 'bujo-capture-bar-draft'

export const CaptureBar = forwardRef<HTMLTextAreaElement, CaptureBarProps>(function CaptureBar(
  {
    onSubmit,
    onSubmitChild,
    onClearParent,
    parentEntry,
    sidebarWidth = 512,
    isSidebarCollapsed = false,
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

  const adjustHeight = useCallback((textarea: HTMLTextAreaElement) => {
    textarea.style.height = 'auto'
    const newHeight = Math.max(textarea.scrollHeight, 24)
    textarea.style.height = `${newHeight}px`
  }, [])

  useEffect(() => {
    if (textareaRef.current) {
      adjustHeight(textareaRef.current)
    }
  }, [content, adjustHeight])

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
    const textarea = e.target
    setContent(textarea.value)
    // Immediately adjust height on input
    adjustHeight(textarea)
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
    <div
      data-testid="capture-bar"
      className="fixed bottom-3 flex flex-col gap-2 p-3 bg-card border rounded-lg"
      style={{
        right: isSidebarCollapsed ? '0.75rem' : `${sidebarWidth + 12}px`,
        left: 'calc(14rem + 0.75rem)' // left-56 + 0.75rem gap
      }}
    >
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
      <textarea
        ref={textareaRef}
        data-testid="capture-bar-input"
        value={content}
        onChange={handleChange}
        onKeyDown={handleKeyDown}
        placeholder="Capture a thought..."
        className="w-full bg-transparent border-none outline-none text-sm placeholder:text-muted-foreground resize-none overflow-hidden font-mono"
      />
    </div>
  )
})

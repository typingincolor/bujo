import { useState, useRef, useEffect, useCallback, forwardRef, useImperativeHandle } from 'react'
import { cn } from '@/lib/utils'
import { Entry } from '@/types/bujo'

type CaptureType = 'task' | 'note' | 'event' | 'question'

interface CaptureBarProps {
  onSubmit: (content: string) => void
  onSubmitChild?: (parentId: number, content: string) => void
  onClearParent?: () => void
  onFileImport?: () => void
  parentEntry?: Entry | null
}

const TYPES: CaptureType[] = ['task', 'note', 'event', 'question']

const TYPE_PREFIXES: Record<CaptureType, string> = {
  task: '. ',
  note: '- ',
  event: 'o ',
  question: '? ',
}

const TYPE_SYMBOLS: Record<CaptureType, string> = {
  task: '.',
  note: '-',
  event: 'o',
  question: '?',
}

const PREFIX_TO_TYPE: Record<string, CaptureType> = {
  '. ': 'task',
  '- ': 'note',
  'o ': 'event',
  '? ': 'question',
}

const DRAFT_KEY = 'bujo-capture-bar-draft'
const TYPE_KEY = 'bujo-capture-bar-type'

function getPlaceholder(type: CaptureType): string {
  return `Add a ${type}...`
}

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
    return localStorage.getItem(DRAFT_KEY) || ''
  })
  const [selectedType, setSelectedType] = useState<CaptureType>(() => {
    const stored = localStorage.getItem(TYPE_KEY) as CaptureType | null
    return stored && TYPES.includes(stored) ? stored : 'task'
  })
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useImperativeHandle(ref, () => textareaRef.current as HTMLTextAreaElement)

  useEffect(() => {
    if (content) {
      localStorage.setItem(DRAFT_KEY, content)
    } else {
      localStorage.removeItem(DRAFT_KEY)
    }
  }, [content])

  useEffect(() => {
    localStorage.setItem(TYPE_KEY, selectedType)
  }, [selectedType])

  const handleSubmit = useCallback(() => {
    if (!content.trim()) return

    const prefixedContent = TYPE_PREFIXES[selectedType] + content

    if (parentEntry && onSubmitChild) {
      onSubmitChild(parentEntry.id, prefixedContent)
    } else {
      onSubmit(prefixedContent)
    }

    setContent('')
    localStorage.removeItem(DRAFT_KEY)
    textareaRef.current?.focus()
  }, [content, selectedType, parentEntry, onSubmitChild, onSubmit])

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newValue = e.target.value

    for (const [prefix, type] of Object.entries(PREFIX_TO_TYPE)) {
      if (newValue === prefix) {
        setSelectedType(type)
        setContent('')
        return
      }
    }

    setContent(newValue)
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
    } else if (e.key === 'Tab' && !content) {
      e.preventDefault()
      const currentIndex = TYPES.indexOf(selectedType)
      const nextIndex = (currentIndex + 1) % TYPES.length
      setSelectedType(TYPES[nextIndex])
    } else if (e.key === 'Escape') {
      e.preventDefault()
      if (content) {
        setContent('')
        localStorage.removeItem(DRAFT_KEY)
      } else {
        textareaRef.current?.blur()
      }
    }
  }

  return (
    <div data-testid="capture-bar" className="flex flex-col gap-2 p-3 bg-card border rounded-lg">
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
        <div className="flex gap-1">
          {TYPES.map((type) => (
            <button
              key={type}
              type="button"
              onClick={() => setSelectedType(type)}
              aria-pressed={selectedType === type}
              aria-label={type}
              className={cn(
                'w-8 h-8 text-sm font-mono rounded',
                selectedType === type
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-muted text-muted-foreground hover:bg-muted/80'
              )}
            >
              {TYPE_SYMBOLS[type]}
            </button>
          ))}
        </div>
        <textarea
          ref={textareaRef}
          data-testid="capture-bar-input"
          value={content}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          placeholder={getPlaceholder(selectedType)}
          rows={1}
          className="flex-1 bg-transparent border-none outline-none text-sm placeholder:text-muted-foreground resize-none"
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

import { useState, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'

interface InlineEntryInputProps {
  onSubmit: (content: string) => void
  onCancel: () => void
  depth?: number
}

export function InlineEntryInput({ onSubmit, onCancel, depth = 0 }: InlineEntryInputProps) {
  const [content, setContent] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    inputRef.current?.focus()
  }, [])

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        onCancel()
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [onCancel])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      if (content.trim()) {
        onSubmit(content)
        setContent('')
      }
    } else if (e.key === 'Escape') {
      e.preventDefault()
      onCancel()
    }
  }

  return (
    <div
      ref={containerRef}
      className={cn(
        'flex items-center gap-2 p-2 rounded-lg border border-primary bg-card'
      )}
      style={{ marginLeft: depth > 0 ? `${depth * 16}px` : undefined }}
    >
      <input
        ref={inputRef}
        type="text"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Type entry (. task, - note, o event, ? question)"
        className="flex-1 bg-transparent border-none outline-none text-sm placeholder:text-muted-foreground"
      />
    </div>
  )
}

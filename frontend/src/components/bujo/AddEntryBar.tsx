import { useState } from 'react'
import { cn } from '@/lib/utils'
import { Plus } from 'lucide-react'
import { EntryType } from '@/types/bujo'

interface AddEntryBarProps {
  onAdd?: (content: string, type: EntryType) => void
}

const entryTypes: { type: EntryType; symbol: string; label: string }[] = [
  { type: 'task', symbol: '•', label: 'Task' },
  { type: 'note', symbol: '–', label: 'Note' },
  { type: 'event', symbol: '○', label: 'Event' },
]

export function AddEntryBar({ onAdd }: AddEntryBarProps) {
  const [content, setContent] = useState('')
  const [selectedType, setSelectedType] = useState<EntryType>('task')
  const [isFocused, setIsFocused] = useState(false)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (content.trim()) {
      onAdd?.(content.trim(), selectedType)
      setContent('')
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className={cn(
        'flex items-center gap-2 p-3 rounded-lg border-2 border-dashed transition-all',
        isFocused
          ? 'border-primary bg-card shadow-sm'
          : 'border-border bg-transparent hover:border-muted-foreground/50'
      )}
    >
      {/* Type selector */}
      <div className="flex items-center gap-1">
        {entryTypes.map((et) => (
          <button
            key={et.type}
            type="button"
            onClick={() => setSelectedType(et.type)}
            className={cn(
              'w-7 h-7 rounded flex items-center justify-center text-lg transition-colors',
              selectedType === et.type
                ? 'bg-primary text-primary-foreground'
                : 'text-muted-foreground hover:bg-secondary'
            )}
            title={et.label}
          >
            {et.symbol}
          </button>
        ))}
      </div>

      {/* Input */}
      <input
        type="text"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        onFocus={() => setIsFocused(true)}
        onBlur={() => setIsFocused(false)}
        placeholder="What's on your mind?"
        className="flex-1 bg-transparent border-none outline-none text-sm placeholder:text-muted-foreground"
      />

      {/* Submit */}
      <button
        type="submit"
        disabled={!content.trim()}
        className={cn(
          'p-2 rounded-md transition-all',
          content.trim()
            ? 'bg-primary text-primary-foreground hover:bg-primary/90'
            : 'text-muted-foreground cursor-not-allowed'
        )}
      >
        <Plus className="w-4 h-4" />
      </button>
    </form>
  )
}

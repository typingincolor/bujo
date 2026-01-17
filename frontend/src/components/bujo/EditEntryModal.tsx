import { useState, useEffect, useRef } from 'react'

interface EditEntryModalProps {
  isOpen: boolean
  initialContent: string
  onSave: (content: string) => void
  onCancel: () => void
}

function EditEntryModalContent({
  initialContent,
  onSave,
  onCancel,
}: Omit<EditEntryModalProps, 'isOpen'>) {
  const [content, setContent] = useState(initialContent)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    inputRef.current?.focus()
  }, [])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && content.trim()) {
      onSave(content)
    } else if (e.key === 'Escape') {
      onCancel()
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div
        className="absolute inset-0 bg-black/50"
        onClick={onCancel}
      />
      <div className="relative bg-card border border-border rounded-lg shadow-lg p-6 max-w-lg w-full mx-4 animate-fade-in">
        <h2 className="text-lg font-display font-semibold mb-4">Edit Entry</h2>
        <input
          ref={inputRef}
          type="text"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          onKeyDown={handleKeyDown}
          className="w-full px-3 py-2 text-sm rounded-md border border-border bg-background focus:outline-none focus:ring-2 focus:ring-primary/50 mb-4"
          placeholder="Entry content"
        />
        <div className="flex justify-end gap-3">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-sm rounded-md border border-border hover:bg-secondary transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={() => onSave(content)}
            disabled={!content.trim()}
            className="px-4 py-2 text-sm rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Save
          </button>
        </div>
      </div>
    </div>
  )
}

export function EditEntryModal({
  isOpen,
  initialContent,
  onSave,
  onCancel,
}: EditEntryModalProps) {
  if (!isOpen) return null

  return (
    <EditEntryModalContent
      key={initialContent}
      initialContent={initialContent}
      onSave={onSave}
      onCancel={onCancel}
    />
  )
}

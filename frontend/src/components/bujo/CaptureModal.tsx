import { useState, useEffect, useCallback } from 'react'
import { cn, startOfDay } from '@/lib/utils'
import { AddEntry, OpenFileDialog, ReadFile } from '@/wailsjs/go/wails/App'
import { OnFileDrop, OnFileDropOff } from '@/wailsjs/runtime/runtime'
import { toWailsTime } from '@/lib/wailsTime'

interface CaptureModalProps {
  isOpen: boolean
  onClose: () => void
  onEntriesCreated: () => void
}

const DRAFT_KEY = 'bujo-capture-draft'

export function CaptureModal({
  isOpen,
  onClose,
  onEntriesCreated,
}: CaptureModalProps) {
  const [content, setContent] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [importError, setImportError] = useState<string | null>(null)

  useEffect(() => {
    if (isOpen) {
      setImportError(null)
      const draft = localStorage.getItem(DRAFT_KEY)
      if (draft) {
        setContent(draft)
      }
    }
  }, [isOpen])

  useEffect(() => {
    if (!isOpen) return

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault()
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, onClose])

  const saveDraft = useCallback((text: string) => {
    if (text) {
      localStorage.setItem(DRAFT_KEY, text)
    } else {
      localStorage.removeItem(DRAFT_KEY)
    }
  }, [])

  const handleContentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newContent = e.target.value
    setContent(newContent)
    saveDraft(newContent)
  }

  const handleFileContent = useCallback((fileContent: string) => {
    if (!fileContent) return

    setContent((prevContent) => {
      const newContent = prevContent.trim()
        ? prevContent + '\n' + fileContent
        : fileContent
      saveDraft(newContent)
      return newContent
    })
  }, [saveDraft])

  const handleImportFile = async () => {
    setImportError(null)
    try {
      const fileContent = await OpenFileDialog()
      handleFileContent(fileContent)
    } catch (err) {
      console.error('Failed to import file:', err)
      setImportError('Failed to import file. Please try again.')
    }
  }

  useEffect(() => {
    if (!isOpen) return

    const handleFileDrop = async (_x: number, _y: number, paths: string[]) => {
      if (paths.length === 0) return
      setImportError(null)
      try {
        const fileContent = await ReadFile(paths[0])
        handleFileContent(fileContent)
      } catch (err) {
        console.error('Failed to read dropped file:', err)
        setImportError('Failed to read file. Please try again.')
      }
    }

    OnFileDrop(handleFileDrop, false)
    return () => {
      OnFileDropOff()
    }
  }, [isOpen, handleFileContent])

  if (!isOpen) return null

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!content.trim()) return

    setIsSubmitting(true)
    try {
      const today = startOfDay(new Date())
      await AddEntry(content.trim(), toWailsTime(today))
      setContent('')
      localStorage.removeItem(DRAFT_KEY)
      onEntriesCreated()
    } catch (err) {
      console.error('Failed to create entries:', err)
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
      <div className="relative z-10 w-full max-w-2xl bg-card border rounded-lg shadow-lg p-6 mx-4 max-h-[80vh] flex flex-col">
        <h2 className="text-lg font-display font-semibold mb-4">Capture Entries</h2>

        {/* Syntax help */}
        <div className="mb-4 p-3 bg-secondary/50 rounded-md text-sm">
          <div className="font-medium mb-2">Entry Prefixes:</div>
          <div className="grid grid-cols-2 gap-2 text-muted-foreground">
            <div><code className="bg-muted px-1 rounded">.</code> Task</div>
            <div><code className="bg-muted px-1 rounded">-</code> Note</div>
            <div><code className="bg-muted px-1 rounded">o</code> Event</div>
            <div><code className="bg-muted px-1 rounded">?</code> Question</div>
          </div>
          <div className="mt-2 text-muted-foreground">
            <span className="font-medium">Tip:</span> Indent with spaces/tabs to create child entries.
          </div>
        </div>

        {importError && (
          <div className="mb-4 p-3 bg-destructive/10 border border-destructive/20 rounded-md text-sm text-destructive">
            {importError}
          </div>
        )}

        <form onSubmit={handleSubmit} className="flex-1 flex flex-col">
          <textarea
            value={content}
            onChange={handleContentChange}
            placeholder="Enter entries (one per line)..."
            rows={10}
            className={cn(
              'w-full flex-1 px-3 py-2 rounded-md border bg-background',
              'focus:outline-none focus:ring-2 focus:ring-primary',
              'placeholder:text-muted-foreground text-sm font-mono resize-none'
            )}
            autoFocus
          />

          <div className="flex justify-end gap-3 mt-4">
            <button
              type="button"
              onClick={handleImportFile}
              className="px-4 py-2 text-sm rounded-md border hover:bg-secondary transition-colors mr-auto"
            >
              Import File
            </button>
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm rounded-md border hover:bg-secondary transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!content.trim() || isSubmitting}
              className={cn(
                'px-4 py-2 text-sm rounded-md transition-colors',
                content.trim() && !isSubmitting
                  ? 'bg-primary text-primary-foreground hover:bg-primary/90'
                  : 'bg-muted text-muted-foreground cursor-not-allowed'
              )}
            >
              {isSubmitting ? 'Saving...' : 'Save Entries'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

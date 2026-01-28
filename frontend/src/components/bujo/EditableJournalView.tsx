import { useState, useRef, useCallback } from 'react'
import { useEditableDocument } from '@/hooks/useEditableDocument'
import { DeletionReviewDialog } from './DeletionReviewDialog'
import { BujoEditor } from '@/lib/codemirror/BujoEditor'

interface EditableJournalViewProps {
  date: Date
}

export function EditableJournalView({ date }: EditableJournalViewProps) {
  const {
    document,
    setDocument,
    isLoading,
    error,
    isDirty,
    validationErrors,
    deletedEntries,
    restoreDeletion,
    save,
    discardChanges,
    lastSaved,
    hasDraft,
    restoreDraft,
    discardDraft,
  } = useEditableDocument(date)

  const [saveError, setSaveError] = useState<string | null>(null)
  const [showDeletionDialog, setShowDeletionDialog] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleSave = useCallback(async () => {
    if (deletedEntries.length > 0) {
      setShowDeletionDialog(true)
      return
    }
    const result = await save()
    if (!result.success && result.error) {
      setSaveError(result.error)
    }
  }, [deletedEntries.length, save])

  const handleImport = useCallback(() => {
    fileInputRef.current?.click()
  }, [])

  const handleFileChange = useCallback(async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    const content = await file.text()
    const separator = document.endsWith('\n') ? '' : '\n'
    setDocument(document + separator + content)

    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }, [document, setDocument])

  const handleEscape = useCallback(() => {
    if (window.document.activeElement instanceof HTMLElement) {
      window.document.activeElement.blur()
    }
  }, [])

  const handleConfirmDeletions = async () => {
    setShowDeletionDialog(false)
    const result = await save()
    if (!result.success && result.error) {
      setSaveError(result.error)
    }
  }

  const handleCancelDeletions = () => {
    setShowDeletionDialog(false)
  }

  const handleRestoreDeletion = (entityId: string) => {
    restoreDeletion(entityId)
  }

  const handleDeleteLine = (lineNumber: number) => {
    const lines = document.split('\n')
    const newLines = lines.filter((_, index) => index !== lineNumber - 1)
    setDocument(newLines.join('\n'))
  }

  const handleChangeToTask = (lineNumber: number) => {
    const lines = document.split('\n')
    const line = lines[lineNumber - 1]
    const leadingSpaces = line.match(/^(\s*)/)?.[1] || ''
    const contentWithoutSymbol = line.replace(/^\s*\S/, leadingSpaces + '.')
    lines[lineNumber - 1] = contentWithoutSymbol
    setDocument(lines.join('\n'))
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64 text-muted-foreground">
        Loading...
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64 text-destructive">
        {error}
      </div>
    )
  }

  return (
    <div className="flex flex-col h-[calc(100vh-220px)]">
      {hasDraft && (
        <div className="flex items-center gap-2 p-3 mb-3 bg-amber-500/10 border border-amber-500/30 rounded-lg text-sm">
          <span className="text-muted-foreground">Unsaved changes found</span>
          <button
            onClick={restoreDraft}
            className="px-2 py-1 text-xs bg-primary text-primary-foreground rounded hover:bg-primary/90"
          >
            Restore
          </button>
          <button
            onClick={discardDraft}
            className="px-2 py-1 text-xs bg-secondary text-secondary-foreground rounded hover:bg-secondary/80"
          >
            Discard Draft
          </button>
        </div>
      )}

      <div className="flex-1 min-h-0 overflow-hidden border border-border">
        <BujoEditor
          value={document}
          onChange={setDocument}
          onSave={handleSave}
          onImport={handleImport}
          onEscape={handleEscape}
        />
      </div>

      <input type="file" ref={fileInputRef} style={{ display: 'none' }} onChange={handleFileChange} />

      {validationErrors.length > 0 && (
        <div className="mt-3 p-3 bg-destructive/10 border border-destructive/30 rounded-lg">
          <span className="text-sm font-medium text-destructive">{validationErrors.length} errors</span>
          {validationErrors.map((err, i) => (
            <div key={i} className="mt-2 text-sm">
              <span className="text-muted-foreground">Line {err.lineNumber}:</span> {err.message}
              {err.quickFixes?.includes('delete') && (
                <button
                  onClick={() => handleDeleteLine(err.lineNumber)}
                  className="ml-2 text-xs text-destructive hover:underline"
                >
                  Delete line
                </button>
              )}
              {err.quickFixes?.includes('change-to-task') && (
                <button
                  onClick={() => handleChangeToTask(err.lineNumber)}
                  className="ml-2 text-xs text-primary hover:underline"
                >
                  Change to task
                </button>
              )}
            </div>
          ))}
        </div>
      )}

      {deletedEntries.length > 0 && (
        <div className="mt-3 text-sm text-muted-foreground">
          {deletedEntries.length} deletions pending
        </div>
      )}

      <DeletionReviewDialog
        isOpen={showDeletionDialog}
        deletedEntries={deletedEntries}
        onConfirm={handleConfirmDeletions}
        onCancel={handleCancelDeletions}
        onRestore={handleRestoreDeletion}
      />

      {saveError && (
        <div className="mt-3 p-3 bg-destructive/10 border border-destructive/30 rounded-lg text-sm text-destructive">
          {saveError}
        </div>
      )}

      <div className="flex items-center justify-between mt-3">
        <div className="text-sm text-muted-foreground">
          {lastSaved && (
            <span>✓ Saved at {lastSaved.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })}</span>
          )}
          {isDirty && <span data-testid="unsaved-indicator" className="ml-2 text-amber-500">● Unsaved changes</span>}
        </div>
        <div className="flex gap-2">
          {isDirty && (
            <button
              onClick={discardChanges}
              className="px-3 py-1.5 text-sm bg-secondary text-secondary-foreground rounded hover:bg-secondary/80"
            >
              Discard
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

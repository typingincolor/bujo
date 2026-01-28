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

  const formatDate = (d: Date) => {
    return d.toLocaleDateString('en-US', {
      weekday: 'long',
      month: 'short',
      day: 'numeric',
    })
  }

  if (isLoading) {
    return <div>Loading...</div>
  }

  if (error) {
    return <div>{error}</div>
  }

  return (
    <div>
      <header>
        <h1>
          {formatDate(date)}
          {isDirty && <span data-testid="unsaved-indicator">●</span>}
        </h1>
      </header>

      {hasDraft && (
        <div>
          <p>Unsaved changes found</p>
          <button onClick={restoreDraft}>Restore</button>
          <button onClick={discardDraft}>Discard Draft</button>
        </div>
      )}

      <BujoEditor
        value={document}
        onChange={setDocument}
        onSave={handleSave}
        onImport={handleImport}
        onEscape={handleEscape}
      />

      <input type="file" ref={fileInputRef} style={{ display: 'none' }} onChange={handleFileChange} />
      <button onClick={() => fileInputRef.current?.click()}>Import</button>

      {validationErrors.length > 0 && (
        <div>
          <span>{validationErrors.length} errors</span>
          {validationErrors.map((err, i) => (
            <div key={i}>
              Line {err.lineNumber}: {err.message}
              {err.quickFixes?.includes('delete') && (
                <button onClick={() => handleDeleteLine(err.lineNumber)}>Delete line</button>
              )}
              {err.quickFixes?.includes('change-to-task') && (
                <button onClick={() => handleChangeToTask(err.lineNumber)}>Change to task</button>
              )}
            </div>
          ))}
        </div>
      )}

      {deletedEntries.length > 0 && (
        <div>{deletedEntries.length} deletions pending</div>
      )}

      <DeletionReviewDialog
        isOpen={showDeletionDialog}
        deletedEntries={deletedEntries}
        onConfirm={handleConfirmDeletions}
        onCancel={handleCancelDeletions}
        onRestore={handleRestoreDeletion}
      />

      {saveError && <div>{saveError}</div>}

      {lastSaved && (
        <div>
          ✓ Saved at {lastSaved.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })}
        </div>
      )}

      {isDirty && (
        <button onClick={discardChanges}>Discard</button>
      )}
    </div>
  )
}

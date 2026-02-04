import { useState, useRef, useCallback } from 'react'
import { useEditableDocument } from '@/hooks/useEditableDocument'
import { BujoEditor } from '@/lib/codemirror/BujoEditor'
import { scanForNewSpecialEntries, SpecialEntries } from '@/hooks/useSaveWithDialogs'
import { MigrateBatchModal } from '@/components/bujo/MigrateBatchModal'
import { ListPickerModal } from '@/components/bujo/ListPickerModal'

interface EditableJournalViewProps {
  date: Date
}

export function EditableJournalView({ date }: EditableJournalViewProps) {
  const {
    document,
    originalDocument,
    setDocument,
    isLoading,
    error,
    isDirty,
    validationErrors,
    save,
    saveWithActions,
    discardChanges,
    lastSaved,
    hasDraft,
    restoreDraft,
    discardDraft,
  } = useEditableDocument(date)

  const [saveError, setSaveError] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const [pendingSpecial, setPendingSpecial] = useState<SpecialEntries | null>(null)
  const [migrateDate, setMigrateDate] = useState<Date | null>(null)

  const doSaveWithActions = async (migDate: Date | null, listId: number | null) => {
    setPendingSpecial(null)
    setMigrateDate(null)

    const actions = {
      migrateDate: migDate ?? undefined,
      listId: listId ?? undefined,
    }

    const result = await saveWithActions(actions)
    if (!result.success && result.error) {
      setSaveError(result.error)
    }
  }

  const handleSave = useCallback(async () => {
    const special = scanForNewSpecialEntries(document, originalDocument)
    if (special.hasSpecialEntries) {
      setPendingSpecial(special)
      setMigrateDate(null)
      return
    }
    const result = await save()
    if (!result.success && result.error) {
      setSaveError(result.error)
    }
  }, [save, document, originalDocument])

  const handleMigrateConfirm = (dateStr: string) => {
    const parsed = new Date(dateStr + 'T00:00:00')

    if (pendingSpecial && pendingSpecial.movedToListEntries.length > 0) {
      setMigrateDate(parsed)
      return
    }

    doSaveWithActions(parsed, null)
  }

  const handleListSelect = (listId: number) => {
    doSaveWithActions(migrateDate, listId)
  }

  const handleDialogCancel = () => {
    setPendingSpecial(null)
    setMigrateDate(null)
  }

  const handleImport = useCallback(() => {
    fileInputRef.current?.click()
  }, [])

  const handleFileChange = useCallback(async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    const MAX_FILE_SIZE = 1_000_000
    if (file.size > MAX_FILE_SIZE) {
      setSaveError('File too large (max 1MB)')
      if (fileInputRef.current) fileInputRef.current.value = ''
      return
    }

    if (file.type && !file.type.startsWith('text/')) {
      setSaveError('Only text files are supported')
      if (fileInputRef.current) fileInputRef.current.value = ''
      return
    }

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
    <div className="flex flex-col flex-1 min-h-0">
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

      <div className="relative flex-1 min-h-0 border border-border">
        <div className="absolute inset-0 overflow-hidden">
          <BujoEditor
            value={document}
            onChange={setDocument}
            onSave={handleSave}
            onImport={handleImport}
            onEscape={handleEscape}
          />
        </div>
      </div>

      <input type="file" ref={fileInputRef} style={{ display: 'none' }} onChange={handleFileChange} />

      <MigrateBatchModal
        isOpen={pendingSpecial !== null && pendingSpecial.migratedEntries.length > 0 && migrateDate === null}
        entries={pendingSpecial?.migratedEntries ?? []}
        onMigrate={handleMigrateConfirm}
        onCancel={handleDialogCancel}
      />

      <ListPickerModal
        isOpen={
          pendingSpecial !== null &&
          pendingSpecial.movedToListEntries.length > 0 &&
          (pendingSpecial.migratedEntries.length === 0 || migrateDate !== null)
        }
        entries={pendingSpecial?.movedToListEntries ?? []}
        onSelect={handleListSelect}
        onCancel={handleDialogCancel}
      />

      <div className="flex flex-wrap gap-x-4 gap-y-1 mt-2 text-xs text-muted-foreground">
        <span><kbd>⌘S</kbd> Save</span>
        <span><kbd>⌘I</kbd> Import</span>
        <span><kbd>⌘F</kbd> Find</span>
        <span><kbd>⌘H</kbd> Replace</span>
        <span><kbd>Tab</kbd> Indent</span>
        <span><kbd>Esc</kbd> Unfocus</span>
        <span className="text-muted-foreground/50">|</span>
        <span><kbd>⌘⏎</kbd> Line below</span>
        <span><kbd>⌘⇧⏎</kbd> Line above</span>
        <span><kbd>⌘⇧K</kbd> Delete line</span>
        <span><kbd>⌘⇧D</kbd> Duplicate</span>
        <span><kbd>⌥↑</kbd> Move up</span>
        <span><kbd>⌥↓</kbd> Move down</span>
        <span><kbd>⌃A</kbd> Line start</span>
        <span><kbd>⌃E</kbd> Line end</span>
      </div>

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

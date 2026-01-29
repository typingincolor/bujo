import type { DeletedEntry } from '@/hooks/useEditableDocument'

function stripEntityIdPrefix(content: string): string {
  return content.replace(/^\s*\[[^\]\n]+\] /, '')
}

interface DeletionReviewDialogProps {
  isOpen: boolean
  deletedEntries: DeletedEntry[]
  onConfirm: () => void
  onCancel: () => void
  onRestore: (entityId: string) => void
  onDiscardAll?: () => void
  date?: string
}

export function DeletionReviewDialog({
  isOpen,
  deletedEntries,
  onConfirm,
  onCancel,
  onRestore,
  onDiscardAll,
  date,
}: DeletionReviewDialogProps) {
  if (!isOpen) {
    return null
  }

  const titleId = 'deletion-review-dialog-title'
  const title = date ? `Save changes to ${date}?` : 'Confirm Deletions'

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onCancel} />
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        className="relative bg-card border border-border rounded-lg shadow-lg p-6 max-w-lg w-full mx-4 animate-fade-in"
      >
        <h2 id={titleId} className="text-lg font-display font-semibold mb-4">
          {title}
        </h2>

        {deletedEntries.length === 0 ? (
          <p className="text-sm text-muted-foreground">No deletions</p>
        ) : (
          <>
            <p className="text-sm text-muted-foreground mb-4">
              {deletedEntries.length} items will be permanently deleted
            </p>
            <ul className="space-y-2 mb-6 max-h-64 overflow-y-auto">
              {deletedEntries.map((entry) => (
                <li
                  key={entry.entityId}
                  className="flex items-center justify-between gap-3 p-2 bg-secondary/50 rounded text-sm"
                >
                  <span className="font-mono truncate">{stripEntityIdPrefix(entry.content)}</span>
                  <button
                    onClick={() => onRestore(entry.entityId)}
                    className="text-xs px-2 py-1 rounded border border-border hover:bg-secondary transition-colors shrink-0"
                  >
                    Restore
                  </button>
                </li>
              ))}
            </ul>
          </>
        )}

        <div className="flex justify-between gap-3">
          <div>
            {onDiscardAll && (
              <button
                onClick={onDiscardAll}
                className="px-4 py-2 text-sm rounded-md text-destructive hover:bg-destructive/10 transition-colors"
              >
                Discard All Changes
              </button>
            )}
          </div>
          <div className="flex gap-3">
            <button
              onClick={onCancel}
              className="px-4 py-2 text-sm rounded-md border border-border hover:bg-secondary transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={onConfirm}
              className="px-4 py-2 text-sm rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
            >
              Save and Delete
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

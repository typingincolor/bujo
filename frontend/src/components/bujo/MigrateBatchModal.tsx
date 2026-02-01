import { useState } from 'react';
import { cn } from '@/lib/utils';

interface MigrateBatchModalProps {
  isOpen: boolean;
  entries: string[];
  onMigrate: (date: string) => void;
  onCancel: () => void;
}

export function MigrateBatchModal({ isOpen, entries, onMigrate, onCancel }: MigrateBatchModalProps) {
  const [selectedDate, setSelectedDate] = useState(() => {
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    return tomorrow.toISOString().split('T')[0];
  });

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (selectedDate) {
      onMigrate(selectedDate);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-background rounded-lg shadow-lg p-6 w-full max-w-md animate-fade-in">
        <h2 className="text-lg font-semibold mb-4">Migrate Entries</h2>
        <p className="text-sm text-muted-foreground mb-2">
          {entries.length === 1 ? 'Migrate this entry' : `Migrate ${entries.length} entries`} to a future date:
        </p>
        <ul className="mb-4 space-y-1">
          {entries.map((entry, i) => (
            <li key={i} className="text-sm text-foreground truncate pl-3 border-l-2 border-primary/50">
              {entry}
            </li>
          ))}
        </ul>
        <form onSubmit={handleSubmit}>
          <input
            type="date"
            value={selectedDate}
            onChange={(e) => setSelectedDate(e.target.value)}
            min={new Date().toISOString().split('T')[0]}
            className={cn(
              'w-full px-3 py-2 rounded-lg border border-border bg-background',
              'focus:outline-none focus:ring-2 focus:ring-primary/50 mb-4'
            )}
            autoFocus
          />
          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={onCancel}
              className="px-4 py-2 rounded-lg text-sm bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 rounded-lg text-sm bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
            >
              Migrate
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

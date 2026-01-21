import { useState, useEffect } from 'react';
import { cn } from '@/lib/utils';
import { GetLists } from '@/wailsjs/go/wails/App';
import { wails } from '@/wailsjs/go/models';
import { List } from 'lucide-react';

interface ListPickerModalProps {
  isOpen: boolean;
  entryContent: string;
  onSelect: (listId: number) => void;
  onCancel: () => void;
}

export function ListPickerModal({ isOpen, entryContent, onSelect, onCancel }: ListPickerModalProps) {
  const [lists, setLists] = useState<wails.ListWithItems[]>([]);
  const [hasFetched, setHasFetched] = useState(false);
  const [selectedListId, setSelectedListId] = useState<number | null>(null);

  useEffect(() => {
    if (!isOpen) {
      // Reset fetch status when modal closes to refetch on next open
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setHasFetched(false);
    }
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen || hasFetched) return;

    GetLists()
      .then((fetchedLists) => {
        setLists(fetchedLists || []);
        if (fetchedLists && fetchedLists.length > 0) {
          setSelectedListId(fetchedLists[0].ID);
        }
        setHasFetched(true);
      })
      .catch((err) => {
        console.error('Failed to fetch lists:', err);
        setLists([]);
        setHasFetched(true);
      });
  }, [isOpen, hasFetched]);

  const loading = isOpen && !hasFetched;

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (selectedListId !== null) {
      onSelect(selectedListId);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-background rounded-lg shadow-lg p-6 w-full max-w-md animate-fade-in">
        <h2 className="text-lg font-semibold mb-4">Move to List</h2>
        <p className="text-sm text-muted-foreground mb-4">
          Move &ldquo;{entryContent}&rdquo; to a list:
        </p>
        {loading ? (
          <p className="text-sm text-muted-foreground py-4 text-center">Loading lists...</p>
        ) : lists.length === 0 ? (
          <p className="text-sm text-muted-foreground py-4 text-center">
            No lists found. Create a list first.
          </p>
        ) : (
          <form onSubmit={handleSubmit}>
            <div className="space-y-2 mb-4 max-h-60 overflow-y-auto">
              {lists.map((list) => (
                <label
                  key={list.ID}
                  className={cn(
                    'flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-colors',
                    selectedListId === list.ID
                      ? 'border-primary bg-primary/10'
                      : 'border-border hover:bg-secondary/50'
                  )}
                >
                  <input
                    type="radio"
                    name="list"
                    value={list.ID}
                    checked={selectedListId === list.ID}
                    onChange={() => setSelectedListId(list.ID)}
                    className="sr-only"
                  />
                  <List className="w-4 h-4 text-muted-foreground" />
                  <span className="flex-1 text-sm">{list.Name}</span>
                  <span className="text-xs text-muted-foreground">
                    {list.Items?.length || 0} items
                  </span>
                </label>
              ))}
            </div>
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
                disabled={selectedListId === null}
                className={cn(
                  'px-4 py-2 rounded-lg text-sm transition-colors',
                  selectedListId !== null
                    ? 'bg-primary text-primary-foreground hover:bg-primary/90'
                    : 'bg-muted text-muted-foreground cursor-not-allowed'
                )}
              >
                Move
              </button>
            </div>
          </form>
        )}
        {lists.length === 0 && !loading && (
          <div className="flex justify-end">
            <button
              type="button"
              onClick={onCancel}
              className="px-4 py-2 rounded-lg text-sm bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors"
            >
              Close
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

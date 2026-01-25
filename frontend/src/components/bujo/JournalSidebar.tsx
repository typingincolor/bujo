import { useState } from 'react';
import * as Collapsible from '@radix-ui/react-collapsible';
import { ChevronDown, ChevronRight } from 'lucide-react';
import { Entry, ENTRY_SYMBOLS } from '@/types/bujo';
import { OverdueItem } from './OverdueItem';
import { cn } from '@/lib/utils';

interface JournalSidebarProps {
  overdueEntries: Entry[];
  now: Date;
  selectedEntry?: Entry;
  ancestors?: Entry[];
  onSelectEntry?: (entry: Entry) => void;
}

export function JournalSidebar({
  overdueEntries,
  now,
  selectedEntry,
  ancestors = [],
  onSelectEntry,
}: JournalSidebarProps) {
  const [isOverdueOpen, setIsOverdueOpen] = useState(false);

  return (
    <div data-testid="overdue-sidebar" className="flex flex-col h-full">
      {/* Overdue Section */}
      <Collapsible.Root open={isOverdueOpen} onOpenChange={setIsOverdueOpen}>
        <Collapsible.Trigger asChild>
          <button
            className="w-full flex items-center gap-2 px-3 py-2 text-sm font-medium hover:bg-secondary/50 rounded-lg"
            aria-label="Overdue"
          >
            {isOverdueOpen ? (
              <ChevronDown className="h-4 w-4 text-muted-foreground" />
            ) : (
              <ChevronRight className="h-4 w-4 text-muted-foreground" />
            )}
            <span>Overdue ({overdueEntries.length})</span>
          </button>
        </Collapsible.Trigger>

        <Collapsible.Content>
          <div className="px-1 py-1">
            {overdueEntries.length === 0 ? (
              <p className="text-sm text-muted-foreground px-2 py-2">
                No overdue items
              </p>
            ) : (
              overdueEntries.map((entry) => (
                <OverdueItem
                  key={entry.id}
                  entry={entry}
                  now={now}
                  onSelect={onSelectEntry}
                  isSelected={selectedEntry?.id === entry.id}
                />
              ))
            )}
          </div>
        </Collapsible.Content>
      </Collapsible.Root>

      {/* Context Section */}
      <div className="mt-4" data-testid="context-section">
        <div className="flex items-center gap-2 px-3 py-2 text-sm font-medium">
          <span>Context</span>
        </div>

        <div className="px-3 py-2">
          {!selectedEntry ? (
            <p className="text-sm text-muted-foreground">No entry selected</p>
          ) : ancestors.length === 0 ? (
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-sm">
                <span className="text-muted-foreground">
                  {ENTRY_SYMBOLS[selectedEntry.type]}
                </span>
                <span>{selectedEntry.content}</span>
              </div>
              <p className="text-sm text-muted-foreground mt-2">No context</p>
            </div>
          ) : (
            <div className="space-y-1">
              {/* Ancestors in order from root to parent */}
              {[...ancestors].reverse().map((ancestor, index) => (
                <div
                  key={ancestor.id}
                  className={cn(
                    'flex items-center gap-2 text-sm',
                    index > 0 && 'ml-3'
                  )}
                >
                  <span className="text-muted-foreground">
                    {ENTRY_SYMBOLS[ancestor.type]}
                  </span>
                  <span className="text-muted-foreground">{ancestor.content}</span>
                </div>
              ))}
              {/* Selected entry */}
              <div
                className={cn(
                  'flex items-center gap-2 text-sm',
                  ancestors.length > 0 && 'ml-3'
                )}
              >
                <span className="text-muted-foreground">
                  {ENTRY_SYMBOLS[selectedEntry.type]}
                </span>
                <span className="font-medium">{selectedEntry.content}</span>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

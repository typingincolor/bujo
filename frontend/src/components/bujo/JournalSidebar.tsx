import { useEffect, useMemo, useState } from 'react';
import { Entry, ENTRY_SYMBOLS, PRIORITY_SYMBOLS } from '@/types/bujo';
import { EntryActionBar } from './EntryActions/EntryActionBar';
import { cn } from '@/lib/utils';
import { calculateAttentionScore } from '@/lib/attentionScore';

interface EntryCallbacks {
  onCancel?: () => void;
  onMigrate?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
  onCyclePriority?: () => void;
  onMoveToList?: () => void;
}

interface OverdueEntryItemProps {
  entry: Entry;
  now: Date;
  isSelected: boolean;
  onSelect?: () => void;
  callbacks: EntryCallbacks;
}

function OverdueEntryItem({ entry, now, isSelected, onSelect, callbacks }: OverdueEntryItemProps) {
  const [isHovered, setIsHovered] = useState(false);
  const attentionResult = calculateAttentionScore(entry, now);
  const symbol = ENTRY_SYMBOLS[entry.type];
  const prioritySymbol = PRIORITY_SYMBOLS[entry.priority];

  // Clear hover state when keyboard navigation occurs
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'ArrowUp' || e.key === 'ArrowDown' || e.key === 'j' || e.key === 'k') {
        setIsHovered(false);
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  return (
    <div
      className={cn(
        'group px-2 py-1.5 rounded-lg text-sm transition-colors',
        !isSelected && isHovered && 'bg-secondary/50',
        isSelected && 'bg-primary/10 ring-1 ring-primary/30'
      )}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <button
        onClick={onSelect}
        className="flex items-center gap-2 text-left min-w-0 w-full"
      >
        <span data-testid="entry-symbol" className="text-muted-foreground flex-shrink-0">
          {symbol}
        </span>

        {prioritySymbol && (
          <span
            data-testid="priority-indicator"
            className="text-orange-500 font-medium flex-shrink-0"
          >
            {prioritySymbol}
          </span>
        )}

        <span className="flex-1 truncate">{entry.content}</span>

        <span
          data-testid="attention-badge"
          className={cn(
            'px-1.5 py-0.5 rounded text-xs font-medium text-white flex-shrink-0',
            attentionResult.score >= 80 ? 'bg-red-500' :
            attentionResult.score >= 50 ? 'bg-orange-500' : 'bg-yellow-500'
          )}
        >
          {attentionResult.score}
        </span>
      </button>

      {/* Action bar below entry - shown on hover */}
      <div
        className={cn(
          'grid transition-all duration-150 ease-out grid-rows-[0fr]',
          isHovered && 'grid-rows-[1fr]'
        )}
      >
        <div className="overflow-hidden">
          <div
            className="pt-1"
            style={{ paddingLeft: 'calc(0.5rem + 0.5rem + 1ch)' }}
          >
            <EntryActionBar
              entry={entry}
              callbacks={callbacks}
              variant="always-visible"
              size="sm"
            />
          </div>
        </div>
      </div>
    </div>
  );
}

export interface JournalSidebarCallbacks {
  onMarkDone?: (entry: Entry) => void;
  onMigrate?: (entry: Entry) => void;
  onEdit?: (entry: Entry) => void;
  onDelete?: (entry: Entry) => void;
  onCyclePriority?: (entry: Entry) => void;
  onMoveToList?: (entry: Entry) => void;
}

interface JournalSidebarProps {
  overdueEntries?: Entry[];
  now?: Date;
  selectedEntry?: Entry;
  contextTree?: Entry[];
  onSelectEntry?: (entry: Entry) => void;
  callbacks?: JournalSidebarCallbacks;
}

interface TreeNode {
  entry: Entry;
  children: TreeNode[];
}

function buildTree(entries: Entry[]): TreeNode[] {
  if (entries.length === 0) return [];

  const entryMap = new Map<number, Entry>();
  const childrenMap = new Map<number | null, Entry[]>();

  for (const entry of entries) {
    entryMap.set(entry.id, entry);
    const parentId = entry.parentId;
    if (!childrenMap.has(parentId)) {
      childrenMap.set(parentId, []);
    }
    childrenMap.get(parentId)!.push(entry);
  }

  function buildNode(entry: Entry): TreeNode {
    const children = childrenMap.get(entry.id) || [];
    return {
      entry,
      children: children.map(buildNode),
    };
  }

  const roots = childrenMap.get(null) || [];
  return roots.map(buildNode);
}

interface ContextTreeProps {
  nodes: TreeNode[];
  selectedEntryId?: number;
  depth?: number;
}

function ContextTree({ nodes, selectedEntryId, depth = 0 }: ContextTreeProps) {
  return (
    <>
      {nodes.map((node) => (
        <div key={node.entry.id}>
          <div
            className={cn(
              'flex items-center gap-2 text-sm py-0.5 font-mono',
              node.entry.id === selectedEntryId
                ? 'font-medium'
                : 'text-muted-foreground'
            )}
            style={{ paddingLeft: `${depth * 12}px` }}
          >
            <span className="text-muted-foreground">
              {ENTRY_SYMBOLS[node.entry.type]}
            </span>
            <span className={cn(
              'truncate',
              node.entry.id === selectedEntryId && 'text-foreground'
            )}>
              {node.entry.content}
            </span>
          </div>
          {node.children.length > 0 && (
            <ContextTree
              nodes={node.children}
              selectedEntryId={selectedEntryId}
              depth={depth + 1}
            />
          )}
        </div>
      ))}
    </>
  );
}

export function JournalSidebar({
  overdueEntries = [],
  now = new Date(),
  selectedEntry,
  contextTree = [],
  onSelectEntry,
  callbacks = {},
}: JournalSidebarProps) {
  const treeNodes = useMemo(() => buildTree(contextTree), [contextTree]);

  // Filter to only show task entries (not notes, events, questions, etc.)
  const taskEntries = useMemo(
    () => overdueEntries.filter((e) => e.type === 'task'),
    [overdueEntries]
  );

  const createEntryCallbacks = (entry: Entry) => ({
    onCancel: callbacks.onMarkDone ? () => callbacks.onMarkDone!(entry) : undefined,
    onMigrate: callbacks.onMigrate ? () => callbacks.onMigrate!(entry) : undefined,
    onEdit: callbacks.onEdit ? () => callbacks.onEdit!(entry) : undefined,
    onDelete: callbacks.onDelete ? () => callbacks.onDelete!(entry) : undefined,
    onCyclePriority: callbacks.onCyclePriority ? () => callbacks.onCyclePriority!(entry) : undefined,
    onMoveToList: callbacks.onMoveToList ? () => callbacks.onMoveToList!(entry) : undefined,
  });

  return (
    <div data-testid="overdue-sidebar" className="flex flex-col h-full">
      {/* Pending Tasks Section */}
      <div>
        <div className="flex items-center gap-2 px-3 py-2 text-sm font-medium">
          <span>Pending Tasks ({taskEntries.length})</span>
        </div>

        <div className="px-1 py-1 max-h-80 overflow-y-auto">
          {taskEntries.length === 0 ? (
            <p className="text-sm text-muted-foreground px-2 py-2">
              No pending tasks
            </p>
          ) : (
            taskEntries.map((entry) => (
              <OverdueEntryItem
                key={entry.id}
                entry={entry}
                now={now}
                isSelected={selectedEntry?.id === entry.id}
                onSelect={() => onSelectEntry?.(entry)}
                callbacks={createEntryCallbacks(entry)}
              />
            ))
          )}
        </div>
      </div>

      {/* Divider */}
      <hr className="my-4 border-border" />

      {/* Context Section */}
      <div data-testid="context-section">
        <div className="flex items-center gap-2 px-3 py-2 text-sm font-medium">
          <span>Context</span>
        </div>

        <div className="px-3 py-2">
          {!selectedEntry ? (
            <p className="text-sm text-muted-foreground">No entry selected</p>
          ) : selectedEntry.parentId === null ? (
            <p className="text-sm text-muted-foreground">No context</p>
          ) : (
            <ContextTree
              nodes={treeNodes}
              selectedEntryId={selectedEntry.id}
            />
          )}
        </div>
      </div>
    </div>
  );
}

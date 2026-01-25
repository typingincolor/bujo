import { useState, useMemo } from 'react';
import * as Collapsible from '@radix-ui/react-collapsible';
import { ChevronDown, ChevronRight } from 'lucide-react';
import { Entry, ENTRY_SYMBOLS } from '@/types/bujo';
import { OverdueItem } from './OverdueItem';
import { cn } from '@/lib/utils';

interface JournalSidebarProps {
  overdueEntries: Entry[];
  now: Date;
  selectedEntry?: Entry;
  contextTree?: Entry[];
  onSelectEntry?: (entry: Entry) => void;
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
  overdueEntries,
  now,
  selectedEntry,
  contextTree = [],
  onSelectEntry,
}: JournalSidebarProps) {
  const [isOverdueOpen, setIsOverdueOpen] = useState(true);

  const treeNodes = useMemo(() => buildTree(contextTree), [contextTree]);

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
          <div className="px-1 py-1 max-h-80 overflow-y-auto">
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
          ) : treeNodes.length === 0 ? (
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

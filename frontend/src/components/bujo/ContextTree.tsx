import { TreeNode } from '@/lib/buildTree';
import { ENTRY_SYMBOLS } from '@/types/bujo';
import { cn } from '@/lib/utils';

interface ContextTreeProps {
  nodes: TreeNode[];
  selectedEntryId?: number;
  depth?: number;
}

/**
 * Renders a hierarchical tree of entries showing parent-child relationships.
 * Used in context panels to display entry ancestry and context.
 *
 * @param nodes - Array of TreeNode objects representing the tree structure
 * @param selectedEntryId - ID of the currently selected entry (highlighted)
 * @param depth - Current nesting depth for indentation (internal use)
 */
export function ContextTree({ nodes, selectedEntryId, depth = 0 }: ContextTreeProps) {
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
            <ContextTree nodes={node.children} selectedEntryId={selectedEntryId} depth={depth + 1} />
          )}
        </div>
      ))}
    </>
  );
}

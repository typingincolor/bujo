import { Entry } from '@/types/bujo';

export interface TreeNode {
  entry: Entry;
  children: TreeNode[];
}

/**
 * Builds a hierarchical tree structure from a flat list of entries.
 * Entries are organized by their parentId relationships.
 *
 * @param entries - Flat array of entries with parent-child relationships
 * @returns Array of root-level TreeNode objects with nested children
 */
export function buildTree(entries: Entry[]): TreeNode[] {
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

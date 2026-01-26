import { useState, useMemo } from 'react';
import { DayEntries, Entry } from '@/types/bujo';
import { DayBox } from './DayBox';
import { WeekendBox } from './WeekendBox';
import { filterWeekEntries, flattenEntries } from '@/lib/weekView';
import { format, parseISO } from 'date-fns';
import { ActionCallbacks } from './EntryActions/types';

export interface WeekViewCallbacks {
  onMarkDone?: (entry: Entry) => void;
  onMigrate?: (entry: Entry) => void;
  onEdit?: (entry: Entry) => void;
  onDelete?: (entry: Entry) => void;
  onCyclePriority?: (entry: Entry) => void;
  onMoveToList?: (entry: Entry) => void;
}

interface WeekViewProps {
  days: DayEntries[];
  callbacks?: WeekViewCallbacks;
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

function ContextTree({ nodes, selectedEntryId, depth = 0 }: { nodes: TreeNode[]; selectedEntryId?: number; depth?: number }) {
  return (
    <>
      {nodes.map((node) => (
        <div key={node.entry.id}>
          <div
            className="flex items-center gap-2 text-sm py-0.5 font-mono"
            style={{ paddingLeft: `${depth * 12}px` }}
          >
            <span className={node.entry.id === selectedEntryId ? 'text-foreground' : 'text-muted-foreground'}>
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

export function WeekView({ days, callbacks = {} }: WeekViewProps) {
  const [selectedEntry, setSelectedEntry] = useState<Entry | undefined>();

  const dayNames = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri'];

  const createEntryCallbacks = (entry: Entry): ActionCallbacks => ({
    onCancel: callbacks.onMarkDone ? () => callbacks.onMarkDone!(entry) : undefined,
    onMigrate: callbacks.onMigrate ? () => callbacks.onMigrate!(entry) : undefined,
    onEdit: callbacks.onEdit ? () => callbacks.onEdit!(entry) : undefined,
    onDelete: callbacks.onDelete ? () => callbacks.onDelete!(entry) : undefined,
    onCyclePriority: callbacks.onCyclePriority ? () => callbacks.onCyclePriority!(entry) : undefined,
    onMoveToList: callbacks.onMoveToList ? () => callbacks.onMoveToList!(entry) : undefined,
  });

  const weekDays = days.slice(0, 5).map((day, index) => ({
    ...day,
    dayName: dayNames[index],
    dayNumber: parseISO(day.date).getDate(),
  }));

  const saturday = days[5];
  const sunday = days[6];

  const filteredWeekDays = weekDays.map(day => ({
    ...day,
    entries: filterWeekEntries(day.entries),
  }));

  const filteredSaturday = saturday ? filterWeekEntries(saturday.entries) : [];
  const filteredSunday = sunday ? filterWeekEntries(sunday.entries) : [];

  const startDate = days[0] ? parseISO(days[0].date) : new Date();
  const endDate = days[6] ? parseISO(days[6].date) : new Date();
  const dateRange = `${format(startDate, 'MMM d')} – ${format(endDate, 'MMM d, yyyy')}`;

  const allEntries = days.flatMap(day => flattenEntries(day.entries));

  const contextTree = useMemo(() => buildTree(allEntries), [allEntries]);

  return (
    <div className="flex h-full gap-4">
      <div className="flex-1 overflow-y-auto">
        <div className="mb-4">
          <h2 className="text-lg font-semibold">Weekly Review</h2>
          <p className="text-sm text-muted-foreground">{dateRange}</p>
        </div>

        <div className="grid grid-cols-3 gap-4">
          {filteredWeekDays.map((day) => (
            <DayBox
              key={day.date}
              dayNumber={day.dayNumber}
              dayName={day.dayName}
              entries={day.entries}
              selectedEntry={selectedEntry}
              onSelectEntry={setSelectedEntry}
              createEntryCallbacks={createEntryCallbacks}
            />
          ))}

          {saturday && sunday && (
            <WeekendBox
              startDay={parseISO(saturday.date).getDate()}
              saturdayEntries={filteredSaturday}
              sundayEntries={filteredSunday}
              selectedEntry={selectedEntry}
              onSelectEntry={setSelectedEntry}
              createEntryCallbacks={createEntryCallbacks}
            />
          )}
        </div>
      </div>

      <div className="w-96 border-l border-border pl-4 overflow-y-auto">
        <div className="mb-3">
          <h3 className="text-sm font-medium">Context</h3>
        </div>

        {!selectedEntry ? (
          <p className="text-sm text-muted-foreground">No entry selected</p>
        ) : selectedEntry.parentId === null ? (
          <p className="text-sm text-muted-foreground">No context</p>
        ) : (
          <ContextTree nodes={contextTree} selectedEntryId={selectedEntry.id} />
        )}
      </div>
    </div>
  );
}

import { Entry } from '@/types/bujo';

export function flattenEntries(entries: Entry[]): Entry[] {
  const flattened: Entry[] = [];

  for (const entry of entries) {
    flattened.push(entry);
    if (entry.children && entry.children.length > 0) {
      flattened.push(...flattenEntries(entry.children));
    }
  }

  return flattened;
}

export function filterWeekEntries(entries: Entry[]): Entry[] {
  const flattened = flattenEntries(entries);

  return flattened.filter(entry =>
    entry.type === 'event' || entry.priority !== 'none'
  );
}

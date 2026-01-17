export type EntryType = 'task' | 'note' | 'event' | 'done' | 'migrated' | 'cancelled' | 'question' | 'answered' | 'answer';

export type Priority = 'none' | 'low' | 'medium' | 'high';

export interface Entry {
  id: number;
  content: string;
  type: EntryType;
  priority: Priority;
  parentId: number | null;
  loggedDate: string;
  children?: Entry[];
  collapsed?: boolean;
}

export interface DayEntries {
  date: string;
  location?: string;
  mood?: string;
  weather?: string;
  entries: Entry[];
}

export interface Habit {
  id: number;
  name: string;
  streak: number;
  completionRate: number;
  goal?: number;
  history: boolean[]; // last 7 or 30 days
  todayLogged: boolean;
  todayCount: number;
}

export interface ListItem {
  id: number;
  content: string;
  type: EntryType;
  done: boolean;
}

export interface BujoList {
  id: number;
  name: string;
  items: ListItem[];
  doneCount: number;
  totalCount: number;
}

export interface Goal {
  id: number;
  content: string;
  month: string; // YYYY-MM
  completed: boolean;
}

export const ENTRY_SYMBOLS: Record<EntryType, string> = {
  task: '.',
  note: '-',
  event: 'o',
  done: 'x',
  migrated: '>',
  cancelled: 'X',
  question: '?',
  answered: '★',
  answer: '↳',
};

export const PRIORITY_SYMBOLS: Record<Priority, string> = {
  none: '',
  low: '!',
  medium: '!!',
  high: '!!!',
};

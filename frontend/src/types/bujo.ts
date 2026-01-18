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

export interface HabitDayStatus {
  date: string;
  completed: boolean;
  count: number;
}

export interface Habit {
  id: number;
  name: string;
  streak: number;
  completionRate: number;
  goal?: number;
  dayHistory: HabitDayStatus[];
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

export type GoalStatus = 'active' | 'done' | 'migrated';

export interface Goal {
  id: number;
  content: string;
  month: string; // YYYY-MM
  status: GoalStatus;
  migratedTo?: string; // YYYY-MM if migrated
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

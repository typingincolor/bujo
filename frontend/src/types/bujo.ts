export type EntryType = 'task' | 'note' | 'event' | 'done' | 'migrated' | 'cancelled' | 'question' | 'answered' | 'answer' | 'movedToList';

export type Priority = 'none' | 'low' | 'medium' | 'high';

export type ActionType = 'done' | 'cancel' | 'priority' | 'migrate';

export interface Entry {
  id: number;
  content: string;
  type: EntryType;
  priority: Priority;
  parentId: number | null;
  loggedDate: string;
  scheduledDate?: string;
  migrationCount?: number;
  tags?: string[];
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
  goalPerWeek?: number;
  goalPerMonth?: number;
  weeklyProgress?: number;
  monthlyProgress?: number;
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

export type GoalStatus = 'active' | 'done' | 'migrated' | 'cancelled';

export interface Goal {
  id: number;
  content: string;
  month: string; // YYYY-MM
  status: GoalStatus;
  migratedTo?: string; // YYYY-MM if migrated
}

export const ENTRY_SYMBOLS: Record<EntryType, string> = {
  task: '•',
  note: '–',
  event: '⚬',
  done: '✓',
  migrated: '→',
  cancelled: '✗',
  question: '?',
  answered: '★',
  answer: '↳',
  movedToList: '^',
};

export const PRIORITY_SYMBOLS: Record<Priority, string> = {
  none: '',
  low: '!',
  medium: '!!',
  high: '!!!',
};

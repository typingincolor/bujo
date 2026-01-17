import { useState, useRef, useEffect } from 'react';
import { Habit } from '@/types/bujo';
import { cn } from '@/lib/utils';
import { Flame, Check, Plus, X } from 'lucide-react';
import { LogHabit, CreateHabit } from '@/wailsjs/go/wails/App';

interface HabitTrackerProps {
  habits: Habit[];
  onHabitChanged?: () => void;
}

interface HabitRowProps {
  habit: Habit;
  onLogHabit: (habitId: number) => void;
}

const DAY_LABELS = ['S', 'M', 'T', 'W', 'T', 'F', 'S']
const DAYS_TO_SHOW = 7

function getRecentDays(): string[] {
  const days: string[] = []
  for (let i = DAYS_TO_SHOW - 1; i >= 0; i--) {
    const d = new Date()
    d.setDate(d.getDate() - i)
    days.push(DAY_LABELS[d.getDay()])
  }
  return days
}

function HabitRow({ habit, onLogHabit }: HabitRowProps) {
  const recentDays = getRecentDays();

  const handleLog = () => {
    onLogHabit(habit.id);
  };

  return (
    <div className="flex items-center gap-4 py-3 px-4 rounded-lg bg-card hover:bg-secondary/30 transition-colors group animate-slide-in">
      {/* Habit name and streak */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="font-medium text-sm truncate">{habit.name}</span>
          {habit.streak > 0 && (
            <span className={cn(
              'flex items-center gap-0.5 text-xs font-semibold text-bujo-streak',
              habit.streak >= 7 && 'animate-streak-glow'
            )}>
              <Flame className="w-3.5 h-3.5" />
              {habit.streak}
            </span>
          )}
        </div>
        <div className="flex items-center gap-1.5 text-xs text-muted-foreground mt-0.5">
          <span>{habit.completionRate}% completion</span>
          {habit.goal && <span>• Goal: {habit.goal}</span>}
        </div>
      </div>

      {/* Sparkline */}
      <div className="flex items-center gap-1">
        {habit.history.map((completed, i) => (
          <div key={i} className="flex flex-col items-center">
            <div
              className={cn(
                'w-5 h-5 rounded-full flex items-center justify-center transition-all',
                completed ? 'bg-bujo-habit-fill' : 'bg-bujo-habit-empty',
                i === habit.history.length - 1 && 'ring-2 ring-primary/30'
              )}
            >
              {completed && (
                <Check className="w-3 h-3 text-primary-foreground" />
              )}
            </div>
            <span className="text-[10px] text-muted-foreground mt-0.5">
              {recentDays[i]}
            </span>
          </div>
        ))}
      </div>

      {/* Today status */}
      <button
        onClick={handleLog}
        className={cn(
          'px-3 py-1.5 rounded-md text-xs font-medium transition-all',
          habit.todayLogged
            ? 'bg-bujo-done text-primary-foreground hover:bg-bujo-done/90'
            : 'bg-primary text-primary-foreground hover:bg-primary/90'
        )}
      >
        {habit.todayLogged ? `+1 (${habit.todayCount})` : 'Log'}
      </button>
    </div>
  );
}

export function HabitTracker({ habits, onHabitChanged }: HabitTrackerProps) {
  const [isAddingHabit, setIsAddingHabit] = useState(false);
  const [newHabitName, setNewHabitName] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isAddingHabit) {
      inputRef.current?.focus();
    }
  }, [isAddingHabit]);

  const handleLogHabit = async (habitId: number) => {
    try {
      await LogHabit(habitId, 1);
      onHabitChanged?.();
    } catch (error) {
      console.error('Failed to log habit:', error);
    }
  };

  const handleCreateHabit = async () => {
    const trimmedName = newHabitName.trim();
    if (!trimmedName) return;

    try {
      await CreateHabit(trimmedName);
      setNewHabitName('');
      onHabitChanged?.();
    } catch (error) {
      console.error('Failed to create habit:', error);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleCreateHabit();
    } else if (e.key === 'Escape') {
      setIsAddingHabit(false);
      setNewHabitName('');
    }
  };

  const handleCancel = () => {
    setIsAddingHabit(false);
    setNewHabitName('');
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2 mb-4">
        <Flame className="w-5 h-5 text-bujo-streak" />
        <h2 className="font-display text-xl font-semibold">Habit Tracker</h2>
        <button
          onClick={() => setIsAddingHabit(true)}
          className="ml-auto px-2 py-1 text-xs rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors flex items-center gap-1"
          aria-label="Add habit"
        >
          <Plus className="w-3 h-3" />
          Add Habit
        </button>
      </div>

      {isAddingHabit && (
        <div className="flex items-center gap-2 py-2 px-4 rounded-lg bg-card border border-border animate-fade-in">
          <input
            ref={inputRef}
            type="text"
            value={newHabitName}
            onChange={(e) => setNewHabitName(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Habit name"
            className="flex-1 px-2 py-1.5 text-sm rounded-md border border-border bg-background focus:outline-none focus:ring-2 focus:ring-primary/50"
          />
          <button
            onClick={handleCancel}
            className="p-1.5 rounded-md hover:bg-secondary transition-colors"
            aria-label="Cancel"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      <div className="space-y-1">
        {habits.map((habit) => (
          <HabitRow key={habit.id} habit={habit} onLogHabit={handleLogHabit} />
        ))}
      </div>
    </div>
  );
}

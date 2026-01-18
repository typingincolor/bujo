import { useState, useRef, useEffect } from 'react';
import { Habit } from '@/types/bujo';
import { cn } from '@/lib/utils';
import { Flame, Check, Plus, X, Trash2, Undo2, Target, ChevronDown } from 'lucide-react';
import { LogHabit, CreateHabit, DeleteHabit, UndoHabitLog, SetHabitGoal, LogHabitForDate } from '@/wailsjs/go/wails/App';
import { ConfirmDialog } from './ConfirmDialog';

type PeriodView = 'week' | 'month' | 'quarter';

interface HabitTrackerProps {
  habits: Habit[];
  onHabitChanged?: () => void;
  onPeriodChange?: (period: PeriodView) => void;
}

interface HabitRowProps {
  habit: Habit;
  onLogHabit: (habitId: number) => void;
  onDeleteHabit: (habitId: number) => void;
  onUndoLog: (habitId: number) => void;
  onSetGoal: (habitId: number) => void;
  onLogForDate: (habitId: number, dayIndex: number) => void;
  settingGoalFor: number | null;
  onGoalSubmit: (habitId: number, goal: number) => void;
  onGoalCancel: () => void;
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

function HabitRow({
  habit,
  onLogHabit,
  onDeleteHabit,
  onUndoLog,
  onSetGoal,
  onLogForDate,
  settingGoalFor,
  onGoalSubmit,
  onGoalCancel
}: HabitRowProps) {
  const recentDays = getRecentDays();
  const [goalInput, setGoalInput] = useState('');

  const handleLog = () => {
    onLogHabit(habit.id);
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDeleteHabit(habit.id);
  };

  const handleUndo = (e: React.MouseEvent) => {
    e.stopPropagation();
    onUndoLog(habit.id);
  };

  const handleGoalKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      const goal = parseInt(goalInput, 10);
      if (!isNaN(goal) && goal > 0) {
        onGoalSubmit(habit.id, goal);
        setGoalInput('');
      }
    } else if (e.key === 'Escape') {
      onGoalCancel();
      setGoalInput('');
    }
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
        {habit.history.map((completed, i) => {
          const isToday = i === habit.history.length - 1;
          return (
            <div key={i} className="flex flex-col items-center">
              <button
                onClick={() => !isToday && onLogForDate(habit.id, i)}
                aria-label={isToday ? undefined : `Log for ${recentDays[i]}`}
                disabled={isToday}
                className={cn(
                  'w-5 h-5 rounded-full flex items-center justify-center transition-all',
                  completed ? 'bg-bujo-habit-fill' : 'bg-bujo-habit-empty',
                  isToday && 'ring-2 ring-primary/30',
                  !isToday && !completed && 'hover:bg-bujo-habit-empty/80 cursor-pointer'
                )}
              >
                {completed && (
                  <Check className="w-3 h-3 text-primary-foreground" />
                )}
              </button>
              <span className="text-[10px] text-muted-foreground mt-0.5">
                {recentDays[i]}
              </span>
            </div>
          );
        })}
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

      {/* Undo button - only show when habit is logged today */}
      {habit.todayLogged && (
        <button
          onClick={handleUndo}
          title="Undo last log"
          className="p-1.5 rounded-md text-muted-foreground hover:text-warning hover:bg-warning/10 transition-colors"
        >
          <Undo2 className="w-4 h-4" />
        </button>
      )}

      {/* Goal button or input */}
      {settingGoalFor === habit.id ? (
        <input
          type="number"
          min="1"
          value={goalInput}
          onChange={(e) => setGoalInput(e.target.value)}
          onKeyDown={handleGoalKeyDown}
          placeholder="Daily goal"
          autoFocus
          className="w-20 px-2 py-1 text-xs rounded-md border border-border bg-background focus:outline-none focus:ring-2 focus:ring-primary/50"
        />
      ) : (
        <button
          onClick={() => onSetGoal(habit.id)}
          title="Set goal"
          className="p-1.5 rounded-md text-muted-foreground hover:text-primary hover:bg-primary/10 transition-colors opacity-0 group-hover:opacity-100"
        >
          <Target className="w-4 h-4" />
        </button>
      )}

      {/* Delete button */}
      <button
        onClick={handleDelete}
        title="Delete habit"
        className="p-1.5 rounded-md text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors opacity-0 group-hover:opacity-100"
      >
        <Trash2 className="w-4 h-4" />
      </button>
    </div>
  );
}

export function HabitTracker({ habits, onHabitChanged, onPeriodChange }: HabitTrackerProps) {
  const [isAddingHabit, setIsAddingHabit] = useState(false);
  const [newHabitName, setNewHabitName] = useState('');
  const [habitToDelete, setHabitToDelete] = useState<Habit | null>(null);
  const [currentPeriod, setCurrentPeriod] = useState<PeriodView>('week');
  const [showPeriodMenu, setShowPeriodMenu] = useState(false);
  const [settingGoalFor, setSettingGoalFor] = useState<number | null>(null);
  const [logForDateInfo, setLogForDateInfo] = useState<{ habitId: number; dayIndex: number; date: Date } | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const periodMenuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (isAddingHabit) {
      inputRef.current?.focus();
    }
  }, [isAddingHabit]);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (periodMenuRef.current && !periodMenuRef.current.contains(e.target as Node)) {
        setShowPeriodMenu(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

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

  const handleDeleteHabit = async () => {
    if (!habitToDelete) return;

    try {
      await DeleteHabit(habitToDelete.id);
      setHabitToDelete(null);
      onHabitChanged?.();
    } catch (error) {
      console.error('Failed to delete habit:', error);
    }
  };

  const handleUndoLog = async (habitId: number) => {
    try {
      await UndoHabitLog(habitId);
      onHabitChanged?.();
    } catch (error) {
      console.error('Failed to undo habit log:', error);
    }
  };

  const handleSetGoal = async (habitId: number, goal: number) => {
    try {
      await SetHabitGoal(habitId, goal);
      setSettingGoalFor(null);
      onHabitChanged?.();
    } catch (error) {
      console.error('Failed to set habit goal:', error);
    }
  };

  const handleLogForDate = (habitId: number, dayIndex: number) => {
    const date = new Date();
    const daysAgo = DAYS_TO_SHOW - 1 - dayIndex;
    date.setDate(date.getDate() - daysAgo);
    setLogForDateInfo({ habitId, dayIndex, date });
  };

  const handleConfirmLogForDate = async () => {
    if (!logForDateInfo) return;
    try {
      await LogHabitForDate(logForDateInfo.habitId, 1, logForDateInfo.date.toISOString());
      setLogForDateInfo(null);
      onHabitChanged?.();
    } catch (error) {
      console.error('Failed to log habit for date:', error);
    }
  };

  const handlePeriodChange = (period: PeriodView) => {
    setCurrentPeriod(period);
    setShowPeriodMenu(false);
    onPeriodChange?.(period);
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

  const handleRequestDelete = (habitId: number) => {
    const habit = habits.find(h => h.id === habitId);
    if (habit) {
      setHabitToDelete(habit);
    }
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2 mb-4">
        <Flame className="w-5 h-5 text-bujo-streak" />
        <h2 className="font-display text-xl font-semibold">Habit Tracker</h2>

        {/* Period selector */}
        <div className="relative ml-2" ref={periodMenuRef}>
          <button
            onClick={() => setShowPeriodMenu(!showPeriodMenu)}
            className="px-2 py-1 text-xs rounded-md bg-secondary/50 hover:bg-secondary transition-colors flex items-center gap-1 capitalize"
          >
            {currentPeriod}
            <ChevronDown className="w-3 h-3" />
          </button>
          {showPeriodMenu && (
            <div className="absolute top-full left-0 mt-1 bg-card border border-border rounded-lg shadow-lg z-50">
              <button
                onClick={() => handlePeriodChange('week')}
                className="w-full px-3 py-1.5 text-xs text-left hover:bg-secondary/50 transition-colors rounded-t-lg"
              >
                Week
              </button>
              <button
                onClick={() => handlePeriodChange('month')}
                className="w-full px-3 py-1.5 text-xs text-left hover:bg-secondary/50 transition-colors"
              >
                Month
              </button>
              <button
                onClick={() => handlePeriodChange('quarter')}
                className="w-full px-3 py-1.5 text-xs text-left hover:bg-secondary/50 transition-colors rounded-b-lg"
              >
                Quarter
              </button>
            </div>
          )}
        </div>

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
          <HabitRow
            key={habit.id}
            habit={habit}
            onLogHabit={handleLogHabit}
            onDeleteHabit={handleRequestDelete}
            onUndoLog={handleUndoLog}
            onSetGoal={(id) => setSettingGoalFor(id)}
            onLogForDate={handleLogForDate}
            settingGoalFor={settingGoalFor}
            onGoalSubmit={handleSetGoal}
            onGoalCancel={() => setSettingGoalFor(null)}
          />
        ))}
      </div>

      <ConfirmDialog
        isOpen={!!habitToDelete}
        title="Delete Habit"
        message={`Are you sure you want to delete "${habitToDelete?.name}"? This will also delete all habit logs.`}
        confirmText="Delete"
        variant="destructive"
        onConfirm={handleDeleteHabit}
        onCancel={() => setHabitToDelete(null)}
      />

      <ConfirmDialog
        isOpen={!!logForDateInfo}
        title="Log Habit for Date"
        message={`Log habit for ${logForDateInfo?.date.toLocaleDateString()}?`}
        confirmText="Confirm"
        onConfirm={handleConfirmLogForDate}
        onCancel={() => setLogForDateInfo(null)}
      />
    </div>
  );
}

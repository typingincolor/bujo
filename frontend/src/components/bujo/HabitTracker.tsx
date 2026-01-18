import { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import { Habit, HabitDayStatus } from '@/types/bujo';
import { cn } from '@/lib/utils';
import { Flame, Plus, X, Trash2, Target, ChevronDown } from 'lucide-react';
import { CreateHabit, DeleteHabit, UndoHabitLogForDate, SetHabitGoal, LogHabitForDate } from '@/wailsjs/go/wails/App';
import { ConfirmDialog } from './ConfirmDialog';
import { CalendarNavigation, CalendarGrid, QuarterGrid } from './calendar';
import {
  getWeekDates,
  getMonthCalendar,
  getQuarterMonths,
  navigatePeriod,
  mapDayHistoryToCalendar,
  formatPeriodLabel,
} from '@/lib/calendarUtils';

type PeriodView = 'week' | 'month' | 'quarter';

interface HabitTrackerProps {
  habits: Habit[];
  onHabitChanged?: () => void;
  period?: PeriodView;
  onPeriodChange?: (period: PeriodView) => void;
  anchorDate?: Date;
  onNavigate?: (newAnchor: Date) => void;
}

interface HabitRowProps {
  habit: Habit;
  onLogForDate: (habitId: number, date: string) => void;
  onDecrementForDate: (habitId: number, date: string) => void;
  onDeleteHabit: (habitId: number) => void;
  onSetGoal: (habitId: number) => void;
  settingGoalFor: number | null;
  onGoalSubmit: (habitId: number, goal: number) => void;
  onGoalCancel: () => void;
  currentPeriod: PeriodView;
  anchorDate: Date;
}

const DAY_LABELS = ['S', 'M', 'T', 'W', 'T', 'F', 'S']

function getDayLabel(dateStr: string): string {
  const date = new Date(dateStr + 'T00:00:00');
  return DAY_LABELS[date.getDay()];
}

function formatShortDate(dateStr: string): string {
  const date = new Date(dateStr + 'T00:00:00');
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

function HabitRow({
  habit,
  onLogForDate,
  onDecrementForDate,
  onDeleteHabit,
  onSetGoal,
  settingGoalFor,
  onGoalSubmit,
  onGoalCancel,
  currentPeriod,
  anchorDate,
}: HabitRowProps) {
  const [goalInput, setGoalInput] = useState('');

  const dayHistory = useMemo(
    () => mapDayHistoryToCalendar(habit.dayHistory),
    [habit.dayHistory]
  );

  const handleLog = useCallback(
    (date: string) => {
      onLogForDate(habit.id, date);
    },
    [habit.id, onLogForDate]
  );

  const handleDecrement = useCallback(
    (date: string) => {
      onDecrementForDate(habit.id, date);
    },
    [habit.id, onDecrementForDate]
  );

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDeleteHabit(habit.id);
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

  const renderCalendar = () => {
    switch (currentPeriod) {
      case 'week': {
        const weekDays = getWeekDates(anchorDate);
        return (
          <CalendarGrid
            calendar={[weekDays]}
            dayHistory={dayHistory}
            onLog={handleLog}
            onDecrement={handleDecrement}
            showHeader={false}
          />
        );
      }
      case 'month': {
        const monthCalendar = getMonthCalendar(anchorDate);
        return (
          <CalendarGrid
            calendar={monthCalendar}
            dayHistory={dayHistory}
            onLog={handleLog}
            onDecrement={handleDecrement}
          />
        );
      }
      case 'quarter': {
        const quarterMonths = getQuarterMonths(anchorDate);
        return (
          <QuarterGrid
            quarters={quarterMonths}
            dayHistory={dayHistory}
            onLog={handleLog}
            onDecrement={handleDecrement}
          />
        );
      }
    }
  };

  return (
    <div className="flex items-center gap-4 py-3 px-4 rounded-lg bg-card hover:bg-secondary/30 transition-colors group animate-slide-in">
      {/* Habit name and streak */}
      <div className="flex-shrink-0 min-w-0 w-32">
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
          {habit.goal && (
            <span
              className="flex items-center gap-0.5"
              aria-label={`Daily goal: ${habit.goal}`}
            >
              • <Target className="w-3 h-3" />{habit.goal}
            </span>
          )}
        </div>
      </div>

      {/* Calendar grid */}
      <div className="flex-1 overflow-x-auto">
        {renderCalendar()}
      </div>

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

export function HabitTracker({ habits, onHabitChanged, period, onPeriodChange, anchorDate, onNavigate }: HabitTrackerProps) {
  const [isAddingHabit, setIsAddingHabit] = useState(false);
  const [newHabitName, setNewHabitName] = useState('');
  const [habitToDelete, setHabitToDelete] = useState<Habit | null>(null);
  const [internalPeriod, setInternalPeriod] = useState<PeriodView>('week');
  const [showPeriodMenu, setShowPeriodMenu] = useState(false);
  const [settingGoalFor, setSettingGoalFor] = useState<number | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const periodMenuRef = useRef<HTMLDivElement>(null);

  // Use controlled period if provided, otherwise use internal state
  const currentPeriod = period ?? internalPeriod;

  const effectiveAnchor = useMemo(() => anchorDate ?? new Date(), [anchorDate]);
  const periodLabel = useMemo(() => formatPeriodLabel(effectiveAnchor, currentPeriod), [effectiveAnchor, currentPeriod]);

  const handleNavigatePrev = useCallback(() => {
    const newAnchor = navigatePeriod(effectiveAnchor, currentPeriod, 'prev');
    onNavigate?.(newAnchor);
  }, [effectiveAnchor, currentPeriod, onNavigate]);

  const handleNavigateNext = useCallback(() => {
    const newAnchor = navigatePeriod(effectiveAnchor, currentPeriod, 'next');
    onNavigate?.(newAnchor);
  }, [effectiveAnchor, currentPeriod, onNavigate]);

  const handleNavigateToday = useCallback(() => {
    onNavigate?.(new Date());
  }, [onNavigate]);

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

  const handleLogForDate = useCallback(async (habitId: number, dateStr: string) => {
    try {
      const date = new Date(dateStr + 'T12:00:00');
      await LogHabitForDate(habitId, 1, date.toISOString());
      onHabitChanged?.();
    } catch (error) {
      console.error('Failed to log habit:', error);
    }
  }, [onHabitChanged]);

  const handleDecrementForDate = useCallback(async (habitId: number, dateStr: string) => {
    try {
      const date = new Date(dateStr + 'T12:00:00');
      await UndoHabitLogForDate(habitId, date.toISOString());
      onHabitChanged?.();
    } catch (error) {
      console.error('Failed to undo habit log:', error);
    }
  }, [onHabitChanged]);

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

  const handleSetGoal = async (habitId: number, goal: number) => {
    try {
      await SetHabitGoal(habitId, goal);
      setSettingGoalFor(null);
      onHabitChanged?.();
    } catch (error) {
      console.error('Failed to set habit goal:', error);
    }
  };

  const handlePeriodChange = (newPeriod: PeriodView) => {
    // Update internal state for uncontrolled mode
    setInternalPeriod(newPeriod);
    setShowPeriodMenu(false);
    // Notify parent (required for controlled mode)
    onPeriodChange?.(newPeriod);
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

        {/* Calendar navigation */}
        <CalendarNavigation
          label={periodLabel}
          onPrev={handleNavigatePrev}
          onNext={handleNavigateNext}
        />

        {/* Today button */}
        <button
          onClick={handleNavigateToday}
          className="px-2 py-1 text-xs rounded-md bg-secondary/50 hover:bg-secondary transition-colors"
        >
          Today
        </button>

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

      {/* Shared day header for week view */}
      {currentPeriod === 'week' && habits.length > 0 && (
        <div className="flex items-center gap-4 px-4">
          <div className="flex-shrink-0 w-32" />
          <div className="flex-1">
            <div className="grid grid-cols-7 gap-0.5 mb-1">
              {['S', 'M', 'T', 'W', 'T', 'F', 'S'].map((label, i) => (
                <div
                  key={i}
                  className="w-full flex justify-center text-[10px] text-muted-foreground font-medium"
                >
                  {label}
                </div>
              ))}
            </div>
          </div>
          {/* Spacer for action buttons */}
          <div className="w-[72px]" />
        </div>
      )}

      <div className="space-y-1">
        {habits.map((habit) => (
          <HabitRow
            key={habit.id}
            habit={habit}
            onLogForDate={handleLogForDate}
            onDecrementForDate={handleDecrementForDate}
            onDeleteHabit={handleRequestDelete}
            onSetGoal={(id) => setSettingGoalFor(id)}
            settingGoalFor={settingGoalFor}
            onGoalSubmit={handleSetGoal}
            onGoalCancel={() => setSettingGoalFor(null)}
            currentPeriod={currentPeriod}
            anchorDate={effectiveAnchor}
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
    </div>
  );
}

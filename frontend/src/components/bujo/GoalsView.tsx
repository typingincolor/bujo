import { Goal } from '@/types/bujo';
import { cn } from '@/lib/utils';
import { Target, CheckCircle2, Circle, ChevronLeft, ChevronRight, Plus, X, Trash2, ArrowRight, Pencil } from 'lucide-react';
import { format, parse, addMonths } from 'date-fns';
import { useState, useRef } from 'react';
import { MarkGoalDone, MarkGoalActive, CreateGoal, DeleteGoal, MigrateGoal, UpdateGoal } from '@/wailsjs/go/wails/App';
import { time } from '@/wailsjs/go/models';
import { ConfirmDialog } from './ConfirmDialog';

interface GoalsViewProps {
  goals: Goal[];
  onGoalChanged?: () => void;
  onError?: (message: string) => void;
}

function getMonthLabel(monthStr: string): string {
  const date = parse(monthStr, 'yyyy-MM', new Date());
  return format(date, 'MMMM yyyy');
}

function toWailsTime(date: Date): time.Time {
  return date.toISOString() as unknown as time.Time
}

export function GoalsView({ goals: initialGoals, onGoalChanged, onError }: GoalsViewProps) {
  const [currentMonth, setCurrentMonth] = useState(() => format(new Date(), 'yyyy-MM'));
  const [isAdding, setIsAdding] = useState(false);
  const [newGoalContent, setNewGoalContent] = useState('');
  const [deleteGoal, setDeleteGoal] = useState<Goal | null>(null);
  const [migrateGoal, setMigrateGoal] = useState<Goal | null>(null);
  const [migrateMonth, setMigrateMonth] = useState<string>('');
  const [editingGoal, setEditingGoal] = useState<Goal | null>(null);
  const [editContent, setEditContent] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);
  const editInputRef = useRef<HTMLInputElement>(null);

  const filteredGoals = initialGoals.filter(g => g.month === currentMonth);
  const completedCount = filteredGoals.filter(g => g.status === 'done').length;
  const progress = filteredGoals.length > 0
    ? Math.round((completedCount / filteredGoals.length) * 100)
    : 0;

  const navigateMonth = (delta: number) => {
    const date = parse(currentMonth, 'yyyy-MM', new Date());
    date.setMonth(date.getMonth() + delta);
    setCurrentMonth(format(date, 'yyyy-MM'));
  };

  const handleToggleGoal = async (goal: Goal) => {
    if (goal.status === 'migrated') return;

    try {
      if (goal.status === 'done') {
        await MarkGoalActive(goal.id);
      } else {
        await MarkGoalDone(goal.id);
      }
      onGoalChanged?.();
    } catch (error) {
      console.error('Failed to toggle goal:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to toggle goal');
    }
  };

  const handleStartMigrate = (goal: Goal) => {
    const nextMonth = format(addMonths(parse(currentMonth, 'yyyy-MM', new Date()), 1), 'yyyy-MM');
    setMigrateMonth(nextMonth);
    setMigrateGoal(goal);
  };

  const handleConfirmMigrate = async () => {
    if (!migrateGoal || !migrateMonth) return;

    try {
      const monthDate = parse(migrateMonth, 'yyyy-MM', new Date());
      await MigrateGoal(migrateGoal.id, toWailsTime(monthDate));
      setMigrateGoal(null);
      setMigrateMonth('');
      onGoalChanged?.();
    } catch (error) {
      console.error('Failed to migrate goal:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to migrate goal');
    }
  };

  const handleStartAdding = () => {
    setIsAdding(true);
    setTimeout(() => inputRef.current?.focus(), 0);
  };

  const handleCancelAdding = () => {
    setIsAdding(false);
    setNewGoalContent('');
  };

  const handleCreateGoal = async () => {
    if (!newGoalContent.trim()) return;

    try {
      const monthDate = parse(currentMonth, 'yyyy-MM', new Date());
      await CreateGoal(newGoalContent.trim(), toWailsTime(monthDate));
      setNewGoalContent('');
      onGoalChanged?.();
    } catch (error) {
      console.error('Failed to create goal:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to create goal');
    }
  };

  const handleAddKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleCreateGoal();
    } else if (e.key === 'Escape') {
      handleCancelAdding();
    }
  };

  const handleConfirmDelete = async () => {
    if (!deleteGoal) return;

    try {
      await DeleteGoal(deleteGoal.id);
      setDeleteGoal(null);
      onGoalChanged?.();
    } catch (error) {
      console.error('Failed to delete goal:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to delete goal');
    }
  };

  const handleStartEdit = (goal: Goal) => {
    setEditingGoal(goal);
    setEditContent(goal.content);
    setTimeout(() => editInputRef.current?.focus(), 0);
  };

  const handleCancelEdit = () => {
    setEditingGoal(null);
    setEditContent('');
  };

  const handleSaveEdit = async () => {
    if (!editingGoal || !editContent.trim()) return;

    try {
      await UpdateGoal(editingGoal.id, editContent.trim());
      setEditingGoal(null);
      setEditContent('');
      onGoalChanged?.();
    } catch (error) {
      console.error('Failed to update goal:', error);
      onError?.(error instanceof Error ? error.message : 'Failed to update goal');
    }
  };

  const handleEditKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSaveEdit();
    } else if (e.key === 'Escape') {
      handleCancelEdit();
    }
  };

  return (
    <div className="space-y-4">
      {/* Header with navigation */}
      <div className="flex items-center gap-2">
        <Target className="w-5 h-5 text-primary" />
        <h2 className="font-display text-xl font-semibold flex-1">Monthly Goals</h2>

        <div className="flex items-center gap-2">
          <button
            onClick={() => navigateMonth(-1)}
            className="p-1.5 rounded hover:bg-secondary transition-colors"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>
          <span className="text-sm font-medium min-w-[120px] text-center">
            {getMonthLabel(currentMonth)}
          </span>
          <button
            onClick={() => navigateMonth(1)}
            className="p-1.5 rounded hover:bg-secondary transition-colors"
          >
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>

        {!isAdding && (
          <button
            onClick={handleStartAdding}
            className="flex items-center gap-1 px-3 py-1.5 text-sm rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            <Plus className="w-4 h-4" />
            Add Goal
          </button>
        )}
      </div>

      {/* Add goal input */}
      {isAdding && (
        <div className="flex items-center gap-2 animate-fade-in">
          <input
            ref={inputRef}
            type="text"
            value={newGoalContent}
            onChange={(e) => setNewGoalContent(e.target.value)}
            onKeyDown={handleAddKeyDown}
            placeholder="New goal..."
            className="flex-1 px-3 py-2 text-sm rounded-md border border-border bg-background focus:outline-none focus:ring-2 focus:ring-primary/50"
          />
          <button
            onClick={handleCancelAdding}
            className="p-2 rounded-md hover:bg-secondary transition-colors text-muted-foreground"
          >
            <X className="w-4 h-4" />
            <span className="sr-only">Cancel</span>
          </button>
        </div>
      )}
      
      {/* Progress bar */}
      <div className="flex items-center gap-3">
        <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
          <div
            className="h-full bg-bujo-done transition-all duration-500"
            style={{ width: `${progress}%` }}
          />
        </div>
        <span className="text-sm text-muted-foreground">
          {completedCount}/{filteredGoals.length}
        </span>
      </div>
      
      {/* Goals list */}
      <div className="space-y-2">
        {filteredGoals.length > 0 ? (
          filteredGoals.map((goal) => (
            <div
              key={goal.id}
              onClick={() => handleToggleGoal(goal)}
              className={cn(
                'flex items-center gap-3 p-3 rounded-lg border border-border',
                'bg-card hover:bg-secondary/30 transition-colors cursor-pointer group animate-fade-in',
                goal.status === 'migrated' && 'opacity-60 cursor-default'
              )}
            >
              {goal.status === 'done' ? (
                <CheckCircle2 className="w-5 h-5 text-bujo-done flex-shrink-0" />
              ) : goal.status === 'migrated' ? (
                <ArrowRight className="w-5 h-5 text-muted-foreground flex-shrink-0" />
              ) : (
                <Circle className="w-5 h-5 text-muted-foreground flex-shrink-0" />
              )}
              {editingGoal?.id === goal.id ? (
                <input
                  ref={editInputRef}
                  type="text"
                  value={editContent}
                  onChange={(e) => setEditContent(e.target.value)}
                  onKeyDown={handleEditKeyDown}
                  onClick={(e) => e.stopPropagation()}
                  className="flex-1 px-2 py-1 text-sm rounded border border-border bg-background focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
              ) : (
                <span className={cn(
                  'flex-1 text-sm',
                  goal.status === 'done' && 'line-through text-muted-foreground',
                  goal.status === 'migrated' && 'text-muted-foreground'
                )}>
                  {goal.content}
                  {goal.status === 'migrated' && goal.migratedTo && (
                    <span className="ml-2 text-xs">
                      (Migrated to {getMonthLabel(goal.migratedTo)})
                    </span>
                  )}
                </span>
              )}
              {goal.status !== 'migrated' && (
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    handleStartMigrate(goal);
                  }}
                  title="Migrate goal"
                  className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary transition-colors opacity-0 group-hover:opacity-100"
                >
                  <ArrowRight className="w-4 h-4" />
                </button>
              )}
              {goal.status !== 'migrated' && (
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    handleStartEdit(goal);
                  }}
                  title="Edit goal"
                  className="p-1 rounded hover:bg-primary/20 text-muted-foreground hover:text-primary transition-colors opacity-0 group-hover:opacity-100"
                >
                  <Pencil className="w-4 h-4" />
                </button>
              )}
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  setDeleteGoal(goal);
                }}
                title="Delete goal"
                className="p-1 rounded hover:bg-destructive/20 text-muted-foreground hover:text-destructive transition-colors opacity-0 group-hover:opacity-100"
              >
                <Trash2 className="w-4 h-4" />
              </button>
              <span className="text-xs text-muted-foreground opacity-0 group-hover:opacity-100">
                #{goal.id}
              </span>
            </div>
          ))
        ) : (
          <p className="text-sm text-muted-foreground italic py-6 text-center">
            No goals for {getMonthLabel(currentMonth)}. Add some!
          </p>
        )}
      </div>

      {/* Delete confirmation dialog */}
      <ConfirmDialog
        isOpen={deleteGoal !== null}
        title="Delete Goal"
        message={`Are you sure you want to delete "${deleteGoal?.content}"?`}
        confirmText="Delete"
        variant="destructive"
        onConfirm={handleConfirmDelete}
        onCancel={() => setDeleteGoal(null)}
      />

      {/* Migrate goal dialog */}
      <ConfirmDialog
        isOpen={migrateGoal !== null}
        title="Migrate Goal"
        message={`Migrate "${migrateGoal?.content}" to ${migrateMonth ? getMonthLabel(migrateMonth) : ''}?`}
        confirmText="Migrate"
        onConfirm={handleConfirmMigrate}
        onCancel={() => {
          setMigrateGoal(null);
          setMigrateMonth('');
        }}
      />
    </div>
  );
}

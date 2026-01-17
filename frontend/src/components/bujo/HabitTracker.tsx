import { Habit } from '@/types/bujo';
import { cn } from '@/lib/utils';
import { Flame, Check } from 'lucide-react';

interface HabitTrackerProps {
  habits: Habit[];
}

interface HabitRowProps {
  habit: Habit;
}

const DAYS = ['S', 'M', 'T', 'W', 'T', 'F', 'S'];

function getRecentDays(): string[] {
  const days: string[] = [];
  for (let i = 6; i >= 0; i--) {
    const d = new Date();
    d.setDate(d.getDate() - i);
    days.push(DAYS[d.getDay()]);
  }
  return days;
}

function HabitRow({ habit }: HabitRowProps) {
  const recentDays = getRecentDays();
  
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
        className={cn(
          'px-3 py-1.5 rounded-md text-xs font-medium transition-all',
          habit.todayLogged
            ? 'bg-bujo-done text-primary-foreground'
            : 'bg-primary text-primary-foreground hover:bg-primary/90'
        )}
      >
        {habit.todayLogged ? 'Done ✓' : 'Log'}
      </button>
    </div>
  );
}

export function HabitTracker({ habits }: HabitTrackerProps) {
  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2 mb-4">
        <Flame className="w-5 h-5 text-bujo-streak" />
        <h2 className="font-display text-xl font-semibold">Habit Tracker</h2>
      </div>
      
      <div className="space-y-1">
        {habits.map((habit) => (
          <HabitRow key={habit.id} habit={habit} />
        ))}
      </div>
    </div>
  );
}

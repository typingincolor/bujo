import { cn } from '@/lib/utils';
import { Calendar, CheckCircle2, AlertCircle } from 'lucide-react';
import { format } from 'date-fns';
import { DayEntries, Habit, Goal } from '@/types/bujo';

interface QuickStatsProps {
  days: DayEntries[];
  habits: Habit[];
  goals: Goal[];
  overdueCount: number;
}

export function QuickStats({ days, habits, goals, overdueCount }: QuickStatsProps) {
  const today = days[0];
  const todayEntries = today?.entries || [];
  const tasksDone = todayEntries.filter(e => e.type === 'done').length;
  
  const habitsLoggedToday = habits.filter(h => h.todayLogged).length;
  const longestStreak = Math.max(...habits.map(h => h.streak), 0);
  
  const currentMonth = format(new Date(), 'yyyy-MM');
  const monthGoals = goals.filter(g => g.month === currentMonth);
  const goalsCompleted = monthGoals.filter(g => g.status === 'done').length;
  
  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
      <StatCard
        icon={CheckCircle2}
        label="Tasks Completed"
        value={tasksDone}
        subtext="today"
        color="text-bujo-done"
      />
      <StatCard
        icon={AlertCircle}
        label="Outstanding Tasks"
        value={overdueCount}
        subtext="need attention"
        color="text-bujo-task"
      />
      <StatCard
        icon={Calendar}
        label="Habits Today"
        value={`${habitsLoggedToday}/${habits.length}`}
        subtext={`${longestStreak} day streak`}
        color="text-bujo-streak"
      />
      <StatCard
        icon={CheckCircle2}
        label="Monthly Goals"
        value={`${goalsCompleted}/${monthGoals.length}`}
        subtext={format(new Date(), 'MMMM')}
        color="text-primary"
      />
    </div>
  );
}

interface StatCardProps {
  icon: React.ElementType;
  label: string;
  value: string | number;
  subtext: string;
  color: string;
}

function StatCard({ icon: Icon, label, value, subtext, color }: StatCardProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-4 hover:bg-secondary/30 transition-colors">
      <div className="flex items-center gap-2 mb-2">
        <Icon className={cn('w-4 h-4', color)} />
        <span className="text-xs text-muted-foreground">{label}</span>
      </div>
      <div className="font-display text-2xl font-semibold">{value}</div>
      <div className="text-xs text-muted-foreground">{subtext}</div>
    </div>
  );
}

import { BarChart3, CheckCircle2, Circle, FileText, Calendar, Flame, Target } from 'lucide-react';
import { cn } from '@/lib/utils';
import { DayEntries, Habit, Goal, Entry } from '@/types/bujo';
import { format } from 'date-fns';

interface StatsViewProps {
  days: DayEntries[];
  habits: Habit[];
  goals: Goal[];
}

function flattenEntries(entries: Entry[]): Entry[] {
  const result: Entry[] = [];
  function traverse(items: Entry[]) {
    for (const entry of items) {
      result.push(entry);
      if (entry.children && entry.children.length > 0) {
        traverse(entry.children);
      }
    }
  }
  traverse(entries);
  return result;
}

export function StatsView({ days, habits, goals }: StatsViewProps) {
  const allEntries = days.flatMap(day => flattenEntries(day.entries));

  const totalEntries = allEntries.length;
  const taskCount = allEntries.filter(e => e.type === 'task' || e.type === 'done').length;
  const doneCount = allEntries.filter(e => e.type === 'done').length;
  const noteCount = allEntries.filter(e => e.type === 'note').length;
  const eventCount = allEntries.filter(e => e.type === 'event').length;

  const taskPercentage = totalEntries > 0 ? Math.round((taskCount / totalEntries) * 100) : 0;
  const completionRate = taskCount > 0 ? Math.round((doneCount / taskCount) * 100) : 0;

  const bestStreak = Math.max(...habits.map(h => h.streak), 0);

  const currentMonth = format(new Date(), 'yyyy-MM');
  const monthGoals = goals.filter(g => g.month === currentMonth);
  const completedGoals = monthGoals.filter(g => g.status === 'done').length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-2">
        <BarChart3 className="w-5 h-5 text-primary" />
        <h2 className="font-display text-xl font-semibold">Insights</h2>
      </div>

      {/* Entry Stats */}
      <div className="space-y-4">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
          Entries Overview
        </h3>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <StatCard
            icon={FileText}
            label="Total Entries"
            value={totalEntries}
            color="text-primary"
          />
          <StatCard
            icon={Circle}
            label="Tasks"
            value={taskCount}
            subtext={`${taskPercentage}%`}
            color="text-bujo-task"
          />
          <StatCard
            icon={CheckCircle2}
            label="Completion Rate"
            value={`${completionRate}%`}
            subtext={`${doneCount}/${taskCount} done`}
            color="text-bujo-done"
          />
          <StatCard
            icon={FileText}
            label="Notes"
            value={noteCount}
            color="text-bujo-note"
          />
        </div>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <StatCard
            icon={Calendar}
            label="Events"
            value={eventCount}
            color="text-foreground"
          />
        </div>
      </div>

      {/* Habit Stats */}
      <div className="space-y-4">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
          Habits
        </h3>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <StatCard
            icon={Flame}
            label="Active Habits"
            value={habits.length}
            color="text-bujo-streak"
          />
          <StatCard
            icon={Flame}
            label="Best Streak"
            value={bestStreak}
            subtext={`${bestStreak} days`}
            color="text-bujo-streak"
          />
        </div>
      </div>

      {/* Goal Stats */}
      <div className="space-y-4">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
          Goals
        </h3>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <StatCard
            icon={Target}
            label="Monthly Goals"
            value={`${completedGoals}/${monthGoals.length}`}
            subtext={format(new Date(), 'MMMM')}
            color="text-primary"
          />
        </div>
      </div>
    </div>
  );
}

interface StatCardProps {
  icon: React.ElementType;
  label: string;
  value: string | number;
  subtext?: string;
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
      {subtext && <div className="text-xs text-muted-foreground">{subtext}</div>}
    </div>
  );
}

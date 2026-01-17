import { Goal } from '@/types/bujo';
import { cn } from '@/lib/utils';
import { Target, CheckCircle2, Circle, ChevronLeft, ChevronRight } from 'lucide-react';
import { format, parse } from 'date-fns';
import { useState } from 'react';

interface GoalsViewProps {
  goals: Goal[];
}

function getMonthLabel(monthStr: string): string {
  const date = parse(monthStr, 'yyyy-MM', new Date());
  return format(date, 'MMMM yyyy');
}

export function GoalsView({ goals: initialGoals }: GoalsViewProps) {
  const [currentMonth, setCurrentMonth] = useState(() => format(new Date(), 'yyyy-MM'));
  
  const filteredGoals = initialGoals.filter(g => g.month === currentMonth);
  const completedCount = filteredGoals.filter(g => g.completed).length;
  const progress = filteredGoals.length > 0 
    ? Math.round((completedCount / filteredGoals.length) * 100)
    : 0;
  
  const navigateMonth = (delta: number) => {
    const date = parse(currentMonth, 'yyyy-MM', new Date());
    date.setMonth(date.getMonth() + delta);
    setCurrentMonth(format(date, 'yyyy-MM'));
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
      </div>
      
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
              className={cn(
                'flex items-center gap-3 p-3 rounded-lg border border-border',
                'bg-card hover:bg-secondary/30 transition-colors cursor-pointer group animate-fade-in'
              )}
            >
              {goal.completed ? (
                <CheckCircle2 className="w-5 h-5 text-bujo-done flex-shrink-0" />
              ) : (
                <Circle className="w-5 h-5 text-muted-foreground flex-shrink-0" />
              )}
              <span className={cn(
                'flex-1 text-sm',
                goal.completed && 'line-through text-muted-foreground'
              )}>
                {goal.content}
              </span>
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
    </div>
  );
}

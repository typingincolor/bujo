import { Activity } from 'lucide-react';

export interface HabitDisplay {
  name: string;
  count: number;
}

interface HabitItemProps extends HabitDisplay {
  datePrefix?: string;
}

export function HabitItem({ name, count, datePrefix }: HabitItemProps) {
  const displayText = count > 1 ? `${name} (${count})` : name;

  return (
    <div className="px-2 py-1.5 rounded-lg text-sm text-muted-foreground flex items-center gap-2">
      {datePrefix && <span className="text-xs opacity-70">{datePrefix}</span>}
      <Activity className="h-3 w-3 flex-shrink-0" />
      <span className="truncate">{displayText}</span>
    </div>
  );
}

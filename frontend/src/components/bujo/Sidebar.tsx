import { cn } from '@/lib/utils';
import {
  CalendarDays,
  Flame,
  List,
  Target,
  Settings,
  BookOpen,
  Search,
  BarChart3,
  HelpCircle,
  FileEdit
} from 'lucide-react';

export type ViewType = 'today' | 'week' | 'questions' | 'habits' | 'lists' | 'goals' | 'search' | 'stats' | 'settings' | 'editable';

interface SidebarProps {
  currentView: ViewType;
  onViewChange: (view: ViewType) => void;
}

const navItems: { view: ViewType; icon: React.ElementType; label: string }[] = [
  { view: 'today', icon: FileEdit, label: 'Edit Journal' },
  { view: 'week', icon: CalendarDays, label: 'Weekly Review' },
  { view: 'questions', icon: HelpCircle, label: 'Open Questions' },
  { view: 'habits', icon: Flame, label: 'Habit Tracker' },
  { view: 'lists', icon: List, label: 'Lists' },
  { view: 'goals', icon: Target, label: 'Monthly Goals' },
  { view: 'search', icon: Search, label: 'Search' },
  { view: 'stats', icon: BarChart3, label: 'Insights' },
];

export function Sidebar({ currentView, onViewChange }: SidebarProps) {
  return (
    <aside className="w-56 h-screen bg-sidebar border-r border-sidebar-border flex flex-col">
      {/* Logo */}
      <div className="p-4 border-b border-sidebar-border">
        <div className="flex items-center gap-2">
          <BookOpen className="w-6 h-6 text-primary" />
          <h1 className="font-display text-xl font-bold tracking-tight">bujo</h1>
        </div>
        <p className="text-xs text-muted-foreground mt-1">Capture. Track. Reflect.</p>
      </div>
      
      {/* Navigation */}
      <nav className="flex-1 p-3 space-y-1">
        {navItems.map(({ view, icon: Icon, label }) => (
          <button
            key={view}
            onClick={() => onViewChange(view)}
            aria-pressed={currentView === view}
            className={cn(
              'w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-all',
              currentView === view
                ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
                : 'text-sidebar-foreground hover:bg-sidebar-accent/50'
            )}
          >
            <Icon className="w-4 h-4" />
            {label}
          </button>
        ))}
      </nav>
      
      {/* Footer */}
      <div className="p-3">
        <button
          onClick={() => onViewChange('settings')}
          aria-pressed={currentView === 'settings'}
          className={cn(
            'w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors',
            currentView === 'settings'
              ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
              : 'text-sidebar-foreground hover:bg-sidebar-accent/50'
          )}
        >
          <Settings className="w-4 h-4" />
          Settings
        </button>
      </div>
    </aside>
  );
}

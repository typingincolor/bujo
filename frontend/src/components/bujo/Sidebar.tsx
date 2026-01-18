import { cn } from '@/lib/utils';
import {
  Calendar,
  CalendarDays,
  Flame,
  List,
  Target,
  Settings,
  BookOpen,
  Search,
  BarChart3,
  Clock,
  HelpCircle
} from 'lucide-react';

export type ViewType = 'today' | 'week' | 'overview' | 'questions' | 'habits' | 'lists' | 'goals' | 'search' | 'stats' | 'settings';

interface SidebarProps {
  currentView: ViewType;
  onViewChange: (view: ViewType) => void;
}

const navItems: { view: ViewType; icon: React.ElementType; label: string }[] = [
  { view: 'today', icon: Calendar, label: 'Today' },
  { view: 'week', icon: CalendarDays, label: 'Review' },
  { view: 'overview', icon: Clock, label: 'Outstanding' },
  { view: 'questions', icon: HelpCircle, label: 'Questions' },
  { view: 'habits', icon: Flame, label: 'Habits' },
  { view: 'lists', icon: List, label: 'Lists' },
  { view: 'goals', icon: Target, label: 'Goals' },
  { view: 'search', icon: Search, label: 'Search' },
  { view: 'stats', icon: BarChart3, label: 'Stats' },
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
        <p className="text-xs text-muted-foreground mt-1">Your digital bullet journal</p>
      </div>
      
      {/* Navigation */}
      <nav className="flex-1 p-3 space-y-1">
        {navItems.map(({ view, icon: Icon, label }) => (
          <button
            key={view}
            onClick={() => onViewChange(view)}
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
      <div className="p-3 border-t border-sidebar-border">
        <button
          onClick={() => onViewChange('settings')}
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

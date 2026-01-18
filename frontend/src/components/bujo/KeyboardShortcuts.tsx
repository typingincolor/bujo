import { Keyboard } from 'lucide-react';

type ViewType = 'today' | 'week' | 'habits' | 'lists' | 'goals' | 'search' | 'stats' | 'settings';

interface KeyboardShortcutsProps {
  view?: ViewType;
}

interface KeyboardHintProps {
  keys: string[];
  action: string;
}

function KeyboardHint({ keys, action }: KeyboardHintProps) {
  return (
    <div className="flex items-center gap-2 text-xs">
      <div className="flex items-center gap-0.5">
        {keys.map((key, i) => (
          <span key={i}>
            <kbd className="px-1.5 py-0.5 rounded bg-muted border border-border font-mono text-[10px]">
              {key}
            </kbd>
            {i < keys.length - 1 && <span className="mx-0.5 text-muted-foreground">/</span>}
          </span>
        ))}
      </div>
      <span className="text-muted-foreground">{action}</span>
    </div>
  );
}

export function KeyboardShortcuts({ view = 'today' }: KeyboardShortcutsProps) {
  const isMac = typeof navigator !== 'undefined' && navigator.platform.includes('Mac');
  const cmdKey = isMac ? '⌘' : 'Ctrl';

  return (
    <div className="rounded-lg border border-border bg-card p-4 space-y-3">
      <div className="flex items-center gap-2 mb-3">
        <Keyboard className="w-4 h-4 text-primary" />
        <h3 className="font-medium text-sm">Keyboard Shortcuts</h3>
      </div>

      {view === 'today' && (
        <div className="grid grid-cols-2 gap-2">
          <KeyboardHint keys={['j', '↓']} action="Move down" />
          <KeyboardHint keys={['k', '↑']} action="Move up" />
          <KeyboardHint keys={['h']} action="Previous day" />
          <KeyboardHint keys={['l']} action="Next day" />
          <KeyboardHint keys={['Space']} action="Toggle done" />
          <KeyboardHint keys={['e']} action="Edit entry" />
          <KeyboardHint keys={['d']} action="Delete entry" />
        </div>
      )}

      {view === 'habits' && (
        <div className="grid grid-cols-1 gap-2">
          <KeyboardHint keys={['w']} action="Cycle view (week/month/quarter)" />
          <KeyboardHint keys={['Click']} action="Log occurrence" />
          <KeyboardHint keys={[`${cmdKey}+Click`]} action="Remove occurrence" />
        </div>
      )}

      {(view !== 'today' && view !== 'habits') && (
        <div className="text-xs text-muted-foreground">
          No shortcuts for this view
        </div>
      )}
    </div>
  );
}

import { Keyboard } from 'lucide-react';
import { ViewType } from './Sidebar';

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
            <kbd className="px-1.5 py-0.5 rounded bg-muted border border-border text-[10px]">
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
          <KeyboardHint keys={['T']} action="Go to today" />
          <KeyboardHint keys={['[']} action="Toggle sidebar" />
          <KeyboardHint keys={['Space']} action="Toggle done" />
          <KeyboardHint keys={['x']} action="Cancel/uncancel" />
          <KeyboardHint keys={['p']} action="Cycle priority" />
          <KeyboardHint keys={['t']} action="Cycle type" />
          <KeyboardHint keys={['Enter']} action="Expand context" />
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

      {view === 'search' && (
        <div className="grid grid-cols-2 gap-2">
          <KeyboardHint keys={['j', '↓']} action="Move down" />
          <KeyboardHint keys={['k', '↑']} action="Move up" />
          <KeyboardHint keys={['Space']} action="Toggle done" />
          <KeyboardHint keys={['x']} action="Cancel/uncancel" />
          <KeyboardHint keys={['p']} action="Cycle priority" />
          <KeyboardHint keys={['t']} action="Cycle type" />
          <KeyboardHint keys={['Enter']} action="Expand context" />
        </div>
      )}

      {view === 'questions' && (
        <div className="grid grid-cols-2 gap-2">
          <KeyboardHint keys={['j', '↓']} action="Move down" />
          <KeyboardHint keys={['k', '↑']} action="Move up" />
          <KeyboardHint keys={['a']} action="Answer question" />
          <KeyboardHint keys={['x']} action="Cancel/uncancel" />
          <KeyboardHint keys={['p']} action="Cycle priority" />
          <KeyboardHint keys={['t']} action="Cycle type" />
          <KeyboardHint keys={['Enter']} action="Expand context" />
        </div>
      )}

      {view === 'editable' && (
        <>
          <div className="grid grid-cols-2 gap-2">
            <KeyboardHint keys={[`${cmdKey}+S`]} action="Save" />
            <KeyboardHint keys={['Tab']} action="Indent" />
            <KeyboardHint keys={['⇧+Tab']} action="Outdent" />
            <KeyboardHint keys={[`${cmdKey}+I`]} action="Import" />
            <KeyboardHint keys={['Esc']} action="Blur editor" />
          </div>
          <div className="border-t border-border pt-3 mt-3">
            <h4 className="font-medium text-xs mb-2">Syntax Reference</h4>
            <div className="grid grid-cols-2 gap-1 text-xs text-muted-foreground">
              <span>. Task</span>
              <span>- Note</span>
              <span>o Event</span>
              <span>x Done</span>
            </div>
            <div className="grid grid-cols-3 gap-1 text-xs text-muted-foreground mt-2">
              <span>!!! Highest</span>
              <span>!! High</span>
              <span>! Low</span>
            </div>
            <div className="text-xs text-muted-foreground mt-2">
              <span>&gt;[date] Migrate to date</span>
            </div>
          </div>
        </>
      )}

      {!['today', 'habits', 'search', 'questions', 'editable'].includes(view) && (
        <div className="text-xs text-muted-foreground">
          No shortcuts for this view
        </div>
      )}

      <div className="border-t border-border pt-2 mt-2">
        <KeyboardHint keys={['?']} action="Toggle this panel" />
      </div>
    </div>
  );
}

import { Keyboard } from 'lucide-react';

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

export function KeyboardShortcuts() {
  return (
    <div className="rounded-lg border border-border bg-card p-4 space-y-3">
      <div className="flex items-center gap-2 mb-3">
        <Keyboard className="w-4 h-4 text-primary" />
        <h3 className="font-medium text-sm">Keyboard Shortcuts</h3>
      </div>
      
      <div className="grid grid-cols-2 gap-2">
        <KeyboardHint keys={['j', '↓']} action="Move down" />
        <KeyboardHint keys={['k', '↑']} action="Move up" />
        <KeyboardHint keys={['Space']} action="Toggle done" />
        <KeyboardHint keys={['x']} action="Cancel task" />
        <KeyboardHint keys={['e']} action="Edit entry" />
        <KeyboardHint keys={['a']} action="Add sibling" />
        <KeyboardHint keys={['A']} action="Add child" />
        <KeyboardHint keys={['d']} action="Delete entry" />
        <KeyboardHint keys={['m']} action="Migrate task" />
        <KeyboardHint keys={['/']} action="Go to date" />
        <KeyboardHint keys={['w']} action="Toggle view" />
        <KeyboardHint keys={['?']} action="Show help" />
      </div>
    </div>
  );
}

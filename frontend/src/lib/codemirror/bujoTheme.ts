import { EditorView } from '@codemirror/view'

export const bujoTheme = EditorView.theme({
  '&': {
    backgroundColor: 'hsl(var(--background))',
    color: 'hsl(var(--foreground))',
    height: '100%',
  },
  '&.cm-focused': {
    outline: 'none',
  },
  '& .cm-scroller': {
    overflow: 'auto',
    fontFamily: 'var(--font-mono, monospace)',
    fontSize: '14px',
    lineHeight: '1.6',
  },
  '& .cm-content': {
    fontFamily: 'var(--font-mono, monospace)',
    fontSize: '14px',
    lineHeight: '1.6',
  },
  '& .cm-gutters': {
    backgroundColor: 'transparent',
    color: 'hsl(var(--muted-foreground))',
    border: 'none',
  },
  '& .cm-activeLineGutter': {
    backgroundColor: 'transparent',
  },
  '& .cm-activeLine': {
    backgroundColor: 'transparent',
  },
  '& .cm-cursor': {
    borderLeftColor: 'hsl(var(--foreground))',
  },
  '&.cm-editor .cm-selectionBackground': {
    backgroundColor: 'hsl(var(--primary) / 0.2)',
  },
  '&.cm-editor.cm-focused > .cm-scroller > .cm-selectionLayer .cm-selectionBackground': {
    backgroundColor: 'hsl(var(--primary) / 0.3)',
  },
  '& .priority-badge': {
    pointerEvents: 'none',
    userSelect: 'none',
    display: 'inline-block',
    borderRadius: '3px',
    padding: '0 4px',
    fontSize: '11px',
    fontWeight: '700',
    lineHeight: '1.4',
    color: 'white',
    verticalAlign: 'baseline',
  },
  '& .priority-badge-1': {
    backgroundColor: 'hsl(var(--priority-high))',
  },
  '& .priority-badge-2': {
    backgroundColor: 'hsl(var(--priority-medium))',
  },
  '& .priority-badge-3': {
    backgroundColor: 'hsl(var(--priority-low))',
  },
  '& .indent-guide': {
    pointerEvents: 'none',
    userSelect: 'none',
  },
  '& .migration-date-preview': {
    pointerEvents: 'none',
    userSelect: 'none',
  },
  '& .cm-foldGutter': {
    width: '16px',
  },
  '& .cm-foldGutter .cm-gutterElement': {
    cursor: 'pointer',
    color: 'hsl(var(--muted-foreground))',
    fontSize: '12px',
    lineHeight: '1.6',
    padding: '0 2px',
    transition: 'color 0.15s',
  },
  '& .cm-foldGutter .cm-gutterElement:hover': {
    color: 'hsl(var(--foreground))',
  },
  '& .cm-foldPlaceholder': {
    backgroundColor: 'hsl(var(--muted) / 0.5)',
    border: '1px solid hsl(var(--border))',
    borderRadius: '3px',
    color: 'hsl(var(--muted-foreground))',
    padding: '0 4px',
    fontSize: '11px',
    cursor: 'pointer',
  },
  '& .cm-entry-task': {
    color: 'hsl(var(--primary))',
    fontWeight: '600',
  },
  '& .cm-entry-event': {
    color: 'hsl(var(--bujo-event))',
    fontWeight: '500',
  },
  '& .cm-entry-question': {
    color: 'hsl(var(--bujo-question))',
    fontWeight: '500',
  },
  '& .cm-entry-answered': {
    color: 'hsl(var(--bujo-done))',
    opacity: '0.7',
  },
  '& .cm-entry-done': {
    color: 'hsl(var(--bujo-done))',
    opacity: '0.7',
  },
  '& .cm-entry-cancelled': {
    color: 'hsl(var(--muted-foreground))',
    textDecoration: 'line-through',
    opacity: '0.5',
  },
  '& .cm-entry-migrated': {
    color: 'hsl(var(--muted-foreground))',
    opacity: '0.6',
  },
  '& .cm-entry-movedToList': {
    color: 'hsl(var(--muted-foreground))',
    opacity: '0.6',
  },
  '& .cm-entry-note': {},
  '& .cm-panels': {
    backgroundColor: 'hsl(var(--card))',
    color: 'hsl(var(--foreground))',
    borderBottom: '1px solid hsl(var(--border))',
  },
  '& .cm-panels.cm-panels-top': {
    borderBottom: '1px solid hsl(var(--border))',
  },
  '& .cm-search': {
    padding: '8px 12px',
    display: 'flex',
    flexWrap: 'wrap',
    gap: '4px',
    alignItems: 'center',
    fontSize: '13px',
  },
  '& .cm-search input': {
    backgroundColor: 'hsl(var(--background))',
    color: 'hsl(var(--foreground))',
    border: '1px solid hsl(var(--border))',
    borderRadius: '4px',
    padding: '2px 6px',
    fontSize: '13px',
    outline: 'none',
  },
  '& .cm-search input:focus': {
    borderColor: 'hsl(var(--primary))',
    boxShadow: '0 0 0 1px hsl(var(--primary) / 0.3)',
  },
  '& .cm-search button': {
    backgroundColor: 'hsl(var(--secondary))',
    color: 'hsl(var(--secondary-foreground))',
    border: '1px solid hsl(var(--border))',
    borderRadius: '4px',
    padding: '2px 8px',
    fontSize: '12px',
    cursor: 'pointer',
  },
  '& .cm-search button:hover': {
    backgroundColor: 'hsl(var(--secondary) / 0.8)',
  },
  '& .cm-search label': {
    fontSize: '12px',
    color: 'hsl(var(--muted-foreground))',
  },
  '& .cm-searchMatch': {
    backgroundColor: 'hsl(var(--primary) / 0.2)',
    outline: '1px solid hsl(var(--primary) / 0.4)',
  },
  '& .cm-searchMatch-selected': {
    backgroundColor: 'hsl(var(--primary) / 0.4)',
  },
})

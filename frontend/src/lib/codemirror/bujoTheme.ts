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
    backgroundColor: 'hsl(var(--background))',
  },
  '& .cm-content': {
    fontFamily: 'var(--font-mono, monospace)',
    fontSize: '14px',
    lineHeight: '1.6',
    backgroundColor: 'hsl(var(--background))',
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
  '& .cm-selectionBackground': {
    backgroundColor: 'hsl(var(--primary) / 0.2)',
  },
  '&.cm-focused .cm-selectionBackground': {
    backgroundColor: 'hsl(var(--primary) / 0.3)',
  },
})

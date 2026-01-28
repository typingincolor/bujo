import { EditorView } from '@codemirror/view'

export const bujoTheme = EditorView.theme({
  '&': {
    backgroundColor: 'var(--background)',
    color: 'var(--foreground)',
  },
  '.cm-content': {
    fontFamily: 'var(--font-mono, monospace)',
    fontSize: '14px',
    lineHeight: '1.6',
  },
  '.cm-gutters': {
    backgroundColor: 'var(--muted)',
    borderRight: '1px solid var(--border)',
  },
  '.cm-activeLineGutter': {
    backgroundColor: 'var(--accent)',
  },
  '.cm-activeLine': {
    backgroundColor: 'var(--accent)',
  },
  '.cm-cursor': {
    borderLeftColor: 'var(--foreground)',
  },
  '.cm-selectionBackground': {
    backgroundColor: 'var(--selection, rgba(0, 0, 0, 0.1))',
  },
  '&.cm-focused .cm-selectionBackground': {
    backgroundColor: 'var(--selection, rgba(0, 0, 0, 0.2))',
  },
})

import { useState, useMemo, useEffect, useRef, useCallback } from 'react'
import CodeMirror, { ReactCodeMirrorRef } from '@uiw/react-codemirror'
import { keymap } from '@codemirror/view'
import { indentWithTab, deleteLine } from '@codemirror/commands'
import { bujoTheme } from './bujoTheme'
import { priorityBadgeExtension } from './priorityBadges'
import { indentGuidesExtension } from './indentGuides'
import { errorHighlightExtension, setErrors } from './errorMarkers'
import { bujoFoldExtension } from './bujoFolding'
import type { DocumentError } from './errorMarkers'

interface BujoEditorProps {
  value: string
  onChange: (value: string) => void
  onSave?: () => void
  onImport?: () => void
  onEscape?: () => void
  errors?: DocumentError[]
}

export function BujoEditor({ value, onChange, onSave, onImport, onEscape, errors = [] }: BujoEditorProps) {
  const editorRef = useRef<ReactCodeMirrorRef>(null)

  const onChangeRef = useRef(onChange)
  const onSaveRef = useRef(onSave)
  const onImportRef = useRef(onImport)
  const onEscapeRef = useRef(onEscape)
  useEffect(() => {
    onChangeRef.current = onChange
    onSaveRef.current = onSave
    onImportRef.current = onImport
    onEscapeRef.current = onEscape
  })

  const stableOnChange = useCallback((val: string) => {
    onChangeRef.current(val)
  }, [])

  // eslint-disable-next-line react-hooks/refs -- refs are only read inside keymap run() callbacks (event handlers), not during render
  const [extensions] = useState(() => {
    const keybindings = keymap.of([
      {
        key: 'Mod-s',
        run: () => {
          onSaveRef.current?.()
          return true
        },
      },
      {
        key: 'Mod-i',
        run: () => {
          onImportRef.current?.()
          return true
        },
      },
      {
        key: 'Escape',
        run: () => {
          onEscapeRef.current?.()
          return true
        },
      },
      {
        key: 'Mod-Shift-k',
        run: deleteLine,
      },
      indentWithTab,
    ])

    return [
      bujoTheme,
      keybindings,
      priorityBadgeExtension(),
      indentGuidesExtension(),
      errorHighlightExtension(),
      bujoFoldExtension(),
    ]
  })

  useEffect(() => {
    const view = editorRef.current?.view
    if (view) {
      view.dispatch({
        effects: setErrors.of(errors),
      })
    }
  }, [errors])

  const basicSetupConfig = useMemo(() => ({
    lineNumbers: false,
    foldGutter: false,
    highlightActiveLine: false,
    highlightSelectionMatches: true,
    bracketMatching: false,
    closeBrackets: false,
    autocompletion: false,
  }), [])

  return (
    <CodeMirror
      ref={editorRef}
      value={value}
      onChange={stableOnChange}
      extensions={extensions}
      theme="none"
      height="100%"
      style={{ height: '100%' }}
      basicSetup={basicSetupConfig}
    />
  )
}

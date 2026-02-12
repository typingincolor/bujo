import { useState, useMemo, useEffect, useRef, useCallback } from 'react'
import CodeMirror, { ReactCodeMirrorRef, EditorView } from '@uiw/react-codemirror'
import { keymap } from '@codemirror/view'
import {
  indentWithTab,
  deleteLine,
  moveLineUp,
  moveLineDown,
  insertBlankLine,
} from '@codemirror/commands'
import { search, searchKeymap } from '@codemirror/search'
import { EditorSelection } from '@codemirror/state'
import { bujoTheme } from './bujoTheme'
import { priorityBadgeExtension } from './priorityBadges'
import { indentGuidesExtension } from './indentGuides'
import { errorHighlightExtension, setErrors } from './errorMarkers'
import { bujoFoldExtension, computeFoldAllEffects } from './bujoFolding'
import { entryTypeStyleExtension } from './entryTypeStyles'
import { highlightLineExtension, setHighlight } from './highlightLine'
import { findEntryLine } from './findEntryLine'
import { tagCompletionSource } from './tagAutocomplete'
import { autocompletion } from '@codemirror/autocomplete'
import { Compartment } from '@codemirror/state'
import { BrowserOpenURL } from '@/wailsjs/runtime/runtime'
import type { DocumentError } from './errorMarkers'

const URL_REGEX = /https?:\/\/[^\s<>"{}|\\^`[\]]+/g

const urlClickHandler = EditorView.domEventHandlers({
  click(event: MouseEvent, view: EditorView) {
    if (!event.metaKey) return false
    const pos = view.posAtCoords({ x: event.clientX, y: event.clientY })
    if (pos === null) return false
    const line = view.state.doc.lineAt(pos)
    for (const match of line.text.matchAll(URL_REGEX)) {
      const start = line.from + match.index!
      const end = start + match[0].length
      if (pos >= start && pos <= end) {
        BrowserOpenURL(match[0])
        event.preventDefault()
        return true
      }
    }
    return false
  },
})

const HIGHLIGHT_DURATION_MS = 2000

interface BujoEditorProps {
  value: string
  onChange: (value: string) => void
  onSave?: () => void
  onImport?: () => void
  onEscape?: () => void
  errors?: DocumentError[]
  highlightText?: string | null
  onHighlightDone?: () => void
  tags?: string[]
}

export function BujoEditor({ value, onChange, onSave, onImport, onEscape, errors = [], highlightText, onHighlightDone, tags = [] }: BujoEditorProps) {
  const editorRef = useRef<ReactCodeMirrorRef>(null)
  const tagCompartment = useRef(new Compartment())

  const isInternalChangeRef = useRef(false)
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
    isInternalChangeRef.current = true
    onChangeRef.current(val)
  }, [])

  // eslint-disable-next-line react-hooks/refs -- refs are only read inside keymap run() callbacks (event handlers), not during render
  const [extensions] = useState(() => {
    const insertBlankLineAbove: typeof insertBlankLine = ({ state, dispatch }) => {
      const changes = state.changeByRange(range => {
        const line = state.doc.lineAt(range.head)
        return {
          range: EditorSelection.cursor(line.from),
          changes: { from: line.from, insert: '\n' },
        }
      })
      dispatch(state.update(changes, { scrollIntoView: true, userEvent: 'input' }))
      return true
    }

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
      {
        key: 'Mod-Enter',
        run: insertBlankLine,
      },
      {
        key: 'Mod-Shift-Enter',
        run: insertBlankLineAbove,
      },
      {
        key: 'Alt-ArrowUp',
        run: moveLineUp,
      },
      {
        key: 'Alt-ArrowDown',
        run: moveLineDown,
      },
      indentWithTab,
    ])

    return [
      bujoTheme,
      EditorView.lineWrapping,
      keybindings,
      search(),
      keymap.of(searchKeymap),
      priorityBadgeExtension(),
      indentGuidesExtension(),
      errorHighlightExtension(),
      bujoFoldExtension(),
      entryTypeStyleExtension(),
      highlightLineExtension(),
      urlClickHandler,
      tagCompartment.current.of([]),
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

  useEffect(() => {
    const view = editorRef.current?.view
    if (view) {
      const ext = tags.length > 0
        ? autocompletion({ override: [tagCompletionSource(tags)] })
        : []
      view.dispatch({
        effects: tagCompartment.current.reconfigure(ext),
      })
    }
  }, [tags])

  const highlightTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const highlightTextRef = useRef(highlightText)
  const onHighlightDoneRef = useRef(onHighlightDone)
  useEffect(() => {
    highlightTextRef.current = highlightText
    onHighlightDoneRef.current = onHighlightDone
  })

  const applyHighlight = useCallback((view: EditorView) => {
    const text = highlightTextRef.current
    if (!text) return

    const doc = view.state.doc.toString()
    const match = findEntryLine(doc, text)
    if (!match) return

    view.dispatch({
      effects: setHighlight.of({ from: match.from, to: match.to }),
      selection: { anchor: match.from },
      scrollIntoView: true,
    })

    if (highlightTimerRef.current) clearTimeout(highlightTimerRef.current)
    highlightTimerRef.current = setTimeout(() => {
      try { view.dispatch({ effects: setHighlight.of(null) }) } catch { /* view may be destroyed */ }
      onHighlightDoneRef.current?.()
    }, HIGHLIGHT_DURATION_MS)
  }, [])

  const handleCreateEditor = useCallback((view: EditorView) => {
    view.contentDOM.spellcheck = true

    const effects = computeFoldAllEffects(view.state)
    if (effects.length > 0) {
      view.dispatch({ effects })
    }

    if (highlightTextRef.current) {
      applyHighlight(view)
    }
  }, [applyHighlight])

  useEffect(() => {
    if (isInternalChangeRef.current) {
      isInternalChangeRef.current = false
      return
    }

    const view = editorRef.current?.view
    if (!view) return

    const effects = computeFoldAllEffects(view.state)
    if (effects.length > 0) {
      view.dispatch({ effects })
    }
  }, [value])

  useEffect(() => {
    if (!highlightText) return

    const view = editorRef.current?.view
    if (!view) return

    applyHighlight(view)

    return () => {
      if (highlightTimerRef.current) clearTimeout(highlightTimerRef.current)
    }
  }, [highlightText, applyHighlight])

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
      onCreateEditor={handleCreateEditor}
      extensions={extensions}
      theme="none"
      height="100%"
      style={{ height: '100%' }}
      basicSetup={basicSetupConfig}
    />
  )
}

import { useMemo, useEffect, useRef, useCallback } from 'react'
import CodeMirror, { ReactCodeMirrorRef } from '@uiw/react-codemirror'
import { keymap } from '@codemirror/view'
import { indentWithTab, deleteLine } from '@codemirror/commands'
import { bujoTheme } from './bujoTheme'
import { priorityBadgeExtension } from './priorityBadges'
import { indentGuidesExtension } from './indentGuides'
import { errorHighlightExtension, setErrors } from './errorMarkers'
import {
  migrationDatePreviewExtension,
  setResolvedDates,
  findMigrationDates,
  ResolvedDateInfo,
} from './migrationDatePreview'
import type { DocumentError } from '../editableParser'

export interface ResolvedDate {
  iso: string
  display: string
}

interface BujoEditorProps {
  value: string
  onChange: (value: string) => void
  onSave?: () => void
  onImport?: () => void
  onEscape?: () => void
  errors?: DocumentError[]
  resolveDate?: (dateString: string) => Promise<ResolvedDate>
}

export function BujoEditor({ value, onChange, onSave, onImport, onEscape, errors = [], resolveDate }: BujoEditorProps) {
  const editorRef = useRef<ReactCodeMirrorRef>(null)
  const lastResolvedValueRef = useRef<string | null>(null)
  const resolvedCacheRef = useRef<Map<string, ResolvedDateInfo>>(new Map())

  const onChangeRef = useRef(onChange)
  const onSaveRef = useRef(onSave)
  const onImportRef = useRef(onImport)
  const onEscapeRef = useRef(onEscape)
  onChangeRef.current = onChange
  onSaveRef.current = onSave
  onImportRef.current = onImport
  onEscapeRef.current = onEscape

  const stableOnChange = useCallback((val: string) => {
    onChangeRef.current(val)
  }, [])

  const extensions = useMemo(() => {
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
      migrationDatePreviewExtension(),
    ]
  }, [])

  useEffect(() => {
    const view = editorRef.current?.view
    if (view) {
      view.dispatch({
        effects: setErrors.of(errors),
      })
    }
  }, [errors])

  const resolveMigrationDates = useCallback(async () => {
    if (!resolveDate) return

    if (lastResolvedValueRef.current === value) return
    lastResolvedValueRef.current = value

    const view = editorRef.current?.view
    if (!view) return

    const migrationDates = findMigrationDates(value)
    if (migrationDates.length === 0) {
      view.dispatch({ effects: setResolvedDates.of([]) })
      return
    }

    const uniqueDateStrings = [...new Set(migrationDates.map((m) => m.dateString))]
    const unresolvedStrings = uniqueDateStrings.filter(
      (ds) => !resolvedCacheRef.current.has(ds)
    )

    for (const dateString of unresolvedStrings) {
      try {
        const resolved = await resolveDate(dateString)
        resolvedCacheRef.current.set(dateString, {
          dateString,
          iso: resolved.iso,
          display: resolved.display,
        })
      } catch (err) {
        resolvedCacheRef.current.set(dateString, {
          dateString,
          iso: null,
          display: null,
          error: err instanceof Error ? err.message : 'Invalid date',
        })
      }
    }

    const resolvedDates = uniqueDateStrings
      .map((ds) => resolvedCacheRef.current.get(ds))
      .filter((info): info is ResolvedDateInfo => info !== undefined)

    view.dispatch({ effects: setResolvedDates.of(resolvedDates) })
  }, [value, resolveDate])

  useEffect(() => {
    // Use a small delay to ensure CodeMirror view is initialized
    const timeoutId = setTimeout(() => {
      resolveMigrationDates()
    }, 0)
    return () => clearTimeout(timeoutId)
  }, [resolveMigrationDates])

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
      basicSetup={basicSetupConfig}
    />
  )
}

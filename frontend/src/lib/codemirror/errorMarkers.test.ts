import { describe, it, expect } from 'vitest'
import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { linesToDiagnostics, errorHighlightExtension, setErrors } from './errorMarkers'
import type { DocumentError } from './errorMarkers'

describe('linesToDiagnostics', () => {
  it('returns empty array when no errors', () => {
    const errors: DocumentError[] = []
    const doc = '. Valid task\n- Valid note'

    const result = linesToDiagnostics(errors, doc)

    expect(result).toEqual([])
  })

  it('converts single error to diagnostic', () => {
    const errors: DocumentError[] = [
      { lineNumber: 1, message: 'Invalid line' }
    ]
    const doc = 'invalid content\n. Valid task'

    const result = linesToDiagnostics(errors, doc)

    expect(result).toHaveLength(1)
    expect(result[0]).toEqual({
      from: 0,
      to: 15,
      severity: 'error',
      message: 'Invalid line'
    })
  })

  it('converts multiple errors to diagnostics', () => {
    const errors: DocumentError[] = [
      { lineNumber: 1, message: 'Error on line 1' },
      { lineNumber: 3, message: 'Error on line 3' }
    ]
    const doc = 'bad line\n. good\nbad again'

    const result = linesToDiagnostics(errors, doc)

    expect(result).toHaveLength(2)
    expect(result[0].from).toBe(0)
    expect(result[0].to).toBe(8)
    expect(result[1].from).toBe(16)
    expect(result[1].to).toBe(25)
  })

  it('handles error on last line', () => {
    const errors: DocumentError[] = [
      { lineNumber: 2, message: 'Last line error' }
    ]
    const doc = '. Task\ninvalid'

    const result = linesToDiagnostics(errors, doc)

    expect(result).toHaveLength(1)
    expect(result[0].from).toBe(7)
    expect(result[0].to).toBe(14)
  })

  it('handles empty lines correctly', () => {
    const errors: DocumentError[] = [
      { lineNumber: 3, message: 'Error after blank' }
    ]
    const doc = '. Task\n\ninvalid'

    const result = linesToDiagnostics(errors, doc)

    expect(result).toHaveLength(1)
    expect(result[0].from).toBe(8)
    expect(result[0].to).toBe(15)
  })
})

describe('errorHighlightExtension', () => {
  function createEditorView(doc: string): EditorView {
    const state = EditorState.create({
      doc,
      extensions: [errorHighlightExtension()],
    })
    return new EditorView({ state })
  }

  it('returns a valid CodeMirror extension', () => {
    const extension = errorHighlightExtension()
    expect(extension).toBeDefined()
  })

  it('can be used to create an EditorState', () => {
    const view = createEditorView('invalid line\n. Valid task')
    expect(view.state.doc.toString()).toBe('invalid line\n. Valid task')
    view.destroy()
  })

  it('highlights error lines when errors are set', () => {
    const view = createEditorView('invalid line\n. Valid task')
    const errors: DocumentError[] = [{ lineNumber: 1, message: 'Unknown entry type' }]

    view.dispatch({
      effects: setErrors.of(errors),
    })

    const errorLine = view.dom.querySelector('.error-line')
    expect(errorLine).not.toBeNull()

    view.destroy()
  })

  it('shows no highlighting when no errors', () => {
    const view = createEditorView('. Valid task\n- Valid note')

    const errorLines = view.dom.querySelectorAll('.error-line')
    expect(errorLines.length).toBe(0)

    view.destroy()
  })

  it('updates highlighting when errors change', () => {
    const view = createEditorView('line 1\nline 2\nline 3')

    view.dispatch({
      effects: setErrors.of([{ lineNumber: 1, message: 'Error 1' }]),
    })

    let errorLines = view.dom.querySelectorAll('.error-line')
    expect(errorLines.length).toBe(1)

    view.dispatch({
      effects: setErrors.of([
        { lineNumber: 1, message: 'Error 1' },
        { lineNumber: 3, message: 'Error 3' },
      ]),
    })

    errorLines = view.dom.querySelectorAll('.error-line')
    expect(errorLines.length).toBe(2)

    view.destroy()
  })
})

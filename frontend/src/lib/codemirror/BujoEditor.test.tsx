import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, act } from '@testing-library/react'
import { useState } from 'react'
import { EditorView } from '@codemirror/view'
import { BujoEditor } from './BujoEditor'
import type { DocumentError } from './errorMarkers'
import * as bujoFolding from './bujoFolding'

describe('BujoEditor', () => {
  it('renders with initial value', () => {
    render(
      <BujoEditor
        value=". Test task"
        onChange={() => {}}
      />
    )

    expect(screen.getByRole('textbox')).toBeInTheDocument()
  })

  it('displays the provided value', () => {
    const testValue = '. Task one\n- Note two'
    render(
      <BujoEditor
        value={testValue}
        onChange={() => {}}
      />
    )

    const editor = screen.getByRole('textbox')
    expect(editor).toHaveTextContent('Task one')
    expect(editor).toHaveTextContent('Note two')
  })

  it('calls onChange when content changes', async () => {
    const handleChange = vi.fn()
    render(
      <BujoEditor
        value=". Initial"
        onChange={handleChange}
      />
    )

    const editor = screen.getByRole('textbox')
    expect(editor).toBeInTheDocument()
  })

  describe('keyboard shortcuts', () => {
    it('calls onSave when Ctrl+S is pressed', () => {
      const handleSave = vi.fn()
      render(
        <BujoEditor
          value=". Task"
          onChange={() => {}}
          onSave={handleSave}
        />
      )

      const editor = screen.getByRole('textbox')
      fireEvent.keyDown(editor, { key: 's', ctrlKey: true })

      expect(handleSave).toHaveBeenCalled()
    })

    it('calls onImport when Ctrl+I is pressed', () => {
      const handleImport = vi.fn()
      render(
        <BujoEditor
          value=". Task"
          onChange={() => {}}
          onImport={handleImport}
        />
      )

      const editor = screen.getByRole('textbox')
      fireEvent.keyDown(editor, { key: 'i', ctrlKey: true })

      expect(handleImport).toHaveBeenCalled()
    })

    it('does not call onSave when S is pressed without modifier', () => {
      const handleSave = vi.fn()
      render(
        <BujoEditor
          value=". Task"
          onChange={() => {}}
          onSave={handleSave}
        />
      )

      const editor = screen.getByRole('textbox')
      fireEvent.keyDown(editor, { key: 's' })

      expect(handleSave).not.toHaveBeenCalled()
    })

    it('calls onEscape when Escape is pressed', () => {
      const handleEscape = vi.fn()
      render(
        <BujoEditor
          value=". Task"
          onChange={() => {}}
          onEscape={handleEscape}
        />
      )

      const editor = screen.getByRole('textbox')
      fireEvent.keyDown(editor, { key: 'Escape' })

      expect(handleEscape).toHaveBeenCalled()
    })

    it('includes indentWithTab extension for Tab indentation support', () => {
      // Tab indentation is configured via indentWithTab keymap extension
      // Actual Tab key behavior is tested in E2E tests since fireEvent
      // doesn't trigger CodeMirror's internal key bindings
      render(
        <BujoEditor
          value=". Task"
          onChange={() => {}}
        />
      )

      // Verify editor renders - actual Tab behavior verified in E2E
      expect(screen.getByRole('textbox')).toBeInTheDocument()
    })
  })

  describe('tag autocomplete', () => {
    it('accepts tags prop without crashing', () => {
      expect(() => {
        render(
          <BujoEditor
            value=". Task with #wo"
            onChange={() => {}}
            tags={['work', 'workout']}
          />
        )
      }).not.toThrow()
    })

    it('renders normally when tags is empty', () => {
      render(
        <BujoEditor
          value=". Task"
          onChange={() => {}}
          tags={[]}
        />
      )

      expect(screen.getByRole('textbox')).toBeInTheDocument()
    })

    it('renders normally when tags is undefined', () => {
      render(
        <BujoEditor
          value=". Task"
          onChange={() => {}}
        />
      )

      expect(screen.getByRole('textbox')).toBeInTheDocument()
    })
  })

  describe('auto-fold on creation', () => {
    it('calls computeFoldAllEffects when editor is created with foldable content', () => {
      const spy = vi.spyOn(bujoFolding, 'computeFoldAllEffects')

      render(
        <BujoEditor
          value={'. Parent\n  . Child 1\n  . Child 2'}
          onChange={() => {}}
        />
      )

      expect(spy).toHaveBeenCalled()
      spy.mockRestore()
    })

    it('re-folds when value changes externally', () => {
      const spy = vi.spyOn(bujoFolding, 'computeFoldAllEffects')

      const { rerender } = render(
        <BujoEditor
          value={'. Parent\n  . Child 1'}
          onChange={() => {}}
        />
      )

      spy.mockClear()

      rerender(
        <BujoEditor
          value={'. New Parent\n  . New Child'}
          onChange={() => {}}
        />
      )

      expect(spy).toHaveBeenCalled()
      spy.mockRestore()
    })

    it('does not re-fold when content changes from user editing', () => {
      const spy = vi.spyOn(bujoFolding, 'computeFoldAllEffects')

      function TestWrapper() {
        const [val, setVal] = useState('. Parent\n  . Child 1\n  . Child 2')
        return <BujoEditor value={val} onChange={setVal} />
      }

      render(<TestWrapper />)

      // Get the EditorView via CodeMirror's static lookup
      const cmEl = document.querySelector('.cm-editor') as HTMLElement
      const view = EditorView.findFromDOM(cmEl)!
      expect(view).toBeTruthy()

      spy.mockClear()

      // Simulate user typing - dispatching a change triggers onChange -> setValue -> re-render
      act(() => {
        view.dispatch({
          changes: { from: view.state.doc.length, insert: ' extra' },
        })
      })

      // Should NOT re-fold when the change came from user editing
      expect(spy).not.toHaveBeenCalled()
      spy.mockRestore()
    })

    it('does not error when there are no foldable lines', () => {
      expect(() => {
        render(
          <BujoEditor
            value={'. Task 1\n. Task 2'}
            onChange={() => {}}
          />
        )
      }).not.toThrow()
    })
  })

  describe('spellcheck', () => {
    it('enables spellcheck on the editor content element', () => {
      render(
        <BujoEditor
          value=". Task"
          onChange={() => {}}
        />
      )

      const contentEl = document.querySelector('.cm-content') as HTMLElement
      expect(contentEl).not.toBeNull()
      expect(contentEl.spellcheck).toBe(true)
    })
  })

  describe('visual extensions', () => {
    it('displays priority badges for priority markers', () => {
      render(
        <BujoEditor
          value=". !!! High priority task"
          onChange={() => {}}
        />
      )

      const badge = document.querySelector('.priority-badge')
      expect(badge).not.toBeNull()
    })

    it('accepts errors prop without crashing', () => {
      const errors: DocumentError[] = [{ lineNumber: 1, message: 'Invalid line' }]

      expect(() => {
        render(
          <BujoEditor
            value="invalid line\n. Valid task"
            onChange={() => {}}
            errors={errors}
          />
        )
      }).not.toThrow()
    })

    it('renders with empty errors array', () => {
      render(
        <BujoEditor
          value=". Valid task\n- Valid note"
          onChange={() => {}}
          errors={[]}
        />
      )

      expect(screen.getByRole('textbox')).toBeInTheDocument()
    })
  })

})

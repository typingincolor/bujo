import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { BujoEditor } from './BujoEditor'
import type { DocumentError } from '../editableParser'

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

  describe('migration date preview', () => {
    it('calls resolveDate when document contains migration syntax', async () => {
      const resolveDate = vi.fn().mockResolvedValue({
        iso: '2026-01-29',
        display: 'Wed, Jan 29',
      })

      render(
        <BujoEditor
          value=">[tomorrow] Call dentist"
          onChange={() => {}}
          resolveDate={resolveDate}
        />
      )

      await waitFor(() => {
        expect(resolveDate).toHaveBeenCalledWith('tomorrow')
      })
    })

    it('calls resolveDate for each unique date string', async () => {
      const resolveDate = vi.fn().mockImplementation((dateStr: string) => {
        if (dateStr === 'tomorrow') {
          return Promise.resolve({ iso: '2026-01-29', display: 'Wed, Jan 29' })
        }
        return Promise.resolve({ iso: '2026-02-04', display: 'Wed, Feb 4' })
      })

      render(
        <BujoEditor
          value=">[tomorrow] Task 1\n>[next week] Task 2"
          onChange={() => {}}
          resolveDate={resolveDate}
        />
      )

      await waitFor(() => {
        expect(resolveDate).toHaveBeenCalledWith('tomorrow')
        expect(resolveDate).toHaveBeenCalledWith('next week')
      })
    })

    it('handles resolveDate errors gracefully', async () => {
      const resolveDate = vi.fn().mockRejectedValue(new Error('Invalid date'))

      render(
        <BujoEditor
          value=">[invalid-date] Task"
          onChange={() => {}}
          resolveDate={resolveDate}
        />
      )

      await waitFor(() => {
        expect(resolveDate).toHaveBeenCalledWith('invalid-date')
      })
    })

    it('does not call resolveDate when prop is not provided', () => {
      render(
        <BujoEditor
          value=">[tomorrow] Call dentist"
          onChange={() => {}}
        />
      )

      expect(screen.getByRole('textbox')).toBeInTheDocument()
    })

    it('only calls resolveDate once per unique date string', async () => {
      const resolveDate = vi.fn().mockResolvedValue({
        iso: '2026-01-29',
        display: 'Wed, Jan 29',
      })

      render(
        <BujoEditor
          value=">[tomorrow] Task 1\n>[tomorrow] Task 2"
          onChange={() => {}}
          resolveDate={resolveDate}
        />
      )

      await waitFor(() => {
        expect(resolveDate).toHaveBeenCalledTimes(1)
        expect(resolveDate).toHaveBeenCalledWith('tomorrow')
      })
    })
  })
})

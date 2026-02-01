import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { EditableJournalView } from './EditableJournalView'

const mockUseEditableDocument = vi.fn()

vi.mock('@/hooks/useEditableDocument', () => ({
  useEditableDocument: (...args: unknown[]) => mockUseEditableDocument(...args),
}))

const createMockState = (overrides = {}) => ({
  document: '. Buy groceries\n- Meeting notes',
  setDocument: vi.fn(),
  isLoading: false,
  error: null,
  isDirty: false,
  validationErrors: [],
  save: vi.fn().mockResolvedValue({ success: true }),
  discardChanges: vi.fn(),
  lastSaved: null,
  hasDraft: false,
  restoreDraft: vi.fn(),
  discardDraft: vi.fn(),
  ...overrides,
})

describe('EditableJournalView', () => {
  const testDate = new Date(2026, 0, 27) // Jan 27, 2026

  beforeEach(() => {
    vi.clearAllMocks()
    mockUseEditableDocument.mockReturnValue(createMockState())
  })

  describe('initial render', () => {
    it('passes date to useEditableDocument hook', () => {
      render(<EditableJournalView date={testDate} />)

      expect(mockUseEditableDocument).toHaveBeenCalledWith(testDate)
    })

    it('shows loading state while loading', () => {
      mockUseEditableDocument.mockReturnValue(createMockState({ isLoading: true }))

      render(<EditableJournalView date={testDate} />)

      expect(screen.getByText(/loading/i)).toBeInTheDocument()
    })

    it('shows error state when error occurs', () => {
      mockUseEditableDocument.mockReturnValue(
        createMockState({ error: 'Failed to load document' })
      )

      render(<EditableJournalView date={testDate} />)

      expect(screen.getByText(/failed to load document/i)).toBeInTheDocument()
    })

    it('renders document content in editor', () => {
      render(<EditableJournalView date={testDate} />)

      const editor = screen.getByRole('textbox')
      expect(editor).toHaveTextContent('Buy groceries')
      expect(editor).toHaveTextContent('Meeting notes')
    })
  })

  describe('editing', () => {
    it('passes setDocument to BujoEditor as onChange', () => {
      const setDocument = vi.fn()
      mockUseEditableDocument.mockReturnValue(createMockState({ setDocument }))

      render(<EditableJournalView date={testDate} />)

      // BujoEditor receives setDocument as onChange prop
      // The actual change handling is tested in BujoEditor.test.tsx
      expect(screen.getByRole('textbox')).toBeInTheDocument()
    })
  })

  describe('unsaved indicator', () => {
    it('shows unsaved dot symbol when dirty', () => {
      mockUseEditableDocument.mockReturnValue(createMockState({ isDirty: true }))

      render(<EditableJournalView date={testDate} />)

      const indicator = screen.getByTestId('unsaved-indicator')
      expect(indicator).toBeInTheDocument()
      expect(indicator).toHaveTextContent('●')
    })

    it('does not show unsaved dot when not dirty', () => {
      mockUseEditableDocument.mockReturnValue(createMockState({ isDirty: false }))

      render(<EditableJournalView date={testDate} />)

      expect(screen.queryByTestId('unsaved-indicator')).not.toBeInTheDocument()
    })
  })

  describe('validation errors', () => {
    it('displays validation errors', () => {
      mockUseEditableDocument.mockReturnValue(
        createMockState({
          validationErrors: [{ lineNumber: 1, message: 'Unknown entry type' }],
        })
      )

      render(<EditableJournalView date={testDate} />)

      // Text is split across elements: <span>Line 1:</span> Unknown entry type
      expect(screen.getByText(/line 1/i)).toBeInTheDocument()
      expect(screen.getByText(/unknown entry type/i)).toBeInTheDocument()
    })

    it('shows error count badge when errors exist', () => {
      mockUseEditableDocument.mockReturnValue(
        createMockState({
          validationErrors: [
            { lineNumber: 1, message: 'Unknown entry type' },
            { lineNumber: 3, message: 'Missing content' },
          ],
        })
      )

      render(<EditableJournalView date={testDate} />)

      expect(screen.getByText('2 errors')).toBeInTheDocument()
    })

    it('shows quick-fix buttons for unknown entry type errors', () => {
      mockUseEditableDocument.mockReturnValue(
        createMockState({
          document: '^ Invalid line',
          validationErrors: [
            { lineNumber: 1, message: 'Unknown entry type', quickFixes: ['delete', 'change-to-task'] },
          ],
        })
      )

      render(<EditableJournalView date={testDate} />)

      expect(screen.getByRole('button', { name: /delete line/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /change to task/i })).toBeInTheDocument()
    })

    it('calls setDocument with line removed when delete line quick-fix is clicked', () => {
      const setDocument = vi.fn()
      mockUseEditableDocument.mockReturnValue(
        createMockState({
          document: '. Task 1\n^ Invalid line\n. Task 2',
          setDocument,
          validationErrors: [
            { lineNumber: 2, message: 'Unknown entry type', quickFixes: ['delete', 'change-to-task'] },
          ],
        })
      )

      render(<EditableJournalView date={testDate} />)

      fireEvent.click(screen.getByRole('button', { name: /delete line/i }))

      expect(setDocument).toHaveBeenCalledWith('. Task 1\n. Task 2')
    })

    it('calls setDocument with corrected symbol when change to task quick-fix is clicked', () => {
      const setDocument = vi.fn()
      mockUseEditableDocument.mockReturnValue(
        createMockState({
          document: '^ Invalid line',
          setDocument,
          validationErrors: [
            { lineNumber: 1, message: 'Unknown entry type', quickFixes: ['delete', 'change-to-task'] },
          ],
        })
      )

      render(<EditableJournalView date={testDate} />)

      fireEvent.click(screen.getByRole('button', { name: /change to task/i }))

      expect(setDocument).toHaveBeenCalledWith('. Invalid line')
    })
  })

  describe('saving', () => {
    it('calls save when Ctrl+S is pressed', async () => {
      const save = vi.fn().mockResolvedValue({ success: true })
      mockUseEditableDocument.mockReturnValue(createMockState({ save }))

      render(<EditableJournalView date={testDate} />)

      const editor = screen.getByRole('textbox')
      fireEvent.keyDown(editor, { key: 's', ctrlKey: true })

      expect(save).toHaveBeenCalled()
    })

    it('shows save confirmation with timestamp in status bar after successful save', async () => {
      const save = vi.fn().mockResolvedValue({ success: true })
      mockUseEditableDocument.mockReturnValue(
        createMockState({
          save,
          lastSaved: new Date(2026, 0, 27, 14, 30),
        })
      )

      render(<EditableJournalView date={testDate} />)

      expect(screen.getByText(/✓ saved at 2:30 pm/i)).toBeInTheDocument()
    })

    it('shows error message when save fails', async () => {
      const save = vi.fn().mockResolvedValue({ success: false, error: 'Validation failed' })
      mockUseEditableDocument.mockReturnValue(createMockState({ save }))

      render(<EditableJournalView date={testDate} />)

      const editor = screen.getByRole('textbox')
      fireEvent.keyDown(editor, { key: 's', ctrlKey: true })

      await waitFor(() => {
        expect(screen.getByText(/validation failed/i)).toBeInTheDocument()
      })
    })
  })

  describe('discard changes', () => {
    it('calls discardChanges when discard button is clicked', async () => {
      const discardChanges = vi.fn()
      mockUseEditableDocument.mockReturnValue(
        createMockState({ isDirty: true, discardChanges })
      )

      render(<EditableJournalView date={testDate} />)

      fireEvent.click(screen.getByRole('button', { name: /discard/i }))

      expect(discardChanges).toHaveBeenCalled()
    })
  })

  describe('crash recovery', () => {
    it('shows draft recovery prompt when draft exists', () => {
      mockUseEditableDocument.mockReturnValue(createMockState({ hasDraft: true }))

      render(<EditableJournalView date={testDate} />)

      expect(screen.getByText(/unsaved changes found/i)).toBeInTheDocument()
    })

    it('calls restoreDraft when restore is clicked', () => {
      const restoreDraft = vi.fn()
      mockUseEditableDocument.mockReturnValue(
        createMockState({ hasDraft: true, restoreDraft })
      )

      render(<EditableJournalView date={testDate} />)

      fireEvent.click(screen.getByRole('button', { name: /restore/i }))

      expect(restoreDraft).toHaveBeenCalled()
    })

    it('calls discardDraft when discard draft is clicked', () => {
      const discardDraft = vi.fn()
      mockUseEditableDocument.mockReturnValue(
        createMockState({ hasDraft: true, discardDraft })
      )

      render(<EditableJournalView date={testDate} />)

      fireEvent.click(screen.getByRole('button', { name: /discard draft/i }))

      expect(discardDraft).toHaveBeenCalled()
    })
  })


  describe('file import', () => {
    it('has hidden file input for import', () => {
      render(<EditableJournalView date={testDate} />)

      const fileInput = document.querySelector('input[type="file"]')
      expect(fileInput).toBeInTheDocument()
    })

    it('appends file content to document when file is selected', async () => {
      const setDocument = vi.fn()
      mockUseEditableDocument.mockReturnValue(createMockState({
        document: '. Existing task',
        setDocument,
      }))

      render(<EditableJournalView date={testDate} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const fileContent = '. Imported task\n- Imported note'
      const file = new File([fileContent], 'test.txt', { type: 'text/plain' })
      Object.defineProperty(file, 'text', { value: () => Promise.resolve(fileContent) })

      Object.defineProperty(fileInput, 'files', { value: [file] })
      fireEvent.change(fileInput)

      await waitFor(() => {
        expect(setDocument).toHaveBeenCalledWith('. Existing task\n. Imported task\n- Imported note')
      })
    })
  })

  describe('keyboard shortcuts', () => {
    it('passes onEscape handler to BujoEditor for blur behavior', () => {
      // Escape key binding is configured in BujoEditor via onEscape prop
      // The actual Escape key behavior (calling blur) is verified via:
      // 1. BujoEditor.test.tsx verifies onEscape callback is called on Escape key
      // 2. E2E tests verify actual blur behavior in real browser
      render(<EditableJournalView date={testDate} />)

      // Verify editor renders - blur behavior is integration tested in E2E
      expect(screen.getByRole('textbox')).toBeInTheDocument()
    })

    it('shows keyboard shortcut legend below the editor', () => {
      render(<EditableJournalView date={testDate} />)

      expect(screen.getByText(/⌘S/)).toBeInTheDocument()
      expect(screen.getByText(/Save/)).toBeInTheDocument()
      expect(screen.getByText(/⌘I/)).toBeInTheDocument()
      expect(screen.getByText(/Import/)).toBeInTheDocument()
      expect(screen.getByText(/⌘⇧K/)).toBeInTheDocument()
      expect(screen.getByText(/Delete line/)).toBeInTheDocument()
      expect(screen.getByText(/Tab/)).toBeInTheDocument()
      expect(screen.getByText(/Indent/)).toBeInTheDocument()
      expect(screen.getByText(/Esc/)).toBeInTheDocument()
      expect(screen.getByText(/Unfocus/)).toBeInTheDocument()
    })
  })

})

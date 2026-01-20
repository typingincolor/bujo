import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CaptureModal } from './CaptureModal'

vi.mock('@/wailsjs/go/wails/App', () => ({
  AddEntry: vi.fn().mockResolvedValue([1]),
  OpenFileDialog: vi.fn().mockResolvedValue(''),
  ReadFile: vi.fn().mockResolvedValue(''),
}))

vi.mock('@/wailsjs/runtime/runtime', () => ({
  OnFileDrop: vi.fn(),
  OnFileDropOff: vi.fn(),
}))

import { AddEntry, OpenFileDialog, ReadFile } from '@/wailsjs/go/wails/App'
import { OnFileDrop, OnFileDropOff } from '@/wailsjs/runtime/runtime'

const mockStorage: Record<string, string> = {}
const mockLocalStorage = {
  getItem: vi.fn((key: string) => mockStorage[key] || null),
  setItem: vi.fn((key: string, value: string) => { mockStorage[key] = value }),
  removeItem: vi.fn((key: string) => { delete mockStorage[key] }),
  clear: vi.fn(() => { Object.keys(mockStorage).forEach(key => delete mockStorage[key]) }),
}

Object.defineProperty(window, 'localStorage', { value: mockLocalStorage })

describe('CaptureModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.clear()
  })

  afterEach(() => {
    mockLocalStorage.clear()
  })

  it('renders nothing when not open', () => {
    const { container } = render(
      <CaptureModal isOpen={false} onClose={() => {}} onEntriesCreated={() => {}} />
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders modal when open', () => {
    render(
      <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
    )
    expect(screen.getByText('Capture Entries')).toBeInTheDocument()
  })

  it('renders textarea for multi-line input', () => {
    render(
      <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
    )
    expect(screen.getByPlaceholderText(/enter entries/i)).toBeInTheDocument()
  })

  it('shows syntax help', () => {
    render(
      <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
    )
    expect(screen.getByText(/task/i)).toBeInTheDocument()
    expect(screen.getByText(/note/i)).toBeInTheDocument()
    expect(screen.getByText(/event/i)).toBeInTheDocument()
  })

  it('calls AddEntry binding when submitting', async () => {
    const onEntriesCreated = vi.fn()
    const user = userEvent.setup()

    render(
      <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={onEntriesCreated} />
    )

    const textarea = screen.getByPlaceholderText(/enter entries/i)
    await user.type(textarea, '. Buy groceries{enter}- Remember to check expiry')
    await user.click(screen.getByText('Save Entries'))

    await waitFor(() => {
      expect(AddEntry).toHaveBeenCalled()
    })
  })

  it('calls onEntriesCreated after successful submission', async () => {
    const onEntriesCreated = vi.fn()
    const user = userEvent.setup()

    render(
      <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={onEntriesCreated} />
    )

    const textarea = screen.getByPlaceholderText(/enter entries/i)
    await user.type(textarea, '. Test entry')
    await user.click(screen.getByText('Save Entries'))

    await waitFor(() => {
      expect(onEntriesCreated).toHaveBeenCalled()
    })
  })

  it('calls onClose when cancel button is clicked', async () => {
    const onClose = vi.fn()
    const user = userEvent.setup()

    render(
      <CaptureModal isOpen={true} onClose={onClose} onEntriesCreated={() => {}} />
    )

    await user.click(screen.getByText('Cancel'))
    expect(onClose).toHaveBeenCalled()
  })

  it('disables save when textarea is empty', () => {
    render(
      <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
    )

    const saveButton = screen.getByText('Save Entries')
    expect(saveButton).toBeDisabled()
  })

  it('clears textarea after successful submission', async () => {
    const user = userEvent.setup()

    render(
      <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
    )

    const textarea = screen.getByPlaceholderText(/enter entries/i)
    await user.type(textarea, '. Test entry')
    await user.click(screen.getByText('Save Entries'))

    await waitFor(() => {
      expect(textarea).toHaveValue('')
    })
  })

  describe('draft auto-save', () => {
    it('saves draft to localStorage as user types', async () => {
      const user = userEvent.setup()

      render(
        <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
      )

      const textarea = screen.getByPlaceholderText(/enter entries/i)
      await user.type(textarea, '. Draft entry')

      await waitFor(() => {
        expect(mockLocalStorage.setItem).toHaveBeenCalledWith('bujo-capture-draft', '. Draft entry')
      })
    })

    it('restores draft from localStorage on open', () => {
      mockStorage['bujo-capture-draft'] = '. Saved draft'

      render(
        <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
      )

      const textarea = screen.getByPlaceholderText(/enter entries/i)
      expect(textarea).toHaveValue('. Saved draft')
    })

    it('clears draft from localStorage after successful save', async () => {
      mockStorage['bujo-capture-draft'] = '. Draft to clear'
      const user = userEvent.setup()

      render(
        <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
      )

      await user.click(screen.getByText('Save Entries'))

      await waitFor(() => {
        expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('bujo-capture-draft')
      })
    })
  })

  describe('syntax support', () => {
    it('shows indentation creates hierarchy hint', () => {
      render(
        <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
      )
      expect(screen.getByText(/indent/i)).toBeInTheDocument()
    })
  })

  describe('file import', () => {
    it('renders import file button', () => {
      render(
        <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
      )
      expect(screen.getByText('Import File')).toBeInTheDocument()
    })

    it('calls OpenFileDialog when import button is clicked', async () => {
      const user = userEvent.setup()

      render(
        <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
      )

      await user.click(screen.getByText('Import File'))

      expect(OpenFileDialog).toHaveBeenCalled()
    })

    it('populates textarea with file contents', async () => {
      const fileContent = `. Task from file
- Note from file
o Event from file`
      vi.mocked(OpenFileDialog).mockResolvedValueOnce(fileContent)
      const user = userEvent.setup()

      render(
        <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
      )

      await user.click(screen.getByText('Import File'))

      await waitFor(() => {
        const textarea = screen.getByPlaceholderText(/enter entries/i)
        expect(textarea).toHaveValue(fileContent)
      })
    })

    it('does not change textarea when user cancels file dialog', async () => {
      vi.mocked(OpenFileDialog).mockResolvedValueOnce('')
      const user = userEvent.setup()

      render(
        <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
      )

      const textarea = screen.getByPlaceholderText(/enter entries/i)
      await user.type(textarea, '. Existing content')
      await user.click(screen.getByText('Import File'))

      await waitFor(() => {
        expect(textarea).toHaveValue('. Existing content')
      })
    })

    it('appends file contents to existing content', async () => {
      const fileContent = `. Task from file`
      vi.mocked(OpenFileDialog).mockResolvedValueOnce(fileContent)
      const user = userEvent.setup()

      render(
        <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
      )

      const textarea = screen.getByPlaceholderText(/enter entries/i)
      await user.type(textarea, '. Existing task')
      await user.click(screen.getByText('Import File'))

      await waitFor(() => {
        expect(textarea).toHaveValue('. Existing task\n. Task from file')
      })
    })
  })

  describe('drag and drop', () => {
    it('registers OnFileDrop handler when modal opens', () => {
      render(
        <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
      )

      expect(OnFileDrop).toHaveBeenCalled()
    })

    it('unregisters OnFileDrop handler when modal closes', () => {
      const { rerender } = render(
        <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
      )

      rerender(
        <CaptureModal isOpen={false} onClose={() => {}} onEntriesCreated={() => {}} />
      )

      expect(OnFileDropOff).toHaveBeenCalled()
    })

    it('reads dropped file and populates textarea', async () => {
      const fileContent = `. Dropped task
- Dropped note`
      vi.mocked(ReadFile).mockResolvedValueOnce(fileContent)

      let dropCallback: (x: number, y: number, paths: string[]) => void = () => {}
      vi.mocked(OnFileDrop).mockImplementation((cb) => {
        dropCallback = cb
      })

      render(
        <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
      )

      dropCallback(100, 100, ['/path/to/file.txt'])

      await waitFor(() => {
        expect(ReadFile).toHaveBeenCalledWith('/path/to/file.txt')
      })

      await waitFor(() => {
        const textarea = screen.getByPlaceholderText(/enter entries/i)
        expect(textarea).toHaveValue(fileContent)
      })
    })

    it('only reads the first dropped file', async () => {
      const fileContent = `. Task from first file`
      vi.mocked(ReadFile).mockResolvedValueOnce(fileContent)

      let dropCallback: (x: number, y: number, paths: string[]) => void = () => {}
      vi.mocked(OnFileDrop).mockImplementation((cb) => {
        dropCallback = cb
      })

      render(
        <CaptureModal isOpen={true} onClose={() => {}} onEntriesCreated={() => {}} />
      )

      dropCallback(100, 100, ['/path/to/first.txt', '/path/to/second.txt'])

      await waitFor(() => {
        expect(ReadFile).toHaveBeenCalledTimes(1)
        expect(ReadFile).toHaveBeenCalledWith('/path/to/first.txt')
      })
    })
  })
})

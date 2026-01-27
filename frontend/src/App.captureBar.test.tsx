import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { SettingsProvider } from './contexts/SettingsContext'
import { createMockEntry, createMockDayEntries, createMockDays, createMockOverdue } from './test/mocks'

const mockDays = createMockDays([createMockDayEntries({
  Entries: [
    createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'First task', CreatedAt: '2026-01-17T10:00:00Z' }),
    createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Note', Content: 'A note', CreatedAt: '2026-01-17T11:00:00Z' }),
    createMockEntry({ ID: 3, EntityID: 'e3', Type: 'Event', Content: 'An event', CreatedAt: '2026-01-17T12:00:00Z' }),
  ],
})])
const mockOverdue = createMockOverdue([])

vi.mock('./wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn().mockReturnValue(() => {}),
  OnFileDrop: vi.fn(),
  OnFileDropOff: vi.fn(),
}))

vi.mock('./wailsjs/go/wails/App', () => ({
  GetDayEntries: vi.fn().mockResolvedValue([{ Date: '2026-01-17T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' }]),
  GetOverdue: vi.fn().mockResolvedValue([]),
  GetHabits: vi.fn().mockResolvedValue({ Habits: [] }),
  GetLists: vi.fn().mockResolvedValue([]),
  GetGoals: vi.fn().mockResolvedValue([]),
  GetOutstandingQuestions: vi.fn().mockResolvedValue([]),
  AddEntry: vi.fn().mockResolvedValue([1]),
  AddChildEntry: vi.fn().mockResolvedValue([1]),
  MarkEntryDone: vi.fn().mockResolvedValue(undefined),
  MarkEntryUndone: vi.fn().mockResolvedValue(undefined),
  EditEntry: vi.fn().mockResolvedValue(undefined),
  DeleteEntry: vi.fn().mockResolvedValue(undefined),
  HasChildren: vi.fn().mockResolvedValue(false),
  CancelEntry: vi.fn().mockResolvedValue(undefined),
  UncancelEntry: vi.fn().mockResolvedValue(undefined),
  CyclePriority: vi.fn().mockResolvedValue(undefined),
  MigrateEntry: vi.fn().mockResolvedValue(100),
  CreateHabit: vi.fn().mockResolvedValue(1),
  SetMood: vi.fn().mockResolvedValue(undefined),
  SetWeather: vi.fn().mockResolvedValue(undefined),
  SetLocation: vi.fn().mockResolvedValue(undefined),
  GetLocationHistory: vi.fn().mockResolvedValue(['Home', 'Office']),
  OpenFileDialog: vi.fn().mockResolvedValue(''),
  ReadFile: vi.fn().mockResolvedValue(''),
}))

import { GetDayEntries, GetOverdue, AddEntry, AddChildEntry, OpenFileDialog } from './wailsjs/go/wails/App'

const mockStorage: Record<string, string> = {}
const mockLocalStorage = {
  getItem: vi.fn((key: string) => mockStorage[key] || null),
  setItem: vi.fn((key: string, value: string) => { mockStorage[key] = value }),
  removeItem: vi.fn((key: string) => { delete mockStorage[key] }),
  clear: vi.fn(() => { Object.keys(mockStorage).forEach(key => delete mockStorage[key]) }),
}

Object.defineProperty(window, 'localStorage', { value: mockLocalStorage })

describe('CaptureBar - Always Visible', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.clear()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  it('shows capture bar at bottom of today view', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
  })

  it('does NOT show type selection buttons (prefix characters are sufficient)', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const captureBar = screen.getByTestId('capture-bar')
    expect(within(captureBar).queryByRole('button', { name: /task/i })).not.toBeInTheDocument()
    expect(within(captureBar).queryByRole('button', { name: /note/i })).not.toBeInTheDocument()
    expect(within(captureBar).queryByRole('button', { name: /event/i })).not.toBeInTheDocument()
    expect(within(captureBar).queryByRole('button', { name: /question/i })).not.toBeInTheDocument()
  })

  it('uses prefix characters for type selection instead of buttons', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, '. Task item')

    // Prefix is kept in input
    expect(input).toHaveValue('. Task item')
  })
})

describe('CaptureBar - Prefix-Based Type Selection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.clear()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  it('Tab blurs input (no type cycling)', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.click(input)

    // Tab should blur the input, not cycle type
    await user.keyboard('{Tab}')
    expect(input).not.toHaveFocus()
  })

  it('typing ". " prefix keeps prefix in input for task', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, '. test task')

    expect(input).toHaveValue('. test task')
  })

  it('typing "- " prefix keeps prefix in input for note', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, '- test note')

    expect(input).toHaveValue('- test note')
  })

  it('typing "o " prefix keeps prefix in input for event', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, 'o test event')

    expect(input).toHaveValue('o test event')
  })

  it('typing "? " prefix keeps prefix in input for question', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, '? test question')

    expect(input).toHaveValue('? test question')
  })
})

describe('CaptureBar - Entry Submission', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.clear()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  it('Enter submits content exactly as typed (user types prefix)', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, '. Buy groceries{Enter}')

    await waitFor(() => {
      expect(AddEntry).toHaveBeenCalledWith('. Buy groceries', expect.any(String))
    })
  })

  it('clears input after submission', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, 'Buy groceries{Enter}')

    await waitFor(() => {
      expect(AddEntry).toHaveBeenCalled()
    })

    expect(input).toHaveValue('')
  })

  it('keeps focus after submission for rapid entry', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, 'Buy groceries{Enter}')

    await waitFor(() => {
      expect(AddEntry).toHaveBeenCalled()
    })

    expect(input).toHaveFocus()
  })

  it('Escape clears input', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, 'Some text{Escape}')

    expect(input).toHaveValue('')
  })
})

describe('CaptureBar - Parent Context (Child Entries)', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.clear()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  it('pressing A on selected entry shows parent context', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    // Select the first entry
    await user.keyboard('j')
    await user.keyboard('A')

    await waitFor(() => {
      expect(screen.getByText(/adding to:/i)).toBeInTheDocument()
      expect(screen.getByText('First task')).toBeInTheDocument()
    })
  })

  it('submitting in child mode calls AddChildEntry', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    // Enter child mode for the first entry (already selected at index 0)
    await user.keyboard('A')

    await waitFor(() => {
      expect(screen.getByText(/adding to:/i)).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, '. Child task{Enter}')

    await waitFor(() => {
      expect(AddChildEntry).toHaveBeenCalledWith(1, '. Child task', expect.any(String))
    })
  })

  it('clicking X clears parent context', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    // Select the first entry and enter child mode
    await user.keyboard('j')
    await user.keyboard('A')

    await waitFor(() => {
      expect(screen.getByText(/adding to:/i)).toBeInTheDocument()
    })

    const clearButton = screen.getByRole('button', { name: /clear parent/i })
    await user.click(clearButton)

    expect(screen.queryByText(/adding to:/i)).not.toBeInTheDocument()
  })
})

describe('CaptureBar - Draft Persistence', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.clear()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  afterEach(() => {
    mockLocalStorage.clear()
  })

  it('saves draft to localStorage on input', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, 'Draft entry')

    await waitFor(() => {
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('bujo-capture-bar-draft', 'Draft entry')
    })
  })

  it('restores draft on mount', async () => {
    mockStorage['bujo-capture-bar-draft'] = 'Saved draft'

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    expect(input).toHaveValue('Saved draft')
  })

  it('clears draft after successful submission', async () => {
    mockStorage['bujo-capture-bar-draft'] = 'Draft to clear'
    const user = userEvent.setup()

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, '{Enter}')

    await waitFor(() => {
      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('bujo-capture-bar-draft')
    })
  })
})

describe('Header - File Upload', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.clear()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  it('shows upload button in header', async () => {
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    expect(screen.getByRole('button', { name: /upload/i })).toBeInTheDocument()
  })

  it('clicking upload button opens file dialog', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const uploadButton = screen.getByRole('button', { name: /upload/i })
    await user.click(uploadButton)

    expect(OpenFileDialog).toHaveBeenCalled()
  })

  it('appends file content to capture bar input', async () => {
    const fileContent = 'Imported content'
    vi.mocked(OpenFileDialog).mockResolvedValueOnce(fileContent)
    const user = userEvent.setup()

    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, 'Existing ')

    const uploadButton = screen.getByRole('button', { name: /upload/i })
    await user.click(uploadButton)

    await waitFor(() => {
      expect(input).toHaveValue('Existing Imported content')
    })
  })
})

describe('CaptureBar - Keyboard Shortcuts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.clear()
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)
  })

  it('i key focuses capture bar', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('i')

    const input = screen.getByTestId('capture-bar-input')
    expect(input).toHaveFocus()
  })

  it('r key focuses capture bar in root mode (clears parent context)', async () => {
    const user = userEvent.setup()
    render(
      <SettingsProvider>
        <App />
      </SettingsProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    // First set up parent context
    await user.keyboard('j')
    await user.keyboard('A')

    await waitFor(() => {
      expect(screen.getByText(/adding to:/i)).toBeInTheDocument()
    })

    // Blur the input (so global keyboard shortcuts work)
    const input = screen.getByTestId('capture-bar-input')
    await user.keyboard('{Escape}')
    input.blur()

    // Press r to clear parent context and focus
    await user.keyboard('r')

    expect(screen.queryByText(/adding to:/i)).not.toBeInTheDocument()
    expect(input).toHaveFocus()
  })

  describe('CaptureBar positioning with dynamic sidebar', () => {
    it('does not use static right-[32rem] class', async () => {
      vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)

      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      )

      await waitFor(() => {
        expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
      })

      const captureBar = screen.getByTestId('capture-bar')

      // Should NOT have the static right-[32rem] class
      expect(captureBar.className).not.toContain('right-[32rem]')
    })

    it('uses dynamic right positioning based on sidebar width', async () => {
      vi.mocked(GetDayEntries).mockResolvedValue(mockDays)
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue)

      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      )

      await waitFor(() => {
        expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
      })

      const captureBar = screen.getByTestId('capture-bar')

      // Sidebar starts collapsed by default (2.5rem wide), capture bar leaves room for it
      expect(captureBar.style.right).toBe('2.5rem')
    })
  })
})

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'
import { createMockEntry, createMockDayEntries, createMockAgenda } from './test/mocks'

const mockEntriesAgenda = createMockAgenda({
  Days: [createMockDayEntries({
    Entries: [
      createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'First task', CreatedAt: '2026-01-17T10:00:00Z' }),
      createMockEntry({ ID: 2, EntityID: 'e2', Type: 'Note', Content: 'A note', CreatedAt: '2026-01-17T11:00:00Z' }),
      createMockEntry({ ID: 3, EntityID: 'e3', Type: 'Event', Content: 'An event', CreatedAt: '2026-01-17T12:00:00Z' }),
    ],
  })],
  Overdue: [],
})

vi.mock('./wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn().mockReturnValue(() => {}),
  OnFileDrop: vi.fn(),
  OnFileDropOff: vi.fn(),
}))

vi.mock('./wailsjs/go/wails/App', () => ({
  GetAgenda: vi.fn().mockResolvedValue({
    Overdue: [],
    Days: [{ Date: '2026-01-17T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' }],
  }),
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

import { GetAgenda, AddEntry, AddChildEntry, OpenFileDialog } from './wailsjs/go/wails/App'

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
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('shows capture bar at bottom of today view', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    expect(screen.getByTestId('capture-bar')).toBeInTheDocument()
  })

  it('shows type selection buttons (Task, Note, Event, Question)', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const captureBar = screen.getByTestId('capture-bar')
    expect(within(captureBar).getByRole('button', { name: /task/i })).toBeInTheDocument()
    expect(within(captureBar).getByRole('button', { name: /note/i })).toBeInTheDocument()
    expect(within(captureBar).getByRole('button', { name: /event/i })).toBeInTheDocument()
    expect(within(captureBar).getByRole('button', { name: /question/i })).toBeInTheDocument()
  })

  it('has Task selected by default', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const captureBar = screen.getByTestId('capture-bar')
    const taskButton = within(captureBar).getByRole('button', { name: /task/i })
    expect(taskButton).toHaveAttribute('aria-pressed', 'true')
  })
})

describe('CaptureBar - Type Selection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.clear()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('clicking type button changes selection', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const captureBar = screen.getByTestId('capture-bar')
    const noteButton = within(captureBar).getByRole('button', { name: /note/i })
    await user.click(noteButton)

    expect(noteButton).toHaveAttribute('aria-pressed', 'true')
    const taskButton = within(captureBar).getByRole('button', { name: /task/i })
    expect(taskButton).toHaveAttribute('aria-pressed', 'false')
  })

  it('Tab cycles through types when input is empty', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const captureBar = screen.getByTestId('capture-bar')
    const input = screen.getByTestId('capture-bar-input')
    await user.click(input)

    // Start with Task selected, Tab should cycle to Note
    await user.keyboard('{Tab}')
    expect(within(captureBar).getByRole('button', { name: /note/i })).toHaveAttribute('aria-pressed', 'true')

    // Tab to Event
    await user.keyboard('{Tab}')
    expect(within(captureBar).getByRole('button', { name: /event/i })).toHaveAttribute('aria-pressed', 'true')

    // Tab to Question
    await user.keyboard('{Tab}')
    expect(within(captureBar).getByRole('button', { name: /question/i })).toHaveAttribute('aria-pressed', 'true')

    // Tab wraps back to Task
    await user.keyboard('{Tab}')
    expect(within(captureBar).getByRole('button', { name: /task/i })).toHaveAttribute('aria-pressed', 'true')
  })

  it('typing ". " prefix changes type to Task', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const captureBar = screen.getByTestId('capture-bar')
    // First select Note type
    const noteButton = within(captureBar).getByRole('button', { name: /note/i })
    await user.click(noteButton)

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, '. ')

    expect(within(captureBar).getByRole('button', { name: /task/i })).toHaveAttribute('aria-pressed', 'true')
  })

  it('typing "- " prefix changes type to Note', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const captureBar = screen.getByTestId('capture-bar')
    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, '- ')

    expect(within(captureBar).getByRole('button', { name: /note/i })).toHaveAttribute('aria-pressed', 'true')
  })

  it('typing "o " prefix changes type to Event', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const captureBar = screen.getByTestId('capture-bar')
    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, 'o ')

    expect(within(captureBar).getByRole('button', { name: /event/i })).toHaveAttribute('aria-pressed', 'true')
  })

  it('typing "? " prefix changes type to Question', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const captureBar = screen.getByTestId('capture-bar')
    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, '? ')

    expect(within(captureBar).getByRole('button', { name: /question/i })).toHaveAttribute('aria-pressed', 'true')
  })
})

describe('CaptureBar - Entry Submission', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.clear()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('Enter submits entry with selected type prefix', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, 'Buy groceries{Enter}')

    await waitFor(() => {
      expect(AddEntry).toHaveBeenCalledWith('. Buy groceries', expect.any(String))
    })
  })

  it('clears input after submission', async () => {
    const user = userEvent.setup()
    render(<App />)

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
    render(<App />)

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
    render(<App />)

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
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('pressing A on selected entry shows parent context', async () => {
    const user = userEvent.setup()
    render(<App />)

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
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    // Enter child mode for the first entry (already selected at index 0)
    await user.keyboard('A')

    await waitFor(() => {
      expect(screen.getByText(/adding to:/i)).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, 'Child task{Enter}')

    await waitFor(() => {
      expect(AddChildEntry).toHaveBeenCalledWith(1, '. Child task', expect.any(String))
    })
  })

  it('clicking X clears parent context', async () => {
    const user = userEvent.setup()
    render(<App />)

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
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  afterEach(() => {
    mockLocalStorage.clear()
  })

  it('saves draft to localStorage on input', async () => {
    const user = userEvent.setup()
    render(<App />)

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

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    expect(input).toHaveValue('Saved draft')
  })

  it('clears draft after successful submission', async () => {
    mockStorage['bujo-capture-bar-draft'] = 'Draft to clear'
    const user = userEvent.setup()

    render(<App />)

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

describe('CaptureBar - File Import', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.clear()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('shows file import button', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    expect(screen.getByRole('button', { name: /import/i })).toBeInTheDocument()
  })

  it('clicking import button opens file dialog', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const importButton = screen.getByRole('button', { name: /import/i })
    await user.click(importButton)

    expect(OpenFileDialog).toHaveBeenCalled()
  })

  it('appends file content to input', async () => {
    const fileContent = 'Imported content'
    vi.mocked(OpenFileDialog).mockResolvedValueOnce(fileContent)
    const user = userEvent.setup()

    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    const input = screen.getByTestId('capture-bar-input')
    await user.type(input, 'Existing ')

    const importButton = screen.getByRole('button', { name: /import/i })
    await user.click(importButton)

    await waitFor(() => {
      expect(input).toHaveValue('Existing Imported content')
    })
  })
})

describe('CaptureBar - Keyboard Shortcuts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockLocalStorage.clear()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesAgenda)
  })

  it('i key focuses capture bar', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('i')

    const input = screen.getByTestId('capture-bar-input')
    expect(input).toHaveFocus()
  })

  it('r key focuses capture bar in root mode (clears parent context)', async () => {
    const user = userEvent.setup()
    render(<App />)

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
})

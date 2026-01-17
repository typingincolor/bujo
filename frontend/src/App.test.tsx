import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import App from './App'

const mockEntriesData = {
  Days: [{
    Date: '2026-01-17T00:00:00Z',
    Entries: [
      { ID: 1, EntityID: 'e1', Type: 'Task', Content: 'First task', Priority: '', ParentID: null, CreatedAt: '2026-01-17T10:00:00Z' },
      { ID: 2, EntityID: 'e2', Type: 'Task', Content: 'Second task', Priority: '', ParentID: null, CreatedAt: '2026-01-17T11:00:00Z' },
      { ID: 3, EntityID: 'e3', Type: 'Note', Content: 'A note', Priority: '', ParentID: null, CreatedAt: '2026-01-17T12:00:00Z' },
    ],
    Location: '',
    Mood: '',
    Weather: '',
  }],
}

const mockEmptyData = {
  Days: [{ Date: '2026-01-17T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' }],
}

vi.mock('./wailsjs/go/wails/App', () => ({
  GetAgenda: vi.fn().mockResolvedValue({
    Days: [{ Date: '2026-01-17T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' }],
  }),
  GetHabits: vi.fn().mockResolvedValue({ Habits: [] }),
  GetLists: vi.fn().mockResolvedValue([]),
  GetGoals: vi.fn().mockResolvedValue([]),
  AddEntry: vi.fn().mockResolvedValue([1]),
  MarkEntryDone: vi.fn().mockResolvedValue(undefined),
  MarkEntryUndone: vi.fn().mockResolvedValue(undefined),
}))

import { GetAgenda, AddEntry, MarkEntryDone } from './wailsjs/go/wails/App'

describe('App - AddEntryBar integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls AddEntry binding when adding entry via AddEntryBar', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText("What's on your mind?")
    fireEvent.change(input, { target: { value: 'Test task' } })

    const submitButton = screen.getByRole('button', { name: '' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(AddEntry).toHaveBeenCalledWith('. Test task', expect.any(String))
    })
  })

  it('refreshes data after adding entry', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    vi.mocked(GetAgenda).mockClear()

    const input = screen.getByPlaceholderText("What's on your mind?")
    fireEvent.change(input, { target: { value: 'Test task' } })

    const submitButton = screen.getByRole('button', { name: '' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(GetAgenda).toHaveBeenCalled()
    })
  })

  it('clears input after successful add', async () => {
    render(<App />)

    await waitFor(() => {
      expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument()
    })

    const input = screen.getByPlaceholderText("What's on your mind?") as HTMLInputElement
    fireEvent.change(input, { target: { value: 'Test task' } })

    const submitButton = screen.getByRole('button', { name: '' })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(input.value).toBe('')
    })
  })
})

describe('App - Keyboard Navigation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GetAgenda).mockResolvedValue(mockEntriesData)
  })

  it('pressing j moves selection down', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('j')

    const secondTask = screen.getByText('Second task').closest('[data-entry-id]')
    expect(secondTask).toHaveAttribute('data-selected', 'true')
  })

  it('pressing k moves selection up', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('jk')

    const firstTask = screen.getByText('First task').closest('[data-entry-id]')
    expect(firstTask).toHaveAttribute('data-selected', 'true')
  })

  it('pressing down arrow moves selection down', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('{ArrowDown}')

    const secondTask = screen.getByText('Second task').closest('[data-entry-id]')
    expect(secondTask).toHaveAttribute('data-selected', 'true')
  })

  it('pressing up arrow moves selection up', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('{ArrowDown}{ArrowUp}')

    const firstTask = screen.getByText('First task').closest('[data-entry-id]')
    expect(firstTask).toHaveAttribute('data-selected', 'true')
  })

  it('pressing Space toggles done on selected task', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard(' ')

    await waitFor(() => {
      expect(MarkEntryDone).toHaveBeenCalledWith(1)
    })
  })

  it('does not go above first entry when pressing k at top', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('kkk')

    const firstTask = screen.getByText('First task').closest('[data-entry-id]')
    expect(firstTask).toHaveAttribute('data-selected', 'true')
  })

  it('does not go below last entry when pressing j at bottom', async () => {
    const user = userEvent.setup()
    render(<App />)

    await waitFor(() => {
      expect(screen.getByText('First task')).toBeInTheDocument()
    })

    await user.keyboard('jjjjj')

    const note = screen.getByText('A note').closest('[data-entry-id]')
    expect(note).toHaveAttribute('data-selected', 'true')
  })
})

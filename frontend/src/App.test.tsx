import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import App from './App'

const mockDayData = {
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

import { GetAgenda, AddEntry } from './wailsjs/go/wails/App'

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

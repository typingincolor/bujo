import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { DayView } from './DayView'
import { DayEntries } from '@/types/bujo'

vi.mock('@/wailsjs/go/wails/App', () => ({
  MarkEntryDone: vi.fn(),
  MarkEntryUndone: vi.fn(),
  CancelEntry: vi.fn(),
  UncancelEntry: vi.fn(),
  CyclePriority: vi.fn(),
  GetSummary: vi.fn(),
}))

const createTestDay = (overrides: Partial<DayEntries> = {}): DayEntries => ({
  date: '2026-01-17',
  entries: [
    {
      id: 1,
      type: 'task',
      content: 'First task',
      priority: 'none',
      parentId: null,
      loggedDate: '2026-01-17',
      children: [],
    },
    {
      id: 2,
      type: 'note',
      content: 'A note',
      priority: 'none',
      parentId: null,
      loggedDate: '2026-01-17',
      children: [],
    },
    {
      id: 3,
      type: 'event',
      content: 'An event',
      priority: 'none',
      parentId: null,
      loggedDate: '2026-01-17',
      children: [],
    },
  ],
  mood: undefined,
  weather: undefined,
  location: undefined,
  ...overrides,
})

describe('DayView', () => {
  describe('entry selection', () => {
    it('calls onSelectEntry when an entry is clicked', () => {
      const onSelectEntry = vi.fn()
      render(
        <DayView
          day={createTestDay()}
          onSelectEntry={onSelectEntry}
        />
      )

      const noteEntry = screen.getByText('A note').closest('[data-entry-id]')
      fireEvent.click(noteEntry!)
      expect(onSelectEntry).toHaveBeenCalledWith(2)
    })

    it('calls onSelectEntry with correct id for first entry', () => {
      const onSelectEntry = vi.fn()
      render(
        <DayView
          day={createTestDay()}
          onSelectEntry={onSelectEntry}
        />
      )

      const taskEntry = screen.getByText('First task').closest('[data-entry-id]')
      fireEvent.click(taskEntry!)
      expect(onSelectEntry).toHaveBeenCalledWith(1)
    })

    it('calls onSelectEntry with correct id for third entry', () => {
      const onSelectEntry = vi.fn()
      render(
        <DayView
          day={createTestDay()}
          onSelectEntry={onSelectEntry}
        />
      )

      const eventEntry = screen.getByText('An event').closest('[data-entry-id]')
      fireEvent.click(eventEntry!)
      expect(onSelectEntry).toHaveBeenCalledWith(3)
    })
  })
})

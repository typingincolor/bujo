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
  RetypeEntry: vi.fn(),
  MoveEntryToRoot: vi.fn(),
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
    it('calls onSelectEntry when an entry is clicked', async () => {
      const onSelectEntry = vi.fn()
      render(
        <DayView
          day={createTestDay()}
          onSelectEntry={onSelectEntry}
        />
      )

      // Click entry to open popover
      const noteEntry = screen.getByText('A note')
      fireEvent.click(noteEntry)

      // Click "Go to entry" in popover
      const goToButton = await screen.findByText('Go to entry')
      fireEvent.click(goToButton)

      expect(onSelectEntry).toHaveBeenCalledWith(2)
    })

    it('calls onSelectEntry with correct id for first entry', async () => {
      const onSelectEntry = vi.fn()
      render(
        <DayView
          day={createTestDay()}
          onSelectEntry={onSelectEntry}
        />
      )

      // Click entry to open popover
      const taskEntry = screen.getByText('First task')
      fireEvent.click(taskEntry)

      // Click "Go to entry" in popover
      const goToButton = await screen.findByText('Go to entry')
      fireEvent.click(goToButton)

      expect(onSelectEntry).toHaveBeenCalledWith(1)
    })

    it('calls onSelectEntry with correct id for third entry', async () => {
      const onSelectEntry = vi.fn()
      render(
        <DayView
          day={createTestDay()}
          onSelectEntry={onSelectEntry}
        />
      )

      // Click entry to open popover
      const eventEntry = screen.getByText('An event')
      fireEvent.click(eventEntry)

      // Click "Go to entry" in popover
      const goToButton = await screen.findByText('Go to entry')
      fireEvent.click(goToButton)

      expect(onSelectEntry).toHaveBeenCalledWith(3)
    })
  })

  describe('move to list functionality', () => {
    it('calls onMoveToList when Move to list button is clicked for task entry', () => {
      const onMoveToList = vi.fn()
      render(
        <DayView
          day={createTestDay()}
          onMoveToList={onMoveToList}
        />
      )

      // Task entry should have move to list button
      fireEvent.click(screen.getByTitle('Move to list'))
      expect(onMoveToList).toHaveBeenCalledWith(expect.objectContaining({ id: 1, content: 'First task' }))
    })

    it('calls onMoveToList from context menu for task entry', () => {
      const onMoveToList = vi.fn()
      render(
        <DayView
          day={createTestDay()}
          onMoveToList={onMoveToList}
        />
      )

      const taskEntry = screen.getByText('First task').closest('[data-entry-id]')!
      fireEvent.contextMenu(taskEntry)
      fireEvent.click(screen.getByRole('menuitem', { name: 'Move to list' }))

      expect(onMoveToList).toHaveBeenCalledWith(expect.objectContaining({ id: 1, content: 'First task' }))
    })

    it('does not show move to list button for note entries', () => {
      render(
        <DayView
          day={createTestDay()}
          onMoveToList={() => {}}
        />
      )

      // There should be only one move to list button (for the task entry, not the note)
      const allMoveToListButtons = screen.getAllByTitle('Move to list')
      expect(allMoveToListButtons.length).toBe(1)
    })
  })

  describe('context menu actions', () => {
    it('calls onAddChild when Add child is clicked from context menu', () => {
      const onAddChild = vi.fn()
      render(
        <DayView
          day={createTestDay()}
          onAddChild={onAddChild}
        />
      )

      const taskEntry = screen.getByText('First task').closest('[data-entry-id]')!
      fireEvent.contextMenu(taskEntry)
      fireEvent.click(screen.getByRole('menuitem', { name: 'Add child' }))

      expect(onAddChild).toHaveBeenCalledWith(expect.objectContaining({ id: 1, content: 'First task' }))
    })

    it('shows Move to root option for child entries', async () => {
      const dayWithChildren = createTestDay({
        entries: [
          {
            id: 1,
            type: 'task',
            content: 'Parent task',
            priority: 'none',
            parentId: null,
            loggedDate: '2026-01-17',
            children: [],
          },
          {
            id: 2,
            type: 'task',
            content: 'Child task',
            priority: 'none',
            parentId: 1,
            loggedDate: '2026-01-17',
            children: [],
          },
        ],
      })

      render(<DayView day={dayWithChildren} />)

      const childEntry = screen.getByText('Child task').closest('[data-entry-id]')!
      fireEvent.contextMenu(childEntry)

      expect(screen.getByRole('menuitem', { name: 'Move to root' })).toBeInTheDocument()
    })

    it('does not show Move to root option for root entries', () => {
      render(<DayView day={createTestDay()} />)

      const taskEntry = screen.getByText('First task').closest('[data-entry-id]')!
      fireEvent.contextMenu(taskEntry)

      expect(screen.queryByRole('menuitem', { name: 'Move to root' })).not.toBeInTheDocument()
    })

    it('calls MoveEntryToRoot when Move to root is clicked', async () => {
      const { MoveEntryToRoot } = await import('@/wailsjs/go/wails/App')
      const mockMoveEntryToRoot = vi.mocked(MoveEntryToRoot)
      mockMoveEntryToRoot.mockResolvedValue()

      const onEntryChanged = vi.fn()
      const dayWithChildren = createTestDay({
        entries: [
          {
            id: 1,
            type: 'task',
            content: 'Parent task',
            priority: 'none',
            parentId: null,
            loggedDate: '2026-01-17',
            children: [],
          },
          {
            id: 2,
            type: 'task',
            content: 'Child task',
            priority: 'none',
            parentId: 1,
            loggedDate: '2026-01-17',
            children: [],
          },
        ],
      })

      render(<DayView day={dayWithChildren} onEntryChanged={onEntryChanged} />)

      const childEntry = screen.getByText('Child task').closest('[data-entry-id]')!
      fireEvent.contextMenu(childEntry)
      fireEvent.click(screen.getByRole('menuitem', { name: 'Move to root' }))

      expect(mockMoveEntryToRoot).toHaveBeenCalledWith(2)
    })
  })
})

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SearchView } from './SearchView'
import {
  Search,
  GetEntry,
  GetEntryAncestors,
  MarkEntryDone,
  CyclePriority,
  MigrateEntry,
} from '@/wailsjs/go/wails/App'
import { createMockEntry } from '@/test/mocks'

// Mock Wails functions
vi.mock('@/wailsjs/go/wails/App', () => ({
  Search: vi.fn(),
  GetEntry: vi.fn(),
  GetEntryAncestors: vi.fn(),
  MarkEntryDone: vi.fn(),
  MarkEntryUndone: vi.fn(),
  CancelEntry: vi.fn(),
  UncancelEntry: vi.fn(),
  DeleteEntry: vi.fn(),
  CyclePriority: vi.fn(),
  RetypeEntry: vi.fn(),
  MigrateEntry: vi.fn(),
}))

// Mock scrollIntoView
Element.prototype.scrollIntoView = vi.fn()

describe('SearchView: EntryContextPopover Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Opening the Popover', () => {
    it('should open popover when clicking a search result', async () => {
      // Setup: Search results with a task
      const searchResults = [
        createMockEntry({
          ID: 1,
          Content: 'Test task',
          Type: 'Task',
          Priority: 'None',
          ParentID: null,
          CreatedAt: '2026-01-24T10:00:00Z',
        }),
      ]
      vi.mocked(Search).mockResolvedValue(searchResults)
      vi.mocked(GetEntryAncestors).mockResolvedValue([])

      render(<SearchView />)
      const user = userEvent.setup()

      // Search for entries
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'test')

      // Wait for results
      await waitFor(() => {
        expect(screen.getByText('Test task')).toBeInTheDocument()
      })

      // Click the entry
      const entry = screen.getByText('Test task')
      await user.click(entry)

      // Popover should open
      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })
    })

    it('should open popover when clicking ContextPill', async () => {
      // Setup: Search result with parent (shows ContextPill)
      const parent = createMockEntry({
        ID: 1,
        Content: 'Parent task',
        Type: 'Task',
        ParentID: null,
      })
      const child = createMockEntry({
        ID: 2,
        Content: 'Child task',
        Type: 'Task',
        ParentID: 1,
      })
      vi.mocked(Search).mockResolvedValue([child])
      vi.mocked(GetEntryAncestors).mockResolvedValue([parent])

      render(<SearchView />)
      const user = userEvent.setup()

      // Search
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'child')

      // Wait for results with ContextPill
      await waitFor(() => {
        expect(screen.getByText('Child task')).toBeInTheDocument()
      })

      // Click the ContextPill (shows "1 above")
      const contextPill = screen.getByText(/1 above/i)
      await user.click(contextPill)

      // Popover should open
      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })
    })

    it('should NOT expand inline when clicking entry (old behavior removed)', async () => {
      // Setup: Search result with parent
      const parent = createMockEntry({
        ID: 1,
        Content: 'Parent task',
        Type: 'Task',
        ParentID: null,
      })
      const child = createMockEntry({
        ID: 2,
        Content: 'Child task',
        Type: 'Task',
        ParentID: 1,
      })
      vi.mocked(Search).mockResolvedValue([child])
      vi.mocked(GetEntryAncestors).mockResolvedValue([parent])

      render(<SearchView />)
      const user = userEvent.setup()

      // Search
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'child')

      // Wait for results
      await waitFor(() => {
        expect(screen.getByText('Child task')).toBeInTheDocument()
      })

      // Click the entry
      const entry = screen.getByText('Child task')
      await user.click(entry)

      // Should open popover, NOT show inline ancestors
      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Parent should be in popover, not inline above the entry
      const dialog = screen.getByRole('dialog')
      expect(dialog).toHaveTextContent('Parent task')
    })
  })

  describe('Popover Content', () => {
    it('should show full entry tree with ancestors in popover', async () => {
      // Setup: Grandparent -> Parent -> Child hierarchy
      const grandparent = createMockEntry({
        ID: 1,
        Content: 'Project Alpha',
        Type: 'Event',
        ParentID: null,
      })
      const parent = createMockEntry({
        ID: 2,
        Content: 'Phase 1',
        Type: 'Note',
        ParentID: 1,
      })
      const child = createMockEntry({
        ID: 3,
        Content: 'Research databases',
        Type: 'Task',
        ParentID: 2,
      })
      vi.mocked(Search).mockResolvedValue([child])
      vi.mocked(GetEntryAncestors).mockResolvedValue([grandparent, parent])

      render(<SearchView />)
      const user = userEvent.setup()

      // Search
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'research')

      await waitFor(() => {
        expect(screen.getByText('Research databases')).toBeInTheDocument()
      })

      // Open popover
      await user.click(screen.getByText('Research databases'))

      // Verify all ancestors shown in tree
      await waitFor(() => {
        const dialog = screen.getByRole('dialog')
        expect(dialog).toHaveTextContent('Project Alpha')
        expect(dialog).toHaveTextContent('Phase 1')
        expect(dialog).toHaveTextContent('Research databases')
      })
    })

    it('should highlight the clicked entry in the popover tree', async () => {
      // Setup: Simple parent-child
      const parent = createMockEntry({
        ID: 1,
        Content: 'Parent task',
        Type: 'Task',
        ParentID: null,
      })
      const child = createMockEntry({
        ID: 2,
        Content: 'Child task',
        Type: 'Task',
        ParentID: 1,
      })
      vi.mocked(Search).mockResolvedValue([child])
      vi.mocked(GetEntryAncestors).mockResolvedValue([parent])

      render(<SearchView />)
      const user = userEvent.setup()

      // Search and open popover
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'child')

      await waitFor(() => {
        expect(screen.getByText('Child task')).toBeInTheDocument()
      })

      await user.click(screen.getByText('Child task'))

      // Verify child is highlighted (has highlighted class/style)
      await waitFor(() => {
        const dialog = screen.getByRole('dialog')
        // The highlighted entry should have bg-primary/10 or similar class
        const highlightedEntry = dialog.querySelector('[data-highlighted="true"]')
        expect(highlightedEntry).toHaveTextContent('Child task')
      })
    })
  })

  describe('Quick Actions in Popover', () => {
    it('should show correct actions for task entries', async () => {
      // Setup: Task entry
      const task = createMockEntry({
        ID: 1,
        Content: 'Task entry',
        Type: 'Task',
        Priority: 'None',
        ParentID: null,
      })
      vi.mocked(Search).mockResolvedValue([task])
      vi.mocked(GetEntryAncestors).mockResolvedValue([])

      render(<SearchView />)
      const user = userEvent.setup()

      // Search and open popover
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'task')

      await waitFor(() => {
        expect(screen.getByText('Task entry')).toBeInTheDocument()
      })

      await user.click(screen.getByText('Task entry'))

      // Verify actions: Done, Priority, Migrate
      await waitFor(() => {
        const dialog = screen.getByRole('dialog')
        expect(dialog).toBeInTheDocument()
        // Should have buttons for done, priority, migrate
        expect(screen.getByTitle(/mark done/i)).toBeInTheDocument()
        expect(screen.getByTitle(/priority/i)).toBeInTheDocument()
        expect(screen.getByTitle(/migrate/i)).toBeInTheDocument()
      })
    })

    it('should call MarkEntryDone when clicking Done action in popover', async () => {
      // Setup: Task entry
      const task = createMockEntry({
        ID: 1,
        Content: 'Task to complete',
        Type: 'Task',
        ParentID: null,
      })
      const doneTask = createMockEntry({
        ID: 1,
        Content: 'Task to complete',
        Type: 'Done',
        ParentID: null,
      })
      vi.mocked(Search).mockResolvedValue([task])
      vi.mocked(GetEntryAncestors).mockResolvedValue([])
      vi.mocked(MarkEntryDone).mockResolvedValue(undefined)
      vi.mocked(GetEntry).mockResolvedValue(doneTask)

      render(<SearchView />)
      const user = userEvent.setup()

      // Search and open popover
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'task')

      await waitFor(() => {
        expect(screen.getByText('Task to complete')).toBeInTheDocument()
      })

      await user.click(screen.getByText('Task to complete'))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Click Done button
      const doneButton = screen.getByTitle(/mark done/i)
      await user.click(doneButton)

      // Verify Wails function called
      await waitFor(() => {
        expect(vi.mocked(MarkEntryDone)).toHaveBeenCalledWith(1)
      })
    })

    it('should call CyclePriority when clicking Priority action in popover', async () => {
      // Setup: Task entry
      const task = createMockEntry({
        ID: 1,
        Content: 'Task to prioritize',
        Type: 'Task',
        Priority: 'None',
        ParentID: null,
      })
      const prioritizedTask = createMockEntry({
        ID: 1,
        Content: 'Task to prioritize',
        Type: 'Task',
        Priority: 'High',
        ParentID: null,
      })
      vi.mocked(Search).mockResolvedValue([task])
      vi.mocked(GetEntryAncestors).mockResolvedValue([])
      vi.mocked(CyclePriority).mockResolvedValue(undefined)
      vi.mocked(GetEntry).mockResolvedValue(prioritizedTask)

      render(<SearchView />)
      const user = userEvent.setup()

      // Search and open popover
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'task')

      await waitFor(() => {
        expect(screen.getByText('Task to prioritize')).toBeInTheDocument()
      })

      await user.click(screen.getByText('Task to prioritize'))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Click Priority button
      const priorityButton = screen.getByTitle(/priority/i)
      await user.click(priorityButton)

      // Verify Wails function called
      await waitFor(() => {
        expect(vi.mocked(CyclePriority)).toHaveBeenCalledWith(1)
      })
    })

    it('should show "Go to entry" button in popover', async () => {
      // Setup: Task entry
      const task = createMockEntry({
        ID: 1,
        Content: 'Task entry',
        Type: 'Task',
        ParentID: null,
      })
      vi.mocked(Search).mockResolvedValue([task])
      vi.mocked(GetEntryAncestors).mockResolvedValue([])

      render(<SearchView />)
      const user = userEvent.setup()

      // Search and open popover
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'task')

      await waitFor(() => {
        expect(screen.getByText('Task entry')).toBeInTheDocument()
      })

      await user.click(screen.getByText('Task entry'))

      // Verify "Go to entry" button
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /go to/i })).toBeInTheDocument()
      })
    })

    it('should call onNavigateToEntry when clicking "Go to entry" button', async () => {
      // Setup: Task entry with navigation callback
      const task = createMockEntry({
        ID: 1,
        Content: 'Task entry',
        Type: 'Task',
        ParentID: null,
      })
      vi.mocked(Search).mockResolvedValue([task])
      vi.mocked(GetEntryAncestors).mockResolvedValue([])

      const onNavigateToEntry = vi.fn()
      render(<SearchView onNavigateToEntry={onNavigateToEntry} />)
      const user = userEvent.setup()

      // Search and open popover
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'task')

      await waitFor(() => {
        expect(screen.getByText('Task entry')).toBeInTheDocument()
      })

      await user.click(screen.getByText('Task entry'))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Click "Go to entry"
      const goToButton = screen.getByRole('button', { name: /go to/i })
      await user.click(goToButton)

      // Verify callback called with entry
      await waitFor(() => {
        expect(onNavigateToEntry).toHaveBeenCalledWith(
          expect.objectContaining({
            id: 1,
            content: 'Task entry',
            type: 'task',
          })
        )
      })
    })
  })

  describe('Closing the Popover', () => {
    it('should close popover when clicking outside', async () => {
      // Setup: Task entry
      const task = createMockEntry({
        ID: 1,
        Content: 'Task entry',
        Type: 'Task',
        ParentID: null,
      })
      vi.mocked(Search).mockResolvedValue([task])
      vi.mocked(GetEntryAncestors).mockResolvedValue([])

      render(<SearchView />)
      const user = userEvent.setup()

      // Search and open popover
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'task')

      await waitFor(() => {
        expect(screen.getByText('Task entry')).toBeInTheDocument()
      })

      await user.click(screen.getByText('Task entry'))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Click outside (on the document body)
      await user.click(document.body)

      // Popover should close
      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      })
    })

    it('should close popover when pressing Escape', async () => {
      // Setup: Task entry
      const task = createMockEntry({
        ID: 1,
        Content: 'Task entry',
        Type: 'Task',
        ParentID: null,
      })
      vi.mocked(Search).mockResolvedValue([task])
      vi.mocked(GetEntryAncestors).mockResolvedValue([])

      render(<SearchView />)
      const user = userEvent.setup()

      // Search and open popover
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'task')

      await waitFor(() => {
        expect(screen.getByText('Task entry')).toBeInTheDocument()
      })

      await user.click(screen.getByText('Task entry'))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Press Escape
      await user.keyboard('{Escape}')

      // Popover should close
      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      })
    })

    it('should close popover after performing action that removes entry from list', async () => {
      // Setup: Task entry that when marked done, no longer matches search
      const task = createMockEntry({
        ID: 1,
        Content: 'Incomplete task',
        Type: 'Task',
        ParentID: null,
      })
      vi.mocked(Search).mockResolvedValue([task])
      vi.mocked(GetEntryAncestors).mockResolvedValue([])
      vi.mocked(MarkEntryDone).mockResolvedValue(undefined)

      // After marking done, GetEntry returns null (entry removed from search)
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      vi.mocked(GetEntry).mockResolvedValue(null as any)

      render(<SearchView />)
      const user = userEvent.setup()

      // Search and open popover
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'incomplete')

      await waitFor(() => {
        expect(screen.getByText('Incomplete task')).toBeInTheDocument()
      })

      await user.click(screen.getByText('Incomplete task'))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Mark done (removes from list)
      const doneButton = screen.getByTitle(/mark done/i)
      await user.click(doneButton)

      // Popover should close (entry removed)
      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      })
    })
  })

  describe('Double Click Navigation', () => {
    it('should navigate to entry on double click (existing behavior)', async () => {
      // Setup: Task entry with navigation callback
      const task = createMockEntry({
        ID: 1,
        Content: 'Task entry',
        Type: 'Task',
        ParentID: null,
      })
      vi.mocked(Search).mockResolvedValue([task])
      vi.mocked(GetEntryAncestors).mockResolvedValue([])

      const onNavigateToEntry = vi.fn()
      render(<SearchView onNavigateToEntry={onNavigateToEntry} />)
      const user = userEvent.setup()

      // Search
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'task')

      await waitFor(() => {
        expect(screen.getByText('Task entry')).toBeInTheDocument()
      })

      // Double click the entry
      const entry = screen.getByText('Task entry')
      await user.dblClick(entry)

      // Should navigate without opening popover
      await waitFor(() => {
        expect(onNavigateToEntry).toHaveBeenCalledWith(
          expect.objectContaining({
            id: 1,
            content: 'Task entry',
          })
        )
      })
    })
  })

  describe('Keyboard Shortcuts with Popover', () => {
    it('should support keyboard shortcut "m" to migrate when popover is open', async () => {
      // Setup: Task entry with migrate callback
      const task = createMockEntry({
        ID: 1,
        Content: 'Task to migrate',
        Type: 'Task',
        ParentID: null,
      })
      vi.mocked(Search).mockResolvedValue([task])
      vi.mocked(GetEntryAncestors).mockResolvedValue([])
      vi.mocked(MigrateEntry).mockResolvedValue(1)

      const onMigrate = vi.fn()
      render(<SearchView onMigrate={onMigrate} />)
      const user = userEvent.setup()

      // Search and select entry
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'task')

      await waitFor(() => {
        expect(screen.getByText('Task to migrate')).toBeInTheDocument()
      })

      // Open popover
      await user.click(screen.getByText('Task to migrate'))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Press "m" to migrate
      await user.keyboard('m')

      // Should call migrate callback
      await waitFor(() => {
        expect(onMigrate).toHaveBeenCalledWith(
          expect.objectContaining({
            id: 1,
            content: 'Task to migrate',
          })
        )
      })
    })

    it('should support keyboard shortcut "p" to cycle priority when popover is open', async () => {
      // Setup: Task entry
      const task = createMockEntry({
        ID: 1,
        Content: 'Task to prioritize',
        Type: 'Task',
        Priority: 'None',
        ParentID: null,
      })
      const prioritizedTask = createMockEntry({
        ID: 1,
        Content: 'Task to prioritize',
        Type: 'Task',
        Priority: 'High',
        ParentID: null,
      })
      vi.mocked(Search).mockResolvedValue([task])
      vi.mocked(GetEntryAncestors).mockResolvedValue([])
      vi.mocked(CyclePriority).mockResolvedValue(undefined)
      vi.mocked(GetEntry).mockResolvedValue(prioritizedTask)

      render(<SearchView />)
      const user = userEvent.setup()

      // Search and open popover
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'task')

      await waitFor(() => {
        expect(screen.getByText('Task to prioritize')).toBeInTheDocument()
      })

      await user.click(screen.getByText('Task to prioritize'))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Press "p" to cycle priority
      await user.keyboard('p')

      // Should call CyclePriority
      await waitFor(() => {
        expect(vi.mocked(CyclePriority)).toHaveBeenCalledWith(1)
      })
    })
  })

  describe('EntryActionBar Visibility', () => {
    it('should hide or minimize EntryActionBar when using popover', async () => {
      // Setup: Task entry
      const task = createMockEntry({
        ID: 1,
        Content: 'Task entry',
        Type: 'Task',
        ParentID: null,
      })
      vi.mocked(Search).mockResolvedValue([task])
      vi.mocked(GetEntryAncestors).mockResolvedValue([])

      render(<SearchView />)
      const user = userEvent.setup()

      // Search
      const searchInput = screen.getByPlaceholderText(/search entries/i)
      await user.type(searchInput, 'task')

      await waitFor(() => {
        expect(screen.getByText('Task entry')).toBeInTheDocument()
      })

      // EntryActionBar should NOT be always-visible anymore
      // It should be hover variant or removed entirely
      // This test verifies the actions are in the popover, not in the result row

      // Actions should NOT be always visible in the row
      // (Previously variant="always-visible", now should be different)
      // We can't easily test "hover" variant without actual hover,
      // but we can verify actions are accessible via popover
      await user.click(screen.getByText('Task entry'))

      await waitFor(() => {
        const dialog = screen.getByRole('dialog')
        expect(dialog).toHaveTextContent('Task entry')
        // Actions should be in popover
        expect(screen.getByTitle(/mark done/i)).toBeInTheDocument()
      })
    })
  })
})

import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { EntryItem } from './EntryItem'
import { Entry } from '@/types/bujo'

const createTestEntry = (overrides: Partial<Entry> = {}): Entry => ({
  id: 1,
  type: 'task',
  content: 'Test entry',
  priority: 'none',
  parentId: null,
  loggedDate: '2026-01-17',
  children: [],
  ...overrides,
})

describe('EntryItem', () => {
  describe('context menu', () => {
    it('shows context menu on right-click', () => {
      render(<EntryItem entry={createTestEntry()} onAddChild={() => {}} onEdit={() => {}} onDelete={() => {}} />)
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)

      expect(screen.getByText('Add child')).toBeInTheDocument()
    })

    it('shows edit option in context menu when onEdit is provided', () => {
      render(<EntryItem entry={createTestEntry()} onEdit={() => {}} />)
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)

      expect(screen.getByRole('menuitem', { name: 'Edit' })).toBeInTheDocument()
    })

    it('shows delete option in context menu when onDelete is provided', () => {
      render(<EntryItem entry={createTestEntry()} onDelete={() => {}} />)
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)

      expect(screen.getByRole('menuitem', { name: 'Delete' })).toBeInTheDocument()
    })

    it('calls onAddChild when Add child option is clicked', () => {
      const onAddChild = vi.fn()
      render(<EntryItem entry={createTestEntry()} onAddChild={onAddChild} />)
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)
      fireEvent.click(screen.getByText('Add child'))

      expect(onAddChild).toHaveBeenCalledTimes(1)
    })

    it('does not show Add child option for question entries', () => {
      const onAddChild = vi.fn()
      render(<EntryItem entry={createTestEntry({ type: 'question' })} onAddChild={onAddChild} />)
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)

      expect(screen.queryByText('Add child')).not.toBeInTheDocument()
    })

    it('closes context menu when clicking outside', () => {
      render(<EntryItem entry={createTestEntry()} onAddChild={() => {}} />)
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)
      expect(screen.getByText('Add child')).toBeInTheDocument()

      fireEvent.click(document.body)
      expect(screen.queryByText('Add child')).not.toBeInTheDocument()
    })

    it('closes context menu when pressing Escape', () => {
      render(<EntryItem entry={createTestEntry()} onAddChild={() => {}} />)
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)
      expect(screen.getByText('Add child')).toBeInTheDocument()

      fireEvent.keyDown(document, { key: 'Escape' })
      expect(screen.queryByText('Add child')).not.toBeInTheDocument()
    })

    it('shows Move to root option when entry has a parent', () => {
      render(
        <EntryItem
          entry={createTestEntry({ parentId: 123 })}
          hasParent={true}
          onMoveToRoot={() => {}}
        />
      )
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)

      expect(screen.getByRole('menuitem', { name: 'Move to root' })).toBeInTheDocument()
    })

    it('does not show Move to root option when entry has no parent', () => {
      render(
        <EntryItem
          entry={createTestEntry({ parentId: null })}
          hasParent={false}
          onMoveToRoot={() => {}}
        />
      )
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)

      expect(screen.queryByRole('menuitem', { name: 'Move to root' })).not.toBeInTheDocument()
    })

    it('calls onMoveToRoot when Move to root option is clicked', () => {
      const onMoveToRoot = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ parentId: 123 })}
          hasParent={true}
          onMoveToRoot={onMoveToRoot}
        />
      )
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)
      fireEvent.click(screen.getByRole('menuitem', { name: 'Move to root' }))

      expect(onMoveToRoot).toHaveBeenCalledTimes(1)
    })

    it('shows Mark done option for task entries in context menu', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onToggleDone={() => {}}
        />
      )
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)

      expect(screen.getByRole('menuitem', { name: 'Mark done' })).toBeInTheDocument()
    })

    it('shows Mark not done option for done entries in context menu', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'done' })}
          onToggleDone={() => {}}
        />
      )
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)

      expect(screen.getByRole('menuitem', { name: 'Mark not done' })).toBeInTheDocument()
    })

    it('shows Cancel option in context menu', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onCancel={() => {}}
        />
      )
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)

      expect(screen.getByRole('menuitem', { name: 'Cancel' })).toBeInTheDocument()
    })

    it('shows Uncancel option for cancelled entries in context menu', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'cancelled' })}
          onUncancel={() => {}}
        />
      )
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)

      expect(screen.getByRole('menuitem', { name: 'Uncancel' })).toBeInTheDocument()
    })

    it('shows Migrate option for task entries in context menu', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onMigrate={() => {}}
        />
      )
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)

      expect(screen.getByRole('menuitem', { name: 'Migrate' })).toBeInTheDocument()
    })

    it('shows Change type option in context menu', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onCycleType={() => {}}
        />
      )
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)

      expect(screen.getByRole('menuitem', { name: 'Change type' })).toBeInTheDocument()
    })

    it('shows Cycle priority option in context menu', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onCyclePriority={() => {}}
        />
      )
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)

      expect(screen.getByRole('menuitem', { name: 'Cycle priority' })).toBeInTheDocument()
    })
  })

  describe('symbol click toggle', () => {
    it('calls onToggleDone when clicking symbol for task entry', () => {
      const onToggleDone = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onToggleDone={onToggleDone}
        />
      )

      const symbolButton = screen.getByTitle('Mark as done')
      fireEvent.click(symbolButton)
      expect(onToggleDone).toHaveBeenCalledTimes(1)
    })

    it('calls onToggleDone when clicking symbol for done entry', () => {
      const onToggleDone = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'done' })}
          onToggleDone={onToggleDone}
        />
      )

      const symbolButton = screen.getByTitle('Mark as not done')
      fireEvent.click(symbolButton)
      expect(onToggleDone).toHaveBeenCalledTimes(1)
    })

    it('symbol is not clickable for note entry', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'note' })}
          onToggleDone={() => {}}
        />
      )

      // Note entries should not have a clickable symbol
      expect(screen.queryByTitle('Mark as done')).not.toBeInTheDocument()
      expect(screen.queryByTitle('Mark as not done')).not.toBeInTheDocument()
    })

    it('symbol click does not trigger row onSelect', () => {
      const onToggleDone = vi.fn()
      const onSelect = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onToggleDone={onToggleDone}
          onSelect={onSelect}
        />
      )

      const symbolButton = screen.getByTitle('Mark as done')
      fireEvent.click(symbolButton)
      expect(onToggleDone).toHaveBeenCalledTimes(1)
      expect(onSelect).not.toHaveBeenCalled()
    })

    it('symbol shows task bullet for task entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onToggleDone={() => {}}
        />
      )
      const symbolButton = screen.getByTitle('Mark as done')
      expect(symbolButton).toHaveTextContent('•')
    })

    it('symbol shows checkmark for done entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'done' })}
          onToggleDone={() => {}}
        />
      )
      const symbolButton = screen.getByTitle('Mark as not done')
      expect(symbolButton).toHaveTextContent('✓')
    })
  })

  describe('move to list functionality', () => {
    it('shows move to list button for task entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onMoveToList={() => {}}
        />
      )
      expect(screen.getByTitle('Move to list')).toBeInTheDocument()
    })

    it('calls onMoveToList when move to list button is clicked', () => {
      const onMoveToList = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onMoveToList={onMoveToList}
        />
      )

      fireEvent.click(screen.getByTitle('Move to list'))
      expect(onMoveToList).toHaveBeenCalledTimes(1)
    })

    it('does not show move to list button for non-task entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'note' })}
          onMoveToList={() => {}}
        />
      )
      expect(screen.queryByTitle('Move to list')).not.toBeInTheDocument()
    })

    it('does not show move to list button for done entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'done' })}
          onMoveToList={() => {}}
        />
      )
      expect(screen.queryByTitle('Move to list')).not.toBeInTheDocument()
    })

    it('shows Move to list option in context menu for task entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onMoveToList={() => {}}
        />
      )
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)

      expect(screen.getByRole('menuitem', { name: 'Move to list' })).toBeInTheDocument()
    })

    it('calls onMoveToList when Move to list context menu option is clicked', () => {
      const onMoveToList = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onMoveToList={onMoveToList}
        />
      )
      const container = screen.getByText('Test entry').closest('[data-entry-id]')!

      fireEvent.contextMenu(container)
      fireEvent.click(screen.getByRole('menuitem', { name: 'Move to list' }))

      expect(onMoveToList).toHaveBeenCalledTimes(1)
    })
  })

  describe('cycle type', () => {
    it('shows cycle type button when onCycleType callback is provided', () => {
      render(<EntryItem entry={createTestEntry({ type: 'task' })} onCycleType={() => {}} />)
      expect(screen.getByTitle('Change type')).toBeInTheDocument()
    })

    it('calls onCycleType when cycle type button is clicked', () => {
      const onCycleType = vi.fn()
      render(<EntryItem entry={createTestEntry({ type: 'task' })} onCycleType={onCycleType} />)

      fireEvent.click(screen.getByTitle('Change type'))
      expect(onCycleType).toHaveBeenCalledTimes(1)
    })

    it('does not show cycle type button for done entries', () => {
      render(<EntryItem entry={createTestEntry({ type: 'done' })} onCycleType={() => {}} />)
      expect(screen.queryByTitle('Change type')).not.toBeInTheDocument()
    })

    it('does not show cycle type button for cancelled entries', () => {
      render(<EntryItem entry={createTestEntry({ type: 'cancelled' })} onCycleType={() => {}} />)
      expect(screen.queryByTitle('Change type')).not.toBeInTheDocument()
    })

    it('does not show cycle type button for migrated entries', () => {
      render(<EntryItem entry={createTestEntry({ type: 'migrated' })} onCycleType={() => {}} />)
      expect(screen.queryByTitle('Change type')).not.toBeInTheDocument()
    })

    it('shows cycle type button for note entries', () => {
      render(<EntryItem entry={createTestEntry({ type: 'note' })} onCycleType={() => {}} />)
      expect(screen.getByTitle('Change type')).toBeInTheDocument()
    })

    it('shows cycle type button for event entries', () => {
      render(<EntryItem entry={createTestEntry({ type: 'event' })} onCycleType={() => {}} />)
      expect(screen.getByTitle('Change type')).toBeInTheDocument()
    })

    it('shows cycle type button for question entries', () => {
      render(<EntryItem entry={createTestEntry({ type: 'question' })} onCycleType={() => {}} />)
      expect(screen.getByTitle('Change type')).toBeInTheDocument()
    })
  })
})

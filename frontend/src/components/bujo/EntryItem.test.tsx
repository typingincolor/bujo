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
  it('renders entry content', () => {
    render(<EntryItem entry={createTestEntry({ content: 'My task' })} />)
    expect(screen.getByText('My task')).toBeInTheDocument()
  })

  it('aligns symbol and text vertically centered', () => {
    render(<EntryItem entry={createTestEntry({ content: 'My task' })} />)
    const container = screen.getByText('My task').closest('[data-entry-id]')
    expect(container).toHaveClass('items-center')
  })

  it('shows edit button on hover', () => {
    render(<EntryItem entry={createTestEntry()} onEdit={() => {}} />)
    const button = screen.getByTitle('Edit entry')
    expect(button).toBeInTheDocument()
  })

  it('shows delete button on hover', () => {
    render(<EntryItem entry={createTestEntry()} onDelete={() => {}} />)
    const button = screen.getByTitle('Delete entry')
    expect(button).toBeInTheDocument()
  })

  it('calls onEdit when edit button is clicked', () => {
    const onEdit = vi.fn()
    render(<EntryItem entry={createTestEntry()} onEdit={onEdit} />)

    fireEvent.click(screen.getByTitle('Edit entry'))
    expect(onEdit).toHaveBeenCalledTimes(1)
  })

  it('calls onDelete when delete button is clicked', () => {
    const onDelete = vi.fn()
    render(<EntryItem entry={createTestEntry()} onDelete={onDelete} />)

    fireEvent.click(screen.getByTitle('Delete entry'))
    expect(onDelete).toHaveBeenCalledTimes(1)
  })

  it('does not render edit button when onEdit is not provided', () => {
    render(<EntryItem entry={createTestEntry()} />)
    expect(screen.queryByTitle('Edit entry')).not.toBeInTheDocument()
  })

  it('does not render delete button when onDelete is not provided', () => {
    render(<EntryItem entry={createTestEntry()} />)
    expect(screen.queryByTitle('Delete entry')).not.toBeInTheDocument()
  })

  it('stops propagation when clicking edit button', () => {
    const onEdit = vi.fn()
    const onToggleDone = vi.fn()
    render(
      <EntryItem
        entry={createTestEntry()}
        onEdit={onEdit}
        onToggleDone={onToggleDone}
      />
    )

    fireEvent.click(screen.getByTitle('Edit entry'))
    expect(onEdit).toHaveBeenCalledTimes(1)
    expect(onToggleDone).not.toHaveBeenCalled()
  })

  it('stops propagation when clicking delete button', () => {
    const onDelete = vi.fn()
    const onToggleDone = vi.fn()
    render(
      <EntryItem
        entry={createTestEntry()}
        onDelete={onDelete}
        onToggleDone={onToggleDone}
      />
    )

    fireEvent.click(screen.getByTitle('Delete entry'))
    expect(onDelete).toHaveBeenCalledTimes(1)
    expect(onToggleDone).not.toHaveBeenCalled()
  })

  describe('question/answer entry types', () => {
    it('renders question entry with ? symbol', () => {
      render(<EntryItem entry={createTestEntry({ type: 'question', content: 'What is TDD?' })} />)
      expect(screen.getByText('?')).toBeInTheDocument()
      expect(screen.getByText('What is TDD?')).toBeInTheDocument()
    })

    it('renders answered entry with ★ symbol', () => {
      render(<EntryItem entry={createTestEntry({ type: 'answered', content: 'What is TDD?' })} />)
      expect(screen.getByText('★')).toBeInTheDocument()
      expect(screen.getByText('What is TDD?')).toBeInTheDocument()
    })

    it('renders answer entry with ↳ symbol', () => {
      render(<EntryItem entry={createTestEntry({ type: 'answer', content: 'Test-Driven Development' })} />)
      expect(screen.getByText('↳')).toBeInTheDocument()
      expect(screen.getByText('Test-Driven Development')).toBeInTheDocument()
    })

    it('question entries are not toggleable', () => {
      const onToggleDone = vi.fn()
      render(<EntryItem entry={createTestEntry({ type: 'question' })} onToggleDone={onToggleDone} />)

      const container = screen.getByText('Test entry').closest('[data-entry-id]')
      fireEvent.click(container!)
      expect(onToggleDone).not.toHaveBeenCalled()
    })

    it('answered entries are not toggleable', () => {
      const onToggleDone = vi.fn()
      render(<EntryItem entry={createTestEntry({ type: 'answered' })} onToggleDone={onToggleDone} />)

      const container = screen.getByText('Test entry').closest('[data-entry-id]')
      fireEvent.click(container!)
      expect(onToggleDone).not.toHaveBeenCalled()
    })

    it('answer entries are not toggleable', () => {
      const onToggleDone = vi.fn()
      render(<EntryItem entry={createTestEntry({ type: 'answer' })} onToggleDone={onToggleDone} />)

      const container = screen.getByText('Test entry').closest('[data-entry-id]')
      fireEvent.click(container!)
      expect(onToggleDone).not.toHaveBeenCalled()
    })

    it('shows answer button for question entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'question', content: 'What is TDD?' })}
          onAnswer={() => {}}
        />
      )
      expect(screen.getByTitle('Answer question')).toBeInTheDocument()
    })

    it('calls onAnswer when answer button is clicked', () => {
      const onAnswer = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'question', content: 'What is TDD?' })}
          onAnswer={onAnswer}
        />
      )

      fireEvent.click(screen.getByTitle('Answer question'))
      expect(onAnswer).toHaveBeenCalledTimes(1)
    })

    it('does not show answer button for non-question entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onAnswer={() => {}}
        />
      )
      expect(screen.queryByTitle('Answer question')).not.toBeInTheDocument()
    })
  })

  describe('cancel/uncancel functionality', () => {
    it('shows cancel button for task entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onCancel={() => {}}
        />
      )
      expect(screen.getByTitle('Cancel entry')).toBeInTheDocument()
    })

    it('shows uncancel button for cancelled entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'cancelled' })}
          onUncancel={() => {}}
        />
      )
      expect(screen.getByTitle('Uncancel entry')).toBeInTheDocument()
    })

    it('calls onCancel when cancel button is clicked', () => {
      const onCancel = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onCancel={onCancel}
        />
      )

      fireEvent.click(screen.getByTitle('Cancel entry'))
      expect(onCancel).toHaveBeenCalledTimes(1)
    })

    it('calls onUncancel when uncancel button is clicked', () => {
      const onUncancel = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'cancelled' })}
          onUncancel={onUncancel}
        />
      )

      fireEvent.click(screen.getByTitle('Uncancel entry'))
      expect(onUncancel).toHaveBeenCalledTimes(1)
    })

    it('renders cancelled entry with strikethrough style', () => {
      render(<EntryItem entry={createTestEntry({ type: 'cancelled', content: 'Cancelled task' })} />)
      const content = screen.getByText('Cancelled task')
      expect(content).toHaveClass('line-through')
    })

    it('does not show cancel button for cancelled entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'cancelled' })}
          onCancel={() => {}}
        />
      )
      expect(screen.queryByTitle('Cancel entry')).not.toBeInTheDocument()
    })

    it('does not show uncancel button for non-cancelled entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onUncancel={() => {}}
        />
      )
      expect(screen.queryByTitle('Uncancel entry')).not.toBeInTheDocument()
    })
  })

  describe('priority functionality', () => {
    it('shows priority button for task entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onCyclePriority={() => {}}
        />
      )
      expect(screen.getByTitle('Cycle priority')).toBeInTheDocument()
    })

    it('cycle priority button shows Flag icon', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onCyclePriority={() => {}}
        />
      )
      const button = screen.getByTitle('Cycle priority')
      const icon = button.querySelector('svg')
      expect(icon).toBeInTheDocument()
    })

    it('calls onCyclePriority when priority button is clicked', () => {
      const onCyclePriority = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onCyclePriority={onCyclePriority}
        />
      )

      fireEvent.click(screen.getByTitle('Cycle priority'))
      expect(onCyclePriority).toHaveBeenCalledTimes(1)
    })

    it('displays priority indicator for high priority', () => {
      render(<EntryItem entry={createTestEntry({ type: 'task', priority: 'high' })} />)
      expect(screen.getByText('!!!')).toBeInTheDocument()
    })

    it('displays priority indicator for medium priority', () => {
      render(<EntryItem entry={createTestEntry({ type: 'task', priority: 'medium' })} />)
      expect(screen.getByText('!!')).toBeInTheDocument()
    })

    it('displays priority indicator for low priority', () => {
      render(<EntryItem entry={createTestEntry({ type: 'task', priority: 'low' })} />)
      expect(screen.getByText('!')).toBeInTheDocument()
    })

    it('does not display priority indicator for none priority', () => {
      render(<EntryItem entry={createTestEntry({ type: 'task', priority: 'none' })} />)
      // None priority shouldn't show any indicator
      expect(screen.queryByText('!')).not.toBeInTheDocument()
      expect(screen.queryByText('!!')).not.toBeInTheDocument()
      expect(screen.queryByText('!!!')).not.toBeInTheDocument()
    })
  })

  describe('selection functionality', () => {
    it('calls onSelect when entry row is clicked', () => {
      const onSelect = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ id: 42, content: 'Click me' })}
          onSelect={onSelect}
        />
      )

      const container = screen.getByText('Click me').closest('[data-entry-id]')
      fireEvent.click(container!)
      expect(onSelect).toHaveBeenCalledTimes(1)
    })

    it('calls onSelect for note entries when clicked', () => {
      const onSelect = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'note', content: 'A note' })}
          onSelect={onSelect}
        />
      )

      const container = screen.getByText('A note').closest('[data-entry-id]')
      fireEvent.click(container!)
      expect(onSelect).toHaveBeenCalledTimes(1)
    })

    it('calls only onSelect (not onToggleDone) when clicking task entry', () => {
      const onSelect = vi.fn()
      const onToggleDone = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task', content: 'A task' })}
          onSelect={onSelect}
          onToggleDone={onToggleDone}
        />
      )

      const container = screen.getByText('A task').closest('[data-entry-id]')
      fireEvent.click(container!)
      expect(onSelect).toHaveBeenCalledTimes(1)
      expect(onToggleDone).not.toHaveBeenCalled()
    })

    it('shows tick button for task entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onToggleDone={() => {}}
        />
      )
      expect(screen.getByTitle('Mark as done')).toBeInTheDocument()
    })

    it('shows untick button for done entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'done' })}
          onToggleDone={() => {}}
        />
      )
      expect(screen.getByTitle('Mark as not done')).toBeInTheDocument()
    })

    it('shows checkmark symbol in untick button for done entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'done' })}
          onToggleDone={() => {}}
        />
      )
      const untickButton = screen.getByTitle('Mark as not done')
      expect(untickButton).toHaveTextContent('✓')
    })

    it('calls onToggleDone when tick button is clicked', () => {
      const onToggleDone = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onToggleDone={onToggleDone}
        />
      )

      fireEvent.click(screen.getByTitle('Mark as done'))
      expect(onToggleDone).toHaveBeenCalledTimes(1)
    })

    it('calls onToggleDone when untick button is clicked', () => {
      const onToggleDone = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'done' })}
          onToggleDone={onToggleDone}
        />
      )

      fireEvent.click(screen.getByTitle('Mark as not done'))
      expect(onToggleDone).toHaveBeenCalledTimes(1)
    })

    it('does not show tick button for note entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'note' })}
          onToggleDone={() => {}}
        />
      )
      expect(screen.queryByTitle('Mark as done')).not.toBeInTheDocument()
      expect(screen.queryByTitle('Mark as not done')).not.toBeInTheDocument()
    })
  })

  describe('migration functionality', () => {
    it('shows migrate button for task entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onMigrate={() => {}}
        />
      )
      expect(screen.getByTitle('Migrate entry')).toBeInTheDocument()
    })

    it('calls onMigrate when migrate button is clicked', () => {
      const onMigrate = vi.fn()
      render(
        <EntryItem
          entry={createTestEntry({ type: 'task' })}
          onMigrate={onMigrate}
        />
      )

      fireEvent.click(screen.getByTitle('Migrate entry'))
      expect(onMigrate).toHaveBeenCalledTimes(1)
    })

    it('does not show migrate button for non-task entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'note' })}
          onMigrate={() => {}}
        />
      )
      expect(screen.queryByTitle('Migrate entry')).not.toBeInTheDocument()
    })

    it('does not show migrate button for done entries', () => {
      render(
        <EntryItem
          entry={createTestEntry({ type: 'done' })}
          onMigrate={() => {}}
        />
      )
      expect(screen.queryByTitle('Migrate entry')).not.toBeInTheDocument()
    })
  })

  describe('visual styling', () => {
    it('renders done entries with success color (not strikethrough)', () => {
      render(<EntryItem entry={createTestEntry({ type: 'done', content: 'Done task' })} />)
      const content = screen.getByText('Done task')
      expect(content).toHaveClass('text-bujo-done')
      expect(content).not.toHaveClass('line-through')
    })
  })

  describe('hover and selection highlighting', () => {
    it('shows hover highlight on non-selected items', () => {
      render(<EntryItem entry={createTestEntry()} isSelected={false} />)
      const container = screen.getByText('Test entry').closest('[data-entry-id]')
      expect(container).toHaveClass('hover:bg-secondary/50')
    })

    it('does not show hover highlight on selected items', () => {
      render(<EntryItem entry={createTestEntry()} isSelected={true} />)
      const container = screen.getByText('Test entry').closest('[data-entry-id]')
      expect(container).not.toHaveClass('hover:bg-secondary/50')
    })

    it('shows selection highlight on selected items', () => {
      render(<EntryItem entry={createTestEntry()} isSelected={true} />)
      const container = screen.getByText('Test entry').closest('[data-entry-id]')
      expect(container).toHaveClass('bg-primary/10')
      expect(container).toHaveClass('ring-1')
    })
  })

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

import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { JournalSidebar } from './JournalSidebar'
import { Entry } from '@/types/bujo'
import { useState } from 'react'

function createTestEntry(overrides: Partial<Entry> = {}): Entry {
  return {
    id: 1,
    content: 'Test entry',
    type: 'task',
    priority: 'none',
    parentId: null,
    loggedDate: '2026-01-25T10:00:00Z',
    ...overrides,
  }
}

function buildContextTree(entry: Entry, entries: Entry[]): Entry[] {
  const entriesById = new Map(entries.map(e => [e.id, e]))
  const path: Entry[] = []
  let current: Entry | undefined = entry

  while (current) {
    path.unshift(current)
    if (current.parentId === null) break
    current = entriesById.get(current.parentId)
  }

  // Return full path including the entry itself
  return path
}

// Integration wrapper that mimics App's selection behavior
function SelectionIntegrationWrapper({
  overdueEntries,
  allEntries,
}: {
  overdueEntries: Entry[]
  allEntries: Entry[]
}) {
  const [selectedEntry, setSelectedEntry] = useState<Entry | null>(null)
  const now = new Date('2026-01-25T12:00:00Z')

  const ancestors = selectedEntry
    ? buildContextTree(selectedEntry, allEntries)
    : []

  return (
    <div>
      <JournalSidebar
        overdueEntries={overdueEntries}
        now={now}
        selectedEntry={selectedEntry ?? undefined}
        contextTree={ancestors}
        onSelectEntry={setSelectedEntry}
      />
      <div data-testid="selection-state">
        {selectedEntry ? `Selected: ${selectedEntry.content}` : 'No selection'}
      </div>
    </div>
  )
}

describe('JournalSidebar Selection Integration', () => {
  it('updates selection state when clicking overdue item', async () => {
    const user = userEvent.setup()
    const overdueEntries = [
      createTestEntry({ id: 1, content: 'Overdue task 1' }),
      createTestEntry({ id: 2, content: 'Overdue task 2' }),
    ]

    render(
      <SelectionIntegrationWrapper
        overdueEntries={overdueEntries}
        allEntries={overdueEntries}
      />
    )

    expect(screen.getByTestId('selection-state')).toHaveTextContent('No selection')

    // Section is expanded by default
    await user.click(screen.getByText('Overdue task 1'))

    expect(screen.getByTestId('selection-state')).toHaveTextContent('Selected: Overdue task 1')
  })

  it('shows ancestor hierarchy in context section when entry with parent selected', async () => {
    const user = userEvent.setup()
    const grandparent = createTestEntry({ id: 1, content: 'Grandparent', parentId: null })
    const parent = createTestEntry({ id: 2, content: 'Parent', parentId: 1 })
    const child = createTestEntry({ id: 3, content: 'Child task', parentId: 2 })

    const allEntries = [grandparent, parent, child]
    const overdueEntries = [child]

    render(
      <SelectionIntegrationWrapper
        overdueEntries={overdueEntries}
        allEntries={allEntries}
      />
    )

    // Section is expanded by default
    await user.click(screen.getByText('Child task'))

    // Context section should show ancestor hierarchy
    const contextSection = screen.getByTestId('context-section')
    expect(contextSection).toHaveTextContent('Grandparent')
    expect(contextSection).toHaveTextContent('Parent')
    expect(contextSection).toHaveTextContent('Child task')
  })

  it('changes selection when clicking different overdue item', async () => {
    const user = userEvent.setup()
    const overdueEntries = [
      createTestEntry({ id: 1, content: 'First task' }),
      createTestEntry({ id: 2, content: 'Second task' }),
    ]

    render(
      <SelectionIntegrationWrapper
        overdueEntries={overdueEntries}
        allEntries={overdueEntries}
      />
    )

    // Section is expanded by default
    await user.click(screen.getByText('First task'))
    expect(screen.getByTestId('selection-state')).toHaveTextContent('Selected: First task')

    await user.click(screen.getByText('Second task'))
    expect(screen.getByTestId('selection-state')).toHaveTextContent('Selected: Second task')
  })

  it('highlights selected entry in overdue list', async () => {
    const user = userEvent.setup()
    const overdueEntries = [
      createTestEntry({ id: 1, content: 'Overdue task' }),
    ]

    render(
      <SelectionIntegrationWrapper
        overdueEntries={overdueEntries}
        allEntries={overdueEntries}
      />
    )

    // Section is expanded by default
    const button = screen.getByRole('button', { name: /overdue task/i })
    const container = button.parentElement!
    expect(container).not.toHaveClass('bg-accent')

    await user.click(button)

    expect(container).toHaveClass('bg-accent')
  })
})

import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { EntryContextPopover } from '../EntryContextPopover'
import { Entry } from '@/types/bujo'

const mockEntries: Entry[] = [
  { id: 1, content: 'Root event', type: 'event', loggedDate: '2026-01-15', priority: 'none', parentId: null, children: [] },
  { id: 2, content: 'Child task', type: 'task', loggedDate: '2026-01-15', priority: 'none', parentId: 1, children: [] },
]

describe('EntryContextPopover', () => {
  it('renders trigger and opens popover on click', async () => {
    render(
      <EntryContextPopover
        entry={mockEntries[1]}
        entries={mockEntries}
        onAction={vi.fn()}
        onNavigate={vi.fn()}
      >
        <button>Click me</button>
      </EntryContextPopover>
    )

    await userEvent.click(screen.getByText('Click me'))

    expect(screen.getByTestId('entry-context-popover')).toBeInTheDocument()
    expect(screen.getByText('Child task')).toBeInTheDocument()
  })

  it('shows quick action buttons based on entry type', async () => {
    render(
      <EntryContextPopover
        entry={mockEntries[1]}
        entries={mockEntries}
        onAction={vi.fn()}
        onNavigate={vi.fn()}
      >
        <button>Click me</button>
      </EntryContextPopover>
    )

    await userEvent.click(screen.getByText('Click me'))

    expect(screen.getByRole('button', { name: /done/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /priority/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /migrate/i })).toBeInTheDocument()
  })

  it('calls onAction when quick action clicked', async () => {
    const onAction = vi.fn()
    render(
      <EntryContextPopover
        entry={mockEntries[1]}
        entries={mockEntries}
        onAction={onAction}
        onNavigate={vi.fn()}
      >
        <button>Click me</button>
      </EntryContextPopover>
    )

    await userEvent.click(screen.getByText('Click me'))
    await userEvent.click(screen.getByRole('button', { name: /done/i }))

    expect(onAction).toHaveBeenCalledWith(mockEntries[1], 'done')
  })

  it('calls onNavigate when "Go to entry" clicked', async () => {
    const onNavigate = vi.fn()
    render(
      <EntryContextPopover
        entry={mockEntries[1]}
        entries={mockEntries}
        onAction={vi.fn()}
        onNavigate={onNavigate}
      >
        <button>Click me</button>
      </EntryContextPopover>
    )

    await userEvent.click(screen.getByText('Click me'))
    await userEvent.click(screen.getByText('Go to entry'))

    expect(onNavigate).toHaveBeenCalledWith(mockEntries[1])
  })

  it('closes on Escape key', async () => {
    render(
      <EntryContextPopover
        entry={mockEntries[1]}
        entries={mockEntries}
        onAction={vi.fn()}
        onNavigate={vi.fn()}
      >
        <button>Click me</button>
      </EntryContextPopover>
    )

    await userEvent.click(screen.getByText('Click me'))
    expect(screen.getByTestId('entry-context-popover')).toBeInTheDocument()

    await userEvent.keyboard('{Escape}')

    expect(screen.queryByTestId('entry-context-popover')).not.toBeInTheDocument()
  })

  it('supports keyboard shortcuts for actions', async () => {
    const onAction = vi.fn()
    render(
      <EntryContextPopover
        entry={mockEntries[1]}
        entries={mockEntries}
        onAction={onAction}
        onNavigate={vi.fn()}
      >
        <button>Click me</button>
      </EntryContextPopover>
    )

    await userEvent.click(screen.getByText('Click me'))
    await userEvent.keyboard(' ') // Space for done

    expect(onAction).toHaveBeenCalledWith(mockEntries[1], 'done')
  })
})

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ListsView } from './ListsView'
import { BujoList } from '@/types/bujo'

vi.mock('@/wailsjs/go/wails/App', () => ({
  MarkListItemDone: vi.fn().mockResolvedValue(undefined),
  MarkListItemUndone: vi.fn().mockResolvedValue(undefined),
  AddListItem: vi.fn().mockResolvedValue(1),
  RemoveListItem: vi.fn().mockResolvedValue(undefined),
}))

import { AddListItem } from '@/wailsjs/go/wails/App'

const createTestList = (overrides: Partial<BujoList> = {}): BujoList => ({
  id: 1,
  name: 'Test List',
  items: [],
  doneCount: 0,
  totalCount: 0,
  ...overrides,
})

const createTestItem = (overrides = {}) => ({
  id: 1,
  content: 'Test Item',
  done: false,
  type: 'task' as const,
  ...overrides,
})

describe('ListsView - Add List Item', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows add item input when list is expanded', () => {
    // First list is auto-expanded
    render(<ListsView lists={[createTestList({ name: 'Shopping' })]} />)

    expect(screen.getByPlaceholderText(/add item/i)).toBeInTheDocument()
  })

  it('calls AddListItem binding when submitting new item', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    // First list is auto-expanded
    render(<ListsView lists={[createTestList({ id: 42, name: 'Shopping' })]} onListChanged={onListChanged} />)

    const input = screen.getByPlaceholderText(/add item/i)
    await user.type(input, 'Buy milk{Enter}')

    await waitFor(() => {
      expect(AddListItem).toHaveBeenCalledWith(42, 'Buy milk')
    })
  })

  it('calls onListChanged after adding item', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    // First list is auto-expanded
    render(<ListsView lists={[createTestList({ name: 'Shopping' })]} onListChanged={onListChanged} />)

    const input = screen.getByPlaceholderText(/add item/i)
    await user.type(input, 'Buy milk{Enter}')

    await waitFor(() => {
      expect(onListChanged).toHaveBeenCalled()
    })
  })

  it('clears input after adding item', async () => {
    const user = userEvent.setup()
    // First list is auto-expanded
    render(<ListsView lists={[createTestList({ name: 'Shopping' })]} />)

    const input = screen.getByPlaceholderText(/add item/i) as HTMLInputElement
    await user.type(input, 'Buy milk{Enter}')

    await waitFor(() => {
      expect(input.value).toBe('')
    })
  })

  it('does not add item with empty content', async () => {
    const user = userEvent.setup()
    // First list is auto-expanded
    render(<ListsView lists={[createTestList({ name: 'Shopping' })]} />)

    const input = screen.getByPlaceholderText(/add item/i)
    await user.type(input, '{Enter}')

    expect(AddListItem).not.toHaveBeenCalled()
  })

  it('does not show add input when list is collapsed', async () => {
    const user = userEvent.setup()
    // First list starts expanded, click to collapse
    render(<ListsView lists={[createTestList({ name: 'Shopping' })]} />)

    // First list is expanded, click to collapse
    await user.click(screen.getByText('Shopping'))

    expect(screen.queryByPlaceholderText(/add item/i)).not.toBeInTheDocument()
  })
})

describe('ListsView - Display Lists', () => {
  it('renders lists', () => {
    render(<ListsView lists={[createTestList({ name: 'Shopping' })]} />)
    expect(screen.getByText('Shopping')).toBeInTheDocument()
  })

  it('shows progress for list items', () => {
    render(<ListsView lists={[createTestList({
      name: 'Tasks',
      doneCount: 2,
      totalCount: 5,
      items: [
        createTestItem({ id: 1, done: true }),
        createTestItem({ id: 2, done: true }),
        createTestItem({ id: 3, done: false }),
        createTestItem({ id: 4, done: false }),
        createTestItem({ id: 5, done: false }),
      ]
    })]} />)
    expect(screen.getByText('2/5')).toBeInTheDocument()
  })
})

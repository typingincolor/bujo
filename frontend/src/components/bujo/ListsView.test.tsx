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
  CreateList: vi.fn().mockResolvedValue(1),
  DeleteList: vi.fn().mockResolvedValue(undefined),
  RenameList: vi.fn().mockResolvedValue(undefined),
  EditListItem: vi.fn().mockResolvedValue(undefined),
  CancelListItem: vi.fn().mockResolvedValue(undefined),
  UncancelListItem: vi.fn().mockResolvedValue(undefined),
}))

import { AddListItem, RemoveListItem, CreateList, DeleteList, RenameList, EditListItem, CancelListItem, UncancelListItem } from '@/wailsjs/go/wails/App'

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

describe('ListsView - Delete List Item', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows delete button on list items', () => {
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ content: 'Buy milk' })]
    })]} />)

    expect(screen.getByTitle('Delete item')).toBeInTheDocument()
  })

  it('calls RemoveListItem binding when delete button is clicked', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ id: 42, content: 'Buy milk' })]
    })]} onListChanged={onListChanged} />)

    await user.click(screen.getByTitle('Delete item'))

    await waitFor(() => {
      expect(RemoveListItem).toHaveBeenCalledWith(42)
    })
  })

  it('calls onListChanged after deleting item', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ content: 'Buy milk' })]
    })]} onListChanged={onListChanged} />)

    await user.click(screen.getByTitle('Delete item'))

    await waitFor(() => {
      expect(onListChanged).toHaveBeenCalled()
    })
  })
})

describe('ListsView - Create List', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows create list button', () => {
    render(<ListsView lists={[]} />)
    expect(screen.getByRole('button', { name: /new list/i })).toBeInTheDocument()
  })

  it('shows input when create list button is clicked', async () => {
    const user = userEvent.setup()
    render(<ListsView lists={[]} />)

    await user.click(screen.getByRole('button', { name: /new list/i }))

    expect(screen.getByPlaceholderText(/list name/i)).toBeInTheDocument()
  })

  it('calls CreateList binding when submitting new list', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    render(<ListsView lists={[]} onListChanged={onListChanged} />)

    await user.click(screen.getByRole('button', { name: /new list/i }))
    const input = screen.getByPlaceholderText(/list name/i)
    await user.type(input, 'Shopping{Enter}')

    await waitFor(() => {
      expect(CreateList).toHaveBeenCalledWith('Shopping')
    })
  })

  it('calls onListChanged after creating list', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    render(<ListsView lists={[]} onListChanged={onListChanged} />)

    await user.click(screen.getByRole('button', { name: /new list/i }))
    const input = screen.getByPlaceholderText(/list name/i)
    await user.type(input, 'Shopping{Enter}')

    await waitFor(() => {
      expect(onListChanged).toHaveBeenCalled()
    })
  })

  it('hides input after creating list', async () => {
    const user = userEvent.setup()
    render(<ListsView lists={[]} />)

    await user.click(screen.getByRole('button', { name: /new list/i }))
    const input = screen.getByPlaceholderText(/list name/i)
    await user.type(input, 'Shopping{Enter}')

    await waitFor(() => {
      expect(screen.queryByPlaceholderText(/list name/i)).not.toBeInTheDocument()
    })
  })

  it('does not create list with empty name', async () => {
    const user = userEvent.setup()
    render(<ListsView lists={[]} />)

    await user.click(screen.getByRole('button', { name: /new list/i }))
    const input = screen.getByPlaceholderText(/list name/i)
    await user.type(input, '{Enter}')

    expect(CreateList).not.toHaveBeenCalled()
  })
})

describe('ListsView - Delete List', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows delete button on list card', () => {
    render(<ListsView lists={[createTestList({ name: 'Shopping' })]} />)
    expect(screen.getByTitle('Delete list')).toBeInTheDocument()
  })

  it('shows confirmation dialog when delete list button is clicked', async () => {
    const user = userEvent.setup()
    render(<ListsView lists={[createTestList({ name: 'Shopping' })]} />)

    await user.click(screen.getByTitle('Delete list'))

    expect(screen.getByText('Delete List')).toBeInTheDocument()
    expect(screen.getByText(/are you sure/i)).toBeInTheDocument()
  })

  it('calls DeleteList binding when confirming delete', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    render(<ListsView lists={[createTestList({ id: 42, name: 'Shopping' })]} onListChanged={onListChanged} />)

    await user.click(screen.getByTitle('Delete list'))
    await user.click(screen.getByRole('button', { name: /^delete$/i }))

    await waitFor(() => {
      expect(DeleteList).toHaveBeenCalledWith(42, true)
    })
  })

  it('calls onListChanged after deleting list', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    render(<ListsView lists={[createTestList({ name: 'Shopping' })]} onListChanged={onListChanged} />)

    await user.click(screen.getByTitle('Delete list'))
    await user.click(screen.getByRole('button', { name: /^delete$/i }))

    await waitFor(() => {
      expect(onListChanged).toHaveBeenCalled()
    })
  })

  it('closes dialog on cancel', async () => {
    const user = userEvent.setup()
    render(<ListsView lists={[createTestList({ name: 'Shopping' })]} />)

    await user.click(screen.getByTitle('Delete list'))
    expect(screen.getByText('Delete List')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /cancel/i }))

    expect(screen.queryByText('Delete List')).not.toBeInTheDocument()
  })
})

describe('ListsView - Rename List', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows rename button on list card', () => {
    render(<ListsView lists={[createTestList({ name: 'Shopping' })]} />)
    expect(screen.getByTitle('Rename list')).toBeInTheDocument()
  })

  it('shows input when rename button is clicked', async () => {
    const user = userEvent.setup()
    render(<ListsView lists={[createTestList({ name: 'Shopping' })]} />)

    await user.click(screen.getByTitle('Rename list'))

    expect(screen.getByDisplayValue('Shopping')).toBeInTheDocument()
  })

  it('calls RenameList binding when submitting new name', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    render(<ListsView lists={[createTestList({ id: 42, name: 'Shopping' })]} onListChanged={onListChanged} />)

    await user.click(screen.getByTitle('Rename list'))
    const input = screen.getByDisplayValue('Shopping')
    await user.clear(input)
    await user.type(input, 'Groceries{Enter}')

    await waitFor(() => {
      expect(RenameList).toHaveBeenCalledWith(42, 'Groceries')
    })
  })

  it('calls onListChanged after renaming list', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    render(<ListsView lists={[createTestList({ name: 'Shopping' })]} onListChanged={onListChanged} />)

    await user.click(screen.getByTitle('Rename list'))
    const input = screen.getByDisplayValue('Shopping')
    await user.clear(input)
    await user.type(input, 'Groceries{Enter}')

    await waitFor(() => {
      expect(onListChanged).toHaveBeenCalled()
    })
  })

  it('cancels rename on Escape', async () => {
    const user = userEvent.setup()
    render(<ListsView lists={[createTestList({ name: 'Shopping' })]} />)

    await user.click(screen.getByTitle('Rename list'))
    const input = screen.getByDisplayValue('Shopping')
    await user.type(input, '{Escape}')

    expect(screen.queryByDisplayValue('Shopping')).not.toBeInTheDocument()
    expect(RenameList).not.toHaveBeenCalled()
  })
})

describe('ListsView - Edit List Item', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows edit button on list items', () => {
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ content: 'Buy milk' })]
    })]} />)

    expect(screen.getByTitle('Edit item')).toBeInTheDocument()
  })

  it('shows input when edit button is clicked', async () => {
    const user = userEvent.setup()
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ content: 'Buy milk' })]
    })]} />)

    await user.click(screen.getByTitle('Edit item'))

    expect(screen.getByDisplayValue('Buy milk')).toBeInTheDocument()
  })

  it('calls EditListItem binding when submitting edit', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ id: 42, content: 'Buy milk' })]
    })]} onListChanged={onListChanged} />)

    await user.click(screen.getByTitle('Edit item'))
    const input = screen.getByDisplayValue('Buy milk')
    await user.clear(input)
    await user.type(input, 'Buy cheese{Enter}')

    await waitFor(() => {
      expect(EditListItem).toHaveBeenCalledWith(42, 'Buy cheese')
    })
  })

  it('calls onListChanged after editing item', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ content: 'Buy milk' })]
    })]} onListChanged={onListChanged} />)

    await user.click(screen.getByTitle('Edit item'))
    const input = screen.getByDisplayValue('Buy milk')
    await user.clear(input)
    await user.type(input, 'Buy cheese{Enter}')

    await waitFor(() => {
      expect(onListChanged).toHaveBeenCalled()
    })
  })

  it('cancels edit on Escape', async () => {
    const user = userEvent.setup()
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ content: 'Buy milk' })]
    })]} />)

    await user.click(screen.getByTitle('Edit item'))
    const input = screen.getByDisplayValue('Buy milk')
    await user.type(input, '{Escape}')

    expect(screen.queryByDisplayValue('Buy milk')).not.toBeInTheDocument()
    expect(EditListItem).not.toHaveBeenCalled()
  })
})

describe('ListsView - Cancel List Item', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows cancel button on task items', () => {
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ content: 'Buy milk', type: 'task' })]
    })]} />)

    expect(screen.getByTitle('Cancel item')).toBeInTheDocument()
  })

  it('calls CancelListItem binding when cancel button is clicked', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ id: 42, content: 'Buy milk', type: 'task' })]
    })]} onListChanged={onListChanged} />)

    await user.click(screen.getByTitle('Cancel item'))

    await waitFor(() => {
      expect(CancelListItem).toHaveBeenCalledWith(42)
    })
  })

  it('calls onListChanged after cancelling item', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ content: 'Buy milk', type: 'task' })]
    })]} onListChanged={onListChanged} />)

    await user.click(screen.getByTitle('Cancel item'))

    await waitFor(() => {
      expect(onListChanged).toHaveBeenCalled()
    })
  })

  it('does not show cancel button on done items', () => {
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ content: 'Buy milk', type: 'done', done: true })]
    })]} />)

    expect(screen.queryByTitle('Cancel item')).not.toBeInTheDocument()
  })
})

describe('ListsView - Uncancel List Item', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows uncancel button on cancelled items', () => {
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ content: 'Buy milk', type: 'cancelled' })]
    })]} />)

    expect(screen.getByTitle('Uncancel item')).toBeInTheDocument()
  })

  it('calls UncancelListItem binding when uncancel button is clicked', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ id: 42, content: 'Buy milk', type: 'cancelled' })]
    })]} onListChanged={onListChanged} />)

    await user.click(screen.getByTitle('Uncancel item'))

    await waitFor(() => {
      expect(UncancelListItem).toHaveBeenCalledWith(42)
    })
  })

  it('calls onListChanged after uncancelling item', async () => {
    const user = userEvent.setup()
    const onListChanged = vi.fn()
    render(<ListsView lists={[createTestList({
      name: 'Shopping',
      items: [createTestItem({ content: 'Buy milk', type: 'cancelled' })]
    })]} onListChanged={onListChanged} />)

    await user.click(screen.getByTitle('Uncancel item'))

    await waitFor(() => {
      expect(onListChanged).toHaveBeenCalled()
    })
  })
})

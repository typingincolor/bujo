import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AddEntryBar } from './AddEntryBar'

describe('AddEntryBar', () => {
  it('renders input placeholder', () => {
    render(<AddEntryBar />)
    expect(screen.getByPlaceholderText(/what's on your mind/i)).toBeInTheDocument()
  })

  it('renders type selector buttons', () => {
    render(<AddEntryBar />)
    expect(screen.getByTitle('Task')).toBeInTheDocument()
    expect(screen.getByTitle('Note')).toBeInTheDocument()
    expect(screen.getByTitle('Event')).toBeInTheDocument()
  })

  it('renders question type button', () => {
    render(<AddEntryBar />)
    expect(screen.getByTitle('Question')).toBeInTheDocument()
  })

  it('calls onAdd with task type by default', async () => {
    const onAdd = vi.fn()
    const user = userEvent.setup()

    render(<AddEntryBar onAdd={onAdd} />)

    const input = screen.getByPlaceholderText(/what's on your mind/i)
    await user.type(input, 'New task')
    await user.click(screen.getByRole('button', { name: '' })) // Plus button

    expect(onAdd).toHaveBeenCalledWith('New task', 'task')
  })

  it('calls onAdd with selected type when submitting', async () => {
    const onAdd = vi.fn()
    const user = userEvent.setup()

    render(<AddEntryBar onAdd={onAdd} />)

    await user.click(screen.getByTitle('Note'))

    const input = screen.getByPlaceholderText(/what's on your mind/i)
    await user.type(input, 'New note')
    await user.click(screen.getByRole('button', { name: '' })) // Plus button

    expect(onAdd).toHaveBeenCalledWith('New note', 'note')
  })

  it('calls onAdd with question type when selected', async () => {
    const onAdd = vi.fn()
    const user = userEvent.setup()

    render(<AddEntryBar onAdd={onAdd} />)

    await user.click(screen.getByTitle('Question'))

    const input = screen.getByPlaceholderText(/what's on your mind/i)
    await user.type(input, 'What is TDD?')
    await user.click(screen.getByRole('button', { name: '' })) // Plus button

    expect(onAdd).toHaveBeenCalledWith('What is TDD?', 'question')
  })

  it('clears input after submission', async () => {
    const onAdd = vi.fn()
    const user = userEvent.setup()

    render(<AddEntryBar onAdd={onAdd} />)

    const input = screen.getByPlaceholderText(/what's on your mind/i)
    await user.type(input, 'Test entry')
    await user.click(screen.getByRole('button', { name: '' })) // Plus button

    expect(input).toHaveValue('')
  })

  it('does not call onAdd when input is empty', async () => {
    const onAdd = vi.fn()

    render(<AddEntryBar onAdd={onAdd} />)

    fireEvent.click(screen.getByRole('button', { name: '' })) // Plus button

    expect(onAdd).not.toHaveBeenCalled()
  })

  it('does not call onAdd when input is whitespace only', async () => {
    const onAdd = vi.fn()
    const user = userEvent.setup()

    render(<AddEntryBar onAdd={onAdd} />)

    const input = screen.getByPlaceholderText(/what's on your mind/i)
    await user.type(input, '   ')
    fireEvent.click(screen.getByRole('button', { name: '' })) // Plus button

    expect(onAdd).not.toHaveBeenCalled()
  })

  it('submits on Enter key', async () => {
    const onAdd = vi.fn()
    const user = userEvent.setup()

    render(<AddEntryBar onAdd={onAdd} />)

    const input = screen.getByPlaceholderText(/what's on your mind/i)
    await user.type(input, 'Test entry{enter}')

    expect(onAdd).toHaveBeenCalledWith('Test entry', 'task')
  })
})

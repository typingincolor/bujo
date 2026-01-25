import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { FileUploadButton } from './FileUploadButton'

describe('FileUploadButton', () => {
  it('renders upload button', () => {
    render(<FileUploadButton />)
    expect(screen.getByRole('button', { name: /upload/i })).toBeInTheDocument()
  })

  it('shows file count badge when files selected', () => {
    const files = [
      new File(['content1'], 'file1.txt', { type: 'text/plain' }),
      new File(['content2'], 'file2.txt', { type: 'text/plain' }),
    ]
    render(<FileUploadButton files={files} />)
    expect(screen.getByText('2')).toBeInTheDocument()
  })

  it('opens popover when clicking button', async () => {
    const user = userEvent.setup()
    render(<FileUploadButton />)

    await user.click(screen.getByRole('button', { name: /upload/i }))
    expect(screen.getByText('Upload Files')).toBeInTheDocument()
  })

  it('shows drag-and-drop zone in popover', async () => {
    const user = userEvent.setup()
    render(<FileUploadButton />)

    await user.click(screen.getByRole('button', { name: /upload/i }))
    expect(screen.getByText(/drag and drop/i)).toBeInTheDocument()
  })

  it('accepts files via click', async () => {
    const onFilesChange = vi.fn()
    const user = userEvent.setup()
    render(<FileUploadButton onFilesChange={onFilesChange} />)

    await user.click(screen.getByRole('button', { name: /upload/i }))
    const input = screen.getByLabelText(/choose files/i)

    const mockFile = new File(['content'], 'test.txt', { type: 'text/plain' })
    await user.upload(input, mockFile)
    expect(onFilesChange).toHaveBeenCalledWith([mockFile])
  })

  it('shows file list with sizes', async () => {
    const user = userEvent.setup()
    const files = [
      new File(['content'], 'test.txt', { type: 'text/plain' }),
    ]
    render(<FileUploadButton files={files} />)

    await user.click(screen.getByRole('button', { name: /upload/i }))
    expect(screen.getByText('test.txt')).toBeInTheDocument()
    expect(screen.getByText(/\d+ B/)).toBeInTheDocument()
  })

  it('removes file when clicking remove button', async () => {
    const onFilesChange = vi.fn()
    const user = userEvent.setup()
    const files = [new File([''], 'test.txt', { type: 'text/plain' })]

    render(<FileUploadButton files={files} onFilesChange={onFilesChange} />)

    await user.click(screen.getByRole('button', { name: /upload/i }))
    await user.click(screen.getByLabelText(/remove test.txt/i))

    expect(onFilesChange).toHaveBeenCalledWith([])
  })

  it('shows storage warning message', async () => {
    const user = userEvent.setup()
    render(<FileUploadButton />)

    await user.click(screen.getByRole('button', { name: /upload/i }))
    expect(screen.getByText(/storage not connected/i)).toBeInTheDocument()
  })

  it('highlights drop zone on drag over', async () => {
    const user = userEvent.setup()
    render(<FileUploadButton />)

    await user.click(screen.getByRole('button', { name: /upload/i }))
    const dropZone = screen.getByTestId('drop-zone')

    fireEvent.dragOver(dropZone)
    expect(dropZone).toHaveClass('border-primary', 'bg-primary/10')
  })

  it('does not show badge when no files', () => {
    render(<FileUploadButton />)
    expect(screen.queryByText('0')).not.toBeInTheDocument()
  })
})

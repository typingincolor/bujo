import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { Header } from '../Header'

describe('Header', () => {
  describe('back button', () => {
    it('does not render back button when canGoBack is false', () => {
      render(<Header title="Test" canGoBack={false} onBack={vi.fn()} />)

      expect(screen.queryByRole('button', { name: /back/i })).not.toBeInTheDocument()
    })

    it('renders back button when canGoBack is true', () => {
      render(<Header title="Test" canGoBack={true} onBack={vi.fn()} />)

      expect(screen.getByRole('button', { name: /back/i })).toBeInTheDocument()
    })

    it('calls onBack when back button clicked', async () => {
      const onBack = vi.fn()
      render(<Header title="Test" canGoBack={true} onBack={onBack} />)

      await userEvent.click(screen.getByRole('button', { name: /back/i }))

      expect(onBack).toHaveBeenCalled()
    })
  })
})

import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { KeyboardShortcuts } from './KeyboardShortcuts'

describe('KeyboardShortcuts', () => {
  describe('editable view', () => {
    it('shows save shortcut', () => {
      render(<KeyboardShortcuts view="editable" />)

      expect(screen.getByText(/save/i)).toBeInTheDocument()
    })

    it('shows indent shortcut', () => {
      render(<KeyboardShortcuts view="editable" />)

      expect(screen.getByText(/indent/i)).toBeInTheDocument()
    })

    it('shows outdent shortcut', () => {
      render(<KeyboardShortcuts view="editable" />)

      expect(screen.getByText(/outdent/i)).toBeInTheDocument()
    })

    it('shows import shortcut', () => {
      render(<KeyboardShortcuts view="editable" />)

      expect(screen.getByText(/import/i)).toBeInTheDocument()
    })

    it('shows escape shortcut', () => {
      render(<KeyboardShortcuts view="editable" />)

      expect(screen.getByText(/esc/i)).toBeInTheDocument()
      expect(screen.getByText(/blur/i)).toBeInTheDocument()
    })

    it('shows syntax reference section', () => {
      render(<KeyboardShortcuts view="editable" />)

      expect(screen.getByText(/syntax reference/i)).toBeInTheDocument()
    })

    it('shows entry type symbols in syntax reference', () => {
      render(<KeyboardShortcuts view="editable" />)

      expect(screen.getByText(/\. task/i)).toBeInTheDocument()
      expect(screen.getByText(/- note/i)).toBeInTheDocument()
      expect(screen.getByText(/o event/i)).toBeInTheDocument()
      expect(screen.getByText(/x done/i)).toBeInTheDocument()
    })

    it('shows priority markers in syntax reference', () => {
      render(<KeyboardShortcuts view="editable" />)

      expect(screen.getByText('!!! Highest')).toBeInTheDocument()
      expect(screen.getByText('!! High')).toBeInTheDocument()
      expect(screen.getByText('! Low')).toBeInTheDocument()
    })

    it('shows migration syntax in syntax reference', () => {
      render(<KeyboardShortcuts view="editable" />)

      expect(screen.getByText(/>\[date\]/i)).toBeInTheDocument()
    })
  })
})

import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
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

describe('EntryItem keyboard navigation', () => {
  describe('action bar visibility when selected', () => {
    it('shows action bar when entry is selected (keyboard navigation)', () => {
      render(
        <EntryItem
          entry={createTestEntry()}
          isSelected={true}
          onEdit={() => {}}
          onDelete={() => {}}
        />
      )

      const actionBar = screen.getByTestId('entry-action-bar')
      // When selected via keyboard, action bar should be visible (opacity-100)
      // not relying on group-hover
      expect(actionBar).not.toHaveClass('opacity-0')
    })

    it('hides action bar when entry is not selected and not hovered', () => {
      render(
        <EntryItem
          entry={createTestEntry()}
          isSelected={false}
          onEdit={() => {}}
          onDelete={() => {}}
        />
      )

      const actionBar = screen.getByTestId('entry-action-bar')
      // When not selected and not hovered, action bar should be hidden
      expect(actionBar).toHaveClass('opacity-0')
    })
  })
})

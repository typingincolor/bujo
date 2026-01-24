import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { WeekSummary } from './WeekSummary'
import { DayEntries, Entry } from '@/types/bujo'

describe('WeekSummary - Popover Integration', () => {
  it('clicking meetings opens EntryContextPopover', async () => {
    const mockEvent: Entry = {
      id: 1,
      content: 'Team standup',
      type: 'event',
      priority: 'none',
      parentId: null,
      loggedDate: '2026-01-24T00:00:00Z',
      children: [
        {
          id: 2,
          content: 'Discussed feature X',
          type: 'note',
          priority: 'none',
          parentId: 1,
          loggedDate: '2026-01-24T00:00:00Z',
        },
        {
          id: 3,
          content: 'Follow up on Y',
          type: 'task',
          priority: 'none',
          parentId: 1,
          loggedDate: '2026-01-24T00:00:00Z',
        },
      ],
    }

    const mockDays: DayEntries[] = [
      {
        date: '2026-01-24',
        entries: [mockEvent],
        location: '',
        mood: '',
        weather: '',
      },
    ]

    const mockOnAction = vi.fn()
    const mockOnNavigate = vi.fn()

    render(
      <WeekSummary
        days={mockDays}
        onAction={mockOnAction}
        onNavigate={mockOnNavigate}
      />
    )

    // Find the meeting item
    const meetingButton = screen.getByText('Team standup')
    expect(meetingButton).toBeInTheDocument()

    // Click it
    await userEvent.click(meetingButton)

    // Popover should open
    let popover: HTMLElement
    await waitFor(() => {
      popover = screen.getByTestId('entry-context-popover')
      expect(popover).toBeInTheDocument()
    })

    // Should show meeting content and children in the popover's EntryTree
    const entryTree = screen.getByTestId('entry-tree')
    expect(entryTree).toHaveTextContent('Team standup')
    expect(entryTree).toHaveTextContent('Discussed feature X')
    expect(entryTree).toHaveTextContent('Follow up on Y')
  })
})

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DateNavigator } from './DateNavigator'

describe('DateNavigator', () => {
  const mockOnDateChange = vi.fn()
  const today = new Date('2026-01-25T12:00:00')

  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(today)
    mockOnDateChange.mockClear()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows "Today" when viewing current date', () => {
    render(<DateNavigator date={today} onDateChange={mockOnDateChange} />)
    // The date button should show "Today" text - use exact aria-label match
    const dateButton = screen.getByLabelText('Today')
    expect(dateButton).toHaveTextContent('Today')
  })

  it('shows formatted date when viewing other dates', () => {
    const otherDate = new Date('2026-01-20T12:00:00')
    render(<DateNavigator date={otherDate} onDateChange={mockOnDateChange} />)
    // The aria-label should contain the formatted date (Jan 20, 2026 is a Tuesday)
    const dateButton = screen.getByLabelText('Tue, Jan 20, 2026')
    expect(dateButton).toHaveTextContent(/Tue, Jan 20, 2026/)
  })

  it('navigates to previous day when clicking prev button', async () => {
    vi.useRealTimers() // Use real timers for userEvent
    const user = userEvent.setup()
    render(<DateNavigator date={today} onDateChange={mockOnDateChange} />)

    await user.click(screen.getByLabelText('Previous day'))
    expect(mockOnDateChange).toHaveBeenCalledTimes(1)
    const calledDate = mockOnDateChange.mock.calls[0][0]
    expect(calledDate.getDate()).toBe(24)
    expect(calledDate.getMonth()).toBe(0) // January
    expect(calledDate.getFullYear()).toBe(2026)
  })

  it('navigates to next day when clicking next button', async () => {
    vi.useRealTimers() // Use real timers for userEvent
    const user = userEvent.setup()
    render(<DateNavigator date={today} onDateChange={mockOnDateChange} />)

    await user.click(screen.getByLabelText('Next day'))
    expect(mockOnDateChange).toHaveBeenCalledTimes(1)
    const calledDate = mockOnDateChange.mock.calls[0][0]
    expect(calledDate.getDate()).toBe(26)
    expect(calledDate.getMonth()).toBe(0) // January
    expect(calledDate.getFullYear()).toBe(2026)
  })

  it('shows Today button when not viewing today', () => {
    const otherDate = new Date('2026-01-20T12:00:00')
    render(<DateNavigator date={otherDate} onDateChange={mockOnDateChange} />)
    const todayButton = screen.getByRole('button', { name: /jump to today/i })
    expect(todayButton).toBeVisible()
    expect(todayButton).not.toHaveClass('invisible')
  })

  it('hides Today button when viewing today (using invisible)', () => {
    render(<DateNavigator date={today} onDateChange={mockOnDateChange} />)
    const todayButton = screen.getByRole('button', { name: /jump to today/i })
    expect(todayButton).toHaveClass('invisible')
  })

  it('navigates to today when clicking Today button', async () => {
    vi.useRealTimers() // Use real timers for userEvent
    const user = userEvent.setup()
    const otherDate = new Date('2026-01-20T12:00:00')
    render(<DateNavigator date={otherDate} onDateChange={mockOnDateChange} />)

    await user.click(screen.getByRole('button', { name: /jump to today/i }))
    expect(mockOnDateChange).toHaveBeenCalledTimes(1)
    const calledDate = mockOnDateChange.mock.calls[0][0]
    expect(calledDate.getDate()).toBe(25)
    expect(calledDate.getMonth()).toBe(0) // January
    expect(calledDate.getFullYear()).toBe(2026)
  })

  it('opens calendar popover when clicking date button', async () => {
    vi.useRealTimers() // Use real timers for userEvent
    const user = userEvent.setup()
    render(<DateNavigator date={today} onDateChange={mockOnDateChange} />)

    // Click the date button (has Calendar icon and "Today" text)
    const dateButton = screen.getByLabelText('Today')
    await user.click(dateButton)
    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })

  it('calls onDateChange when selecting date from calendar', async () => {
    vi.useRealTimers() // Use real timers for userEvent
    const user = userEvent.setup()
    render(<DateNavigator date={today} onDateChange={mockOnDateChange} />)

    // Click the date button to open calendar
    const dateButton = screen.getByLabelText('Today')
    await user.click(dateButton)
    await user.click(screen.getByRole('gridcell', { name: '15' }))

    expect(mockOnDateChange).toHaveBeenCalled()
  })
})

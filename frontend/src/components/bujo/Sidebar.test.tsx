import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { Sidebar } from './Sidebar';

describe('Sidebar', () => {
  const mockOnViewChange = vi.fn();

  it('renders improved navigation labels', () => {
    render(<Sidebar currentView="today" onViewChange={mockOnViewChange} />);

    // Improved labels (Option B)
    expect(screen.getByText('Journal')).toBeInTheDocument();
    expect(screen.getByText('Weekly Review')).toBeInTheDocument();
    expect(screen.getByText('Open Questions')).toBeInTheDocument();
    expect(screen.getByText('Habit Tracker')).toBeInTheDocument();
    expect(screen.getByText('Lists')).toBeInTheDocument();
    expect(screen.getByText('Monthly Goals')).toBeInTheDocument();
    expect(screen.getByText('Search')).toBeInTheDocument();
    expect(screen.getByText('Insights')).toBeInTheDocument();
    expect(screen.getByText('Settings')).toBeInTheDocument();
  });

  it('renders improved tagline', () => {
    render(<Sidebar currentView="today" onViewChange={mockOnViewChange} />);

    // Improved tagline (Option C)
    expect(screen.getByText('Capture. Track. Reflect.')).toBeInTheDocument();
  });

  it('settings button container has no border above it', () => {
    render(<Sidebar currentView="today" onViewChange={mockOnViewChange} />);

    const settingsButton = screen.getByRole('button', { name: /settings/i });
    const footer = settingsButton.parentElement;

    // Footer should NOT have border-t class (no line above settings)
    expect(footer).not.toHaveClass('border-t');
  });
});

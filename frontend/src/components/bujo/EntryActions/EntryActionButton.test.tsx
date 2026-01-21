import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { EntryActionButton } from './EntryActionButton';
import { ACTION_REGISTRY } from './types';

describe('EntryActionButton', () => {
  it('renders button with correct title', () => {
    const onClick = vi.fn();
    render(
      <EntryActionButton
        config={ACTION_REGISTRY.edit}
        onClick={onClick}
      />
    );

    const button = screen.getByTitle('Edit entry');
    expect(button).toBeInTheDocument();
  });

  it('calls onClick when clicked', () => {
    const onClick = vi.fn();
    render(
      <EntryActionButton
        config={ACTION_REGISTRY.delete}
        onClick={onClick}
      />
    );

    const button = screen.getByTitle('Delete entry');
    fireEvent.click(button);

    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('stops event propagation on click', () => {
    const onClick = vi.fn();
    const parentClick = vi.fn();

    render(
      <div onClick={parentClick}>
        <EntryActionButton
          config={ACTION_REGISTRY.cancel}
          onClick={onClick}
        />
      </div>
    );

    const button = screen.getByTitle('Cancel entry');
    fireEvent.click(button);

    expect(onClick).toHaveBeenCalledTimes(1);
    expect(parentClick).not.toHaveBeenCalled();
  });

  it('renders with data-action-slot attribute', () => {
    const onClick = vi.fn();
    render(
      <EntryActionButton
        config={ACTION_REGISTRY.migrate}
        onClick={onClick}
      />
    );

    const button = screen.getByTitle('Migrate entry');
    expect(button).toHaveAttribute('data-action-slot');
  });

  it('renders icon with correct size for sm variant', () => {
    const onClick = vi.fn();
    render(
      <EntryActionButton
        config={ACTION_REGISTRY.edit}
        onClick={onClick}
        size="sm"
      />
    );

    const button = screen.getByTitle('Edit entry');
    const svg = button.querySelector('svg');
    expect(svg).toHaveClass('w-3.5', 'h-3.5');
  });

  it('renders icon with correct size for md variant', () => {
    const onClick = vi.fn();
    render(
      <EntryActionButton
        config={ACTION_REGISTRY.edit}
        onClick={onClick}
        size="md"
      />
    );

    const button = screen.getByTitle('Edit entry');
    const svg = button.querySelector('svg');
    expect(svg).toHaveClass('w-4', 'h-4');
  });

  it('applies hover classes from config', () => {
    const onClick = vi.fn();
    render(
      <EntryActionButton
        config={ACTION_REGISTRY.delete}
        onClick={onClick}
      />
    );

    const button = screen.getByTitle('Delete entry');
    expect(button).toHaveClass('hover:bg-destructive/20');
    expect(button).toHaveClass('hover:text-destructive');
  });
});

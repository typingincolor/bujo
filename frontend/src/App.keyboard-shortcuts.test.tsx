import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { userEvent } from '@testing-library/user-event';
import { SettingsProvider } from './contexts/SettingsContext';
import App from './App';
import { createMockEntry, createMockDayEntries, createMockDays, createMockOverdue } from './test/mocks';

const mockDays = createMockDays([createMockDayEntries({
  Entries: [
    createMockEntry({ ID: 1, EntityID: 'e1', Type: 'Task', Content: 'Test task', CreatedAt: '2026-01-17T10:00:00Z' }),
  ],
})]);
const mockOverdue = createMockOverdue([]);

vi.mock('./wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn().mockReturnValue(() => {}),
  OnFileDrop: vi.fn(),
  OnFileDropOff: vi.fn(),
}));

vi.mock('./wailsjs/go/wails/App', () => ({
  GetDayEntries: vi.fn().mockResolvedValue([{ Date: '2026-01-17T00:00:00Z', Entries: [], Location: '', Mood: '', Weather: '' }]),
  GetOverdue: vi.fn().mockResolvedValue([]),
  GetHabits: vi.fn().mockResolvedValue({ Habits: [] }),
  GetLists: vi.fn().mockResolvedValue([]),
  GetGoals: vi.fn().mockResolvedValue([]),
  GetOutstandingQuestions: vi.fn().mockResolvedValue([]),
  GetWeekSummary: vi.fn().mockResolvedValue({ Days: [] }),
  AddEntry: vi.fn().mockResolvedValue([1]),
  MarkEntryDone: vi.fn().mockResolvedValue(undefined),
  MarkEntryUndone: vi.fn().mockResolvedValue(undefined),
  EditEntry: vi.fn().mockResolvedValue(undefined),
  DeleteEntry: vi.fn().mockResolvedValue(undefined),
  HasChildren: vi.fn().mockResolvedValue(false),
  CancelEntry: vi.fn().mockResolvedValue(undefined),
  UncancelEntry: vi.fn().mockResolvedValue(undefined),
  CyclePriority: vi.fn().mockResolvedValue(undefined),
  MigrateEntry: vi.fn().mockResolvedValue(100),
  CreateHabit: vi.fn().mockResolvedValue(1),
  SetMood: vi.fn().mockResolvedValue(undefined),
  SetWeather: vi.fn().mockResolvedValue(undefined),
  SetLocation: vi.fn().mockResolvedValue(undefined),
  GetLocationHistory: vi.fn().mockResolvedValue(['Home', 'Office']),
  OpenFileDialog: vi.fn().mockResolvedValue(''),
  ReadFile: vi.fn().mockResolvedValue(''),
  GetEditableDocumentWithEntries: vi.fn().mockResolvedValue({ document: '', entries: [] }),
  ValidateEditableDocument: vi.fn().mockResolvedValue({ isValid: true, errors: [] }),
  ApplyEditableDocument: vi.fn().mockResolvedValue({ inserted: 0, updated: 0, deleted: 0, migrated: 0 }),
  SearchEntries: vi.fn().mockResolvedValue([]),
  GetStats: vi.fn().mockResolvedValue({
    TotalEntries: 0,
    TasksCompleted: 0,
    ActiveHabits: 0,
    CurrentStreak: 0,
  }),
  GetVersion: vi.fn().mockResolvedValue('1.0.0'),
}));

import { GetDayEntries, GetOverdue } from './wailsjs/go/wails/App';

describe('App keyboard shortcuts', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(GetDayEntries).mockResolvedValue(mockDays);
    vi.mocked(GetOverdue).mockResolvedValue(mockOverdue);
  });

  describe('View navigation shortcuts', () => {
    it('CMD+1 switches to Journal view', async () => {
      const user = userEvent.setup();
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      );

      // Wait for initial load
      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument();
      });

      // Switch away first
      await user.keyboard('{Meta>}2{/Meta}');

      // Then switch back with CMD+1
      await user.keyboard('{Meta>}1{/Meta}');

      // Check sidebar button is active
      const journalButton = screen.getByRole('button', { name: /Journal/i });
      expect(journalButton).toHaveAttribute('aria-pressed', 'true');
    });

    it('CMD+2 switches to Weekly Review view', async () => {
      const user = userEvent.setup();
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      );

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument();
      });

      await user.keyboard('{Meta>}2{/Meta}');

      const weekButton = screen.getByRole('button', { name: /Weekly Review/i });
      expect(weekButton).toHaveAttribute('aria-pressed', 'true');
    });

    it('CMD+3 switches to Open Questions view', async () => {
      const user = userEvent.setup();
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      );

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument();
      });

      await user.keyboard('{Meta>}3{/Meta}');

      const questionsButton = screen.getByRole('button', { name: /Open Questions/i });
      expect(questionsButton).toHaveAttribute('aria-pressed', 'true');
    });

    it('CMD+4 switches to Habit Tracker view', async () => {
      const user = userEvent.setup();
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      );

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument();
      });

      await user.keyboard('{Meta>}4{/Meta}');

      const habitsButton = screen.getByRole('button', { name: /Habit Tracker/i });
      expect(habitsButton).toHaveAttribute('aria-pressed', 'true');
    });

    it('CMD+5 switches to Lists view', async () => {
      const user = userEvent.setup();
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      );

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument();
      });

      await user.keyboard('{Meta>}5{/Meta}');

      const listsButton = screen.getByRole('button', { name: /Lists/i });
      expect(listsButton).toHaveAttribute('aria-pressed', 'true');
    });

    it('CMD+6 switches to Monthly Goals view', async () => {
      const user = userEvent.setup();
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      );

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument();
      });

      await user.keyboard('{Meta>}6{/Meta}');

      const goalsButton = screen.getByRole('button', { name: /Monthly Goals/i });
      expect(goalsButton).toHaveAttribute('aria-pressed', 'true');
    });

    it('CMD+7 switches to Search view', async () => {
      const user = userEvent.setup();
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      );

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument();
      });

      await user.keyboard('{Meta>}7{/Meta}');

      const searchButton = screen.getByRole('button', { name: /Search/i });
      expect(searchButton).toHaveAttribute('aria-pressed', 'true');
    });

    it('CMD+8 switches to Insights view', async () => {
      const user = userEvent.setup();
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      );

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument();
      });

      await user.keyboard('{Meta>}8{/Meta}');

      const insightsButton = screen.getByRole('button', { name: /Insights/i });
      expect(insightsButton).toHaveAttribute('aria-pressed', 'true');
    });

    it('CMD+9 switches to Settings view', async () => {
      const user = userEvent.setup();
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      );

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument();
      });

      await user.keyboard('{Meta>}9{/Meta}');

      const settingsButton = screen.getByRole('button', { name: /Settings/i });
      expect(settingsButton).toHaveAttribute('aria-pressed', 'true');
    });

    it('Ctrl+1 switches to Journal view on Windows/Linux', async () => {
      const user = userEvent.setup();
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      );

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument();
      });

      // Switch away first
      await user.keyboard('{Control>}2{/Control}');

      // Then switch back with Ctrl+1
      await user.keyboard('{Control>}1{/Control}');

      const journalButton = screen.getByRole('button', { name: /Journal/i });
      expect(journalButton).toHaveAttribute('aria-pressed', 'true');
    });

    it('CMD+0 does nothing', async () => {
      const user = userEvent.setup();
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      );

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument();
      });

      // Should start in Journal view
      const journalButton = screen.getByRole('button', { name: /Journal/i });
      expect(journalButton).toHaveAttribute('aria-pressed', 'true');

      // Try CMD+0
      await user.keyboard('{Meta>}0{/Meta}');

      // Should still be in Journal view
      expect(journalButton).toHaveAttribute('aria-pressed', 'true');
    });

    it('switches between multiple views using keyboard shortcuts', async () => {
      const user = userEvent.setup();
      render(
        <SettingsProvider>
          <App />
        </SettingsProvider>
      );

      await waitFor(() => {
        expect(screen.queryByText('Loading your journal...')).not.toBeInTheDocument();
      });

      // Start in Journal
      const journalButton = screen.getByRole('button', { name: /Journal/i });
      expect(journalButton).toHaveAttribute('aria-pressed', 'true');

      // Switch to Weekly Review
      await user.keyboard('{Meta>}2{/Meta}');
      const weekButton = screen.getByRole('button', { name: /Weekly Review/i });
      expect(weekButton).toHaveAttribute('aria-pressed', 'true');

      // Switch to Habits
      await user.keyboard('{Meta>}4{/Meta}');
      const habitsButton = screen.getByRole('button', { name: /Habit Tracker/i });
      expect(habitsButton).toHaveAttribute('aria-pressed', 'true');

      // Switch back to Journal
      await user.keyboard('{Meta>}1{/Meta}');
      expect(journalButton).toHaveAttribute('aria-pressed', 'true');
    });
  });
});

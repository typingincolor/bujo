package sqlite

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func setupInsightsTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	schema := `
		CREATE TABLE summaries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			week_start TEXT NOT NULL,
			week_end TEXT NOT NULL,
			summary_text TEXT NOT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE topics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			summary_id INTEGER NOT NULL,
			topic TEXT NOT NULL,
			content TEXT,
			importance TEXT CHECK(importance IN ('high', 'medium', 'low')),
			FOREIGN KEY (summary_id) REFERENCES summaries(id)
		);

		CREATE TABLE initiatives (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			status TEXT CHECK(status IN ('active', 'planning', 'blocked', 'completed', 'on-hold')),
			description TEXT,
			last_updated TEXT DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE initiative_mentions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			summary_id INTEGER NOT NULL,
			initiative_id INTEGER NOT NULL,
			update_text TEXT,
			FOREIGN KEY (summary_id) REFERENCES summaries(id),
			FOREIGN KEY (initiative_id) REFERENCES initiatives(id)
		);

		CREATE TABLE actions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			summary_id INTEGER NOT NULL,
			action_text TEXT NOT NULL,
			priority TEXT CHECK(priority IN ('high', 'medium', 'low')),
			status TEXT CHECK(status IN ('pending', 'completed', 'blocked', 'cancelled')),
			due_date TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (summary_id) REFERENCES summaries(id)
		);

		CREATE TABLE decisions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			decision_text TEXT NOT NULL,
			rationale TEXT,
			participants TEXT,
			expected_outcomes TEXT,
			decision_date TEXT NOT NULL,
			summary_id INTEGER,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (summary_id) REFERENCES summaries(id)
		);

		CREATE TABLE decision_initiatives (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			decision_id INTEGER NOT NULL,
			initiative_id INTEGER NOT NULL,
			FOREIGN KEY (decision_id) REFERENCES decisions(id),
			FOREIGN KEY (initiative_id) REFERENCES initiatives(id)
		);

		CREATE TABLE metadata (
			key TEXT PRIMARY KEY,
			value TEXT
		);

		INSERT INTO metadata (key, value) VALUES ('version', '1.1');
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	seedInsightsData(t, db)
	return db
}

func setupEmptyInsightsTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	schema := `
		CREATE TABLE summaries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			week_start TEXT NOT NULL,
			week_end TEXT NOT NULL,
			summary_text TEXT NOT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)
	return db
}

func seedInsightsData(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO summaries (id, week_start, week_end, summary_text, created_at) VALUES
		(1, '2026-01-13', '2026-01-19', 'Week of Jan 13: Focused on GenAI integration and team planning.', '2026-01-20 09:00:00'),
		(2, '2026-01-20', '2026-01-26', 'Week of Jan 20: Major progress on tech scorecard. Team retrospective.', '2026-01-27 09:00:00'),
		(3, '2026-01-27', '2026-02-02', 'Week of Jan 27: Sprint completion and quarterly planning kickoff.', '2026-02-03 09:00:00')
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO topics (summary_id, topic, content, importance) VALUES
		(1, 'GenAI', 'Integration planning for AI features', 'high'),
		(1, 'Team Planning', 'Q1 roadmap discussion', 'medium'),
		(2, 'Tech Scorecard', 'Completed initial assessment', 'high'),
		(3, 'Quarterly Planning', 'Q2 goals defined', 'high'),
		(3, 'Sprint Review', 'All stories completed', 'medium')
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO initiatives (id, name, status, description, last_updated) VALUES
		(1, 'GenAI Integration', 'active', 'Integrate AI capabilities into core platform', '2026-01-27'),
		(2, 'Tech Scorecard', 'active', 'Technology assessment framework', '2026-01-26'),
		(3, 'Q1 OKRs', 'completed', 'First quarter objectives', '2026-01-20')
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO initiative_mentions (summary_id, initiative_id, update_text) VALUES
		(1, 1, 'Started planning AI integration approach'),
		(2, 2, 'Completed tech scorecard assessment'),
		(3, 1, 'AI integration sprint completed')
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO actions (summary_id, action_text, priority, status, due_date, created_at) VALUES
		(1, 'Review AI vendor proposals', 'high', 'completed', '2026-01-20', '2026-01-20 09:00:00'),
		(2, 'Schedule tech scorecard review', 'medium', 'pending', '2026-02-10', '2026-01-27 09:00:00'),
		(3, 'Prepare Q2 planning materials', 'high', 'pending', '2026-02-05', '2026-02-03 09:00:00'),
		(3, 'Update team onboarding docs', 'low', 'pending', NULL, '2026-02-03 09:00:00')
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO decisions (decision_text, rationale, participants, expected_outcomes, decision_date, summary_id, created_at) VALUES
		('Adopt Claude as primary AI provider', 'Best performance on code tasks', 'Engineering team', 'Improved developer productivity', '2026-01-15', 1, '2026-01-20 09:00:00'),
		('Move to biweekly sprints', 'Better planning cadence', 'Team leads', 'More predictable delivery', '2026-01-28', 3, '2026-02-03 09:00:00')
	`)
	require.NoError(t, err)
}

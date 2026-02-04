package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/typingincolor/bujo/internal/domain"
)

type InsightsRepository struct {
	db *sql.DB
}

func NewInsightsRepository(db *sql.DB) *InsightsRepository {
	return &InsightsRepository{db: db}
}

func (r *InsightsRepository) IsAvailable() bool {
	return r.db != nil
}

func (r *InsightsRepository) GetLatestSummary(ctx context.Context) (*domain.InsightsSummary, error) {
	if r.db == nil {
		return nil, nil
	}

	row := r.db.QueryRowContext(ctx,
		`SELECT id, week_start, week_end, summary_text, created_at
		 FROM summaries ORDER BY week_start DESC LIMIT 1`)

	var s domain.InsightsSummary
	err := row.Scan(&s.ID, &s.WeekStart, &s.WeekEnd, &s.SummaryText, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *InsightsRepository) GetSummaries(ctx context.Context, limit int) ([]domain.InsightsSummary, error) {
	if r.db == nil {
		return []domain.InsightsSummary{}, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, week_start, week_end, summary_text, created_at
		 FROM summaries ORDER BY week_start DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []domain.InsightsSummary
	for rows.Next() {
		var s domain.InsightsSummary
		if err := rows.Scan(&s.ID, &s.WeekStart, &s.WeekEnd, &s.SummaryText, &s.CreatedAt); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	if summaries == nil {
		summaries = []domain.InsightsSummary{}
	}
	return summaries, rows.Err()
}

func (r *InsightsRepository) GetTopicsForSummary(ctx context.Context, summaryID int64) ([]domain.InsightsTopic, error) {
	if r.db == nil {
		return []domain.InsightsTopic{}, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, summary_id, topic, content, importance
		 FROM topics WHERE summary_id = ?
		 ORDER BY CASE importance
			WHEN 'high' THEN 1
			WHEN 'medium' THEN 2
			WHEN 'low' THEN 3
		 END`, summaryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []domain.InsightsTopic
	for rows.Next() {
		var t domain.InsightsTopic
		var content sql.NullString
		if err := rows.Scan(&t.ID, &t.SummaryID, &t.Topic, &content, &t.Importance); err != nil {
			return nil, err
		}
		t.Content = content.String
		topics = append(topics, t)
	}
	if topics == nil {
		topics = []domain.InsightsTopic{}
	}
	return topics, rows.Err()
}

func (r *InsightsRepository) GetActiveInitiatives(ctx context.Context, limit int) ([]domain.InsightsInitiative, error) {
	if r.db == nil {
		return []domain.InsightsInitiative{}, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, status, description, last_updated
		 FROM initiatives WHERE status = 'active'
		 ORDER BY last_updated DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var initiatives []domain.InsightsInitiative
	for rows.Next() {
		var i domain.InsightsInitiative
		var desc sql.NullString
		if err := rows.Scan(&i.ID, &i.Name, &i.Status, &desc, &i.LastUpdated); err != nil {
			return nil, err
		}
		i.Description = desc.String
		initiatives = append(initiatives, i)
	}
	if initiatives == nil {
		initiatives = []domain.InsightsInitiative{}
	}
	return initiatives, rows.Err()
}

func (r *InsightsRepository) GetPendingActions(ctx context.Context) ([]domain.InsightsAction, error) {
	if r.db == nil {
		return []domain.InsightsAction{}, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT a.id, a.summary_id, a.action_text, a.priority, a.status,
		        COALESCE(a.due_date, ''), a.created_at, s.week_start
		 FROM actions a
		 JOIN summaries s ON a.summary_id = s.id
		 WHERE a.status = 'pending'
		 ORDER BY
			CASE a.priority
				WHEN 'high' THEN 1
				WHEN 'medium' THEN 2
				WHEN 'low' THEN 3
			END,
			a.due_date ASC NULLS LAST,
			s.week_start DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []domain.InsightsAction
	for rows.Next() {
		var a domain.InsightsAction
		if err := rows.Scan(&a.ID, &a.SummaryID, &a.ActionText, &a.Priority,
			&a.Status, &a.DueDate, &a.CreatedAt, &a.WeekStart); err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	if actions == nil {
		actions = []domain.InsightsAction{}
	}
	return actions, rows.Err()
}

func (r *InsightsRepository) GetRecentDecisions(ctx context.Context, limit int) ([]domain.InsightsDecision, error) {
	if r.db == nil {
		return []domain.InsightsDecision{}, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, decision_text, rationale, participants,
		        expected_outcomes, decision_date, summary_id, created_at
		 FROM decisions
		 ORDER BY decision_date DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []domain.InsightsDecision
	for rows.Next() {
		var d domain.InsightsDecision
		var rationale, participants, outcomes sql.NullString
		if err := rows.Scan(&d.ID, &d.DecisionText, &rationale, &participants,
			&outcomes, &d.DecisionDate, &d.SummaryID, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.Rationale = rationale.String
		d.Participants = participants.String
		d.ExpectedOutcomes = outcomes.String
		decisions = append(decisions, d)
	}
	if decisions == nil {
		decisions = []domain.InsightsDecision{}
	}
	return decisions, rows.Err()
}

func (r *InsightsRepository) GetSummaryForWeek(ctx context.Context, weekStart, nextWeekStart string) (*domain.InsightsSummary, error) {
	if r.db == nil {
		return nil, nil
	}

	row := r.db.QueryRowContext(ctx,
		`SELECT id, week_start, week_end, summary_text, created_at
		 FROM summaries WHERE week_start >= ? AND week_start < ?
		 ORDER BY week_start DESC LIMIT 1`, weekStart, nextWeekStart)

	var s domain.InsightsSummary
	err := row.Scan(&s.ID, &s.WeekStart, &s.WeekEnd, &s.SummaryText, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *InsightsRepository) GetActionsForWeek(ctx context.Context, weekStart, nextWeekStart string) ([]domain.InsightsAction, error) {
	if r.db == nil {
		return []domain.InsightsAction{}, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT a.id, a.summary_id, a.action_text, a.priority, a.status,
		        COALESCE(a.due_date, ''), a.created_at, s.week_start
		 FROM actions a
		 JOIN summaries s ON a.summary_id = s.id
		 WHERE s.week_start >= ? AND s.week_start < ?
		 ORDER BY
			CASE a.priority
				WHEN 'high' THEN 1
				WHEN 'medium' THEN 2
				WHEN 'low' THEN 3
			END,
			a.due_date ASC NULLS LAST`, weekStart, nextWeekStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []domain.InsightsAction
	for rows.Next() {
		var a domain.InsightsAction
		if err := rows.Scan(&a.ID, &a.SummaryID, &a.ActionText, &a.Priority,
			&a.Status, &a.DueDate, &a.CreatedAt, &a.WeekStart); err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	if actions == nil {
		actions = []domain.InsightsAction{}
	}
	return actions, rows.Err()
}

func (r *InsightsRepository) GetDaysSinceLastSummary(ctx context.Context) (int, error) {
	if r.db == nil {
		return -1, nil
	}

	row := r.db.QueryRowContext(ctx,
		`SELECT CAST(julianday('now') - julianday(MAX(week_start)) AS INTEGER)
		 FROM summaries`)

	var days sql.NullInt64
	if err := row.Scan(&days); err != nil {
		return 0, fmt.Errorf("failed to query days since last summary: %w", err)
	}
	if !days.Valid {
		return 0, nil
	}
	return int(days.Int64), nil
}

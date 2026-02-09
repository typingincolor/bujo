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
	defer func() { _ = rows.Close() }()

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
	defer func() { _ = rows.Close() }()

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
	defer func() { _ = rows.Close() }()

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
	defer func() { _ = rows.Close() }()

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
	defer func() { _ = rows.Close() }()

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
	defer func() { _ = rows.Close() }()

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

func (r *InsightsRepository) GetInitiativePortfolio(ctx context.Context) ([]domain.InsightsInitiativePortfolio, error) {
	if r.db == nil {
		return []domain.InsightsInitiativePortfolio{}, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT
			i.id, i.name, i.status, COALESCE(i.description, ''), i.last_updated,
			COUNT(im.id) as mention_count,
			COALESCE(MAX(s.week_start), '') as last_mentioned_week,
			COALESCE(GROUP_CONCAT(DISTINCT s.week_start), '') as activity_weeks
		 FROM initiatives i
		 LEFT JOIN initiative_mentions im ON i.id = im.initiative_id
		 LEFT JOIN summaries s ON im.summary_id = s.id
		 GROUP BY i.id
		 ORDER BY
			CASE i.status
				WHEN 'active' THEN 1
				WHEN 'planning' THEN 2
				WHEN 'blocked' THEN 3
				WHEN 'on-hold' THEN 4
				WHEN 'completed' THEN 5
			END,
			mention_count DESC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var portfolio []domain.InsightsInitiativePortfolio
	for rows.Next() {
		var p domain.InsightsInitiativePortfolio
		if err := rows.Scan(&p.ID, &p.Name, &p.Status, &p.Description, &p.LastUpdated,
			&p.MentionCount, &p.LastMentionWeek, &p.ActivityWeeks); err != nil {
			return nil, err
		}
		portfolio = append(portfolio, p)
	}
	if portfolio == nil {
		portfolio = []domain.InsightsInitiativePortfolio{}
	}
	return portfolio, rows.Err()
}

func (r *InsightsRepository) GetInitiativeDetail(ctx context.Context, initiativeID int64) (*domain.InsightsInitiativeDetail, error) {
	if r.db == nil {
		return nil, nil
	}

	var init domain.InsightsInitiative
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, status, COALESCE(description, ''), last_updated
		 FROM initiatives WHERE id = ?`, initiativeID).
		Scan(&init.ID, &init.Name, &init.Status, &desc, &init.LastUpdated)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	init.Description = desc.String

	updates, err := r.getInitiativeUpdates(ctx, initiativeID)
	if err != nil {
		return nil, err
	}

	pendingActions, err := r.getInitiativePendingActions(ctx, initiativeID)
	if err != nil {
		return nil, err
	}

	decisions, err := r.getInitiativeDecisions(ctx, initiativeID)
	if err != nil {
		return nil, err
	}

	return &domain.InsightsInitiativeDetail{
		Initiative:     init,
		Updates:        updates,
		PendingActions: pendingActions,
		Decisions:      decisions,
	}, nil
}

func (r *InsightsRepository) getInitiativeUpdates(ctx context.Context, initiativeID int64) ([]domain.InsightsInitiativeUpdate, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT s.week_start, s.week_end, im.update_text
		 FROM initiative_mentions im
		 JOIN summaries s ON im.summary_id = s.id
		 WHERE im.initiative_id = ?
		 ORDER BY s.week_start ASC`, initiativeID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var updates []domain.InsightsInitiativeUpdate
	for rows.Next() {
		var u domain.InsightsInitiativeUpdate
		var text sql.NullString
		if err := rows.Scan(&u.WeekStart, &u.WeekEnd, &text); err != nil {
			return nil, err
		}
		u.UpdateText = text.String
		updates = append(updates, u)
	}
	if updates == nil {
		updates = []domain.InsightsInitiativeUpdate{}
	}
	return updates, rows.Err()
}

func (r *InsightsRepository) getInitiativePendingActions(ctx context.Context, initiativeID int64) ([]domain.InsightsAction, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT a.id, a.summary_id, a.action_text, a.priority, a.status,
		        COALESCE(a.due_date, ''), a.created_at, s.week_start
		 FROM actions a
		 JOIN summaries s ON a.summary_id = s.id
		 JOIN initiative_mentions im ON im.summary_id = s.id AND im.initiative_id = ?
		 WHERE a.status = 'pending'
		 ORDER BY
			CASE a.priority
				WHEN 'high' THEN 1
				WHEN 'medium' THEN 2
				WHEN 'low' THEN 3
			END,
			a.due_date ASC NULLS LAST`, initiativeID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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

func (r *InsightsRepository) getInitiativeDecisions(ctx context.Context, initiativeID int64) ([]domain.InsightsDecision, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT d.id, d.decision_text, COALESCE(d.rationale, ''),
		        COALESCE(d.participants, ''), COALESCE(d.expected_outcomes, ''),
		        d.decision_date, d.summary_id, d.created_at
		 FROM decisions d
		 JOIN decision_initiatives di ON d.id = di.decision_id
		 WHERE di.initiative_id = ?
		 ORDER BY d.decision_date DESC`, initiativeID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var decisions []domain.InsightsDecision
	for rows.Next() {
		var d domain.InsightsDecision
		if err := rows.Scan(&d.ID, &d.DecisionText, &d.Rationale, &d.Participants,
			&d.ExpectedOutcomes, &d.DecisionDate, &d.SummaryID, &d.CreatedAt); err != nil {
			return nil, err
		}
		decisions = append(decisions, d)
	}
	if decisions == nil {
		decisions = []domain.InsightsDecision{}
	}
	return decisions, rows.Err()
}

func (r *InsightsRepository) GetDistinctTopics(ctx context.Context) ([]string, error) {
	if r.db == nil {
		return []string{}, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT topic FROM topics ORDER BY topic ASC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var topics []string
	for rows.Next() {
		var topic string
		if err := rows.Scan(&topic); err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}
	if topics == nil {
		topics = []string{}
	}
	return topics, rows.Err()
}

func (r *InsightsRepository) GetTopicTimeline(ctx context.Context, topic string) ([]domain.InsightsTopicTimeline, error) {
	if r.db == nil {
		return []domain.InsightsTopicTimeline{}, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT t.topic, COALESCE(t.content, ''), t.importance, s.week_start, s.week_end
		 FROM topics t
		 JOIN summaries s ON t.summary_id = s.id
		 WHERE t.topic = ?
		 ORDER BY s.week_start ASC`, topic)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var timeline []domain.InsightsTopicTimeline
	for rows.Next() {
		var tl domain.InsightsTopicTimeline
		if err := rows.Scan(&tl.Topic, &tl.Content, &tl.Importance, &tl.WeekStart, &tl.WeekEnd); err != nil {
			return nil, err
		}
		timeline = append(timeline, tl)
	}
	if timeline == nil {
		timeline = []domain.InsightsTopicTimeline{}
	}
	return timeline, rows.Err()
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

func (r *InsightsRepository) GetWeeklyReport(ctx context.Context, weekStart, nextWeekStart string) (*domain.InsightsWeeklyReport, error) {
	if r.db == nil {
		return nil, nil
	}

	summary, err := r.GetSummaryForWeek(ctx, weekStart, nextWeekStart)
	if err != nil {
		return nil, err
	}
	if summary == nil {
		return nil, nil
	}

	topics, err := r.GetTopicsForSummary(ctx, summary.ID)
	if err != nil {
		return nil, err
	}

	actions, err := r.GetActionsForWeek(ctx, weekStart, nextWeekStart)
	if err != nil {
		return nil, err
	}

	initiativeUpdates, err := r.getInitiativeUpdatesForSummary(ctx, summary.ID)
	if err != nil {
		return nil, err
	}

	return &domain.InsightsWeeklyReport{
		Summary:           summary,
		Topics:            topics,
		InitiativeUpdates: initiativeUpdates,
		Actions:           actions,
	}, nil
}

func (r *InsightsRepository) getInitiativeUpdatesForSummary(ctx context.Context, summaryID int64) ([]domain.InsightsInitiativeWeekUpdate, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT i.name, COALESCE(im.update_text, '')
		 FROM initiative_mentions im
		 JOIN initiatives i ON im.initiative_id = i.id
		 WHERE im.summary_id = ?
		 ORDER BY i.name ASC`, summaryID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var updates []domain.InsightsInitiativeWeekUpdate
	for rows.Next() {
		var u domain.InsightsInitiativeWeekUpdate
		if err := rows.Scan(&u.InitiativeName, &u.UpdateText); err != nil {
			return nil, err
		}
		updates = append(updates, u)
	}
	if updates == nil {
		updates = []domain.InsightsInitiativeWeekUpdate{}
	}
	return updates, rows.Err()
}

func (r *InsightsRepository) GetDecisionsWithInitiatives(ctx context.Context) ([]domain.InsightsDecisionWithInitiatives, error) {
	if r.db == nil {
		return []domain.InsightsDecisionWithInitiatives{}, nil
	}

	query := `
		SELECT
			d.id,
			d.decision_text,
			COALESCE(d.rationale, ''),
			COALESCE(d.participants, ''),
			COALESCE(d.expected_outcomes, ''),
			d.decision_date,
			d.summary_id,
			d.created_at,
			GROUP_CONCAT(i.name) as initiatives
		FROM decisions d
		LEFT JOIN decision_initiatives di ON d.id = di.decision_id
		LEFT JOIN initiatives i ON di.initiative_id = i.id
		GROUP BY d.id
		ORDER BY d.decision_date DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query decisions with initiatives: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []domain.InsightsDecisionWithInitiatives
	for rows.Next() {
		var d domain.InsightsDecisionWithInitiatives
		var initiatives sql.NullString
		err := rows.Scan(
			&d.ID,
			&d.DecisionText,
			&d.Rationale,
			&d.Participants,
			&d.ExpectedOutcomes,
			&d.DecisionDate,
			&d.SummaryID,
			&d.CreatedAt,
			&initiatives,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan decision: %w", err)
		}
		if initiatives.Valid {
			d.Initiatives = initiatives.String
		}
		results = append(results, d)
	}

	if results == nil {
		return []domain.InsightsDecisionWithInitiatives{}, nil
	}
	return results, rows.Err()
}

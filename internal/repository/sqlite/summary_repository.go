package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type SummaryRepository struct {
	db *sql.DB
}

func NewSummaryRepository(db *sql.DB) *SummaryRepository {
	return &SummaryRepository{db: db}
}

func (r *SummaryRepository) Insert(ctx context.Context, summary domain.Summary) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO summaries (horizon, content, start_date, end_date, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, summary.Horizon, summary.Content,
		summary.StartDate.Format("2006-01-02"),
		summary.EndDate.Format("2006-01-02"),
		summary.CreatedAt.Format(time.RFC3339))

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *SummaryRepository) Get(ctx context.Context, horizon domain.SummaryHorizon, start, end time.Time) (*domain.Summary, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, horizon, content, start_date, end_date, created_at
		FROM summaries
		WHERE horizon = ? AND start_date = ? AND end_date = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, horizon, start.Format("2006-01-02"), end.Format("2006-01-02"))

	return r.scanSummary(row)
}

func (r *SummaryRepository) GetByHorizon(ctx context.Context, horizon domain.SummaryHorizon) ([]domain.Summary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, horizon, content, start_date, end_date, created_at
		FROM summaries
		WHERE horizon = ?
		ORDER BY start_date DESC
	`, horizon)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var summaries []domain.Summary
	for rows.Next() {
		var s domain.Summary
		var horizonStr, startDate, endDate, createdAt string

		err := rows.Scan(&s.ID, &horizonStr, &s.Content, &startDate, &endDate, &createdAt)
		if err != nil {
			return nil, err
		}

		s.Horizon = domain.SummaryHorizon(horizonStr)
		s.StartDate, _ = time.Parse("2006-01-02", startDate)
		s.EndDate, _ = time.Parse("2006-01-02", endDate)
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

		summaries = append(summaries, s)
	}

	return summaries, rows.Err()
}

func (r *SummaryRepository) GetAll(ctx context.Context) ([]domain.Summary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, horizon, content, start_date, end_date, created_at
		FROM summaries
		ORDER BY start_date DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var summaries []domain.Summary
	for rows.Next() {
		var s domain.Summary
		var horizonStr, startDate, endDate, createdAt string

		err := rows.Scan(&s.ID, &horizonStr, &s.Content, &startDate, &endDate, &createdAt)
		if err != nil {
			return nil, err
		}

		s.Horizon = domain.SummaryHorizon(horizonStr)
		s.StartDate, _ = time.Parse("2006-01-02", startDate)
		s.EndDate, _ = time.Parse("2006-01-02", endDate)
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

		summaries = append(summaries, s)
	}

	return summaries, rows.Err()
}

func (r *SummaryRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM summaries WHERE id = ?", id)
	return err
}

func (r *SummaryRepository) DeleteAll(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM summaries")
	return err
}

func (r *SummaryRepository) scanSummary(row *sql.Row) (*domain.Summary, error) {
	var s domain.Summary
	var horizonStr, startDate, endDate, createdAt string

	err := row.Scan(&s.ID, &horizonStr, &s.Content, &startDate, &endDate, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	s.Horizon = domain.SummaryHorizon(horizonStr)
	s.StartDate, _ = time.Parse("2006-01-02", startDate)
	s.EndDate, _ = time.Parse("2006-01-02", endDate)
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	return &s, nil
}

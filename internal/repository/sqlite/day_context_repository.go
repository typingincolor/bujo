package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type DayContextRepository struct {
	db *sql.DB
}

func NewDayContextRepository(db *sql.DB) *DayContextRepository {
	return &DayContextRepository{db: db}
}

func (r *DayContextRepository) Upsert(ctx context.Context, dayCtx domain.DayContext) error {
	dateStr := dayCtx.Date.Format("2006-01-02")

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO day_context (date, location, mood, weather)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(date) DO UPDATE SET
			location = excluded.location,
			mood = excluded.mood,
			weather = excluded.weather
	`, dateStr, dayCtx.Location, dayCtx.Mood, dayCtx.Weather)

	return err
}

func (r *DayContextRepository) GetByDate(ctx context.Context, date time.Time) (*domain.DayContext, error) {
	dateStr := date.Format("2006-01-02")

	row := r.db.QueryRowContext(ctx, `
		SELECT date, location, mood, weather
		FROM day_context WHERE date = ?
	`, dateStr)

	return r.scanDayContext(row)
}

func (r *DayContextRepository) Delete(ctx context.Context, date time.Time) error {
	dateStr := date.Format("2006-01-02")
	_, err := r.db.ExecContext(ctx, "DELETE FROM day_context WHERE date = ?", dateStr)
	return err
}

func (r *DayContextRepository) GetRange(ctx context.Context, start, end time.Time) ([]domain.DayContext, error) {
	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")

	rows, err := r.db.QueryContext(ctx, `
		SELECT date, location, mood, weather
		FROM day_context
		WHERE date >= ? AND date <= ?
		ORDER BY date
	`, startStr, endStr)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var contexts []domain.DayContext
	for rows.Next() {
		var dayCtx domain.DayContext
		var dateStr string
		var location, mood, weather sql.NullString

		err := rows.Scan(&dateStr, &location, &mood, &weather)
		if err != nil {
			return nil, err
		}

		dayCtx.Date, _ = time.Parse("2006-01-02", dateStr)
		if location.Valid {
			dayCtx.Location = &location.String
		}
		if mood.Valid {
			dayCtx.Mood = &mood.String
		}
		if weather.Valid {
			dayCtx.Weather = &weather.String
		}

		contexts = append(contexts, dayCtx)
	}

	return contexts, rows.Err()
}

func (r *DayContextRepository) scanDayContext(row *sql.Row) (*domain.DayContext, error) {
	var dayCtx domain.DayContext
	var dateStr string
	var location, mood, weather sql.NullString

	err := row.Scan(&dateStr, &location, &mood, &weather)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	dayCtx.Date, _ = time.Parse("2006-01-02", dateStr)
	if location.Valid {
		dayCtx.Location = &location.String
	}
	if mood.Valid {
		dayCtx.Mood = &mood.String
	}
	if weather.Valid {
		dayCtx.Weather = &weather.String
	}

	return &dayCtx, nil
}

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
	now := time.Now().Format(time.RFC3339)

	existing, err := r.GetByDate(ctx, dayCtx.Date)
	if err != nil {
		return err
	}

	if existing == nil {
		entityID := dayCtx.EntityID
		if entityID.IsEmpty() {
			entityID = domain.NewEntityID()
		}

		_, err = r.db.ExecContext(ctx, `
			INSERT INTO day_context (date, location, mood, weather, entity_id, version, valid_from, op_type)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, dateStr, dayCtx.Location, dayCtx.Mood, dayCtx.Weather,
			entityID.String(), 1, now, domain.OpTypeInsert.String())
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		UPDATE day_context SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, existing.EntityID.String())
	if err != nil {
		return err
	}

	var maxVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM day_context WHERE entity_id = ?
	`, existing.EntityID.String()).Scan(&maxVersion)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO day_context (date, location, mood, weather, entity_id, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, dateStr, dayCtx.Location, dayCtx.Mood, dayCtx.Weather,
		existing.EntityID.String(), maxVersion+1, now, domain.OpTypeUpdate.String())
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *DayContextRepository) GetByDate(ctx context.Context, date time.Time) (*domain.DayContext, error) {
	dateStr := date.Format("2006-01-02")

	row := r.db.QueryRowContext(ctx, `
		SELECT date, location, mood, weather, entity_id
		FROM day_context WHERE date = ? AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
	`, dateStr)

	return r.scanDayContext(row)
}

func (r *DayContextRepository) Delete(ctx context.Context, date time.Time) error {
	dayCtx, err := r.GetByDate(ctx, date)
	if err != nil {
		return err
	}
	if dayCtx == nil {
		return nil // Already deleted or doesn't exist
	}

	now := time.Now().Format(time.RFC3339)
	dateStr := date.Format("2006-01-02")

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		UPDATE day_context SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, dayCtx.EntityID.String())
	if err != nil {
		return err
	}

	var maxVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM day_context WHERE entity_id = ?
	`, dayCtx.EntityID.String()).Scan(&maxVersion)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO day_context (date, location, mood, weather, entity_id, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, dateStr, dayCtx.Location, dayCtx.Mood, dayCtx.Weather,
		dayCtx.EntityID.String(), maxVersion+1, now, domain.OpTypeDelete.String())
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *DayContextRepository) GetRange(ctx context.Context, start, end time.Time) ([]domain.DayContext, error) {
	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")

	rows, err := r.db.QueryContext(ctx, `
		SELECT date, location, mood, weather, entity_id
		FROM day_context
		WHERE date >= ? AND date <= ?
		AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
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
		var location, mood, weather, entityID sql.NullString

		err := rows.Scan(&dateStr, &location, &mood, &weather, &entityID)
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
		if entityID.Valid {
			dayCtx.EntityID = domain.EntityID(entityID.String)
		}

		contexts = append(contexts, dayCtx)
	}

	return contexts, rows.Err()
}

func (r *DayContextRepository) GetAll(ctx context.Context) ([]domain.DayContext, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT date, location, mood, weather, entity_id
		FROM day_context
		WHERE (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
		ORDER BY date
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var contexts []domain.DayContext
	for rows.Next() {
		var dayCtx domain.DayContext
		var dateStr string
		var location, mood, weather, entityID sql.NullString

		err := rows.Scan(&dateStr, &location, &mood, &weather, &entityID)
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
		if entityID.Valid {
			dayCtx.EntityID = domain.EntityID(entityID.String)
		}

		contexts = append(contexts, dayCtx)
	}

	return contexts, rows.Err()
}

func (r *DayContextRepository) DeleteAll(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM day_context")
	return err
}

func (r *DayContextRepository) scanDayContext(row *sql.Row) (*domain.DayContext, error) {
	var dayCtx domain.DayContext
	var dateStr string
	var location, mood, weather, entityID sql.NullString

	err := row.Scan(&dateStr, &location, &mood, &weather, &entityID)
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
	if entityID.Valid {
		dayCtx.EntityID = domain.EntityID(entityID.String)
	}

	return &dayCtx, nil
}

func (r *DayContextRepository) GetDeleted(ctx context.Context) ([]domain.DayContext, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT date, location, mood, weather, entity_id
		FROM day_context
		WHERE op_type = 'DELETE'
		AND valid_to IS NULL
		ORDER BY valid_from DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var contexts []domain.DayContext
	for rows.Next() {
		var dayCtx domain.DayContext
		var dateStr string
		var location, mood, weather, entityID sql.NullString

		err := rows.Scan(&dateStr, &location, &mood, &weather, &entityID)
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
		if entityID.Valid {
			dayCtx.EntityID = domain.EntityID(entityID.String)
		}

		contexts = append(contexts, dayCtx)
	}

	return contexts, rows.Err()
}

func (r *DayContextRepository) GetLastModified(ctx context.Context) (time.Time, error) {
	var validFrom sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT MAX(valid_from) FROM day_context`).Scan(&validFrom)
	if err != nil {
		return time.Time{}, err
	}
	if !validFrom.Valid {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, validFrom.String)
}

func (r *DayContextRepository) Restore(ctx context.Context, entityID domain.EntityID) error {
	now := time.Now().Format(time.RFC3339)

	var lastCtx struct {
		Date     string
		Location sql.NullString
		Mood     sql.NullString
		Weather  sql.NullString
		Version  int
		OpType   string
	}

	err := r.db.QueryRowContext(ctx, `
		SELECT date, location, mood, weather, version, op_type
		FROM day_context WHERE entity_id = ?
		ORDER BY version DESC LIMIT 1
	`, entityID.String()).Scan(
		&lastCtx.Date, &lastCtx.Location, &lastCtx.Mood, &lastCtx.Weather,
		&lastCtx.Version, &lastCtx.OpType)
	if err != nil {
		return err
	}

	if lastCtx.OpType != domain.OpTypeDelete.String() {
		return nil // Not deleted, nothing to restore
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		UPDATE day_context SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, entityID.String())
	if err != nil {
		return err
	}

	var location, mood, weather *string
	if lastCtx.Location.Valid {
		location = &lastCtx.Location.String
	}
	if lastCtx.Mood.Valid {
		mood = &lastCtx.Mood.String
	}
	if lastCtx.Weather.Valid {
		weather = &lastCtx.Weather.String
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO day_context (date, location, mood, weather, entity_id, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, lastCtx.Date, location, mood, weather,
		entityID.String(), lastCtx.Version+1, now, domain.OpTypeInsert.String())
	if err != nil {
		return err
	}

	return tx.Commit()
}

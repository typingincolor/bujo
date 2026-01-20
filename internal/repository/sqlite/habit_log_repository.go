package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type HabitLogRepository struct {
	db *sql.DB
}

func NewHabitLogRepository(db *sql.DB) *HabitLogRepository {
	return &HabitLogRepository{db: db}
}

func (r *HabitLogRepository) Insert(ctx context.Context, log domain.HabitLog) (int64, error) {
	entityID := log.EntityID
	if entityID.IsEmpty() {
		entityID = domain.NewEntityID()
	}
	now := time.Now().Format(time.RFC3339)

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO habit_logs (habit_id, count, logged_at, entity_id, habit_entity_id, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, log.HabitID, log.Count, log.LoggedAt.Format(time.RFC3339),
		entityID.String(), log.HabitEntityID.String(), 1, now, domain.OpTypeInsert.String())

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *HabitLogRepository) GetByID(ctx context.Context, id int64) (*domain.HabitLog, error) {
	var entityID string
	err := r.db.QueryRowContext(ctx, `
		SELECT entity_id FROM habit_logs WHERE id = ?
	`, id).Scan(&entityID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return r.GetByEntityID(ctx, domain.EntityID(entityID))
}

func (r *HabitLogRepository) GetByEntityID(ctx context.Context, entityID domain.EntityID) (*domain.HabitLog, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, habit_id, count, logged_at, entity_id, habit_entity_id
		FROM habit_logs WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
	`, entityID.String())

	var log domain.HabitLog
	var loggedAt string
	var scannedEntityID, habitEntityID sql.NullString

	err := row.Scan(&log.ID, &log.HabitID, &log.Count, &loggedAt, &scannedEntityID, &habitEntityID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	log.LoggedAt, _ = time.Parse(time.RFC3339, loggedAt)
	if scannedEntityID.Valid {
		log.EntityID = domain.EntityID(scannedEntityID.String)
	}
	if habitEntityID.Valid {
		log.HabitEntityID = domain.EntityID(habitEntityID.String)
	}
	return &log, nil
}

func (r *HabitLogRepository) GetByHabitID(ctx context.Context, habitID int64) ([]domain.HabitLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, habit_id, count, logged_at, entity_id, habit_entity_id
		FROM habit_logs WHERE habit_id = ? AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
		ORDER BY logged_at
	`, habitID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanLogs(rows)
}

func (r *HabitLogRepository) GetRange(ctx context.Context, habitID int64, start, end time.Time) ([]domain.HabitLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, habit_id, count, logged_at, entity_id, habit_entity_id
		FROM habit_logs
		WHERE habit_id = ? AND logged_at >= ? AND logged_at <= ?
		AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
		ORDER BY logged_at
	`, habitID, start.Format(time.RFC3339), end.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanLogs(rows)
}

func (r *HabitLogRepository) GetRangeByEntityID(ctx context.Context, habitEntityID domain.EntityID, start, end time.Time) ([]domain.HabitLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, habit_id, count, logged_at, entity_id, habit_entity_id
		FROM habit_logs
		WHERE habit_entity_id = ? AND logged_at >= ? AND logged_at <= ?
		AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
		ORDER BY logged_at
	`, habitEntityID.String(), start.Format(time.RFC3339), end.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanLogs(rows)
}

func (r *HabitLogRepository) GetAllRange(ctx context.Context, start, end time.Time) ([]domain.HabitLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, habit_id, count, logged_at, entity_id, habit_entity_id
		FROM habit_logs
		WHERE logged_at >= ? AND logged_at <= ?
		AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
		ORDER BY logged_at
	`, start.Format(time.RFC3339), end.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanLogs(rows)
}

func (r *HabitLogRepository) GetAll(ctx context.Context) ([]domain.HabitLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, habit_id, count, logged_at, entity_id, habit_entity_id
		FROM habit_logs
		WHERE (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
		ORDER BY logged_at
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanLogs(rows)
}

func (r *HabitLogRepository) DeleteAll(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM habit_logs")
	return err
}

func (r *HabitLogRepository) Delete(ctx context.Context, id int64) error {
	log, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if log == nil {
		return nil // Already deleted or doesn't exist
	}

	now := time.Now().Format(time.RFC3339)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		UPDATE habit_logs SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, log.EntityID.String())
	if err != nil {
		return err
	}

	var maxVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM habit_logs WHERE entity_id = ?
	`, log.EntityID.String()).Scan(&maxVersion)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO habit_logs (habit_id, count, logged_at, entity_id, habit_entity_id, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, log.HabitID, log.Count, log.LoggedAt.Format(time.RFC3339),
		log.EntityID.String(), log.HabitEntityID.String(), maxVersion+1, now, domain.OpTypeDelete.String())
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *HabitLogRepository) GetLastByHabitID(ctx context.Context, habitID int64) (*domain.HabitLog, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, habit_id, count, logged_at, entity_id, habit_entity_id
		FROM habit_logs
		WHERE habit_id = ?
		AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
		ORDER BY logged_at DESC, id DESC
		LIMIT 1
	`, habitID)

	var log domain.HabitLog
	var loggedAt string
	var entityID, habitEntityID sql.NullString

	err := row.Scan(&log.ID, &log.HabitID, &log.Count, &loggedAt, &entityID, &habitEntityID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	log.LoggedAt, _ = time.Parse(time.RFC3339, loggedAt)
	if entityID.Valid {
		log.EntityID = domain.EntityID(entityID.String)
	}
	if habitEntityID.Valid {
		log.HabitEntityID = domain.EntityID(habitEntityID.String)
	}
	return &log, nil
}

func (r *HabitLogRepository) GetDeleted(ctx context.Context) ([]domain.HabitLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, habit_id, count, logged_at, entity_id, habit_entity_id
		FROM habit_logs
		WHERE op_type = 'DELETE'
		AND valid_to IS NULL
		ORDER BY valid_from DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanLogs(rows)
}

func (r *HabitLogRepository) Restore(ctx context.Context, entityID domain.EntityID) (int64, error) {
	now := time.Now().Format(time.RFC3339)

	var lastLog struct {
		HabitID       int64
		Count         int
		LoggedAt      string
		HabitEntityID sql.NullString
		Version       int
		OpType        string
	}

	err := r.db.QueryRowContext(ctx, `
		SELECT habit_id, count, logged_at, habit_entity_id, version, op_type
		FROM habit_logs WHERE entity_id = ?
		ORDER BY version DESC LIMIT 1
	`, entityID.String()).Scan(
		&lastLog.HabitID, &lastLog.Count, &lastLog.LoggedAt,
		&lastLog.HabitEntityID, &lastLog.Version, &lastLog.OpType)
	if err != nil {
		return 0, err
	}

	if lastLog.OpType != domain.OpTypeDelete.String() {
		return 0, nil // Not deleted, nothing to restore
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		UPDATE habit_logs SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, entityID.String())
	if err != nil {
		return 0, err
	}

	var habitEntityID *string
	if lastLog.HabitEntityID.Valid {
		habitEntityID = &lastLog.HabitEntityID.String
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO habit_logs (habit_id, count, logged_at, entity_id, habit_entity_id, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, lastLog.HabitID, lastLog.Count, lastLog.LoggedAt,
		entityID.String(), habitEntityID, lastLog.Version+1, now, domain.OpTypeInsert.String())
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *HabitLogRepository) GetLastModified(ctx context.Context) (time.Time, error) {
	var validFrom sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT MAX(valid_from) FROM habit_logs`).Scan(&validFrom)
	if err != nil {
		return time.Time{}, err
	}
	if !validFrom.Valid {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, validFrom.String)
}

func (r *HabitLogRepository) scanLogs(rows *sql.Rows) ([]domain.HabitLog, error) {
	var logs []domain.HabitLog

	for rows.Next() {
		var log domain.HabitLog
		var loggedAt string
		var entityID, habitEntityID sql.NullString

		err := rows.Scan(&log.ID, &log.HabitID, &log.Count, &loggedAt, &entityID, &habitEntityID)
		if err != nil {
			return nil, err
		}

		log.LoggedAt, _ = time.Parse(time.RFC3339, loggedAt)
		if entityID.Valid {
			log.EntityID = domain.EntityID(entityID.String)
		}
		if habitEntityID.Valid {
			log.HabitEntityID = domain.EntityID(habitEntityID.String)
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

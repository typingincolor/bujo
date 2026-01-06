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
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO habit_logs (habit_id, count, logged_at)
		VALUES (?, ?, ?)
	`, log.HabitID, log.Count, log.LoggedAt.Format(time.RFC3339))

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *HabitLogRepository) GetByHabitID(ctx context.Context, habitID int64) ([]domain.HabitLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, habit_id, count, logged_at
		FROM habit_logs WHERE habit_id = ?
		ORDER BY logged_at
	`, habitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

func (r *HabitLogRepository) GetRange(ctx context.Context, habitID int64, start, end time.Time) ([]domain.HabitLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, habit_id, count, logged_at
		FROM habit_logs
		WHERE habit_id = ? AND logged_at >= ? AND logged_at <= ?
		ORDER BY logged_at
	`, habitID, start.Format(time.RFC3339), end.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

func (r *HabitLogRepository) GetAllRange(ctx context.Context, start, end time.Time) ([]domain.HabitLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, habit_id, count, logged_at
		FROM habit_logs
		WHERE logged_at >= ? AND logged_at <= ?
		ORDER BY logged_at
	`, start.Format(time.RFC3339), end.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

func (r *HabitLogRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM habit_logs WHERE id = ?", id)
	return err
}

func (r *HabitLogRepository) scanLogs(rows *sql.Rows) ([]domain.HabitLog, error) {
	var logs []domain.HabitLog

	for rows.Next() {
		var log domain.HabitLog
		var loggedAt string

		err := rows.Scan(&log.ID, &log.HabitID, &log.Count, &loggedAt)
		if err != nil {
			return nil, err
		}

		log.LoggedAt, _ = time.Parse(time.RFC3339, loggedAt)
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

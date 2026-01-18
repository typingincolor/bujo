package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type GoalRepository struct {
	db *sql.DB
}

func NewGoalRepository(db *sql.DB) *GoalRepository {
	return &GoalRepository{db: db}
}

func (r *GoalRepository) Insert(ctx context.Context, goal domain.Goal) (int64, error) {
	entityID := goal.EntityID
	if entityID.IsEmpty() {
		entityID = domain.NewEntityID()
	}
	now := time.Now().Format(time.RFC3339)
	monthKey := goal.Month.Format("2006-01")

	status := goal.Status
	if status == "" {
		status = domain.GoalStatusActive
	}

	var migratedTo *string
	if goal.MigratedTo != nil {
		mt := goal.MigratedTo.Format("2006-01")
		migratedTo = &mt
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO goals (entity_id, content, month, status, migrated_to, created_at, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, entityID.String(), goal.Content, monthKey, string(status), migratedTo, goal.CreatedAt.Format(time.RFC3339),
		1, now, domain.OpTypeInsert.String())

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *GoalRepository) GetByID(ctx context.Context, id int64) (*domain.Goal, error) {
	var entityID string
	err := r.db.QueryRowContext(ctx, `SELECT entity_id FROM goals WHERE id = ?`, id).Scan(&entityID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT id, entity_id, content, month, status, migrated_to, created_at
		FROM goals WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
	`, entityID)

	return r.scanGoal(row)
}

func (r *GoalRepository) GetByMonth(ctx context.Context, month time.Time) ([]domain.Goal, error) {
	monthKey := month.Format("2006-01")

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, entity_id, content, month, status, migrated_to, created_at
		FROM goals WHERE month = ? AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
		ORDER BY created_at
	`, monthKey)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var goals []domain.Goal
	for rows.Next() {
		var goal domain.Goal
		var entityID sql.NullString
		var migratedTo sql.NullString
		var monthStr, statusStr, createdAt string

		err := rows.Scan(&goal.ID, &entityID, &goal.Content, &monthStr, &statusStr, &migratedTo, &createdAt)
		if err != nil {
			return nil, err
		}

		if entityID.Valid {
			goal.EntityID = domain.EntityID(entityID.String)
		}
		goal.Month, _ = time.Parse("2006-01", monthStr)
		goal.Status = domain.GoalStatus(statusStr)
		if migratedTo.Valid {
			mt, _ := time.Parse("2006-01", migratedTo.String)
			goal.MigratedTo = &mt
		}
		goal.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

		goals = append(goals, goal)
	}

	return goals, rows.Err()
}

func (r *GoalRepository) GetAll(ctx context.Context) ([]domain.Goal, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, entity_id, content, month, status, migrated_to, created_at
		FROM goals WHERE (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
		ORDER BY month DESC, created_at
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var goals []domain.Goal
	for rows.Next() {
		var goal domain.Goal
		var entityID sql.NullString
		var migratedTo sql.NullString
		var monthStr, statusStr, createdAt string

		err := rows.Scan(&goal.ID, &entityID, &goal.Content, &monthStr, &statusStr, &migratedTo, &createdAt)
		if err != nil {
			return nil, err
		}

		if entityID.Valid {
			goal.EntityID = domain.EntityID(entityID.String)
		}
		goal.Month, _ = time.Parse("2006-01", monthStr)
		goal.Status = domain.GoalStatus(statusStr)
		if migratedTo.Valid {
			mt, _ := time.Parse("2006-01", migratedTo.String)
			goal.MigratedTo = &mt
		}
		goal.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

		goals = append(goals, goal)
	}

	return goals, rows.Err()
}

func (r *GoalRepository) Update(ctx context.Context, goal domain.Goal) error {
	current, err := r.GetByID(ctx, goal.ID)
	if err != nil {
		return err
	}
	if current == nil {
		return nil
	}

	now := time.Now().Format(time.RFC3339)
	monthKey := goal.Month.Format("2006-01")

	var migratedTo *string
	if goal.MigratedTo != nil {
		mt := goal.MigratedTo.Format("2006-01")
		migratedTo = &mt
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		UPDATE goals SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, current.EntityID.String())
	if err != nil {
		return err
	}

	var maxVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM goals WHERE entity_id = ?
	`, current.EntityID.String()).Scan(&maxVersion)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO goals (entity_id, content, month, status, migrated_to, created_at, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, current.EntityID.String(), goal.Content, monthKey, string(goal.Status), migratedTo,
		current.CreatedAt.Format(time.RFC3339), maxVersion+1, now, domain.OpTypeUpdate.String())
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *GoalRepository) Delete(ctx context.Context, id int64) error {
	goal, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if goal == nil {
		return nil
	}

	now := time.Now().Format(time.RFC3339)

	var migratedTo *string
	if goal.MigratedTo != nil {
		mt := goal.MigratedTo.Format("2006-01")
		migratedTo = &mt
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
		UPDATE goals SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, goal.EntityID.String())
	if err != nil {
		return err
	}

	var maxVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM goals WHERE entity_id = ?
	`, goal.EntityID.String()).Scan(&maxVersion)
	if err != nil {
		return err
	}

	monthKey := goal.Month.Format("2006-01")
	_, err = tx.ExecContext(ctx, `
		INSERT INTO goals (entity_id, content, month, status, migrated_to, created_at, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, goal.EntityID.String(), goal.Content, monthKey, string(goal.Status), migratedTo,
		goal.CreatedAt.Format(time.RFC3339), maxVersion+1, now, domain.OpTypeDelete.String())
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *GoalRepository) DeleteAll(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM goals")
	return err
}

func (r *GoalRepository) GetByEntityID(ctx context.Context, entityID domain.EntityID) (*domain.Goal, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, entity_id, content, month, status, migrated_to, created_at
		FROM goals WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
	`, entityID.String())

	return r.scanGoal(row)
}

func (r *GoalRepository) MoveToMonth(ctx context.Context, id int64, newMonth time.Time) error {
	goal, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if goal == nil {
		return nil
	}

	goal.Month = newMonth
	return r.Update(ctx, *goal)
}

func (r *GoalRepository) scanGoal(row *sql.Row) (*domain.Goal, error) {
	var goal domain.Goal
	var entityID sql.NullString
	var migratedTo sql.NullString
	var monthStr, statusStr, createdAt string

	err := row.Scan(&goal.ID, &entityID, &goal.Content, &monthStr, &statusStr, &migratedTo, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if entityID.Valid {
		goal.EntityID = domain.EntityID(entityID.String)
	}
	goal.Month, _ = time.Parse("2006-01", monthStr)
	goal.Status = domain.GoalStatus(statusStr)
	if migratedTo.Valid {
		mt, _ := time.Parse("2006-01", migratedTo.String)
		goal.MigratedTo = &mt
	}
	goal.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	return &goal, nil
}

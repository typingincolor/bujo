package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type HabitRepository struct {
	db *sql.DB
}

func NewHabitRepository(db *sql.DB) *HabitRepository {
	return &HabitRepository{db: db}
}

func (r *HabitRepository) Insert(ctx context.Context, habit domain.Habit) (int64, error) {
	entityID := habit.EntityID
	if entityID.IsEmpty() {
		entityID = domain.NewEntityID()
	}
	now := time.Now().Format(time.RFC3339)

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO habits (name, goal_per_day, goal_per_week, goal_per_month, created_at, entity_id, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, habit.Name, habit.GoalPerDay, habit.GoalPerWeek, habit.GoalPerMonth, habit.CreatedAt.Format(time.RFC3339),
		entityID.String(), 1, now, domain.OpTypeInsert.String())

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *HabitRepository) GetByID(ctx context.Context, id int64) (*domain.Habit, error) {
	// First, get the entity_id for this ID (may be from a closed version)
	var entityID string
	err := r.db.QueryRowContext(ctx, `
		SELECT entity_id FROM habits WHERE id = ?
	`, id).Scan(&entityID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Then get the current version for that entity
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, goal_per_day, goal_per_week, goal_per_month, created_at, entity_id
		FROM habits WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
	`, entityID)

	return r.scanHabit(row)
}

func (r *HabitRepository) GetByName(ctx context.Context, name string) (*domain.Habit, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, goal_per_day, goal_per_week, goal_per_month, created_at, entity_id
		FROM habits WHERE name = ? AND (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
	`, name)

	return r.scanHabit(row)
}

func (r *HabitRepository) GetOrCreate(ctx context.Context, name string, goalPerDay int) (*domain.Habit, error) {
	existing, err := r.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	habit := domain.Habit{
		Name:       name,
		GoalPerDay: goalPerDay,
		CreatedAt:  time.Now(),
		EntityID:   domain.NewEntityID(),
	}

	id, err := r.Insert(ctx, habit)
	if err != nil {
		return nil, err
	}

	habit.ID = id
	return &habit, nil
}

func (r *HabitRepository) GetAll(ctx context.Context) ([]domain.Habit, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, goal_per_day, goal_per_week, goal_per_month, created_at, entity_id
		FROM habits WHERE (valid_to IS NULL OR valid_to = '') AND op_type != 'DELETE'
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var habits []domain.Habit
	for rows.Next() {
		var habit domain.Habit
		var createdAt string
		var entityID sql.NullString

		err := rows.Scan(&habit.ID, &habit.Name, &habit.GoalPerDay, &habit.GoalPerWeek, &habit.GoalPerMonth, &createdAt, &entityID)
		if err != nil {
			return nil, err
		}

		habit.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		if entityID.Valid {
			habit.EntityID = domain.EntityID(entityID.String)
		}
		habits = append(habits, habit)
	}

	return habits, rows.Err()
}

func (r *HabitRepository) Update(ctx context.Context, habit domain.Habit) error {
	current, err := r.GetByID(ctx, habit.ID)
	if err != nil {
		return err
	}
	if current == nil {
		return fmt.Errorf("habit not found: %d", habit.ID)
	}

	now := time.Now().Format(time.RFC3339)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Close current version
	_, err = tx.ExecContext(ctx, `
		UPDATE habits SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, current.EntityID.String())
	if err != nil {
		return err
	}

	// Get next version number
	var maxVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM habits WHERE entity_id = ?
	`, current.EntityID.String()).Scan(&maxVersion)
	if err != nil {
		return err
	}

	// Insert new version with UPDATE op_type
	_, err = tx.ExecContext(ctx, `
		INSERT INTO habits (name, goal_per_day, goal_per_week, goal_per_month, created_at, entity_id, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, habit.Name, habit.GoalPerDay, habit.GoalPerWeek, habit.GoalPerMonth, current.CreatedAt.Format(time.RFC3339),
		current.EntityID.String(), maxVersion+1, now, domain.OpTypeUpdate.String())
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *HabitRepository) Delete(ctx context.Context, id int64) error {
	habit, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if habit == nil {
		return nil // Already deleted or doesn't exist
	}

	now := time.Now().Format(time.RFC3339)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Close current version
	_, err = tx.ExecContext(ctx, `
		UPDATE habits SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, habit.EntityID.String())
	if err != nil {
		return err
	}

	// Get next version number
	var maxVersion int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM habits WHERE entity_id = ?
	`, habit.EntityID.String()).Scan(&maxVersion)
	if err != nil {
		return err
	}

	// Insert delete marker
	_, err = tx.ExecContext(ctx, `
		INSERT INTO habits (name, goal_per_day, goal_per_week, goal_per_month, created_at, entity_id, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, habit.Name, habit.GoalPerDay, habit.GoalPerWeek, habit.GoalPerMonth, habit.CreatedAt.Format(time.RFC3339),
		habit.EntityID.String(), maxVersion+1, now, domain.OpTypeDelete.String())
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *HabitRepository) GetDeleted(ctx context.Context) ([]domain.Habit, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, goal_per_day, goal_per_week, goal_per_month, created_at, entity_id
		FROM habits
		WHERE op_type = 'DELETE'
		AND valid_to IS NULL
		ORDER BY valid_from DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var habits []domain.Habit
	for rows.Next() {
		var habit domain.Habit
		var createdAt string
		var entityID sql.NullString

		err := rows.Scan(&habit.ID, &habit.Name, &habit.GoalPerDay, &habit.GoalPerWeek, &habit.GoalPerMonth, &createdAt, &entityID)
		if err != nil {
			return nil, err
		}

		habit.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		if entityID.Valid {
			habit.EntityID = domain.EntityID(entityID.String)
		}
		habits = append(habits, habit)
	}

	return habits, rows.Err()
}

func (r *HabitRepository) Restore(ctx context.Context, entityID domain.EntityID) (int64, error) {
	now := time.Now().Format(time.RFC3339)

	// Get the most recent version (which should be a DELETE marker)
	var lastHabit struct {
		Name         string
		GoalPerDay   int
		GoalPerWeek  int
		GoalPerMonth int
		CreatedAt    string
		Version      int
		OpType       string
	}

	err := r.db.QueryRowContext(ctx, `
		SELECT name, goal_per_day, goal_per_week, goal_per_month, created_at, version, op_type
		FROM habits WHERE entity_id = ?
		ORDER BY version DESC LIMIT 1
	`, entityID.String()).Scan(
		&lastHabit.Name, &lastHabit.GoalPerDay, &lastHabit.GoalPerWeek, &lastHabit.GoalPerMonth, &lastHabit.CreatedAt,
		&lastHabit.Version, &lastHabit.OpType)
	if err != nil {
		return 0, err
	}

	if lastHabit.OpType != domain.OpTypeDelete.String() {
		return 0, nil // Not deleted, nothing to restore
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	// Close the DELETE marker
	_, err = tx.ExecContext(ctx, `
		UPDATE habits SET valid_to = ? WHERE entity_id = ? AND (valid_to IS NULL OR valid_to = '')
	`, now, entityID.String())
	if err != nil {
		return 0, err
	}

	// Insert a new version with INSERT op_type to restore
	result, err := tx.ExecContext(ctx, `
		INSERT INTO habits (name, goal_per_day, goal_per_week, goal_per_month, created_at, entity_id, version, valid_from, op_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, lastHabit.Name, lastHabit.GoalPerDay, lastHabit.GoalPerWeek, lastHabit.GoalPerMonth, lastHabit.CreatedAt,
		entityID.String(), lastHabit.Version+1, now, domain.OpTypeInsert.String())
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *HabitRepository) scanHabit(row *sql.Row) (*domain.Habit, error) {
	var habit domain.Habit
	var createdAt string
	var entityID sql.NullString

	err := row.Scan(&habit.ID, &habit.Name, &habit.GoalPerDay, &habit.GoalPerWeek, &habit.GoalPerMonth, &createdAt, &entityID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	habit.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	if entityID.Valid {
		habit.EntityID = domain.EntityID(entityID.String)
	}
	return &habit, nil
}

package sqlite

import (
	"context"
	"database/sql"
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
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO habits (name, goal_per_day, created_at)
		VALUES (?, ?, ?)
	`, habit.Name, habit.GoalPerDay, habit.CreatedAt.Format(time.RFC3339))

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *HabitRepository) GetByID(ctx context.Context, id int64) (*domain.Habit, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, goal_per_day, created_at
		FROM habits WHERE id = ?
	`, id)

	return r.scanHabit(row)
}

func (r *HabitRepository) GetByName(ctx context.Context, name string) (*domain.Habit, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, goal_per_day, created_at
		FROM habits WHERE name = ?
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
		SELECT id, name, goal_per_day, created_at
		FROM habits ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var habits []domain.Habit
	for rows.Next() {
		var habit domain.Habit
		var createdAt string

		err := rows.Scan(&habit.ID, &habit.Name, &habit.GoalPerDay, &createdAt)
		if err != nil {
			return nil, err
		}

		habit.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		habits = append(habits, habit)
	}

	return habits, rows.Err()
}

func (r *HabitRepository) Update(ctx context.Context, habit domain.Habit) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE habits SET name = ?, goal_per_day = ? WHERE id = ?
	`, habit.Name, habit.GoalPerDay, habit.ID)

	return err
}

func (r *HabitRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM habits WHERE id = ?", id)
	return err
}

func (r *HabitRepository) scanHabit(row *sql.Row) (*domain.Habit, error) {
	var habit domain.Habit
	var createdAt string

	err := row.Scan(&habit.ID, &habit.Name, &habit.GoalPerDay, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	habit.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &habit, nil
}

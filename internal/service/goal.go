package service

import (
	"context"
	"fmt"
	"time"

	"github.com/typingincolor/bujo/internal/domain"
)

type GoalRepository interface {
	Insert(ctx context.Context, goal domain.Goal) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Goal, error)
	GetByMonth(ctx context.Context, month time.Time) ([]domain.Goal, error)
	GetAll(ctx context.Context) ([]domain.Goal, error)
	Update(ctx context.Context, goal domain.Goal) error
	Delete(ctx context.Context, id int64) error
	MoveToMonth(ctx context.Context, id int64, newMonth time.Time) error
}

type GoalService struct {
	goalRepo GoalRepository
}

func NewGoalService(goalRepo GoalRepository) *GoalService {
	return &GoalService{
		goalRepo: goalRepo,
	}
}

func (s *GoalService) CreateGoal(ctx context.Context, content string, month time.Time) (int64, error) {
	goal := domain.Goal{
		Content:   content,
		Month:     month,
		Status:    domain.GoalStatusActive,
		CreatedAt: time.Now(),
	}

	if err := goal.Validate(); err != nil {
		return 0, err
	}

	return s.goalRepo.Insert(ctx, goal)
}

func (s *GoalService) GetGoal(ctx context.Context, id int64) (*domain.Goal, error) {
	return s.goalRepo.GetByID(ctx, id)
}

func (s *GoalService) GetGoalsForMonth(ctx context.Context, month time.Time) ([]domain.Goal, error) {
	return s.goalRepo.GetByMonth(ctx, month)
}

func (s *GoalService) GetCurrentMonthGoals(ctx context.Context) ([]domain.Goal, error) {
	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return s.goalRepo.GetByMonth(ctx, currentMonth)
}

func (s *GoalService) GetAllGoals(ctx context.Context) ([]domain.Goal, error) {
	return s.goalRepo.GetAll(ctx)
}

func (s *GoalService) MarkDone(ctx context.Context, id int64) error {
	goal, err := s.goalRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if goal == nil {
		return fmt.Errorf("goal not found: %d", id)
	}

	updated := goal.MarkDone()
	return s.goalRepo.Update(ctx, updated)
}

func (s *GoalService) MarkActive(ctx context.Context, id int64) error {
	goal, err := s.goalRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if goal == nil {
		return fmt.Errorf("goal not found: %d", id)
	}

	updated := goal.MarkActive()
	return s.goalRepo.Update(ctx, updated)
}

func (s *GoalService) MoveToMonth(ctx context.Context, id int64, newMonth time.Time) error {
	goal, err := s.goalRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if goal == nil {
		return fmt.Errorf("goal not found: %d", id)
	}

	return s.goalRepo.MoveToMonth(ctx, id, newMonth)
}

func (s *GoalService) DeleteGoal(ctx context.Context, id int64) error {
	goal, err := s.goalRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if goal == nil {
		return fmt.Errorf("goal not found: %d", id)
	}

	return s.goalRepo.Delete(ctx, id)
}

func (s *GoalService) UpdateContent(ctx context.Context, id int64, content string) error {
	goal, err := s.goalRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if goal == nil {
		return fmt.Errorf("goal not found: %d", id)
	}

	goal.Content = content
	return s.goalRepo.Update(ctx, *goal)
}

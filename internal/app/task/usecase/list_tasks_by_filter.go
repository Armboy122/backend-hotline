package usecase

import (
	"context"

	taskdomain "backend-hotlines3/internal/domain/task"
	"backend-hotlines3/internal/port/outbound/repository"
)

type ListTasksByFilterInput struct {
	Year   string
	Month  string
	TeamID *int64
}

type ListTasksByFilterOutput struct {
	Tasks []taskdomain.Entity
}

type ListTasksByFilterUseCase struct {
	repo repository.TaskRepository
}

func NewListTasksByFilterUseCase(repo repository.TaskRepository) *ListTasksByFilterUseCase {
	return &ListTasksByFilterUseCase{repo: repo}
}

func (uc *ListTasksByFilterUseCase) Execute(ctx context.Context, input ListTasksByFilterInput) (*ListTasksByFilterOutput, error) {
	if input.Year == "" || input.Month == "" {
		return nil, ErrYearMonthRequired
	}

	tasks, err := uc.repo.ListByFilter(ctx, repository.TaskByFilterQuery{
		Year:   input.Year,
		Month:  input.Month,
		TeamID: input.TeamID,
	})
	if err != nil {
		return nil, err
	}

	return &ListTasksByFilterOutput{Tasks: tasks}, nil
}

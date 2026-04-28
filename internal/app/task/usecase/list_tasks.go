package usecase

import (
	"context"

	taskdomain "backend-hotlines3/internal/domain/task"
	"backend-hotlines3/internal/port/outbound/repository"
)

type ListTasksInput struct {
	Page   int
	Limit  int
	Filter repository.TaskListFilter
}

type ListTasksOutput struct {
	Tasks []taskdomain.Entity
	Total int64
	Page  int
	Limit int
}

type ListTasksUseCase struct {
	repo repository.TaskRepository
}

func NewListTasksUseCase(repo repository.TaskRepository) *ListTasksUseCase {
	return &ListTasksUseCase{repo: repo}
}

func (uc *ListTasksUseCase) Execute(ctx context.Context, input ListTasksInput) (*ListTasksOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 || input.Limit > 100 {
		input.Limit = 50
	}

	query := repository.TaskListQuery{
		Page:   input.Page,
		Limit:  input.Limit,
		Filter: input.Filter,
	}

	tasks, total, err := uc.repo.List(ctx, query)
	if err != nil {
		return nil, err
	}

	return &ListTasksOutput{
		Tasks: tasks,
		Total: total,
		Page:  input.Page,
		Limit: input.Limit,
	}, nil
}

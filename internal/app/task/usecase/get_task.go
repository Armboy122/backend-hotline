package usecase

import (
	"context"

	taskdomain "backend-hotlines3/internal/domain/task"
	"backend-hotlines3/internal/port/outbound/repository"
)

type GetTaskInput struct {
	ID int64
}

type GetTaskUseCase struct {
	repo repository.TaskRepository
}

func NewGetTaskUseCase(repo repository.TaskRepository) *GetTaskUseCase {
	return &GetTaskUseCase{repo: repo}
}

func (uc *GetTaskUseCase) Execute(ctx context.Context, input GetTaskInput) (*taskdomain.Entity, error) {
	if input.ID <= 0 {
		return nil, ErrInvalidTaskID
	}
	entity, err := uc.repo.GetByID(ctx, repository.TaskGetQuery{ID: input.ID})
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, ErrTaskNotFound
	}
	return entity, nil
}

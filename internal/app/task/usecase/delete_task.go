package usecase

import (
	"context"

	"backend-hotlines3/internal/port/outbound/repository"
)

type DeleteTaskInput struct {
	ID int64
}

type DeleteTaskUseCase struct {
	repo repository.TaskRepository
}

func NewDeleteTaskUseCase(repo repository.TaskRepository) *DeleteTaskUseCase {
	return &DeleteTaskUseCase{repo: repo}
}

func (uc *DeleteTaskUseCase) Execute(ctx context.Context, input DeleteTaskInput) error {
	if input.ID <= 0 {
		return ErrInvalidTaskID
	}
	return uc.repo.SoftDelete(ctx, repository.TaskDeleteCommand{ID: input.ID})
}

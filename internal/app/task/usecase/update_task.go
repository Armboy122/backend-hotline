package usecase

import (
	"context"
	"time"

	taskdomain "backend-hotlines3/internal/domain/task"
	"backend-hotlines3/internal/port/outbound/repository"
)

type UpdateTaskInput struct {
	ID          int64
	WorkDate    *time.Time
	TeamID      *int64
	JobTypeID   *int64
	JobDetailID *int64
	FeederID    *int64
	NumPole     *string
	DeviceCode  *string
	Detail      *string
	URLsBefore  []string
	URLsAfter   []string
	Latitude    *float64
	Longitude   *float64
}

type UpdateTaskUseCase struct {
	repo repository.TaskRepository
}

func NewUpdateTaskUseCase(repo repository.TaskRepository) *UpdateTaskUseCase {
	return &UpdateTaskUseCase{repo: repo}
}

func (uc *UpdateTaskUseCase) Execute(ctx context.Context, input UpdateTaskInput) (*taskdomain.Entity, error) {
	if input.ID <= 0 {
		return nil, ErrInvalidTaskID
	}

	repoInput := repository.TaskUpdateInput{
		ID:          input.ID,
		WorkDate:    input.WorkDate,
		TeamID:      input.TeamID,
		JobTypeID:   input.JobTypeID,
		JobDetailID: input.JobDetailID,
		FeederID:    input.FeederID,
		NumPole:     input.NumPole,
		DeviceCode:  input.DeviceCode,
		Detail:      input.Detail,
		URLsBefore:  input.URLsBefore,
		URLsAfter:   input.URLsAfter,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
	}
	entity, err := uc.repo.Update(ctx, repoInput)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, ErrTaskNotFound
	}
	return entity, nil
}

package usecase

import (
	"context"
	"time"

	taskdomain "backend-hotlines3/internal/domain/task"
	"backend-hotlines3/internal/port/outbound/repository"
)

type CreateTaskInput struct {
	WorkDate    time.Time
	TeamID      int64
	JobTypeID   int64
	JobDetailID int64
	FeederID    *int64
	NumPole     *string
	DeviceCode  *string
	Detail      *string
	URLsBefore  []string
	URLsAfter   []string
	Latitude    *float64
	Longitude   *float64
}

type CreateTaskUseCase struct {
	repo repository.TaskRepository
}

func NewCreateTaskUseCase(repo repository.TaskRepository) *CreateTaskUseCase {
	return &CreateTaskUseCase{repo: repo}
}

func (uc *CreateTaskUseCase) Execute(ctx context.Context, input CreateTaskInput) (*taskdomain.Entity, error) {
	if input.WorkDate.IsZero() {
		return nil, ErrWorkDateRequired
	}
	if input.TeamID <= 0 {
		return nil, ErrTeamIDRequired
	}
	if input.JobTypeID <= 0 {
		return nil, ErrJobTypeIDRequired
	}
	if input.JobDetailID <= 0 {
		return nil, ErrJobDetailIDRequired
	}

	repoInput := repository.TaskCreateInput{
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
	return uc.repo.Create(ctx, repoInput)
}

package usecase

import (
	"context"
	"time"

	taskdomain "backend-hotlines3/internal/domain/task"
	"backend-hotlines3/internal/port/outbound/repository"
)

type ListTasksByTeamInput struct {
	Page      int
	Limit     int
	WorkDate  *time.Time
	TeamID    *int64
	JobTypeID *int64
	FeederID  *int64
}

type ListTasksByTeamOutput struct {
	Tasks []taskdomain.Entity
	Total int64
	Page  int
	Limit int
}

type ListTasksByTeamUseCase struct {
	repo repository.TaskRepository
}

func NewListTasksByTeamUseCase(repo repository.TaskRepository) *ListTasksByTeamUseCase {
	return &ListTasksByTeamUseCase{repo: repo}
}

func (uc *ListTasksByTeamUseCase) Execute(ctx context.Context, input ListTasksByTeamInput) (*ListTasksByTeamOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 || input.Limit > 500 {
		input.Limit = 100
	}

	query := repository.TaskByTeamQuery{
		Page:      input.Page,
		Limit:     input.Limit,
		WorkDate:  input.WorkDate,
		TeamID:    input.TeamID,
		JobTypeID: input.JobTypeID,
		FeederID:  input.FeederID,
	}

	tasks, total, err := uc.repo.ListByTeam(ctx, query)
	if err != nil {
		return nil, err
	}

	return &ListTasksByTeamOutput{
		Tasks: tasks,
		Total: total,
		Page:  input.Page,
		Limit: input.Limit,
	}, nil
}

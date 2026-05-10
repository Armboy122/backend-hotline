package service

import (
	"context"
	"errors"
	"time"

	"backend-hotlines3/internal/feature/task/entity"
	taskrepository "backend-hotlines3/internal/feature/task/repository"
)

var (
	ErrTaskNotFound        = errors.New("task not found")
	ErrInvalidTaskID       = errors.New("invalid task ID")
	ErrWorkDateRequired    = errors.New("workDate is required")
	ErrTeamIDRequired      = errors.New("teamId is required")
	ErrJobTypeIDRequired   = errors.New("jobTypeId is required")
	ErrJobDetailIDRequired = errors.New("jobDetailId is required")
	ErrYearMonthRequired   = errors.New("year and month are required")
	ErrForbidden           = errors.New("forbidden")
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

type ListTasksInput struct {
	Page   int
	Limit  int
	Filter taskrepository.ListFilter
}

type ListTasksOutput struct {
	Tasks []entity.Task
	Total int64
	Page  int
	Limit int
}

type ListTasksByTeamInput struct {
	Page      int
	Limit     int
	WorkDate  *time.Time
	TeamID    *int64
	JobTypeID *int64
	FeederID  *int64
}

type ListTasksByTeamOutput struct {
	Tasks []entity.Task
	Total int64
	Page  int
	Limit int
}

type ListTasksByFilterInput struct {
	Year   string
	Month  string
	TeamID *int64
}

type ListTasksByFilterOutput struct {
	Tasks []entity.Task
}

type ServiceInterface interface {
	List(ctx context.Context, actor entity.Actor, input ListTasksInput) (*ListTasksOutput, error)
	GetByID(ctx context.Context, actor entity.Actor, id int64) (*entity.Task, error)
	Create(ctx context.Context, actor entity.Actor, input CreateTaskInput) (*entity.Task, error)
	Update(ctx context.Context, actor entity.Actor, input UpdateTaskInput) (*entity.Task, error)
	Delete(ctx context.Context, actor entity.Actor, id int64) error
	ListByTeam(ctx context.Context, actor entity.Actor, input ListTasksByTeamInput) (*ListTasksByTeamOutput, error)
	ListByFilter(ctx context.Context, actor entity.Actor, input ListTasksByFilterInput) (*ListTasksByFilterOutput, error)
}

type service struct {
	repository taskrepository.Repository
}

func NewService(repository taskrepository.Repository) ServiceInterface {
	return &service{repository: repository}
}

func (s *service) List(ctx context.Context, actor entity.Actor, input ListTasksInput) (*ListTasksOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 || input.Limit > 100 {
		input.Limit = 50
	}
	input.Filter.TeamID = actor.ScopedTeamID(input.Filter.TeamID)

	tasks, total, err := s.repository.List(ctx, taskrepository.ListQuery{
		Page:   input.Page,
		Limit:  input.Limit,
		Filter: input.Filter,
	})
	if err != nil {
		return nil, err
	}

	return &ListTasksOutput{Tasks: tasks, Total: total, Page: input.Page, Limit: input.Limit}, nil
}

func (s *service) GetByID(ctx context.Context, actor entity.Actor, id int64) (*entity.Task, error) {
	if id <= 0 {
		return nil, ErrInvalidTaskID
	}
	task, err := s.repository.GetByID(ctx, taskrepository.GetQuery{ID: id})
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	if !actor.CanReadTeam(task.TeamID) {
		return nil, ErrForbidden
	}
	return task, nil
}

func (s *service) Create(ctx context.Context, actor entity.Actor, input CreateTaskInput) (*entity.Task, error) {
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
	if !actor.CanWriteTeam(input.TeamID) {
		return nil, ErrForbidden
	}

	return s.repository.Create(ctx, taskrepository.CreateInput(input))
}

func (s *service) Update(ctx context.Context, actor entity.Actor, input UpdateTaskInput) (*entity.Task, error) {
	if input.ID <= 0 {
		return nil, ErrInvalidTaskID
	}
	existing, err := s.repository.GetByID(ctx, taskrepository.GetQuery{ID: input.ID})
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrTaskNotFound
	}
	if !actor.CanWriteTeam(existing.TeamID) {
		return nil, ErrForbidden
	}
	if input.TeamID != nil && !actor.CanWriteTeam(*input.TeamID) {
		return nil, ErrForbidden
	}
	task, err := s.repository.Update(ctx, taskrepository.UpdateInput(input))
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (s *service) Delete(ctx context.Context, actor entity.Actor, id int64) error {
	if id <= 0 {
		return ErrInvalidTaskID
	}
	existing, err := s.repository.GetByID(ctx, taskrepository.GetQuery{ID: id})
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrTaskNotFound
	}
	if !actor.CanWriteTeam(existing.TeamID) {
		return ErrForbidden
	}
	return s.repository.SoftDelete(ctx, taskrepository.DeleteCommand{ID: id})
}

func (s *service) ListByTeam(ctx context.Context, actor entity.Actor, input ListTasksByTeamInput) (*ListTasksByTeamOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 || input.Limit > 500 {
		input.Limit = 100
	}
	input.TeamID = actor.ScopedTeamID(input.TeamID)

	tasks, total, err := s.repository.ListByTeam(ctx, taskrepository.ByTeamQuery{
		Page:      input.Page,
		Limit:     input.Limit,
		WorkDate:  input.WorkDate,
		TeamID:    input.TeamID,
		JobTypeID: input.JobTypeID,
		FeederID:  input.FeederID,
	})
	if err != nil {
		return nil, err
	}

	return &ListTasksByTeamOutput{Tasks: tasks, Total: total, Page: input.Page, Limit: input.Limit}, nil
}

func (s *service) ListByFilter(ctx context.Context, actor entity.Actor, input ListTasksByFilterInput) (*ListTasksByFilterOutput, error) {
	if input.Year == "" || input.Month == "" {
		return nil, ErrYearMonthRequired
	}
	input.TeamID = actor.ScopedTeamID(input.TeamID)

	tasks, err := s.repository.ListByFilter(ctx, taskrepository.ByFilterQuery{
		Year:   input.Year,
		Month:  input.Month,
		TeamID: input.TeamID,
	})
	if err != nil {
		return nil, err
	}

	return &ListTasksByFilterOutput{Tasks: tasks}, nil
}

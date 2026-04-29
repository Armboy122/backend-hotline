package repository

import (
	"context"
	"time"

	taskdomain "backend-hotlines3/internal/domain/task"
)

// --- List (existing) ---

type TaskListFilter struct {
	WorkDate  *time.Time
	TeamID    *int64
	JobTypeID *int64
	FeederID  *int64
}

type TaskListQuery struct {
	Page   int
	Limit  int
	Filter TaskListFilter
}

// --- GetByID ---

type TaskGetQuery struct {
	ID int64
}

// --- Create ---

type TaskCreateInput struct {
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

// --- Update ---

type TaskUpdateInput struct {
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

// --- SoftDelete ---

type TaskDeleteCommand struct {
	ID int64
}

// --- ListByTeam ---

type TaskByTeamQuery struct {
	Page      int
	Limit     int
	WorkDate  *time.Time
	TeamID    *int64
	JobTypeID *int64
	FeederID  *int64
}

// --- ListByFilter ---

type TaskByFilterQuery struct {
	Year   string
	Month  string
	TeamID *int64
}

// TaskRepository defines all persistence operations for TaskDaily.
type TaskRepository interface {
	List(ctx context.Context, query TaskListQuery) ([]taskdomain.Entity, int64, error)
	GetByID(ctx context.Context, query TaskGetQuery) (*taskdomain.Entity, error)
	Create(ctx context.Context, input TaskCreateInput) (*taskdomain.Entity, error)
	Update(ctx context.Context, input TaskUpdateInput) (*taskdomain.Entity, error)
	SoftDelete(ctx context.Context, cmd TaskDeleteCommand) error
	ListByTeam(ctx context.Context, query TaskByTeamQuery) ([]taskdomain.Entity, int64, error)
	ListByFilter(ctx context.Context, query TaskByFilterQuery) ([]taskdomain.Entity, error)
}

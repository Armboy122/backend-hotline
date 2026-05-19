package repository

import (
	"context"
	"time"

	"backend-hotlines3/internal/feature/task/entity"

	"gorm.io/gorm"
)

type ListFilter struct {
	WorkDate  *time.Time
	TeamID    *int64
	JobTypeID *int64
	FeederID  *int64
}

type ListQuery struct {
	Page   int
	Limit  int
	Filter ListFilter
}

type GetQuery struct {
	ID int64
}

type CreateInput struct {
	WorkDate        time.Time
	TeamID          int64
	JobTypeID       int64
	JobDetailID     int64
	FeederID        *int64
	NumPole         *string
	DeviceCode      *string
	Detail          *string
	URLsBefore      []string
	URLsAfter       []string
	Latitude        *float64
	Longitude       *float64
	SourceType      *string
	SourceID        *int64
	LargeWorkTaskID *int64
}

type UpdateInput struct {
	ID              int64
	WorkDate        *time.Time
	TeamID          *int64
	JobTypeID       *int64
	JobDetailID     *int64
	FeederID        *int64
	NumPole         *string
	DeviceCode      *string
	Detail          *string
	URLsBefore      []string
	URLsAfter       []string
	Latitude        *float64
	Longitude       *float64
	SourceType      *string
	SourceID        *int64
	LargeWorkTaskID *int64
}

type DeleteCommand struct {
	ID int64
}

type ByTeamQuery struct {
	Page      int
	Limit     int
	WorkDate  *time.Time
	TeamID    *int64
	JobTypeID *int64
	FeederID  *int64
}

type ByFilterQuery struct {
	Year   string
	Month  string
	TeamID *int64
}

type Repository interface {
	List(ctx context.Context, query ListQuery) ([]entity.Task, int64, error)
	GetByID(ctx context.Context, query GetQuery) (*entity.Task, error)
	Create(ctx context.Context, input CreateInput) (*entity.Task, error)
	Update(ctx context.Context, input UpdateInput) (*entity.Task, error)
	SoftDelete(ctx context.Context, cmd DeleteCommand) error
	ListByTeam(ctx context.Context, query ByTeamQuery) ([]entity.Task, int64, error)
	ListByFilter(ctx context.Context, query ByFilterQuery) ([]entity.Task, error)
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

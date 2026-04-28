package repository

import (
	"context"
	"time"

	taskdomain "backend-hotlines3/internal/domain/task"
)

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

type TaskRepository interface {
	List(ctx context.Context, query TaskListQuery) ([]taskdomain.Entity, int64, error)
}

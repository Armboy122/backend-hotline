package repository

import (
	"context"

	taskentity "backend-hotlines3/internal/feature/task/entity"
	taskrepository "backend-hotlines3/internal/feature/task/repository"

	"gorm.io/gorm"
)

// Repository is the read-only TaskDaily port needed by work-report aggregation.
type Repository interface {
	ListByFilter(ctx context.Context, query taskrepository.ByFilterQuery) ([]taskentity.Task, error)
}

// NewRepository wires the work-report read model to the canonical TaskDaily repository.
func NewRepository(db *gorm.DB) Repository {
	return taskrepository.NewRepository(db)
}

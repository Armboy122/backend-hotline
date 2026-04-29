package task

import (
	"context"
	"time"

	taskdomain "backend-hotlines3/internal/domain/task"
	"backend-hotlines3/internal/port/outbound/repository"
)

// Repository is the module-local persistence interface alias.
type Repository = repository.TaskRepository

type ListFilter = repository.TaskListFilter
type ListQuery = repository.TaskListQuery
type GetQuery = repository.TaskGetQuery
type CreateInput = repository.TaskCreateInput
type UpdateInput = repository.TaskUpdateInput
type DeleteCommand = repository.TaskDeleteCommand
type ByTeamQuery = repository.TaskByTeamQuery
type ByFilterQuery = repository.TaskByFilterQuery


// RepositoryService is a local alias for the repository contract that keeps
// module code self-contained during the migration. It is a transitional
// placeholder and should disappear once Phase 1 moves behavior into the module.
type RepositoryService interface {
	List(ctx context.Context, query ListQuery) ([]taskdomain.Entity, int64, error)
	GetByID(ctx context.Context, query GetQuery) (*taskdomain.Entity, error)
	Create(ctx context.Context, input CreateInput) (*taskdomain.Entity, error)
	Update(ctx context.Context, input UpdateInput) (*taskdomain.Entity, error)
	SoftDelete(ctx context.Context, cmd DeleteCommand) error
	ListByTeam(ctx context.Context, query ByTeamQuery) ([]taskdomain.Entity, int64, error)
	ListByFilter(ctx context.Context, query ByFilterQuery) ([]taskdomain.Entity, error)
}

// Normalized pagination defaults used by the module-local service.
func NormalizePagination(page, limit int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func NormalizeTeamPagination(page, limit int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func StartOfDay(value time.Time) time.Time {
	return value.UTC().Truncate(24 * time.Hour)
}

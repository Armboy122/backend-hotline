package repository

import (
	"context"

	"backend-hotlines3/internal/feature/monthlyschedule/entity"
)

type Repository interface {
	FindOrCreatePeriod(ctx context.Context, year, month int) (entity.Period, error)
	FindPeriod(ctx context.Context, year, month int) (entity.Period, error)
	ListVisibleTeams(ctx context.Context) ([]entity.Team, error)
	GetSchedule(ctx context.Context, monthlyPlanID int64, status string) (*entity.Schedule, error)
	ReplaceDraft(ctx context.Context, period entity.Period, actor entity.Actor, expectedRevisionNo *int, assignments []entity.Assignment) (*entity.Schedule, error)
	PublishDraft(ctx context.Context, period entity.Period, actor entity.Actor, draftRevisionID int64, checksum string, projection []byte) (*entity.Schedule, error)
}

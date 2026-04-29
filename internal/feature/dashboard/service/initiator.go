package service

import (
	"context"

	"backend-hotlines3/internal/feature/dashboard/entity"
	"backend-hotlines3/internal/feature/dashboard/repository"
)

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Summary(ctx context.Context, f entity.SummaryFilter) (entity.SummaryResult, error) {
	return s.repo.Summary(ctx, f)
}

func (s *Service) TopJobs(ctx context.Context, f entity.TopFilter) ([]entity.TopJobResult, error) {
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 10
	}
	return s.repo.TopJobs(ctx, f)
}

func (s *Service) TopFeeders(ctx context.Context, f entity.TopFilter) ([]entity.TopFeederResult, error) {
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 10
	}
	return s.repo.TopFeeders(ctx, f)
}

func (s *Service) FeederMatrix(ctx context.Context, f entity.MatrixFilter) (entity.FeederMatrixResult, error) {
	return s.repo.FeederMatrix(ctx, f)
}

func (s *Service) Stats(ctx context.Context, f entity.StatsFilter) (entity.StatsResult, error) {
	return s.repo.Stats(ctx, f)
}

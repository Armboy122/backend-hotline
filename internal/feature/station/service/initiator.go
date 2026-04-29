package service

import (
	"context"

	"backend-hotlines3/internal/feature/station/entity"
	"backend-hotlines3/internal/feature/station/repository"
)

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]entity.Entity, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*entity.Entity, error) {
	if id <= 0 {
		return nil, entity.ErrInvalidID
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, input entity.CreateInput) (*entity.Entity, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, input entity.UpdateInput) (*entity.Entity, error) {
	if input.ID <= 0 {
		return nil, entity.ErrInvalidID
	}
	return s.repo.Update(ctx, input)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return entity.ErrInvalidID
	}
	return s.repo.Delete(ctx, id)
}

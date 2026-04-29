package service

import (
	"context"

	"backend-hotlines3/internal/feature/pea/entity"
	"backend-hotlines3/internal/feature/pea/repository"
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

func (s *Service) Create(ctx context.Context, shortname, fullname string, operationID int64) (*entity.Entity, error) {
	return s.repo.Create(ctx, shortname, fullname, operationID)
}

func (s *Service) BulkCreate(ctx context.Context, items []entity.Entity) ([]entity.Entity, error) {
	if len(items) == 0 {
		return nil, entity.ErrInvalidID
	}
	return s.repo.BulkCreate(ctx, items)
}

func (s *Service) Update(ctx context.Context, id int64, shortname, fullname string, operationID int64) (*entity.Entity, error) {
	if id <= 0 {
		return nil, entity.ErrInvalidID
	}
	return s.repo.Update(ctx, id, shortname, fullname, operationID)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return entity.ErrInvalidID
	}
	return s.repo.Delete(ctx, id)
}

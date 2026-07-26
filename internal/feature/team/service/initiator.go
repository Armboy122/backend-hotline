package service

import (
	"context"
	"strings"

	"backend-hotlines3/internal/feature/team/entity"
	"backend-hotlines3/internal/feature/team/repository"
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

func (s *Service) Create(ctx context.Context, input entity.UpsertInput) (*entity.Entity, error) {
	normalized, err := normalizeInput(input)
	if err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, normalized)
}

func (s *Service) Update(ctx context.Context, id int64, input entity.UpsertInput) (*entity.Entity, error) {
	if id <= 0 {
		return nil, entity.ErrInvalidID
	}
	normalized, err := normalizeInput(input)
	if err != nil {
		return nil, err
	}
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current.Code != nil && normalized.Code != nil && *current.Code != *normalized.Code {
		return nil, entity.ErrCodeImmutable
	}
	return s.repo.Update(ctx, id, normalized)
}

func normalizeInput(input entity.UpsertInput) (entity.UpsertInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return entity.UpsertInput{}, entity.ErrInvalidInput
	}
	input.Code = normalizedPointer(input.Code, true)
	input.BaseArea = normalizedPointer(input.BaseArea, false)
	input.CrewType = normalizedPointer(input.CrewType, false)
	if input.DisplayOrder != nil && *input.DisplayOrder < 0 {
		return entity.UpsertInput{}, entity.ErrInvalidInput
	}
	return input, nil
}

func normalizedPointer(value *string, uppercase bool) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if uppercase {
		normalized = strings.ToUpper(normalized)
	}
	if normalized == "" {
		return nil
	}
	return &normalized
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return entity.ErrInvalidID
	}
	return s.repo.Delete(ctx, id)
}

package service

import (
	"context"

	"backend-hotlines3/internal/feature/user/dto"
	"backend-hotlines3/internal/feature/user/entity"
	"backend-hotlines3/internal/feature/user/repository"
	"backend-hotlines3/pkg/password"
)

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, page, limit int) (entity.ListResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	items, total, err := s.repo.List(ctx, page, limit)
	if err != nil {
		return entity.ListResult{}, err
	}
	return entity.ListResult{Items: items, Total: total, Page: page, Limit: limit}, nil
}

func (s *Service) GetByID(ctx context.Context, id uint) (entity.UserInfo, error) {
	if id == 0 {
		return entity.UserInfo{}, entity.ErrInvalidID
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, req dto.CreateUserRequest) (entity.UserInfo, error) {
	hashed, err := password.HashPassword(req.Password)
	if err != nil {
		return entity.UserInfo{}, err
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return s.repo.Create(ctx, entity.CreateInput{
		Username:       req.Username,
		HashedPassword: hashed,
		Role:           req.Role,
		TeamID:         req.TeamID,
		IsActive:       isActive,
	})
}

func (s *Service) Update(ctx context.Context, id uint, req dto.UpdateUserRequest) (entity.UserInfo, error) {
	if id == 0 {
		return entity.UserInfo{}, entity.ErrInvalidID
	}
	return s.repo.Update(ctx, id, entity.UpdateInput{
		Username: req.Username,
		Role:     req.Role,
		TeamID:   req.TeamID,
		IsActive: req.IsActive,
	})
}

func (s *Service) Delete(ctx context.Context, id, callerID uint) error {
	if id == 0 {
		return entity.ErrInvalidID
	}
	if id == callerID {
		return entity.ErrCannotDeleteSelf
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) ChangePassword(ctx context.Context, id uint, req dto.ChangePasswordRequest) error {
	if id == 0 {
		return entity.ErrInvalidID
	}
	hash, err := s.repo.GetPasswordHash(ctx, id)
	if err != nil {
		return err
	}
	if !password.CheckPassword(req.OldPassword, hash) {
		return entity.ErrInvalidPassword
	}
	newHash, err := password.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, id, newHash)
}

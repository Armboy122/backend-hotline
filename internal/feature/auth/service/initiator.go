package service

import (
	"context"
	"errors"
	"time"

	"backend-hotlines3/internal/feature/auth/entity"
	"backend-hotlines3/internal/feature/auth/repository"
	"backend-hotlines3/pkg/jwt"
	"backend-hotlines3/pkg/password"
)

type Service struct {
	repo       repository.Repository
	jwtManager *jwt.JWTManager
}

func NewService(repo repository.Repository, jwtManager *jwt.JWTManager) *Service {
	return &Service{repo: repo, jwtManager: jwtManager}
}

func (s *Service) Register(ctx context.Context, req entity.RegisterInput) (entity.UserInfo, error) {
	existing, err := s.repo.FindByUsername(ctx, req.Username)
	if err == nil && existing != nil {
		return entity.UserInfo{}, entity.ErrUserExists
	}
	if req.Role == "user" && req.TeamID == nil {
		return entity.UserInfo{}, entity.ErrTeamRequired
	}

	hashed, err := password.HashPassword(req.Password)
	if err != nil {
		return entity.UserInfo{}, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	return s.repo.Create(ctx, entity.CreateUserInput{
		Username:       req.Username,
		HashedPassword: hashed,
		Role:           req.Role,
		TeamID:         req.TeamID,
		IsActive:       isActive,
	})
}

func (s *Service) Login(ctx context.Context, username, pwd string) (entity.LoginResult, error) {
	authUser, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidCredentials) {
			return entity.LoginResult{}, entity.ErrInvalidCredentials
		}
		return entity.LoginResult{}, err
	}

	if !password.CheckPassword(pwd, authUser.PasswordHash) {
		return entity.LoginResult{}, entity.ErrInvalidCredentials
	}

	if !authUser.IsActive {
		return entity.LoginResult{}, entity.ErrAccountDisabled
	}

	s.repo.UpdateLastLogin(ctx, authUser.ID, time.Now())

	access, refresh, err := s.jwtManager.GenerateTokenPair(authUser.ID, authUser.Username, authUser.Role, authUser.TeamID)
	if err != nil {
		return entity.LoginResult{}, err
	}

	return entity.LoginResult{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         authUser.ToUserInfo(),
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (entity.TokenPair, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return entity.TokenPair{}, entity.ErrInvalidToken
	}

	authUser, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return entity.TokenPair{}, entity.ErrUserNotFound
		}
		return entity.TokenPair{}, err
	}

	if !authUser.IsActive {
		return entity.TokenPair{}, entity.ErrAccountDisabled
	}

	access, refresh, err := s.jwtManager.GenerateTokenPair(authUser.ID, authUser.Username, authUser.Role, authUser.TeamID)
	if err != nil {
		return entity.TokenPair{}, err
	}
	return entity.TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *Service) Me(ctx context.Context, userID uint) (entity.UserInfo, error) {
	return s.repo.FindByIDWithTeam(ctx, userID)
}

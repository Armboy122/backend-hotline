package repository

import (
	"context"
	"errors"
	"time"

	"backend-hotlines3/internal/feature/auth/entity"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	FindByUsername(ctx context.Context, username string) (*entity.AuthUser, error)
	FindByID(ctx context.Context, id uint) (*entity.AuthUser, error)
	FindByIDWithTeam(ctx context.Context, id uint) (entity.UserInfo, error)
	Create(ctx context.Context, in entity.CreateUserInput) (entity.UserInfo, error)
	UpdateLastLogin(ctx context.Context, id uint, t time.Time) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func toAuthUser(m models.User) *entity.AuthUser {
	lastLogin := ""
	if m.LastLogin != nil {
		lastLogin = m.LastLogin.Format(time.RFC3339)
	}
	return &entity.AuthUser{
		ID:                 m.ID,
		Username:           m.Username,
		Role:               m.Role,
		TeamID:             m.TeamID,
		Capabilities:       activeCapabilityCodes(m.Capabilities),
		IsActive:           m.IsActive,
		MustChangePassword: m.MustChangePassword,
		LastLogin:          &lastLogin,
		CreatedAt:          m.CreatedAt.Format(time.RFC3339),
		PasswordHash:       m.Password,
	}
}

func (r *repository) FindByUsername(ctx context.Context, username string) (*entity.AuthUser, error) {
	var m models.User
	if err := r.db.WithContext(ctx).Preload("Capabilities", "revoked_at IS NULL").Where(`"username" = ?`, username).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrInvalidCredentials
		}
		return nil, err
	}
	return toAuthUser(m), nil
}

func (r *repository) FindByID(ctx context.Context, id uint) (*entity.AuthUser, error) {
	var m models.User
	if err := r.db.WithContext(ctx).Preload("Capabilities", "revoked_at IS NULL").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}
	return toAuthUser(m), nil
}

func (r *repository) FindByIDWithTeam(ctx context.Context, id uint) (entity.UserInfo, error) {
	var m models.User
	if err := r.db.WithContext(ctx).Preload("Capabilities", "revoked_at IS NULL").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.UserInfo{}, entity.ErrUserNotFound
		}
		return entity.UserInfo{}, err
	}
	return toAuthUser(m).ToUserInfo(), nil
}

func activeCapabilityCodes(items []models.UserCapability) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item.RevokedAt == nil {
			out = append(out, item.Code)
		}
	}
	return out
}

func (r *repository) Create(ctx context.Context, in entity.CreateUserInput) (entity.UserInfo, error) {
	m := models.User{
		Username:           in.Username,
		Password:           in.HashedPassword,
		Role:               in.Role,
		TeamID:             in.TeamID,
		IsActive:           in.IsActive,
		MustChangePassword: in.MustChangePassword,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return entity.UserInfo{}, err
	}
	return toAuthUser(m).ToUserInfo(), nil
}

func (r *repository) UpdateLastLogin(ctx context.Context, id uint, t time.Time) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update(`"lastLogin"`, t).Error
}

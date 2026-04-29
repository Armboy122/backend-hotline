package repository

import (
	"context"
	"errors"
	"time"

	"backend-hotlines3/internal/feature/user/entity"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	List(ctx context.Context, page, limit int) ([]entity.UserInfo, int64, error)
	GetByID(ctx context.Context, id uint) (entity.UserInfo, error)
	Create(ctx context.Context, in entity.CreateInput) (entity.UserInfo, error)
	Update(ctx context.Context, id uint, in entity.UpdateInput) (entity.UserInfo, error)
	Delete(ctx context.Context, id uint) error
	GetPasswordHash(ctx context.Context, id uint) (string, error)
	UpdatePassword(ctx context.Context, id uint, hashed string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func toInfo(m models.User) entity.UserInfo {
	lastLogin := ""
	if m.LastLogin != nil {
		lastLogin = m.LastLogin.Format(time.RFC3339)
	}
	return entity.UserInfo{
		ID:        m.ID,
		Username:  m.Username,
		Role:      m.Role,
		TeamID:    m.TeamID,
		IsActive:  m.IsActive,
		LastLogin: &lastLogin,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
	}
}

func (r *repository) List(ctx context.Context, page, limit int) ([]entity.UserInfo, int64, error) {
	base := r.db.WithContext(ctx).Scopes(models.UserNotDeleted)
	var total int64
	base.Model(&models.User{}).Count(&total)
	var items []models.User
	if err := base.Preload("Team").Order(`"createdAt" DESC`).Offset((page - 1) * limit).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	out := make([]entity.UserInfo, 0, len(items))
	for _, m := range items {
		out = append(out, toInfo(m))
	}
	return out, total, nil
}

func (r *repository) GetByID(ctx context.Context, id uint) (entity.UserInfo, error) {
	var m models.User
	if err := r.db.WithContext(ctx).Scopes(models.UserNotDeleted).Preload("Team").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.UserInfo{}, entity.ErrNotFound
		}
		return entity.UserInfo{}, err
	}
	return toInfo(m), nil
}

func (r *repository) Create(ctx context.Context, in entity.CreateInput) (entity.UserInfo, error) {
	m := models.User{
		Username: in.Username,
		Password: in.HashedPassword,
		Role:     in.Role,
		TeamID:   in.TeamID,
		IsActive: in.IsActive,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return entity.UserInfo{}, err
	}
	if err := r.db.WithContext(ctx).Preload("Team").First(&m, m.ID).Error; err != nil {
		return entity.UserInfo{}, err
	}
	return toInfo(m), nil
}

func (r *repository) Update(ctx context.Context, id uint, in entity.UpdateInput) (entity.UserInfo, error) {
	var m models.User
	if err := r.db.WithContext(ctx).Scopes(models.UserNotDeleted).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.UserInfo{}, entity.ErrNotFound
		}
		return entity.UserInfo{}, err
	}
	if in.Username != nil {
		m.Username = *in.Username
	}
	if in.Role != nil {
		m.Role = *in.Role
	}
	if in.TeamID != nil {
		m.TeamID = in.TeamID
	}
	if in.IsActive != nil {
		m.IsActive = *in.IsActive
	}
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return entity.UserInfo{}, err
	}
	if err := r.db.WithContext(ctx).Preload("Team").First(&m, m.ID).Error; err != nil {
		return entity.UserInfo{}, err
	}
	return toInfo(m), nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	var m models.User
	if err := r.db.WithContext(ctx).Scopes(models.UserNotDeleted).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.ErrNotFound
		}
		return err
	}
	now := time.Now()
	m.DeletedAt = &now
	return r.db.WithContext(ctx).Save(&m).Error
}

func (r *repository) GetPasswordHash(ctx context.Context, id uint) (string, error) {
	var m models.User
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", entity.ErrNotFound
		}
		return "", err
	}
	return m.Password, nil
}

func (r *repository) UpdatePassword(ctx context.Context, id uint, hashed string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("password", hashed).Error
}

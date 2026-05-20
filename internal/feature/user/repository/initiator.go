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
	UpdatePassword(ctx context.Context, id uint, hashed string, mustChangePassword bool) error
	MustChangePassword(ctx context.Context, id uint) (bool, error)
	CountActiveSuperAdmins(ctx context.Context, excludeID *uint) (int64, error)
	ListContacts(ctx context.Context, q entity.ContactListQuery) ([]entity.ContactInfo, int64, error)
	GetContactByID(ctx context.Context, id uint) (entity.ContactInfo, error)
	UpdateContact(ctx context.Context, id uint, in entity.UpdateContactInput) (entity.ContactInfo, error)
	CreateExternalContact(ctx context.Context, input entity.ExternalContactInput) (entity.ContactInfo, error)
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
	info := entity.UserInfo{
		ID:                 m.ID,
		Username:           m.Username,
		Role:               m.Role,
		TeamID:             m.TeamID,
		DisplayName:        m.DisplayName,
		Position:           m.Position,
		PhoneNumber:        m.PhoneNumber,
		IsActive:           m.IsActive,
		MustChangePassword: m.MustChangePassword,
		LastLogin:          &lastLogin,
		CreatedAt:          m.CreatedAt.Format(time.RFC3339),
	}
	if m.Team != nil {
		info.Team = &entity.TeamInfo{ID: m.Team.ID, Name: m.Team.Name}
	}
	return info
}

func toContactInfo(m models.User) entity.ContactInfo {
	info := entity.ContactInfo{
		ID:          m.ID,
		Source:      entity.ContactSourceUser,
		Type:        entity.ContactTypeUser,
		Username:    m.Username,
		DisplayName: m.DisplayName,
		Position:    m.Position,
		PhoneNumber: m.PhoneNumber,
		Role:        m.Role,
		TeamID:      m.TeamID,
		IsActive:    m.IsActive,
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339),
	}
	if m.Team != nil {
		info.Team = &entity.TeamInfo{ID: m.Team.ID, Name: m.Team.Name}
	}
	return info
}

func toExternalContactInfo(m models.ExternalContact) entity.ContactInfo {
	info := entity.ContactInfo{
		ExternalID:   m.ID,
		Source:       entity.ContactSourceExternal,
		Type:         m.Type,
		DisplayName:  &m.DisplayName,
		Position:     m.Position,
		PhoneNumber:  &m.PhoneNumber,
		Organization: m.Organization,
		Notes:        m.Notes,
		TeamID:       m.TeamID,
		IsActive:     m.IsActive,
		UpdatedAt:    m.UpdatedAt.Format(time.RFC3339),
	}
	if m.Team != nil {
		info.Team = &entity.TeamInfo{ID: m.Team.ID, Name: m.Team.Name}
	}
	return info
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

func (r *repository) UpdatePassword(ctx context.Context, id uint, hashed string, mustChangePassword bool) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{"password": hashed, "must_change_password": mustChangePassword}).Error
}

func (r *repository) MustChangePassword(ctx context.Context, id uint) (bool, error) {
	var m models.User
	if err := r.db.WithContext(ctx).Select("must_change_password").Scopes(models.UserNotDeleted).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, entity.ErrNotFound
		}
		return false, err
	}
	return m.MustChangePassword, nil
}

func (r *repository) CountActiveSuperAdmins(ctx context.Context, excludeID *uint) (int64, error) {
	query := r.db.WithContext(ctx).Model(&models.User{}).
		Scopes(models.UserNotDeleted).
		Where("role = ? AND \"isActive\" = ?", "super_admin", true)
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *repository) ListContacts(ctx context.Context, q entity.ContactListQuery) ([]entity.ContactInfo, int64, error) {
	base := r.db.WithContext(ctx).Model(&models.User{}).Scopes(models.UserNotDeleted)
	includeUsers := q.Type == "" || q.Type == entity.ContactTypeUser
	includeExternal := q.Type == "" || q.Type == entity.ContactTypeExternal || q.Type == entity.ContactTypeEmergency || q.Type == entity.ContactTypeOperationCenter
	if !includeUsers {
		base = base.Where("1 = 0")
	}
	if !q.IncludeInactive {
		base = base.Where(`"isActive" = ?`, true)
	}
	if q.TeamID != nil {
		base = base.Where(`"teamId" = ?`, *q.TeamID)
	}
	if q.Role != "" {
		base = base.Where("role = ?", q.Role)
	}
	if q.Query != "" {
		like := "%" + q.Query + "%"
		base = base.Where(`username ILIKE ? OR "displayName" ILIKE ? OR position ILIKE ? OR "phoneNumber" ILIKE ?`, like, like, like, like)
	}
	var userTotal int64
	if err := base.Count(&userTotal).Error; err != nil {
		return nil, 0, err
	}
	var users []models.User
	if includeUsers {
		if err := base.Preload("Team").Order(`"isActive" DESC, "displayName" ASC NULLS LAST, username ASC`).Find(&users).Error; err != nil {
			return nil, 0, err
		}
	}

	extBase := r.db.WithContext(ctx).Model(&models.ExternalContact{}).Where("deleted_at IS NULL")
	if !includeExternal {
		extBase = extBase.Where("1 = 0")
	}
	if q.Type != "" && q.Type != entity.ContactTypeUser {
		extBase = extBase.Where("type = ?", q.Type)
	}
	if !q.IncludeInactive {
		extBase = extBase.Where("is_active = ?", true)
	}
	if q.TeamID != nil {
		extBase = extBase.Where("(team_id = ? OR team_id IS NULL)", *q.TeamID)
	}
	if q.Query != "" {
		like := "%" + q.Query + "%"
		extBase = extBase.Where("display_name ILIKE ? OR phone_number ILIKE ? OR organization ILIKE ? OR position ILIKE ? OR notes ILIKE ?", like, like, like, like, like)
	}
	var extTotal int64
	if err := extBase.Count(&extTotal).Error; err != nil {
		return nil, 0, err
	}
	var external []models.ExternalContact
	if includeExternal {
		if err := extBase.Preload("Team").Order("is_active DESC, display_name ASC").Find(&external).Error; err != nil {
			return nil, 0, err
		}
	}

	out := make([]entity.ContactInfo, 0, len(users)+len(external))
	for _, item := range users {
		out = append(out, toContactInfo(item))
	}
	for _, item := range external {
		out = append(out, toExternalContactInfo(item))
	}
	start := (q.Page - 1) * q.Limit
	if start > len(out) {
		return []entity.ContactInfo{}, userTotal + extTotal, nil
	}
	end := start + q.Limit
	if end > len(out) {
		end = len(out)
	}
	return out[start:end], userTotal + extTotal, nil
}

func (r *repository) GetContactByID(ctx context.Context, id uint) (entity.ContactInfo, error) {
	var m models.User
	if err := r.db.WithContext(ctx).Scopes(models.UserNotDeleted).Preload("Team").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.ContactInfo{}, entity.ErrNotFound
		}
		return entity.ContactInfo{}, err
	}
	return toContactInfo(m), nil
}

func (r *repository) UpdateContact(ctx context.Context, id uint, in entity.UpdateContactInput) (entity.ContactInfo, error) {
	updates := map[string]interface{}{}
	if in.DisplayName != nil {
		updates["displayName"] = *in.DisplayName
	}
	if in.Position != nil {
		updates["position"] = *in.Position
	}
	if in.PhoneNumber != nil {
		updates["phoneNumber"] = *in.PhoneNumber
	}
	if len(updates) > 0 {
		updates["updatedAt"] = time.Now()
	}
	result := r.db.WithContext(ctx).Model(&models.User{}).Scopes(models.UserNotDeleted).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return entity.ContactInfo{}, result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ContactInfo{}, entity.ErrNotFound
	}
	return r.GetContactByID(ctx, id)
}

func (r *repository) CreateExternalContact(ctx context.Context, input entity.ExternalContactInput) (entity.ContactInfo, error) {
	active := true
	if input.IsActive != nil {
		active = *input.IsActive
	}
	m := models.ExternalContact{
		Type:         input.Type,
		DisplayName:  input.DisplayName,
		PhoneNumber:  input.PhoneNumber,
		Organization: input.Organization,
		Position:     input.Position,
		Notes:        input.Notes,
		TeamID:       input.TeamID,
		IsActive:     active,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return entity.ContactInfo{}, err
	}
	if err := r.db.WithContext(ctx).Preload("Team").First(&m, "id = ?", m.ID).Error; err != nil {
		return entity.ContactInfo{}, err
	}
	return toExternalContactInfo(m), nil
}

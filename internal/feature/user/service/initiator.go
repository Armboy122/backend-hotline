package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/feature/user/dto"
	"backend-hotlines3/internal/feature/user/entity"
	"backend-hotlines3/internal/feature/user/repository"
	"backend-hotlines3/pkg/password"
)

const adminCreatedDefaultPassword = "123456"

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListContacts(ctx context.Context, actor entity.Actor, q entity.ContactListQuery) (entity.ContactListResult, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 50
	}
	q.Query = strings.TrimSpace(q.Query)
	q.Type = strings.TrimSpace(q.Type)
	if actor.Role != policy.RoleSuperAdmin {
		q.IncludeInactive = false
	}
	items, total, err := s.repo.ListContacts(ctx, q)
	if err != nil {
		return entity.ContactListResult{}, err
	}
	return entity.ContactListResult{Items: items, Total: total, Page: q.Page, Limit: q.Limit}, nil
}

func (s *Service) CreateExternalContact(ctx context.Context, actor entity.Actor, input entity.ExternalContactInput) (entity.ContactInfo, error) {
	if !canManageExternalContact(actor, input.TeamID) {
		return entity.ContactInfo{}, entity.ErrForbidden
	}
	input.Type = strings.TrimSpace(input.Type)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.PhoneNumber = strings.TrimSpace(input.PhoneNumber)
	if !isAllowedContactType(input.Type) || input.DisplayName == "" || input.PhoneNumber == "" {
		return entity.ContactInfo{}, entity.ErrInvalidContactField
	}
	return s.repo.CreateExternalContact(ctx, input)
}

func canManageExternalContact(actor entity.Actor, teamID *int64) bool {
	if actor.Role == policy.RoleSuperAdmin {
		return true
	}
	if actor.Role != policy.RoleTeamLead && actor.Role != policy.RoleUser {
		return false
	}
	return teamID != nil && *teamID > 0
}

func isAllowedContactType(t string) bool {
	switch strings.TrimSpace(t) {
	case entity.ContactTypeExternal, entity.ContactTypeEmergency, entity.ContactTypeOperationCenter:
		return true
	default:
		return false
	}
}

func (s *Service) GetContactByID(ctx context.Context, id uint) (entity.ContactInfo, error) {
	if id == 0 {
		return entity.ContactInfo{}, entity.ErrInvalidID
	}
	return s.repo.GetContactByID(ctx, id)
}

func (s *Service) UpdateContact(ctx context.Context, actor entity.Actor, id uint, req dto.UpdateContactRequest) (entity.ContactInfo, error) {
	if id == 0 {
		return entity.ContactInfo{}, entity.ErrInvalidID
	}
	if actor.ID != id && actor.Role != policy.RoleSuperAdmin {
		return entity.ContactInfo{}, entity.ErrForbidden
	}
	if req.DisplayName == nil && req.Position == nil && req.PhoneNumber == nil {
		return entity.ContactInfo{}, entity.ErrNoContactFields
	}
	in := entity.UpdateContactInput{
		DisplayName: normalizeContactField(req.DisplayName),
		Position:    normalizeContactField(req.Position),
		PhoneNumber: normalizeContactField(req.PhoneNumber),
	}
	if exceedsMax(in.DisplayName, 120) || exceedsMax(in.Position, 120) || exceedsMax(in.PhoneNumber, 40) {
		return entity.ContactInfo{}, entity.ErrInvalidContactField
	}
	return s.repo.UpdateContact(ctx, id, in)
}

func normalizeContactField(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	return &trimmed
}

func exceedsMax(v *string, max int) bool {
	return v != nil && utf8.RuneCountInString(*v) > max
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

func (s *Service) Create(ctx context.Context, actor entity.Actor, req dto.CreateUserRequest) (entity.UserInfo, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if err := s.authorizeCreate(ctx, actor, req.Role, isActive); err != nil {
		return entity.UserInfo{}, err
	}
	hashed, err := password.HashPassword(adminCreatedDefaultPassword)
	if err != nil {
		return entity.UserInfo{}, err
	}
	return s.repo.Create(ctx, entity.CreateInput{
		Username:           req.Username,
		HashedPassword:     hashed,
		Role:               req.Role,
		TeamID:             req.TeamID,
		IsActive:           isActive,
		MustChangePassword: true,
	})
}

func (s *Service) Update(ctx context.Context, actor entity.Actor, id uint, req dto.UpdateUserRequest) (entity.UserInfo, error) {
	if id == 0 {
		return entity.UserInfo{}, entity.ErrInvalidID
	}
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return entity.UserInfo{}, err
	}
	if err := s.authorizeUpdate(ctx, actor, current, req); err != nil {
		return entity.UserInfo{}, err
	}
	return s.repo.Update(ctx, id, entity.UpdateInput{
		Username: req.Username,
		Role:     req.Role,
		TeamID:   req.TeamID,
		IsActive: req.IsActive,
	})
}

func (s *Service) Delete(ctx context.Context, actor entity.Actor, id uint) error {
	if id == 0 {
		return entity.ErrInvalidID
	}
	if id == actor.ID {
		return entity.ErrCannotDeleteSelf
	}
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !policy.CanManageRole(actor.Role, current.Role) {
		return entity.ErrForbidden
	}
	if current.Role == policy.RoleSuperAdmin && current.IsActive {
		count, err := s.repo.CountActiveSuperAdmins(ctx, &id)
		if err != nil {
			return err
		}
		if count == 0 {
			return entity.ErrCannotRemoveOnlySuperAdmin
		}
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
	return s.repo.UpdatePassword(ctx, id, newHash, false)
}

func (s *Service) ResetPassword(ctx context.Context, actor entity.Actor, id uint, _ dto.ResetPasswordRequest) error {
	if id == 0 {
		return entity.ErrInvalidID
	}
	if !policy.CanResetPasswords(actor.Role) {
		return entity.ErrForbidden
	}
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}
	newHash, err := password.HashPassword(adminCreatedDefaultPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, id, newHash, true)
}

func (s *Service) authorizeCreate(ctx context.Context, actor entity.Actor, targetRole string, targetActive bool) error {
	if !policy.IsValidRole(targetRole) {
		return entity.ErrInvalidRole
	}
	if !policy.CanCreateRole(actor.Role, targetRole) {
		return entity.ErrForbidden
	}
	if targetRole == policy.RoleSuperAdmin && targetActive {
		count, err := s.repo.CountActiveSuperAdmins(ctx, nil)
		if err != nil {
			return err
		}
		if count > 0 {
			return entity.ErrSuperAdminAlreadyExists
		}
	}
	return nil
}

func (s *Service) authorizeUpdate(ctx context.Context, actor entity.Actor, current entity.UserInfo, req dto.UpdateUserRequest) error {
	if !policy.CanManageRole(actor.Role, current.Role) {
		return entity.ErrForbidden
	}
	targetRole := current.Role
	if req.Role != nil {
		if !policy.IsValidRole(*req.Role) {
			return entity.ErrInvalidRole
		}
		targetRole = *req.Role
		if !policy.CanManageRole(actor.Role, targetRole) {
			return entity.ErrForbidden
		}
	}
	targetActive := current.IsActive
	if req.IsActive != nil {
		targetActive = *req.IsActive
	}

	if current.Role == policy.RoleSuperAdmin && current.IsActive && (targetRole != policy.RoleSuperAdmin || !targetActive) {
		count, err := s.repo.CountActiveSuperAdmins(ctx, &current.ID)
		if err != nil {
			return err
		}
		if count == 0 {
			return entity.ErrCannotRemoveOnlySuperAdmin
		}
	}
	if targetRole == policy.RoleSuperAdmin && targetActive && (current.Role != policy.RoleSuperAdmin || !current.IsActive) {
		count, err := s.repo.CountActiveSuperAdmins(ctx, &current.ID)
		if err != nil {
			return err
		}
		if count > 0 {
			return entity.ErrSuperAdminAlreadyExists
		}
	}
	return nil
}

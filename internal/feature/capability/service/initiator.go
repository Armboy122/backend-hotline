package service

import (
	"context"
	"errors"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/feature/capability/entity"
	"backend-hotlines3/internal/feature/capability/repository"
)

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListAvailable() []entity.Capability {
	return entity.AvailableCapabilities()
}

func (s *Service) ListForUser(ctx context.Context, userID uint) (entity.UserCapabilities, error) {
	if userID == 0 {
		return entity.UserCapabilities{}, entity.ErrInvalidUserID
	}
	codes, err := s.repo.ListCodesByUserID(ctx, userID)
	if err != nil {
		return entity.UserCapabilities{}, err
	}
	return entity.UserCapabilities{UserID: userID, Capabilities: codes}, nil
}

func (s *Service) ReplaceForUser(ctx context.Context, actorID uint, userID uint, codes []string) (entity.UserCapabilities, error) {
	if userID == 0 {
		return entity.UserCapabilities{}, entity.ErrInvalidUserID
	}
	deduped := make([]string, 0, len(codes))
	seen := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		if !entity.IsValidCapability(code) {
			return entity.UserCapabilities{}, entity.ErrInvalidCapability
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		deduped = append(deduped, code)
	}
	target, err := s.repo.GetTargetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrTargetUserNotFound) {
			return entity.UserCapabilities{}, entity.ErrTargetUserNotFound
		}
		return entity.UserCapabilities{}, err
	}
	if target == nil {
		return entity.UserCapabilities{}, entity.ErrTargetUserNotFound
	}
	if !target.IsActive {
		return entity.UserCapabilities{}, entity.ErrInactiveTargetUser
	}
	if target.Role == policy.RoleViewer {
		return entity.UserCapabilities{}, entity.ErrViewerTargetForbidden
	}
	grantedBy := actorID
	if err := s.repo.ReplaceCodes(ctx, userID, deduped, &grantedBy); err != nil {
		return entity.UserCapabilities{}, err
	}
	return entity.UserCapabilities{UserID: userID, Capabilities: deduped}, nil
}

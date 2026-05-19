package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/feature/teamplan/entity"
	"backend-hotlines3/internal/feature/teamplan/repository"
)

var (
	ErrInvalidTeamPlanID = entity.ErrInvalidID
)

type ListInput struct {
	From   time.Time
	To     time.Time
	TeamID *int64
	Status []string
	Page   int
	Limit  int
}

type ListOutput struct {
	Items []entity.TeamPlan
	Total int64
	Page  int
	Limit int
}

type CreateInput struct {
	TeamID            int64
	Title             string
	WorkType          *string
	StartDate         *time.Time
	EndDate           *time.Time
	WorkTime          *string
	LocationText      string
	PEAID             *int64
	OperationCenterID *int64
	FeederID          *int64
	StationID         *int64
	Notes             *string
}

type UpdateInput struct {
	ID                int64
	TeamID            *int64
	Title             *string
	WorkType          *string
	StartDate         *time.Time
	EndDate           *time.Time
	WorkTime          *string
	LocationText      *string
	PEAID             *int64
	OperationCenterID *int64
	FeederID          *int64
	StationID         *int64
	Notes             *string
	Status            *string
}

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, actor entity.Actor, input ListInput) (*ListOutput, error) {
	hasFrom := !input.From.IsZero()
	hasTo := !input.To.IsZero()
	if hasFrom != hasTo {
		return nil, entity.ErrInvalidRange
	}
	var from time.Time
	var to time.Time
	if hasFrom {
		from = dateOnly(input.From)
		to = dateOnly(input.To)
		if to.Before(from) {
			return nil, entity.ErrInvalidRange
		}
	}
	teamID, err := scopedListTeamID(actor, input.TeamID)
	if err != nil {
		return nil, err
	}
	page := normalizePage(input.Page)
	limit := normalizeLimit(input.Limit)
	items, total, err := s.repo.List(ctx, repository.ListQuery{
		From:   from,
		To:     to,
		TeamID: teamID,
		Status: input.Status,
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	return &ListOutput{Items: items, Total: total, Page: page, Limit: limit}, nil
}

func (s *Service) GetByID(ctx context.Context, actor entity.Actor, id int64) (*entity.TeamPlan, error) {
	if id <= 0 {
		return nil, entity.ErrInvalidID
	}
	item, err := s.repo.GetByID(ctx, repository.GetQuery{ID: id})
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, entity.ErrNotFound
	}
	if !canViewTeamPlan(actor, *item) {
		return nil, entity.ErrForbidden
	}
	return item, nil
}

func (s *Service) Create(ctx context.Context, actor entity.Actor, input CreateInput) (*entity.TeamPlan, error) {
	if input.TeamID <= 0 || strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.LocationText) == "" {
		return nil, entity.ErrInvalidInput
	}
	if input.EndDate != nil && input.StartDate == nil {
		return nil, entity.ErrInvalidRange
	}
	if input.StartDate != nil && input.EndDate != nil && dateOnly(*input.EndDate).Before(dateOnly(*input.StartDate)) {
		return nil, entity.ErrInvalidRange
	}
	if !actor.CanCreateTeamPlan(input.TeamID) {
		return nil, entity.ErrForbidden
	}
	created, err := s.repo.Create(ctx, repository.CreateInput{
		TeamID:            input.TeamID,
		CreatedByUserID:   actor.UserID,
		Title:             strings.TrimSpace(input.Title),
		WorkType:          input.WorkType,
		StartDate:         dateOnlyPtr(input.StartDate),
		EndDate:           input.EndDate,
		WorkTime:          input.WorkTime,
		LocationText:      strings.TrimSpace(input.LocationText),
		PEAID:             input.PEAID,
		OperationCenterID: input.OperationCenterID,
		FeederID:          input.FeederID,
		StationID:         input.StationID,
		Notes:             input.Notes,
		Status:            initialStatusForStartDate(input.StartDate),
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) Update(ctx context.Context, actor entity.Actor, input UpdateInput) (*entity.TeamPlan, error) {
	if input.ID <= 0 {
		return nil, entity.ErrInvalidID
	}
	existing, err := s.repo.GetByID(ctx, repository.GetQuery{ID: input.ID})
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, entity.ErrNotFound
	}
	if !actor.CanUpdateTeamPlan(*existing) {
		return nil, entity.ErrForbidden
	}
	if input.TeamID != nil && *input.TeamID != existing.TeamID && actor.Role != "super_admin" {
		return nil, entity.ErrForbidden
	}
	if input.Status != nil {
		status := strings.TrimSpace(*input.Status)
		if !entity.IsValidStatus(status) {
			return nil, entity.ErrInvalidStatus
		}
	}
	mergedStartDate := existing.StartDate
	if input.StartDate != nil {
		mergedStartDate = dateOnlyPtr(input.StartDate)
	}
	mergedEndDate := existing.EndDate
	if input.EndDate != nil {
		mergedEndDate = input.EndDate
	}
	if mergedEndDate != nil && mergedStartDate == nil {
		return nil, entity.ErrInvalidRange
	}
	if mergedStartDate != nil && mergedEndDate != nil && dateOnly(*mergedEndDate).Before(dateOnly(*mergedStartDate)) {
		return nil, entity.ErrInvalidRange
	}
	if input.Status != nil && strings.TrimSpace(*input.Status) == entity.StatusPlanned && mergedStartDate == nil {
		return nil, entity.ErrInvalidRange
	}
	updated, err := s.repo.Update(ctx, repository.UpdateInput{
		ID:                input.ID,
		TeamID:            input.TeamID,
		Title:             input.Title,
		WorkType:          input.WorkType,
		StartDate:         input.StartDate,
		EndDate:           input.EndDate,
		WorkTime:          input.WorkTime,
		LocationText:      input.LocationText,
		PEAID:             input.PEAID,
		OperationCenterID: input.OperationCenterID,
		FeederID:          input.FeederID,
		StationID:         input.StationID,
		Notes:             input.Notes,
		Status:            input.Status,
	})
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, entity.ErrNotFound
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, actor entity.Actor, id int64) error {
	if id <= 0 {
		return entity.ErrInvalidID
	}
	existing, err := s.repo.GetByID(ctx, repository.GetQuery{ID: id})
	if err != nil {
		return err
	}
	if existing == nil {
		return entity.ErrNotFound
	}
	if !actor.CanDeleteTeamPlan(*existing) {
		return entity.ErrForbidden
	}
	return s.repo.SoftDelete(ctx, repository.DeleteCommand{ID: id})
}

func canViewTeamPlan(actor entity.Actor, plan entity.TeamPlan) bool {
	if actor.Role == policy.RoleSuperAdmin {
		return true
	}
	if actor.TeamID == nil || *actor.TeamID != plan.TeamID {
		return false
	}
	switch actor.Role {
	case policy.RoleTeamLead, policy.RoleUser, policy.RoleViewer:
		return true
	default:
		return false
	}
}

func initialStatusForStartDate(start *time.Time) string {
	if start == nil || start.IsZero() {
		return entity.StatusDraft
	}
	return entity.StatusPlanned
}

func dateOnlyPtr(t *time.Time) *time.Time {
	if t == nil || t.IsZero() {
		return nil
	}
	out := dateOnly(*t)
	return &out
}

func scopedListTeamID(actor entity.Actor, requested *int64) (*int64, error) {
	if actor.Role == policy.RoleSuperAdmin {
		return requested, nil
	}
	if actor.TeamID == nil || *actor.TeamID <= 0 {
		return nil, entity.ErrForbidden
	}
	switch actor.Role {
	case policy.RoleTeamLead, policy.RoleUser, policy.RoleViewer:
		if requested != nil && *requested != *actor.TeamID {
			return nil, entity.ErrForbidden
		}
		teamID := *actor.TeamID
		return &teamID, nil
	default:
		return nil, entity.ErrForbidden
	}
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizeLimit(limit int) int {
	if limit < 1 || limit > 100 {
		return 50
	}
	return limit
}

func dateOnly(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func (s *Service) MustNotUse() error {
	return errors.New("unused")
}

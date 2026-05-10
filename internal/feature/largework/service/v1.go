package service

import (
	"context"
	"errors"
	"time"

	"backend-hotlines3/internal/feature/largework/entity"
	repo "backend-hotlines3/internal/feature/largework/repository"
)

var (
	ErrForbidden               = errors.New("forbidden")
	ErrNotFound                = errors.New("large work item not found")
	ErrInvalidID               = errors.New("invalid large work item ID")
	ErrInvalidParticipantTeams = errors.New("participant team list must contain at least two distinct teams including owner")
	ErrInvalidStartDate        = errors.New("startDate is required")
	ErrInvalidTitle            = errors.New("title is required")
	ErrInvalidLocation         = errors.New("locationText is required")
	ErrInvalidOwnerTeam        = errors.New("ownerTeamId is required")
	ErrInvalidStatus           = errors.New("invalid status")
	ErrInvalidDateRange        = errors.New("endDate must be on or after startDate")
	ErrInvalidStateTransition  = errors.New("large work item state does not allow this operation")
)

type CreateInput struct {
	OwnerTeamID        int64
	ParticipantTeamIDs []int64
	Title              string
	WorkType           *string
	StartDate          time.Time
	EndDate            *time.Time
	WorkTime           *string
	LocationText       string
	PEAID              *int64
	OperationCenterID  *int64
	FeederID           *int64
	StationID          *int64
	Notes              *string
	Status             string
}

type UpdateInput struct {
	ID                 int64
	OwnerTeamID        *int64
	ParticipantTeamIDs []int64
	Title              *string
	WorkType           *string
	StartDate          *time.Time
	EndDate            *time.Time
	WorkTime           *string
	LocationText       *string
	PEAID              *int64
	OperationCenterID  *int64
	FeederID           *int64
	StationID          *int64
	Notes              *string
	Status             *string
}

type ListInput struct {
	Page     int
	Limit    int
	From     *time.Time
	To       *time.Time
	TeamID   *int64
	Statuses []string
}

type ListOutput struct {
	Items []entity.LargeWorkItem
	Page  int
	Limit int
	Total int64
}

type Service interface {
	Create(context.Context, entity.Actor, CreateInput) (*entity.LargeWorkItem, error)
	GetByID(context.Context, entity.Actor, int64) (*entity.LargeWorkItem, error)
	List(context.Context, entity.Actor, ListInput) (*ListOutput, error)
	Update(context.Context, entity.Actor, UpdateInput) (*entity.LargeWorkItem, error)
	Cancel(context.Context, entity.Actor, int64) (*entity.LargeWorkItem, error)
}

type service struct {
	repo repo.Repository
}

func NewService(repository repo.Repository) Service {
	return &service{repo: repository}
}

func (s *service) Create(ctx context.Context, actor entity.Actor, input CreateInput) (*entity.LargeWorkItem, error) {
	if input.OwnerTeamID <= 0 {
		return nil, ErrInvalidOwnerTeam
	}
	if input.Title == "" {
		return nil, ErrInvalidTitle
	}
	if input.LocationText == "" {
		return nil, ErrInvalidLocation
	}
	if input.StartDate.IsZero() {
		return nil, ErrInvalidStartDate
	}
	if input.EndDate != nil && input.EndDate.Before(input.StartDate) {
		return nil, ErrInvalidDateRange
	}
	participantIDs := normalizeTeamIDs(append(input.ParticipantTeamIDs, input.OwnerTeamID))
	if len(participantIDs) < 2 {
		return nil, ErrInvalidParticipantTeams
	}
	if !actor.CanCreateLargeWork(&input.OwnerTeamID) {
		return nil, ErrForbidden
	}
	status := input.Status
	if status == "" {
		status = entity.LargeWorkStatusPlanned
	}
	if !isAllowedStatus(status) {
		return nil, ErrInvalidStatus
	}
	return s.repo.Create(ctx, repo.CreateInput{
		OwnerTeamID:        input.OwnerTeamID,
		CreatedByUserID:    actor.UserID,
		ParticipantTeamIDs: participantIDs,
		Title:              input.Title,
		WorkType:           input.WorkType,
		StartDate:          input.StartDate,
		EndDate:            input.EndDate,
		WorkTime:           input.WorkTime,
		LocationText:       input.LocationText,
		PEAID:              input.PEAID,
		OperationCenterID:  input.OperationCenterID,
		FeederID:           input.FeederID,
		StationID:          input.StationID,
		Notes:              input.Notes,
		Status:             status,
	})
}

func (s *service) GetByID(ctx context.Context, actor entity.Actor, id int64) (*entity.LargeWorkItem, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	item, err := s.repo.GetByID(ctx, repo.GetQuery{ID: id})
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	if !actor.CanViewLargeWork(item) {
		return nil, ErrForbidden
	}
	return item, nil
}

func (s *service) List(ctx context.Context, actor entity.Actor, input ListInput) (*ListOutput, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 || limit > 100 {
		limit = 100
	}
	input.TeamID = actor.ScopedTeamID(input.TeamID)
	items, total, err := s.repo.List(ctx, repo.ListQuery{Page: page, Limit: limit, From: input.From, To: input.To, TeamID: input.TeamID, Statuses: input.Statuses})
	if err != nil {
		return nil, err
	}
	visible := make([]entity.LargeWorkItem, 0, len(items))
	for _, item := range items {
		if actor.CanViewLargeWork(&item) {
			visible = append(visible, item)
		}
	}
	return &ListOutput{Items: visible, Page: page, Limit: limit, Total: total}, nil
}

func (s *service) Update(ctx context.Context, actor entity.Actor, input UpdateInput) (*entity.LargeWorkItem, error) {
	if input.ID <= 0 {
		return nil, ErrInvalidID
	}
	existing, err := s.repo.GetByID(ctx, repo.GetQuery{ID: input.ID})
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if !actor.IsPrivileged() {
		return nil, ErrForbidden
	}
	if !isEditableStatus(existing.Status) {
		return nil, ErrInvalidStateTransition
	}
	if input.OwnerTeamID != nil && !actor.CanCreateLargeWork(input.OwnerTeamID) {
		return nil, ErrForbidden
	}
	if input.Status != nil && !isAllowedStatus(*input.Status) {
		return nil, ErrInvalidStatus
	}
	if input.ParticipantTeamIDs != nil {
		participantIDs := normalizeTeamIDs(append(input.ParticipantTeamIDs, func() int64 {
			if input.OwnerTeamID != nil {
				return *input.OwnerTeamID
			}
			return existing.OwnerTeamID
		}()))
		if len(participantIDs) < 2 {
			return nil, ErrInvalidParticipantTeams
		}
		input.ParticipantTeamIDs = participantIDs
	}
	if input.Title != nil && *input.Title == "" {
		return nil, ErrInvalidTitle
	}
	if input.LocationText != nil && *input.LocationText == "" {
		return nil, ErrInvalidLocation
	}
	startDate := existing.StartDate
	if input.StartDate != nil {
		startDate = *input.StartDate
	}
	endDate := existing.EndDate
	if input.EndDate != nil {
		endDate = input.EndDate
	}
	if endDate != nil && endDate.Before(startDate) {
		return nil, ErrInvalidDateRange
	}
	updated, err := s.repo.Update(ctx, repo.UpdateInput{
		ID:                 input.ID,
		OwnerTeamID:        input.OwnerTeamID,
		ParticipantTeamIDs: input.ParticipantTeamIDs,
		Title:              input.Title,
		WorkType:           input.WorkType,
		StartDate:          input.StartDate,
		EndDate:            input.EndDate,
		WorkTime:           input.WorkTime,
		LocationText:       input.LocationText,
		PEAID:              input.PEAID,
		OperationCenterID:  input.OperationCenterID,
		FeederID:           input.FeederID,
		StationID:          input.StationID,
		Notes:              input.Notes,
		Status:             input.Status,
	})
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrNotFound
	}
	return updated, nil
}

func (s *service) Cancel(ctx context.Context, actor entity.Actor, id int64) (*entity.LargeWorkItem, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	item, err := s.repo.GetByID(ctx, repo.GetQuery{ID: id})
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	if !actor.CanViewLargeWork(item) {
		return nil, ErrForbidden
	}
	if !actor.IsPrivileged() {
		return nil, ErrForbidden
	}
	if !isEditableStatus(item.Status) {
		return nil, ErrInvalidStateTransition
	}
	status := entity.LargeWorkStatusCancelled
	return s.repo.Update(ctx, repo.UpdateInput{ID: id, Status: &status})
}

func isEditableStatus(status string) bool {
	switch status {
	case entity.LargeWorkStatusDraft, entity.LargeWorkStatusPlanned:
		return true
	default:
		return false
	}
}

func normalizeTeamIDs(ids []int64) []int64 {
	seen := map[int64]struct{}{}
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func isAllowedStatus(status string) bool {
	switch status {
	case entity.LargeWorkStatusDraft, entity.LargeWorkStatusPlanned, entity.LargeWorkStatusInProgress, entity.LargeWorkStatusCompleted, entity.LargeWorkStatusCancelled:
		return true
	default:
		return false
	}
}

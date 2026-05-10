package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/feature/largework/entity"
	repo "backend-hotlines3/internal/feature/largework/repository"
)

type fakeRepository struct {
	capturedList   repo.ListQuery
	capturedGet    repo.GetQuery
	capturedCreate repo.CreateInput
	capturedUpdate repo.UpdateInput

	listItems []entity.LargeWorkItem
	listTotal int64
	listErr   error
	listFunc  func(repo.ListQuery) ([]entity.LargeWorkItem, int64, error)

	getItem *entity.LargeWorkItem
	getErr  error

	createItem *entity.LargeWorkItem
	createErr  error

	updateItem *entity.LargeWorkItem
	updateErr  error
}

func (f *fakeRepository) List(_ context.Context, q repo.ListQuery) ([]entity.LargeWorkItem, int64, error) {
	f.capturedList = q
	if f.listFunc != nil {
		return f.listFunc(q)
	}
	return f.listItems, f.listTotal, f.listErr
}
func (f *fakeRepository) GetByID(_ context.Context, q repo.GetQuery) (*entity.LargeWorkItem, error) {
	f.capturedGet = q
	return f.getItem, f.getErr
}
func (f *fakeRepository) Create(_ context.Context, input repo.CreateInput) (*entity.LargeWorkItem, error) {
	f.capturedCreate = input
	return f.createItem, f.createErr
}
func (f *fakeRepository) Update(_ context.Context, input repo.UpdateInput) (*entity.LargeWorkItem, error) {
	f.capturedUpdate = input
	return f.updateItem, f.updateErr
}

func actor(role string, teamID *int64) entity.Actor {
	return entity.Actor{UserID: 1, Role: role, TeamID: teamID}
}

func TestCreateAutoAddsOwnerToParticipantsAndDedupes(t *testing.T) {
	ownerTeamID := int64(7)
	participantA := int64(8)
	participantDup := int64(7)
	repo := &fakeRepository{createItem: &entity.LargeWorkItem{ID: 1, OwnerTeamID: 7}}
	svc := NewService(repo)

	out, err := svc.Create(context.Background(), actor(policy.RoleAdmin, &ownerTeamID), CreateInput{
		OwnerTeamID:        7,
		ParticipantTeamIDs: []int64{participantA, participantDup},
		Title:              "งานระดมทีม เปลี่ยนอุปกรณ์หลัก",
		StartDate:          mustDate(t, "2026-06-10"),
		LocationText:       "Station A",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil || out.ID != 1 {
		t.Fatalf("unexpected create result: %#v", out)
	}
	if len(repo.capturedCreate.ParticipantTeamIDs) != 2 {
		t.Fatalf("participant ids = %#v, want owner auto-added and deduped", repo.capturedCreate.ParticipantTeamIDs)
	}
	if repo.capturedCreate.ParticipantTeamIDs[0] != 8 || repo.capturedCreate.ParticipantTeamIDs[1] != 7 {
		t.Fatalf("participant ids order/value = %#v, want [8 7] or equivalent dedupe with owner", repo.capturedCreate.ParticipantTeamIDs)
	}
}

func TestCreateRejectsInsufficientTeams(t *testing.T) {
	ownerTeamID := int64(7)
	svc := NewService(&fakeRepository{})
	_, err := svc.Create(context.Background(), actor(policy.RoleTeamLead, &ownerTeamID), CreateInput{
		OwnerTeamID:        7,
		ParticipantTeamIDs: []int64{7},
		Title:              "งานระดมทีม",
		StartDate:          mustDate(t, "2026-06-10"),
		LocationText:       "Station A",
	})
	if !errors.Is(err, ErrInvalidParticipantTeams) {
		t.Fatalf("got %v, want ErrInvalidParticipantTeams", err)
	}
}

func TestCreateRejectsTeamLeadForAnyOwnerTeamInMVP(t *testing.T) {
	ownerTeamID := int64(7)
	svc := NewService(&fakeRepository{})
	_, err := svc.Create(context.Background(), actor(policy.RoleTeamLead, &ownerTeamID), CreateInput{
		OwnerTeamID:        ownerTeamID,
		ParticipantTeamIDs: []int64{ownerTeamID, 9},
		Title:              "งานระดมทีม",
		StartDate:          mustDate(t, "2026-06-10"),
		LocationText:       "Station A",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("got %v, want ErrForbidden", err)
	}
}

func TestCreateRejectsEndDateBeforeStartDate(t *testing.T) {
	ownerTeamID := int64(7)
	endDate := mustDate(t, "2026-06-09")
	svc := NewService(&fakeRepository{})

	_, err := svc.Create(context.Background(), actor(policy.RoleAdmin, &ownerTeamID), CreateInput{
		OwnerTeamID:        ownerTeamID,
		ParticipantTeamIDs: []int64{ownerTeamID, 9},
		Title:              "งานระดมทีม",
		StartDate:          mustDate(t, "2026-06-10"),
		EndDate:            &endDate,
		LocationText:       "Station A",
	})

	if !errors.Is(err, ErrInvalidDateRange) {
		t.Fatalf("got %v, want ErrInvalidDateRange", err)
	}
}

func TestListScopesNonPrivilegedActorsToVisibleTeam(t *testing.T) {
	teamID := int64(7)
	repo := &fakeRepository{listItems: []entity.LargeWorkItem{{ID: 1, OwnerTeamID: 7}, {ID: 2, OwnerTeamID: 8}}, listTotal: 2}
	svc := NewService(repo)

	out, err := svc.List(context.Background(), actor(policy.RoleUser, &teamID), ListInput{Page: 0, Limit: 500})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Page != 1 || out.Limit != 100 {
		t.Fatalf("page/limit = %d/%d, want 1/100", out.Page, out.Limit)
	}
	if repo.capturedList.TeamID == nil || *repo.capturedList.TeamID != teamID {
		t.Fatalf("captured team filter = %#v, want %d", repo.capturedList.TeamID, teamID)
	}
}

func TestListDoesNotRepaginateRepositoryPageForPrivilegedActors(t *testing.T) {
	seed := []entity.LargeWorkItem{
		{ID: 1, OwnerTeamID: 7},
		{ID: 2, OwnerTeamID: 8},
		{ID: 3, OwnerTeamID: 9},
		{ID: 4, OwnerTeamID: 10},
		{ID: 5, OwnerTeamID: 11},
	}
	for _, role := range []string{policy.RoleAdmin, policy.RoleSuperAdmin} {
		t.Run(role, func(t *testing.T) {
			repo := &fakeRepository{listFunc: paginatedLargeWorkSeed(seed)}
			svc := NewService(repo)

			out, err := svc.List(context.Background(), actor(role, nil), ListInput{Page: 2, Limit: 2})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.Total != 5 {
				t.Fatalf("total = %d, want repository total 5", out.Total)
			}
			if len(out.Items) != 2 || out.Items[0].ID != 3 || out.Items[1].ID != 4 {
				t.Fatalf("items = %#v, want repository page items 3 and 4", out.Items)
			}
		})
	}
}

func TestListUsesRepositoryScopedPaginationForTeamActors(t *testing.T) {
	teamID := int64(7)
	seed := []entity.LargeWorkItem{
		{ID: 1, OwnerTeamID: 7},
		{ID: 2, OwnerTeamID: 8},
		{ID: 3, OwnerTeamID: 9, Teams: []entity.LargeWorkTeam{{ID: teamID}}},
		{ID: 4, OwnerTeamID: 10},
		{ID: 5, OwnerTeamID: 7},
		{ID: 6, OwnerTeamID: 11, Teams: []entity.LargeWorkTeam{{ID: teamID}}},
		{ID: 7, OwnerTeamID: 12},
	}
	repo := &fakeRepository{listFunc: paginatedLargeWorkSeed(seed)}
	svc := NewService(repo)

	out, err := svc.List(context.Background(), actor(policy.RoleUser, &teamID), ListInput{Page: 2, Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.capturedList.TeamID == nil || *repo.capturedList.TeamID != teamID {
		t.Fatalf("captured team filter = %#v, want %d", repo.capturedList.TeamID, teamID)
	}
	if out.Total != 4 {
		t.Fatalf("total = %d, want repository total 4", out.Total)
	}
	if len(out.Items) != 2 || out.Items[0].ID != 5 || out.Items[1].ID != 6 {
		t.Fatalf("items = %#v, want repository-scoped page items 5 and 6", out.Items)
	}
}

func TestUpdateAndCancelRestrictByActorRoleAndEditableState(t *testing.T) {
	ownerTeamID := int64(7)
	existing := &entity.LargeWorkItem{ID: 99, OwnerTeamID: 7, Status: entity.LargeWorkStatusPlanned}
	repo := &fakeRepository{getItem: existing, updateItem: existing}
	svc := NewService(repo)

	if _, err := svc.Update(context.Background(), actor(policy.RoleAdmin, &ownerTeamID), UpdateInput{ID: 99, Title: strPtr("updated")}); err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}
	if repo.capturedUpdate.ID != 99 {
		t.Fatalf("update id = %d, want 99", repo.capturedUpdate.ID)
	}
	if _, err := svc.Cancel(context.Background(), actor(policy.RoleAdmin, &ownerTeamID), 99); err != nil {
		t.Fatalf("unexpected cancel error: %v", err)
	}
	if _, err := svc.Update(context.Background(), actor(policy.RoleTeamLead, &ownerTeamID), UpdateInput{ID: 99, Title: strPtr("updated")}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("team lead update got %v, want ErrForbidden", err)
	}

	for _, status := range []string{entity.LargeWorkStatusInProgress, entity.LargeWorkStatusCompleted, entity.LargeWorkStatusCancelled} {
		t.Run("reject update and cancel for "+status, func(t *testing.T) {
			blocked := &entity.LargeWorkItem{ID: 99, OwnerTeamID: 7, Status: status}
			repo := &fakeRepository{getItem: blocked, updateItem: blocked}
			svc := NewService(repo)

			if _, err := svc.Update(context.Background(), actor(policy.RoleAdmin, &ownerTeamID), UpdateInput{ID: 99, Title: strPtr("updated")}); !errors.Is(err, ErrInvalidStateTransition) {
				t.Fatalf("update got %v, want ErrInvalidStateTransition", err)
			}
			if _, err := svc.Cancel(context.Background(), actor(policy.RoleAdmin, &ownerTeamID), 99); !errors.Is(err, ErrInvalidStateTransition) {
				t.Fatalf("cancel got %v, want ErrInvalidStateTransition", err)
			}
		})
	}
}

func TestUpdateRejectsEffectiveEndDateBeforeStartDate(t *testing.T) {
	ownerTeamID := int64(7)
	existingEnd := mustDate(t, "2026-06-12")
	existing := &entity.LargeWorkItem{ID: 99, OwnerTeamID: 7, Status: entity.LargeWorkStatusPlanned, StartDate: mustDate(t, "2026-06-10"), EndDate: &existingEnd}
	svc := NewService(&fakeRepository{getItem: existing, updateItem: existing})

	newEnd := mustDate(t, "2026-06-09")
	if _, err := svc.Update(context.Background(), actor(policy.RoleAdmin, &ownerTeamID), UpdateInput{ID: 99, EndDate: &newEnd}); !errors.Is(err, ErrInvalidDateRange) {
		t.Fatalf("end date update got %v, want ErrInvalidDateRange", err)
	}

	newStart := mustDate(t, "2026-06-13")
	if _, err := svc.Update(context.Background(), actor(policy.RoleAdmin, &ownerTeamID), UpdateInput{ID: 99, StartDate: &newStart}); !errors.Is(err, ErrInvalidDateRange) {
		t.Fatalf("start date update got %v, want ErrInvalidDateRange", err)
	}
}

func paginatedLargeWorkSeed(seed []entity.LargeWorkItem) func(repo.ListQuery) ([]entity.LargeWorkItem, int64, error) {
	return func(q repo.ListQuery) ([]entity.LargeWorkItem, int64, error) {
		filtered := make([]entity.LargeWorkItem, 0, len(seed))
		for _, item := range seed {
			if q.TeamID != nil && !largeWorkItemBelongsToTeam(item, *q.TeamID) {
				continue
			}
			filtered = append(filtered, item)
		}
		page := q.Page
		if page < 1 {
			page = 1
		}
		limit := q.Limit
		if limit < 1 {
			limit = len(filtered)
		}
		start := (page - 1) * limit
		if start >= len(filtered) {
			return nil, int64(len(filtered)), nil
		}
		end := start + limit
		if end > len(filtered) {
			end = len(filtered)
		}
		return filtered[start:end], int64(len(filtered)), nil
	}
}

func largeWorkItemBelongsToTeam(item entity.LargeWorkItem, teamID int64) bool {
	if item.OwnerTeamID == teamID {
		return true
	}
	for _, team := range item.Teams {
		if team.ID == teamID {
			return true
		}
	}
	return false
}

func mustDate(t *testing.T, raw string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		t.Fatalf("parse date %s: %v", raw, err)
	}
	return parsed
}

func strPtr(v string) *string { return &v }

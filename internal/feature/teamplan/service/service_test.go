package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/feature/teamplan/entity"
	"backend-hotlines3/internal/feature/teamplan/repository"
)

func TestServiceCreateEnforcesRoleAndTeamScope(t *testing.T) {
	ctx := context.Background()
	teamID := int64(7)
	otherTeamID := int64(8)
	lead := entity.Actor{UserID: 2, Role: policy.RoleTeamLead, TeamID: &teamID}
	user := entity.Actor{UserID: 3, Role: policy.RoleUser, TeamID: &teamID}
	admin := entity.Actor{UserID: 4, Role: policy.RoleAdmin, TeamID: &teamID}
	superAdmin := entity.Actor{UserID: 1, Role: policy.RoleSuperAdmin}

	repo := &fakeRepo{}
	svc := NewService(repo)
	startDate := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)

	input := CreateInput{
		TeamID:       teamID,
		Title:        "Patrol feeder A",
		StartDate:    &startDate,
		LocationText: "Bang Khen",
	}
	if _, err := svc.Create(ctx, lead, input); err != nil {
		t.Fatalf("team lead create: %v", err)
	}
	if repo.createInput == nil || repo.createInput.TeamID != teamID || repo.createInput.CreatedByUserID != lead.UserID {
		t.Fatalf("expected create input captured for own team, got %+v", repo.createInput)
	}

	repo = &fakeRepo{}
	svc = NewService(repo)
	unscheduledInput := input
	unscheduledInput.StartDate = nil
	if created, err := svc.Create(ctx, lead, unscheduledInput); err != nil {
		t.Fatalf("team lead create unscheduled draft: %v", err)
	} else if created.Status != entity.StatusDraft {
		t.Fatalf("unscheduled plan status = %q, want draft", created.Status)
	}
	if repo.createInput == nil || repo.createInput.StartDate != nil || repo.createInput.Status != entity.StatusDraft {
		t.Fatalf("unscheduled create input = %+v, want nil startDate and draft status", repo.createInput)
	}

	repo = &fakeRepo{}
	svc = NewService(repo)
	if _, err := svc.Create(ctx, user, input); err != nil {
		t.Fatalf("user create: %v", err)
	}
	if repo.createInput == nil || repo.createInput.TeamID != teamID || repo.createInput.CreatedByUserID != user.UserID {
		t.Fatalf("expected user create input captured, got %+v", repo.createInput)
	}

	repo = &fakeRepo{}
	svc = NewService(repo)
	input.TeamID = otherTeamID
	if _, err := svc.Create(ctx, user, input); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("user cross-team create expected forbidden, got %v", err)
	}
	if repo.createCalled {
		t.Fatalf("repository should not be called on forbidden create")
	}

	repo = &fakeRepo{}
	svc = NewService(repo)
	if _, err := svc.Create(ctx, admin, input); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("admin cross-team create expected forbidden, got %v", err)
	}
	if repo.createCalled {
		t.Fatalf("repository should not be called on forbidden create")
	}

	repo = &fakeRepo{}
	svc = NewService(repo)
	input.TeamID = teamID
	if _, err := svc.Create(ctx, admin, input); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("legacy admin own-team create expected forbidden, got %v", err)
	}
	if repo.createCalled {
		t.Fatalf("repository should not be called for legacy admin create")
	}

	viewer := entity.Actor{UserID: 5, Role: policy.RoleViewer, TeamID: &teamID}

	repo = &fakeRepo{}
	svc = NewService(repo)
	if _, err := svc.Create(ctx, viewer, input); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("viewer create expected forbidden, got %v", err)
	}
	if _, err := svc.Create(ctx, superAdmin, input); err != nil {
		t.Fatalf("super admin create: %v", err)
	}
}

func TestServiceListScopesTeamActorsToTheirOwnTeam(t *testing.T) {
	ctx := context.Background()
	teamID := int64(7)
	otherTeamID := int64(8)
	actor := entity.Actor{UserID: 3, Role: policy.RoleUser, TeamID: &teamID}
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	repo := &fakeRepo{}
	svc := NewService(repo)
	if _, err := svc.List(ctx, actor, ListInput{From: from, To: to}); err != nil {
		t.Fatalf("list own team scoped implicitly: %v", err)
	}
	if repo.listInput.TeamID == nil || *repo.listInput.TeamID != teamID {
		t.Fatalf("list should constrain user actor to own team, got query %+v", repo.listInput)
	}

	repo = &fakeRepo{}
	svc = NewService(repo)
	if _, err := svc.List(ctx, actor, ListInput{From: from, To: to, TeamID: &otherTeamID}); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("cross-team list expected forbidden, got %v", err)
	}
	if repo.listCalled {
		t.Fatalf("repository should not be called on forbidden cross-team list")
	}
}

func TestServiceListRejectsLegacyAdmin(t *testing.T) {
	ctx := context.Background()
	teamID := int64(7)
	admin := entity.Actor{UserID: 4, Role: policy.RoleAdmin, TeamID: &teamID}
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	repo := &fakeRepo{}
	svc := NewService(repo)
	if _, err := svc.List(ctx, admin, ListInput{From: from, To: to}); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("legacy admin list expected forbidden, got %v", err)
	}
	if repo.listCalled {
		t.Fatalf("repository should not be called for legacy admin list")
	}
}

func TestServiceListAllowsDraftBoardWithoutDateRange(t *testing.T) {
	ctx := context.Background()
	teamID := int64(7)
	actor := entity.Actor{UserID: 3, Role: policy.RoleUser, TeamID: &teamID}
	repo := &fakeRepo{}
	svc := NewService(repo)

	out, err := svc.List(ctx, actor, ListInput{Status: []string{entity.StatusDraft}})
	if err != nil {
		t.Fatalf("draft board list without date range: %v", err)
	}
	if out.Page != 1 || out.Limit != 50 {
		t.Fatalf("pagination defaults = page %d limit %d, want 1/50", out.Page, out.Limit)
	}
	if repo.listInput.TeamID == nil || *repo.listInput.TeamID != teamID {
		t.Fatalf("draft board should be scoped to own team, got query %+v", repo.listInput)
	}
	if !repo.listInput.From.IsZero() || !repo.listInput.To.IsZero() {
		t.Fatalf("draft board should not force date range, got query %+v", repo.listInput)
	}
}

func TestServiceGetByIDRejectsCrossTeamAccess(t *testing.T) {
	ctx := context.Background()
	teamID := int64(7)
	otherTeamID := int64(8)
	actor := entity.Actor{UserID: 3, Role: policy.RoleUser, TeamID: &teamID}
	otherTeamPlan := entity.TeamPlan{ID: 12, TeamID: otherTeamID, CreatedByUserID: 92, Status: entity.StatusPlanned}

	repo := &fakeRepo{getByIDResult: &otherTeamPlan}
	svc := NewService(repo)
	if _, err := svc.GetByID(ctx, actor, otherTeamPlan.ID); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("cross-team get expected forbidden, got %v", err)
	}
}

func TestServiceGetByIDRejectsAdminCrossTeamAccess(t *testing.T) {
	ctx := context.Background()
	teamID := int64(7)
	otherTeamID := int64(8)
	admin := entity.Actor{UserID: 4, Role: policy.RoleAdmin, TeamID: &teamID}
	otherTeamPlan := entity.TeamPlan{ID: 12, TeamID: otherTeamID, CreatedByUserID: 92, Status: entity.StatusPlanned}

	repo := &fakeRepo{getByIDResult: &otherTeamPlan}
	svc := NewService(repo)
	if _, err := svc.GetByID(ctx, admin, otherTeamPlan.ID); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("admin cross-team get expected forbidden, got %v", err)
	}
}

func TestServiceUpdateRejectsStartOnlyDateAfterExistingEndDate(t *testing.T) {
	ctx := context.Background()
	teamID := int64(7)
	creatorID := int64(91)
	user := entity.Actor{UserID: creatorID, Role: policy.RoleUser, TeamID: &teamID}
	endDate := time.Date(2026, 6, 5, 15, 30, 0, 0, time.Local)
	newStartDate := time.Date(2026, 6, 6, 8, 0, 0, 0, time.Local)
	plan := entity.TeamPlan{
		ID:              10,
		TeamID:          teamID,
		CreatedByUserID: creatorID,
		Status:          entity.StatusPlanned,
		StartDate:       ptrTime(time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)),
		EndDate:         &endDate,
	}
	repo := &fakeRepo{getByIDResult: &plan}
	svc := NewService(repo)

	_, err := svc.Update(ctx, user, UpdateInput{ID: plan.ID, StartDate: &newStartDate})

	if !errors.Is(err, entity.ErrInvalidRange) {
		t.Fatalf("start-only update error = %v, want invalid range", err)
	}
	if repo.updateCalled {
		t.Fatalf("repository should not be called on invalid merged date range")
	}
}

func TestServiceUpdateRejectsEndOnlyDateBeforeExistingStartDate(t *testing.T) {
	ctx := context.Background()
	teamID := int64(7)
	creatorID := int64(91)
	user := entity.Actor{UserID: creatorID, Role: policy.RoleUser, TeamID: &teamID}
	newEndDate := time.Date(2026, 6, 2, 23, 59, 0, 0, time.Local)
	plan := entity.TeamPlan{
		ID:              10,
		TeamID:          teamID,
		CreatedByUserID: creatorID,
		Status:          entity.StatusPlanned,
		StartDate:       ptrTime(time.Date(2026, 6, 3, 17, 0, 0, 0, time.Local)),
	}
	repo := &fakeRepo{getByIDResult: &plan}
	svc := NewService(repo)

	_, err := svc.Update(ctx, user, UpdateInput{ID: plan.ID, EndDate: &newEndDate})

	if !errors.Is(err, entity.ErrInvalidRange) {
		t.Fatalf("end-only update error = %v, want invalid range", err)
	}
	if repo.updateCalled {
		t.Fatalf("repository should not be called on invalid merged date range")
	}
}

func TestServiceUpdateAllowsAuthorizedDateAndWorkTimeEdit(t *testing.T) {
	ctx := context.Background()
	teamID := int64(7)
	creatorID := int64(91)
	actor := entity.Actor{UserID: 1, Role: policy.RoleSuperAdmin}
	existingStart := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)
	existingEnd := time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)
	plan := entity.TeamPlan{
		ID:              10,
		TeamID:          teamID,
		CreatedByUserID: creatorID,
		Title:           "Patrol feeder A",
		LocationText:    "Bang Khen",
		Status:          entity.StatusPlanned,
		StartDate:       &existingStart,
		EndDate:         &existingEnd,
	}
	repo := &fakeRepo{getByIDResult: &plan}
	svc := NewService(repo)
	newStart := time.Date(2026, 6, 7, 9, 30, 0, 0, time.Local)
	newEnd := time.Date(2026, 6, 9, 18, 15, 0, 0, time.Local)
	workTime := "09:00-16:00"

	updated, err := svc.Update(ctx, actor, UpdateInput{ID: plan.ID, StartDate: &newStart, EndDate: &newEnd, WorkTime: &workTime})

	if err != nil {
		t.Fatalf("super admin update dates/work time: %v", err)
	}
	if repo.updateInput == nil || repo.updateInput.StartDate == nil || repo.updateInput.EndDate == nil || repo.updateInput.WorkTime == nil {
		t.Fatalf("update input missing dates/workTime: %+v", repo.updateInput)
	}
	wantStart := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, 6, 9, 0, 0, 0, 0, time.UTC)
	if !repo.updateInput.StartDate.Equal(wantStart) || !repo.updateInput.EndDate.Equal(wantEnd) {
		t.Fatalf("update dates = %v/%v, want %v/%v", repo.updateInput.StartDate, repo.updateInput.EndDate, wantStart, wantEnd)
	}
	if *repo.updateInput.WorkTime != workTime || updated.WorkTime == nil || *updated.WorkTime != workTime {
		t.Fatalf("workTime not preserved through update: input=%+v updated=%+v", repo.updateInput.WorkTime, updated.WorkTime)
	}
}

func TestServiceUpdateOmittedDateFieldsPreserveExistingValues(t *testing.T) {
	ctx := context.Background()
	teamID := int64(7)
	creatorID := int64(91)
	actor := entity.Actor{UserID: creatorID, Role: policy.RoleUser, TeamID: &teamID}
	existingStart := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)
	existingEnd := time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC)
	existingWorkTime := "เช้า"
	plan := entity.TeamPlan{
		ID:              10,
		TeamID:          teamID,
		CreatedByUserID: creatorID,
		Title:           "Patrol feeder A",
		LocationText:    "Bang Khen",
		Status:          entity.StatusPlanned,
		StartDate:       &existingStart,
		EndDate:         &existingEnd,
		WorkTime:        &existingWorkTime,
	}
	repo := &fakeRepo{getByIDResult: &plan}
	svc := NewService(repo)
	title := "Patrol feeder A - updated"

	updated, err := svc.Update(ctx, actor, UpdateInput{ID: plan.ID, Title: &title})

	if err != nil {
		t.Fatalf("user update title only: %v", err)
	}
	if repo.updateInput == nil || repo.updateInput.StartDate != nil || repo.updateInput.EndDate != nil {
		t.Fatalf("omitted dates should be sent as nil update fields, got %+v", repo.updateInput)
	}
	if updated.StartDate == nil || updated.EndDate == nil || !updated.StartDate.Equal(existingStart) || !updated.EndDate.Equal(existingEnd) {
		t.Fatalf("omitted dates not preserved: got start=%v end=%v", updated.StartDate, updated.EndDate)
	}
	if updated.WorkTime == nil || *updated.WorkTime != existingWorkTime {
		t.Fatalf("omitted workTime not preserved: got %v", updated.WorkTime)
	}
}

func TestServiceUpdateAndDeleteEnforceOwnershipRules(t *testing.T) {
	ctx := context.Background()
	teamID := int64(7)
	otherTeamID := int64(8)
	creatorID := int64(91)
	otherCreatorID := int64(92)
	lead := entity.Actor{UserID: 2, Role: policy.RoleTeamLead, TeamID: &teamID}
	user := entity.Actor{UserID: creatorID, Role: policy.RoleUser, TeamID: &teamID}
	superAdmin := entity.Actor{UserID: 1, Role: policy.RoleSuperAdmin}

	plan := entity.TeamPlan{ID: 10, TeamID: teamID, CreatedByUserID: creatorID, Status: entity.StatusPlanned}
	completed := entity.TeamPlan{ID: 11, TeamID: teamID, CreatedByUserID: creatorID, Status: entity.StatusCompleted}
	otherTeamPlan := entity.TeamPlan{ID: 12, TeamID: otherTeamID, CreatedByUserID: otherCreatorID, Status: entity.StatusPlanned}

	repo := &fakeRepo{getByIDResult: &plan}
	svc := NewService(repo)
	newTitle := "Updated"
	if _, err := svc.Update(ctx, user, UpdateInput{ID: plan.ID, Title: &newTitle}); err != nil {
		t.Fatalf("user update own item: %v", err)
	}
	if repo.updateInput == nil || repo.updateInput.ID != plan.ID {
		t.Fatalf("expected update input captured, got %+v", repo.updateInput)
	}

	repo = &fakeRepo{getByIDResult: &plan}
	svc = NewService(repo)
	if _, err := svc.Update(ctx, user, UpdateInput{ID: plan.ID, Title: &newTitle, TeamID: &otherTeamID}); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("user update other team expected forbidden, got %v", err)
	}
	if repo.updateCalled {
		t.Fatalf("repository should not be called on forbidden update")
	}

	repo = &fakeRepo{getByIDResult: &completed}
	svc = NewService(repo)
	if _, err := svc.Update(ctx, user, UpdateInput{ID: completed.ID, Title: &newTitle}); err != nil {
		t.Fatalf("user update completed own-team item: %v", err)
	}

	repo = &fakeRepo{getByIDResult: &otherTeamPlan}
	svc = NewService(repo)
	if _, err := svc.Update(ctx, lead, UpdateInput{ID: otherTeamPlan.ID, Title: &newTitle}); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("team lead update other team expected forbidden, got %v", err)
	}

	repo = &fakeRepo{getByIDResult: &otherTeamPlan}
	svc = NewService(repo)
	if _, err := svc.Update(ctx, superAdmin, UpdateInput{ID: otherTeamPlan.ID, Title: &newTitle, TeamID: &otherTeamID}); err != nil {
		t.Fatalf("super admin update: %v", err)
	}

	viewer := entity.Actor{UserID: 5, Role: policy.RoleViewer, TeamID: &teamID}
	repo = &fakeRepo{getByIDResult: &plan}
	svc = NewService(repo)
	if _, err := svc.Update(ctx, viewer, UpdateInput{ID: plan.ID, Title: &newTitle}); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("viewer update expected forbidden, got %v", err)
	}
	if repo.updateCalled {
		t.Fatalf("repository should not be called on viewer update")
	}

	repo = &fakeRepo{getByIDResult: &plan}
	svc = NewService(repo)
	legacyAdmin := entity.Actor{UserID: 4, Role: policy.RoleAdmin, TeamID: &teamID}
	if _, err := svc.Update(ctx, legacyAdmin, UpdateInput{ID: plan.ID, Title: &newTitle}); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("legacy admin update expected forbidden, got %v", err)
	}
	if repo.updateCalled {
		t.Fatalf("repository should not be called on legacy admin update")
	}

	repo = &fakeRepo{getByIDResult: &plan}
	svc = NewService(repo)
	if err := svc.Delete(ctx, user, plan.ID); err != nil {
		t.Fatalf("user delete own-team item: %v", err)
	}
	if repo.deleteInput == nil || repo.deleteInput.ID != plan.ID {
		t.Fatalf("expected user delete input captured, got %+v", repo.deleteInput)
	}

	repo = &fakeRepo{getByIDResult: &plan}
	svc = NewService(repo)
	if err := svc.Delete(ctx, lead, plan.ID); err != nil {
		t.Fatalf("team lead delete: %v", err)
	}
	if repo.deleteInput == nil || repo.deleteInput.ID != plan.ID {
		t.Fatalf("expected delete input captured, got %+v", repo.deleteInput)
	}

	repo = &fakeRepo{getByIDResult: &otherTeamPlan}
	svc = NewService(repo)
	if err := svc.Delete(ctx, superAdmin, otherTeamPlan.ID); err != nil {
		t.Fatalf("super admin delete: %v", err)
	}
}

func TestServiceLegacyAdminCannotCRUDOwnTeam(t *testing.T) {
	ctx := context.Background()
	teamID := int64(7)
	otherTeamID := int64(8)
	admin := entity.Actor{UserID: 4, Role: policy.RoleAdmin, TeamID: &teamID}

	repo := &fakeRepo{}
	svc := NewService(repo)
	input := CreateInput{
		TeamID:       teamID,
		Title:        "Legacy admin plan",
		StartDate:    ptrTime(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)),
		LocationText: "HQ",
	}
	if _, err := svc.Create(ctx, admin, input); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("legacy admin create own team expected forbidden, got %v", err)
	}

	repo = &fakeRepo{}
	svc = NewService(repo)
	input = CreateInput{
		TeamID:       otherTeamID,
		Title:        "Sneaky plan",
		StartDate:    ptrTime(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)),
		LocationText: "Elsewhere",
	}
	if _, err := svc.Create(ctx, admin, input); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("legacy admin create other team expected forbidden, got %v", err)
	}

	ownPlan := entity.TeamPlan{ID: 50, TeamID: teamID, CreatedByUserID: 4, Status: entity.StatusPlanned}
	repo = &fakeRepo{getByIDResult: &ownPlan}
	svc = NewService(repo)
	if _, err := svc.GetByID(ctx, admin, ownPlan.ID); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("legacy admin get own team expected forbidden, got %v", err)
	}

	repo = &fakeRepo{getByIDResult: &ownPlan}
	svc = NewService(repo)
	newTitle := "Admin updated"
	if _, err := svc.Update(ctx, admin, UpdateInput{ID: ownPlan.ID, Title: &newTitle}); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("legacy admin update own team expected forbidden, got %v", err)
	}

	otherPlan := entity.TeamPlan{ID: 51, TeamID: otherTeamID, CreatedByUserID: 99, Status: entity.StatusPlanned}
	repo = &fakeRepo{getByIDResult: &otherPlan}
	svc = NewService(repo)
	if _, err := svc.Update(ctx, admin, UpdateInput{ID: otherPlan.ID, Title: &newTitle}); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("legacy admin update other team expected forbidden, got %v", err)
	}

	repo = &fakeRepo{getByIDResult: &ownPlan}
	svc = NewService(repo)
	if err := svc.Delete(ctx, admin, ownPlan.ID); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("legacy admin delete own team expected forbidden, got %v", err)
	}

	repo = &fakeRepo{getByIDResult: &otherPlan}
	svc = NewService(repo)
	if err := svc.Delete(ctx, admin, otherPlan.ID); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("legacy admin delete other team expected forbidden, got %v", err)
	}
}

func TestServiceLegacyAdminListForbidden(t *testing.T) {
	ctx := context.Background()
	teamID := int64(7)
	admin := entity.Actor{UserID: 4, Role: policy.RoleAdmin, TeamID: &teamID}

	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	repo := &fakeRepo{}
	svc := NewService(repo)
	if _, err := svc.List(ctx, admin, ListInput{From: from, To: to, TeamID: &teamID}); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("legacy admin list expected forbidden, got %v", err)
	}
	if repo.listCalled {
		t.Fatalf("repository should not be called for legacy admin list")
	}
}

type fakeRepo struct {
	listCalled    bool
	createCalled  bool
	updateCalled  bool
	deleteCalled  bool
	listInput     repository.ListQuery
	createInput   *repository.CreateInput
	updateInput   *repository.UpdateInput
	deleteInput   *repository.DeleteCommand
	getByIDResult *entity.TeamPlan
}

func (f *fakeRepo) List(_ context.Context, query repository.ListQuery) ([]entity.TeamPlan, int64, error) {
	f.listCalled = true
	f.listInput = query
	return nil, 0, nil
}

func (f *fakeRepo) ListAll(context.Context, repository.ListQuery) ([]entity.TeamPlan, error) {
	return nil, nil
}

func (f *fakeRepo) GetByID(context.Context, repository.GetQuery) (*entity.TeamPlan, error) {
	if f.getByIDResult == nil {
		return nil, entity.ErrNotFound
	}
	item := *f.getByIDResult
	return &item, nil
}

func (f *fakeRepo) Create(_ context.Context, input repository.CreateInput) (*entity.TeamPlan, error) {
	f.createCalled = true
	f.createInput = &input
	return &entity.TeamPlan{ID: 100, TeamID: input.TeamID, CreatedByUserID: input.CreatedByUserID, Title: input.Title, StartDate: input.StartDate, LocationText: input.LocationText, Status: input.Status}, nil
}

func (f *fakeRepo) Update(_ context.Context, input repository.UpdateInput) (*entity.TeamPlan, error) {
	f.updateCalled = true
	f.updateInput = &input
	teamID := int64(0)
	if input.TeamID != nil {
		teamID = *input.TeamID
	} else if f.getByIDResult != nil {
		teamID = f.getByIDResult.TeamID
	}
	var startDate *time.Time
	if input.StartDate != nil {
		startDate = input.StartDate
	} else if f.getByIDResult != nil {
		startDate = f.getByIDResult.StartDate
	}
	var endDate *time.Time
	if input.EndDate != nil {
		endDate = input.EndDate
	} else if f.getByIDResult != nil {
		endDate = f.getByIDResult.EndDate
	}
	var workTime *string
	if input.WorkTime != nil {
		workTime = input.WorkTime
	} else if f.getByIDResult != nil {
		workTime = f.getByIDResult.WorkTime
	}
	title := ""
	if input.Title != nil {
		title = *input.Title
	} else if f.getByIDResult != nil {
		title = f.getByIDResult.Title
	}
	location := ""
	if input.LocationText != nil {
		location = *input.LocationText
	} else if f.getByIDResult != nil {
		location = f.getByIDResult.LocationText
	}
	status := entity.StatusPlanned
	if input.Status != nil {
		status = *input.Status
	} else if f.getByIDResult != nil {
		status = f.getByIDResult.Status
	}
	return &entity.TeamPlan{ID: input.ID, TeamID: teamID, CreatedByUserID: 1, Title: title, StartDate: startDate, EndDate: endDate, WorkTime: workTime, LocationText: location, Status: status}, nil
}

func (f *fakeRepo) SoftDelete(_ context.Context, cmd repository.DeleteCommand) error {
	f.deleteCalled = true
	f.deleteInput = &cmd
	return nil
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

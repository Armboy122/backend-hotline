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

	input := CreateInput{
		TeamID:       teamID,
		Title:        "Patrol feeder A",
		StartDate:    time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC),
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
		t.Fatalf("admin create expected forbidden, got %v", err)
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
		StartDate:       time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC),
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
		StartDate:       time.Date(2026, 6, 3, 17, 0, 0, 0, time.Local),
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
	if _, err := svc.Update(ctx, user, UpdateInput{ID: completed.ID, Title: &newTitle}); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("user update completed expected forbidden, got %v", err)
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

	repo = &fakeRepo{getByIDResult: &plan}
	svc = NewService(repo)
	if err := svc.Delete(ctx, user, plan.ID); !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("user delete expected forbidden, got %v", err)
	}
	if repo.deleteCalled {
		t.Fatalf("repository should not be called on forbidden delete")
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
	return &entity.TeamPlan{ID: 100, TeamID: input.TeamID, CreatedByUserID: input.CreatedByUserID, Title: input.Title, StartDate: input.StartDate, LocationText: input.LocationText, Status: entity.StatusPlanned}, nil
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
	startDate := time.Time{}
	if input.StartDate != nil {
		startDate = *input.StartDate
	} else if f.getByIDResult != nil {
		startDate = f.getByIDResult.StartDate
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
	return &entity.TeamPlan{ID: input.ID, TeamID: teamID, CreatedByUserID: 1, Title: title, StartDate: startDate, LocationText: location, Status: status}, nil
}

func (f *fakeRepo) SoftDelete(_ context.Context, cmd repository.DeleteCommand) error {
	f.deleteCalled = true
	f.deleteInput = &cmd
	return nil
}

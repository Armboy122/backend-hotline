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
	capturedList                    repo.ListQuery
	capturedGet                     repo.GetQuery
	capturedCreate                  repo.CreateInput
	capturedUpdate                  repo.UpdateInput
	capturedReplace                 repo.ReplaceTasksInput
	capturedReplaceWithParticipants repo.ReplaceTasksAndParticipantsInput
	capturedListTasks               repo.ListTasksQuery
	capturedListAssigned            repo.ListAssignedTasksQuery
	capturedUpdateTask              repo.UpdateTaskInput

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

	replaceTasks                 []entity.LargeWorkTask
	replaceErr                   error
	replaceWithParticipantsTasks []entity.LargeWorkTask
	replaceWithParticipantsErr   error
	listTasks                    []entity.LargeWorkTask
	listTasksErr                 error
	assignedTasks                []entity.LargeWorkTask
	assignedTotal                int64
	assignedErr                  error
	getTask                      *entity.LargeWorkTask
	getTaskErr                   error
	updateTask                   *entity.LargeWorkTask
	updateTaskErr                error
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
func (f *fakeRepository) ReplaceTasks(_ context.Context, input repo.ReplaceTasksInput) ([]entity.LargeWorkTask, error) {
	f.capturedReplace = input
	return f.replaceTasks, f.replaceErr
}
func (f *fakeRepository) ReplaceTasksAndParticipants(_ context.Context, input repo.ReplaceTasksAndParticipantsInput) ([]entity.LargeWorkTask, error) {
	f.capturedReplaceWithParticipants = input
	return f.replaceWithParticipantsTasks, f.replaceWithParticipantsErr
}
func (f *fakeRepository) ListTasksByPlan(_ context.Context, q repo.ListTasksQuery) ([]entity.LargeWorkTask, error) {
	f.capturedListTasks = q
	return f.listTasks, f.listTasksErr
}
func (f *fakeRepository) ListAssignedTasks(_ context.Context, q repo.ListAssignedTasksQuery) ([]entity.LargeWorkTask, int64, error) {
	f.capturedListAssigned = q
	return f.assignedTasks, f.assignedTotal, f.assignedErr
}
func (f *fakeRepository) GetTaskByID(_ context.Context, q repo.GetTaskQuery) (*entity.LargeWorkTask, error) {
	return f.getTask, f.getTaskErr
}
func (f *fakeRepository) UpdateTask(_ context.Context, input repo.UpdateTaskInput) (*entity.LargeWorkTask, error) {
	f.capturedUpdateTask = input
	return f.updateTask, f.updateTaskErr
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

func TestCreateAllowsTeamLeadForOwnTeamAndRejectsOtherTeam(t *testing.T) {
	ownerTeamID := int64(7)
	repo := &fakeRepository{createItem: &entity.LargeWorkItem{ID: 1, OwnerTeamID: ownerTeamID}}
	svc := NewService(repo)

	out, err := svc.Create(context.Background(), actor(policy.RoleTeamLead, &ownerTeamID), CreateInput{
		OwnerTeamID:        ownerTeamID,
		ParticipantTeamIDs: []int64{ownerTeamID, 9},
		Title:              "งานระดมทีม",
		StartDate:          mustDate(t, "2026-06-10"),
		LocationText:       "Station A",
	})
	if err != nil {
		t.Fatalf("team lead own-team create got %v, want nil", err)
	}
	if out == nil || repo.capturedCreate.CreatedByUserID != 1 {
		t.Fatalf("team lead create result = %#v captured = %#v, want created item with actor id", out, repo.capturedCreate)
	}

	otherTeamID := int64(8)
	_, err = svc.Create(context.Background(), actor(policy.RoleTeamLead, &ownerTeamID), CreateInput{
		OwnerTeamID:        otherTeamID,
		ParticipantTeamIDs: []int64{otherTeamID, 9},
		Title:              "งานระดมทีม",
		StartDate:          mustDate(t, "2026-06-10"),
		LocationText:       "Station A",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("team lead other-team create got %v, want ErrForbidden", err)
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

func TestUpdateAndCancelAllowTeamLeadForOwnTeamEditableWork(t *testing.T) {
	ownerTeamID := int64(7)
	existing := &entity.LargeWorkItem{ID: 99, OwnerTeamID: ownerTeamID, CreatedByUserID: 1, Status: entity.LargeWorkStatusPlanned}
	repo := &fakeRepository{getItem: existing, updateItem: existing}
	svc := NewService(repo)

	if _, err := svc.Update(context.Background(), actor(policy.RoleTeamLead, &ownerTeamID), UpdateInput{ID: 99, Title: strPtr("updated")}); err != nil {
		t.Fatalf("team lead own-team update error: %v", err)
	}
	if repo.capturedUpdate.ID != 99 {
		t.Fatalf("update id = %d, want 99", repo.capturedUpdate.ID)
	}
	if _, err := svc.Cancel(context.Background(), actor(policy.RoleTeamLead, &ownerTeamID), 99); err != nil {
		t.Fatalf("team lead own-team cancel error: %v", err)
	}

	otherTeamID := int64(8)
	createdByLead := &entity.LargeWorkItem{ID: 100, OwnerTeamID: otherTeamID, CreatedByUserID: 1, Status: entity.LargeWorkStatusPlanned}
	if _, err := NewService(&fakeRepository{getItem: createdByLead, updateItem: createdByLead}).Update(context.Background(), actor(policy.RoleTeamLead, &ownerTeamID), UpdateInput{ID: 100, Title: strPtr("creator update")}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("team lead creator update for another team got %v, want ErrForbidden", err)
	}

	unrelated := &entity.LargeWorkItem{ID: 101, OwnerTeamID: otherTeamID, CreatedByUserID: 999, Status: entity.LargeWorkStatusPlanned}
	if _, err := NewService(&fakeRepository{getItem: unrelated, updateItem: unrelated}).Update(context.Background(), actor(policy.RoleTeamLead, &ownerTeamID), UpdateInput{ID: 101, Title: strPtr("blocked")}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("team lead unrelated update got %v, want ErrForbidden", err)
	}

	if _, err := NewService(&fakeRepository{getItem: existing, updateItem: existing}).Update(context.Background(), actor(policy.RoleUser, &ownerTeamID), UpdateInput{ID: 99, Title: strPtr("blocked")}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("user update got %v, want ErrForbidden", err)
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

func TestOverviewIncludesProgressAndAllowsAllTeamsToView(t *testing.T) {
	ownerTeamID := int64(7)
	workerTeamID := int64(8)
	unrelatedTeamID := int64(99)
	plan := &entity.LargeWorkItem{ID: 501, OwnerTeamID: ownerTeamID, Status: entity.LargeWorkStatusPlanned, Teams: []entity.LargeWorkTeam{{ID: ownerTeamID}, {ID: workerTeamID}}}
	repo := &fakeRepository{getItem: plan, listTasks: []entity.LargeWorkTask{
		{ID: 1, LargeWorkItemID: 501, AssignedTeamID: workerTeamID, Status: entity.LargeWorkTaskStatusTodo},
		{ID: 2, LargeWorkItemID: 501, AssignedTeamID: workerTeamID, Status: entity.LargeWorkTaskStatusDone},
		{ID: 3, LargeWorkItemID: 501, AssignedTeamID: ownerTeamID, Status: entity.LargeWorkTaskStatusBlocked},
	}}
	svc := NewService(repo)

	out, err := svc.GetOverview(context.Background(), actor(policy.RoleUser, &unrelatedTeamID), 501)
	if err != nil {
		t.Fatalf("overview error: %v", err)
	}
	if out.Plan.ID != 501 || out.Progress.Total != 3 || out.Progress.Done != 1 || out.Progress.Blocked != 1 || out.Progress.Todo != 1 {
		t.Fatalf("overview = %#v", out)
	}
	if len(out.TeamProgress) != 2 || out.TeamProgress[0].AssignedTeamID != workerTeamID || out.TeamProgress[0].Total != 2 {
		t.Fatalf("team progress = %#v", out.TeamProgress)
	}
}

func TestReplaceTasksAllowsTeamLeadForParticipatingTeamsAndPreservesPlanningBoardFields(t *testing.T) {
	ownerTeamID := int64(7)
	participantTeamID := int64(8)
	plan := &entity.LargeWorkItem{ID: 501, OwnerTeamID: ownerTeamID, Status: entity.LargeWorkStatusPlanned, Teams: []entity.LargeWorkTeam{{ID: ownerTeamID}, {ID: participantTeamID}}}
	workDetail := "trim branches near primary line"
	notes := "bring cones"
	pointCount := 2
	treeCount := 5
	itemCount := 1
	lat := 13.756331
	lng := 100.501762
	repo := &fakeRepository{getItem: plan, replaceWithParticipantsTasks: []entity.LargeWorkTask{{ID: 11, LargeWorkItemID: 501, AssignedTeamID: participantTeamID}}}
	svc := NewService(repo)

	_, err := svc.ReplaceTasks(context.Background(), actor(policy.RoleTeamLead, &ownerTeamID), 501, ReplaceTasksInput{Tasks: []TaskPointInput{{
		AssignedTeamID: participantTeamID,
		Sequence:       3,
		PointLabel:     "P3",
		Latitude:       &lat,
		Longitude:      &lng,
		WorkType:       "tree_trimming",
		WorkDetail:     &workDetail,
		PointCount:     &pointCount,
		TreeCount:      &treeCount,
		ItemCount:      &itemCount,
		Notes:          &notes,
		Metadata:       map[string]any{"source": "planning-board", "color": "red"},
	}}})
	if err != nil {
		t.Fatalf("replace tasks error: %v", err)
	}
	if repo.capturedReplaceWithParticipants.LargeWorkItemID != 501 || len(repo.capturedReplaceWithParticipants.Tasks) != 1 {
		t.Fatalf("replace capture = %#v", repo.capturedReplaceWithParticipants)
	}
	if len(repo.capturedReplaceWithParticipants.ParticipantTeamIDs) != 2 {
		t.Fatalf("participant update = %#v, want existing participants only", repo.capturedReplaceWithParticipants.ParticipantTeamIDs)
	}
	got := repo.capturedReplaceWithParticipants.Tasks[0]
	if got.AssignedTeamID != participantTeamID || got.Sequence != 3 || got.PointLabel != "P3" || got.WorkType != "tree_trimming" || got.WorkDetail == nil || *got.WorkDetail != workDetail {
		t.Fatalf("basic task fields not preserved: %#v", got)
	}
	if got.Latitude == nil || *got.Latitude != lat || got.Longitude == nil || *got.Longitude != lng || got.PointCount == nil || *got.PointCount != pointCount || got.TreeCount == nil || *got.TreeCount != treeCount || got.ItemCount == nil || *got.ItemCount != itemCount || got.Notes == nil || *got.Notes != notes {
		t.Fatalf("planning board task fields not preserved: %#v", got)
	}
	if got.Metadata["source"] != "planning-board" || got.Metadata["color"] != "red" || got.Status != entity.LargeWorkTaskStatusTodo {
		t.Fatalf("metadata/status not preserved/defaulted: %#v", got)
	}

	otherTeamID := int64(10)
	_, err = NewService(&fakeRepository{getItem: plan}).ReplaceTasks(context.Background(), actor(policy.RoleTeamLead, &otherTeamID), 501, ReplaceTasksInput{Tasks: []TaskPointInput{{AssignedTeamID: participantTeamID, Sequence: 1, PointLabel: "P1", WorkType: "tree"}}})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("other team lead replace got %v, want ErrForbidden", err)
	}
}

func TestReplaceTasksRejectsUnassignedTaskTeams(t *testing.T) {
	ownerTeamID := int64(7)
	unassignedTeamID := int64(9)
	plan := &entity.LargeWorkItem{ID: 501, OwnerTeamID: ownerTeamID, Status: entity.LargeWorkStatusPlanned, Teams: []entity.LargeWorkTeam{{ID: ownerTeamID}, {ID: 8}}}
	repo := &fakeRepository{getItem: plan}
	svc := NewService(repo)

	_, err := svc.ReplaceTasks(context.Background(), actor(policy.RoleTeamLead, &ownerTeamID), 501, ReplaceTasksInput{Tasks: []TaskPointInput{{AssignedTeamID: unassignedTeamID, Sequence: 1, PointLabel: "P1", WorkType: "tree"}}})
	if !errors.Is(err, ErrInvalidTask) {
		t.Fatalf("unassigned task team got %v, want ErrInvalidTask", err)
	}
	if repo.capturedReplaceWithParticipants.LargeWorkItemID != 0 {
		t.Fatalf("replace mutated repository despite unassigned task team: %#v", repo.capturedReplaceWithParticipants)
	}
}

func TestMyTodosStartPhotoCompleteAndBlockEnforceOwnTeamAndPhotos(t *testing.T) {
	teamID := int64(8)
	workerActor := actor(policy.RoleUser, &teamID)
	task := &entity.LargeWorkTask{ID: 11, LargeWorkItemID: 501, AssignedTeamID: teamID, Status: entity.LargeWorkTaskStatusTodo}
	plan := &entity.LargeWorkItem{ID: 501, OwnerTeamID: 7, Status: entity.LargeWorkStatusInProgress, Teams: []entity.LargeWorkTeam{{ID: teamID}}}
	repo := &fakeRepository{getItem: plan, assignedTasks: []entity.LargeWorkTask{*task}, assignedTotal: 1, getTask: task, updateTask: task}
	svc := NewService(repo)

	todos, err := svc.MyTodos(context.Background(), workerActor, MyTodosInput{Page: 0, Limit: 500})
	if err != nil || todos.Total != 1 || repo.capturedListAssigned.AssignedTeamID != teamID || todos.Limit != 100 {
		t.Fatalf("my todos out=%#v err=%v captured=%#v", todos, err, repo.capturedListAssigned)
	}
	started, err := svc.StartTask(context.Background(), workerActor, 11)
	if err != nil || started == nil || repo.capturedUpdateTask.Status == nil || *repo.capturedUpdateTask.Status != entity.LargeWorkTaskStatusInProgress || repo.capturedUpdateTask.StartedByUserID == nil {
		t.Fatalf("start result=%#v err=%v update=%#v", started, err, repo.capturedUpdateTask)
	}

	if _, err := svc.AttachTaskPhoto(context.Background(), workerActor, 11, PhotoKindBefore, "   "); !errors.Is(err, ErrPhotoRequired) {
		t.Fatalf("attach blank before got %v, want ErrPhotoRequired", err)
	}
	beforeURL := "https://cdn.example/before.jpg"
	if _, err := svc.AttachTaskPhoto(context.Background(), workerActor, 11, PhotoKindBefore, beforeURL); err != nil {
		t.Fatalf("attach before: %v", err)
	}
	if len(repo.capturedUpdateTask.BeforePhotoURLs) != 1 || repo.capturedUpdateTask.BeforePhotoURLs[0] != beforeURL {
		t.Fatalf("before update = %#v", repo.capturedUpdateTask)
	}

	inProgress := &entity.LargeWorkTask{ID: 11, AssignedTeamID: teamID, Status: entity.LargeWorkTaskStatusInProgress, BeforePhotoURLs: []string{beforeURL}}
	repo.getTask = inProgress
	if _, err := svc.CompleteTask(context.Background(), workerActor, 11, CompleteTaskInput{CompletionNote: "done"}); !errors.Is(err, ErrPhotoRequired) {
		t.Fatalf("complete without after got %v, want ErrPhotoRequired", err)
	}
	if _, err := svc.CompleteTask(context.Background(), workerActor, 11, CompleteTaskInput{AfterPhotoURLs: []string{""}, CompletionNote: "done"}); !errors.Is(err, ErrPhotoRequired) {
		t.Fatalf("complete with empty after photo got %v, want ErrPhotoRequired", err)
	}
	done, err := svc.CompleteTask(context.Background(), workerActor, 11, CompleteTaskInput{AfterPhotoURLs: []string{"https://cdn.example/after.jpg"}, CompletionNote: "done"})
	if err != nil || done == nil || repo.capturedUpdateTask.Status == nil || *repo.capturedUpdateTask.Status != entity.LargeWorkTaskStatusDone || repo.capturedUpdateTask.CompletedByUserID == nil {
		t.Fatalf("complete result=%#v err=%v update=%#v", done, err, repo.capturedUpdateTask)
	}

	repo.getTask = task
	blocked, err := svc.BlockTask(context.Background(), workerActor, 11, "access denied")
	if err != nil || blocked == nil || repo.capturedUpdateTask.Status == nil || *repo.capturedUpdateTask.Status != entity.LargeWorkTaskStatusBlocked || repo.capturedUpdateTask.Notes == nil || *repo.capturedUpdateTask.Notes != "access denied" {
		t.Fatalf("block result=%#v err=%v update=%#v", blocked, err, repo.capturedUpdateTask)
	}

	otherTeamID := int64(99)
	_, err = svc.StartTask(context.Background(), actor(policy.RoleUser, &otherTeamID), 11)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("other team start got %v, want ErrForbidden", err)
	}
}

func TestReplaceTasksReturnsClearSchemaUnavailableError(t *testing.T) {
	ownerTeamID := int64(7)
	plan := &entity.LargeWorkItem{ID: 501, OwnerTeamID: ownerTeamID, Status: entity.LargeWorkStatusPlanned, Teams: []entity.LargeWorkTeam{{ID: ownerTeamID}, {ID: 8}}}
	repo := &fakeRepository{getItem: plan, listTasksErr: repo.ErrSchemaUnavailable}
	svc := NewService(repo)

	_, err := svc.ReplaceTasks(context.Background(), actor(policy.RoleTeamLead, &ownerTeamID), 501, ReplaceTasksInput{Tasks: []TaskPointInput{{AssignedTeamID: 8, Sequence: 1, PointLabel: "P1", WorkType: "tree"}}})
	if !errors.Is(err, ErrTaskSchemaUnavailable) {
		t.Fatalf("schema unavailable got %v, want ErrTaskSchemaUnavailable", err)
	}
}

func TestReplaceTasksRejectsExecutedTasksBeforeMutatingParticipants(t *testing.T) {
	ownerTeamID := int64(7)
	plan := &entity.LargeWorkItem{ID: 501, OwnerTeamID: ownerTeamID, Status: entity.LargeWorkStatusPlanned, Teams: []entity.LargeWorkTeam{{ID: ownerTeamID}, {ID: 8}}}
	repo := &fakeRepository{getItem: plan, listTasks: []entity.LargeWorkTask{{ID: 1, LargeWorkItemID: 501, AssignedTeamID: 8, Status: entity.LargeWorkTaskStatusInProgress}}}
	svc := NewService(repo)

	_, err := svc.ReplaceTasks(context.Background(), actor(policy.RoleTeamLead, &ownerTeamID), 501, ReplaceTasksInput{Tasks: []TaskPointInput{{AssignedTeamID: 8, Sequence: 1, PointLabel: "P1", WorkType: "tree"}}})
	if !errors.Is(err, ErrInvalidStateTransition) {
		t.Fatalf("replace with executed tasks got %v, want ErrInvalidStateTransition", err)
	}
	if repo.capturedReplaceWithParticipants.LargeWorkItemID != 0 || repo.capturedUpdate.ID != 0 {
		t.Fatalf("mutation captured despite rejection: update=%#v replace=%#v", repo.capturedUpdate, repo.capturedReplaceWithParticipants)
	}
}

func TestTaskExecutionRejectsInactiveParentPlan(t *testing.T) {
	teamID := int64(8)
	workerActor := actor(policy.RoleUser, &teamID)
	for _, status := range []string{entity.LargeWorkStatusCompleted, entity.LargeWorkStatusCancelled} {
		t.Run(status, func(t *testing.T) {
			parent := &entity.LargeWorkItem{ID: 501, OwnerTeamID: 7, Status: status, Teams: []entity.LargeWorkTeam{{ID: teamID}}}
			task := &entity.LargeWorkTask{ID: 11, LargeWorkItemID: 501, AssignedTeamID: teamID, Status: entity.LargeWorkTaskStatusTodo, BeforePhotoURLs: []string{"https://cdn.example/before.jpg"}}
			repo := &fakeRepository{getItem: parent, getTask: task, updateTask: task}
			svc := NewService(repo)

			if _, err := svc.StartTask(context.Background(), workerActor, 11); !errors.Is(err, ErrInvalidStateTransition) {
				t.Fatalf("start with parent %s got %v, want ErrInvalidStateTransition", status, err)
			}
			if _, err := svc.AttachTaskPhoto(context.Background(), workerActor, 11, PhotoKindBefore, "https://cdn.example/before2.jpg"); !errors.Is(err, ErrInvalidStateTransition) {
				t.Fatalf("attach with parent %s got %v, want ErrInvalidStateTransition", status, err)
			}

			task.Status = entity.LargeWorkTaskStatusInProgress
			if _, err := svc.CompleteTask(context.Background(), workerActor, 11, CompleteTaskInput{AfterPhotoURLs: []string{"https://cdn.example/after.jpg"}, CompletionNote: "done"}); !errors.Is(err, ErrInvalidStateTransition) {
				t.Fatalf("complete with parent %s got %v, want ErrInvalidStateTransition", status, err)
			}
			if _, err := svc.BlockTask(context.Background(), workerActor, 11, "blocked"); !errors.Is(err, ErrInvalidStateTransition) {
				t.Fatalf("block with parent %s got %v, want ErrInvalidStateTransition", status, err)
			}
			if repo.capturedUpdateTask.ID != 0 {
				t.Fatalf("update captured despite inactive parent: %#v", repo.capturedUpdateTask)
			}
		})
	}
}

func TestTaskMutationsRejectTerminalStatuses(t *testing.T) {
	teamID := int64(8)
	for _, status := range []string{entity.LargeWorkTaskStatusDone, entity.LargeWorkTaskStatusCancelled} {
		t.Run(status, func(t *testing.T) {
			task := &entity.LargeWorkTask{ID: 11, LargeWorkItemID: 501, AssignedTeamID: teamID, Status: status}
			plan := &entity.LargeWorkItem{ID: 501, OwnerTeamID: 7, Status: entity.LargeWorkStatusInProgress, Teams: []entity.LargeWorkTeam{{ID: teamID}}}
			repo := &fakeRepository{getItem: plan, getTask: task, updateTask: task}
			svc := NewService(repo)
			_, err := svc.BlockTask(context.Background(), actor(policy.RoleUser, &teamID), 11, "blocked")
			if !errors.Is(err, ErrInvalidStateTransition) {
				t.Fatalf("block %s got %v, want ErrInvalidStateTransition", status, err)
			}
			if repo.capturedUpdateTask.ID != 0 {
				t.Fatalf("update captured despite terminal status: %#v", repo.capturedUpdateTask)
			}

			_, err = svc.AttachTaskPhoto(context.Background(), actor(policy.RoleUser, &teamID), 11, PhotoKindAfter, "https://cdn.example/after.jpg")
			if !errors.Is(err, ErrInvalidStateTransition) {
				t.Fatalf("attach photo to %s got %v, want ErrInvalidStateTransition", status, err)
			}
			if repo.capturedUpdateTask.ID != 0 {
				t.Fatalf("photo update captured despite terminal status: %#v", repo.capturedUpdateTask)
			}
		})
	}
}

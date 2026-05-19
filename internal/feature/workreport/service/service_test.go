package service

import (
	"context"
	"errors"
	"testing"
	"time"

	taskentity "backend-hotlines3/internal/feature/task/entity"
	taskrepository "backend-hotlines3/internal/feature/task/repository"
)

// --- fake taskLister for tests ---

type fakeTaskLister struct {
	capturedQuery taskrepository.ByFilterQuery
	tasks         []taskentity.Task
	err           error
}

func (f *fakeTaskLister) ListByFilter(_ context.Context, query taskrepository.ByFilterQuery) ([]taskentity.Task, error) {
	f.capturedQuery = query
	return f.tasks, f.err
}

// --- helpers ---

func strPtr(s string) *string     { return &s }
func int64Ptr(i int64) *int64     { return &i }
func floatPtr(f float64) *float64 { return &f }

func makeActor(role string, teamID *int64) taskentity.Actor {
	return taskentity.Actor{UserID: 1, Role: role, TeamID: teamID}
}

func makeTask(id int64, teamID int64, teamName *string, workDate time.Time) taskentity.Task {
	return taskentity.Task{
		ID:          id,
		WorkDate:    workDate,
		TeamID:      teamID,
		JobTypeID:   1,
		JobDetailID: 2,
		TeamName:    teamName,
		CreatedAt:   time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC),
	}
}

// --- Tests (RED phase: these should pass once service is implemented) ---

func TestGetReportRequiresYearAndMonth(t *testing.T) {
	svc := NewService(&fakeTaskLister{})

	_, err := svc.GetReport(context.Background(), makeActor("super_admin", nil), GetReportInput{Year: "", Month: "5"})
	if !errors.Is(err, ErrYearMonthRequired) {
		t.Fatalf("got %v, want ErrYearMonthRequired", err)
	}

	_, err = svc.GetReport(context.Background(), makeActor("super_admin", nil), GetReportInput{Year: "2026", Month: ""})
	if !errors.Is(err, ErrYearMonthRequired) {
		t.Fatalf("got %v, want ErrYearMonthRequired", err)
	}
}

func TestGetReportPassesYearMonthAndTeamToRepository(t *testing.T) {
	fake := &fakeTaskLister{}
	svc := NewService(fake)

	teamID := int64(7)
	_, _ = svc.GetReport(context.Background(), makeActor("super_admin", nil), GetReportInput{
		Year:   "2026",
		Month:  "5",
		TeamID: &teamID,
	})

	if fake.capturedQuery.Year != "2026" {
		t.Fatalf("Year = %q, want 2026", fake.capturedQuery.Year)
	}
	if fake.capturedQuery.Month != "5" {
		t.Fatalf("Month = %q, want 5", fake.capturedQuery.Month)
	}
	if fake.capturedQuery.TeamID == nil || *fake.capturedQuery.TeamID != 7 {
		t.Fatalf("TeamID = %#v, want 7", fake.capturedQuery.TeamID)
	}
}

func TestGetReportScopesTeamForNonPrivilegedUser(t *testing.T) {
	fake := &fakeTaskLister{}
	svc := NewService(fake)

	ownTeamID := int64(42)
	requestedTeamID := int64(99)
	_, _ = svc.GetReport(context.Background(), makeActor("user", &ownTeamID), GetReportInput{
		Year:   "2026",
		Month:  "5",
		TeamID: &requestedTeamID,
	})

	// Non-privileged user should be scoped to their own team.
	if fake.capturedQuery.TeamID == nil || *fake.capturedQuery.TeamID != ownTeamID {
		t.Fatalf("TeamID = %#v, want own team %d", fake.capturedQuery.TeamID, ownTeamID)
	}
}

func TestGetReportAllowsAdminToReadRequestedTeam(t *testing.T) {
	fake := &fakeTaskLister{}
	svc := NewService(fake)

	ownTeamID := int64(42)
	requestedTeamID := int64(99)
	_, _ = svc.GetReport(context.Background(), makeActor("admin", &ownTeamID), GetReportInput{
		Year:   "2026",
		Month:  "5",
		TeamID: &requestedTeamID,
	})

	if fake.capturedQuery.TeamID == nil || *fake.capturedQuery.TeamID != requestedTeamID {
		t.Fatalf("TeamID = %#v, want requested team %d", fake.capturedQuery.TeamID, requestedTeamID)
	}
}

func TestGetReportReturnsSummaryAndItems(t *testing.T) {
	date1 := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)

	fake := &fakeTaskLister{
		tasks: []taskentity.Task{
			makeTask(1, 10, strPtr("Ops East"), date1),
			makeTask(2, 10, strPtr("Ops East"), date2),
			makeTask(3, 20, strPtr("Ops West"), date1),
		},
	}
	svc := NewService(fake)

	out, err := svc.GetReport(context.Background(), makeActor("super_admin", nil), GetReportInput{
		Year:  "2026",
		Month: "5",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Summary
	if out.Summary.TotalTasks != 3 {
		t.Fatalf("TotalTasks = %d, want 3", out.Summary.TotalTasks)
	}
	if out.Summary.UniqueTeamCount != 2 {
		t.Fatalf("UniqueTeamCount = %d, want 2", out.Summary.UniqueTeamCount)
	}

	// Team summaries
	teamMap := map[int64]int64{}
	for _, ts := range out.Summary.TeamSummaries {
		teamMap[ts.TeamID] = ts.Count
	}
	if teamMap[10] != 2 {
		t.Fatalf("team 10 count = %d, want 2", teamMap[10])
	}
	if teamMap[20] != 1 {
		t.Fatalf("team 20 count = %d, want 1", teamMap[20])
	}

	// Items
	if len(out.Items) != 3 {
		t.Fatalf("items len = %d, want 3", len(out.Items))
	}
	if out.Items[0].WorkDate != "2026-05-01" {
		t.Fatalf("item[0] WorkDate = %q, want 2026-05-01", out.Items[0].WorkDate)
	}
	if out.Items[0].TeamName == nil || *out.Items[0].TeamName != "Ops East" {
		t.Fatalf("item[0] TeamName = %#v, want Ops East", out.Items[0].TeamName)
	}
}

func TestGetReportReturnsEmptyWhenNoTasks(t *testing.T) {
	fake := &fakeTaskLister{tasks: []taskentity.Task{}}
	svc := NewService(fake)

	out, err := svc.GetReport(context.Background(), makeActor("super_admin", nil), GetReportInput{
		Year:  "2026",
		Month: "5",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Summary.TotalTasks != 0 {
		t.Fatalf("TotalTasks = %d, want 0", out.Summary.TotalTasks)
	}
	if out.Summary.UniqueTeamCount != 0 {
		t.Fatalf("UniqueTeamCount = %d, want 0", out.Summary.UniqueTeamCount)
	}
	if len(out.Items) != 0 {
		t.Fatalf("items len = %d, want 0", len(out.Items))
	}
}

func TestGetReportPropagatesRepositoryError(t *testing.T) {
	fake := &fakeTaskLister{err: errors.New("db failure")}
	svc := NewService(fake)

	_, err := svc.GetReport(context.Background(), makeActor("super_admin", nil), GetReportInput{
		Year:  "2026",
		Month: "5",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetReportMapsAllTaskFields(t *testing.T) {
	workDate := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 15, 8, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 15, 9, 0, 0, 0, time.UTC)

	task := taskentity.Task{
		ID:                  42,
		WorkDate:            workDate,
		TeamID:              7,
		JobTypeID:           3,
		JobDetailID:         11,
		FeederID:            int64Ptr(99),
		NumPole:             strPtr("A1-B2"),
		DeviceCode:          strPtr("DC-001"),
		Detail:              strPtr("checked and repaired"),
		URLsBefore:          []string{"before1.jpg"},
		URLsAfter:           []string{"after1.jpg", "after2.jpg"},
		Latitude:            floatPtr(13.7563),
		Longitude:           floatPtr(100.5018),
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
		TeamName:            strPtr("Ops East"),
		JobTypeName:         strPtr("Inspection"),
		JobDetailName:       strPtr("Visual check"),
		FeederCode:          strPtr("F-001"),
		StationName:         strPtr("Bang Khen"),
		OperationCenterID:   int64Ptr(5),
		OperationCenterName: strPtr("North Center"),
	}

	fake := &fakeTaskLister{tasks: []taskentity.Task{task}}
	svc := NewService(fake)

	out, err := svc.GetReport(context.Background(), makeActor("super_admin", nil), GetReportInput{
		Year: "2026", Month: "5",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(out.Items))
	}
	it := out.Items[0]

	assertField := func(name string, got, want interface{}) {
		t.Helper()
		switch g := got.(type) {
		case string:
			if w, ok := want.(string); ok && g != w {
				t.Fatalf("%s = %q, want %q", name, g, w)
			}
		case int64:
			if w, ok := want.(int64); ok && g != w {
				t.Fatalf("%s = %d, want %d", name, g, w)
			}
		case *int64:
			if want == nil {
				if g != nil {
					t.Fatalf("%s = %#v, want nil", name, g)
				}
			} else {
				if g == nil {
					t.Fatalf("%s = nil, want %v", name, want)
				}
				w := want.(int64)
				if *g != w {
					t.Fatalf("%s = %d, want %d", name, *g, w)
				}
			}
		case *string:
			if want == nil {
				if g != nil {
					t.Fatalf("%s = %#v, want nil", name, g)
				}
			} else {
				if g == nil {
					t.Fatalf("%s = nil, want %v", name, want)
				}
				w := want.(string)
				if *g != w {
					t.Fatalf("%s = %q, want %q", name, *g, w)
				}
			}
		case *float64:
			if g == nil {
				t.Fatalf("%s = nil, want non-nil", name)
			}
			w := want.(float64)
			if *g != w {
				t.Fatalf("%s = %v, want %v", name, *g, w)
			}
		case []string:
			w := want.([]string)
			if len(g) != len(w) {
				t.Fatalf("%s len = %d, want %d", name, len(g), len(w))
			}
			for i := range g {
				if g[i] != w[i] {
					t.Fatalf("%s[%d] = %q, want %q", name, i, g[i], w[i])
				}
			}
		}
	}

	assertField("ID", it.ID, int64(42))
	assertField("WorkDate", it.WorkDate, "2026-05-15")
	assertField("TeamID", it.TeamID, int64(7))
	assertField("JobTypeID", it.JobTypeID, int64(3))
	assertField("JobDetailID", it.JobDetailID, int64(11))
	assertField("FeederID", it.FeederID, int64(99))
	assertField("NumPole", it.NumPole, "A1-B2")
	assertField("DeviceCode", it.DeviceCode, "DC-001")
	assertField("Detail", it.Detail, "checked and repaired")
	assertField("Latitude", it.Latitude, 13.7563)
	assertField("Longitude", it.Longitude, 100.5018)
	assertField("TeamName", it.TeamName, "Ops East")
	assertField("JobTypeName", it.JobTypeName, "Inspection")
	assertField("JobDetailName", it.JobDetailName, "Visual check")
	assertField("FeederCode", it.FeederCode, "F-001")
	assertField("StationName", it.StationName, "Bang Khen")
	assertField("OperationCenterID", it.OperationCenterID, int64(5))
	assertField("OperationCenterName", it.OperationCenterName, "North Center")
	assertField("URLsBefore", it.URLsBefore, []string{"before1.jpg"})
	assertField("URLsAfter", it.URLsAfter, []string{"after1.jpg", "after2.jpg"})
}

func TestGetReportNilURLsDefaultToEmptySlice(t *testing.T) {
	task := makeTask(1, 7, strPtr("Team A"), time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC))
	task.URLsBefore = nil
	task.URLsAfter = nil

	fake := &fakeTaskLister{tasks: []taskentity.Task{task}}
	svc := NewService(fake)

	out, err := svc.GetReport(context.Background(), makeActor("super_admin", nil), GetReportInput{
		Year: "2026", Month: "5",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Items[0].URLsBefore == nil {
		t.Fatal("URLsBefore should not be nil")
	}
	if len(out.Items[0].URLsBefore) != 0 {
		t.Fatalf("URLsBefore len = %d, want 0", len(out.Items[0].URLsBefore))
	}
	if out.Items[0].URLsAfter == nil {
		t.Fatal("URLsAfter should not be nil")
	}
	if len(out.Items[0].URLsAfter) != 0 {
		t.Fatalf("URLsAfter len = %d, want 0", len(out.Items[0].URLsAfter))
	}
}

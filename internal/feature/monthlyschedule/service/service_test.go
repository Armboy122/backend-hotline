package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"backend-hotlines3/internal/feature/monthlyschedule/entity"
	"backend-hotlines3/internal/models"
)

func TestSaveDraftRejectsOverlapsAndNonSuperAdmin(t *testing.T) {
	repo := newFakeRepository()
	svc := New(repo)
	start := mustDate(t, "2026-07-02")
	end := mustDate(t, "2026-07-05")

	_, err := svc.SaveDraft(context.Background(), entity.Actor{UserID: 1, Role: "team_lead"}, 2026, 7, nil, []entity.Assignment{{
		TeamID: 1, AssignmentType: "field", StartDate: start, EndDate: end, Destination: "กฟส. บ้านไผ่",
	}})
	if !errors.Is(err, entity.ErrForbidden) {
		t.Fatalf("non-super-admin error = %v, want forbidden", err)
	}

	_, err = svc.SaveDraft(context.Background(), superAdmin(), 2026, 7, nil, []entity.Assignment{
		{TeamID: 1, AssignmentType: "field", StartDate: start, EndDate: end, Destination: "กฟส. บ้านไผ่"},
		{TeamID: 1, AssignmentType: "support", StartDate: mustDate(t, "2026-07-05"), EndDate: mustDate(t, "2026-07-06"), Destination: "กฟจ. ขอนแก่น"},
	})
	if !errors.Is(err, entity.ErrOverlappingAssignment) {
		t.Fatalf("overlap error = %v, want ErrOverlappingAssignment", err)
	}
}

func TestPublishRequiresCompleteTeamMetadata(t *testing.T) {
	repo := newFakeRepository()
	repo.teams[0].Code = nil
	repo.draft = schedule(1, []entity.Assignment{})
	svc := New(repo)

	_, err := svc.Publish(context.Background(), superAdmin(), 2026, 7, 1)
	if !errors.Is(err, entity.ErrTeamMetadataMissing) {
		t.Fatalf("publish error = %v, want ErrTeamMetadataMissing", err)
	}
	if repo.publishCalls != 0 {
		t.Fatalf("invalid draft must not be published")
	}
}

func TestPublishDerivesHomeGapsAndKeepsAwaySegments(t *testing.T) {
	repo := newFakeRepository()
	repo.draft = schedule(1, []entity.Assignment{{
		ID:             88,
		TeamID:         1,
		AssignmentType: models.MonthlyPlanAssignmentTypeRemote,
		StartDate:      mustDate(t, "2026-07-10"),
		EndDate:        mustDate(t, "2026-07-12"),
		Destination:    "กฟอ. ชุมแพ",
		SourceType:     models.MonthlyPlanAssignmentSourceManual,
	}})
	svc := New(repo)

	got, err := svc.Publish(context.Background(), superAdmin(), 2026, 7, 1)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if repo.publishCalls != 1 {
		t.Fatalf("publish calls = %d, want 1", repo.publishCalls)
	}
	if len(got.Teams) != 2 {
		t.Fatalf("teams = %d, want 2", len(got.Teams))
	}
	first := got.Teams[0]
	if len(first.Segments) != 3 {
		t.Fatalf("team 1 segments = %+v, want home-away-home", first.Segments)
	}
	want := []struct {
		kind, start, end string
	}{
		{"home", "2026-07-01", "2026-07-09"},
		{"remote", "2026-07-10", "2026-07-12"},
		{"home", "2026-07-13", "2026-07-31"},
	}
	for i, expected := range want {
		segment := first.Segments[i]
		if segment.AssignmentType != expected.kind || segment.StartDate != expected.start || segment.EndDate != expected.end {
			t.Fatalf("segment %d = %+v, want %+v", i, segment, expected)
		}
	}
	if got.Checksum == "" || len(got.Checksum) != 64 {
		t.Fatalf("checksum = %q, want sha256 hex", got.Checksum)
	}
}

func TestPublishedProjectionChecksumAndOrderingAreStable(t *testing.T) {
	repo := newFakeRepository()
	checksum := strings.Repeat("a", 64)
	publishedAt := time.Date(2026, 6, 25, 10, 0, 0, 0, time.UTC)
	repo.published = schedule(3, []entity.Assignment{})
	repo.published.Revision.Status = models.MonthlyPlanScheduleStatusPublished
	repo.published.Revision.PublishedAt = &publishedAt
	repo.published.Revision.Checksum = &checksum
	repo.teams[0], repo.teams[1] = repo.teams[1], repo.teams[0]
	svc := New(repo)

	got, err := svc.GetPublished(context.Background(), 2026, 7)
	if err != nil {
		t.Fatalf("get published: %v", err)
	}
	if got.Checksum != checksum {
		t.Fatalf("checksum = %q, want persisted checksum", got.Checksum)
	}
	if got.Teams[0].DisplayOrder > got.Teams[1].DisplayOrder {
		t.Fatalf("teams not sorted by display order: %+v", got.Teams)
	}
	if got.Teams[0].Segments[0].StartDate != "2026-07-01" || got.Teams[0].Segments[0].EndDate != "2026-07-31" {
		t.Fatalf("empty team must project as full-month home: %+v", got.Teams[0].Segments)
	}
}

func TestPublishedProjectionKeepsTeamSnapshotAfterMasterDataChanges(t *testing.T) {
	repo := newFakeRepository()
	repo.draft = schedule(1, []entity.Assignment{})
	svc := New(repo)

	published, err := svc.Publish(context.Background(), superAdmin(), 2026, 7, 1)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	originalBaseArea := published.Teams[0].BaseArea
	changed := "พื้นที่ฐานใหม่"
	repo.teams[0].BaseArea = &changed

	got, err := svc.GetPublished(context.Background(), 2026, 7)
	if err != nil {
		t.Fatalf("get published: %v", err)
	}
	if got.Teams[0].BaseArea != originalBaseArea {
		t.Fatalf("published base area changed from %q to %q", originalBaseArea, got.Teams[0].BaseArea)
	}
}

type fakeRepository struct {
	period       entity.Period
	teams        []entity.Team
	draft        *entity.Schedule
	published    *entity.Schedule
	publishCalls int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		period: entity.Period{ID: 10, Year: 2026, Month: 7},
		teams: []entity.Team{
			{ID: 1, Name: "ชุด 1", Code: stringPointer("T01"), BaseArea: stringPointer("ขอนแก่น"), CrewType: stringPointer("ฮอทไลน์"), DisplayOrder: 1, MonthlyPlanVisible: true},
			{ID: 2, Name: "ชุด 2", Code: stringPointer("T02"), BaseArea: stringPointer("บ้านไผ่"), CrewType: stringPointer("ฮอทไลน์"), DisplayOrder: 2, MonthlyPlanVisible: true},
		},
	}
}

func (f *fakeRepository) FindOrCreatePeriod(_ context.Context, year, month int) (entity.Period, error) {
	f.period.Year, f.period.Month = year, month
	return f.period, nil
}

func (f *fakeRepository) FindPeriod(_ context.Context, _, _ int) (entity.Period, error) {
	return f.period, nil
}

func (f *fakeRepository) ListVisibleTeams(context.Context) ([]entity.Team, error) {
	return append([]entity.Team(nil), f.teams...), nil
}

func (f *fakeRepository) GetSchedule(_ context.Context, _ int64, status string) (*entity.Schedule, error) {
	if status == models.MonthlyPlanScheduleStatusDraft {
		if f.draft == nil {
			return nil, entity.ErrDraftNotFound
		}
		return f.draft, nil
	}
	if f.published == nil {
		return nil, entity.ErrPublishedNotFound
	}
	return f.published, nil
}

func (f *fakeRepository) ReplaceDraft(_ context.Context, period entity.Period, actor entity.Actor, _ *int, assignments []entity.Assignment) (*entity.Schedule, error) {
	f.draft = schedule(1, assignments)
	f.draft.Revision.MonthlyPlanID = period.ID
	f.draft.Revision.CreatedByUserID = actor.UserID
	return f.draft, nil
}

func (f *fakeRepository) PublishDraft(_ context.Context, _ entity.Period, actor entity.Actor, _ int64, checksum string, projection []byte) (*entity.Schedule, error) {
	f.publishCalls++
	now := time.Date(2026, 6, 25, 10, 0, 0, 0, time.UTC)
	f.published = f.draft
	f.published.Revision.Status = models.MonthlyPlanScheduleStatusPublished
	f.published.Revision.PublishedAt = &now
	f.published.Revision.PublishedByUserID = &actor.UserID
	f.published.Revision.Checksum = &checksum
	f.published.Revision.Projection = append([]byte(nil), projection...)
	return f.published, nil
}

func schedule(revisionNo int, assignments []entity.Assignment) *entity.Schedule {
	return &entity.Schedule{
		Revision:    &entity.Revision{ID: int64(revisionNo), MonthlyPlanID: 10, RevisionNo: revisionNo, Status: models.MonthlyPlanScheduleStatusDraft},
		Assignments: assignments,
	}
}

func superAdmin() entity.Actor {
	return entity.Actor{UserID: 99, Role: "super_admin"}
}

func mustDate(t *testing.T, raw string) time.Time {
	t.Helper()
	value, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		t.Fatalf("parse date %q: %v", raw, err)
	}
	return value
}

func stringPointer(value string) *string {
	return &value
}

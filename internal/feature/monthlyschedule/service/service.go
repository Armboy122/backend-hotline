package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"backend-hotlines3/internal/feature/auth/policy"
	"backend-hotlines3/internal/feature/monthlyschedule/entity"
	"backend-hotlines3/internal/feature/monthlyschedule/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetWorkspace(ctx context.Context, year, month int) (entity.Workspace, error) {
	if !validPeriod(year, month) {
		return entity.Workspace{}, entity.ErrInvalidPeriod
	}
	period, err := s.repo.FindOrCreatePeriod(ctx, year, month)
	if err != nil {
		return entity.Workspace{}, err
	}
	teams, err := s.repo.ListVisibleTeams(ctx)
	if err != nil {
		return entity.Workspace{}, err
	}
	draft, err := s.repo.GetSchedule(ctx, period.ID, entity.ScheduleStatusDraft)
	if err != nil && !errors.Is(err, entity.ErrDraftNotFound) {
		return entity.Workspace{}, err
	}
	published, err := s.repo.GetSchedule(ctx, period.ID, entity.ScheduleStatusPublished)
	if err != nil && !errors.Is(err, entity.ErrPublishedNotFound) {
		return entity.Workspace{}, err
	}
	return entity.Workspace{Period: period, Teams: teams, Draft: draft, Published: published}, nil
}

func (s *Service) SaveDraft(
	ctx context.Context,
	actor entity.Actor,
	year, month int,
	expectedRevisionNo *int,
	assignments []entity.Assignment,
) (entity.Workspace, error) {
	if actor.Role != policy.RoleSuperAdmin {
		return entity.Workspace{}, entity.ErrForbidden
	}
	if !validPeriod(year, month) {
		return entity.Workspace{}, entity.ErrInvalidPeriod
	}
	period, err := s.repo.FindOrCreatePeriod(ctx, year, month)
	if err != nil {
		return entity.Workspace{}, err
	}
	teams, err := s.repo.ListVisibleTeams(ctx)
	if err != nil {
		return entity.Workspace{}, err
	}
	normalized, err := validateAssignments(period, teams, assignments, false)
	if err != nil {
		return entity.Workspace{}, err
	}
	if _, err := s.repo.ReplaceDraft(ctx, period, actor, expectedRevisionNo, normalized); err != nil {
		return entity.Workspace{}, err
	}
	return s.GetWorkspace(ctx, year, month)
}

func (s *Service) Publish(ctx context.Context, actor entity.Actor, year, month int, expectedRevisionNo int) (entity.Projection, error) {
	if actor.Role != policy.RoleSuperAdmin {
		return entity.Projection{}, entity.ErrForbidden
	}
	if !validPeriod(year, month) {
		return entity.Projection{}, entity.ErrInvalidPeriod
	}
	period, err := s.repo.FindOrCreatePeriod(ctx, year, month)
	if err != nil {
		return entity.Projection{}, err
	}
	teams, err := s.repo.ListVisibleTeams(ctx)
	if err != nil {
		return entity.Projection{}, err
	}
	draft, err := s.repo.GetSchedule(ctx, period.ID, entity.ScheduleStatusDraft)
	if err != nil {
		return entity.Projection{}, err
	}
	if draft.Revision == nil || draft.Revision.RevisionNo != expectedRevisionNo {
		return entity.Projection{}, entity.ErrRevisionConflict
	}
	normalized, err := validateAssignments(period, teams, draft.Assignments, true)
	if err != nil {
		return entity.Projection{}, err
	}
	draft.Assignments = normalized

	projection, err := buildProjection(period, *draft.Revision, teams, normalized)
	if err != nil {
		return entity.Projection{}, err
	}
	checksum, err := projectionChecksum(projection)
	if err != nil {
		return entity.Projection{}, err
	}
	projectionSnapshot, err := json.Marshal(projection.Teams)
	if err != nil {
		return entity.Projection{}, fmt.Errorf("encode monthly schedule projection snapshot: %w", err)
	}
	published, err := s.repo.PublishDraft(ctx, period, actor, draft.Revision.ID, checksum, projectionSnapshot)
	if err != nil {
		return entity.Projection{}, err
	}
	if published.Revision == nil || published.Revision.PublishedAt == nil {
		return entity.Projection{}, fmt.Errorf("published monthly schedule is missing publication metadata")
	}
	projection.Revision = entity.ProjectionRevision{ID: published.Revision.ID, No: published.Revision.RevisionNo}
	projection.PublishedAt = *published.Revision.PublishedAt
	projection.Checksum = checksum
	return projection, nil
}

func (s *Service) GetPublished(ctx context.Context, year, month int) (entity.Projection, error) {
	if !validPeriod(year, month) {
		return entity.Projection{}, entity.ErrInvalidPeriod
	}
	period, err := s.repo.FindPeriod(ctx, year, month)
	if err != nil {
		return entity.Projection{}, err
	}
	teams, err := s.repo.ListVisibleTeams(ctx)
	if err != nil {
		return entity.Projection{}, err
	}
	published, err := s.repo.GetSchedule(ctx, period.ID, entity.ScheduleStatusPublished)
	if err != nil {
		return entity.Projection{}, err
	}
	if published.Revision == nil || published.Revision.PublishedAt == nil || published.Revision.Checksum == nil {
		return entity.Projection{}, fmt.Errorf("published monthly schedule is missing integrity metadata")
	}
	projection := entity.Projection{
		Period:   entity.ProjectionPeriod{Year: period.Year, Month: period.Month},
		Revision: entity.ProjectionRevision{ID: published.Revision.ID, No: published.Revision.RevisionNo},
	}
	if len(published.Revision.Projection) > 0 {
		if err := json.Unmarshal(published.Revision.Projection, &projection.Teams); err != nil {
			return entity.Projection{}, fmt.Errorf("decode monthly schedule projection snapshot: %w", err)
		}
	} else {
		projection, err = buildProjection(period, *published.Revision, teams, published.Assignments)
		if err != nil {
			return entity.Projection{}, err
		}
	}
	projection.PublishedAt = *published.Revision.PublishedAt
	projection.Checksum = *published.Revision.Checksum
	return projection, nil
}

func validPeriod(year, month int) bool {
	return year >= 2000 && year <= 2100 && month >= 1 && month <= 12
}

func validateAssignments(period entity.Period, teams []entity.Team, assignments []entity.Assignment, requireMetadata bool) ([]entity.Assignment, error) {
	teamByID := make(map[int64]entity.Team, len(teams))
	for _, team := range teams {
		teamByID[team.ID] = team
		if requireMetadata && (blank(team.Code) || blank(team.BaseArea) || blank(team.CrewType)) {
			return nil, fmt.Errorf("%w: team %d (%s)", entity.ErrTeamMetadataMissing, team.ID, team.Name)
		}
	}

	first := time.Date(period.Year, time.Month(period.Month), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	allowedTypes := map[string]bool{
		entity.AssignmentTypeField:   true,
		entity.AssignmentTypeRemote:  true,
		entity.AssignmentTypeSupport: true,
		entity.AssignmentTypeSpecial: true,
	}
	allowedSources := map[string]bool{
		entity.AssignmentSourceManual:       true,
		entity.AssignmentSourceLargeWork:    true,
		entity.AssignmentSourceApprovedFile: true,
	}

	out := make([]entity.Assignment, 0, len(assignments))
	for i, assignment := range assignments {
		if _, ok := teamByID[assignment.TeamID]; !ok {
			return nil, fmt.Errorf("%w: assignment %d uses a hidden or unknown team", entity.ErrInvalidAssignment, i+1)
		}
		assignment.AssignmentType = strings.TrimSpace(assignment.AssignmentType)
		assignment.Destination = strings.TrimSpace(assignment.Destination)
		assignment.StartDate = dateOnly(assignment.StartDate)
		assignment.EndDate = dateOnly(assignment.EndDate)
		if assignment.SourceType == "" {
			assignment.SourceType = entity.AssignmentSourceManual
		}
		if !allowedTypes[assignment.AssignmentType] || !allowedSources[assignment.SourceType] {
			return nil, fmt.Errorf("%w: assignment %d has an unsupported type or source", entity.ErrInvalidAssignment, i+1)
		}
		if assignment.Destination == "" {
			return nil, fmt.Errorf("%w: assignment %d destination is required", entity.ErrInvalidAssignment, i+1)
		}
		if assignment.StartDate.Before(first) || assignment.EndDate.After(last) || assignment.EndDate.Before(assignment.StartDate) {
			return nil, fmt.Errorf("%w: assignment %d date range is outside %04d-%02d", entity.ErrInvalidAssignment, i+1, period.Year, period.Month)
		}
		out = append(out, assignment)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].TeamID != out[j].TeamID {
			return out[i].TeamID < out[j].TeamID
		}
		if !out[i].StartDate.Equal(out[j].StartDate) {
			return out[i].StartDate.Before(out[j].StartDate)
		}
		return out[i].ID < out[j].ID
	})
	for i := 1; i < len(out); i++ {
		previous, current := out[i-1], out[i]
		if previous.TeamID == current.TeamID && !current.StartDate.After(previous.EndDate) {
			return nil, fmt.Errorf("%w: team %d (%s to %s)", entity.ErrOverlappingAssignment, current.TeamID, previous.StartDate.Format(time.DateOnly), current.EndDate.Format(time.DateOnly))
		}
	}
	return out, nil
}

func buildProjection(period entity.Period, revision entity.Revision, teams []entity.Team, assignments []entity.Assignment) (entity.Projection, error) {
	first := time.Date(period.Year, time.Month(period.Month), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	byTeam := make(map[int64][]entity.Assignment)
	for _, assignment := range assignments {
		byTeam[assignment.TeamID] = append(byTeam[assignment.TeamID], assignment)
	}
	sortedTeams := append([]entity.Team(nil), teams...)
	sort.SliceStable(sortedTeams, func(i, j int) bool {
		if sortedTeams[i].DisplayOrder != sortedTeams[j].DisplayOrder {
			return sortedTeams[i].DisplayOrder < sortedTeams[j].DisplayOrder
		}
		return sortedTeams[i].ID < sortedTeams[j].ID
	})
	projection := entity.Projection{
		Period:   entity.ProjectionPeriod{Year: period.Year, Month: period.Month},
		Revision: entity.ProjectionRevision{ID: revision.ID, No: revision.RevisionNo},
		Teams:    make([]entity.ProjectedTeam, 0, len(sortedTeams)),
	}
	for _, team := range sortedTeams {
		if blank(team.Code) || blank(team.BaseArea) || blank(team.CrewType) {
			return entity.Projection{}, fmt.Errorf("%w: team %d (%s)", entity.ErrTeamMetadataMissing, team.ID, team.Name)
		}
		projected := entity.ProjectedTeam{
			ID:           team.ID,
			Code:         strings.TrimSpace(*team.Code),
			Name:         team.Name,
			BaseArea:     strings.TrimSpace(*team.BaseArea),
			CrewType:     strings.TrimSpace(*team.CrewType),
			DisplayOrder: team.DisplayOrder,
			Segments:     []entity.ProjectedSegment{},
		}
		cursor := first
		for _, assignment := range byTeam[team.ID] {
			if cursor.Before(assignment.StartDate) {
				projected.Segments = append(projected.Segments, homeSegment(cursor, assignment.StartDate.AddDate(0, 0, -1), projected.BaseArea))
			}
			id := assignment.ID
			projected.Segments = append(projected.Segments, entity.ProjectedSegment{
				AssignmentID:   &id,
				AssignmentType: assignment.AssignmentType,
				StartDate:      assignment.StartDate.Format(time.DateOnly),
				EndDate:        assignment.EndDate.Format(time.DateOnly),
				Destination:    assignment.Destination,
				Note:           assignment.Note,
				SourceType:     assignment.SourceType,
				SourceID:       assignment.SourceID,
			})
			cursor = assignment.EndDate.AddDate(0, 0, 1)
		}
		if !cursor.After(last) {
			projected.Segments = append(projected.Segments, homeSegment(cursor, last, projected.BaseArea))
		}
		projection.Teams = append(projection.Teams, projected)
	}
	return projection, nil
}

func homeSegment(start, end time.Time, baseArea string) entity.ProjectedSegment {
	return entity.ProjectedSegment{
		AssignmentType: "home",
		StartDate:      start.Format(time.DateOnly),
		EndDate:        end.Format(time.DateOnly),
		Destination:    baseArea,
		SourceType:     "derived",
	}
}

func projectionChecksum(projection entity.Projection) (string, error) {
	stable := projection
	stable.PublishedAt = time.Time{}
	stable.Checksum = ""
	payload, err := json.Marshal(stable)
	if err != nil {
		return "", fmt.Errorf("encode monthly schedule checksum payload: %w", err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func blank(value *string) bool {
	return value == nil || strings.TrimSpace(*value) == ""
}

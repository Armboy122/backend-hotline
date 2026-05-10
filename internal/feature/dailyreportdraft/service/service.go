package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	largeworkentity "backend-hotlines3/internal/feature/largework/entity"
	largeworkrepository "backend-hotlines3/internal/feature/largework/repository"
	monthlyentity "backend-hotlines3/internal/feature/monthlyplan/entity"
	monthlyrepository "backend-hotlines3/internal/feature/monthlyplan/repository"
	taskentity "backend-hotlines3/internal/feature/task/entity"
	teamplanentity "backend-hotlines3/internal/feature/teamplan/entity"
	teamplanrepository "backend-hotlines3/internal/feature/teamplan/repository"
)

const (
	SourceTypeTeamPlan    = "team_plan"
	SourceTypeMonthlyPlan = "monthly_plan"
	SourceTypeLargeWork   = "large_work"
)

var (
	ErrForbidden             = errors.New("forbidden")
	ErrNotFound              = errors.New("not found")
	ErrUnsupportedSourceType = errors.New("unsupported source type")
	ErrWorkDateRequired      = errors.New("work date is required")
	ErrTeamIDRequired        = errors.New("team id is required")
)

type SourceRef struct {
	SourceType string `json:"sourceType"`
	SourceID   int64  `json:"sourceId"`
	Title      string `json:"title"`
}

type SourceCandidate struct {
	SourceRef
	WorkDate string  `json:"workDate,omitempty"`
	TeamID   *int64  `json:"teamId,omitempty"`
	TeamName *string `json:"teamName,omitempty"`
	Location *string `json:"location,omitempty"`
	Detail   *string `json:"detail,omitempty"`
}

type Prefill struct {
	WorkDate    string   `json:"workDate"`
	TeamID      *int64   `json:"teamId"`
	JobTypeID   *int64   `json:"jobTypeId"`
	JobDetailID *int64   `json:"jobDetailId"`
	FeederID    *int64   `json:"feederId"`
	NumPole     *string  `json:"numPole"`
	DeviceCode  *string  `json:"deviceCode"`
	Detail      *string  `json:"detail"`
	URLsBefore  []string `json:"urlsBefore"`
	URLsAfter   []string `json:"urlsAfter"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
}

type FromPlanInput struct {
	SourceType string
	SourceID   int64
	WorkDate   *time.Time
}

type ListSourcesInput struct {
	WorkDate *time.Time
	TeamID   *int64
}

type FromPlanOutput struct {
	Source   SourceRef `json:"source"`
	Prefill  Prefill   `json:"prefill"`
	Warnings []string  `json:"warnings"`
}

type ListSourcesOutput struct {
	Items []SourceCandidate `json:"items"`
}

type teamPlanRepo interface {
	List(ctx context.Context, query teamplanrepository.ListQuery) ([]teamplanentity.TeamPlan, int64, error)
	GetByID(ctx context.Context, query teamplanrepository.GetQuery) (*teamplanentity.TeamPlan, error)
}

type monthlyRepo interface {
	FindPeriod(ctx context.Context, q monthlyrepository.PeriodFindQuery) (*monthlyentity.Entity, error)
	ListFiles(ctx context.Context, q monthlyrepository.PlanFileListQuery) ([]monthlyentity.PlanFileEntity, error)
	GetFileByID(ctx context.Context, id int64) (*monthlyentity.PlanFileEntity, error)
}

type largeWorkRepo interface {
	List(ctx context.Context, q largeworkrepository.ListQuery) ([]largeworkentity.LargeWorkItem, int64, error)
	GetByID(ctx context.Context, q largeworkrepository.GetQuery) (*largeworkentity.LargeWorkItem, error)
}

type Service struct {
	teamPlans  teamPlanRepo
	monthly    monthlyRepo
	largeWorks largeWorkRepo
}

func NewService(teamPlans teamPlanRepo, monthly monthlyRepo, largeWorks largeWorkRepo) *Service {
	return &Service{teamPlans: teamPlans, monthly: monthly, largeWorks: largeWorks}
}

func (s *Service) ListSources(ctx context.Context, actor taskentity.Actor, input ListSourcesInput) (*ListSourcesOutput, error) {
	workDate := input.WorkDate
	if workDate == nil || workDate.IsZero() {
		return nil, ErrWorkDateRequired
	}
	teamID, err := s.resolveTeamScope(actor, input.TeamID)
	if err != nil {
		return nil, err
	}
	dayStart := dateOnly(*workDate)
	dayEnd := dayStart

	items := make([]SourceCandidate, 0, 8)
	teamPlans, _, err := s.teamPlans.List(ctx, teamplanrepository.ListQuery{
		From:   dayStart,
		To:     dayEnd,
		TeamID: teamID,
		Page:   1,
		Limit:  100,
	})
	if err != nil {
		return nil, err
	}
	for _, plan := range teamPlans {
		items = append(items, SourceCandidate{
			SourceRef: SourceRef{SourceType: SourceTypeTeamPlan, SourceID: plan.ID, Title: plan.Title},
			WorkDate:  dayStart.Format("2006-01-02"),
			TeamID:    ptrInt64(plan.TeamID),
			TeamName:  teamName(plan.Team),
			Location:  strPtr(plan.LocationText),
			Detail:    draftDetailFromTeamPlan(plan),
		})
	}

	period, err := s.monthly.FindPeriod(ctx, monthlyrepository.PeriodFindQuery{Year: dayStart.Year(), Month: int(dayStart.Month())})
	if err == nil && period != nil {
		files, err := s.monthly.ListFiles(ctx, monthlyrepository.PlanFileListQuery{MonthlyPlanID: period.ID, TeamID: teamQueryTeamID(actor, teamID), Search: ""})
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			items = append(items, SourceCandidate{
				SourceRef: SourceRef{SourceType: SourceTypeMonthlyPlan, SourceID: file.ID, Title: monthlyTitle(file)},
				WorkDate:  chooseDate(file.WorkStartDate, dayStart),
				TeamID:    file.TeamID,
				TeamName:  file.TeamName,
				Detail:    draftDetailFromMonthlyFile(file),
			})
		}
	} else if err != nil && !errors.Is(err, monthlyentity.ErrPeriodNotFound) {
		return nil, err
	}

	if s.largeWorks != nil {
		largeWorks, _, err := s.largeWorks.List(ctx, largeworkrepository.ListQuery{
			From:    &dayStart,
			To:      &dayEnd,
			TeamID:  teamID,
			Statuses: []string{largeworkentity.LargeWorkStatusPlanned, largeworkentity.LargeWorkStatusInProgress, largeworkentity.LargeWorkStatusCompleted},
		})
		if err != nil {
			return nil, err
		}
		for _, work := range largeWorks {
			items = append(items, SourceCandidate{
				SourceRef: SourceRef{SourceType: SourceTypeLargeWork, SourceID: work.ID, Title: work.Title},
				WorkDate:  chooseDate(&work.StartDate, dayStart),
				TeamID:    ptrInt64(work.OwnerTeamID),
				TeamName:  largeWorkOwnerTeamName(work.Teams),
				Detail:    draftDetailFromLargeWork(work),
			})
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].SourceType != items[j].SourceType {
			return items[i].SourceType < items[j].SourceType
		}
		if items[i].Title != items[j].Title {
			return items[i].Title < items[j].Title
		}
		return items[i].SourceID < items[j].SourceID
	})
	return &ListSourcesOutput{Items: items}, nil
}

func (s *Service) FromPlan(ctx context.Context, actor taskentity.Actor, input FromPlanInput) (*FromPlanOutput, error) {
	if input.SourceID <= 0 {
		return nil, ErrNotFound
	}
	switch input.SourceType {
	case SourceTypeTeamPlan:
		return s.fromTeamPlan(ctx, actor, input)
	case SourceTypeMonthlyPlan:
		return s.fromMonthlyPlan(ctx, actor, input)
	case SourceTypeLargeWork:
		return s.fromLargeWork(ctx, actor, input)
	default:
		return nil, ErrUnsupportedSourceType
	}
}

func (s *Service) fromTeamPlan(ctx context.Context, actor taskentity.Actor, input FromPlanInput) (*FromPlanOutput, error) {
	plan, err := s.teamPlans.GetByID(ctx, teamplanrepository.GetQuery{ID: input.SourceID})
	if err != nil || plan == nil {
		return nil, ErrNotFound
	}
	if !actor.CanReadTeam(plan.TeamID) {
		return nil, ErrForbidden
	}
	workDate := input.WorkDate
	if workDate == nil || workDate.IsZero() {
		workDate = &plan.StartDate
	}
	prefill := Prefill{
		WorkDate:    dateOnly(*workDate).Format("2006-01-02"),
		TeamID:      ptrInt64(plan.TeamID),
		JobTypeID:   nil,
		JobDetailID: nil,
		FeederID:    plan.FeederID,
		NumPole:     nil,
		DeviceCode:  nil,
		Detail:      stringPtr(draftDetailFromTeamPlan(*plan)),
		URLsBefore:  []string{},
		URLsAfter:   []string{},
		Latitude:    nil,
		Longitude:   nil,
	}
	return &FromPlanOutput{
		Source:   SourceRef{SourceType: SourceTypeTeamPlan, SourceID: plan.ID, Title: plan.Title},
		Prefill:  prefill,
		Warnings: []string{"jobTypeId and jobDetailId must be selected before saving"},
	}, nil
}

func (s *Service) fromMonthlyPlan(ctx context.Context, actor taskentity.Actor, input FromPlanInput) (*FromPlanOutput, error) {
	file, err := s.monthly.GetFileByID(ctx, input.SourceID)
	if err != nil || file == nil {
		return nil, ErrNotFound
	}
	monthlyActor := monthlyentity.Actor{UserID: actor.UserID, Role: actor.Role, TeamID: actor.TeamID}
	if !file.IsMasterPlan && !monthlyActor.CanAccessFile(file.TeamID) {
		return nil, ErrForbidden
	}
	workDate := input.WorkDate
	if workDate == nil || workDate.IsZero() {
		if file.WorkStartDate != nil {
			workDate = file.WorkStartDate
		} else {
			return nil, ErrWorkDateRequired
		}
	}
	teamID := file.TeamID
	if teamID == nil {
		teamID = actor.TeamID
	}
	prefill := Prefill{
		WorkDate:    dateOnly(*workDate).Format("2006-01-02"),
		TeamID:      teamID,
		JobTypeID:   nil,
		JobDetailID: nil,
		FeederID:    nil,
		NumPole:     nil,
		DeviceCode:  nil,
		Detail:      stringPtr(draftDetailFromMonthlyFile(*file)),
		URLsBefore:  []string{},
		URLsAfter:   []string{},
		Latitude:    nil,
		Longitude:   nil,
	}
	return &FromPlanOutput{
		Source:   SourceRef{SourceType: SourceTypeMonthlyPlan, SourceID: file.ID, Title: monthlyTitle(*file)},
		Prefill:  prefill,
		Warnings: []string{"jobTypeId and jobDetailId must be selected before saving"},
	}, nil
}

func (s *Service) fromLargeWork(ctx context.Context, actor taskentity.Actor, input FromPlanInput) (*FromPlanOutput, error) {
	if s.largeWorks == nil {
		return nil, ErrUnsupportedSourceType
	}
	item, err := s.largeWorks.GetByID(ctx, largeworkrepository.GetQuery{ID: input.SourceID})
	if err != nil || item == nil {
		return nil, ErrNotFound
	}
	lwActor := largeworkentity.Actor{UserID: actor.UserID, Role: actor.Role, TeamID: actor.TeamID}
	if !lwActor.CanViewLargeWork(item) {
		return nil, ErrForbidden
	}
	workDate := input.WorkDate
	if workDate == nil || workDate.IsZero() {
		workDate = &item.StartDate
	}
	teamID := ptrInt64(item.OwnerTeamID)
	prefill := Prefill{
		WorkDate:    dateOnly(*workDate).Format("2006-01-02"),
		TeamID:      teamID,
		JobTypeID:   nil,
		JobDetailID: nil,
		FeederID:    item.FeederID,
		NumPole:     nil,
		DeviceCode:  nil,
		Detail:      stringPtr(draftDetailFromLargeWork(*item)),
		URLsBefore:  []string{},
		URLsAfter:   []string{},
		Latitude:    nil,
		Longitude:   nil,
	}
	return &FromPlanOutput{
		Source:   SourceRef{SourceType: SourceTypeLargeWork, SourceID: item.ID, Title: item.Title},
		Prefill:  prefill,
		Warnings: []string{"jobTypeId and jobDetailId must be selected before saving"},
	}, nil
}

func (s *Service) resolveTeamScope(actor taskentity.Actor, requested *int64) (*int64, error) {
	if actor.IsPrivileged() {
		if requested != nil {
			return requested, nil
		}
		return actor.TeamID, nil
	}
	if requested != nil && actor.TeamID != nil && *requested != *actor.TeamID {
		return nil, ErrForbidden
	}
	if requested != nil {
		return requested, nil
	}
	if actor.TeamID == nil {
		return nil, ErrTeamIDRequired
	}
	return actor.TeamID, nil
}

func teamQueryTeamID(actor taskentity.Actor, teamID *int64) int64 {
	if actor.IsPrivileged() && teamID == nil {
		return 0
	}
	if teamID == nil {
		return 0
	}
	return *teamID
}

func draftDetailFromTeamPlan(plan teamplanentity.TeamPlan) *string {
	parts := make([]string, 0, 3)
	if title := strings.TrimSpace(plan.Title); title != "" {
		parts = append(parts, title)
	}
	if location := strings.TrimSpace(plan.LocationText); location != "" {
		parts = append(parts, location)
	}
	if plan.Notes != nil {
		if note := strings.TrimSpace(*plan.Notes); note != "" {
			parts = append(parts, note)
		}
	}
	if len(parts) == 0 {
		return nil
	}
	joined := strings.Join(parts, "\n")
	return &joined
}

func draftDetailFromMonthlyFile(file monthlyentity.PlanFileEntity) *string {
	parts := make([]string, 0, 3)
	if file.Description != nil {
		if v := strings.TrimSpace(*file.Description); v != "" {
			parts = append(parts, v)
		}
	}
	if file.Destination != nil {
		if v := strings.TrimSpace(*file.Destination); v != "" {
			parts = append(parts, v)
		}
	}
	if file.Remarks != nil {
		if v := strings.TrimSpace(*file.Remarks); v != "" {
			parts = append(parts, v)
		}
	}
	if len(parts) == 0 {
		return nil
	}
	joined := strings.Join(parts, "\n")
	return &joined
}

func draftDetailFromLargeWork(item largeworkentity.LargeWorkItem) *string {
	parts := make([]string, 0, 4)
	if title := strings.TrimSpace(item.Title); title != "" {
		parts = append(parts, title)
	}
	if item.WorkType != nil {
		if v := strings.TrimSpace(*item.WorkType); v != "" {
			parts = append(parts, v)
		}
	}
	if location := strings.TrimSpace(item.LocationText); location != "" {
		parts = append(parts, location)
	}
	if item.Notes != nil {
		if v := strings.TrimSpace(*item.Notes); v != "" {
			parts = append(parts, v)
		}
	}
	if len(parts) == 0 {
		return nil
	}
	joined := strings.Join(parts, "\n")
	return &joined
}

func largeWorkOwnerTeamName(teams []largeworkentity.LargeWorkTeam) *string {
	for _, team := range teams {
		if strings.TrimSpace(team.Role) == largeworkentity.LargeWorkTeamRoleOwner && strings.TrimSpace(team.Name) != "" {
			name := strings.TrimSpace(team.Name)
			return &name
		}
	}
	for _, team := range teams {
		if strings.TrimSpace(team.Name) != "" {
			name := strings.TrimSpace(team.Name)
			return &name
		}
	}
	return nil
}

func monthlyTitle(file monthlyentity.PlanFileEntity) string {
	if file.Description != nil && strings.TrimSpace(*file.Description) != "" {
		return strings.TrimSpace(*file.Description)
	}
	if strings.TrimSpace(file.FileName) != "" {
		return strings.TrimSpace(file.FileName)
	}
	return "monthly plan"
}

func chooseDate(candidate *time.Time, fallback time.Time) string {
	if candidate == nil || candidate.IsZero() {
		return dateOnly(fallback).Format("2006-01-02")
	}
	return dateOnly(*candidate).Format("2006-01-02")
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func teamName(team *teamplanentity.TeamSummary) *string {
	if team == nil {
		return nil
	}
	return &team.Name
}

func ptrInt64(v int64) *int64 { return &v }
func strPtr(v string) *string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	s := strings.TrimSpace(v)
	return &s
}
func stringPtr(v *string) *string { return v }

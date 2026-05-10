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
	teamplanentity "backend-hotlines3/internal/feature/teamplan/entity"
	teamplanrepository "backend-hotlines3/internal/feature/teamplan/repository"
)

const (
	dateLayout          = "2006-01-02"
	SourceTypeLargeWork = "large_work"
)

var ErrInvalidPeriod = errors.New("invalid planning period")

type Actor struct {
	UserID int64
	Role   string
	TeamID *int64
}

func (a Actor) toTeamPlanActor() teamplanentity.Actor {
	return teamplanentity.Actor{UserID: a.UserID, Role: a.Role, TeamID: a.TeamID}
}

func (a Actor) toMonthlyPlanActor() monthlyentity.Actor {
	return monthlyentity.Actor{UserID: a.UserID, Role: a.Role, TeamID: a.TeamID}
}

type TeamPlanReader interface {
	ListAll(ctx context.Context, query teamplanrepository.ListQuery) ([]teamplanentity.TeamPlan, error)
}

type MonthlyPlanReader interface {
	FindPeriod(ctx context.Context, q monthlyrepository.PeriodFindQuery) (*monthlyentity.Entity, error)
	GetOrCreateSettings(ctx context.Context) (*monthlyentity.SettingsEntity, error)
	ListFiles(ctx context.Context, q monthlyrepository.PlanFileListQuery) ([]monthlyentity.PlanFileEntity, error)
}

type LargeWorkReader interface {
	List(ctx context.Context, q largeworkrepository.ListQuery) ([]largeworkentity.LargeWorkItem, int64, error)
}

type Service struct {
	teamPlans    TeamPlanReader
	monthlyPlans MonthlyPlanReader
	largeWorks   LargeWorkReader
	clock        func() time.Time
}

type MonthProjection struct {
	Year       int           `json:"year"`
	Month      int           `json:"month"`
	MonthStart string        `json:"monthStart"`
	MonthEnd   string        `json:"monthEnd"`
	Days       []CalendarDay `json:"days"`
}

type CalendarDay struct {
	Date  string         `json:"date"`
	Items []CalendarItem `json:"items"`
}

type CalendarItem struct {
	SourceType string      `json:"sourceType"`
	SourceID   int64       `json:"sourceId"`
	Title      string      `json:"title"`
	Team       *TeamRef    `json:"team,omitempty"`
	Teams      []TeamRef   `json:"teams,omitempty"`
	WorkType   *string     `json:"workType,omitempty"`
	WorkTime   *string     `json:"workTime,omitempty"`
	Location   string      `json:"location"`
	Status     string      `json:"status,omitempty"`
	DateRange  DateRange   `json:"dateRange"`
	Actions    ItemActions `json:"actions"`
	FileName   string      `json:"fileName,omitempty"`
	Notes      *string     `json:"notes,omitempty"`
}

type TeamRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type DateRange struct {
	StartDate string  `json:"startDate"`
	EndDate   *string `json:"endDate,omitempty"`
}

type ItemActions struct {
	CanEdit     bool `json:"canEdit"`
	CanDelete   bool `json:"canDelete"`
	CanDownload bool `json:"canDownload"`
}

func NewService(teamPlans TeamPlanReader, monthlyPlans MonthlyPlanReader, largeWorks LargeWorkReader, clock func() time.Time) *Service {
	if clock == nil {
		clock = time.Now
	}
	return &Service{teamPlans: teamPlans, monthlyPlans: monthlyPlans, largeWorks: largeWorks, clock: clock}
}

func (s *Service) GetMonth(ctx context.Context, actor Actor, year, month int) (*MonthProjection, error) {
	if year < 1 || month < 1 || month > 12 {
		return nil, ErrInvalidPeriod
	}

	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, -1)

	teamPlans, err := s.teamPlans.ListAll(ctx, teamplanrepository.ListQuery{From: monthStart, To: monthEnd})
	if err != nil {
		return nil, err
	}

	days := make([]CalendarDay, 0, monthEnd.Day())
	dayIndex := make(map[string]int, monthEnd.Day())
	for d := monthStart; !d.After(monthEnd); d = d.AddDate(0, 0, 1) {
		key := d.Format(dateLayout)
		dayIndex[key] = len(days)
		days = append(days, CalendarDay{Date: key, Items: []CalendarItem{}})
	}

	for _, plan := range teamPlans {
		item := buildTeamPlanItem(actor.toTeamPlanActor(), plan)
		for _, date := range plan.ExpandedDates() {
			if date.Before(monthStart) || date.After(monthEnd) {
				continue
			}
			key := date.Format(dateLayout)
			idx, ok := dayIndex[key]
			if !ok {
				continue
			}
			days[idx].Items = append(days[idx].Items, item)
		}
	}

	period, err := s.monthlyPlans.FindPeriod(ctx, monthlyrepository.PeriodFindQuery{Year: year, Month: month})
	if err != nil {
		return nil, err
	}
	if period != nil {
		files, err := s.monthlyPlans.ListFiles(ctx, monthlyrepository.PlanFileListQuery{MonthlyPlanID: period.ID})
		if err != nil {
			return nil, err
		}
		if len(files) > 0 {
			settings, err := s.monthlyPlans.GetOrCreateSettings(ctx)
			if err != nil {
				return nil, err
			}
			monthlyActor := actor.toMonthlyPlanActor()
			for _, file := range files {
				item := buildMonthlyPlanItem(monthlyActor, file, period, settings, s.clock)
				for _, date := range expandDateRange(file.WorkStartDate, file.WorkEndDate, monthStart, monthEnd) {
					idx, ok := dayIndex[date.Format(dateLayout)]
					if !ok {
						continue
					}
					days[idx].Items = append(days[idx].Items, item)
				}
			}
		}
	}

	if s.largeWorks != nil {
		largeWorks, _, err := s.largeWorks.List(ctx, largeworkrepository.ListQuery{From: &monthStart, To: &monthEnd, TeamID: actor.TeamID, Statuses: []string{largeworkentity.LargeWorkStatusPlanned, largeworkentity.LargeWorkStatusInProgress, largeworkentity.LargeWorkStatusCompleted}})
		if err != nil {
			return nil, err
		}
		for _, work := range largeWorks {
			item := buildLargeWorkItem(actor.toLargeWorkActor(), work)
			for _, date := range expandDateRange(&work.StartDate, work.EndDate, monthStart, monthEnd) {
				idx, ok := dayIndex[date.Format(dateLayout)]
				if !ok {
					continue
				}
				days[idx].Items = append(days[idx].Items, item)
			}
		}
	}

	for i := range days {
		sort.SliceStable(days[i].Items, func(a, b int) bool {
			if days[i].Items[a].SourceType != days[i].Items[b].SourceType {
				return days[i].Items[a].SourceType < days[i].Items[b].SourceType
			}
			return days[i].Items[a].SourceID < days[i].Items[b].SourceID
		})
	}

	return &MonthProjection{
		Year:       year,
		Month:      month,
		MonthStart: monthStart.Format(dateLayout),
		MonthEnd:   monthEnd.Format(dateLayout),
		Days:       days,
	}, nil
}

func buildTeamPlanItem(actor teamplanentity.Actor, plan teamplanentity.TeamPlan) CalendarItem {
	item := CalendarItem{
		SourceType: "team_plan",
		SourceID:   plan.ID,
		Title:      plan.Title,
		Location:   plan.LocationText,
		WorkType:   plan.WorkType,
		WorkTime:   plan.WorkTime,
		Status:     plan.Status,
		DateRange:  DateRange{StartDate: plan.StartDate.Format(dateLayout)},
		Actions: ItemActions{
			CanEdit:   actor.CanUpdateTeamPlan(plan),
			CanDelete: actor.CanDeleteTeamPlan(plan),
		},
	}
	if plan.Team != nil {
		item.Team = &TeamRef{ID: plan.Team.ID, Name: plan.Team.Name}
	}
	if plan.EndDate != nil && !plan.EndDate.IsZero() {
		end := plan.EndDate.Format(dateLayout)
		item.DateRange.EndDate = &end
	} else {
		end := plan.StartDate.Format(dateLayout)
		item.DateRange.EndDate = &end
	}
	return item
}

func buildMonthlyPlanItem(actor monthlyentity.Actor, file monthlyentity.PlanFileEntity, period *monthlyentity.Entity, settings *monthlyentity.SettingsEntity, clock func() time.Time) CalendarItem {
	monthStart := time.Date(period.Year, time.Month(period.Month), 1, 0, 0, 0, 0, time.UTC)
	item := CalendarItem{
		SourceType: "monthly_plan",
		SourceID:   file.ID,
		Title:      pickTitle(file.Description, file.FileName),
		Location:   pickText(file.Destination),
		Status:     "submitted",
		DateRange:  DateRange{StartDate: monthStart.Format(dateLayout)},
		Actions:    ItemActions{},
		FileName:   file.FileName,
		Notes:      file.Remarks,
	}
	if file.TeamID != nil {
		item.Team = &TeamRef{ID: *file.TeamID, Name: pickTextPtr(file.TeamName)}
	}
	start := pickTime(file.WorkStartDate, monthStart)
	end := pickTime(file.WorkEndDate, start)
	if end.Before(start) {
		end = start
	}
	item.DateRange.StartDate = start.Format(dateLayout)
	endText := end.Format(dateLayout)
	item.DateRange.EndDate = &endText

	locked := clock().After(monthlyLockDeadline(period.Year, period.Month, settings.LockDay))
	canUpload := true
	if locked {
		canUpload = actor.CanUploadAfterLock(settings.AdminCanUploadAfterLock)
	}
	canEdit := false
	if file.IsMasterPlan {
		canEdit = actor.CanUploadMasterPlan() && canUpload
		item.Actions.CanDownload = actor.IsAdmin()
	} else {
		canEdit = actor.CanUploadForTeam(file.TeamID) && canUpload
		item.Actions.CanDownload = actor.CanAccessFile(file.TeamID)
	}
	item.Actions.CanEdit = canEdit
	item.Actions.CanDelete = false
	return item
}

func buildLargeWorkItem(actor largeworkentity.Actor, work largeworkentity.LargeWorkItem) CalendarItem {
	item := CalendarItem{
		SourceType: SourceTypeLargeWork,
		SourceID:   work.ID,
		Title:      work.Title,
		Location:   work.LocationText,
		WorkType:   work.WorkType,
		WorkTime:   work.WorkTime,
		Status:     work.Status,
		DateRange:  DateRange{StartDate: work.StartDate.Format(dateLayout)},
		Actions: ItemActions{
			CanEdit:   true,
			CanDelete: true,
		},
	}
	if work.EndDate != nil && !work.EndDate.IsZero() {
		end := work.EndDate.Format(dateLayout)
		item.DateRange.EndDate = &end
	} else {
		end := work.StartDate.Format(dateLayout)
		item.DateRange.EndDate = &end
	}
	if len(work.Teams) > 0 {
		item.Team = &TeamRef{ID: work.OwnerTeamID, Name: largeWorkOwnerTeamName(work.Teams)}
		item.Teams = make([]TeamRef, 0, len(work.Teams))
		for _, team := range work.Teams {
			item.Teams = append(item.Teams, TeamRef{ID: team.ID, Name: team.Name})
		}
	}
	item.Actions.CanEdit = actor.CanUpdateLargeWork(&work)
	item.Actions.CanDelete = actor.CanCancelLargeWork(&work)
	return item
}

func (a Actor) toLargeWorkActor() largeworkentity.Actor {
	return largeworkentity.Actor{UserID: a.UserID, Role: a.Role, TeamID: a.TeamID}
}

func largeWorkOwnerTeamName(teams []largeworkentity.LargeWorkTeam) string {
	for _, team := range teams {
		if strings.TrimSpace(team.Role) == largeworkentity.LargeWorkTeamRoleOwner && strings.TrimSpace(team.Name) != "" {
			return strings.TrimSpace(team.Name)
		}
	}
	for _, team := range teams {
		if strings.TrimSpace(team.Name) != "" {
			return strings.TrimSpace(team.Name)
		}
	}
	return ""
}

func monthlyLockDeadline(year, month, lockDay int) time.Time {
	if lockDay < 1 {
		lockDay = 1
	}
	planMonthStart := time.Date(year, time.Month(month), 1, 23, 59, 59, 0, time.UTC)
	deadlineMonthStart := planMonthStart.AddDate(0, -1, 0)
	lastDay := deadlineMonthStart.AddDate(0, 1, -1).Day()
	if lockDay > lastDay {
		lockDay = lastDay
	}
	return time.Date(deadlineMonthStart.Year(), deadlineMonthStart.Month(), lockDay, 23, 59, 59, 0, time.UTC)
}

func expandDateRange(start, end *time.Time, monthStart, monthEnd time.Time) []time.Time {
	if start == nil {
		fallback := monthStart
		start = &fallback
	}
	current := pickTime(start, monthStart)
	finish := pickTime(end, current)
	if finish.Before(current) {
		finish = current
	}
	out := make([]time.Time, 0, int(finish.Sub(current)/(24*time.Hour))+1)
	for d := current; !d.After(finish); d = d.AddDate(0, 0, 1) {
		if d.Before(monthStart) || d.After(monthEnd) {
			continue
		}
		out = append(out, d)
	}
	return out
}

func pickTime(raw *time.Time, fallback time.Time) time.Time {
	if raw == nil || raw.IsZero() {
		return fallback
	}
	return time.Date(raw.Year(), raw.Month(), raw.Day(), 0, 0, 0, 0, time.UTC)
}

func pickText(raw *string) string {
	if raw == nil {
		return ""
	}
	return strings.TrimSpace(*raw)
}

func pickTextPtr(raw *string) string {
	if raw == nil {
		return ""
	}
	return strings.TrimSpace(*raw)
}

func pickTitle(description *string, fallback string) string {
	desc := pickText(description)
	if desc != "" {
		return desc
	}
	if fallback != "" {
		return fallback
	}
	return "monthly plan"
}

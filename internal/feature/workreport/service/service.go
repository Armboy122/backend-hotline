package service

import (
	"context"
	"errors"
	"sort"

	"backend-hotlines3/internal/feature/auth/policy"
	taskentity "backend-hotlines3/internal/feature/task/entity"
	taskrepository "backend-hotlines3/internal/feature/task/repository"
)

var (
	ErrYearMonthRequired = errors.New("year and month are required")
	ErrForbidden         = errors.New("forbidden")
)

// ReportItem is a single task entry in the work report.
type ReportItem struct {
	ID                  int64    `json:"id"`
	WorkDate            string   `json:"workDate"`
	TeamID              int64    `json:"teamId"`
	TeamName            *string  `json:"teamName,omitempty"`
	JobTypeID           int64    `json:"jobTypeId"`
	JobTypeName         *string  `json:"jobTypeName,omitempty"`
	JobDetailID         int64    `json:"jobDetailId"`
	JobDetailName       *string  `json:"jobDetailName,omitempty"`
	FeederID            *int64   `json:"feederId,omitempty"`
	FeederCode          *string  `json:"feederCode,omitempty"`
	StationName         *string  `json:"stationName,omitempty"`
	OperationCenterID   *int64   `json:"operationCenterId,omitempty"`
	OperationCenterName *string  `json:"operationCenterName,omitempty"`
	NumPole             *string  `json:"numPole,omitempty"`
	DeviceCode          *string  `json:"deviceCode,omitempty"`
	Detail              *string  `json:"detail,omitempty"`
	URLsBefore          []string `json:"urlsBefore,omitempty"`
	URLsAfter           []string `json:"urlsAfter,omitempty"`
	Latitude            *float64 `json:"latitude,omitempty"`
	Longitude           *float64 `json:"longitude,omitempty"`
	CreatedAt           string   `json:"createdAt"`
	UpdatedAt           string   `json:"updatedAt"`
}

// TeamSummary holds per-team task counts for the report.
type TeamSummary struct {
	TeamID   int64  `json:"teamId"`
	TeamName string `json:"teamName"`
	Count    int64  `json:"count"`
}

// Summary contains aggregate statistics for the work report.
type Summary struct {
	TotalTasks      int64         `json:"totalTasks"`
	UniqueTeamCount int64         `json:"uniqueTeamCount"`
	TeamSummaries   []TeamSummary `json:"teamSummaries,omitempty"`
	// Future fields (need schema migration):
	// TotalFromPlanning   int64 `json:"totalFromPlanning"`
	// TotalFromMonthlyPlan int64 `json:"totalFromMonthlyPlan"`
	// TotalOutsidePlan    int64 `json:"totalOutsidePlan"`
}

// GetReportInput is the query input for the work report.
type GetReportInput struct {
	Year   string
	Month  string
	TeamID *int64
}

// GetReportOutput is the response payload.
type GetReportOutput struct {
	Summary Summary      `json:"summary"`
	Items   []ReportItem `json:"items"`
}

type taskLister interface {
	ListByFilter(ctx context.Context, query taskrepository.ByFilterQuery) ([]taskentity.Task, error)
}

// Service provides work report aggregation over TaskDaily.
type Service struct {
	tasks taskLister
}

// NewService creates a new work report service.
func NewService(tasks taskLister) *Service {
	return &Service{tasks: tasks}
}

// GetReport returns a monthly work report with summary statistics.
func (s *Service) GetReport(ctx context.Context, actor taskentity.Actor, input GetReportInput) (*GetReportOutput, error) {
	if input.Year == "" || input.Month == "" {
		return nil, ErrYearMonthRequired
	}

	// Scope team to actor's own team for non-privileged users. Work reports allow
	// both super_admin and admin to read all teams; team_lead/user/viewer remain
	// own-team scoped.
	teamID := input.TeamID
	if actor.Role != policy.RoleSuperAdmin && actor.Role != policy.RoleAdmin {
		teamID = actor.ScopedTeamID(input.TeamID)
	}

	tasks, err := s.tasks.ListByFilter(ctx, taskrepository.ByFilterQuery{
		Year:   input.Year,
		Month:  input.Month,
		TeamID: teamID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]ReportItem, 0, len(tasks))
	teamCounts := make(map[int64]*TeamSummary)

	for _, t := range tasks {
		item := ReportItem{
			ID:                  t.ID,
			WorkDate:            t.WorkDate.Format("2006-01-02"),
			TeamID:              t.TeamID,
			TeamName:            t.TeamName,
			JobTypeID:           t.JobTypeID,
			JobTypeName:         t.JobTypeName,
			JobDetailID:         t.JobDetailID,
			JobDetailName:       t.JobDetailName,
			FeederID:            t.FeederID,
			FeederCode:          t.FeederCode,
			StationName:         t.StationName,
			OperationCenterID:   t.OperationCenterID,
			OperationCenterName: t.OperationCenterName,
			NumPole:             t.NumPole,
			DeviceCode:          t.DeviceCode,
			Detail:              t.Detail,
			URLsBefore:          t.URLsBefore,
			URLsAfter:           t.URLsAfter,
			Latitude:            t.Latitude,
			Longitude:           t.Longitude,
			CreatedAt:           t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:           t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if t.URLsBefore == nil {
			item.URLsBefore = []string{}
		}
		if t.URLsAfter == nil {
			item.URLsAfter = []string{}
		}
		items = append(items, item)

		// Accumulate team counts.
		name := ""
		if t.TeamName != nil {
			name = *t.TeamName
		}
		if ts, ok := teamCounts[t.TeamID]; ok {
			ts.Count++
		} else {
			teamCounts[t.TeamID] = &TeamSummary{
				TeamID:   t.TeamID,
				TeamName: name,
				Count:    1,
			}
		}
	}

	teamSums := make([]TeamSummary, 0, len(teamCounts))
	for _, ts := range teamCounts {
		teamSums = append(teamSums, *ts)
	}
	sort.Slice(teamSums, func(i, j int) bool { return teamSums[i].TeamID < teamSums[j].TeamID })

	return &GetReportOutput{
		Summary: Summary{
			TotalTasks:      int64(len(tasks)),
			UniqueTeamCount: int64(len(teamCounts)),
			TeamSummaries:   teamSums,
		},
		Items: items,
	}, nil
}

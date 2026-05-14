package entity

import (
	"errors"
	"strings"
	"time"

	"backend-hotlines3/internal/feature/auth/policy"
)

var (
	ErrNotFound      = errors.New("team plan not found")
	ErrInvalidID     = errors.New("invalid team plan ID")
	ErrForbidden     = errors.New("forbidden")
	ErrInvalidInput   = errors.New("invalid team plan input")
	ErrInvalidStatus  = errors.New("invalid team plan status")
	ErrInvalidRange   = errors.New("invalid team plan date range")
)

const (
	StatusPlanned   = "planned"
	StatusCancelled = "cancelled"
	StatusCompleted = "completed"
)

var allowedStatuses = map[string]struct{}{
	StatusPlanned:   {},
	StatusCancelled: {},
	StatusCompleted: {},
}

type Actor struct {
	UserID int64
	Role   string
	TeamID *int64
}

type TeamSummary struct {
	ID   int64
	Name string
}

type UserSummary struct {
	ID          int64
	Username    string
	DisplayName *string
}

type TeamPlan struct {
	ID                int64
	TeamID            int64
	CreatedByUserID   int64
	Title             string
	WorkType          *string
	StartDate         time.Time
	EndDate           *time.Time
	WorkTime          *string
	LocationText      string
	PEAID             *int64
	OperationCenterID *int64
	FeederID          *int64
	StationID         *int64
	Notes             *string
	Status            string
	DailyTaskID       *int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	Team              *TeamSummary
	CreatedBy         *UserSummary
}

func (a Actor) CanCreateTeamPlan(teamID int64) bool {
	if a.Role == policy.RoleSuperAdmin {
		return true
	}
	if teamID <= 0 || a.TeamID == nil {
		return false
	}
	switch a.Role {
	case policy.RoleAdmin, policy.RoleTeamLead, policy.RoleUser:
		return *a.TeamID == teamID
	default:
		return false
	}
}

func (a Actor) CanUpdateTeamPlan(plan TeamPlan) bool {
	if a.Role == policy.RoleSuperAdmin {
		return true
	}
	if a.TeamID == nil || *a.TeamID != plan.TeamID {
		return false
	}
	switch a.Role {
	case policy.RoleAdmin, policy.RoleTeamLead:
		return true
	case policy.RoleUser:
		return plan.CreatedByUserID == a.UserID && plan.Status == StatusPlanned
	default:
		return false
	}
}

func (a Actor) CanDeleteTeamPlan(plan TeamPlan) bool {
	if a.Role == policy.RoleSuperAdmin {
		return true
	}
	if a.TeamID == nil || *a.TeamID != plan.TeamID {
		return false
	}
	switch a.Role {
	case policy.RoleAdmin, policy.RoleTeamLead:
		return true
	default:
		return false
	}
}

func (p TeamPlan) ExpandedDates() []time.Time {
	start := dateOnly(p.StartDate)
	end := start
	if p.EndDate != nil && !p.EndDate.IsZero() {
		end = dateOnly(*p.EndDate)
	}
	if end.Before(start) {
		end = start
	}

	dates := make([]time.Time, 0, int(end.Sub(start)/(24*time.Hour))+1)
	for current := start; !current.After(end); current = current.Add(24 * time.Hour) {
		dates = append(dates, current)
	}
	return dates
}

func NormalizeStatus(status string) (string, bool) {
	if status == "" {
		return StatusPlanned, true
	}
	_, ok := allowedStatuses[strings.TrimSpace(status)]
	if !ok {
		return "", false
	}
	return strings.TrimSpace(status), true
}

func IsValidStatus(status string) bool {
	_, ok := allowedStatuses[status]
	return ok
}

func dateOnly(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

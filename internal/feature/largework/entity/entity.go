package entity

import "time"

type Actor struct {
	UserID int64
	Role   string
	TeamID *int64
}

type LargeWorkItem struct {
	ID                int64
	OwnerTeamID       int64
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
	Teams             []LargeWorkTeam
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

type LargeWorkTeam struct {
	ID                int64
	Name              string
	Role              string
	ParticipantStatus string
}

const (
	LargeWorkStatusDraft      = "draft"
	LargeWorkStatusPlanned    = "planned"
	LargeWorkStatusInProgress = "in_progress"
	LargeWorkStatusCompleted  = "completed"
	LargeWorkStatusCancelled  = "cancelled"

	LargeWorkTeamRoleOwner       = "owner"
	LargeWorkTeamRoleParticipant = "participant"

	LargeWorkParticipantStatusAssigned     = "assigned"
	LargeWorkParticipantStatusAcknowledged = "acknowledged"
)

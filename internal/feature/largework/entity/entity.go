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

type LargeWorkTask struct {
	ID                int64
	LargeWorkItemID   int64
	AssignedTeamID    int64
	Sequence          int
	PointLabel        string
	Latitude          *float64
	Longitude         *float64
	WorkType          string
	WorkDetail        *string
	PointCount        *int
	TreeCount         *int
	ItemCount         *int
	Notes             *string
	Status            string
	BeforePhotoURLs   []string
	AfterPhotoURLs    []string
	CompletionNote    *string
	StartedByUserID   *int64
	StartedAt         *time.Time
	CompletedByUserID *int64
	CompletedAt       *time.Time
	Metadata          map[string]any
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
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

	LargeWorkTaskStatusTodo       = "todo"
	LargeWorkTaskStatusInProgress = "in_progress"
	LargeWorkTaskStatusDone       = "done"
	LargeWorkTaskStatusBlocked    = "blocked"
	LargeWorkTaskStatusCancelled  = "cancelled"
)

package entity

import (
	"errors"
	"time"
)

var (
	ErrInvalidPeriod         = errors.New("invalid monthly schedule period")
	ErrForbidden             = errors.New("forbidden: monthly schedule requires super_admin")
	ErrDraftNotFound         = errors.New("monthly schedule draft not found")
	ErrPublishedNotFound     = errors.New("published monthly schedule not found")
	ErrRevisionConflict      = errors.New("monthly schedule revision conflict")
	ErrInvalidAssignment     = errors.New("invalid monthly schedule assignment")
	ErrOverlappingAssignment = errors.New("monthly schedule assignments overlap")
	ErrTeamMetadataMissing   = errors.New("monthly schedule team metadata is incomplete")
)

const (
	ScheduleStatusDraft     = "draft"
	ScheduleStatusPublished = "published"

	AssignmentTypeField   = "field"
	AssignmentTypeRemote  = "remote"
	AssignmentTypeSupport = "support"
	AssignmentTypeSpecial = "special"

	AssignmentSourceManual       = "manual"
	AssignmentSourceLargeWork    = "large_work"
	AssignmentSourceApprovedFile = "approved_file"
)

type Actor struct {
	UserID int64
	Role   string
}

type Period struct {
	ID    int64
	Year  int
	Month int
}

type Team struct {
	ID                 int64
	Name               string
	Code               *string
	BaseArea           *string
	CrewType           *string
	DisplayOrder       int
	MonthlyPlanVisible bool
}

type Revision struct {
	ID                int64
	MonthlyPlanID     int64
	RevisionNo        int
	Status            string
	CreatedByUserID   int64
	PublishedByUserID *int64
	PublishedAt       *time.Time
	Checksum          *string
	Projection        []byte
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Assignment struct {
	ID             int64
	RevisionID     int64
	TeamID         int64
	AssignmentType string
	StartDate      time.Time
	EndDate        time.Time
	Destination    string
	Note           *string
	SourceType     string
	SourceID       *int64
}

type Schedule struct {
	Revision    *Revision
	Assignments []Assignment
}

type Workspace struct {
	Period    Period
	Teams     []Team
	Draft     *Schedule
	Published *Schedule
}

type Projection struct {
	Period      ProjectionPeriod
	Revision    ProjectionRevision
	PublishedAt time.Time
	Teams       []ProjectedTeam
	Checksum    string
}

type ProjectionPeriod struct {
	Year  int
	Month int
}

type ProjectionRevision struct {
	ID int64
	No int
}

type ProjectedTeam struct {
	ID           int64
	Code         string
	Name         string
	BaseArea     string
	CrewType     string
	DisplayOrder int
	Segments     []ProjectedSegment
}

type ProjectedSegment struct {
	AssignmentID   *int64
	AssignmentType string
	StartDate      string
	EndDate        string
	Destination    string
	Note           *string
	SourceType     string
	SourceID       *int64
}

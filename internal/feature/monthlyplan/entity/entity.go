package entity

import "time"

// Entity represents a monthly plan period.
type Entity struct {
	ID        int64
	Year      int
	Month     int
	IsLocked  bool
	CreatedAt time.Time
}

// PlanFileEntity represents a file uploaded to a monthly plan period.
type PlanFileEntity struct {
	ID            int64
	MonthlyPlanID int64
	TeamID        *int64
	UploadedByID  int64
	FileKey       string
	FileURL       string
	FileName      string
	FileSizeBytes int64
	Description   *string
	WorkStartDate *time.Time
	WorkEndDate   *time.Time
	Destination   *string
	Remarks       *string
	IsMasterPlan  bool
	IsDeleted     bool
	DeletedAt     *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	// Joined relations (populated by adapter)
	TeamName     *string
	UploaderName *string
}

// SettingsEntity represents the single-row monthly plan settings.
type SettingsEntity struct {
	ID                      int64
	LockDay                 int
	AutoCreateDay           int
	AutoCreateTarget        string
	AllowedFileTypes        []string
	MaxFileSizeMB           *int
	ReminderStartDay        int
	AdminCanUploadAfterLock bool
	UpdatedAt               time.Time
}

// FileSizeLogEntity records file size at upload time.
type FileSizeLogEntity struct {
	ID            int64
	PlanFileID    int64
	UploadedByID  int64
	FileSizeBytes int64
	CreatedAt     time.Time
}

// SubmissionStatusEntry is one team's submission status for a period.
type SubmissionStatusEntry struct {
	TeamID    int64
	TeamName  string
	Status    string // "submitted" | "pending" | "missed"
	FileCount int
}

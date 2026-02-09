package models

// Column name constants for type-safe SQL queries.
// PostgreSQL requires quoted identifiers for camelCase column names.

var TaskCol = struct {
	TeamID, JobTypeID, JobDetailID, FeederID, WorkDate, DeletedAt string
}{
	TeamID:      `"teamId"`,
	JobTypeID:   `"jobTypeId"`,
	JobDetailID: `"jobDetailId"`,
	FeederID:    `"feederId"`,
	WorkDate:    `"workdate"`,
	DeletedAt:   `"deletedat"`,
}

var JobDetailCol = struct {
	DeletedAt string
}{
	DeletedAt: `"deletedAt"`,
}

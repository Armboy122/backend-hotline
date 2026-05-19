package entity

import "time"

type Task struct {
	ID              int64
	WorkDate        time.Time
	TeamID          int64
	JobTypeID       int64
	JobDetailID     int64
	FeederID        *int64
	NumPole         *string
	DeviceCode      *string
	Detail          *string
	URLsBefore      []string
	URLsAfter       []string
	Latitude        *float64
	Longitude       *float64
	SourceType      *string
	SourceID        *int64
	LargeWorkTaskID *int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time

	TeamName            *string
	JobTypeName         *string
	JobDetailName       *string
	FeederCode          *string
	StationName         *string
	OperationCenterID   *int64
	OperationCenterName *string
}

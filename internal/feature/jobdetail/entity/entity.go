package entity

import (
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("job detail not found")
	ErrInvalidID    = errors.New("invalid job detail ID")
	ErrNotDeleted   = errors.New("job detail is not deleted")
)

type Entity struct {
	ID        int64
	Name      string
	JobTypeID *int64
	Tasks     int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

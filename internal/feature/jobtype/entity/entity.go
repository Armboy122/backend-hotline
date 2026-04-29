package entity

import "errors"

var (
	ErrNotFound  = errors.New("job type not found")
	ErrInvalidID = errors.New("invalid job type ID")
)

type Entity struct {
	ID    int64
	Name  string
	Tasks int64
}

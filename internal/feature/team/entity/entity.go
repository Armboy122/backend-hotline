package entity

import "errors"

var (
	ErrNotFound  = errors.New("team not found")
	ErrInvalidID = errors.New("invalid team ID")
)

type Entity struct {
	ID    int64
	Name  string
	Tasks int64
}

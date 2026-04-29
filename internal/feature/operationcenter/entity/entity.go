package entity

import "errors"

var (
	ErrNotFound  = errors.New("operation center not found")
	ErrInvalidID = errors.New("invalid operation center ID")
)

type Entity struct {
	ID   int64
	Name string
}

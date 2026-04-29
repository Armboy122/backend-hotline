package entity

import "errors"

var (
	ErrNotFound  = errors.New("PEA not found")
	ErrInvalidID = errors.New("invalid PEA ID")
)

type NestedOperationCenter struct {
	ID   int64
	Name string
}

type Entity struct {
	ID              int64
	Shortname       string
	Fullname        string
	OperationID     int64
	OperationCenter *NestedOperationCenter
}

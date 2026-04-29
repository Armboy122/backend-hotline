package entity

import "errors"

var (
	ErrNotFound  = errors.New("feeder not found")
	ErrInvalidID = errors.New("invalid feeder ID")
)

type NestedStation struct {
	ID              int64
	Name            string
	CodeName        string
	OperationCenter *NestedOperationCenter
}

type NestedOperationCenter struct {
	ID   int64
	Name string
}

type Entity struct {
	ID        int64
	Code      string
	StationID int64
	Tasks     int64
	Station   *NestedStation
}

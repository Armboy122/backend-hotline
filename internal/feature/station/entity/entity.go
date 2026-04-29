package entity

import "errors"

var (
	ErrNotFound  = errors.New("station not found")
	ErrInvalidID = errors.New("invalid station ID")
)

type OperationCenter struct {
	ID   int64
	Name string
}

type Entity struct {
	ID              int64
	Name            string
	CodeName        string
	OperationID     int64
	OperationCenter *OperationCenter
}

type CreateInput struct {
	Name        string
	CodeName    string
	OperationID int64
}

type UpdateInput struct {
	ID          int64
	Name        string
	CodeName    string
	OperationID int64
}

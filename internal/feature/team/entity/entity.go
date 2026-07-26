package entity

import "errors"

var (
	ErrNotFound      = errors.New("team not found")
	ErrInvalidID     = errors.New("invalid team ID")
	ErrInvalidInput  = errors.New("invalid team input")
	ErrCodeImmutable = errors.New("team code cannot be changed after it is assigned")
)

type Entity struct {
	ID                 int64
	Name               string
	Code               *string
	BaseArea           *string
	CrewType           *string
	DisplayOrder       int
	MonthlyPlanVisible bool
	Tasks              int64
}

type UpsertInput struct {
	Name               string
	Code               *string
	BaseArea           *string
	CrewType           *string
	DisplayOrder       *int
	MonthlyPlanVisible *bool
}

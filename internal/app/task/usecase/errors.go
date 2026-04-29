package usecase

import "errors"

// Shared usecase errors for task operations.
var (
	ErrTaskNotFound     = errors.New("task not found")
	ErrInvalidTaskID    = errors.New("invalid task ID")
	ErrWorkDateRequired = errors.New("workDate is required")
	ErrTeamIDRequired   = errors.New("teamId is required")
	ErrJobTypeIDRequired  = errors.New("jobTypeId is required")
	ErrJobDetailIDRequired = errors.New("jobDetailId is required")
	ErrYearMonthRequired  = errors.New("year and month are required")
)

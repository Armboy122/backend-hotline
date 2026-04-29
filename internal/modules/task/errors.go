package task

import taskusecase "backend-hotlines3/internal/app/task/usecase"

var (
	ErrTaskNotFound  = taskusecase.ErrTaskNotFound
	ErrInvalidTaskID = taskusecase.ErrInvalidTaskID
	ErrWorkDateRequired = taskusecase.ErrWorkDateRequired
	ErrTeamIDRequired = taskusecase.ErrTeamIDRequired
	ErrJobTypeIDRequired = taskusecase.ErrJobTypeIDRequired
	ErrJobDetailIDRequired = taskusecase.ErrJobDetailIDRequired
	ErrYearMonthRequired = taskusecase.ErrYearMonthRequired
)

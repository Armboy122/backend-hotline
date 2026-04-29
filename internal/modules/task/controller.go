package task

import (
	v1 "backend-hotlines3/internal/handlers/v1"
	"backend-hotlines3/internal/port/outbound/repository"
)

// Controller is the module-first entrypoint for TaskDaily routes.
// It currently wraps the existing v1 handler while the module is being migrated.
type Controller struct {
	*v1.TaskHandler
}

func NewController(repo repository.TaskRepository) *Controller {
	return &Controller{TaskHandler: v1.NewTaskHandler(repo)}
}

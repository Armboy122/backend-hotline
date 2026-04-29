package task

import "backend-hotlines3/internal/port/outbound/repository"

// Service is the module-local service boundary for TaskDaily.
// It is a Phase 1 placeholder for the SCC-style home for task business logic
// while the legacy v1 handler still owns the HTTP behavior.
type Service struct {
	repo repository.TaskRepository
}

func NewService(repo repository.TaskRepository) *Service {
	return &Service{repo: repo}
}

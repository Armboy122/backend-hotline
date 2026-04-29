package service

import (
	"time"

	"backend-hotlines3/internal/feature/monthlyplan/repository"
)

type Service struct {
	repo    repository.Repository
	storage repository.StoragePort
	clock   func() time.Time
}

func NewService(repo repository.Repository, storage repository.StoragePort) *Service {
	return &Service{repo: repo, storage: storage, clock: time.Now}
}

type PlanFileCreateInput = repository.PlanFileCreateInput

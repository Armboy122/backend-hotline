package service

import (
	"sync"
	"time"

	"backend-hotlines3/internal/feature/monthlyplan/repository"
)

type Service struct {
	repo          repository.Repository
	storage       repository.StoragePort
	clock         func() time.Time
	settingsMu    sync.RWMutex
	settingsCache *repositoryCachedSettings
}

type repositoryCachedSettings struct {
	value     any
	expiresAt time.Time
}

func NewService(repo repository.Repository, storage repository.StoragePort) *Service {
	return &Service{repo: repo, storage: storage, clock: time.Now}
}

func (s *Service) SetClockForTest(clock func() time.Time) {
	s.clock = clock
}

type PlanFileCreateInput = repository.PlanFileCreateInput

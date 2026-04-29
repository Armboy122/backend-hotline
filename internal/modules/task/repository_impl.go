package task

import (
	gormadapter "backend-hotlines3/internal/adapter/out/persistence/gorm"
	"gorm.io/gorm"
)

func NewRepositoryImpl(db *gorm.DB) Repository {
	return gormadapter.NewTaskRepository(db)
}

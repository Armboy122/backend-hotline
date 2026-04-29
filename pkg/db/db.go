package db

import (
	"context"
	"fmt"
	"time"

	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// CamelCaseNamingStrategy keeps column names aligned with existing struct tags.
type CamelCaseNamingStrategy struct {
	schema.NamingStrategy
}

func (s CamelCaseNamingStrategy) ColumnName(table, column string) string {
	return column
}

// Connect establishes the PostgreSQL connection used by the app and CLI tools.
func Connect(ctx context.Context, cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
		cfg.Database.TimeZone,
	)

	logLevel := logger.Warn
	if cfg.Server.Mode == "debug" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NamingStrategy: CamelCaseNamingStrategy{
			schema.NamingStrategy{
				SingularTable: true,
				NoLowerCase:   true,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying database connection: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	return db, nil
}

// AutoMigrate runs the current schema migration set.
func AutoMigrate(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).AutoMigrate(
		&models.OperationCenter{},
		&models.PEA{},
		&models.Station{},
		&models.Feeder{},
		&models.JobType{},
		&models.JobDetail{},
		&models.Team{},
		&models.TaskDaily{},
		&models.User{},
		&models.MonthlyPlan{},
		&models.PlanFile{},
		&models.FileSizeLog{},
		&models.MonthlyPlanSetting{},
	); err != nil {
		return fmt.Errorf("failed to auto-migrate database: %w", err)
	}
	return nil
}

package db

import (
	"context"
	"fmt"
	"strings"
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
	if err := db.WithContext(ctx).AutoMigrate(MigrationModels()...); err != nil {
		return fmt.Errorf("failed to auto-migrate database: %w", err)
	}
	if err := ensureSingleActiveSuperAdminIndex(ctx, db); err != nil {
		return err
	}
	if err := ensurePerformanceIndexes(ctx, db); err != nil {
		return err
	}
	return nil
}

// MigrationModels returns the ordered GORM model set owned by AutoMigrate.
func MigrationModels() []any {
	return []any{
		&models.OperationCenter{},
		&models.PEA{},
		&models.Station{},
		&models.Feeder{},
		&models.JobType{},
		&models.JobDetail{},
		&models.Team{},
		&models.TaskDaily{},
		&models.User{},
		&models.UserCapability{},
		&models.MonthlyPlan{},
		&models.PlanFile{},
		&models.FileSizeLog{},
		&models.MonthlyPlanSetting{},
		&models.TeamPlan{},
		&models.ExternalContact{},
		&models.LargeWorkItem{},
		&models.LargeWorkItemTeam{},
		&models.LargeWorkTask{},
	}
}

func ensurePerformanceIndexes(ctx context.Context, db *gorm.DB) error {
	statements := []string{
		`CREATE INDEX IF NOT EXISTS taskdaily_active_workdate_idx ON "TaskDaily" ("deletedat", "workdate")`,
		`CREATE INDEX IF NOT EXISTS taskdaily_active_workdate_team_idx ON "TaskDaily" ("deletedat", "workdate", "teamId")`,
		`CREATE INDEX IF NOT EXISTS planfile_plan_deleted_team_idx ON "PlanFile" ("monthlyPlanId", "isDeleted", "teamId")`,
		`CREATE INDEX IF NOT EXISTS monthlyplan_year_month_idx ON "MonthlyPlan" ("year", "month")`,
		`CREATE UNIQUE INDEX IF NOT EXISTS user_capabilities_active_user_code_idx ON user_capabilities (user_id, code) WHERE revoked_at IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS taskdaily_large_work_task_unique_idx ON "TaskDaily" (large_work_task_id) WHERE large_work_task_id IS NOT NULL AND deletedat IS NULL`,
	}
	for _, stmt := range statements {
		if err := db.WithContext(ctx).Exec(stmt).Error; err != nil {
			return fmt.Errorf("failed to ensure performance index: %w", err)
		}
	}
	return nil
}

func ensureSingleActiveSuperAdminIndex(ctx context.Context, db *gorm.DB) error {
	stmt := `CREATE UNIQUE INDEX IF NOT EXISTS user_single_active_super_admin_idx
		ON "User" ((role))
		WHERE role = 'super_admin' AND "isActive" = true AND "deletedAt" IS NULL`
	if err := db.WithContext(ctx).Exec(stmt).Error; err != nil {
		if strings.Contains(err.Error(), "could not create unique index") || strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("active super_admin invariant violated; resolve duplicate active super_admin users before migrating: %w", err)
		}
		return fmt.Errorf("failed to ensure single active super_admin index: %w", err)
	}
	return nil
}

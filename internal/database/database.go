package database

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

// CamelCaseNamingStrategy - Custom naming strategy to use camelCase column names
type CamelCaseNamingStrategy struct {
	schema.NamingStrategy
}

func (s CamelCaseNamingStrategy) ColumnName(table, column string) string {
	// Use column names as defined in struct tags (camelCase)
	return column
}

// Connect establishes a connection to the PostgreSQL database using the provided configuration.
// It configures connection pooling, logging level, and custom naming strategy for camelCase columns.
// Returns a GORM database instance or an error if connection fails.
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

	// ใช้ Warn level สำหรับ release mode, Info สำหรับ debug
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

	// Connection Pool settings สำหรับ Neon PostgreSQL
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying database connection: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}

// AutoMigrate runs automatic migration for all models in the database.
// It creates or updates tables to match the model definitions without data loss.
// Returns an error if migration fails for any model.
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
	); err != nil {
		return fmt.Errorf("failed to auto-migrate database: %w", err)
	}
	return nil
}

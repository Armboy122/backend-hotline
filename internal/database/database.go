package database

import (
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

func Connect(cfg *config.Config) (*gorm.DB, error) {
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
		return nil, err
	}

	// Connection Pool settings สำหรับ Neon PostgreSQL
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.OperationCenter{},
		&models.PEA{},
		&models.Station{},
		&models.Feeder{},
		&models.JobType{},
		&models.JobDetail{},
		&models.Team{},
		&models.TaskDaily{},
		&models.User{},
	)
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/database"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

func main() {
	// Parse command line flags
	dropTables := flag.Bool("drop", false, "Drop all existing tables before migration")
	flag.Parse()

	ctx := context.Background()

	// โหลด configuration
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// เชื่อมต่อ database
	db, err := database.Connect(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Drop tables ถ้าต้องการ
	if *dropTables {
		log.Println("Dropping all tables...")
		if err := dropAllTables(db); err != nil {
			log.Fatalf("Failed to drop tables: %v", err)
		}
		log.Println("All tables dropped successfully")
	}

	// Auto migrate models
	log.Println("Creating tables...")
	if err := database.AutoMigrate(ctx, db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Migration completed successfully!")
}

func dropAllTables(db *gorm.DB) error {
	// Drop tables in reverse order to handle foreign key constraints
	models := []interface{}{
		&models.TaskDaily{},
		&models.JobDetail{},
		&models.JobType{},
		&models.Feeder{},
		&models.Station{},
		&models.Team{},
		&models.PEA{},
		&models.OperationCenter{},
	}

	for _, model := range models {
		if err := db.Migrator().DropTable(model); err != nil {
			return fmt.Errorf("failed to drop table: %v", err)
		}
	}

	return nil
}

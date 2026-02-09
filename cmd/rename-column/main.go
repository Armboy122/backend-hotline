package main

import (
	"log"

	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/database"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Check current column name
	type ColInfo struct {
		ColumnName string `gorm:"column:column_name"`
	}

	var colInfo ColInfo
	db.Raw(`
		SELECT column_name
		FROM INFORMATION_SCHEMA.columns
		WHERE table_schema = CURRENT_SCHEMA()
		AND table_name = 'TaskDaily'
		AND column_name ILIKE 'deleted%'
	`).Scan(&colInfo)

	log.Printf("Current deleted column: %s", colInfo.ColumnName)

	// Rename if it's deleted_at
	if colInfo.ColumnName == "deleted_at" {
		log.Println("Renaming deleted_at to deletedAt...")
		if err := db.Exec(`ALTER TABLE "TaskDaily" RENAME COLUMN "deleted_at" TO "deletedAt"`).Error; err != nil {
			log.Fatalf("Failed to rename column: %v", err)
		}
		log.Println("✓ Column renamed successfully")
	} else if colInfo.ColumnName == "deletedAt" {
		log.Println("✓ Column name is already correct (deletedAt)")
	} else {
		log.Printf("Warning: Found unexpected column name: %s", colInfo.ColumnName)
	}
}

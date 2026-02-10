package main

import (
	"context"
	"log"

	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/database"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(ctx, cfg)
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

	// Rename to deletedat (lowercase) to match GORM's expected naming
	if colInfo.ColumnName != "deletedat" {
		log.Printf("Renaming %s to deletedat...", colInfo.ColumnName)
		if err := db.Exec(`ALTER TABLE "TaskDaily" RENAME COLUMN "` + colInfo.ColumnName + `" TO "deletedat"`).Error; err != nil {
			log.Fatalf("Failed to rename column: %v", err)
		}
		log.Println("✓ Column renamed successfully")
	} else {
		log.Println("✓ Column name is already correct (deletedat)")
	}

	// Also rename createdAt/updatedAt to lowercase if needed
	log.Println("Checking createdAt/updatedAt columns...")
	var createdAtCol, updatedAtCol string
	db.Raw(`SELECT column_name FROM INFORMATION_SCHEMA.columns WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'TaskDaily' AND column_name ILIKE 'creat%'`).Scan(&createdAtCol)
	db.Raw(`SELECT column_name FROM INFORMATION_SCHEMA.columns WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'TaskDaily' AND column_name ILIKE 'updat%'`).Scan(&updatedAtCol)

	log.Printf("Current createdAt: %s, updatedAt: %s", createdAtCol, updatedAtCol)

	if createdAtCol != "createdat" && createdAtCol != "" {
		log.Println("Renaming createdAt to createdat...")
		db.Exec(`ALTER TABLE "TaskDaily" RENAME COLUMN "` + createdAtCol + `" TO "createdat"`)
		log.Println("✓ createdAt renamed")
	}

	if updatedAtCol != "updatedat" && updatedAtCol != "" {
		log.Println("Renaming updatedAt to updatedat...")
		db.Exec(`ALTER TABLE "TaskDaily" RENAME COLUMN "` + updatedAtCol + `" TO "updatedat"`)
		log.Println("✓ updatedAt renamed")
	}

	log.Println("All column names updated to match GORM's lowercase naming")
}

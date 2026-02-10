package main

import (
	"context"
	"log"

	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/database"
)

func main() {
	ctx := context.Background()
	cfg, _ := config.LoadConfig(ctx)
	db, _ := database.Connect(ctx, cfg)

	// Rename workDate to workdate
	log.Println("Renaming workDate to workdate...")
	if err := db.Exec(`ALTER TABLE "TaskDaily" RENAME COLUMN "workDate" TO "workdate"`).Error; err != nil {
		log.Printf("Warning: %v", err)
	} else {
		log.Println("✓ workDate renamed to workdate")
	}
}

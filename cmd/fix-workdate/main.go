package main

import (
	"context"
	"log"

	"backend-hotlines3/internal/config"
	dbpkg "backend-hotlines3/pkg/db"
)

func main() {
	ctx := context.Background()
	cfg, _ := config.LoadConfig(ctx)
	db, _ := dbpkg.Connect(ctx, cfg)

	// Rename workDate to workdate
	log.Println("Renaming workDate to workdate...")
	if err := db.Exec(`ALTER TABLE "TaskDaily" RENAME COLUMN "workDate" TO "workdate"`).Error; err != nil {
		log.Printf("Warning: %v", err)
	} else {
		log.Println("✓ workDate renamed to workdate")
	}
}

package main

import (
	"log"

	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/database"

	"gorm.io/gorm"
)

func main() {
	// โหลด configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// เชื่อมต่อ database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// แก้ไข schema
	if err := fixTaskDailySchema(db); err != nil {
		log.Fatalf("Failed to fix TaskDaily schema: %v", err)
	}

	log.Println("Schema fix completed successfully!")
}

func fixTaskDailySchema(db *gorm.DB) error {
	log.Println("Fixing TaskDaily schema...")

	// 1. เช็คว่ามี deleted_at column หรือยัง
	hasDeletedAt := false
	db.Raw(`
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.columns 
		WHERE table_schema = CURRENT_SCHEMA() 
		AND table_name = 'TaskDaily' 
		AND column_name = 'deleted_at'
	`).Scan(&hasDeletedAt)

	if !hasDeletedAt {
		log.Println("Adding deleted_at column to TaskDaily...")
		if err := db.Exec(`ALTER TABLE "TaskDaily" ADD "deleted_at" timestamptz(6)`).Error; err != nil {
			return err
		}
		log.Println("✓ deleted_at column added")
	} else {
		log.Println("✓ deleted_at column already exists")
	}

	// 2. เช็คว่ามี work_date column หรือยัง
	hasWorkDate := false
	db.Raw(`
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.columns 
		WHERE table_schema = CURRENT_SCHEMA() 
		AND table_name = 'TaskDaily' 
		AND column_name = 'workDate'
	`).Scan(&hasWorkDate)

	if !hasWorkDate {
		log.Println("Adding workDate column to TaskDaily...")
		if err := db.Exec(`ALTER TABLE "TaskDaily" ADD "workDate" date NOT NULL DEFAULT CURRENT_DATE`).Error; err != nil {
			log.Printf("Warning: Failed to add workDate column: %v", err)
		} else {
			log.Println("✓ workDate column added")
		}
	} else {
		log.Println("✓ workDate column already exists")
	}

	// 3. เช็คว่ามี task_date column เก่าหรือไม่ ถ้ามีให้ migrate data
	hasTaskDate := false
	db.Raw(`
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.columns 
		WHERE table_schema = CURRENT_SCHEMA() 
		AND table_name = 'TaskDaily' 
		AND column_name = 'task_date'
	`).Scan(&hasTaskDate)

	if hasTaskDate {
		log.Println("Found old task_date column, migrating data to workDate...")
		// Migrate data from task_date to workDate
		if err := db.Exec(`
			UPDATE "TaskDaily"
			SET "workDate" = "task_date"
			WHERE "task_date" IS NOT NULL
		`).Error; err != nil {
			log.Printf("Warning: Failed to migrate data from task_date to workDate: %v", err)
		} else {
			log.Println("✓ Data migrated from task_date to workDate")
		}

		// Drop old column
		log.Println("Dropping old task_date column...")
		if err := db.Exec(`ALTER TABLE "TaskDaily" DROP COLUMN IF EXISTS "task_date"`).Error; err != nil {
			log.Printf("Warning: Failed to drop task_date column: %v", err)
		} else {
			log.Println("✓ Old task_date column dropped")
		}
	}

	// 4. สร้าง indexes ที่จำเป็น
	log.Println("Creating indexes...")

	// Index on workDate
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS "TaskDaily_workDate_idx" ON "TaskDaily"("workDate")
	`).Error; err != nil {
		log.Printf("Warning: Failed to create workDate index: %v", err)
	} else {
		log.Println("✓ workDate index created/verified")
	}

	// Index on jobTypeId and jobDetailId
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS "TaskDaily_jobTypeId_jobDetailId_idx" ON "TaskDaily"("jobTypeId", "jobDetailId")
	`).Error; err != nil {
		log.Printf("Warning: Failed to create jobTypeId/jobDetailId index: %v", err)
	} else {
		log.Println("✓ jobTypeId/jobDetailId index created/verified")
	}

	// Index on feederId
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS "TaskDaily_feederId_idx" ON "TaskDaily"("feederId")
	`).Error; err != nil {
		log.Printf("Warning: Failed to create feederId index: %v", err)
	} else {
		log.Println("✓ feederId index created/verified")
	}

	// Index on latitude and longitude
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS "TaskDaily_latitude_longitude_idx" ON "TaskDaily"("latitude", "longitude")
	`).Error; err != nil {
		log.Printf("Warning: Failed to create latitude/longitude index: %v", err)
	} else {
		log.Println("✓ latitude/longitude index created/verified")
	}

	return nil
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/database"
	"backend-hotlines3/internal/models"

	"gorm.io/gorm"
)

func main() {
	ctx := context.Background()

	// โหลด configuration
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// เชื่อมต่อ database (ไม่ auto migrate)
	db, err := database.Connect(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	fmt.Println("✅ เชื่อมต่อฐานข้อมูลสำเร็จ")
	fmt.Println()

	// ทดสอบอ่านข้อมูลจากแต่ละตาราง
	testOperationCenter(db)
	testPea(db)
	testStation(db)
	testFeeder(db)
	testJobType(db)
	testJobDetail(db)
	testTeam(db)
	testTaskDaily(db)

	fmt.Println("✅ ทดสอบเสร็จสมบูรณ์!")
}

func testOperationCenter(db *gorm.DB) {
	var operationCenters []models.OperationCenter
	result := db.Find(&operationCenters)
	if result.Error != nil {
		log.Printf("❌ Error reading OperationCenter: %v", result.Error)
		return
	}
	fmt.Printf("📋 OperationCenter: %d รายการ\n", len(operationCenters))
	if len(operationCenters) > 0 {
		data, _ := json.MarshalIndent(operationCenters[0], "", "  ")
		fmt.Printf("   ตัวอย่าง: %s\n", data)
	}
	fmt.Println()
}

func testPea(db *gorm.DB) {
	var peas []models.PEA
	result := db.Find(&peas)
	if result.Error != nil {
		log.Printf("❌ Error reading Pea: %v", result.Error)
		return
	}
	fmt.Printf("📋 Pea: %d รายการ\n", len(peas))
	if len(peas) > 0 {
		data, _ := json.MarshalIndent(peas[0], "", "  ")
		fmt.Printf("   ตัวอย่าง: %s\n", data)
	}
	fmt.Println()
}

func testStation(db *gorm.DB) {
	var stations []models.Station
	result := db.Find(&stations)
	if result.Error != nil {
		log.Printf("❌ Error reading Station: %v", result.Error)
		return
	}
	fmt.Printf("📋 Station: %d รายการ\n", len(stations))
	if len(stations) > 0 {
		data, _ := json.MarshalIndent(stations[0], "", "  ")
		fmt.Printf("   ตัวอย่าง: %s\n", data)
	}
	fmt.Println()
}

func testFeeder(db *gorm.DB) {
	var feeders []models.Feeder
	result := db.Find(&feeders)
	if result.Error != nil {
		log.Printf("❌ Error reading Feeder: %v", result.Error)
		return
	}
	fmt.Printf("📋 Feeder: %d รายการ\n", len(feeders))
	if len(feeders) > 0 {
		data, _ := json.MarshalIndent(feeders[0], "", "  ")
		fmt.Printf("   ตัวอย่าง: %s\n", data)
	}
	fmt.Println()
}

func testJobType(db *gorm.DB) {
	var jobTypes []models.JobType
	result := db.Find(&jobTypes)
	if result.Error != nil {
		log.Printf("❌ Error reading JobType: %v", result.Error)
		return
	}
	fmt.Printf("📋 JobType: %d รายการ\n", len(jobTypes))
	if len(jobTypes) > 0 {
		data, _ := json.MarshalIndent(jobTypes[0], "", "  ")
		fmt.Printf("   ตัวอย่าง: %s\n", data)
	}
	fmt.Println()
}

func testJobDetail(db *gorm.DB) {
	var jobDetails []models.JobDetail
	result := db.Find(&jobDetails)
	if result.Error != nil {
		log.Printf("❌ Error reading JobDetail: %v", result.Error)
		return
	}
	fmt.Printf("📋 JobDetail: %d รายการ\n", len(jobDetails))
	if len(jobDetails) > 0 {
		data, _ := json.MarshalIndent(jobDetails[0], "", "  ")
		fmt.Printf("   ตัวอย่าง: %s\n", data)
	}
	fmt.Println()
}

func testTeam(db *gorm.DB) {
	var teams []models.Team
	result := db.Find(&teams)
	if result.Error != nil {
		log.Printf("❌ Error reading Team: %v", result.Error)
		return
	}
	fmt.Printf("📋 Team: %d รายการ\n", len(teams))
	if len(teams) > 0 {
		data, _ := json.MarshalIndent(teams[0], "", "  ")
		fmt.Printf("   ตัวอย่าง: %s\n", data)
	}
	fmt.Println()
}

func testTaskDaily(db *gorm.DB) {
	var tasks []models.TaskDaily
	result := db.Limit(5).Find(&tasks)
	if result.Error != nil {
		log.Printf("❌ Error reading TaskDaily: %v", result.Error)
		return
	}
	fmt.Printf("📋 TaskDaily: แสดง %d รายการแรก\n", len(tasks))
	if len(tasks) > 0 {
		data, _ := json.MarshalIndent(tasks[0], "", "  ")
		fmt.Printf("   ตัวอย่าง: %s\n", data)
	}
	fmt.Println()
}

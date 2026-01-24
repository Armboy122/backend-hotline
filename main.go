package main

import (
	"fmt"
	"log"

	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/database"
	"backend-hotlines3/internal/router"

	"github.com/gin-gonic/gin"
)

func main() {
	// โหลด configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ตั้งค่า Gin mode
	gin.SetMode(cfg.Server.Mode)

	// เชื่อมต่อ database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Auto migrate models
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// สร้าง router
	r := router.SetupRouter(cfg, db)

	// เริ่ม server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

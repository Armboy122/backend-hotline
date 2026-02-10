package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend-hotlines3/internal/config"
	"backend-hotlines3/internal/database"
	"backend-hotlines3/internal/middleware"
	"backend-hotlines3/internal/router"
	"backend-hotlines3/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create root context for application startup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// โหลด configuration
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ตั้งค่า Gin mode
	gin.SetMode(cfg.Server.Mode)

	// เชื่อมต่อ database
	db, err := database.Connect(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Auto migrate models (เฉพาะเมื่อ auto_migrate: true ใน config)
	if cfg.Database.AutoMigrate {
		log.Println("Running AutoMigrate...")
		if err := database.AutoMigrate(ctx, db); err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}
		log.Println("AutoMigrate completed")
	}

	// Initialize JWT Manager
	accessTokenExpiry, err := time.ParseDuration(cfg.JWT.AccessTokenExpiry)
	if err != nil {
		log.Fatalf("Failed to parse access token expiry: %v", err)
	}
	refreshTokenExpiry, err := time.ParseDuration(cfg.JWT.RefreshTokenExpiry)
	if err != nil {
		log.Fatalf("Failed to parse refresh token expiry: %v", err)
	}

	jwtManager := jwt.NewJWTManager(cfg.JWT.Secret, accessTokenExpiry, refreshTokenExpiry)

	// สร้าง router
	r := router.SetupRouter(cfg, db, jwtManager)

	// Global Recovery Middleware (Handle Panics)
	r.Use(middleware.RecoveryMiddleware())

	// สร้าง HTTP server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Setup graceful shutdown
	// Channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sig := <-sigChan
	log.Printf("Received signal: %v. Shutting down gracefully...", sig)

	// Create a deadline for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

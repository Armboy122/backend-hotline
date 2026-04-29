package router

import (
	"context"
	"time"

	gormadapter "backend-hotlines3/internal/adapter/out/persistence/gorm"
	storageadapter "backend-hotlines3/internal/adapter/out/storage"
	"backend-hotlines3/internal/config"
	mpusecase "backend-hotlines3/internal/app/monthlyplan/usecase"
	v1 "backend-hotlines3/internal/handlers/v1"
	taskmodule "backend-hotlines3/internal/modules/task"
	"backend-hotlines3/internal/middleware"
	"backend-hotlines3/pkg/jwt"
	"backend-hotlines3/pkg/s3"
	"log"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(cfg *config.Config, db *gorm.DB, jwtManager *jwt.JWTManager) *gin.Engine {
	r := gin.Default()

	// Global Recovery Middleware (must be first to catch panics in all routes)
	r.Use(middleware.RecoveryMiddleware())

	// Request Timeout Middleware (30 seconds)
	r.Use(middleware.TimeoutMiddleware(30 * time.Second))

	// Rate Limiting Middleware (100 requests per minute per IP)
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	r.Use(rateLimiter.Middleware())

	// Gzip Compression Middleware
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// CORS middleware
	r.Use(CORSMiddleware(cfg))

	// Health check — cache 1 minute
	r.GET("/health", middleware.CachePublic(60), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	// ============================================
	// API v1 Routes (New Standard Format)
	// ============================================
	apiV1 := r.Group("/v1")
	{
		// Auth middleware
		authMw := middleware.NewAuthMiddleware(jwtManager)

		// Auth Routes — no CDN cache (mutations + user-specific)
		authHandler := v1.NewAuthHandler(db, jwtManager)
		authGroup := apiV1.Group("/auth")
		{
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/refresh", authHandler.RefreshToken)
			authGroup.POST("/logout", authMw.RequireAuth(), authHandler.Logout)
			authGroup.GET("/me", authMw.RequireAuth(), middleware.CachePrivate(), authHandler.Me)
		}

		// Teams — cache 2 minutes (has task counts that update with new tasks)
		teamsV1 := apiV1.Group("/teams")
		{
			handler := v1.NewTeamHandler(db)
			teamsV1.GET("", middleware.CachePublic(120), handler.List)
			teamsV1.GET("/:id", middleware.CachePublic(120), handler.GetByID)
			teamsV1.POST("", handler.Create)
			teamsV1.PUT("/:id", handler.Update)
			teamsV1.DELETE("/:id", handler.Delete)
		}

		// Job Types — cache 5 minutes (admin-only edits, changes infrequently)
		jobTypesV1 := apiV1.Group("/job-types")
		{
			handler := v1.NewJobTypeHandler(db)
			jobTypesV1.GET("", middleware.CachePublic(300), handler.List)
			jobTypesV1.GET("/:id", middleware.CachePublic(300), handler.GetByID)
			jobTypesV1.POST("", handler.Create)
			jobTypesV1.PUT("/:id", handler.Update)
			jobTypesV1.DELETE("/:id", handler.Delete)
		}

		// Job Details — cache 5 minutes (admin-only edits, changes infrequently)
		jobDetailsV1 := apiV1.Group("/job-details")
		{
			handler := v1.NewJobDetailHandler(db)
			jobDetailsV1.GET("", middleware.CachePublic(300), handler.List)
			jobDetailsV1.GET("/:id", middleware.CachePublic(300), handler.GetByID)
			jobDetailsV1.POST("", handler.Create)
			jobDetailsV1.PUT("/:id", handler.Update)
			jobDetailsV1.DELETE("/:id", handler.Delete)
			jobDetailsV1.POST("/:id/restore", handler.Restore)
		}

		// Feeders — cache 2 minutes (has task counts that update with new tasks)
		feedersV1 := apiV1.Group("/feeders")
		{
			handler := v1.NewFeederHandler(db)
			feedersV1.GET("", middleware.CachePublic(120), handler.List)
			feedersV1.GET("/:id", middleware.CachePublic(120), handler.GetByID)
			feedersV1.POST("", handler.Create)
			feedersV1.PUT("/:id", handler.Update)
			feedersV1.DELETE("/:id", handler.Delete)
		}

		// Stations — cache 10 minutes (static reference data)
		stationsV1 := apiV1.Group("/stations")
		{
			handler := v1.NewStationHandler(db)
			stationsV1.GET("", middleware.CachePublic(600), handler.List)
			stationsV1.GET("/:id", middleware.CachePublic(600), handler.GetByID)
			stationsV1.POST("", handler.Create)
			stationsV1.PUT("/:id", handler.Update)
			stationsV1.DELETE("/:id", handler.Delete)
		}

		// PEAs — cache 10 minutes (static reference data)
		peasV1 := apiV1.Group("/peas")
		{
			handler := v1.NewPEAHandler(db)
			peasV1.GET("", middleware.CachePublic(600), handler.List)
			peasV1.GET("/:id", middleware.CachePublic(600), handler.GetByID)
			peasV1.POST("", handler.Create)
			peasV1.POST("/bulk", handler.BulkCreate)
			peasV1.PUT("/:id", handler.Update)
			peasV1.DELETE("/:id", handler.Delete)
		}

		// Operation Centers — cache 10 minutes (static reference data, rarely changes)
		operationCentersV1 := apiV1.Group("/operation-centers")
		{
			handler := v1.NewOperationCenterHandler(db)
			operationCentersV1.GET("", middleware.CachePublic(600), handler.List)
			operationCentersV1.GET("/:id", middleware.CachePublic(600), handler.GetByID)
			operationCentersV1.POST("", handler.Create)
			operationCentersV1.PUT("/:id", handler.Update)
			operationCentersV1.DELETE("/:id", handler.Delete)
		}

		// Tasks
		tasksV1 := apiV1.Group("/tasks")
		{
			handler := taskmodule.NewController(taskmodule.NewRepositoryImpl(db))
			tasksV1.GET("", middleware.CachePublic(60), handler.List)                    // cache 1 min (paginated, dynamic filters)
			tasksV1.GET("/by-team", middleware.CachePublic(120), handler.ListByTeam)     // cache 2 min
			tasksV1.GET("/by-filter", middleware.CachePublic(180), handler.ListByFilter) // cache 3 min (per year/month combo)
			tasksV1.GET("/:id", middleware.CachePublic(60), handler.GetByID)             // cache 1 min
			tasksV1.POST("", handler.Create)
			tasksV1.PUT("/:id", handler.Update)
			tasksV1.DELETE("/:id", handler.Delete)
		}

		// Upload — no cache (presigned URLs are unique per request)
		uploadHandler, err := v1.NewUploadHandler(cfg)
		if err != nil {
			log.Printf("Warning: Upload handler initialization failed: %v", err)
		} else {
			uploadV1 := apiV1.Group("/upload")
			{
				uploadV1.POST("/image", uploadHandler.GetPresignedURL)
				uploadV1.DELETE("/*key", uploadHandler.DeleteFile)
			}
		}

		// Monthly Plan File Management — wired via use cases
		mpRepo := gormadapter.NewMonthlyPlanRepository(db)
		r2Client, mpR2Err := s3.NewR2Client(s3.R2Config{
			AccountID:       cfg.Cloudflare.R2.AccountID,
			AccessKeyID:     cfg.Cloudflare.R2.AccessKeyID,
			SecretAccessKey: cfg.Cloudflare.R2.SecretAccessKey,
			BucketName:      cfg.Cloudflare.R2.BucketName,
			PublicURL:       cfg.Cloudflare.R2.PublicURL,
		})
		if mpR2Err != nil {
			log.Printf("Warning: R2 client init failed: %v", mpR2Err)
		}
		var monthlyPlanHandler *v1.MonthlyPlanHandler
		if mpR2Err == nil {
			r2Storage := storageadapter.NewR2StorageAdapter(r2Client)
			monthlyPlanHandler = v1.NewMonthlyPlanHandler(v1.MonthlyPlanHandlerDeps{
				GetSettingsUC:      mpusecase.NewGetSettingsUseCase(mpRepo),
				UpdateSettingsUC:   mpusecase.NewUpdateSettingsUseCase(mpRepo),
				EnsurePeriodUC:     mpusecase.NewEnsurePeriodUseCase(mpRepo),
				PresignUploadUC:    mpusecase.NewPresignUploadUseCase(mpRepo, r2Storage),
				ConfirmUploadUC:    mpusecase.NewConfirmUploadUseCase(mpRepo),
				ListFilesUC:        mpusecase.NewListFilesUseCase(mpRepo),
				GetFileUC:          mpusecase.NewGetFileUseCase(mpRepo),
				SoftDeleteFileUC:   mpusecase.NewSoftDeleteFileUseCase(mpRepo),
				RestoreFileUC:     mpusecase.NewRestoreFileUseCase(mpRepo),
				HardDeleteFileUC:   mpusecase.NewHardDeleteFileUseCase(mpRepo, r2Storage),
				SubmissionStatusUC: mpusecase.NewGetSubmissionStatusUseCase(mpRepo),
				DownloadURLFunc: func(fileKey string) (string, error) {
					return r2Storage.PresignDownload(context.Background(), fileKey, 15*time.Minute)
				},
			})
		}
		if monthlyPlanHandler != nil {
			monthlyPlansV1 := apiV1.Group("/monthly-plans")
			monthlyPlansV1.Use(authMw.RequireAuth())
			{
				// Settings (admin only)
				adminOnly := monthlyPlansV1.Group("")
				adminOnly.Use(authMw.RequireRole("admin"))
				{
					adminOnly.GET("/settings", monthlyPlanHandler.GetSettings)
					adminOnly.PUT("/settings", monthlyPlanHandler.UpdateSettings)
					adminOnly.DELETE("/files/:id/permanent", monthlyPlanHandler.HardDeleteFile)
					adminOnly.POST("/files/:id/restore", monthlyPlanHandler.RestoreFile)
				}

				// Period endpoints (all authenticated users)
				monthlyPlansV1.GET("/:year/:month", monthlyPlanHandler.GetOrCreatePeriod)
				monthlyPlansV1.GET("/:year/:month/files", monthlyPlanHandler.ListFiles)
				monthlyPlansV1.GET("/:year/:month/status", monthlyPlanHandler.GetSubmissionStatus)

				// Upload flow (presign + confirm)
				monthlyPlansV1.POST("/:year/:month/files/presign", monthlyPlanHandler.PresignUpload)
				monthlyPlansV1.POST("/:year/:month/files", monthlyPlanHandler.ConfirmUpload)

				// File actions
				monthlyPlansV1.DELETE("/files/:id", monthlyPlanHandler.SoftDeleteFile)
				monthlyPlansV1.GET("/files/:id/download", monthlyPlanHandler.GetDownloadURL)
			}
		}

		// Dashboard — cache 5 minutes (heavy aggregation queries, stale data acceptable)
		dashboardV1 := apiV1.Group("/dashboard")
		{
			handler := v1.NewDashboardHandler(db)
			dashboardV1.GET("/summary", middleware.CachePublic(300), handler.Summary)
			dashboardV1.GET("/top-jobs", middleware.CachePublic(300), handler.TopJobs)
			dashboardV1.GET("/top-feeders", middleware.CachePublic(300), handler.TopFeeders)
			dashboardV1.GET("/feeder-matrix", middleware.CachePublic(300), handler.FeederMatrix)
			dashboardV1.GET("/stats", middleware.CachePublic(300), handler.Stats)
		}

		// Users — no cache (admin-only + user-specific context)
		usersV1 := apiV1.Group("/users")
		{
			handler := v1.NewUserHandler(db)

			// Apply authentication middleware to all user routes
			usersV1.Use(authMw.RequireAuth())

			// Admin only routes
			adminUsers := usersV1.Group("")
			adminUsers.Use(authMw.RequireRole("admin"))
			{
				adminUsers.GET("", middleware.CachePrivate(), handler.List)
				adminUsers.GET("/:id", middleware.CachePrivate(), handler.GetByID)
				adminUsers.POST("", handler.Create)
				adminUsers.PUT("/:id", handler.Update)
				adminUsers.DELETE("/:id", handler.Delete)
			}

			// User can change their own password (authenticated, but not necessarily admin)
			usersV1.PUT("/:id/password", handler.ChangePassword)
		}
	}

	return r
}

func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowedOrigin := "*" // fallback for dev

		if len(cfg.CORS.AllowedOrigins) > 0 {
			allowedOrigin = "" // deny by default if origins are configured
			for _, o := range cfg.CORS.AllowedOrigins {
				if o == origin || o == "*" {
					allowedOrigin = origin
					break
				}
			}
		}

		if allowedOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

package router

import (
	"log"
	"time"

	"backend-hotlines3/internal/config"

	authcontroller "backend-hotlines3/internal/feature/auth/controller"
	"backend-hotlines3/internal/feature/auth/policy"
	authrepository "backend-hotlines3/internal/feature/auth/repository"
	authservice "backend-hotlines3/internal/feature/auth/service"
	capabilitycontroller "backend-hotlines3/internal/feature/capability/controller"
	capabilityrepository "backend-hotlines3/internal/feature/capability/repository"
	capabilityservice "backend-hotlines3/internal/feature/capability/service"
	dailyreportdraftcontroller "backend-hotlines3/internal/feature/dailyreportdraft/controller"
	dailyreportdraftservice "backend-hotlines3/internal/feature/dailyreportdraft/service"
	dashboardcontroller "backend-hotlines3/internal/feature/dashboard/controller"
	dashboardrepository "backend-hotlines3/internal/feature/dashboard/repository"
	dashboardservice "backend-hotlines3/internal/feature/dashboard/service"
	feedercontroller "backend-hotlines3/internal/feature/feeder/controller"
	feederrepository "backend-hotlines3/internal/feature/feeder/repository"
	feederservice "backend-hotlines3/internal/feature/feeder/service"
	jobdetailcontroller "backend-hotlines3/internal/feature/jobdetail/controller"
	jobdetailrepository "backend-hotlines3/internal/feature/jobdetail/repository"
	jobdetailservice "backend-hotlines3/internal/feature/jobdetail/service"
	jobtypecontroller "backend-hotlines3/internal/feature/jobtype/controller"
	jobtyperepository "backend-hotlines3/internal/feature/jobtype/repository"
	jobtypeservice "backend-hotlines3/internal/feature/jobtype/service"
	largeworkcontroller "backend-hotlines3/internal/feature/largework/controller"
	largeworkrepository "backend-hotlines3/internal/feature/largework/repository"
	largeworkservice "backend-hotlines3/internal/feature/largework/service"
	monthlyplancontroller "backend-hotlines3/internal/feature/monthlyplan/controller"
	monthlyplanrepository "backend-hotlines3/internal/feature/monthlyplan/repository"
	monthlyplanservice "backend-hotlines3/internal/feature/monthlyplan/service"
	opcontroller "backend-hotlines3/internal/feature/operationcenter/controller"
	oprepository "backend-hotlines3/internal/feature/operationcenter/repository"
	opservice "backend-hotlines3/internal/feature/operationcenter/service"
	peacontroller "backend-hotlines3/internal/feature/pea/controller"
	pearepository "backend-hotlines3/internal/feature/pea/repository"
	peaservicepkg "backend-hotlines3/internal/feature/pea/service"
	planningcalendarcontroller "backend-hotlines3/internal/feature/planningcalendar/controller"
	planningcalendarservice "backend-hotlines3/internal/feature/planningcalendar/service"
	stationcontroller "backend-hotlines3/internal/feature/station/controller"
	stationrepository "backend-hotlines3/internal/feature/station/repository"
	stationservice "backend-hotlines3/internal/feature/station/service"
	taskcontroller "backend-hotlines3/internal/feature/task/controller"
	taskrepository "backend-hotlines3/internal/feature/task/repository"
	taskservice "backend-hotlines3/internal/feature/task/service"
	teamcontroller "backend-hotlines3/internal/feature/team/controller"
	teamrepository "backend-hotlines3/internal/feature/team/repository"
	teamservice "backend-hotlines3/internal/feature/team/service"
	teamplancontroller "backend-hotlines3/internal/feature/teamplan/controller"
	teamplanrepository "backend-hotlines3/internal/feature/teamplan/repository"
	teamplanservice "backend-hotlines3/internal/feature/teamplan/service"
	uploadcontroller "backend-hotlines3/internal/feature/upload/controller"
	usercontroller "backend-hotlines3/internal/feature/user/controller"
	userrepository "backend-hotlines3/internal/feature/user/repository"
	userservice "backend-hotlines3/internal/feature/user/service"
	workreportcontroller "backend-hotlines3/internal/feature/workreport/controller"
	workreportrepository "backend-hotlines3/internal/feature/workreport/repository"
	workreportservice "backend-hotlines3/internal/feature/workreport/service"
	"backend-hotlines3/internal/middleware"
	"backend-hotlines3/pkg/jwt"
	"backend-hotlines3/pkg/s3"

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
		var uRepo userrepository.Repository
		if db != nil {
			uRepo = userrepository.NewRepository(db)
			authMw.SetPasswordStatusStore(uRepo)
		}
		capabilityRepo := capabilityrepository.NewRepository(db)
		capabilitySvc := capabilityservice.NewService(capabilityRepo)
		capabilityCtrl := capabilitycontroller.NewController(capabilitySvc)

		// Auth Routes — no CDN cache (mutations + user-specific)
		{
			aRepo := authrepository.NewRepository(db)
			aSvc := authservice.NewService(aRepo, jwtManager)
			aCtrl := authcontroller.NewController(aSvc)
			authGroup := apiV1.Group("/auth")
			authGroup.POST("/login", aCtrl.Login)
			authGroup.POST("/register", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), aCtrl.Register)
			authGroup.POST("/refresh", aCtrl.RefreshToken)
			authGroup.POST("/logout", authMw.RequireAuth(), aCtrl.Logout)
			authGroup.GET("/me", authMw.RequireAuth(), middleware.CachePrivate(30), aCtrl.Me)
		}

		// ── Master Data Feature (Phase D) ───────────────────────────────────

		// Capabilities — round-1 capability grant/revoke for approved monthly plan upload.
		{
			capabilitiesV1 := apiV1.Group("/capabilities")
			capabilitiesV1.Use(authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...))
			capabilitiesV1.GET("", capabilityCtrl.List)
		}

		// Teams — cache 2 minutes (has task counts that update with new tasks)
		teamsV1 := apiV1.Group("/teams")
		{
			repository := teamrepository.NewRepository(db)
			service := teamservice.NewService(repository)
			handler := teamcontroller.NewController(service)
			teamsV1.GET("", middleware.CachePublic(120), handler.List)
			teamsV1.GET("/:id", middleware.CachePublic(120), handler.GetByID)
			teamsV1.POST("", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), handler.Create)
			teamsV1.PUT("/:id", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), handler.Update)
			teamsV1.DELETE("/:id", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), handler.Delete)
		}

		// Job Types — cache 5 minutes (admin-only edits, changes infrequently)
		jobTypesV1 := apiV1.Group("/job-types")
		{
			repository := jobtyperepository.NewRepository(db)
			service := jobtypeservice.NewService(repository)
			handler := jobtypecontroller.NewController(service)
			jobTypesV1.GET("", middleware.CachePublic(300), handler.List)
			jobTypesV1.GET("/:id", middleware.CachePublic(300), handler.GetByID)
			jobTypesV1.POST("", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), handler.Create)
			jobTypesV1.PUT("/:id", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), handler.Update)
			jobTypesV1.DELETE("/:id", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), handler.Delete)
		}

		// Job Details — cache 5 minutes (admin-only edits, changes infrequently)
		jobDetailsV1 := apiV1.Group("/job-details")
		{
			jdRepo := jobdetailrepository.NewRepository(db)
			jdSvc := jobdetailservice.NewService(jdRepo)
			jdCtrl := jobdetailcontroller.NewController(jdSvc)
			jobDetailsV1.GET("", middleware.CachePublic(300), jdCtrl.List)
			jobDetailsV1.GET("/:id", middleware.CachePublic(300), jdCtrl.GetByID)
			jobDetailsV1.POST("", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), jdCtrl.Create)
			jobDetailsV1.PUT("/:id", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), jdCtrl.Update)
			jobDetailsV1.DELETE("/:id", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), jdCtrl.Delete)
			jobDetailsV1.POST("/:id/restore", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), jdCtrl.Restore)
		}

		// Feeders — cache 2 minutes (has task counts that update with new tasks)
		feedersV1 := apiV1.Group("/feeders")
		{
			fRepo := feederrepository.NewRepository(db)
			fSvc := feederservice.NewService(fRepo)
			fCtrl := feedercontroller.NewController(fSvc)
			feedersV1.GET("", middleware.CachePublic(120), fCtrl.List)
			feedersV1.GET("/:id", middleware.CachePublic(120), fCtrl.GetByID)
			feedersV1.POST("", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), fCtrl.Create)
			feedersV1.PUT("/:id", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), fCtrl.Update)
			feedersV1.DELETE("/:id", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), fCtrl.Delete)
		}

		// Stations — cache 10 minutes (static reference data)
		stationsV1 := apiV1.Group("/stations")
		{
			repository := stationrepository.NewRepository(db)
			service := stationservice.NewService(repository)
			handler := stationcontroller.NewController(service)
			stationsV1.GET("", middleware.CachePublic(600), handler.List)
			stationsV1.GET("/:id", middleware.CachePublic(600), handler.GetByID)
			stationsV1.POST("", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), handler.Create)
			stationsV1.PUT("/:id", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), handler.Update)
			stationsV1.DELETE("/:id", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), handler.Delete)
		}

		// PEAs — cache 10 minutes (static reference data)
		peasV1 := apiV1.Group("/peas")
		{
			peaRepo := pearepository.NewRepository(db)
			peaSvc := peaservicepkg.NewService(peaRepo)
			peaCtrl := peacontroller.NewController(peaSvc)
			peasV1.GET("", middleware.CachePublic(600), peaCtrl.List)
			peasV1.GET("/:id", middleware.CachePublic(600), peaCtrl.GetByID)
			peasV1.POST("", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), peaCtrl.Create)
			peasV1.POST("/bulk", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), peaCtrl.BulkCreate)
			peasV1.PUT("/:id", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), peaCtrl.Update)
			peasV1.DELETE("/:id", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), peaCtrl.Delete)
		}

		// Operation Centers — cache 10 minutes (static reference data, rarely changes)
		operationCentersV1 := apiV1.Group("/operation-centers")
		{
			ocRepo := oprepository.NewRepository(db)
			ocSvc := opservice.NewService(ocRepo)
			ocCtrl := opcontroller.NewController(ocSvc)
			operationCentersV1.GET("", middleware.CachePublic(600), ocCtrl.List)
			operationCentersV1.GET("/:id", middleware.CachePublic(600), ocCtrl.GetByID)
			operationCentersV1.POST("", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), ocCtrl.Create)
			operationCentersV1.PUT("/:id", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), ocCtrl.Update)
			operationCentersV1.DELETE("/:id", authMw.RequireAuth(), authMw.RequireRole(superAdminOnlyRoles()...), ocCtrl.Delete)
		}

		// Tasks — authenticated and team-scoped; non-privileged users only see their own team.
		tasksV1 := apiV1.Group("/tasks")
		tasksV1.Use(authMw.RequireAuth())
		{
			repository := taskrepository.NewRepository(db)
			service := taskservice.NewService(repository)
			handler := taskcontroller.NewController(service)
			tasksV1.GET("", middleware.CachePrivate(), handler.List)
			tasksV1.GET("/by-team", middleware.CachePrivate(), handler.ListByTeam)
			tasksV1.GET("/by-filter", middleware.CachePrivate(), handler.ListByFilter)
			tasksV1.GET("/:id", middleware.CachePrivate(), handler.GetByID)
			tasksV1.POST("", handler.Create)
			tasksV1.PUT("/:id", handler.Update)
			tasksV1.DELETE("/:id", handler.Delete)
		}

		// Work reports — read-only monthly aggregation over TaskDaily.
		{
			workReportRepo := workreportrepository.NewRepository(db)
			workReportSvc := workreportservice.NewService(workReportRepo)
			workReportCtrl := workreportcontroller.NewController(workReportSvc)
			workReportsV1 := apiV1.Group("/work-reports")
			workReportsV1.Use(authMw.RequireAuth())
			workReportsV1.GET("", middleware.CachePrivate(), workReportCtrl.GetReport)
		}

		// Daily-report draft / prefill — authenticated and team-scoped.
		{
			teamPlanRepo := teamplanrepository.NewRepository(db)
			monthlyPlanRepo := monthlyplanrepository.NewRepository(db)
			lwRepo := largeworkrepository.NewRepository(db)
			draftSvc := dailyreportdraftservice.NewService(teamPlanRepo, monthlyPlanRepo, lwRepo)
			draftCtrl := dailyreportdraftcontroller.NewController(draftSvc)
			dailyDraftV1 := apiV1.Group("/daily-report-drafts")
			dailyDraftV1.Use(authMw.RequireAuth())
			{
				dailyDraftV1.GET("/sources", middleware.CachePrivate(), draftCtrl.ListSources)
				dailyDraftV1.GET("/from-plan", middleware.CachePrivate(), draftCtrl.FromPlan)
			}
		}

		// Upload — no cache (presigned URLs are unique per request)
		uploadHandler, err := uploadcontroller.NewController(cfg)
		if err != nil {
			log.Printf("Warning: Upload handler initialization failed: %v", err)
		} else {
			uploadV1 := apiV1.Group("/upload")
			uploadV1.Use(authMw.RequireAuth())
			{
				uploadV1.POST("/image", uploadHandler.GetPresignedURL)
				uploadV1.DELETE("/*key", uploadHandler.DeleteFile)
			}
		}

		// Team Plans
		{
			teamPlanRepo := teamplanrepository.NewRepository(db)
			teamPlanSvc := teamplanservice.NewService(teamPlanRepo)
			teamPlanCtrl := teamplancontroller.NewController(teamPlanSvc)
			teamPlansV1 := apiV1.Group("/team-plans")
			teamPlansV1.Use(authMw.RequireAuth())
			teamPlansV1.GET("", middleware.CachePrivate(), teamPlanCtrl.List)
			teamPlansV1.GET("/:id", middleware.CachePrivate(), teamPlanCtrl.GetByID)
			teamPlansV1.POST("", teamPlanCtrl.Create)
			teamPlansV1.PUT("/:id", teamPlanCtrl.Update)
			teamPlansV1.DELETE("/:id", teamPlanCtrl.Delete)
		}

		// Planning Calendar
		{
			teamPlanRepo := teamplanrepository.NewRepository(db)
			monthlyPlanRepo := monthlyplanrepository.NewRepository(db)
			lwRepo := largeworkrepository.NewRepository(db)
			planningSvc := planningcalendarservice.NewService(teamPlanRepo, monthlyPlanRepo, lwRepo, time.Now)
			planningCtrl := planningcalendarcontroller.NewController(planningSvc)
			planningV1 := apiV1.Group("/planning")
			planningV1.Use(authMw.RequireAuth())
			planningV1.GET("/calendar/:year/:month", middleware.CachePrivate(), planningCtrl.GetMonth)
		}

		// Large Work Items (งานระดมทีม)
		{
			lwRepo := largeworkrepository.NewRepository(db)
			lwDailyReports := largeworkrepository.NewDailyReportCreator(db)
			lwSvc := largeworkservice.NewServiceWithDailyReport(lwRepo, lwDailyReports)
			lwCtrl := largeworkcontroller.NewController(lwSvc)
			largeWorksV1 := apiV1.Group("/large-work-items")
			largeWorksV1.Use(authMw.RequireAuth())
			largeWorksV1.GET("", middleware.CachePrivate(), lwCtrl.List)
			largeWorksV1.GET("/:id", middleware.CachePrivate(), lwCtrl.GetByID)
			largeWorksV1.GET("/:id/overview", middleware.CachePrivate(), lwCtrl.GetOverview)
			largeWorksV1.GET("/:id/tasks", middleware.CachePrivate(), lwCtrl.ListTasks)
			largeWorksV1.PUT("/:id/tasks", lwCtrl.ReplaceTasks)
			largeWorksV1.POST("/:id/tasks", lwCtrl.ReplaceTasks)
			largeWorksTasksV1 := apiV1.Group("/large-work-tasks")
			largeWorksTasksV1.Use(authMw.RequireAuth())
			largeWorksTasksV1.GET("/my-todos", middleware.CachePrivate(), lwCtrl.MyTodos)
			largeWorksTasksV1.PATCH("/:taskId/start", lwCtrl.StartTask)
			largeWorksTasksV1.POST("/:taskId/photos", lwCtrl.AttachTaskPhoto)
			largeWorksTasksV1.PATCH("/:taskId/complete", lwCtrl.CompleteTask)
			largeWorksTasksV1.PATCH("/:taskId/block", lwCtrl.BlockTask)
			largeWorksV1.POST("", lwCtrl.Create)
			largeWorksV1.PATCH("/:id", lwCtrl.Update)
			largeWorksV1.POST("/:id/cancel", lwCtrl.Cancel)
			largeWorksV1.POST("/:id/attachments", lwCtrl.Attachments)
		}

		// Monthly Plan File Management
		monthlyPlanRepo := monthlyplanrepository.NewRepository(db)
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
		var monthlyPlanController *monthlyplancontroller.Controller
		if mpR2Err == nil {
			r2Storage := monthlyplanrepository.NewR2Storage(r2Client)
			monthlyPlanService := monthlyplanservice.NewService(monthlyPlanRepo, r2Storage)
			monthlyPlanController = monthlyplancontroller.NewController(monthlyPlanService)
		}
		if monthlyPlanController != nil {
			monthlyPlansV1 := apiV1.Group("/monthly-plans")
			monthlyPlansV1.Use(authMw.RequireAuth(), func(c *gin.Context) {
				if uid, ok := c.Get("user_id"); ok {
					if userID, ok := uid.(uint); ok {
						if caps, err := capabilityRepo.ListCodesByUserID(c.Request.Context(), userID); err == nil {
							c.Set("capabilities", caps)
						}
					}
				}
				c.Next()
			})
			{
				// Settings (admin only)
				adminOnly := monthlyPlansV1.Group("")
				adminOnly.Use(authMw.RequireRole(monthlyPlanManagerRoles()...))
				{
					adminOnly.GET("/settings", monthlyPlanController.GetSettings)
					adminOnly.PUT("/settings", monthlyPlanController.UpdateSettings)
					adminOnly.DELETE("/files/:id/permanent", monthlyPlanController.HardDeleteFile)
					adminOnly.POST("/files/:id/restore", monthlyPlanController.RestoreFile)
				}

				// Period endpoints (all authenticated users)
				monthlyPlansV1.GET("/:year/overview", middleware.CachePrivate(60), monthlyPlanController.GetYearOverview)
				monthlyPlansV1.GET("/:year/:month", middleware.CachePrivate(60), monthlyPlanController.GetOrCreatePeriod)
				monthlyPlansV1.GET("/:year/:month/files", middleware.CachePrivate(60), monthlyPlanController.ListFiles)
				monthlyPlansV1.GET("/:year/:month/status", middleware.CachePrivate(60), monthlyPlanController.GetSubmissionStatus)

				// Approved/monthly-plan conversion into planning projection
				monthlyPlansV1.POST("/:year/:month/convert-to-planning", monthlyPlanController.ConvertApprovedToPlanning)

				// Upload flow (presign + confirm)
				monthlyPlansV1.POST("/:year/:month/files/presign", monthlyPlanController.PresignUpload)
				monthlyPlansV1.POST("/:year/:month/files", monthlyPlanController.ConfirmUpload)

				// File actions
				monthlyPlansV1.DELETE("/files/:id", monthlyPlanController.SoftDeleteFile)
				monthlyPlansV1.GET("/files/:id/download", monthlyPlanController.GetDownloadURL)
			}
		}

		// ── Dashboard Feature (Phase D) ─────────────────────────────────────
		{
			dbRepo := dashboardrepository.NewRepository(db)
			dbSvc := dashboardservice.NewService(dbRepo)
			dbCtrl := dashboardcontroller.NewController(dbSvc)
			dashboardV1 := apiV1.Group("/dashboard")
			dashboardV1.Use(authMw.RequireAuth(), authMw.RequireRole(dashboardReadRoles()...))
			dashboardV1.GET("/summary", middleware.CachePrivate(60), dbCtrl.Summary)
			dashboardV1.GET("/top-jobs", middleware.CachePrivate(60), dbCtrl.TopJobs)
			dashboardV1.GET("/top-feeders", middleware.CachePrivate(60), dbCtrl.TopFeeders)
			dashboardV1.GET("/feeder-matrix", middleware.CachePrivate(60), dbCtrl.FeederMatrix)
			dashboardV1.GET("/stats", middleware.CachePrivate(60), dbCtrl.Stats)
		}

		// Users — no cache (admin-only + user-specific context)
		usersV1 := apiV1.Group("/users")
		{
			if uRepo == nil {
				uRepo = userrepository.NewRepository(db)
			}
			uSvc := userservice.NewService(uRepo)
			uCtrl := usercontroller.NewController(uSvc)
			// Apply authentication middleware to all user routes
			usersV1.Use(authMw.RequireAuth())
			// Contact directory is visible to every authenticated user.
			contactsV1 := apiV1.Group("/contact-directory")
			contactsV1.Use(authMw.RequireAuth())
			{
				contactsV1.GET("", middleware.CachePrivate(), uCtrl.ListContacts)
				contactsV1.GET("/:userId", middleware.CachePrivate(), uCtrl.GetContactByID)
			}

			// Own contact info update is available to every authenticated user.
			usersV1.PATCH("/me/contact", uCtrl.UpdateOwnContact)

			// Admin only routes
			adminUsers := usersV1.Group("")
			adminUsers.Use(authMw.RequireRole(superAdminOnlyRoles()...))
			{
				adminUsers.GET("", middleware.CachePrivate(), uCtrl.List)
				adminUsers.GET("/:id/capabilities", middleware.CachePrivate(), capabilityCtrl.ListForUser)
				adminUsers.PUT("/:id/capabilities", capabilityCtrl.ReplaceForUser)
				adminUsers.GET("/:id", middleware.CachePrivate(), uCtrl.GetByID)
				adminUsers.POST("", uCtrl.Create)
				adminUsers.PUT("/:id", uCtrl.Update)
				adminUsers.PATCH("/:id/contact", uCtrl.UpdateContact)
				adminUsers.PUT("/:id/reset-password", uCtrl.ResetPassword)
				adminUsers.DELETE("/:id", uCtrl.Delete)
			}

			// User can change their own password (authenticated, but not necessarily admin)
			usersV1.PUT("/:id/password", uCtrl.ChangePassword)
		}
	}

	return r
}

func dashboardReadRoles() []string {
	return []string{policy.RoleSuperAdmin, policy.RoleViewer}
}

func superAdminOnlyRoles() []string {
	return []string{policy.RoleSuperAdmin}
}

func monthlyPlanManagerRoles() []string {
	return []string{policy.RoleSuperAdmin}
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
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

package routes

import (
	"wallet-point/internal/audit"
	"wallet-point/internal/auth"
	"wallet-point/internal/external" // Add this
	"wallet-point/internal/marketplace"
	"wallet-point/internal/mission"
	"wallet-point/internal/transfer"
	"wallet-point/internal/user"
	"wallet-point/internal/wallet"
	"wallet-point/middleware"
	"wallet-point/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, allowedOrigins string, jwtExpiry int) {
	// Apply global middleware
	r.Use(middleware.CORS(allowedOrigins))
	r.Use(middleware.Logger())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.IPBasedRateLimiter())

	// Publicly accessible uploaded files
	// Backend handler saves to ./public/uploads, so we must serve from there
	r.Static("/uploads", "./public/uploads")

	// API v1 group
	api := r.Group("/api/v1")

	// Global Upload Endpoint
	api.POST("/upload", middleware.AuthMiddleware(), utils.HandleFileUpload)

	// Initialize repositories
	authRepo := auth.NewAuthRepository(db)
	userRepo := user.NewUserRepository(db)
	walletRepo := wallet.NewWalletRepository(db)
	marketplaceRepo := marketplace.NewMarketplaceRepository(db)
	auditRepo := audit.NewAuditRepository(db)
	missionRepo := mission.NewMissionRepository(db)
	transferRepo := transfer.NewRepository(db)
	externalRepo := external.NewRepository(db) // Add this

	// Initialize services
	authService := auth.NewAuthService(authRepo, jwtExpiry)
	userService := user.NewUserService(userRepo)
	walletService := wallet.NewWalletService(walletRepo, db)
	marketplaceService := marketplace.NewMarketplaceService(marketplaceRepo, walletService, db)
	auditService := audit.NewAuditService(auditRepo)
	missionService := mission.NewMissionService(missionRepo, walletService, db)
	transferService := transfer.NewService(transferRepo, walletRepo, walletService, db)
	externalService := external.NewService(externalRepo, walletRepo, walletService, marketplaceService, missionService, auditService, db) // Add this

	// Initialize handlers
	authHandler := auth.NewAuthHandler(authService, auditService)
	userHandler := user.NewUserHandler(userService, auditService)
	walletHandler := wallet.NewWalletHandler(walletService, auditService)
	marketplaceHandler := marketplace.NewMarketplaceHandler(marketplaceService, auditService)
	auditHandler := audit.NewAuditHandler(auditService)
	missionHandler := mission.NewMissionHandler(missionService, auditService)
	transferHandler := transfer.NewHandler(transferService, auditService)
	externalHandler := external.NewHandler(externalService, auditService) // Add this

	// ========================================
	// PUBLIC ROUTES
	// ========================================
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", middleware.AuthRateLimiter(), authHandler.Login)
		authGroup.GET("/me", middleware.AuthMiddleware(), authHandler.Me)
		authGroup.PUT("/profile", middleware.AuthMiddleware(), authHandler.UpdateProfile)
		authGroup.PUT("/password", middleware.AuthMiddleware(), authHandler.UpdatePassword)
	}

	// ========================================
	// ADMIN ROUTES
	// ========================================
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware())
	adminGroup.Use(middleware.RoleMiddleware("admin"))
	{
		// User Management
		adminGroup.POST("/users", authHandler.Register) // Create new user
		adminGroup.GET("/users", userHandler.GetAll)
		adminGroup.GET("/users/:id", userHandler.GetByID)
		adminGroup.PUT("/users/:id", userHandler.Update)
		adminGroup.DELETE("/users/:id", userHandler.Deactivate)
		adminGroup.PUT("/users/:id/password", userHandler.ChangePassword)

		// Wallet Management
		adminGroup.GET("/wallets", walletHandler.GetAllWallets)
		adminGroup.GET("/wallets/:id", walletHandler.GetWalletByID)
		adminGroup.GET("/wallets/:id/transactions", walletHandler.GetWalletTransactions)
		adminGroup.POST("/wallet/adjustment", walletHandler.AdjustPoints)
		adminGroup.POST("/wallet/reset", walletHandler.ResetWallet)

		// Transaction Monitoring
		adminGroup.GET("/transactions", walletHandler.GetAllTransactions)
		adminGroup.GET("/transfers", transferHandler.GetAllTransfers)

		// Marketplace Management
		adminGroup.GET("/marketplace/transactions", marketplaceHandler.GetTransactions) // Add this
		adminGroup.GET("/products", marketplaceHandler.GetAll)
		adminGroup.POST("/products", marketplaceHandler.Create)
		adminGroup.GET("/products/:id", marketplaceHandler.GetByID)
		adminGroup.PUT("/products/:id", marketplaceHandler.Update)
		adminGroup.DELETE("/products/:id", marketplaceHandler.Delete)

		// Audit Logs
		adminGroup.GET("/audit-logs", auditHandler.GetAll)

		// External Sources Management
		adminGroup.GET("/external/sources", externalHandler.ListSources)
		adminGroup.POST("/external/sources", externalHandler.RegisterSource)
		adminGroup.POST("/external/products", externalHandler.RegisterProduct)
		adminGroup.POST("/external/missions", externalHandler.RegisterMission)

		// Admin Dashboard Stats
		adminGroup.GET("/stats", walletHandler.GetAdminStats)
	}

	// ========================================
	// DOSEN ROUTES
	// ========================================
	dosenGroup := api.Group("/dosen")
	dosenGroup.Use(middleware.AuthMiddleware())
	dosenGroup.Use(middleware.RoleMiddleware("dosen", "admin"))
	{
		// Mission & Task Management
		dosenGroup.POST("/missions", missionHandler.CreateMission)
		dosenGroup.PUT("/missions/:id", missionHandler.UpdateMission)
		dosenGroup.DELETE("/missions/:id", missionHandler.DeleteMission)
		dosenGroup.GET("/missions", missionHandler.GetAllMissions)
		dosenGroup.GET("/missions/:id", missionHandler.GetMissionByID)

		// Submission Validation
		dosenGroup.GET("/submissions", missionHandler.GetAllSubmissions)
		dosenGroup.POST("/submissions/:id/review", missionHandler.ReviewSubmission)

		// Dashboard Stats
		dosenGroup.GET("/stats", missionHandler.GetDosenStats)

		// Monitoring & Manual Rewards
		dosenGroup.GET("/students", userHandler.GetAll) // Reuse GetAll but restricted to Dosen
		dosenGroup.POST("/reward", walletHandler.AdjustPoints)
	}

	// ========================================
	// MAHASISWA ROUTES
	// ========================================
	mahasiswaGroup := api.Group("/mahasiswa")
	mahasiswaGroup.Use(middleware.AuthMiddleware())
	mahasiswaGroup.Use(middleware.RoleMiddleware("mahasiswa"))
	{
		// Mission & Task Submission
		mahasiswaGroup.GET("/missions", missionHandler.GetAllMissions)
		mahasiswaGroup.GET("/missions/:id", missionHandler.GetMissionByID)
		mahasiswaGroup.POST("/missions/submit", missionHandler.SubmitMission)
		mahasiswaGroup.GET("/submissions", missionHandler.GetAllSubmissions)

		// Transfer Points
		mahasiswaGroup.POST("/transfer", transferHandler.CreateTransfer)
		mahasiswaGroup.GET("/transfer/history", transferHandler.GetMyTransfers)
		mahasiswaGroup.GET("/transfer/recipient/:id", transferHandler.GetRecipientInfo)
		mahasiswaGroup.GET("/transfer/sent", transferHandler.GetSentTransfers)
		mahasiswaGroup.GET("/transfer/received", transferHandler.GetReceivedTransfers)
		mahasiswaGroup.GET("/users/lookup", userHandler.LookupUser) // Lookup user for transfer verification

		// Marketplace Purchase
		mahasiswaGroup.POST("/marketplace/purchase", marketplaceHandler.Purchase)
		mahasiswaGroup.GET("/marketplace/products", marketplaceHandler.GetAll) // Reuse GetAll, maybe add status filter later
		mahasiswaGroup.GET("/marketplace/products/:id", marketplaceHandler.GetByID)

		// Gamification
		mahasiswaGroup.GET("/leaderboard", walletHandler.GetLeaderboard)

		// Personal Wallet
		mahasiswaGroup.GET("/wallet", walletHandler.GetMyWallet)
		mahasiswaGroup.GET("/transactions", walletHandler.GetMyTransactions) // Replaces old getTransactions use case
		mahasiswaGroup.POST("/payment/token", walletHandler.GeneratePaymentToken)
		mahasiswaGroup.POST("/payment/execute", walletHandler.ExecuteStudentPayment)

		// External Point Sync
		mahasiswaGroup.POST("/external/sync", externalHandler.SyncPoints)
	}

	// ========================================
	// MERCHANT ROUTES
	// ========================================
	merchantGroup := api.Group("/merchant")
	merchantGroup.Use(middleware.AuthMiddleware())
	merchantGroup.Use(middleware.RoleMiddleware("merchant", "admin"))
	{
		merchantGroup.POST("/payment/scan", walletHandler.MerchantScan)
		merchantGroup.GET("/stats", walletHandler.GetMerchantStats)
	}

	// Global QR Status Check
	api.GET("/payment/status/:token", walletHandler.CheckTokenStatus)

	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Wallet Point API is running",
		})
	})
}

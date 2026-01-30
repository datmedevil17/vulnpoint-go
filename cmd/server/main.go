package main

import (
	"log"

	"github.com/datmedevil17/go-vuln/internal/config"
	"github.com/datmedevil17/go-vuln/internal/database"
	"github.com/datmedevil17/go-vuln/internal/handlers"
	"github.com/datmedevil17/go-vuln/internal/middleware"
	"github.com/datmedevil17/go-vuln/internal/routes"
	"github.com/datmedevil17/go-vuln/internal/services"
	"github.com/datmedevil17/go-vuln/internal/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connections
	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	redisClient, err := database.NewRedis(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize utilities
	jwtUtil := utils.NewJWTUtil(cfg.JWT.Secret, cfg.JWT.Expiration)

	// Initialize services
	authService := services.NewAuthService(db, cfg)
	scannerService := services.NewScannerService(db)
	notificationService := services.NewNotificationService(cfg)
	aiService := services.NewAIService(cfg)
	githubService := services.NewGitHubService(db)
	workflowService := services.NewWorkflowService(db, scannerService, notificationService, aiService, githubService)
	embeddingService := services.NewEmbeddingService()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, jwtUtil, cfg)
	workflowHandler := handlers.NewWorkflowHandler(workflowService)
	githubHandler := handlers.NewGitHubHandler(githubService, authService)
	scannerHandler := handlers.NewScannerHandler(scannerService)
	codeHandler := handlers.NewCodeHandler(aiService, embeddingService)
	chatbotHandler := handlers.NewChatbotHandler(aiService)

	// Create Gin router
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.CORSMiddleware(cfg))
	router.Use(middleware.LoggerMiddleware())
	router.Use(gin.Recovery())

	// Setup routes
	routes.SetupRoutes(router, &routes.RouterConfig{
		AuthHandler:     authHandler,
		WorkflowHandler: workflowHandler,
		GitHubHandler:   githubHandler,
		ScannerHandler:  scannerHandler,
		CodeHandler:     codeHandler,
		ChatbotHandler:  chatbotHandler,
		JWTUtil:         jwtUtil,
	})

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("🚀 VulnPilot server starting on %s", addr)
	log.Printf("📝 Mode: %s", cfg.Server.Mode)
	log.Printf("🔒 CORS Origins: %v", cfg.Frontend.CORSOrigins)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

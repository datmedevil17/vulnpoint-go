package routes

import (
	"github.com/datmedevil17/go-vuln/internal/handlers"
	"github.com/datmedevil17/go-vuln/internal/middleware"
	"github.com/datmedevil17/go-vuln/internal/utils"
	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	AuthHandler       *handlers.AuthHandler
	WorkflowHandler   *handlers.WorkflowHandler
	GitHubHandler     *handlers.GitHubHandler
	ScannerHandler    *handlers.ScannerHandler
	CodeHandler       *handlers.CodeHandler
	ChatbotHandler    *handlers.ChatbotHandler
	AIWorkflowHandler *handlers.AIWorkflowHandler
	JWTUtil           *utils.JWTUtil
}

func SetupRoutes(router *gin.Engine, cfg *RouterConfig) {
	// Health check
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "vulnpilot-go",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.GET("/github", cfg.AuthHandler.GetAuthURL)
			auth.GET("/github/callback", cfg.AuthHandler.HandleCallback)
			auth.POST("/logout", cfg.AuthHandler.Logout)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTUtil))
		{
			// AI Workflow Generation
			protected.POST("/workflow/ai-generate", cfg.AIWorkflowHandler.GenerateWorkflow)

			// User routes
			protected.GET("/user", cfg.AuthHandler.GetCurrentUser)

			// Workflow routes
			workflows := protected.Group("/workflows")
			{
				workflows.POST("", cfg.WorkflowHandler.CreateWorkflow)
				workflows.GET("", cfg.WorkflowHandler.ListWorkflows)
				workflows.GET("/executions/:id", cfg.WorkflowHandler.GetExecution)
				workflows.GET("/reports", cfg.WorkflowHandler.ListWorkflowExecutions)
				workflows.DELETE("/reports/:id", cfg.WorkflowHandler.DeleteWorkflowExecution)
				workflows.GET("/:id", cfg.WorkflowHandler.GetWorkflow)
				workflows.PUT("/:id", cfg.WorkflowHandler.UpdateWorkflow)
				workflows.DELETE("/:id", cfg.WorkflowHandler.DeleteWorkflow)
				workflows.POST("/:id/execute", cfg.WorkflowHandler.ExecuteWorkflow)
			}

			// GitHub routes
			github := protected.Group("/github")
			{
				github.GET("/repositories", cfg.GitHubHandler.ListRepositories)
				github.GET("/repositories/:owner/:repo/files", cfg.GitHubHandler.GetRepositoryFiles)
				github.GET("/repositories/:owner/:repo/content", cfg.GitHubHandler.GetFileContent)
			}

			// Scanner routes
			scan := protected.Group("/scan")
			{
				scan.POST("/nmap", cfg.ScannerHandler.NmapScan)
				scan.POST("/nikto", cfg.ScannerHandler.NiktoScan)
				scan.POST("/gobuster", cfg.ScannerHandler.GobusterScan)
				scan.GET("/results", cfg.ScannerHandler.ListScanResults)
				scan.GET("/results/:id", cfg.ScannerHandler.GetScanResult)
			}

			// Code analysis routes
			code := protected.Group("/code")
			{
				code.POST("/analyze", cfg.CodeHandler.AnalyzeCode)
				code.POST("/quick-scan", cfg.CodeHandler.QuickScan)
				code.POST("/compare", cfg.CodeHandler.CompareCode)
			}

			// Chatbot routes
			chatbot := protected.Group("/chatbot")
			{
				chatbot.POST("/chat", cfg.ChatbotHandler.Chat)
				chatbot.POST("/explain", cfg.ChatbotHandler.ExplainVulnerability)
				chatbot.POST("/remediate", cfg.ChatbotHandler.SuggestRemediation)
				chatbot.POST("/ask", cfg.ChatbotHandler.AskSecurityQuestion)
			}
		}
	}
}

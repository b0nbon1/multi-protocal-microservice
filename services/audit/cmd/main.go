package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"audit-service/internal/database"
	"audit-service/internal/handlers"
	"audit-service/shared/logger"
	"audit-service/shared/middleware"
)

func main() {
	// Initialize logger
	logger.Init("audit-service")
	defer logger.Sync()

	// Connect to MongoDB
	if err := database.Connect(); err != nil {
		zap.L().Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer database.Close()

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize handlers
	auditHandler := handlers.NewAuditHandler()

	// Setup router
	router := gin.New()

	// Middleware
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandler())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	}))

	// Routes
	api := router.Group("/api/v1")
	{
		api.POST("/logs", auditHandler.CreateLog)
		api.GET("/logs", auditHandler.GetAllLogs)
		api.GET("/logs/user/:userId", auditHandler.GetUserLogs)
		api.GET("/logs/analytics", auditHandler.GetLogAnalytics)
		api.GET("/logs/search", auditHandler.SearchLogs)
	}

	// Health check
	router.GET("/health", auditHandler.HealthCheck)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	// Validate port
	if _, err := strconv.Atoi(port); err != nil {
		zap.L().Fatal("Invalid port number", zap.String("port", port))
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		zap.L().Info("Starting Audit Service",
			zap.String("port", port),
			zap.String("env", os.Getenv("GIN_MODE")),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server forced to shutdown", zap.Error(err))
	}

	zap.L().Info("Server exited")
}

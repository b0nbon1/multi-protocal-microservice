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

	"api-gateway/internal/handlers"
	authMiddleware "api-gateway/internal/middleware"
	"api-gateway/internal/proxy"
	"api-gateway/shared/logger"
	"api-gateway/shared/middleware"
)

func main() {
	// Initialize logger
	logger.Init("api-gateway")
	defer logger.Sync()

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Service configuration
	serviceConfig := proxy.ServiceConfig{
		AuthServiceURL:         getEnv("AUTH_SERVICE_URL", "http://localhost:3001"),
		ContractServiceURL:     getEnv("CONTRACT_SERVICE_URL", "http://localhost:3002"),
		PaymentServiceURL:      getEnv("PAYMENT_SERVICE_URL", "http://localhost:3003"),
		DisputeServiceURL:      getEnv("DISPUTE_SERVICE_URL", "http://localhost:3004"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8081"),
		AuditServiceURL:        getEnv("AUDIT_SERVICE_URL", "http://localhost:8082"),
	}

	// Initialize handlers
	proxyHandler := proxy.NewProxyHandler(serviceConfig)
	healthHandler := handlers.NewHealthHandler()

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

	// Health check endpoint
	router.GET("/api/v1/health", healthHandler.HealthCheck)

	// Auth routes (no authentication required)
	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.POST("/register", proxyHandler.ProxyRequest(serviceConfig.AuthServiceURL+"/api/v1"))
		authGroup.POST("/login", proxyHandler.ProxyRequest(serviceConfig.AuthServiceURL+"/api/v1"))
	}

	// Protected routes (authentication required)
	protected := router.Group("/api/v1")
	protected.Use(authMiddleware.AuthMiddleware())
	{
		// User routes
		protected.GET("/users/*path", proxyHandler.ProxyRequest(serviceConfig.AuthServiceURL+"/api/v1"))

		// Contract routes
		protected.POST("/contracts", proxyHandler.ProxyRequest(serviceConfig.ContractServiceURL+"/api/v1"))
		protected.GET("/contracts/*path", proxyHandler.ProxyRequest(serviceConfig.ContractServiceURL+"/api/v1"))
		protected.PUT("/contracts/*path", proxyHandler.ProxyRequest(serviceConfig.ContractServiceURL+"/api/v1"))
		protected.DELETE("/contracts/*path", proxyHandler.ProxyRequest(serviceConfig.ContractServiceURL+"/api/v1"))

		// Payment routes
		protected.GET("/wallets/*path", proxyHandler.ProxyRequest(serviceConfig.PaymentServiceURL+"/api/v1"))
		protected.POST("/wallets/*path", proxyHandler.ProxyRequest(serviceConfig.PaymentServiceURL+"/api/v1"))
		protected.POST("/transfers", proxyHandler.ProxyRequest(serviceConfig.PaymentServiceURL+"/api/v1"))
		protected.GET("/transactions/*path", proxyHandler.ProxyRequest(serviceConfig.PaymentServiceURL+"/api/v1"))

		// Dispute routes
		protected.POST("/disputes", proxyHandler.ProxyRequest(serviceConfig.DisputeServiceURL+"/api/v1"))
		protected.GET("/disputes/*path", proxyHandler.ProxyRequest(serviceConfig.DisputeServiceURL+"/api/v1"))
		protected.PUT("/disputes/*path", proxyHandler.ProxyRequest(serviceConfig.DisputeServiceURL+"/api/v1"))
		protected.DELETE("/disputes/*path", proxyHandler.ProxyRequest(serviceConfig.DisputeServiceURL+"/api/v1"))

		// Notification routes
		protected.POST("/notifications/*path", proxyHandler.ProxyRequest(serviceConfig.NotificationServiceURL+"/api/v1"))
		protected.GET("/notifications/*path", proxyHandler.ProxyRequest(serviceConfig.NotificationServiceURL+"/api/v1"))

		// Audit routes (admin only in real implementation)
		protected.GET("/audit/*path", proxyHandler.ProxyRequest(serviceConfig.AuditServiceURL+"/api/v1"))
	}

	// WebSocket routes (no auth middleware for simplicity)
	router.GET("/ws", proxyHandler.ProxyRequest(serviceConfig.NotificationServiceURL))

	// Get port from environment
	port := getEnv("PORT", "8080")

	// Validate port
	if _, err := strconv.Atoi(port); err != nil {
		zap.L().Fatal("Invalid port number", zap.String("port", port))
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		zap.L().Info("Starting API Gateway",
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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}


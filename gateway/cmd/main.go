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

	grpcClients "api-gateway/internal/grpc"
	"api-gateway/internal/handlers"
	authMiddleware "api-gateway/internal/middleware"
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

	// gRPC Service configuration
	grpcConfig := grpcClients.GRPCConfig{
		AuthServiceAddr:         getEnv("AUTH_GRPC_ADDR", "localhost:50051"),
		ContractServiceAddr:     getEnv("CONTRACT_GRPC_ADDR", "localhost:50052"),
		PaymentServiceAddr:      getEnv("PAYMENT_GRPC_ADDR", "localhost:50053"),
		DisputeServiceAddr:      getEnv("DISPUTE_GRPC_ADDR", "localhost:50054"),
		NotificationServiceAddr: getEnv("NOTIFICATION_GRPC_ADDR", "localhost:50055"),
		AuditServiceAddr:        getEnv("AUDIT_GRPC_ADDR", "localhost:50056"),
	}

	// Initialize gRPC clients
	grpcClientManager, err := grpcClients.NewGRPCClients(grpcConfig)
	if err != nil {
		zap.L().Fatal("Failed to initialize gRPC clients", zap.Error(err))
	}
	defer grpcClientManager.Close()

	// Initialize handlers
	grpcProxyHandler := grpcClients.NewGRPCProxyHandler(grpcClientManager)
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
		authGroup.POST("/register", grpcProxyHandler.Register)
		authGroup.POST("/login", grpcProxyHandler.Login)
		authGroup.POST("/validate", grpcProxyHandler.ValidateToken)
	}

	// Protected routes (authentication required)
	protected := router.Group("/api/v1")
	protected.Use(authMiddleware.AuthMiddleware())
	{
		// User routes
		protected.GET("/users/:userId", grpcProxyHandler.GetUser)

		// Contract routes
		protected.POST("/contracts", grpcProxyHandler.CreateContract)
		protected.GET("/contracts", grpcProxyHandler.GetContracts)
		protected.GET("/contracts/:contractId", grpcProxyHandler.GetContract)

		// Payment routes
		protected.POST("/wallets", grpcProxyHandler.CreateWallet)
		protected.POST("/transfers", grpcProxyHandler.CreateTransfer)

		// Dispute routes
		protected.POST("/disputes", grpcProxyHandler.CreateDispute)

		// Notification routes
		protected.GET("/notifications", grpcProxyHandler.GetNotifications)
		protected.PUT("/notifications/:notificationId/read", grpcProxyHandler.MarkNotificationAsRead)

		// Audit routes (admin only in real implementation)
		protected.GET("/audit/logs", grpcProxyHandler.GetAuditLogs)
	}

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

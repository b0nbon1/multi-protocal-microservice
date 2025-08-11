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

	"notification-service/internal/handlers"
	"notification-service/internal/websocket"
	"notification-service/shared/logger"
	"notification-service/shared/middleware"
)

func main() {
	// Initialize logger
	logger.Init("notification-service")
	defer logger.Sync()

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize handlers
	notificationHandler := handlers.NewNotificationHandler(hub)

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
		api.POST("/notify", notificationHandler.SendNotification)
		api.GET("/connections", notificationHandler.GetActiveConnections)
	}

	// WebSocket endpoint
	router.GET("/ws", hub.ServeWS)

	// Health check
	router.GET("/health", notificationHandler.HealthCheck)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
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
		zap.L().Info("Starting Notification Service",
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


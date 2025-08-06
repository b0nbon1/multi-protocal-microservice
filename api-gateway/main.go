package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

type Claims struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// WebSocket connections
	wsConnections = make(map[*websocket.Conn]bool)
	config        Config
)

func main() {
	config = LoadConfig()
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "api-gateway",
		})
	})

	// WebSocket endpoint
	r.GET("/ws", handleWebSocket)

	// Broadcast endpoint
	r.POST("/api/broadcast", func(c *gin.Context) {
		var message map[string]interface{}
		if err := c.ShouldBindJSON(&message); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		broadcastToWebSockets(message)
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Service routing
	userServiceURL, _ := url.Parse(getEnv("USER_SERVICE_URL", "http://user-service:3001"))
	orderServiceURL, _ := url.Parse(getEnv("ORDER_SERVICE_URL", "http://order-service:3002"))

	// Authentication routes (no auth required)
	authGroup := r.Group("/api/auth")
	authGroup.Any("/register", reverseProxy(userServiceURL))
	authGroup.Any("/login", reverseProxy(userServiceURL))

	// Protected user routes (auth required)
	userGroup := r.Group("/api/users")
	userGroup.Use(authMiddleware())
	userGroup.Any("/*path", reverseProxy(userServiceURL))

	// Protected order routes
	orderGroup := r.Group("/api/orders")
	orderGroup.Use(authMiddleware())
	orderGroup.Any("/*path", reverseProxy(orderServiceURL))

	// GraphQL endpoint for analytics
	analyticsURL, _ := url.Parse("http://analytics-service:4000")
	r.Any("/graphql", reverseProxy(analyticsURL))

	log.Println("API Gateway starting on port 3800")
	log.Fatal(r.Run(":3800"))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

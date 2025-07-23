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
    UserID int    `json:"userId"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

var (
    jwtSecret = []byte(getEnv("JWT_SECRET", "your-secret-key"))
    upgrader  = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            return true // Allow all origins in development
        },
    }
    
    // WebSocket connections
    wsConnections = make(map[*websocket.Conn]bool)
)

func main() {
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

    // User service routes (no auth required for login/register)
    r.Any("/api/users/register", reverseProxy(userServiceURL))
    r.Any("/api/users/login", reverseProxy(userServiceURL))
    
    // Protected user routes
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

    log.Println("API Gateway starting on port 3000")
    log.Fatal(r.Run(":3000"))
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
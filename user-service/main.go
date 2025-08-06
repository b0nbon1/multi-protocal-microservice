package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

var (
	db          *sql.DB
	redisClient *redis.Client
	config      Config
	// jwtSecret   = []byte(getEnv("JWT_SECRET", "your-secret-key"))
)

func main() {
	// Initialize database
	var err error
	config = LoadConfig()
	db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName, config.DBSSLMode))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient = redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_URL", "redis://localhost:6379")[8:], // Remove redis:// prefix
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "user-service",
		})
	})

	// Routes
	r.POST("/register", registerHandler)
	r.POST("/login", loginHandler)
	r.GET("/users/:id", getUserHandler)

	log.Println("User Service starting on port 3001")
	log.Fatal(r.Run(":3001"))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

package main

import (
    "database/sql"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    _ "github.com/lib/pq"
)

var (
    db        *sql.DB
    redisClient *redis.Client
    jwtSecret = []byte(getEnv("JWT_SECRET", "your-secret-key"))
)

func main() {
    // Initialize database
    var err error
    db, err = sql.Open("postgres", getEnv("DATABASE_URL", "postgres://admin:admin123@postgres:5432/microservices_db?sslmode=disable"))
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

    // Initialize Redis
    redisClient = redis.NewClient(&redis.Options{
        Addr: getEnv("REDIS_URL", "redis://localhost:6379")[8:], // Remove redis:// prefix
    })

    // Create tables
    initDatabase()

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

func initDatabase() {
    query := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        email VARCHAR(255) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        name VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`

    if _, err := db.Exec(query); err != nil {
        log.Fatal("Failed to create users table:", err)
    }

    log.Println("Database initialized successfully")
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

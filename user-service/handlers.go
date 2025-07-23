package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
    ID        int       `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    Name      string    `json:"name" db:"name"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type UserWithPassword struct {
    User
    PasswordHash string `db:"password_hash"`
}

type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}

type Claims struct {
    UserID int    `json:"userId"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func registerHandler(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    // Insert user
    var user User
    query := `
        INSERT INTO users (email, password_hash, name) 
        VALUES ($1, $2, $3) 
        RETURNING id, email, name, created_at`

    err = db.QueryRow(query, req.Email, string(hashedPassword), req.Name).
        Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)

    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists or database error"})
        return
    }

    // Cache user data
    cacheUser(user)

    c.JSON(http.StatusCreated, user)
}

func loginHandler(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Get user from database
    var userWithPassword UserWithPassword
    query := `SELECT id, email, name, password_hash, created_at FROM users WHERE email = $1`

    err := db.QueryRow(query, req.Email).Scan(
        &userWithPassword.ID,
        &userWithPassword.Email,
        &userWithPassword.Name,
        &userWithPassword.PasswordHash,
        &userWithPassword.CreatedAt,
    )

    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(userWithPassword.PasswordHash), []byte(req.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    // Generate JWT token
    token, err := generateToken(userWithPassword.User)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, LoginResponse{
        Token: token,
        User:  userWithPassword.User,
    })
}

func getUserHandler(c *gin.Context) {
    userID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    // Try cache first
    cacheKey := "user:" + strconv.Itoa(userID)
    cached, err := redisClient.Get(context.Background(), cacheKey).Result()
    if err == nil {
        var user User
        if json.Unmarshal([]byte(cached), &user) == nil {
            c.JSON(http.StatusOK, user)
            return
        }
    }

    // Query database
    var user User
    query := `SELECT id, email, name, created_at FROM users WHERE id = $1`

    err = db.QueryRow(query, userID).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    // Cache the result
    cacheUser(user)

    c.JSON(http.StatusOK, user)
}

func cacheUser(user User) {
    data, _ := json.Marshal(user)
    redisClient.Set(context.Background(), "user:"+strconv.Itoa(user.ID), data, time.Hour)
}

func generateToken(user User) (string, error) {
    claims := Claims{
        UserID: user.ID,
        Email:  user.Email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

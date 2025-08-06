package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
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

	supabaseKey := config.SupabaseKey
	supabaseURL := config.SupabaseURL

	client := resty.New()
	resp, err := client.R().
		SetHeader("apikey", supabaseKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"email":    req.Email,
			"password": req.Password,
		}).
		Post(supabaseURL + "/auth/v1/signup")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Signup Response:", resp)
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	supabaseKey := config.SupabaseKey
	supabaseURL := config.SupabaseURL

	client := resty.New()
	resp, err := client.R().
		SetHeader("apikey", supabaseKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"email":    req.Email,
			"password": req.Password,
		}).
		Post(supabaseURL + "/auth/v1/token?grant_type=password")

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Login Response:", resp)
	fmt.Println("Status Code:", resp.StatusCode())
	fmt.Println("Body:", resp.String())

	// Parse the response body as JSON
	var authResponse map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &authResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse authentication response"})
		return
	}

	// Return the parsed JSON response
	c.JSON(http.StatusOK, gin.H{"message": "User logged in successfully", "access_token": authResponse["access_token"]})
}

func getUserHandler(c *gin.Context) {
	userID := c.Param("id")

	// Try cache first
	cacheKey := "user:" + userID
	cached, err := redisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var user User
		if json.Unmarshal([]byte(cached), &user) == nil {
			c.JSON(http.StatusOK, user)
			return
		}
	}

	client := resty.New()
	resp, err := client.R().
		SetHeader("apikey", config.SupabaseAdminKey).
		SetHeader("Authorization", "Bearer "+config.SupabaseAdminKey).
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("%s/auth/v1/admin/users/%s", config.SupabaseURL, userID))

	if err != nil {
		log.Fatal(err)
	}

    // Parse the response body as JSON
	var userResponse map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &userResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse authentication response"})
		return
	}

	fmt.Println("Status:", resp.Status())
	fmt.Println("Body:", resp.String())
	cacheUser(userID, userResponse)

	c.JSON(http.StatusOK, userResponse)
}

func cacheUser(userID string, user map[string]interface{}) {
	data, _ := json.Marshal(user)
	redisClient.Set(context.Background(), "user:"+userID, data, time.Hour)
}


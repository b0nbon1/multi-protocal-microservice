package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

type HealthCheck struct {
	Service      string            `json:"service"`
	Status       string            `json:"status"`
	Timestamp    string            `json:"timestamp"`
	Version      string            `json:"version,omitempty"`
	Uptime       int64             `json:"uptime,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
	})
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Error:   "BAD_REQUEST",
		Message: message,
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, APIResponse{
		Success: false,
		Error:   "UNAUTHORIZED",
		Message: message,
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, APIResponse{
		Success: false,
		Error:   "NOT_FOUND",
		Message: message,
	})
}

func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Error:   "INTERNAL_SERVER_ERROR",
		Message: message,
	})
}

func Health(c *gin.Context, service string, version string, uptime int64) {
	c.JSON(http.StatusOK, HealthCheck{
		Service:   service,
		Status:    "healthy",
		Timestamp: "2024-01-01T00:00:00Z", // This should be current timestamp
		Version:   version,
		Uptime:    uptime,
	})
}

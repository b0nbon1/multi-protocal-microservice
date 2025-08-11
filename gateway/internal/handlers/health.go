package handlers

import (
	"time"

	"api-gateway/shared/response"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	uptime := time.Since(startTime)

	response.Health(c, "api-gateway", "1.0.0", int64(uptime.Seconds()))
}

var startTime = time.Now()


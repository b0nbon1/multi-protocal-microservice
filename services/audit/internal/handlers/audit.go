package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"audit-service/internal/database"
	"audit-service/internal/models"
	"audit-service/shared/response"
)

type AuditHandler struct{}

func NewAuditHandler() *AuditHandler {
	return &AuditHandler{}
}

func (h *AuditHandler) CreateLog(c *gin.Context) {
	var req models.CreateLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("Invalid log request", zap.Error(err))
		response.BadRequest(c, "Invalid request payload")
		return
	}

	log := &models.Log{
		UserID:    req.UserID,
		Action:    req.Action,
		Metadata:  req.Metadata,
		Timestamp: time.Now(),
	}

	if err := database.DB.Create(log).Error; err != nil {
		zap.L().Error("Failed to create log", zap.Error(err))
		response.InternalError(c, "Failed to create audit log")
		return
	}

	zap.L().Info("Audit log created",
		zap.Uint("id", log.ID),
		zap.String("action", log.Action),
		zap.Stringp("userID", log.UserID),
	)

	response.Created(c, log.ToResponse())
}

func (h *AuditHandler) GetUserLogs(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		response.BadRequest(c, "User ID is required")
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 50

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	var logs []models.Log
	var total int64

	// Get total count
	if err := database.DB.Model(&models.Log{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		zap.L().Error("Failed to count user logs", zap.Error(err))
		response.InternalError(c, "Failed to retrieve logs")
		return
	}

	// Get logs with pagination
	if err := database.DB.Where("user_id = ?", userID).
		Order("timestamp DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error; err != nil {
		zap.L().Error("Failed to get user logs", zap.Error(err))
		response.InternalError(c, "Failed to retrieve logs")
		return
	}

	// Convert to response format
	logResponses := make([]models.LogResponse, len(logs))
	for i, log := range logs {
		logResponses[i] = log.ToResponse()
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	result := gin.H{
		"logs": logResponses,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	response.Success(c, result)
}

func (h *AuditHandler) GetAllLogs(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	limit := 50

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	var logs []models.Log
	var total int64

	query := database.DB.Model(&models.Log{})

	// Filter by action if provided
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		zap.L().Error("Failed to count logs", zap.Error(err))
		response.InternalError(c, "Failed to retrieve logs")
		return
	}

	// Get logs with pagination
	if err := query.Order("timestamp DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error; err != nil {
		zap.L().Error("Failed to get logs", zap.Error(err))
		response.InternalError(c, "Failed to retrieve logs")
		return
	}

	// Convert to response format
	logResponses := make([]models.LogResponse, len(logs))
	for i, log := range logs {
		logResponses[i] = log.ToResponse()
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	result := gin.H{
		"logs": logResponses,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	response.Success(c, result)
}

// Health check endpoint
func (h *AuditHandler) HealthCheck(c *gin.Context) {
	uptime := time.Since(startTime)

	response.Health(c, "audit-service", "1.0.0", int64(uptime.Seconds()))
}

var startTime = time.Now()


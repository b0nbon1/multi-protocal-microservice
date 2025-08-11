package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"notification-service/internal/models"
	"notification-service/internal/websocket"
	"notification-service/shared/response"
)

type NotificationHandler struct {
	hub *websocket.Hub
}

func NewNotificationHandler(hub *websocket.Hub) *NotificationHandler {
	return &NotificationHandler{
		hub: hub,
	}
}

func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var req models.NotifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("Invalid notification request", zap.Error(err))
		response.BadRequest(c, "Invalid request payload")
		return
	}

	notification := &models.Notification{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Type:      req.Type,
		Title:     req.Title,
		Message:   req.Message,
		Data:      req.Data,
		Timestamp: time.Now(),
	}

	if req.UserID != "" {
		h.hub.BroadcastToUser(req.UserID, notification)
	} else {
		h.hub.BroadcastToAll(notification)
	}

	zap.L().Info("Notification sent",
		zap.String("id", notification.ID),
		zap.String("type", notification.Type),
		zap.String("userID", notification.UserID),
	)

	response.Success(c, gin.H{
		"id":        notification.ID,
		"timestamp": notification.Timestamp,
		"status":    "sent",
	})
}

func (h *NotificationHandler) GetActiveConnections(c *gin.Context) {
	userID := c.Query("userId")

	// This is a simplified implementation
	// In a real application, you might want to expose more detailed connection statistics
	stats := gin.H{
		"totalConnections": len(h.hub.GetClients()),
	}

	if userID != "" {
		userConnections := h.hub.GetUserConnectionCount(userID)
		stats["userConnections"] = userConnections
		stats["userId"] = userID
	}

	response.Success(c, stats)
}

// Health check endpoint
func (h *NotificationHandler) HealthCheck(c *gin.Context) {
	uptime := time.Since(startTime)

	response.Health(c, "notification-service", "1.0.0", int64(uptime.Seconds()))
}

var startTime = time.Now()

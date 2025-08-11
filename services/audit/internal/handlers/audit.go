package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		ID:       primitive.NewObjectID(),
		UserID:   req.UserID,
		Action:   req.Action,
		Metadata: req.Metadata,
	}
	log.SetDefaults()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := database.LogsCollection.InsertOne(ctx, log)
	if err != nil {
		zap.L().Error("Failed to create log", zap.Error(err))
		response.InternalError(c, "Failed to create audit log")
		return
	}

	// Update the ID from the insert result
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		log.ID = oid
	}

	zap.L().Info("Audit log created",
		zap.String("id", log.ID.Hex()),
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

	skip := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build filter
	filter := bson.M{"userId": userID}

	// Get total count
	total, err := database.LogsCollection.CountDocuments(ctx, filter)
	if err != nil {
		zap.L().Error("Failed to count user logs", zap.Error(err))
		response.InternalError(c, "Failed to retrieve logs")
		return
	}

	// Find options with pagination and sorting
	findOptions := options.Find().
		SetSort(bson.D{{"timestamp", -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := database.LogsCollection.Find(ctx, filter, findOptions)
	if err != nil {
		zap.L().Error("Failed to get user logs", zap.Error(err))
		response.InternalError(c, "Failed to retrieve logs")
		return
	}
	defer cursor.Close(ctx)

	var logs []models.Log
	if err := cursor.All(ctx, &logs); err != nil {
		zap.L().Error("Failed to decode logs", zap.Error(err))
		response.InternalError(c, "Failed to retrieve logs")
		return
	}

	// Convert to response format
	logResponses := make([]models.LogResponse, len(logs))
	for i, log := range logs {
		logResponses[i] = log.ToResponse()
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	result := models.LogsWithPagination{
		Logs: logResponses,
		Pagination: models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
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

	skip := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build filter
	filter := bson.M{}

	// Filter by action if provided
	if action := c.Query("action"); action != "" {
		filter["action"] = action
	}

	// Filter by date range if provided
	if startDate := c.Query("startDate"); startDate != "" {
		if start, err := time.Parse(time.RFC3339, startDate); err == nil {
			if filter["timestamp"] == nil {
				filter["timestamp"] = bson.M{}
			}
			filter["timestamp"].(bson.M)["$gte"] = start
		}
	}

	if endDate := c.Query("endDate"); endDate != "" {
		if end, err := time.Parse(time.RFC3339, endDate); err == nil {
			if filter["timestamp"] == nil {
				filter["timestamp"] = bson.M{}
			}
			filter["timestamp"].(bson.M)["$lte"] = end
		}
	}

	// Get total count
	total, err := database.LogsCollection.CountDocuments(ctx, filter)
	if err != nil {
		zap.L().Error("Failed to count logs", zap.Error(err))
		response.InternalError(c, "Failed to retrieve logs")
		return
	}

	// Find options with pagination and sorting
	findOptions := options.Find().
		SetSort(bson.D{{"timestamp", -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := database.LogsCollection.Find(ctx, filter, findOptions)
	if err != nil {
		zap.L().Error("Failed to get logs", zap.Error(err))
		response.InternalError(c, "Failed to retrieve logs")
		return
	}
	defer cursor.Close(ctx)

	var logs []models.Log
	if err := cursor.All(ctx, &logs); err != nil {
		zap.L().Error("Failed to decode logs", zap.Error(err))
		response.InternalError(c, "Failed to retrieve logs")
		return
	}

	// Convert to response format
	logResponses := make([]models.LogResponse, len(logs))
	for i, log := range logs {
		logResponses[i] = log.ToResponse()
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	result := models.LogsWithPagination{
		Logs: logResponses,
		Pagination: models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	response.Success(c, result)
}

func (h *AuditHandler) GetLogAnalytics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Aggregation pipeline for action analytics
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$action",
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
		{
			"$limit": 10,
		},
	}

	cursor, err := database.LogsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		zap.L().Error("Failed to get analytics", zap.Error(err))
		response.InternalError(c, "Failed to retrieve analytics")
		return
	}
	defer cursor.Close(ctx)

	var analytics []models.LogAggregationResult
	if err := cursor.All(ctx, &analytics); err != nil {
		zap.L().Error("Failed to decode analytics", zap.Error(err))
		response.InternalError(c, "Failed to retrieve analytics")
		return
	}

	// Get total logs count
	total, err := database.LogsCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		zap.L().Error("Failed to count total logs", zap.Error(err))
		response.InternalError(c, "Failed to retrieve analytics")
		return
	}

	result := gin.H{
		"totalLogs":   total,
		"topActions":  analytics,
		"generatedAt": time.Now(),
	}

	response.Success(c, result)
}

func (h *AuditHandler) SearchLogs(c *gin.Context) {
	searchTerm := c.Query("q")
	if searchTerm == "" {
		response.BadRequest(c, "Search term is required")
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

	skip := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build text search filter
	filter := bson.M{
		"$or": []bson.M{
			{"action": bson.M{"$regex": searchTerm, "$options": "i"}},
			{"metadata": bson.M{"$regex": searchTerm, "$options": "i"}},
		},
	}

	// Get total count
	total, err := database.LogsCollection.CountDocuments(ctx, filter)
	if err != nil {
		zap.L().Error("Failed to count search results", zap.Error(err))
		response.InternalError(c, "Failed to search logs")
		return
	}

	// Find options with pagination and sorting
	findOptions := options.Find().
		SetSort(bson.D{{"timestamp", -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := database.LogsCollection.Find(ctx, filter, findOptions)
	if err != nil {
		zap.L().Error("Failed to search logs", zap.Error(err))
		response.InternalError(c, "Failed to search logs")
		return
	}
	defer cursor.Close(ctx)

	var logs []models.Log
	if err := cursor.All(ctx, &logs); err != nil {
		zap.L().Error("Failed to decode search results", zap.Error(err))
		response.InternalError(c, "Failed to search logs")
		return
	}

	// Convert to response format
	logResponses := make([]models.LogResponse, len(logs))
	for i, log := range logs {
		logResponses[i] = log.ToResponse()
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	result := models.LogsWithPagination{
		Logs: logResponses,
		Pagination: models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	response.Success(c, result)
}

// Health check endpoint
func (h *AuditHandler) HealthCheck(c *gin.Context) {
	uptime := time.Since(startTime)

	// Check MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := database.Client.Ping(ctx, nil); err != nil {
		zap.L().Error("MongoDB health check failed", zap.Error(err))
		response.InternalError(c, "Database connection failed")
		return
	}

	response.Health(c, "audit-service", "1.0.0", int64(uptime.Seconds()))
}

var startTime = time.Now()

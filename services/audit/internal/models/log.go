package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Log struct {
	ID        primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	UserID    *string                `json:"userId,omitempty" bson:"userId,omitempty"`
	Action    string                 `json:"action" bson:"action"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp" bson:"timestamp"`
	CreatedAt time.Time              `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt" bson:"updatedAt"`
}

type CreateLogRequest struct {
	UserID   *string                `json:"userId,omitempty" binding:"omitempty"`
	Action   string                 `json:"action" binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type LogResponse struct {
	ID        string                 `json:"id"`
	UserID    *string                `json:"userId,omitempty"`
	Action    string                 `json:"action"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	CreatedAt time.Time              `json:"createdAt"`
}

type LogsWithPagination struct {
	Logs       []LogResponse `json:"logs"`
	Pagination Pagination    `json:"pagination"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type LogAggregationResult struct {
	Action string `json:"action" bson:"_id"`
	Count  int    `json:"count" bson:"count"`
}

func (l *Log) ToResponse() LogResponse {
	return LogResponse{
		ID:        l.ID.Hex(),
		UserID:    l.UserID,
		Action:    l.Action,
		Metadata:  l.Metadata,
		Timestamp: l.Timestamp,
		CreatedAt: l.CreatedAt,
	}
}

// SetDefaults sets default values for the log entry
func (l *Log) SetDefaults() {
	now := time.Now()
	l.Timestamp = now
	l.CreatedAt = now
	l.UpdatedAt = now
}

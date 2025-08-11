package models

import (
	"time"

	"gorm.io/gorm"
)

type Log struct {
	ID        uint                   `json:"id" gorm:"primaryKey"`
	UserID    *string                `json:"userId,omitempty" gorm:"type:uuid"`
	Action    string                 `json:"action" gorm:"not null"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	Timestamp time.Time              `json:"timestamp" gorm:"autoCreateTime"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
	DeletedAt gorm.DeletedAt         `json:"-" gorm:"index"`
}

type CreateLogRequest struct {
	UserID   *string                `json:"userId,omitempty" binding:"omitempty,uuid"`
	Action   string                 `json:"action" binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type LogResponse struct {
	ID        uint                   `json:"id"`
	UserID    *string                `json:"userId,omitempty"`
	Action    string                 `json:"action"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	CreatedAt time.Time              `json:"createdAt"`
}

func (l *Log) ToResponse() LogResponse {
	return LogResponse{
		ID:        l.ID,
		UserID:    l.UserID,
		Action:    l.Action,
		Metadata:  l.Metadata,
		Timestamp: l.Timestamp,
		CreatedAt: l.CreatedAt,
	}
}


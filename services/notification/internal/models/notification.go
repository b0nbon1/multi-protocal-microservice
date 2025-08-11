package models

import (
	"time"
)

type Notification struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"userId,omitempty"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type NotifyRequest struct {
	UserID  string                 `json:"userId,omitempty" binding:"omitempty,uuid"`
	Type    string                 `json:"type" binding:"required"`
	Title   string                 `json:"title" binding:"required"`
	Message string                 `json:"message" binding:"required"`
	Data    map[string]interface{} `json:"data,omitempty"`
}


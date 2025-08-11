package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ServiceConfig struct {
	AuthServiceURL         string
	ContractServiceURL     string
	PaymentServiceURL      string
	DisputeServiceURL      string
	NotificationServiceURL string
	AuditServiceURL        string
}

type ProxyHandler struct {
	config     ServiceConfig
	httpClient *http.Client
}

func NewProxyHandler(config ServiceConfig) *ProxyHandler {
	return &ProxyHandler{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *ProxyHandler) ProxyRequest(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Build target URL
		fullURL := targetURL + c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			fullURL += "?" + c.Request.URL.RawQuery
		}

		zap.L().Info("Proxying request",
			zap.String("method", c.Request.Method),
			zap.String("originalPath", c.Request.URL.Path),
			zap.String("targetURL", fullURL),
		)

		// Create new request
		req, err := http.NewRequest(c.Request.Method, fullURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			zap.L().Error("Failed to create proxy request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "INTERNAL_SERVER_ERROR",
				"message": "Failed to create proxy request",
			})
			return
		}

		// Copy headers
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Make the request
		resp, err := p.httpClient.Do(req)
		if err != nil {
			zap.L().Error("Proxy request failed", zap.Error(err))
			c.JSON(http.StatusBadGateway, gin.H{
				"success": false,
				"error":   "BAD_GATEWAY",
				"message": "Service unavailable",
			})
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			zap.L().Error("Failed to read proxy response", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "INTERNAL_SERVER_ERROR",
				"message": "Failed to read service response",
			})
			return
		}

		// Send response
		c.Status(resp.StatusCode)
		c.Writer.Write(respBody)

		// Log audit event for non-auth requests
		if targetURL != p.config.AuthServiceURL {
			go p.logAuditEvent(c, targetURL, resp.StatusCode)
		}
	}
}

func (p *ProxyHandler) logAuditEvent(c *gin.Context, targetURL string, statusCode int) {
	userID, exists := c.Get("userID")
	if !exists {
		return // Skip audit for unauthenticated requests
	}

	auditData := map[string]interface{}{
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"targetURL":  targetURL,
		"statusCode": statusCode,
		"userAgent":  c.Request.UserAgent(),
		"ip":         c.ClientIP(),
	}

	logPayload := map[string]interface{}{
		"userId":   userID,
		"action":   fmt.Sprintf("API_REQUEST_%s", c.Request.Method),
		"metadata": auditData,
	}

	payloadBytes, err := json.Marshal(logPayload)
	if err != nil {
		zap.L().Error("Failed to marshal audit log", zap.Error(err))
		return
	}

	req, err := http.NewRequest("POST", p.config.AuditServiceURL+"/api/v1/logs", bytes.NewBuffer(payloadBytes))
	if err != nil {
		zap.L().Error("Failed to create audit request", zap.Error(err))
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		zap.L().Error("Failed to send audit log", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		zap.L().Warn("Audit service returned error", zap.Int("statusCode", resp.StatusCode))
	}
}


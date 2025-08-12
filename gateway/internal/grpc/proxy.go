package grpc

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	auditpb "api-gateway/internal/grpc/audit/proto"
	authpb "api-gateway/internal/grpc/auth/proto"
	contractpb "api-gateway/internal/grpc/contract/proto"
	disputepb "api-gateway/internal/grpc/dispute/proto"
	notificationpb "api-gateway/internal/grpc/notification/proto"
	paymentpb "api-gateway/internal/grpc/payment/proto"
)

type GRPCProxyHandler struct {
	clients *GRPCClients
}

func NewGRPCProxyHandler(clients *GRPCClients) *GRPCProxyHandler {
	return &GRPCProxyHandler{
		clients: clients,
	}
}

// Auth handlers
func (h *GRPCProxyHandler) Register(c *gin.Context) {
	var reqData struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &authpb.RegisterRequest{
		Email:     reqData.Email,
		Password:  reqData.Password,
		FirstName: reqData.FirstName,
		LastName:  reqData.LastName,
	}

	resp, err := h.clients.AuthClient.Register(ctx, req)
	if err != nil {
		zap.L().Error("gRPC Register failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Service unavailable",
		})
		return
	}

	statusCode := http.StatusOK
	if !resp.Success {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, resp)
}

func (h *GRPCProxyHandler) Login(c *gin.Context) {
	var reqData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &authpb.LoginRequest{
		Email:    reqData.Email,
		Password: reqData.Password,
	}

	resp, err := h.clients.AuthClient.Login(ctx, req)
	if err != nil {
		zap.L().Error("gRPC Login failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Service unavailable",
		})
		return
	}

	statusCode := http.StatusOK
	if !resp.Success {
		statusCode = http.StatusUnauthorized
	}

	c.JSON(statusCode, resp)
}

func (h *GRPCProxyHandler) ValidateToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "UNAUTHORIZED",
			"message": "No token provided",
		})
		return
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &authpb.ValidateTokenRequest{Token: token}
	resp, err := h.clients.AuthClient.ValidateToken(ctx, req)
	if err != nil {
		zap.L().Error("gRPC ValidateToken failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Service unavailable",
		})
		return
	}

	if !resp.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":  true,
		"userId": resp.UserId,
		"email":  resp.Email,
	})
}

func (h *GRPCProxyHandler) GetUser(c *gin.Context) {
	userId := c.Param("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "INVALID_REQUEST",
			"message": "User ID is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &authpb.GetUserRequest{UserId: userId}
	resp, err := h.clients.AuthClient.GetUser(ctx, req)
	if err != nil {
		zap.L().Error("gRPC GetUser failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Service unavailable",
		})
		return
	}

	statusCode := http.StatusOK
	if !resp.Success {
		statusCode = http.StatusNotFound
	}

	c.JSON(statusCode, resp)
}

// Contract handlers
func (h *GRPCProxyHandler) CreateContract(c *gin.Context) {
	var reqData struct {
		Title        string  `json:"title"`
		Description  string  `json:"description"`
		Amount       float64 `json:"amount"`
		Currency     string  `json:"currency"`
		ClientId     string  `json:"clientId"`
		FreelancerId string  `json:"freelancerId"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &contractpb.CreateContractRequest{
		UserId:       userID.(string),
		Title:        reqData.Title,
		Description:  reqData.Description,
		Amount:       reqData.Amount,
		Currency:     reqData.Currency,
		ClientId:     reqData.ClientId,
		FreelancerId: reqData.FreelancerId,
	}

	resp, err := h.clients.ContractClient.CreateContract(ctx, req)
	if err != nil {
		zap.L().Error("gRPC CreateContract failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Service unavailable",
		})
		return
	}

	statusCode := http.StatusCreated
	if !resp.Success {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, resp)
}

func (h *GRPCProxyHandler) GetContracts(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &contractpb.GetContractsRequest{
		UserId: userID.(string),
		Page:   int32(page),
		Limit:  int32(limit),
	}

	resp, err := h.clients.ContractClient.GetContracts(ctx, req)
	if err != nil {
		zap.L().Error("gRPC GetContracts failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Service unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *GRPCProxyHandler) GetContract(c *gin.Context) {
	contractID := c.Param("contractId")
	if contractID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "INVALID_REQUEST",
			"message": "Contract ID is required",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &contractpb.GetContractRequest{
		ContractId: contractID,
		UserId:     userID.(string),
	}

	resp, err := h.clients.ContractClient.GetContract(ctx, req)
	if err != nil {
		zap.L().Error("gRPC GetContract failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Service unavailable",
		})
		return
	}

	statusCode := http.StatusOK
	if !resp.Success {
		statusCode = http.StatusNotFound
	}

	c.JSON(statusCode, resp)
}

// Payment handlers
func (h *GRPCProxyHandler) CreateWallet(c *gin.Context) {
	var reqData struct {
		Currency string `json:"currency"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &paymentpb.CreateWalletRequest{
		UserId:   userID.(string),
		Currency: reqData.Currency,
	}

	resp, err := h.clients.PaymentClient.CreateWallet(ctx, req)
	if err != nil {
		zap.L().Error("gRPC CreateWallet failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Service unavailable",
		})
		return
	}

	statusCode := http.StatusCreated
	if !resp.Success {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, resp)
}

func (h *GRPCProxyHandler) CreateTransfer(c *gin.Context) {
	var reqData struct {
		ToUserId    string  `json:"toUserId"`
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`
		Description string  `json:"description"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &paymentpb.CreateTransferRequest{
		FromUserId:  userID.(string),
		ToUserId:    reqData.ToUserId,
		Amount:      reqData.Amount,
		Currency:    reqData.Currency,
		Description: reqData.Description,
	}

	resp, err := h.clients.PaymentClient.CreateTransfer(ctx, req)
	if err != nil {
		zap.L().Error("gRPC CreateTransfer failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Service unavailable",
		})
		return
	}

	statusCode := http.StatusCreated
	if !resp.Success {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, resp)

	// Log audit event
	go h.logAuditEvent(userID.(string), "CREATE_TRANSFER", "payment", map[string]string{
		"amount":   strconv.FormatFloat(reqData.Amount, 'f', 2, 64),
		"currency": reqData.Currency,
		"toUserId": reqData.ToUserId,
	}, c.ClientIP(), c.Request.UserAgent())
}

// Dispute handlers
func (h *GRPCProxyHandler) CreateDispute(c *gin.Context) {
	var reqData struct {
		ContractId  string `json:"contractId"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Category    string `json:"category"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &disputepb.CreateDisputeRequest{
		UserId:      userID.(string),
		ContractId:  reqData.ContractId,
		Title:       reqData.Title,
		Description: reqData.Description,
		Category:    reqData.Category,
	}

	resp, err := h.clients.DisputeClient.CreateDispute(ctx, req)
	if err != nil {
		zap.L().Error("gRPC CreateDispute failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Service unavailable",
		})
		return
	}

	statusCode := http.StatusCreated
	if !resp.Success {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, resp)

	// Log audit event
	go h.logAuditEvent(userID.(string), "CREATE_DISPUTE", "dispute", map[string]string{
		"contractId": reqData.ContractId,
		"category":   reqData.Category,
	}, c.ClientIP(), c.Request.UserAgent())
}

// Notification handlers
func (h *GRPCProxyHandler) GetNotifications(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	unreadOnly, _ := strconv.ParseBool(c.DefaultQuery("unreadOnly", "false"))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &notificationpb.GetNotificationsRequest{
		UserId:     userID.(string),
		Page:       int32(page),
		Limit:      int32(limit),
		UnreadOnly: unreadOnly,
	}

	resp, err := h.clients.NotificationClient.GetNotifications(ctx, req)
	if err != nil {
		zap.L().Error("gRPC GetNotifications failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Service unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *GRPCProxyHandler) MarkNotificationAsRead(c *gin.Context) {
	notificationID := c.Param("notificationId")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "INVALID_REQUEST",
			"message": "Notification ID is required",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &notificationpb.MarkAsReadRequest{
		NotificationId: notificationID,
		UserId:         userID.(string),
	}

	resp, err := h.clients.NotificationClient.MarkAsRead(ctx, req)
	if err != nil {
		zap.L().Error("gRPC MarkAsRead failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Service unavailable",
		})
		return
	}

	statusCode := http.StatusOK
	if !resp.Success {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, resp)
}

// Audit handlers
func (h *GRPCProxyHandler) GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	action := c.Query("action")
	resource := c.Query("resource")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := &auditpb.GetLogsRequest{
		Page:      int32(page),
		Limit:     int32(limit),
		Action:    action,
		Resource:  resource,
		StartDate: startDate,
		EndDate:   endDate,
	}

	resp, err := h.clients.AuditClient.GetLogs(ctx, req)
	if err != nil {
		zap.L().Error("gRPC GetLogs failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Service unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *GRPCProxyHandler) logAuditEvent(userID, action, resource string, metadata map[string]string, ipAddress, userAgent string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &auditpb.CreateLogRequest{
		UserId:    userID,
		Action:    action,
		Resource:  resource,
		Metadata:  metadata,
		IpAddress: ipAddress,
		UserAgent: userAgent,
	}

	_, err := h.clients.AuditClient.CreateLog(ctx, req)
	if err != nil {
		zap.L().Error("Failed to create audit log", zap.Error(err))
	}
}

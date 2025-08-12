package grpc

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	auditpb "api-gateway/internal/grpc/audit/proto"
	authpb "api-gateway/internal/grpc/auth/proto"
	contractpb "api-gateway/internal/grpc/contract/proto"
	disputepb "api-gateway/internal/grpc/dispute/proto"
	notificationpb "api-gateway/internal/grpc/notification/proto"
	paymentpb "api-gateway/internal/grpc/payment/proto"
)

type GRPCClients struct {
	AuthClient         authpb.AuthServiceClient
	ContractClient     contractpb.ContractServiceClient
	PaymentClient      paymentpb.PaymentServiceClient
	DisputeClient      disputepb.DisputeServiceClient
	NotificationClient notificationpb.NotificationServiceClient
	AuditClient        auditpb.AuditServiceClient
	connections        []*grpc.ClientConn
}

type GRPCConfig struct {
	AuthServiceAddr         string
	ContractServiceAddr     string
	PaymentServiceAddr      string
	DisputeServiceAddr      string
	NotificationServiceAddr string
	AuditServiceAddr        string
}

func NewGRPCClients(config GRPCConfig) (*GRPCClients, error) {
	clients := &GRPCClients{
		connections: make([]*grpc.ClientConn, 0),
	}

	// Create auth service client
	authConn, err := createConnection(config.AuthServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}
	clients.AuthClient = authpb.NewAuthServiceClient(authConn)
	clients.connections = append(clients.connections, authConn)

	// Create contract service client
	contractConn, err := createConnection(config.ContractServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to contract service: %w", err)
	}
	clients.ContractClient = contractpb.NewContractServiceClient(contractConn)
	clients.connections = append(clients.connections, contractConn)

	// Create payment service client
	paymentConn, err := createConnection(config.PaymentServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}
	clients.PaymentClient = paymentpb.NewPaymentServiceClient(paymentConn)
	clients.connections = append(clients.connections, paymentConn)

	// Create dispute service client
	disputeConn, err := createConnection(config.DisputeServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to dispute service: %w", err)
	}
	clients.DisputeClient = disputepb.NewDisputeServiceClient(disputeConn)
	clients.connections = append(clients.connections, disputeConn)

	// Create notification service client
	notificationConn, err := createConnection(config.NotificationServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to notification service: %w", err)
	}
	clients.NotificationClient = notificationpb.NewNotificationServiceClient(notificationConn)
	clients.connections = append(clients.connections, notificationConn)

	// Create audit service client
	auditConn, err := createConnection(config.AuditServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to audit service: %w", err)
	}
	clients.AuditClient = auditpb.NewAuditServiceClient(auditConn)
	clients.connections = append(clients.connections, auditConn)

	zap.L().Info("All gRPC clients initialized successfully")
	return clients, nil
}

func createConnection(address string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	return conn, nil
}

func (c *GRPCClients) Close() error {
	for _, conn := range c.connections {
		if err := conn.Close(); err != nil {
			zap.L().Error("Failed to close gRPC connection", zap.Error(err))
			return err
		}
	}
	zap.L().Info("All gRPC connections closed")
	return nil
}

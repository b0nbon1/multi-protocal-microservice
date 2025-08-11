package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	Client         *mongo.Client
	Database       *mongo.Database
	LogsCollection *mongo.Collection
)

const (
	DatabaseName       = "audit_db"
	LogsCollectionName = "logs"
)

func Connect() error {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		return fmt.Errorf("MONGODB_URI environment variable is required")
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Set timeouts
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Set global variables
	Client = client
	Database = client.Database(DatabaseName)
	LogsCollection = Database.Collection(LogsCollectionName)

	// Create indexes for better performance
	if err := createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	zap.L().Info("Connected to MongoDB successfully",
		zap.String("database", DatabaseName),
		zap.String("uri", mongoURI),
	)

	return nil
}

func createIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"userId": 1,
			},
		},
		{
			Keys: map[string]interface{}{
				"action": 1,
			},
		},
		{
			Keys: map[string]interface{}{
				"timestamp": -1,
			},
		},
		{
			Keys: map[string]interface{}{
				"userId":    1,
				"timestamp": -1,
			},
		},
		{
			Keys: map[string]interface{}{
				"action":    1,
				"timestamp": -1,
			},
		},
		{
			Keys: map[string]interface{}{
				"userId":    1,
				"action":    1,
				"timestamp": -1,
			},
		},
	}

	// Create indexes
	_, err := LogsCollection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}

	zap.L().Info("MongoDB indexes created successfully")
	return nil
}

func Close() error {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := Client.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
		}

		zap.L().Info("Disconnected from MongoDB")
	}
	return nil
}

// GetCollection returns a specific collection
func GetCollection(name string) *mongo.Collection {
	return Database.Collection(name)
}

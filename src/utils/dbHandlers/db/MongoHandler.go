package db

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoHandler struct {
	Client   *mongo.Client
	Database *mongo.Database
}

type MongoConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
}

func NewMongoHandler(connString, dbName string) (*MongoHandler, error) {
	clientOptions := options.Client().ApplyURI(connString)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("could not connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(dbName)
	return &MongoHandler{
		Client:   client,
		Database: database,
	}, nil
}

func (h *MongoHandler) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return h.Client.Disconnect(ctx)
}

// BuildMongoConnectionString creates a MongoDB connection string from config
// Format: mongodb://[username:password@]host[:port][/database]
func BuildMongoConnectionString(config MongoConfig) string {
	// URL encode username and password to handle special characters
	user := url.QueryEscape(config.User)
	password := url.QueryEscape(config.Password)

	var connStr string

	// Build connection string based on available credentials
	if user != "" && password != "" {
		connStr = fmt.Sprintf("mongodb://%s:%s@%s", user, password, config.Host)
	} else {
		connStr = fmt.Sprintf("mongodb://%s", config.Host)
	}

	// Add port if specified
	if config.Port != "" {
		connStr = fmt.Sprintf("%s:%s", connStr, config.Port)
	}

	// Add database if specified
	if config.Database != "" {
		connStr = fmt.Sprintf("%s/%s", connStr, config.Database)
	}

	return connStr
}

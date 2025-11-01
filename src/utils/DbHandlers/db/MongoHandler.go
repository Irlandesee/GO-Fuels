package db

import (
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

func NewMongoHandler(config MongoConfig) (*MongoHandler, error) {
	clientOpts := options.Client().

}

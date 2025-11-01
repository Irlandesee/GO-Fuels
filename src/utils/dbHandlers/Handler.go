package dbHandlers

import (
	"Irlandesee/GO-Fuels/src/utils/dbHandlers/db"
	"Irlandesee/GO-Fuels/src/utils/dbHandlers/rabbit"
)

type Handler struct {
	Postgres *db.PostgresHandler
	Mongo    *db.MongoHandler
	Rabbit   *rabbit.RabbitHandler
}

func NewDbHandler() *Handler {
	return &Handler{}
}

func (h *Handler) InitPostgres(host, port, dbName, user, password string) error {
	pgHandler, err := db.NewPostgresHandler(host, port, dbName, user, password)
	if err != nil {
		return err
	}
	h.Postgres = pgHandler
	return nil
}

func (h *Handler) InitMongo(connectionString, dbName string) error {
	mongoHandler, err := db.NewMongoHandler(connectionString, dbName)
	if err != nil {
		return err
	}
	h.Mongo = mongoHandler
	return nil
}

func (h *Handler) InitRabbit(uri string) error {
	rabbitHandler, err := rabbit.NewRabbitHandler(uri)
	if err != nil {
		return err
	}
	h.Rabbit = rabbitHandler
	return nil
}

func (h *Handler) Close() error {
	if h.Mongo != nil {
		if err := h.Mongo.Close(); err != nil {
			return err
		}
	}
	if h.Rabbit != nil {
		if err := h.Rabbit.Close(); err != nil {
			return err
		}
	}
	// Postgres connection is managed by GORM, close if needed
	if h.Postgres != nil {
		if err := h.Postgres.Close(); err != nil {
			return err
		}
	}
	return nil
}

package main

import (
	"log"
	"os"

	"Irlandesee/GO-Fuels/src/routers"
	"Irlandesee/GO-Fuels/src/utils/dbHandlers"

	"github.com/labstack/echo/v4"
)

func main() {
	handler := dbHandlers.NewDbHandler()

	// Init Postgres
	if err := handler.InitPostgres(
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "6543"),
		getEnv("DB_NAME", "postgres"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASS", "mysecretpassword"),
	); err != nil {
		log.Fatalf("Failed to init Postgres: %v", err)
	}

	// Init MongoDB
	if err := handler.InitMongo(
		getEnv("MONGO_URI", "mongodb://admin:password@localhost:27017"),
		getEnv("MONGO_DB", "go_fuels"),
	); err != nil {
		log.Fatalf("Failed to init MongoDB: %v", err)
	}

	// Init RabbitMQ
	if err := handler.InitRabbit(
		getEnv("RABBIT_URI", "amqp://guest:guest@localhost:5672/"),
	); err != nil {
		log.Fatalf("Failed to init RabbitMQ: %v", err)
	}

	defer func(handler *dbHandlers.Handler) {
		err := handler.Close()
		if err != nil {
			log.Fatalf("Failed to close DB handler: %v", err)
		}
	}(handler)

	e := echo.New()
	routers.RegisterRoutes(e, handler)

	e.Logger.Fatal(e.Start(":8080"))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

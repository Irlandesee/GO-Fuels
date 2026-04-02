package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"Irlandesee/GO-Fuels/src/ingester"
	"Irlandesee/GO-Fuels/src/models"
	"Irlandesee/GO-Fuels/src/utils/dbHandlers/db"
	"Irlandesee/GO-Fuels/src/utils/dbHandlers/rabbit"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "6543")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASS", "mysecretpassword")
	dbName := getEnv("DB_NAME", "postgres")
	rabbitURI := getEnv("RABBIT_URI", "amqp://guest:guest@localhost:5672/")

	handler, err := db.NewPostgresHandler(dbHost, dbPort, dbName, dbUser, dbPass)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer handler.Close()

	rabbitHandler, err := rabbit.NewRabbitHandler(rabbitURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer func(rabbitHandler *rabbit.RabbitHandler) {
		err := rabbitHandler.Close()
		if err != nil {

		}
	}(rabbitHandler)

	fuelIngester := ingester.NewIngester(handler)

	msgs, err := rabbitHandler.Consume(models.IngestionQueueName)
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Println("Ingester started, waiting for a single job from RabbitMQ...")

	select {
	case msg, ok := <-msgs:
		if !ok {
			log.Fatal("RabbitMQ channel closed before receiving a message")
		}

		var jobMsg models.IngestionJobMessage
		if err := json.Unmarshal(msg.Body, &jobMsg); err != nil {
			log.Fatalf("Failed to unmarshal message: %v", err)
		}

		log.Printf("Received job %d", jobMsg.JobID)

		job, err := handler.GetIngestionJob(ctx, jobMsg.JobID)
		if err != nil {
			log.Fatalf("Failed to load job %d: %v", jobMsg.JobID, err)
		}

		log.Printf("Starting ingestion for lat: %f, lng: %f, radius: %d", job.Lat, job.Lng, job.Radius)
		fuelIngester.ProcessJob(ctx, job)

		if err := msg.Ack(false); err != nil {
			log.Printf("Failed to ack message: %v", err)
		}

		log.Printf("Job %d completed with status %s", job.ID, job.Status)

	case <-ctx.Done():
		log.Println("Shutting down ingester...")
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

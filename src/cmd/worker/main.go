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
	// Database config
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "6543")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASS", "mysecretpassword")
	dbName := getEnv("DB_NAME", "postgres")

	// RabbitMQ config
	rabbitURI := getEnv("RABBIT_URI", "amqp://guest:guest@localhost:5672/")

	// Init Postgres
	pgHandler, err := db.NewPostgresHandler(dbHost, dbPort, dbName, dbUser, dbPass)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pgHandler.Close()

	// Init RabbitMQ
	rabbitHandler, err := rabbit.NewRabbitHandler(rabbitURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitHandler.Close()

	// Init Ingester
	fuelIngester := ingester.NewIngester(pgHandler)

	// Consume from queue
	msgs, err := rabbitHandler.Consume(models.IngestionQueueName)
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Println("Worker started, waiting for ingestion jobs...")

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				log.Println("RabbitMQ channel closed, shutting down")
				return
			}

			var jobMsg models.IngestionJobMessage
			if err := json.Unmarshal(msg.Body, &jobMsg); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				_ = msg.Nack(false, false)
				continue
			}

			log.Printf("Received job %d", jobMsg.JobID)

			job, err := pgHandler.GetIngestionJob(ctx, jobMsg.JobID)
			if err != nil {
				log.Printf("Failed to load job %d: %v", jobMsg.JobID, err)
				_ = msg.Nack(false, false)
				continue
			}

			fuelIngester.ProcessJob(ctx, job)

			if err := msg.Ack(false); err != nil {
				log.Printf("Failed to ack message for job %d: %v", jobMsg.JobID, err)
			}

			log.Printf("Job %d completed with status %s", job.ID, job.Status)

		case <-ctx.Done():
			log.Println("Shutting down worker...")
			return
		}
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}


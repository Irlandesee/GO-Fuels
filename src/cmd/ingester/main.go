package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"Irlandesee/GO-Fuels/src/ingester"
	"Irlandesee/GO-Fuels/src/utils/dbHandlers/db"
)

func main() {
	// Configurazione database da variabili d'ambiente o default (come da docker-compose)
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "6543") // Porta esterna definita nel docker-compose
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASS", "mysecretpassword")
	dbName := getEnv("DB_NAME", "postgres") // Default per timescale image se non specificato altrimenti

	// Inizializzazione PostgresHandler
	handler, err := db.NewPostgresHandler(dbHost, dbPort, dbName, dbUser, dbPass)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer handler.Close()

	// Inizializzazione Ingester
	fuelIngester := ingester.NewIngester(handler)

	// Context con cancellazione su segnali di sistema
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Esempio di coordinate (dal curl fornito)
	lat := 45.6893093
	lng := 8.7359091
	radius := 5

	log.Printf("Starting ingestion for lat: %f, lng: %f, radius: %d", lat, lng, radius)

	// Esecuzione dell'ingestion
	records, err := fuelIngester.FetchAndIngest(ctx, lat, lng, radius)
	if err != nil {
		log.Printf("Ingestion failed: %v", err)
	} else {
		log.Printf("Ingestion completed successfully, %d records processed", records)
	}

	// Se si volesse far girare periodicamente:
	/*
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := fuelIngester.FetchAndIngest(ctx, lat, lng, radius); err != nil {
					log.Printf("Periodic ingestion failed: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	*/
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

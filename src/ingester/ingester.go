package ingester

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"Irlandesee/GO-Fuels/src/models"
	"Irlandesee/GO-Fuels/src/utils/dbHandlers/db"
)

// API models
type SearchZoneRequest struct {
	Points []Point `json:"points"`
	Radius int     `json:"radius"`
}

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type SearchZoneResponse struct {
	Success bool            `json:"success"`
	Center  Point           `json:"center"`
	Results []StationResult `json:"results"`
}

type StationResult struct {
	ID         int           `json:"id"`
	Name       string        `json:"name"`
	Fuels      []FuelPrice   `json:"fuels"`
	Location   Point         `json:"location"`
	InsertDate time.Time     `json:"insertDate"`
	Address    *string       `json:"address"`
	Brand      string        `json:"brand"`
	Distance   string        `json:"distance"`
}

type FuelPrice struct {
	ID       int     `json:"id"`
	Price    float64 `json:"price"`
	Name     string  `json:"name"`
	FuelID   int     `json:"fuelId"`
	IsSelf   bool    `json:"isSelf"`
}

type Ingester struct {
	dbHandler *db.PostgresHandler
}

func NewIngester(dbHandler *db.PostgresHandler) *Ingester {
	return &Ingester{
		dbHandler: dbHandler,
	}
}

func (i *Ingester) FetchAndIngest(ctx context.Context, lat, lng float64, radius int) error {
	url := "https://carburanti.mise.gov.it/ospzApi/search/zone"
	
	reqBody := SearchZoneRequest{
		Points: []Point{{Lat: lat, Lng: lng}},
		Radius: radius,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:148.0) Gecko/20100101 Firefox/148.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://carburanti.mise.gov.it")
	req.Header.Set("Referer", "https://carburanti.mise.gov.it/ospzSearch/zona")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp SearchZoneResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("API response indicates failure")
	}

	for _, station := range apiResp.Results {
		for _, fuel := range station.Fuels {
			fuelData := &models.FuelData{
				LocationKey:  fmt.Sprintf("station_%d", station.ID),
				FuelKey:      fmt.Sprintf("fuel_%d_%t", fuel.FuelID, fuel.IsSelf),
				FuelCategory: fuel.Name, // Using the name as category for now
				Price:        fuel.Price,
				LastUpdate:   station.InsertDate,
				Currency:     "EUR",
				IsActive:     true,
			}

			if err := i.dbHandler.UpsertFuelData(ctx, fuelData); err != nil {
				fmt.Printf("Error upserting fuel data for station %d, fuel %d: %v\n", station.ID, fuel.FuelID, err)
			}
		}
	}

	return nil
}

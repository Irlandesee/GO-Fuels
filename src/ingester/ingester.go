package ingester

import (
	"context"
	"fmt"
	"time"

	"resty.dev/v3"

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
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	Fuels      []FuelPrice `json:"fuels"`
	Location   Point       `json:"location"`
	InsertDate time.Time   `json:"insertDate"`
	Address    *string     `json:"address"`
	Brand      string      `json:"brand"`
	Distance   string      `json:"distance"`
}

type FuelPrice struct {
	ID     int     `json:"id"`
	Price  float64 `json:"price"`
	Name   string  `json:"name"`
	FuelID int     `json:"fuelId"`
	IsSelf bool    `json:"isSelf"`
}

type Ingester struct {
	dbHandler *db.PostgresHandler
	client    *resty.Client
}

func NewIngester(dbHandler *db.PostgresHandler) *Ingester {
	client := resty.New().
		SetBaseURL("https://carburanti.mise.gov.it").
		SetTimeout(30 * time.Second).
		SetHeaders(map[string]string{
			"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:148.0) Gecko/20100101 Firefox/148.0",
			"Accept":     "application/json",
			"Origin":     "https://carburanti.mise.gov.it",
			"Referer":    "https://carburanti.mise.gov.it/ospzSearch/zona",
		})

	return &Ingester{
		dbHandler: dbHandler,
		client:    client,
	}
}

// ProcessJob manages the full lifecycle of an ingestion job: RUNNING -> DONE/FAILED
func (i *Ingester) ProcessJob(ctx context.Context, job *models.IngestionJob) {
	// Mark as RUNNING
	if err := i.dbHandler.UpdateIngestionJobStatus(ctx, job.ID, models.JobStatusRunning, "", 0); err != nil {
		fmt.Printf("Error updating job %d to RUNNING: %v\n", job.ID, err)
		return
	}

	records, err := i.FetchAndIngest(ctx, job.Lat, job.Lng, job.Radius)
	if err != nil {
		_ = i.dbHandler.UpdateIngestionJobStatus(ctx, job.ID, models.JobStatusFailed, err.Error(), 0)
		return
	}

	_ = i.dbHandler.UpdateIngestionJobStatus(ctx, job.ID, models.JobStatusDone, "", records)
}

// FetchAndIngest calls the MISE API and upserts fuel data. Returns the number of records processed.
func (i *Ingester) FetchAndIngest(ctx context.Context, lat, lng float64, radius int) (int, error) {
	reqBody := SearchZoneRequest{
		Points: []Point{{Lat: lat, Lng: lng}},
		Radius: radius,
	}

	var apiResp SearchZoneResponse

	resp, err := i.client.R().
		SetContext(ctx).
		SetBody(reqBody).
		SetResult(&apiResp).
		Post("/ospzApi/search/zone")
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.IsError() {
		return 0, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.String())
	}

	if !apiResp.Success {
		return 0, fmt.Errorf("API response indicates failure")
	}

	records := 0
	for _, station := range apiResp.Results {
		for _, fuel := range station.Fuels {
			fuelData := &models.FuelData{
				LocationKey:  fmt.Sprintf("station_%d", station.ID),
				FuelKey:      fmt.Sprintf("fuel_%d_%t", fuel.FuelID, fuel.IsSelf),
				FuelCategory: fuel.Name,
				Price:        fuel.Price,
				LastUpdate:   station.InsertDate,
				Currency:     "EUR",
				IsActive:     true,
			}

			if err := i.dbHandler.UpsertFuelData(ctx, fuelData); err != nil {
				fmt.Printf("Error upserting fuel data for station %d, fuel %d: %v\n", station.ID, fuel.FuelID, err)
				continue
			}
			records++
		}
	}

	return records, nil
}

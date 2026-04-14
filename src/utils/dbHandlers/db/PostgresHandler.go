package db

import (
	"Irlandesee/GO-Fuels/src/models"
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresHandler struct {
	DB *gorm.DB
}

func (h *PostgresHandler) Close() error {
	sqlDB, err := h.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func NewPostgresHandler(host, port, dbName, user, password string) (*PostgresHandler, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s dbname=%s port=%s password=%s sslmode=disable",
		host, user, dbName, port, password,
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		PrepareStmt:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GORM with postgres driver: %w", err)
	}

	// Auto-migrate models
	if err := gormDB.AutoMigrate(&models.FuelData{}, &models.IngestionJob{}); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	// Create hypertable for TimescaleDB if it doesn't exist
	// We use the 'last_update' column as the time dimension
	// IMPORTANT: Unique constraints on hypertables must include the partitioning column (last_update).
	err = gormDB.Exec("SELECT create_hypertable('fuel_data', 'last_update', if_not_exists => TRUE);").Error
	if err != nil {
		// Log error but don't fail if it's already a hypertable or other minor issue
		// In some versions/setups of TimescaleDB 'if_not_exists' might behave differently
		fmt.Printf("Warning: failed to create hypertable (might already exist): %v\n", err)
	}

	return &PostgresHandler{DB: gormDB}, nil
}

// ─── FuelData ─────────────────────────────────────────────────────────────────

func (h *PostgresHandler) CreateFuelData(ctx context.Context, fd *models.FuelData) error {
	fd.LastUpdate = time.Now()
	return h.DB.WithContext(ctx).Create(fd).Error
}

// UpsertFuelData inserts a new fuel price row or updates an existing one.
// Uses ON CONFLICT on the unique index (location_key, fuel_key, last_update)
// for atomic, idempotent upserts — no precision issues with timestamp matching.
func (h *PostgresHandler) UpsertFuelData(ctx context.Context, fd *models.FuelData) error {
	if fd.LastUpdate.IsZero() {
		fd.LastUpdate = time.Now()
	}

	return h.DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "location_key"},
				{Name: "fuel_key"},
				{Name: "last_update"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"price", "fuel_category", "currency", "is_active", "updated_at",
			}),
		}).
		Create(fd).Error
}

func (h *PostgresHandler) GetFuelDataByLocation(ctx context.Context, locationKey string) ([]models.FuelData, error) {
	var results []models.FuelData
	err := h.DB.WithContext(ctx).
		Where("location_key = ? AND is_active = true", locationKey).
		Find(&results).Error
	return results, err
}

func (h *PostgresHandler) GetFuelDataByLocationAndFuel(ctx context.Context, locationKey, fuelKey string) (*models.FuelData, error) {
	var result models.FuelData
	err := h.DB.WithContext(ctx).
		Where("location_key = ? AND fuel_key = ? AND is_active = true", locationKey, fuelKey).
		First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ─── Price History ────────────────────────────────────────────────────────────

func (h *PostgresHandler) GetPriceHistory(ctx context.Context, locationKey, fuelKey string, since time.Time) ([]models.FuelData, error) {
	var results []models.FuelData
	err := h.DB.WithContext(ctx).
		Where("location_key = ? AND fuel_key = ? AND last_update >= ?", locationKey, fuelKey, since).
		Order("last_update ASC").
		Find(&results).Error
	return results, err
}

func (h *PostgresHandler) GetAveragePrice(ctx context.Context, fuelKey string, since time.Time) (float64, error) {
	var avg float64
	err := h.DB.WithContext(ctx).
		Model(&models.FuelData{}).
		Select("AVG(price)").
		Where("fuel_key = ? AND last_update >= ? AND is_active = true", fuelKey, since).
		Scan(&avg).Error
	return avg, err
}

func (h *PostgresHandler) GetLatestPricesByFuelType(ctx context.Context, fuelKey string, limit int) ([]models.FuelData, error) {
	var results []models.FuelData
	err := h.DB.WithContext(ctx).
		Where("fuel_key = ? AND is_active = true", fuelKey).
		Order("last_update DESC").
		Limit(limit).
		Find(&results).Error
	return results, err
}

// ─── Price Update ─────────────────────────────────────────────────────────────

func (h *PostgresHandler) UpdateFuelPrice(ctx context.Context, locationKey, fuelKey string, price float64) error {
	result := h.DB.WithContext(ctx).
		Model(&models.FuelData{}).
		Where("location_key = ? AND fuel_key = ?", locationKey, fuelKey).
		Updates(map[string]interface{}{
			"price":       price,
			"last_update": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no fuel data found for location_key=%s fuel_key=%s", locationKey, fuelKey)
	}
	return nil
}

// ─── Ingestion Jobs ──────────────────────────────────────────────────────────

func (h *PostgresHandler) CreateIngestionJob(ctx context.Context, job *models.IngestionJob) error {
	job.Status = models.JobStatusStart
	return h.DB.WithContext(ctx).Create(job).Error
}

func (h *PostgresHandler) GetIngestionJob(ctx context.Context, jobID uint) (*models.IngestionJob, error) {
	var job models.IngestionJob
	err := h.DB.WithContext(ctx).First(&job, jobID).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (h *PostgresHandler) UpdateIngestionJobStatus(ctx context.Context, jobID uint, status models.JobStatus, errMsg string, records int) error {
	updates := map[string]interface{}{
		"status": status,
	}

	now := time.Now()
	switch status {
	case models.JobStatusRunning:
		updates["started_at"] = now
	case models.JobStatusDone, models.JobStatusFailed:
		updates["ended_at"] = now
		updates["records"] = records
		if errMsg != "" {
			updates["error"] = errMsg
		}
	}

	return h.DB.WithContext(ctx).Model(&models.IngestionJob{}).Where("id = ?", jobID).Updates(updates).Error
}

func (h *PostgresHandler) ListIngestionJobs(ctx context.Context, limit, offset int) ([]models.IngestionJob, error) {
	var jobs []models.IngestionJob
	err := h.DB.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&jobs).Error
	return jobs, err
}

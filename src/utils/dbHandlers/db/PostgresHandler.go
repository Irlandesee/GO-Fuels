package db

import (
	"Irlandesee/GO-Fuels/src/models"
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	// Auto-migrate the FuelData model
	if err := gormDB.AutoMigrate(&models.FuelData{}); err != nil {
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

// UpsertFuelData updates the price if (location_key, fuel_key) already exists,
// otherwise inserts a new row.
func (h *PostgresHandler) UpsertFuelData(ctx context.Context, fd *models.FuelData) error {
	fd.LastUpdate = time.Now()

	return h.DB.WithContext(ctx).
		Where(models.FuelData{LocationKey: fd.LocationKey, FuelKey: fd.FuelKey}).
		Assign(models.FuelData{
			Price:        fd.Price,
			FuelCategory: fd.FuelCategory,
			Currency:     fd.Currency,
			IsActive:     fd.IsActive,
			LastUpdate:   fd.LastUpdate,
		}).
		FirstOrCreate(fd).Error
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

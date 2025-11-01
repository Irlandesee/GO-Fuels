package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gorm"
)

// UserPreferences represents user settings and preferences
type UserPreferences struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID            string             `bson:"user_id" json:"user_id"`
	PreferredBrands   []string           `bson:"preferred_brands" json:"preferred_brands"`
	PreferredFuelType string             `bson:"preferred_fuel_type" json:"preferred_fuel_type"`
	NotifyOnPrice     bool               `bson:"notify_on_price" json:"notify_on_price"`
	PriceThreshold    float64            `bson:"price_threshold" json:"price_threshold"`
	DefaultLocation   string             `bson:"default_location" json:"default_location"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
}

// Location represents a fuel station location
type Location struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	LocationKey string             `bson:"location_key" json:"location_key"` // Unique identifier (indexed)
	Name        string             `bson:"name" json:"name"`
	Address     string             `bson:"address" json:"address"`
	City        string             `bson:"city" json:"city"`
	Province    string             `bson:"province" json:"province"`
	ZipCode     string             `bson:"zip_code" json:"zip_code"`
	Country     string             `bson:"country" json:"country"`
	Coordinates struct {
		Latitude  float64 `bson:"latitude" json:"latitude"`
		Longitude float64 `bson:"longitude" json:"longitude"`
	} `bson:"coordinates" json:"coordinates"`
	BrandKey  string    `bson:"brand_key" json:"brand_key"` // Reference to Brand
	IsActive  bool      `bson:"is_active" json:"is_active"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// Brand represents a fuel brand/company
type Brand struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	BrandKey    string             `bson:"brand_key" json:"brand_key"` // Unique identifier (indexed)
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// FuelType represents different types of fuel available
type FuelType struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FuelKey      string             `bson:"fuel_key" json:"fuel_key"`           // Unique identifier (indexed) - e.g., "diesel", "gasoline_95"
	FuelCategory string             `bson:"fuel_category" json:"fuel_category"` // e.g., "Diesel", "Gasoline", "Electric"
	Name         string             `bson:"name" json:"name"`
	Description  string             `bson:"description" json:"description"`
	Unit         string             `bson:"unit" json:"unit"` // e.g., "liter", "gallon", "kWh"
	IsActive     bool               `bson:"is_active" json:"is_active"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// FuelData represents the price data stored in PostgreSQL
// Uses string keys to reference MongoDB collections
type FuelData struct {
	gorm.Model
	FuelKey      string    `gorm:"type:varchar(100);not null;index" json:"fuel_key"`     // References FuelType.fuel_key in MongoDB
	FuelCategory string    `gorm:"type:varchar(50);not null;index" json:"fuel_category"` // References FuelType.fuel_category in MongoDB
	Price        float64   `gorm:"type:decimal(10,3);not null" json:"price"`             // Current price
	LocationKey  string    `gorm:"type:varchar(100);not null;index" json:"location_key"` // References Location.location_key in MongoDB
	LastUpdate   time.Time `gorm:"type:timestamp;not null;index" json:"last_update"`     // When price was last updated
	Currency     string    `gorm:"type:varchar(3);default:'EUR'" json:"currency"`        // e.g., "EUR", "USD"
	IsActive     bool      `gorm:"default:true;index" json:"is_active"`                  // Soft delete flag
}

const (
	UserPreferencesCollection = "user_preferences"
	LocationsCollection       = "locations"
	BrandsCollection          = "brands"
	FuelTypesCollection       = "fuel_types"
)

// GenerateFuelKey creates a consistent fuel key from category and type
// Example: "gasoline_95", "diesel_premium", "electric_fast"
func GenerateFuelKey(category, fuelType string) string {
	//TODO
	return "" // placeholder
}

// GenerateLocationKey creates a consistent location key
// Example: "shell_main_street_123", "eni_highway_a1_km45"
func GenerateLocationKey(brand, address string) string {
	//TODO
	return "" // placeholder
}

package db

import (
	"Irlandesee/GO-Fuels/src/models"
	"context"
	"fmt"
	"math"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoHandler struct {
	Client   *mongo.Client
	Database *mongo.Database
}

type MongoConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
}

func NewMongoHandler(connString, dbName string) (*MongoHandler, error) {
	clientOptions := options.Client().ApplyURI(connString)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("could not connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(dbName)
	return &MongoHandler{
		Client:   client,
		Database: database,
	}, nil
}

func (h *MongoHandler) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return h.Client.Disconnect(ctx)
}

func BuildMongoConnectionString(config MongoConfig) string {
	user := url.QueryEscape(config.User)
	password := url.QueryEscape(config.Password)

	var connStr string
	if user != "" && password != "" {
		connStr = fmt.Sprintf("mongodb://%s:%s@%s", user, password, config.Host)
	} else {
		connStr = fmt.Sprintf("mongodb://%s", config.Host)
	}

	if config.Port != "" {
		connStr = fmt.Sprintf("%s:%s", connStr, config.Port)
	}
	if config.Database != "" {
		connStr = fmt.Sprintf("%s/%s", connStr, config.Database)
	}
	return connStr
}

// ─── FuelType ────────────────────────────────────────────────────────────────

func (h *MongoHandler) CreateFuelType(ctx context.Context, ft *models.FuelType) error {
	ft.ID = primitive.NewObjectID()
	ft.CreatedAt = time.Now()
	ft.UpdatedAt = time.Now()

	_, err := h.Database.Collection(models.FuelTypesCollection).InsertOne(ctx, ft)
	return err
}

func (h *MongoHandler) GetAllFuelTypes(ctx context.Context) ([]models.FuelType, error) {
	cursor, err := h.Database.Collection(models.FuelTypesCollection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.FuelType
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (h *MongoHandler) GetFuelTypeByKey(ctx context.Context, fuelKey string) (*models.FuelType, error) {
	var result models.FuelType
	err := h.Database.Collection(models.FuelTypesCollection).
		FindOne(ctx, bson.M{"fuel_key": fuelKey}).
		Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *MongoHandler) GetFuelTypesByCategory(ctx context.Context, category string) ([]models.FuelType, error) {
	cursor, err := h.Database.Collection(models.FuelTypesCollection).
		Find(ctx, bson.M{"fuel_category": category})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.FuelType
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// ─── Brand ───────────────────────────────────────────────────────────────────

func (h *MongoHandler) CreateBrand(ctx context.Context, b *models.Brand) error {
	b.ID = primitive.NewObjectID()
	b.CreatedAt = time.Now()
	b.UpdatedAt = time.Now()

	_, err := h.Database.Collection(models.BrandsCollection).InsertOne(ctx, b)
	return err
}

func (h *MongoHandler) GetBrandByKey(ctx context.Context, brandKey string) (*models.Brand, error) {
	var result models.Brand
	err := h.Database.Collection(models.BrandsCollection).
		FindOne(ctx, bson.M{"brand_key": brandKey}).
		Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *MongoHandler) GetAllBrands(ctx context.Context) ([]models.Brand, error) {
	cursor, err := h.Database.Collection(models.BrandsCollection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.Brand
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (h *MongoHandler) UpdateBrand(ctx context.Context, brandKey string, b *models.Brand) error {
	b.UpdatedAt = time.Now()
	_, err := h.Database.Collection(models.BrandsCollection).UpdateOne(
		ctx,
		bson.M{"brand_key": brandKey},
		bson.M{"$set": b},
	)
	return err
}

// ─── Location ────────────────────────────────────────────────────────────────

func (h *MongoHandler) CreateLocation(ctx context.Context, l *models.Location) error {
	l.ID = primitive.NewObjectID()
	l.CreatedAt = time.Now()
	l.UpdatedAt = time.Now()

	_, err := h.Database.Collection(models.LocationsCollection).InsertOne(ctx, l)
	return err
}

func (h *MongoHandler) GetLocationByKey(ctx context.Context, locationKey string) (*models.Location, error) {
	var result models.Location
	err := h.Database.Collection(models.LocationsCollection).
		FindOne(ctx, bson.M{"location_key": locationKey}).
		Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *MongoHandler) UpdateLocation(ctx context.Context, locationKey string, l *models.Location) error {
	l.UpdatedAt = time.Now()
	_, err := h.Database.Collection(models.LocationsCollection).UpdateOne(
		ctx,
		bson.M{"location_key": locationKey},
		bson.M{"$set": l},
	)
	return err
}

func (h *MongoHandler) GetLocationsByBrand(ctx context.Context, brandKey string) ([]models.Location, error) {
	cursor, err := h.Database.Collection(models.LocationsCollection).
		Find(ctx, bson.M{"brand_key": brandKey})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.Location
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetNearbyLocations uses a bounding-box pre-filter then Haversine for accuracy.
func (h *MongoHandler) GetNearbyLocations(ctx context.Context, lat, lng, radiusKm float64) ([]models.Location, error) {
	// Rough degree deltas for the bounding box
	latDelta := radiusKm / 111.0
	lngDelta := radiusKm / (111.0 * math.Cos(lat*math.Pi/180))

	filter := bson.M{
		"coordinates.latitude":  bson.M{"$gte": lat - latDelta, "$lte": lat + latDelta},
		"coordinates.longitude": bson.M{"$gte": lng - lngDelta, "$lte": lng + lngDelta},
		"is_active":             true,
	}

	cursor, err := h.Database.Collection(models.LocationsCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var candidates []models.Location
	if err := cursor.All(ctx, &candidates); err != nil {
		return nil, err
	}

	// Haversine refinement
	var results []models.Location
	for _, loc := range candidates {
		if haversineKm(lat, lng, loc.Coordinates.Latitude, loc.Coordinates.Longitude) <= radiusKm {
			results = append(results, loc)
		}
	}
	return results, nil
}

// LocationWithPrices is the combined response for cross-DB queries.
type LocationWithPrices struct {
	Location models.Location   `json:"location"`
	Prices   []models.FuelData `json:"prices"`
}

// GetLocationWithPrices fetches location from Mongo; the handler layer must
// inject Postgres prices. Here we return just the location so the router can
// enrich it.
func (h *MongoHandler) GetLocationWithPrices(ctx context.Context, locationKey string) (*LocationWithPrices, error) {
	location, err := h.GetLocationByKey(ctx, locationKey)
	if err != nil {
		return nil, err
	}
	return &LocationWithPrices{Location: *location}, nil
}

// GetNearbyLocationsWithPrices returns nearby locations; prices must be
// enriched by the handler using Postgres (same pattern as above).
func (h *MongoHandler) GetNearbyLocationsWithPrices(ctx context.Context, lat, lng, radiusKm float64, fuelKey string) ([]LocationWithPrices, error) {
	locations, err := h.GetNearbyLocations(ctx, lat, lng, radiusKm)
	if err != nil {
		return nil, err
	}

	results := make([]LocationWithPrices, len(locations))
	for i, loc := range locations {
		results[i] = LocationWithPrices{Location: loc}
	}
	return results, nil
}

// ─── UserPreferences ─────────────────────────────────────────────────────────

func (h *MongoHandler) CreateUserPreference(ctx context.Context, p *models.UserPreferences) error {
	p.ID = primitive.NewObjectID()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	_, err := h.Database.Collection(models.UserPreferencesCollection).InsertOne(ctx, p)
	return err
}

func (h *MongoHandler) GetUserPreferenceByUserID(ctx context.Context, userID string) (*models.UserPreferences, error) {
	var result models.UserPreferences
	err := h.Database.Collection(models.UserPreferencesCollection).
		FindOne(ctx, bson.M{"user_id": userID}).
		Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (h *MongoHandler) UpdateUserPreference(ctx context.Context, userID string, p *models.UserPreferences) error {
	p.UpdatedAt = time.Now()
	_, err := h.Database.Collection(models.UserPreferencesCollection).UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{"$set": p},
	)
	return err
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

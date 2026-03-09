package routers

import (
	"Irlandesee/GO-Fuels/src/models"
	"Irlandesee/GO-Fuels/src/utils/dbHandlers"
	"net/http"

	"github.com/labstack/echo/v4"
)

type LocationHandler struct {
	DB *dbHandlers.Handler
}

func NewLocationHandler(db *dbHandlers.Handler) *LocationHandler {
	return &LocationHandler{DB: db}
}

func (h *LocationHandler) CreateLocation(c echo.Context) error {
	var req models.Location
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.DB.Mongo.CreateLocation(c.Request().Context(), &req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, req)
}

func (h *LocationHandler) GetLocationByKey(c echo.Context) error {
	locationKey := c.Param("location_key")

	location, err := h.DB.Mongo.GetLocationByKey(c.Request().Context(), locationKey)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, location)
}

func (h *LocationHandler) UpdateLocation(c echo.Context) error {
	locationKey := c.Param("location_key")

	var req models.Location
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.DB.Mongo.UpdateLocation(c.Request().Context(), locationKey, &req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, req)
}

func (h *LocationHandler) GetLocationsByBrand(c echo.Context) error {
	brandKey := c.Param("brand_key")

	locations, err := h.DB.Mongo.GetLocationsByBrand(c.Request().Context(), brandKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, locations)
}

type NearbyLocationsRequest struct {
	Lat    float64 `query:"lat" validate:"required"`
	Lng    float64 `query:"lng" validate:"required"`
	Radius float64 `query:"radius"` // in kilometers
}

func (h *LocationHandler) GetNearbyLocations(c echo.Context) error {
	var req NearbyLocationsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid query parameters"})
	}

	if req.Radius <= 0 {
		req.Radius = 5.0 // Default radius: 5km
	}

	locations, err := h.DB.Mongo.GetNearbyLocations(c.Request().Context(), req.Lat, req.Lng, req.Radius)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, locations)
}

// GetLocationWithPrices fetches the location from Mongo, then enriches it
// with current prices from Postgres — the correct pattern for cross-DB queries.
func (h *LocationHandler) GetLocationWithPrices(c echo.Context) error {
	locationKey := c.Param("location_key")

	// 1. Fetch location metadata from MongoDB
	partial, err := h.DB.Mongo.GetLocationWithPrices(c.Request().Context(), locationKey)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	// 2. Enrich with prices from PostgreSQL
	prices, err := h.DB.Postgres.GetFuelDataByLocation(c.Request().Context(), locationKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	partial.Prices = prices

	return c.JSON(http.StatusOK, partial)
}

type NearbyLocationsWithPricesRequest struct {
	Lat     float64 `query:"lat" validate:"required"`
	Lng     float64 `query:"lng" validate:"required"`
	Radius  float64 `query:"radius"` // in kilometers
	FuelKey string  `query:"fuel_key" validate:"required"`
}

// GetNearbyLocationsWithPrices fetches nearby locations, then enriches each
// with its current fuel prices from Postgres.
func (h *LocationHandler) GetNearbyLocationsWithPrices(c echo.Context) error {
	var req NearbyLocationsWithPricesRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid query parameters"})
	}
	if req.Radius <= 0 {
		req.Radius = 5.0
	}

	// 1. Fetch nearby locations from MongoDB
	partials, err := h.DB.Mongo.GetNearbyLocationsWithPrices(
		c.Request().Context(), req.Lat, req.Lng, req.Radius, req.FuelKey,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// 2. Enrich each location with Postgres prices, filtered by fuel_key
	for i := range partials {
		locationKey := partials[i].Location.LocationKey
		var prices []models.FuelData
		var pgErr error

		if req.FuelKey != "" {
			fd, e := h.DB.Postgres.GetFuelDataByLocationAndFuel(
				c.Request().Context(), locationKey, req.FuelKey,
			)
			if e == nil {
				prices = []models.FuelData{*fd}
			} else {
				pgErr = e
			}
		} else {
			prices, pgErr = h.DB.Postgres.GetFuelDataByLocation(c.Request().Context(), locationKey)
		}

		if pgErr == nil {
			partials[i].Prices = prices
		}
		// If Postgres lookup fails for one location, continue — don't abort the whole response
	}

	return c.JSON(http.StatusOK, partials)
}

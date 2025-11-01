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

func (h *LocationHandler) GetLocationWithPrices(c echo.Context) error {
	locationKey := c.Param("location_key")

	//TODO: handle cross db query
	result, err := h.DB.Mongo.GetLocationWithPrices(c.Request().Context(), locationKey)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

type NearbyLocationsWithPricesRequest struct {
	Lat     float64 `query:"lat" validate:"required"`
	Lng     float64 `query:"lng" validate:"required"`
	Radius  float64 `query:"radius"` // in kilometers
	FuelKey string  `query:"fuel_key" validate:"required"`
}

func (h *LocationHandler) GetNearbyLocationsWithPrices(c echo.Context) error {
	var req NearbyLocationsWithPricesRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid query parameters"})
	}

	if req.Radius <= 0 {
		req.Radius = 5.0 // Default radius: 5km
	}

	result, err := h.DB.Mongo.GetNearbyLocationsWithPrices(c.Request().Context(), req.Lat, req.Lng, req.Radius, req.FuelKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

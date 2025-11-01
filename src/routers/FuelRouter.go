package routers

import (
	"Irlandesee/GO-Fuels/src/models"
	"Irlandesee/GO-Fuels/src/utils/dbHandlers"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type FuelHandler struct {
	DB *dbHandlers.Handler
}

func NewFuelHandler(db *dbHandlers.Handler) *FuelHandler {
	return &FuelHandler{DB: db}
}

func (h *FuelHandler) CreateFuelType(c echo.Context) error {
	var req models.FuelType
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.DB.Mongo.CreateFuelType(c.Request().Context(), &req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, req)
}

func (h *FuelHandler) GetAllFuelTypes(c echo.Context) error {
	fuelTypes, err := h.DB.Mongo.GetAllFuelTypes(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, fuelTypes)
}

func (h *FuelHandler) GetFuelTypeByKey(c echo.Context) error {
	fuelKey := c.Param("fuel_key")

	fuelType, err := h.DB.Mongo.GetFuelTypeByKey(c.Request().Context(), fuelKey)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, fuelType)
}

func (h *FuelHandler) GetFuelTypesByCategory(c echo.Context) error {
	category := c.Param("category")

	fuelTypes, err := h.DB.Mongo.GetFuelTypesByCategory(c.Request().Context(), category)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, fuelTypes)
}

func (h *FuelHandler) CreateFuelData(c echo.Context) error {
	var req models.FuelData
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.DB.Postgres.CreateFuelData(c.Request().Context(), &req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, req)
}

func (h *FuelHandler) UpsertFuelData(c echo.Context) error {
	var req models.FuelData
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.DB.Postgres.UpsertFuelData(c.Request().Context(), &req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, req)
}

func (h *FuelHandler) GetFuelDataByLocation(c echo.Context) error {
	locationKey := c.Param("location_key")

	data, err := h.DB.Postgres.GetFuelDataByLocation(c.Request().Context(), locationKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, data)
}

func (h *FuelHandler) GetFuelDataByLocationAndFuel(c echo.Context) error {
	locationKey := c.Param("location_key")
	fuelKey := c.Param("fuel_key")

	data, err := h.DB.Postgres.GetFuelDataByLocationAndFuel(c.Request().Context(), locationKey, fuelKey)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, data)
}

type PriceHistoryRequest struct {
	LocationKey string `query:"location_key" validate:"required"`
	FuelKey     string `query:"fuel_key" validate:"required"`
	Since       string `query:"since"` // ISO 8601 format
}

func (h *FuelHandler) GetPriceHistory(c echo.Context) error {
	var req PriceHistoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid query parameters"})
	}

	// Parse since date, default to 30 days ago
	var since time.Time
	if req.Since != "" {
		parsed, err := time.Parse(time.RFC3339, req.Since)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid date format, use ISO 8601"})
		}
		since = parsed
	} else {
		since = time.Now().AddDate(0, 0, -30) // Default: last 30 days
	}

	data, err := h.DB.Postgres.GetPriceHistory(c.Request().Context(), req.LocationKey, req.FuelKey, since)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, data)
}

func (h *FuelHandler) GetAveragePrice(c echo.Context) error {
	fuelKey := c.Param("fuel_key")
	sinceStr := c.QueryParam("since")

	// Parse since date, default to 7 days ago
	var since time.Time
	if sinceStr != "" {
		parsed, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid date format, use ISO 8601"})
		}
		since = parsed
	} else {
		since = time.Now().AddDate(0, 0, -7) // Default: last 7 days
	}

	avgPrice, err := h.DB.Postgres.GetAveragePrice(c.Request().Context(), fuelKey, since)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"fuel_key":      fuelKey,
		"average_price": avgPrice,
		"since":         since,
	})
}

func (h *FuelHandler) GetLatestPrices(c echo.Context) error {
	fuelKey := c.Param("fuel_key")
	limitStr := c.QueryParam("limit")

	limit := 10 // Default limit
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}

	data, err := h.DB.Postgres.GetLatestPricesByFuelType(c.Request().Context(), fuelKey, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, data)
}

type UpdatePriceRequest struct {
	LocationKey string  `json:"location_key" validate:"required"`
	FuelKey     string  `json:"fuel_key" validate:"required"`
	Price       float64 `json:"price" validate:"required,gt=0"`
}

func (h *FuelHandler) UpdateFuelPrice(c echo.Context) error {
	var req UpdatePriceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.DB.Postgres.UpdateFuelPrice(c.Request().Context(), req.LocationKey, req.FuelKey, req.Price); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Price updated successfully"})
}

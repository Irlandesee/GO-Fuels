package routers

import (
	"Irlandesee/GO-Fuels/src/models"
	"Irlandesee/GO-Fuels/src/utils/dbHandlers"
	"net/http"

	"github.com/labstack/echo/v4"
)

type BrandHandler struct {
	DB *dbHandlers.Handler
}

func NewBrandHandler(db *dbHandlers.Handler) *BrandHandler {
	return &BrandHandler{DB: db}
}

func (h *BrandHandler) CreateBrand(c echo.Context) error {
	var req models.Brand
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.DB.Mongo.CreateBrand(c.Request().Context(), &req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, req)
}

func (h *BrandHandler) GetBrandByKey(c echo.Context) error {
	brandKey := c.Param("brand_key")

	brand, err := h.DB.Mongo.GetBrandByKey(c.Request().Context(), brandKey)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, brand)
}

func (h *BrandHandler) GetAllBrands(c echo.Context) error {
	brands, err := h.DB.Mongo.GetAllBrands(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, brands)
}

func (h *BrandHandler) UpdateBrand(c echo.Context) error {
	brandKey := c.Param("brand_key")

	var req models.Brand
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.DB.Mongo.UpdateBrand(c.Request().Context(), brandKey, &req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, req)
}

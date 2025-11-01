package routers

import (
	"Irlandesee/GO-Fuels/src/models"
	"Irlandesee/GO-Fuels/src/utils/dbHandlers"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	DB *dbHandlers.Handler
}

func NewUserHandler(db *dbHandlers.Handler) *UserHandler {
	return &UserHandler{DB: db}
}

func (h *UserHandler) CreateUserPreference(c echo.Context) error {
	var req models.UserPreferences
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.DB.Mongo.CreateUserPreference(c.Request().Context(), &req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, req)
}

func (h *UserHandler) GetUserPreference(c echo.Context) error {
	userID := c.Param("user_id")

	pref, err := h.DB.Mongo.GetUserPreferenceByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, pref)
}

func (h *UserHandler) UpdateUserPreference(c echo.Context) error {
	userID := c.Param("user_id")

	var req models.UserPreferences
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.DB.Mongo.UpdateUserPreference(c.Request().Context(), userID, &req); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, req)
}

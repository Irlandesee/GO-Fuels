package routers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"Irlandesee/GO-Fuels/src/models"
	"Irlandesee/GO-Fuels/src/utils/dbHandlers"

	"github.com/labstack/echo/v4"
)

type IngestionHandler struct {
	DB *dbHandlers.Handler
}

func NewIngestionHandler(db *dbHandlers.Handler) *IngestionHandler {
	return &IngestionHandler{DB: db}
}

type CreateIngestionJobRequest struct {
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	Radius int     `json:"radius"`
}

func (h *IngestionHandler) CreateJob(c echo.Context) error {
	var req CreateIngestionJobRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Radius <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Radius must be positive"})
	}

	job := &models.IngestionJob{
		Lat:    req.Lat,
		Lng:    req.Lng,
		Radius: req.Radius,
	}

	if err := h.DB.Postgres.CreateIngestionJob(c.Request().Context(), job); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Publish job ID to RabbitMQ
	msg := models.IngestionJobMessage{JobID: job.ID}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal job message"})
	}

	if err := h.DB.Rabbit.Publish(models.IngestionQueueName, msgBytes); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to enqueue job: " + err.Error()})
	}

	return c.JSON(http.StatusCreated, job)
}

func (h *IngestionHandler) GetJob(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid job ID"})
	}

	job, err := h.DB.Postgres.GetIngestionJob(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Job not found"})
	}

	return c.JSON(http.StatusOK, job)
}

func (h *IngestionHandler) ListJobs(c echo.Context) error {
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit := 20
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	jobs, err := h.DB.Postgres.ListIngestionJobs(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, jobs)
}

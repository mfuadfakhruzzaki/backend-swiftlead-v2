package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
	"github.com/swiftlead/backend-swiftlet/internal/services"
	"github.com/swiftlead/backend-swiftlet/pkg/response"
)

// SensorHandler handles sensor endpoints
type SensorHandler struct {
	sensorService    *services.SensorService
	telemetryService *services.TelemetryService
}

func NewSensorHandler(sensorService *services.SensorService, telemetryService *services.TelemetryService) *SensorHandler {
	return &SensorHandler{
		sensorService:    sensorService,
		telemetryService: telemetryService,
	}
}

// Get handles GET /sensors/{sensor_id}
func (h *SensorHandler) Get(w http.ResponseWriter, r *http.Request) {
	sensorID := chi.URLParam(r, "sensor_id")

	sensor, err := h.sensorService.GetByID(r.Context(), sensorID)
	if err != nil {
		if err == repository.ErrSensorNotFound {
			response.NotFound(w, "Sensor not found")
		} else {
			response.InternalError(w, "Failed to get sensor")
		}
		return
	}

	response.Success(w, "", sensor)
}

// Update handles PATCH /sensors/{sensor_id}
func (h *SensorHandler) Update(w http.ResponseWriter, r *http.Request) {
	sensorID := chi.URLParam(r, "sensor_id")

	var req models.UpdateSensorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	sensor, err := h.sensorService.Update(r.Context(), sensorID, &req)
	if err != nil {
		if err == repository.ErrSensorNotFound {
			response.NotFound(w, "Sensor not found")
		} else {
			response.InternalError(w, "Failed to update sensor")
		}
		return
	}

	response.Success(w, "Sensor updated", sensor)
}

// GetReadings handles GET /sensors/{sensor_id}/readings
func (h *SensorHandler) GetReadings(w http.ResponseWriter, r *http.Request) {
	sensorID := chi.URLParam(r, "sensor_id")

	// Parse query params
	var from, to time.Time
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		from, _ = time.Parse(time.RFC3339, fromStr)
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		to, _ = time.Parse(time.RFC3339, toStr)
	}

	_, limit := getPagination(r)

	readings, err := h.telemetryService.GetReadings(r.Context(), sensorID, from, to, limit)
	if err != nil {
		response.InternalError(w, "Failed to get readings")
		return
	}

	response.Success(w, "", readings)
}

// CreateReading handles POST /sensors/{sensor_id}/readings
func (h *SensorHandler) CreateReading(w http.ResponseWriter, r *http.Request) {
	sensorID := chi.URLParam(r, "sensor_id")

	var req models.CreateSensorReadingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	reading, err := h.telemetryService.CreateReading(r.Context(), sensorID, &req)
	if err != nil {
		response.InternalError(w, "Failed to create reading")
		return
	}

	response.Created(w, "Reading created", reading)
}

package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/swiftlead/backend-swiftlet/internal/ai"
	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/services"
	"github.com/swiftlead/backend-swiftlet/pkg/logger"
	"github.com/swiftlead/backend-swiftlet/pkg/response"
)

// AIHandler handles AI proxy endpoints
type AIHandler struct {
	aiClient        *ai.Client
	sensorService   *services.SensorService
	telemetryService *services.TelemetryService
}

func NewAIHandler(aiClient *ai.Client, sensorSvc *services.SensorService, telemetrySvc *services.TelemetryService) *AIHandler {
	return &AIHandler{
		aiClient:        aiClient,
		sensorService:   sensorSvc,
		telemetryService: telemetrySvc,
	}
}

// HealthCheck handles GET /ai/health
func (h *AIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health, err := h.aiClient.HealthCheck(r.Context())
	if err != nil {
		logger.Error("AI Engine health check failed: %v", err)
		response.InternalError(w, "AI Engine health check failed")
		return
	}
	response.Success(w, "", health)
}

// PredictGrade handles POST /ai/predict-grade
func (h *AIHandler) PredictGrade(w http.ResponseWriter, r *http.Request) {
	if !h.aiClient.IsEnabled() {
		response.Error(w, http.StatusServiceUnavailable, "AI Engine is disabled", "ai_disabled")
		return
	}

	var req ai.GradePredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	result, err := h.aiClient.PredictGrade(r.Context(), &req)
	if err != nil {
		logger.Error("AI grade prediction failed: %v", err)
		response.InternalError(w, "AI grade prediction failed")
		return
	}

	response.Success(w, "", result)
}

// PredictPump handles POST /ai/predict-pump
func (h *AIHandler) PredictPump(w http.ResponseWriter, r *http.Request) {
	if !h.aiClient.IsEnabled() {
		response.Error(w, http.StatusServiceUnavailable, "AI Engine is disabled", "ai_disabled")
		return
	}

	var req ai.PumpPredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	result, err := h.aiClient.PredictPump(r.Context(), &req)
	if err != nil {
		logger.Error("AI pump prediction failed: %v", err)
		response.InternalError(w, "AI pump prediction failed")
		return
	}

	response.Success(w, "", result)
}

// Analyze handles POST /ai/analyze — caller supplies sensor values manually
func (h *AIHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	if !h.aiClient.IsEnabled() {
		response.Error(w, http.StatusServiceUnavailable, "AI Engine is disabled", "ai_disabled")
		return
	}

	var req ai.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	result, err := h.aiClient.Analyze(r.Context(), &req)
	if err != nil {
		logger.Error("AI analysis failed: %v", err)
		response.InternalError(w, "AI analysis failed")
		return
	}

	response.Success(w, "", result)
}

// AnalyzeNode handles POST /nodes/{node_id}/ai/analyze
// Automatically fetches the latest sensor readings for the node from DB,
// then requests a comprehensive AI analysis — no need to supply sensor values manually.
func (h *AIHandler) AnalyzeNode(w http.ResponseWriter, r *http.Request) {
	if !h.aiClient.IsEnabled() {
		response.Error(w, http.StatusServiceUnavailable, "AI Engine is disabled", "ai_disabled")
		return
	}

	nodeID := chi.URLParam(r, "node_id")
	if nodeID == "" {
		response.BadRequest(w, "node_id is required")
		return
	}

	// Fetch all sensors for this node
	sensors, err := h.sensorService.ListByNode(r.Context(), nodeID)
	if err != nil {
		logger.Error("Failed to list sensors for node %s: %v", nodeID, err)
		response.InternalError(w, "Failed to fetch sensor list")
		return
	}

	var temp, humid, ammonia float64
	var rbwID string

	// Fetch latest reading per sensor type
	for _, sensor := range sensors {
		val, err := h.fetchLatestValue(r.Context(), sensor.ID)
		if err != nil || val == 0 {
			continue
		}
		switch sensor.SensorType {
		case models.SensorTypeTemp:
			temp = val
		case models.SensorTypeHumid:
			humid = val
		case models.SensorTypeAmmonia:
			ammonia = val
		}
	}

	// Optional: allow caller to pass rbw_id via query param
	rbwID = r.URL.Query().Get("rbw_id")

	req := ai.AnalyzeRequest{
		NodeID:      nodeID,
		RBWID:       rbwID,
		Temperature: temp,
		Humidity:    humid,
		Ammonia:     ammonia,
	}

	result, err := h.aiClient.Analyze(r.Context(), &req)
	if err != nil {
		logger.Error("AI node analysis failed: %v", err)
		response.InternalError(w, "AI analysis failed")
		return
	}

	response.Success(w, "", result)
}

// AnomalyDetect handles POST /ai/anomaly-detect
func (h *AIHandler) AnomalyDetect(w http.ResponseWriter, r *http.Request) {
	if !h.aiClient.IsEnabled() {
		response.Error(w, http.StatusServiceUnavailable, "AI Engine is disabled", "ai_disabled")
		return
	}

	var req ai.AnomalyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	result, err := h.aiClient.DetectAnomaly(r.Context(), &req)
	if err != nil {
		logger.Error("AI anomaly detection failed: %v", err)
		response.InternalError(w, "AI anomaly detection failed")
		return
	}

	response.Success(w, "", result)
}

// fetchLatestValue returns the latest reading value for a sensor, or 0 if not found
func (h *AIHandler) fetchLatestValue(ctx context.Context, sensorID string) (float64, error) {
	reading, err := h.telemetryService.GetLatestReading(ctx, sensorID)
	if err != nil {
		return 0, err
	}
	if reading == nil {
		return 0, nil
	}
	return reading.Value, nil
}

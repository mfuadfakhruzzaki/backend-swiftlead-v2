package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/swiftlead/backend-swiftlet/internal/ai"
	"github.com/swiftlead/backend-swiftlet/pkg/response"
)

// AIHandler handles AI proxy endpoints
type AIHandler struct {
	aiClient *ai.Client
}

func NewAIHandler(aiClient *ai.Client) *AIHandler {
	return &AIHandler{aiClient: aiClient}
}

// HealthCheck handles GET /ai/health
func (h *AIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health, err := h.aiClient.HealthCheck(r.Context())
	if err != nil {
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
		response.InternalError(w, "AI pump prediction failed")
		return
	}

	response.Success(w, "", result)
}

// Analyze handles POST /ai/analyze
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
		response.InternalError(w, "AI anomaly detection failed")
		return
	}

	response.Success(w, "", result)
}

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/swiftlead/backend-swiftlet/internal/ai"
	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/services"
	"github.com/swiftlead/backend-swiftlet/pkg/logger"
	"github.com/swiftlead/backend-swiftlet/pkg/response"
)

// AIHandler handles AI proxy endpoints
type AIHandler struct {
	aiClient         *ai.Client
	sensorService    *services.SensorService
	telemetryService *services.TelemetryService
	nodeService      *services.NodeService
}

func NewAIHandler(
	aiClient *ai.Client,
	sensorSvc *services.SensorService,
	telemetrySvc *services.TelemetryService,
	nodeSvc *services.NodeService,
) *AIHandler {
	return &AIHandler{
		aiClient:         aiClient,
		sensorService:    sensorSvc,
		telemetryService: telemetrySvc,
		nodeService:      nodeSvc,
	}
}

// ────────────────────────────────────────────────────────────
// Manual endpoints (caller supplies sensor values)
// ────────────────────────────────────────────────────────────

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

// PredictGrade handles POST /ai/predict-grade (manual body)
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

// PredictPump handles POST /ai/predict-pump (manual body)
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

// Analyze handles POST /ai/analyze (manual body)
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

// AnomalyDetect handles POST /ai/anomaly-detect (manual body)
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

// ────────────────────────────────────────────────────────────
// Node-based endpoints (auto-fetch real-time data from DB)
// ────────────────────────────────────────────────────────────

// AnalyzeNode handles POST /nodes/{node_id}/ai/analyze
// Fetches latest sensor readings from DB, no request body needed.
func (h *AIHandler) AnalyzeNode(w http.ResponseWriter, r *http.Request) {
	if !h.aiClient.IsEnabled() {
		response.Error(w, http.StatusServiceUnavailable, "AI Engine is disabled", "ai_disabled")
		return
	}
	nodeID := chi.URLParam(r, "node_id")

	temp, humid, ammonia, rbwID, err := h.fetchLatestReadings(r.Context(), nodeID)
	if err != nil {
		logger.Error("Failed to fetch readings for node %s: %v", nodeID, err)
		response.InternalError(w, "Failed to fetch sensor readings")
		return
	}

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

// PredictPumpNode handles POST /nodes/{node_id}/ai/predict-pump
// Fetches latest sensor readings + node pump state from DB automatically.
func (h *AIHandler) PredictPumpNode(w http.ResponseWriter, r *http.Request) {
	if !h.aiClient.IsEnabled() {
		response.Error(w, http.StatusServiceUnavailable, "AI Engine is disabled", "ai_disabled")
		return
	}
	nodeID := chi.URLParam(r, "node_id")

	temp, humid, ammonia, rbwID, err := h.fetchLatestReadings(r.Context(), nodeID)
	if err != nil {
		logger.Error("Failed to fetch readings for node %s: %v", nodeID, err)
		response.InternalError(w, "Failed to fetch sensor readings")
		return
	}

	// Fetch pump state from node record
	pumpOn := false
	node, err := h.nodeService.GetByID(r.Context(), nodeID)
	if err == nil && node != nil && node.StatePump != nil {
		pumpOn = *node.StatePump
	}

	req := ai.PumpPredictionRequest{
		NodeID:          nodeID,
		RBWID:           rbwID,
		CurrentTemp:     temp,
		CurrentHumid:    humid,
		CurrentAmmonia:  &ammonia,
		PumpCurrentlyOn: pumpOn,
		UseML:           true,
	}
	result, err := h.aiClient.PredictPump(r.Context(), &req)
	if err != nil {
		logger.Error("AI pump prediction failed for node %s: %v", nodeID, err)
		response.InternalError(w, "AI pump prediction failed")
		return
	}
	response.Success(w, "", result)
}

// PredictGradeNode handles POST /nodes/{node_id}/ai/predict-grade
// Computes 7-day sensor averages from DB. Caller may supply harvest data in body.
func (h *AIHandler) PredictGradeNode(w http.ResponseWriter, r *http.Request) {
	if !h.aiClient.IsEnabled() {
		response.Error(w, http.StatusServiceUnavailable, "AI Engine is disabled", "ai_disabled")
		return
	}
	nodeID := chi.URLParam(r, "node_id")

	// Optional harvest context from request body
	var body struct {
		FloorNo          int     `json:"floor_no"`
		NestsCount       int     `json:"nests_count"`
		WeightKg         float64 `json:"weight_kg"`
		DaysSinceHarvest *int    `json:"days_since_last_harvest,omitempty"`
	}
	// Ignore decode error — fields have zero-value defaults
	_ = json.NewDecoder(r.Body).Decode(&body)

	// Fetch 7-day sensor averages from DB
	avgTemp, avgHumid, avgAmmonia, rbwID, err := h.compute7DayAverages(r.Context(), nodeID)
	if err != nil {
		logger.Error("Failed to compute 7-day averages for node %s: %v", nodeID, err)
		response.InternalError(w, "Failed to fetch sensor history")
		return
	}

	req := ai.GradePredictionRequest{
		RBWID:            rbwID,
		NodeID:           nodeID,
		FloorNo:          body.FloorNo,
		NestsCount:       body.NestsCount,
		WeightKg:         body.WeightKg,
		AvgTemp7Days:     &avgTemp,
		AvgHumid7Days:    &avgHumid,
		AvgAmmonia7Days:  &avgAmmonia,
		DaysSinceHarvest: body.DaysSinceHarvest,
	}
	result, err := h.aiClient.PredictGrade(r.Context(), &req)
	if err != nil {
		logger.Error("AI grade prediction failed for node %s: %v", nodeID, err)
		response.InternalError(w, "AI grade prediction failed")
		return
	}
	response.Success(w, "", result)
}

// AnomalyDetectNode handles POST /nodes/{node_id}/ai/anomaly-detect
// Runs anomaly detection for each sensor using the latest DB readings.
func (h *AIHandler) AnomalyDetectNode(w http.ResponseWriter, r *http.Request) {
	if !h.aiClient.IsEnabled() {
		response.Error(w, http.StatusServiceUnavailable, "AI Engine is disabled", "ai_disabled")
		return
	}
	nodeID := chi.URLParam(r, "node_id")

	sensors, err := h.sensorService.ListByNode(r.Context(), nodeID)
	if err != nil {
		logger.Error("Failed to list sensors for node %s: %v", nodeID, err)
		response.InternalError(w, "Failed to fetch sensor list")
		return
	}

	// Resolve node for rbw_id
	rbwID := ""
	if node, err := h.nodeService.GetByID(r.Context(), nodeID); err == nil && node != nil {
		rbwID = node.RBWID
	}

	type sensorResult struct {
		SensorID   string  `json:"sensor_id"`
		SensorType string  `json:"sensor_type"`
		Value      float64 `json:"value"`
		IsAnomaly  bool    `json:"is_anomaly"`
		Score      float64 `json:"score"`
		Reason     *string `json:"reason,omitempty"`
	}

	var results []sensorResult
	overallAnomaly := false

	for _, sensor := range sensors {
		reading, err := h.telemetryService.GetLatestReading(r.Context(), sensor.ID)
		if err != nil || reading == nil {
			continue
		}

		req := ai.AnomalyRequest{
			SensorID:   sensor.ID,
			SensorType: sensor.SensorType,
			RBWID:      rbwID,
			NodeID:     nodeID,
			RecordedAt: reading.RecordedAt,
			Value:      reading.Value,
		}
		resp, err := h.aiClient.DetectAnomaly(r.Context(), &req)
		if err != nil {
			logger.Warn("Anomaly detect failed for sensor %s: %v", sensor.ID, err)
			continue
		}

		if resp.IsAnomaly {
			overallAnomaly = true
		}
		results = append(results, sensorResult{
			SensorID:   sensor.ID,
			SensorType: sensor.SensorType,
			Value:      reading.Value,
			IsAnomaly:  resp.IsAnomaly,
			Score:      resp.Score,
			Reason:     resp.Reason,
		})
	}

	response.Success(w, "", map[string]interface{}{
		"node_id":         nodeID,
		"rbw_id":          rbwID,
		"overall_anomaly": overallAnomaly,
		"sensors":         results,
	})
}

// ────────────────────────────────────────────────────────────
// Private helpers
// ────────────────────────────────────────────────────────────

// fetchLatestReadings returns the latest temp, humid, ammonia values and rbw_id for a node
func (h *AIHandler) fetchLatestReadings(ctx context.Context, nodeID string) (temp, humid, ammonia float64, rbwID string, err error) {
	node, nodeErr := h.nodeService.GetByID(ctx, nodeID)
	if nodeErr == nil && node != nil {
		rbwID = node.RBWID
	}

	sensors, err := h.sensorService.ListByNode(ctx, nodeID)
	if err != nil {
		return
	}
	for _, sensor := range sensors {
		reading, rErr := h.telemetryService.GetLatestReading(ctx, sensor.ID)
		if rErr != nil || reading == nil {
			continue
		}
		switch sensor.SensorType {
		case models.SensorTypeTemp:
			temp = reading.Value
		case models.SensorTypeHumid:
			humid = reading.Value
		case models.SensorTypeAmmonia:
			ammonia = reading.Value
		}
	}
	return
}

// compute7DayAverages computes the arithmetic mean of readings over the last 7 days per sensor
func (h *AIHandler) compute7DayAverages(ctx context.Context, nodeID string) (avgTemp, avgHumid, avgAmmonia float64, rbwID string, err error) {
	node, nodeErr := h.nodeService.GetByID(ctx, nodeID)
	if nodeErr == nil && node != nil {
		rbwID = node.RBWID
	}

	sensors, err := h.sensorService.ListByNode(ctx, nodeID)
	if err != nil {
		return
	}

	from := time.Now().AddDate(0, 0, -7)
	to := time.Now()

	avg := func(readings []*models.SensorReading) float64 {
		if len(readings) == 0 {
			return 0
		}
		sum := 0.0
		for _, r := range readings {
			sum += r.Value
		}
		return sum / float64(len(readings))
	}

	for _, sensor := range sensors {
		readings, rErr := h.telemetryService.GetReadings(ctx, sensor.ID, from, to, 10080)
		if rErr != nil || len(readings) == 0 {
			continue
		}
		switch sensor.SensorType {
		case models.SensorTypeTemp:
			avgTemp = avg(readings)
		case models.SensorTypeHumid:
			avgHumid = avg(readings)
		case models.SensorTypeAmmonia:
			avgAmmonia = avg(readings)
		}
	}
	return
}

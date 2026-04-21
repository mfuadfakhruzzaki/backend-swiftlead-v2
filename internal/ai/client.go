package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/swiftlead/backend-swiftlet/pkg/logger"
)

// Client is an HTTP client for the AI Engine
type Client struct {
	baseURL    string
	httpClient *http.Client
	enabled    bool
}

// NewClient creates a new AI Engine client
func NewClient(baseURL string, timeout int, enabled bool) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		enabled: enabled,
	}
}

// IsEnabled returns whether AI Engine is enabled
func (c *Client) IsEnabled() bool {
	return c.enabled
}

// SetTimeout sets the HTTP client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// HealthCheck checks if AI Engine is healthy — calls GET /health
func (c *Client) HealthCheck(ctx context.Context) (*HealthResponse, error) {
	if !c.enabled {
		return &HealthResponse{Status: "disabled"}, nil
	}

	resp, err := c.doRequest(ctx, "GET", "/health", nil)
	if err != nil {
		logger.Warn("AI Engine health check failed to connect: %v", err)
		return &HealthResponse{Status: "unhealthy"}, nil
	}
	defer resp.Body.Close()

	var result HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Warn("AI Engine health check decoding failed: %v", err)
		return &HealthResponse{Status: "unhealthy"}, nil
	}
	return &result, nil
}

// DetectAnomaly detects anomalies in sensor readings — calls POST /v1/anomaly-detect
func (c *Client) DetectAnomaly(ctx context.Context, req *AnomalyRequest) (*AnomalyResponse, error) {
	if !c.enabled {
		return &AnomalyResponse{IsAnomaly: false}, nil
	}

	resp, err := c.doRequest(ctx, "POST", "/v1/anomaly-detect", req)
	if err != nil {
		logger.Warn("AI anomaly detection failed: %v", err)
		return &AnomalyResponse{IsAnomaly: false}, nil
	}
	defer resp.Body.Close()

	var result AnomalyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Warn("AI anomaly detection decoding failed: %v", err)
		return &AnomalyResponse{IsAnomaly: false}, nil
	}
	return &result, nil
}

// PredictGrade predicts harvest grade — calls POST /v1/predict-nest-grade
func (c *Client) PredictGrade(ctx context.Context, req *GradePredictionRequest) (*GradePredictionResponse, error) {
	if !c.enabled {
		return c.fallbackGradePredict(req), nil
	}

	resp, err := c.doRequest(ctx, "POST", "/v1/predict-nest-grade", req)
	if err != nil {
		logger.Warn("AI grade prediction failed: %v", err)
		return c.fallbackGradePredict(req), nil
	}
	defer resp.Body.Close()

	var raw aiGradeResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		logger.Warn("AI grade prediction decoding failed: %v", err)
		return c.fallbackGradePredict(req), nil
	}

	result := &GradePredictionResponse{
		Grade:          MapGradeToDBEnum(raw.PredictedGrade),
		Confidence:     raw.Confidence,
		Recommendation: raw.Recommendation,
	}
	return result, nil
}

// PredictPump recommends pump action — calls POST /v1/recommend-pump-action
func (c *Client) PredictPump(ctx context.Context, req *PumpPredictionRequest) (*PumpPredictionResponse, error) {
	if !c.enabled {
		return &PumpPredictionResponse{PumpState: "off", Confidence: 0}, nil
	}

	// Default use_ml to true if not explicitly set
	body := struct {
		NodeID          string   `json:"node_id"`
		RBWID           string   `json:"rbw_id,omitempty"`
		CurrentTemp     float64  `json:"current_temp"`
		CurrentHumid    float64  `json:"current_humid"`
		CurrentAmmonia  *float64 `json:"current_ammonia,omitempty"`
		PumpCurrentlyOn bool     `json:"pump_currently_on"`
		UseML           bool     `json:"use_ml"`
	}{
		NodeID:          req.NodeID,
		RBWID:           req.RBWID,
		CurrentTemp:     req.CurrentTemp,
		CurrentHumid:    req.CurrentHumid,
		CurrentAmmonia:  req.CurrentAmmonia,
		PumpCurrentlyOn: req.PumpCurrentlyOn,
		UseML:           true,
	}

	resp, err := c.doRequest(ctx, "POST", "/v1/recommend-pump-action", body)
	if err != nil {
		logger.Warn("AI pump prediction failed: %v", err)
		return &PumpPredictionResponse{PumpState: "off", Confidence: 0}, nil
	}
	defer resp.Body.Close()

	var raw aiPumpResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		logger.Warn("AI pump prediction decoding failed: %v", err)
		return &PumpPredictionResponse{PumpState: "off", Confidence: 0}, nil
	}

	return mapPumpResponse(&raw), nil
}

// Analyze performs comprehensive analysis — calls POST /v2/decide
func (c *Client) Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("AI Engine is disabled")
	}

	// Map to /v2/decide request format
	body := struct {
		NodeID      string  `json:"node_id"`
		TemperatureC float64 `json:"temperature_c"`
		HumidityRH  float64 `json:"humidity_rh"`
		NH3PPM      float64 `json:"nh3_ppm"`
		UseBuffer   bool    `json:"use_buffer"`
	}{
		NodeID:      req.NodeID,
		TemperatureC: req.Temperature,
		HumidityRH:  req.Humidity,
		NH3PPM:      req.Ammonia,
		UseBuffer:   true,
	}

	resp, err := c.doRequest(ctx, "POST", "/v2/decide", body)
	if err != nil {
		logger.Warn("AI Engine analyze request failed: %v", err)
		return c.fallbackAnalyze(req), nil
	}
	defer resp.Body.Close()

	var raw aiDecideV2Response
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		logger.Warn("AI Engine analyze decoding failed: %v", err)
		return c.fallbackAnalyze(req), nil
	}

	return c.mapDecideV2ToAnalyze(req, &raw), nil
}

// mapDecideV2ToAnalyze converts /v2/decide response to AnalyzeResponse
func (c *Client) mapDecideV2ToAnalyze(req *AnalyzeRequest, raw *aiDecideV2Response) *AnalyzeResponse {
	grade := MapGradeToDBEnum(raw.Grade)
	mappedProbs := make(map[string]float64)
	for k, v := range raw.Probabilities {
		mappedProbs[MapGradeToDBEnum(k)] = v
	}

	gradePred := &GradePredictionResponse{
		Grade:         grade,
		Probabilities: mappedProbs,
	}
	if len(mappedProbs) > 0 {
		for _, v := range mappedProbs {
			if v > gradePred.Confidence {
				gradePred.Confidence = v
			}
		}
	}

	pumpState := "off"
	if raw.SprayerOn {
		pumpState = "on"
	}
	pumpRec := &PumpPredictionResponse{
		PumpState:  pumpState,
		Confidence: 0.9,
		Reason:     raw.SprayerReason,
	}

	sTemp := evalSensorHealth(req.Temperature, 27, 30, "°C")
	sHumid := evalSensorHealth(req.Humidity, 70, 80, "%")
	sAmm := evalSensorHealth(req.Ammonia, 0, 20, "ppm")
	overall := (sTemp.HealthScore + sHumid.HealthScore + sAmm.HealthScore) / 3.0

	var recs []*Recommendation
	if anomVerdict, ok := raw.Anomaly["verdict"].(string); ok && anomVerdict == "anomaly" {
		recs = append(recs, &Recommendation{
			Priority: "high",
			Type:     "anomaly",
			Message:  "AI mendeteksi pembacaan sensor yang tidak normal. Periksa sensor.",
		})
	}
	if sAmm.Status != "normal" {
		pri := "medium"
		if sAmm.Status == "critical" {
			pri = "high"
		}
		recs = append(recs, &Recommendation{
			Priority: pri,
			Type:     "ventilation",
			Message:  "Tingkatkan ventilasi untuk menurunkan kadar amonia.",
		})
	}
	if sTemp.Status == "critical" && req.Temperature > 30 {
		recs = append(recs, &Recommendation{
			Priority: "high",
			Type:     "temperature",
			Message:  "Suhu kritis tinggi. Pastikan sistem pendingin aktif.",
		})
	}

	return &AnalyzeResponse{
		OverallHealthScore: overall,
		Sensors: map[string]*SensorStatus{
			"temperature": sTemp,
			"humidity":    sHumid,
			"ammonia":     sAmm,
		},
		GradePrediction:    gradePred,
		PumpRecommendation: pumpRec,
		Recommendations:    recs,
	}
}

// doRequest performs an HTTP request to the AI Engine
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, fmt.Errorf("AI Engine returned status %d", resp.StatusCode)
	}

	return resp, nil
}

// mapPumpResponse maps AI engine pump action response to PumpPredictionResponse
func mapPumpResponse(raw *aiPumpResponse) *PumpPredictionResponse {
	pumpState := "off"
	switch raw.Action {
	case "turn_on":
		pumpState = "on"
	case "turn_off":
		pumpState = "off"
	case "keep_current":
		pumpState = "keep_current"
	}

	durationMin := 0.0
	if raw.RecommendedDurationMinutes != nil {
		durationMin = float64(*raw.RecommendedDurationMinutes)
	} else if raw.RecommendedDurationSeconds != nil {
		durationMin = *raw.RecommendedDurationSeconds / 60.0
	}

	return &PumpPredictionResponse{
		PumpState:       pumpState,
		Action:          raw.Action,
		Confidence:      raw.Confidence,
		DurationMinutes: durationMin,
		Reason:          raw.Reason,
		Engine:          raw.Engine,
	}
}

// gradeFromThresholds is a simple threshold-based grade classifier used in fallbacks
func gradeFromThresholds(temp, humid, ammonia float64) *GradePredictionResponse {
	score := 100.0

	if temp < 20 || temp > 35 {
		score -= 40
	} else if temp < 25 || temp > 32 {
		score -= 20
	}

	if humid < 50 || humid > 95 {
		score -= 30
	} else if humid < 65 || humid > 85 {
		score -= 15
	}

	if ammonia > 25 {
		score -= 40
	} else if ammonia > 15 {
		score -= 20
	}

	var grade string
	if score >= 70 {
		grade = "good"
	} else if score >= 40 {
		grade = "medium"
	} else {
		grade = "poor"
	}

	return &GradePredictionResponse{
		Grade:      grade,
		Confidence: 0.5,
		Probabilities: map[string]float64{
			"good":   0.33,
			"medium": 0.34,
			"poor":   0.33,
		},
	}
}

// fallbackGradePredict provides threshold-based grade prediction using 7-day averages if available
func (c *Client) fallbackGradePredict(req *GradePredictionRequest) *GradePredictionResponse {
	temp, humid, ammonia := 29.0, 75.0, 5.0
	if req.AvgTemp7Days != nil {
		temp = *req.AvgTemp7Days
	}
	if req.AvgHumid7Days != nil {
		humid = *req.AvgHumid7Days
	}
	if req.AvgAmmonia7Days != nil {
		ammonia = *req.AvgAmmonia7Days
	}
	return gradeFromThresholds(temp, humid, ammonia)
}

// fallbackAnalyze provides robust fallback analysis if AI Engine is unreachable
func (c *Client) fallbackAnalyze(req *AnalyzeRequest) *AnalyzeResponse {
	grade := gradeFromThresholds(req.Temperature, req.Humidity, req.Ammonia)

	pump := &PumpPredictionResponse{
		PumpState:       "off",
		Confidence:      0.9,
		DurationMinutes: 0.0,
	}

	sTemp := evalSensorHealth(req.Temperature, 28.0, 30.0, "°C")
	sHumid := evalSensorHealth(req.Humidity, 70.0, 80.0, "%")
	sAmm := evalSensorHealth(req.Ammonia, 0.0, 20.0, "ppm")
	overall := (sTemp.HealthScore + sHumid.HealthScore + sAmm.HealthScore) / 3.0

	var recs []*Recommendation
	if sAmm.Status != "normal" {
		pri := "medium"
		if sAmm.Status == "critical" {
			pri = "high"
		}
		recs = append(recs, &Recommendation{
			Priority: pri,
			Type:     "ventilation",
			Message:  "Tingkatkan ventilasi untuk menurunkan kadar amonia (aturan fallback).",
		})
	}
	if sTemp.Status == "critical" && req.Temperature > 30.0 {
		recs = append(recs, &Recommendation{
			Priority: "high",
			Type:     "temperature",
			Message:  "Suhu kritis tinggi. Pastikan sistem pendingin aktif (aturan fallback).",
		})
	}

	return &AnalyzeResponse{
		OverallHealthScore: overall,
		Sensors: map[string]*SensorStatus{
			"temperature": sTemp,
			"humidity":    sHumid,
			"ammonia":     sAmm,
		},
		GradePrediction:    grade,
		PumpRecommendation: pump,
		Recommendations:    recs,
	}
}

// evalSensorHealth computes a SensorStatus based on optimal range
func evalSensorHealth(val, optMin, optMax float64, unit string) *SensorStatus {
	status := "normal"
	score := 100.0

	if val < optMin {
		score -= (optMin - val) * 5
	} else if val > optMax {
		score -= (val - optMax) * 5
	}

	if score < 0 {
		score = 0
	}
	if score <= 50 {
		status = "critical"
	} else if score < 100 {
		status = "warning"
	}

	return &SensorStatus{
		Value:       val,
		Unit:        unit,
		Status:      status,
		HealthScore: score,
	}
}

package ai

import "time"

// AnomalyRequest for anomaly detection — matches /v1/anomaly-detect
type AnomalyRequest struct {
	SensorID   string    `json:"sensor_id"`
	SensorType string    `json:"sensor_type"`
	RBWID      string    `json:"rbw_id"`
	NodeID     string    `json:"node_id"`
	RecordedAt time.Time `json:"recorded_at"`
	Value      float64   `json:"value"`
}

// AnomalyResponse from anomaly detection
type AnomalyResponse struct {
	IsAnomaly bool    `json:"is_anomaly"`
	Score     float64 `json:"score"`
	Reason    *string `json:"reason,omitempty"`
}

// GradePredictionRequest for harvest grade prediction — matches /v1/predict-nest-grade
type GradePredictionRequest struct {
	RBWID            string   `json:"rbw_id"`
	NodeID           string   `json:"node_id,omitempty"`
	FloorNo          int      `json:"floor_no"`
	NestsCount       int      `json:"nests_count"`
	WeightKg         float64  `json:"weight_kg"`
	AvgTemp7Days     *float64 `json:"avg_temp_7days,omitempty"`
	AvgHumid7Days    *float64 `json:"avg_humid_7days,omitempty"`
	AvgAmmonia7Days  *float64 `json:"avg_ammonia_7days,omitempty"`
	DaysSinceHarvest *int     `json:"days_since_last_harvest,omitempty"`
}

// aiGradeResponse is the internal type matching AI engine /v1/predict-nest-grade response
type aiGradeResponse struct {
	PredictedGrade string                 `json:"predicted_grade"`
	Confidence     float64                `json:"confidence"`
	Factors        map[string]interface{} `json:"factors"`
	Recommendation string                 `json:"recommendation"`
	PredictedAt    string                 `json:"predicted_at"`
}

// GradePredictionResponse from grade prediction (returned to frontend)
type GradePredictionResponse struct {
	Grade          string             `json:"grade"`
	Confidence     float64            `json:"confidence"`
	Probabilities  map[string]float64 `json:"probabilities,omitempty"`
	Recommendation string             `json:"recommendation,omitempty"`
}

// PumpPredictionRequest for pump control recommendation — matches /v1/recommend-pump-action
type PumpPredictionRequest struct {
	NodeID          string   `json:"node_id"`
	RBWID           string   `json:"rbw_id,omitempty"`
	CurrentTemp     float64  `json:"current_temp"`
	CurrentHumid    float64  `json:"current_humid"`
	CurrentAmmonia  *float64 `json:"current_ammonia,omitempty"`
	PumpCurrentlyOn bool     `json:"pump_currently_on"`
	UseML           bool     `json:"use_ml"`
}

// aiPumpResponse is the internal type matching AI engine /v1/recommend-pump-action response
type aiPumpResponse struct {
	Action                     string   `json:"action"`
	Reason                     string   `json:"reason"`
	Confidence                 float64  `json:"confidence"`
	RecommendedDurationMinutes *int     `json:"recommended_duration_minutes"`
	RecommendedDurationSeconds *float64 `json:"recommended_duration_seconds"`
	Engine                     string   `json:"engine"`
}

// PumpPredictionResponse from pump prediction (returned to frontend)
type PumpPredictionResponse struct {
	PumpState       string  `json:"pump_state"`       // "on" | "off" | "keep_current"
	Action          string  `json:"action,omitempty"` // raw: "turn_on" | "turn_off" | "keep_current"
	Confidence      float64 `json:"confidence"`
	DurationMinutes float64 `json:"duration_minutes"`
	Reason          string  `json:"reason,omitempty"`
	Engine          string  `json:"engine,omitempty"`
}

// AnalyzeRequest for comprehensive analysis — maps to /v2/decide
type AnalyzeRequest struct {
	NodeID      string  `json:"node_id"`
	RBWID       string  `json:"rbw_id,omitempty"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Ammonia     float64 `json:"ammonia"`
}

// aiDecideV2Response is the internal type matching AI engine /v2/decide response
type aiDecideV2Response struct {
	Grade          string                 `json:"grade"`
	Probabilities  map[string]float64     `json:"probabilities"`
	SprayerOn      bool                   `json:"sprayer_on"`
	SprayerReason  string                 `json:"sprayer_reason"`
	Anomaly        map[string]interface{} `json:"anomaly"`
	UsedThresholds map[string]float64     `json:"used_thresholds"`
	BufferStats    map[string]interface{} `json:"buffer_stats"`
	FeaturesUsed   map[string]float64     `json:"features_used"`
	Note           string                 `json:"note"`
}

// SensorStatus in analysis response
type SensorStatus struct {
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Status      string  `json:"status"`
	HealthScore float64 `json:"health_score"`
}

// Recommendation in analysis response
type Recommendation struct {
	Priority string `json:"priority"`
	Type     string `json:"type"`
	Message  string `json:"message"`
}

// AnalyzeResponse from comprehensive analysis (returned to frontend)
type AnalyzeResponse struct {
	OverallHealthScore float64                  `json:"overall_health_score"`
	Sensors            map[string]*SensorStatus `json:"sensors"`
	GradePrediction    *GradePredictionResponse `json:"grade_prediction"`
	PumpRecommendation *PumpPredictionResponse  `json:"pump_recommendation"`
	Recommendations    []*Recommendation        `json:"recommendations"`
}

// HealthResponse from health check
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}

// MapGradeToDBEnum maps AI grade to database enum
func MapGradeToDBEnum(aiGrade string) string {
	switch aiGrade {
	case "bagus":
		return "good"
	case "sedang":
		return "medium"
	case "buruk":
		return "poor"
	default:
		return aiGrade
	}
}

// MapDBEnumToAIGrade maps database enum to AI grade
func MapDBEnumToAIGrade(dbGrade string) string {
	switch dbGrade {
	case "good":
		return "bagus"
	case "medium":
		return "sedang"
	case "poor":
		return "buruk"
	default:
		return dbGrade
	}
}

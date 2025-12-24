package ai

import "time"

// AnomalyRequest for anomaly detection
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

// GradePredictionRequest for harvest grade prediction
type GradePredictionRequest struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Ammonia     float64 `json:"ammonia"`
	RBWID       string  `json:"rbw_id,omitempty"`
	NodeID      string  `json:"node_id,omitempty"`
}

// GradePredictionResponse from grade prediction
type GradePredictionResponse struct {
	Grade         string             `json:"grade"`
	Confidence    float64            `json:"confidence"`
	Probabilities map[string]float64 `json:"probabilities"`
}

// PumpPredictionRequest for pump control recommendation
type PumpPredictionRequest struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Ammonia     float64 `json:"ammonia"`
	RBWID       string  `json:"rbw_id,omitempty"`
	NodeID      string  `json:"node_id,omitempty"`
}

// PumpPredictionResponse from pump prediction
type PumpPredictionResponse struct {
	PumpState       string  `json:"pump_state"`
	Confidence      float64 `json:"confidence"`
	DurationMinutes float64 `json:"duration_minutes"`
}

// AnalyzeRequest for comprehensive analysis
type AnalyzeRequest struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Ammonia     float64 `json:"ammonia"`
	RBWID       string  `json:"rbw_id,omitempty"`
	NodeID      string  `json:"node_id,omitempty"`
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

// AnalyzeResponse from comprehensive analysis
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

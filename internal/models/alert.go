package models

import "time"

// Alert represents a system alert
type Alert struct {
	ID         string     `json:"id"`
	RBWID      string     `json:"rbw_id"`
	NodeID     *string    `json:"node_id,omitempty"`
	SensorID   *string    `json:"sensor_id,omitempty"`
	AlertType  string     `json:"alert_type"`
	Severity   int        `json:"severity"`
	Message    *string    `json:"message,omitempty"`
	IsRead     bool       `json:"is_read"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy *string    `json:"resolved_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// AlertType constants
const (
	AlertTypeTempHigh    = "temp_high"
	AlertTypeTempLow     = "temp_low"
	AlertTypeHumidHigh   = "humid_high"
	AlertTypeHumidLow    = "humid_low"
	AlertTypeAmmoniaHigh = "ammonia_high"
	AlertTypeNodeOffline = "node_offline"
	AlertTypeAIAnomaly   = "ai_anomaly"
)

// CreateAlertRequest for creating an alert
type CreateAlertRequest struct {
	RBWID     string  `json:"rbw_id"`
	NodeID    *string `json:"node_id,omitempty"`
	SensorID  *string `json:"sensor_id,omitempty"`
	AlertType string  `json:"alert_type"`
	Severity  int     `json:"severity"`
	Message   string  `json:"message,omitempty"`
}

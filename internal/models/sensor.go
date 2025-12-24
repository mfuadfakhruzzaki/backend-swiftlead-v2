package models

import "time"

// Sensor represents a sensor attached to a node
type Sensor struct {
	ID         string    `json:"id"`
	NodeID     string    `json:"node_id"`
	SensorType string    `json:"sensor_type"`
	SensorName *string   `json:"sensor_name,omitempty"`
	Unit       *string   `json:"unit,omitempty"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SensorType constants
const (
	SensorTypeTemp    = "temp"
	SensorTypeHumid   = "humid"
	SensorTypeAmmonia = "ammonia"
)

// CreateSensorRequest for creating a new sensor
type CreateSensorRequest struct {
	SensorType string `json:"sensor_type"`
	SensorName string `json:"sensor_name,omitempty"`
	Unit       string `json:"unit,omitempty"`
}

// UpdateSensorRequest for updating a sensor
type UpdateSensorRequest struct {
	SensorName *string `json:"sensor_name,omitempty"`
	Unit       *string `json:"unit,omitempty"`
	IsActive   *bool   `json:"is_active,omitempty"`
}

// SensorReading represents a sensor reading
type SensorReading struct {
	ID         int64     `json:"id"`
	SensorID   string    `json:"sensor_id"`
	RecordedAt time.Time `json:"recorded_at"`
	Value      float64   `json:"value"`
	IsAnomaly  bool      `json:"is_anomaly"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateSensorReadingRequest for ingesting sensor data
type CreateSensorReadingRequest struct {
	Value      float64   `json:"value"`
	RecordedAt time.Time `json:"recorded_at,omitempty"`
}

// SensorPayload from MQTT
type SensorPayload struct {
	ESP32UID  string  `json:"esp32_uid"`
	Temp      float64 `json:"temp"`
	Humidity  float64 `json:"rh"`
	Ammonia   float64 `json:"nh3"`
	RSSI      int     `json:"rssi"`
	Timestamp int64   `json:"timestamp"`
	Seq       int     `json:"seq"`
}

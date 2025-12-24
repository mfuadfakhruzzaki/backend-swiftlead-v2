package models

import "time"

// Node represents an IoT device (ESP32)
type Node struct {
	ID             string     `json:"id"`
	RBWID          string     `json:"rbw_id"`
	NodeType       string     `json:"node_type"`
	NodeCode       string     `json:"node_code"`
	ESP32UID       *string    `json:"esp32_uid,omitempty"`
	StatusNode     string     `json:"status_node"`
	LastSeen       *time.Time `json:"last_seen,omitempty"`
	HasAudio       bool       `json:"has_audio"`
	StateAudioLMB  *bool      `json:"state_audio_lmb,omitempty"`
	StateAudioNest *bool      `json:"state_audio_nest,omitempty"`
	HasPump        bool       `json:"has_pump"`
	StatePump      *bool      `json:"state_pump,omitempty"`
	InstalledAt    *time.Time `json:"installed_at,omitempty"`
	UninstalledAt  *time.Time `json:"uninstalled_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// NodeType constants
const (
	NodeTypeGateway = "gateway"
	NodeTypeNest    = "nest"
	NodeTypeLMB     = "lmb"
	NodeTypePump    = "pump"
)

// NodeStatus constants
const (
	NodeStatusOnline  = "online"
	NodeStatusOffline = "offline"
	NodeStatusError   = "error"
)

// CreateNodeRequest for creating a new node
type CreateNodeRequest struct {
	NodeType string  `json:"node_type"`
	NodeCode string  `json:"node_code"`
	ESP32UID *string `json:"esp32_uid,omitempty"`
	HasAudio bool    `json:"has_audio"`
	HasPump  bool    `json:"has_pump"`
}

// UpdateNodeRequest for updating a node
type UpdateNodeRequest struct {
	NodeCode *string `json:"node_code,omitempty"`
	ESP32UID *string `json:"esp32_uid,omitempty"`
	HasAudio *bool   `json:"has_audio,omitempty"`
	HasPump  *bool   `json:"has_pump,omitempty"`
}

// AudioControlRequest for controlling audio
type AudioControlRequest struct {
	Action string `json:"action"` // audio_set_lmb, audio_set_nest, call_bird
	Value  int    `json:"value"`  // 0 or 1
}

// PumpControlRequest for controlling pump
type PumpControlRequest struct {
	Action string `json:"action"` // sprayer_set
	Value  int    `json:"value"`  // 0 or 1
}

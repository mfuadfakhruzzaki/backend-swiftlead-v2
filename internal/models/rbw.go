package models

import "time"

// RBW represents a Rumah Burung Walet (Swiftlet House)
type RBW struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Address     *string   `json:"address,omitempty"`
	Latitude    *float64  `json:"latitude,omitempty"`
	Longitude   *float64  `json:"longitude,omitempty"`
	TotalFloors int       `json:"total_floors"`
	Description *string   `json:"description,omitempty"`
	PhotoURL    *string   `json:"photo_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateRBWRequest for creating a new RBW
type CreateRBWRequest struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Address     string   `json:"address,omitempty"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	TotalFloors int      `json:"total_floors"`
	Description string   `json:"description,omitempty"`
}

// UpdateRBWRequest for updating an RBW
type UpdateRBWRequest struct {
	Name        *string  `json:"name,omitempty"`
	Address     *string  `json:"address,omitempty"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	TotalFloors *int     `json:"total_floors,omitempty"`
	Description *string  `json:"description,omitempty"`
	PhotoURL    *string  `json:"photo_url,omitempty"`
}

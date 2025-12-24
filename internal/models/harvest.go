package models

import "time"

// Harvest represents a harvest record
type Harvest struct {
	ID          string    `json:"id"`
	RBWID       string    `json:"rbw_id"`
	NodeID      *string   `json:"node_id,omitempty"`
	FloorNo     int       `json:"floor_no"`
	HarvestedAt time.Time `json:"harvested_at"`
	NestsCount  int       `json:"nests_count"`
	WeightKg    *float64  `json:"weight_kg,omitempty"`
	Grade       *string   `json:"grade,omitempty"`
	Notes       *string   `json:"notes,omitempty"`
	CreatedBy   *string   `json:"created_by,omitempty"`
	CycleDays   *int      `json:"cycle_days,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// HarvestGrade constants
const (
	HarvestGradeGood   = "good"
	HarvestGradeMedium = "medium"
	HarvestGradePoor   = "poor"
)

// CreateHarvestRequest for creating a harvest
type CreateHarvestRequest struct {
	RBWID       string    `json:"rbw_id"`
	NodeID      *string   `json:"node_id,omitempty"`
	FloorNo     int       `json:"floor_no"`
	HarvestedAt time.Time `json:"harvested_at"`
	NestsCount  int       `json:"nests_count"`
	WeightKg    *float64  `json:"weight_kg,omitempty"`
	Grade       *string   `json:"grade,omitempty"`
	Notes       string    `json:"notes,omitempty"`
}

// UpdateHarvestRequest for updating a harvest
type UpdateHarvestRequest struct {
	FloorNo    *int     `json:"floor_no,omitempty"`
	NestsCount *int     `json:"nests_count,omitempty"`
	WeightKg   *float64 `json:"weight_kg,omitempty"`
	Grade      *string  `json:"grade,omitempty"`
	Notes      *string  `json:"notes,omitempty"`
}

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
	RBWID       string    `json:"rbw_id" validate:"required"`
	NodeID      *string   `json:"node_id,omitempty"`
	FloorNo     int       `json:"floor_no" validate:"required,min=1"`
	HarvestedAt time.Time `json:"harvested_at" validate:"required"`
	NestsCount  int       `json:"nests_count" validate:"min=0"`
	WeightKg    *float64  `json:"weight_kg,omitempty"`
	Grade       *string   `json:"grade,omitempty" validate:"omitempty,oneof=good medium poor"`
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

// HarvestStats represents aggregate statistics for harvests
type HarvestStats struct {
	TotalHarvests      int     `json:"total_harvests"`
	TotalNests         int     `json:"total_nests"`
	TotalWeightKg      float64 `json:"total_weight_kg"`
	AvgNestsPerHarvest float64 `json:"avg_nests_per_harvest"`
	AvgWeightKg        float64 `json:"avg_weight_kg"`
	AvgCycleDays       float64 `json:"avg_cycle_days"`
}

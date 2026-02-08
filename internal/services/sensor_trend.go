package services

import (
	"context"
	"math"
	"time"

	"github.com/swiftlead/backend-swiftlet/internal/repository"
)

// TrendDirection represents the direction of a sensor trend
type TrendDirection string

const (
	TrendRising  TrendDirection = "rising"
	TrendFalling TrendDirection = "falling"
	TrendStable  TrendDirection = "stable"
)

// TrendResult represents the result of a trend analysis
type TrendResult struct {
	SensorID   string         `json:"sensor_id"`
	SensorType string         `json:"sensor_type"`
	Direction  TrendDirection `json:"direction"`
	Slope      float64        `json:"slope"`
	AvgValue   float64        `json:"avg_value"`
	MinValue   float64        `json:"min_value"`
	MaxValue   float64        `json:"max_value"`
	DataPoints int            `json:"data_points"`
	Period     string         `json:"period"`
}

// SensorTrendCalculator calculates sensor data trends
type SensorTrendCalculator struct {
	telemetryRepo repository.TelemetryRepository
}

// NewSensorTrendCalculator creates a new sensor trend calculator
func NewSensorTrendCalculator(telemetryRepo repository.TelemetryRepository) *SensorTrendCalculator {
	return &SensorTrendCalculator{
		telemetryRepo: telemetryRepo,
	}
}

// CalculateTrend calculates the trend for a sensor over a time period
func (c *SensorTrendCalculator) CalculateTrend(ctx context.Context, sensorID, sensorType string, duration time.Duration) (*TrendResult, error) {
	from := time.Now().Add(-duration)
	to := time.Now()

	readings, err := c.telemetryRepo.GetReadings(ctx, sensorID, from, to, 1000)
	if err != nil {
		return nil, err
	}

	if len(readings) < 2 {
		return &TrendResult{
			SensorID:   sensorID,
			SensorType: sensorType,
			Direction:  TrendStable,
			DataPoints: len(readings),
			Period:     duration.String(),
		}, nil
	}

	// Calculate stats
	var sum, minVal, maxVal float64
	minVal = math.MaxFloat64
	maxVal = -math.MaxFloat64

	xs := make([]float64, len(readings))
	ys := make([]float64, len(readings))

	for i, r := range readings {
		val := r.Value
		sum += val
		if val < minVal {
			minVal = val
		}
		if val > maxVal {
			maxVal = val
		}
		xs[i] = float64(i)
		ys[i] = val
	}

	avgValue := sum / float64(len(readings))

	// Simple linear regression to get slope
	slope := linearRegressionSlope(xs, ys)

	// Determine direction based on slope threshold
	threshold := 0.01 * avgValue // 1% of average as threshold
	if threshold < 0.001 {
		threshold = 0.001
	}

	var direction TrendDirection
	if slope > threshold {
		direction = TrendRising
	} else if slope < -threshold {
		direction = TrendFalling
	} else {
		direction = TrendStable
	}

	return &TrendResult{
		SensorID:   sensorID,
		SensorType: sensorType,
		Direction:  direction,
		Slope:      math.Round(slope*1000) / 1000,
		AvgValue:   math.Round(avgValue*100) / 100,
		MinValue:   math.Round(minVal*100) / 100,
		MaxValue:   math.Round(maxVal*100) / 100,
		DataPoints: len(readings),
		Period:     duration.String(),
	}, nil
}

// linearRegressionSlope computes the slope of a simple linear regression
func linearRegressionSlope(xs, ys []float64) float64 {
	n := float64(len(xs))
	if n < 2 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i := 0; i < len(xs); i++ {
		sumX += xs[i]
		sumY += ys[i]
		sumXY += xs[i] * ys[i]
		sumX2 += xs[i] * xs[i]
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0
	}

	return (n*sumXY - sumX*sumY) / denominator
}

package services

import (
	"context"
	"time"

	"github.com/swiftlead/backend-swiftlet/internal/config"
	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
)

// TelemetryService handles sensor data ingestion
type TelemetryService struct {
	nodeRepo      repository.NodeRepository
	sensorRepo    repository.SensorRepository
	telemetryRepo repository.TelemetryRepository
	alertRepo     repository.AlertRepository
	cfg           *config.Config
}

// NewTelemetryService creates a new telemetry service
func NewTelemetryService(
	nodeRepo repository.NodeRepository,
	sensorRepo repository.SensorRepository,
	telemetryRepo repository.TelemetryRepository,
	alertRepo repository.AlertRepository,
	cfg *config.Config,
) *TelemetryService {
	return &TelemetryService{
		nodeRepo:      nodeRepo,
		sensorRepo:    sensorRepo,
		telemetryRepo: telemetryRepo,
		alertRepo:     alertRepo,
		cfg:           cfg,
	}
}

// ProcessSensorPayload processes incoming MQTT sensor data
func (s *TelemetryService) ProcessSensorPayload(ctx context.Context, payload *models.SensorPayload) error {
	// Find node by ESP32 UID
	node, err := s.nodeRepo.GetByESP32UID(ctx, payload.ESP32UID)
	if err != nil {
		return err
	}

	// Update node last_seen
	if err := s.nodeRepo.UpdateLastSeen(ctx, node.ID); err != nil {
		return err
	}

	// Get sensors for this node
	sensors, err := s.sensorRepo.ListByNode(ctx, node.ID)
	if err != nil {
		return err
	}

	// Process each sensor reading
	recordedAt := time.Unix(payload.Timestamp, 0)
	if payload.Timestamp == 0 {
		recordedAt = time.Now()
	}

	for _, sensor := range sensors {
		var value float64
		var alertType string

		switch sensor.SensorType {
		case models.SensorTypeTemp:
			value = payload.Temp
			if value > s.cfg.TempHighThreshold {
				alertType = models.AlertTypeTempHigh
			} else if value < s.cfg.TempLowThreshold {
				alertType = models.AlertTypeTempLow
			}
		case models.SensorTypeHumid:
			value = payload.Humidity
			if value > s.cfg.HumidHighThreshold {
				alertType = models.AlertTypeHumidHigh
			} else if value < s.cfg.HumidLowThreshold {
				alertType = models.AlertTypeHumidLow
			}
		case models.SensorTypeAmmonia:
			value = payload.Ammonia
			if value > s.cfg.AmmoniaHighThreshold {
				alertType = models.AlertTypeAmmoniaHigh
			}
		default:
			continue
		}

		// Create reading
		reading := &models.SensorReading{
			SensorID:   sensor.ID,
			RecordedAt: recordedAt,
			Value:      value,
			IsAnomaly:  alertType != "",
		}
		if err := s.telemetryRepo.CreateReading(ctx, reading); err != nil {
			return err
		}

		// Create alert if threshold exceeded
		if alertType != "" {
			alert := &models.Alert{
				RBWID:     node.RBWID,
				NodeID:    &node.ID,
				SensorID:  &sensor.ID,
				AlertType: alertType,
				Severity:  s.getSeverity(alertType, value),
				Message:   s.getAlertMessage(alertType, value),
			}
			if err := s.alertRepo.Create(ctx, alert); err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateReading creates a manual sensor reading
func (s *TelemetryService) CreateReading(ctx context.Context, sensorID string, req *models.CreateSensorReadingRequest) (*models.SensorReading, error) {
	reading := &models.SensorReading{
		SensorID:   sensorID,
		RecordedAt: req.RecordedAt,
		Value:      req.Value,
	}
	if reading.RecordedAt.IsZero() {
		reading.RecordedAt = time.Now()
	}

	if err := s.telemetryRepo.CreateReading(ctx, reading); err != nil {
		return nil, err
	}
	return reading, nil
}

// GetReadings retrieves sensor readings with time range
func (s *TelemetryService) GetReadings(ctx context.Context, sensorID string, from, to time.Time, limit int) ([]*models.SensorReading, error) {
	if from.IsZero() {
		from = time.Now().Add(-24 * time.Hour)
	}
	if to.IsZero() {
		to = time.Now()
	}
	if limit == 0 {
		limit = 100
	}
	return s.telemetryRepo.GetReadings(ctx, sensorID, from, to, limit)
}

func (s *TelemetryService) getSeverity(alertType string, value float64) int {
	switch alertType {
	case models.AlertTypeAmmoniaHigh:
		if value > 35 {
			return 5
		}
		return 4
	case models.AlertTypeTempHigh, models.AlertTypeTempLow:
		return 3
	default:
		return 2
	}
}

func (s *TelemetryService) getAlertMessage(alertType string, value float64) *string {
	var msg string
	switch alertType {
	case models.AlertTypeTempHigh:
		msg = "Temperature too high"
	case models.AlertTypeTempLow:
		msg = "Temperature too low"
	case models.AlertTypeHumidHigh:
		msg = "Humidity too high"
	case models.AlertTypeHumidLow:
		msg = "Humidity too low"
	case models.AlertTypeAmmoniaHigh:
		msg = "Ammonia level high, improve ventilation"
	}
	return &msg
}

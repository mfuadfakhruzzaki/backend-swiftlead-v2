package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/swiftlead/backend-swiftlet/internal/ai"
	"github.com/swiftlead/backend-swiftlet/internal/config"
	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
	"github.com/swiftlead/backend-swiftlet/internal/websocket"
	"github.com/swiftlead/backend-swiftlet/pkg/logger"
)

// TelemetryService handles sensor data ingestion
type TelemetryService struct {
	nodeRepo      repository.NodeRepository
	sensorRepo    repository.SensorRepository
	telemetryRepo repository.TelemetryRepository
	alertRepo     repository.AlertRepository
	aiClient      *ai.Client
	cfg           *config.Config
	wsHub         *websocket.Hub
	audioSvc      *AudioService
}

// NewTelemetryService creates a new telemetry service.
// audioSvc may be nil; when set, AI pump decisions are automatically actuated
// and an auto-off timer is scheduled per cfg.PumpAutoOffSeconds.
func NewTelemetryService(
	nodeRepo repository.NodeRepository,
	sensorRepo repository.SensorRepository,
	telemetryRepo repository.TelemetryRepository,
	alertRepo repository.AlertRepository,
	aiClient *ai.Client,
	cfg *config.Config,
	wsHub *websocket.Hub,
	audioSvc *AudioService,
) *TelemetryService {
	return &TelemetryService{
		nodeRepo:      nodeRepo,
		sensorRepo:    sensorRepo,
		telemetryRepo: telemetryRepo,
		alertRepo:     alertRepo,
		aiClient:      aiClient,
		cfg:           cfg,
		wsHub:         wsHub,
		audioSvc:      audioSvc,
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

	// If this node is not a gateway, also mark the gateway of the same RBW as online
	// because data from nest/lmb/pump nodes passes through the gateway (ESP-NOW → MQTT)
	if node.NodeType != models.NodeTypeGateway {
		gateway, err := s.nodeRepo.GetGatewayByRBW(ctx, node.RBWID)
		if err != nil {
			logger.Error("Failed to get gateway for RBW %s: %v", node.RBWID, err)
		} else if gateway != nil {
			if err := s.nodeRepo.UpdateLastSeen(ctx, gateway.ID); err != nil {
				logger.Error("Failed to update gateway last_seen: %v", err)
			}
		}
	}

	// If this node IS a gateway, also mark pump and LMB nodes in the same RBW as online
	// because pump/LMB nodes communicate via ESP-NOW through the gateway (they don't send MQTT data directly)
	if node.NodeType == models.NodeTypeGateway {
		espNowTypes := []string{models.NodeTypePump, models.NodeTypeLMB}
		if err := s.nodeRepo.UpdateLastSeenByRBWAndTypes(ctx, node.RBWID, espNowTypes); err != nil {
			logger.Error("Failed to update pump/LMB last_seen for RBW %s: %v", node.RBWID, err)
		}
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

		// Extract value per sensor type — no hardcoded thresholds, AI decides what's anomalous
		switch sensor.SensorType {
		case models.SensorTypeTemp:
			value = payload.Temp
		case models.SensorTypeHumid:
			value = payload.Humidity
		case models.SensorTypeAmmonia:
			value = payload.Ammonia
		default:
			continue
		}

		// AI anomaly detection is the sole alert trigger
		isAnomaly := false
		if s.aiClient != nil && s.aiClient.IsEnabled() {
			anomalyReq := &ai.AnomalyRequest{
				SensorID:   sensor.ID,
				SensorType: sensor.SensorType,
				RBWID:      node.RBWID,
				NodeID:     node.ID,
				RecordedAt: recordedAt,
				Value:      value,
			}
			anomalyResp, err := s.aiClient.DetectAnomaly(ctx, anomalyReq)
			if err == nil && anomalyResp.IsAnomaly {
				isAnomaly = true
				alertType = models.AlertTypeAIAnomaly
				logger.Info("AI anomaly detected: sensor=%s value=%.2f score=%.2f", sensor.ID, value, anomalyResp.Score)
			}
		}

		// Create reading
		reading := &models.SensorReading{
			SensorID:   sensor.ID,
			RecordedAt: recordedAt,
			Value:      value,
			IsAnomaly:  isAnomaly,
		}
		if err := s.telemetryRepo.CreateReading(ctx, reading); err != nil {
			return err
		}

		// Broadcast to WebSocket clients
		if s.wsHub != nil {
			sensorData, _ := json.Marshal(map[string]interface{}{
				"sensor_id":   sensor.ID,
				"node_id":     node.ID,
				"rbw_id":      node.RBWID,
				"sensor_type": sensor.SensorType,
				"value":       value,
				"unit":        sensor.Unit,
				"is_anomaly":  isAnomaly,
				"recorded_at": recordedAt.Format(time.RFC3339),
			})
			s.wsHub.BroadcastAll(websocket.Message{
				Type:      "sensor_reading",
				Data:      sensorData,
				Timestamp: time.Now(),
			})
		}

		// Create alert if threshold exceeded or AI anomaly detected (with deduplication)
		if alertType != "" {
			exists, _ := s.alertRepo.HasUnresolved(ctx, alertType, &node.ID, &sensor.ID)
			if !exists {
				severity := s.getSeverity(alertType, value)
				msg := s.getAlertMessage(alertType, value)
				alert := &models.Alert{
					RBWID:     node.RBWID,
					NodeID:    &node.ID,
					SensorID:  &sensor.ID,
					AlertType: alertType,
					Severity:  severity,
					Message:   msg,
				}
				if err := s.alertRepo.Create(ctx, alert); err != nil {
					return err
				}
				s.broadcastAlert(alert)
			}
		}
	}

	// AI: push to buffer + comprehensive multivariate decision.
	// Runs in a goroutine so MQTT processing is never blocked by AI latency.
	if s.aiClient != nil && s.aiClient.IsEnabled() {
		nodeID := node.ID
		rbwID := node.RBWID
		temp := payload.Temp
		humid := payload.Humidity
		ammonia := payload.Ammonia
		ts := float64(recordedAt.Unix())
		hub := s.wsHub

		go func() {
			aiCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// 1. Push to rolling buffer (enables 1-hour rolling features for /v2/decide)
			_ = s.aiClient.PushReading(aiCtx, nodeID, temp, humid, ammonia, &ts)

			// 2. Multivariate decision: grade + pump + anomaly sekaligus
			decision, err := s.aiClient.DecideRealtime(aiCtx, nodeID, temp, humid, ammonia)
			if err != nil || decision == nil {
				return
			}

			// 3. Broadcast hasil AI ke WebSocket clients
			if hub != nil {
				decisionData, _ := json.Marshal(map[string]interface{}{
					"node_id":        nodeID,
					"rbw_id":         rbwID,
					"grade":          decision.Grade,
					"sprayer_on":     decision.SprayerOn,
					"sprayer_reason": decision.SprayerReason,
					"anomaly":        decision.AnomalyVerdict,
					"confidence":     decision.Confidence,
					"recorded_at":    recordedAt.Format(time.RFC3339),
				})
				hub.BroadcastAll(websocket.Message{
					Type:      "ai_decision",
					Data:      decisionData,
					Timestamp: time.Now(),
				})
			}

			// 4. Auto-actuate pump based on AI recommendation.
			// Skipped entirely when pump_auto_mode = false (manual override active).
			// Only issues an MQTT command when the desired state differs from the
			// current state to avoid redundant commands.
			//
			// Race-condition note: GetPumpNodesByRBW is called after DecideRealtime
			// which can take several seconds. A manual override (pump_auto_mode=false)
			// issued during that window would not be visible in pumpNode. To close
			// this race, each node is re-read individually with GetByID right before
			// actuation — the window is now effectively zero (two consecutive statements).
			if s.audioSvc != nil {
				pumpNodes, pErr := s.nodeRepo.GetPumpNodesByRBW(aiCtx, rbwID)
				if pErr != nil {
					logger.Warn("AI auto-actuate: failed to get pump nodes for RBW %s: %v", rbwID, pErr)
				} else {
					for _, pumpNode := range pumpNodes {
						// First pass: skip obvious non-candidates without an extra DB round trip
						if !pumpNode.PumpAutoMode {
							logger.Debug("AI auto-actuate skipped: node=%s is in manual mode", pumpNode.ID)
							continue
						}

						// Fresh read — eliminates the race between GetPumpNodesByRBW
						// (pre-DecideRealtime) and a manual override that arrived during AI latency.
						fresh, err := s.nodeRepo.GetByID(aiCtx, pumpNode.ID)
						if err != nil {
							logger.Warn("AI auto-actuate: fresh read failed: node=%s err=%v", pumpNode.ID, err)
							continue
						}
						if !fresh.PumpAutoMode {
							logger.Debug("AI auto-actuate skipped (fresh): node=%s manual override active", pumpNode.ID)
							continue
						}

						currentlyOn := fresh.StatePump != nil && *fresh.StatePump
						if decision.SprayerOn == currentlyOn {
							continue // no state change needed
						}
						value := 0
						if decision.SprayerOn {
							value = 1
						}
						req := &models.PumpControlRequest{Action: "sprayer_set", Value: value}
						if err := s.audioSvc.ControlPumpAI(aiCtx, pumpNode.ID, req); err != nil {
							logger.Error("AI auto-actuate pump failed: node=%s err=%v", pumpNode.ID, err)
							continue
						}
						logger.Info("AI auto-actuated pump: node=%s sprayer_on=%v reason=%q",
							pumpNode.ID, decision.SprayerOn, decision.SprayerReason)

						// Schedule auto-off when turning on so the pump doesn't run indefinitely.
						// Only schedule if AI mode is active (fresh.PumpAutoMode == true),
						// meaning ControlPumpAI actually actuated. When manual mode is active,
						// ControlPumpAI returns nil (no-op) and we must NOT touch the scheduler
						// to avoid overwriting the user's manual timer.
						if fresh.PumpAutoMode && decision.SprayerOn && s.cfg.PumpAutoOffSeconds > 0 {
							duration := time.Duration(s.cfg.PumpAutoOffSeconds * float64(time.Second))
							s.audioSvc.ScheduleAutoOff(pumpNode.ID, duration)
						}
					}
				}
			}

			// 6. Node-level alert from multivariate anomaly verdict
			if decision.AnomalyVerdict == "anomaly" {
				nid := nodeID
				exists, _ := s.alertRepo.HasUnresolved(aiCtx, models.AlertTypeAIAnomaly, &nid, nil)
				if !exists {
					msg := "AI mendeteksi anomali lingkungan secara multivariate — kondisi RBW tidak normal"
					alert := &models.Alert{
						RBWID:     rbwID,
						NodeID:    &nid,
						AlertType: models.AlertTypeAIAnomaly,
						Severity:  3,
						Message:   &msg,
					}
					if err := s.alertRepo.Create(aiCtx, alert); err == nil {
						s.broadcastAlert(alert)
					}
				}
			}
		}()
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

func (s *TelemetryService) getSeverity(alertType string, _ float64) int {
	switch alertType {
	case models.AlertTypeAIAnomaly:
		return 3
	default:
		return 2
	}
}

func (s *TelemetryService) getAlertMessage(alertType string, _ float64) *string {
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
	case models.AlertTypeAIAnomaly:
		msg = "AI detected anomalous sensor reading"
	}
	return &msg
}

// GetLatestReading retrieves the latest reading for a sensor
func (s *TelemetryService) GetLatestReading(ctx context.Context, sensorID string) (*models.SensorReading, error) {
	return s.telemetryRepo.GetLatestReading(ctx, sensorID)
}

func (s *TelemetryService) broadcastAlert(alert *models.Alert) {
	if s.wsHub == nil {
		return
	}
	data, _ := json.Marshal(map[string]interface{}{
		"id":         alert.ID,
		"rbw_id":     alert.RBWID,
		"node_id":    alert.NodeID,
		"sensor_id":  alert.SensorID,
		"alert_type": alert.AlertType,
		"severity":   alert.Severity,
		"message":    alert.Message,
		"created_at": alert.CreatedAt.Format(time.RFC3339),
	})
	s.wsHub.BroadcastAll(websocket.Message{
		Type:      "new_alert",
		Data:      data,
		Timestamp: time.Now(),
	})
}

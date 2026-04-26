package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/mqtt"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
	"github.com/swiftlead/backend-swiftlet/pkg/logger"
)

// pumpScheduler tracks per-node auto-off cancel functions.
// When the pump is turned on with a duration, a goroutine sleeps then issues
// a turn-off command. Calling cancelAutoOff before it fires cancels the command.
type pumpScheduler struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func (s *pumpScheduler) schedule(nodeID string, duration time.Duration, mqttClient *mqtt.Client, nodeRepo repository.NodeRepository) {
	s.mu.Lock()
	if cancel, ok := s.cancels[nodeID]; ok {
		cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancels[nodeID] = cancel
	s.mu.Unlock()

	go func() {
		select {
		case <-time.After(duration):
		case <-ctx.Done():
			return
		}

		s.mu.Lock()
		delete(s.cancels, nodeID)
		s.mu.Unlock()

		if err := mqttClient.PublishPumpCommand(0); err != nil {
			logger.Error("Pump auto-off MQTT publish failed for node %s: %v", nodeID, err)
		}
		offCtx, offCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer offCancel()
		if err := nodeRepo.UpdatePumpState(offCtx, nodeID, false); err != nil {
			logger.Warn("Pump auto-off DB update failed for node %s: %v", nodeID, err)
		}
		logger.Info("Pump auto-off fired: node=%s", nodeID)
	}()
}

func (s *pumpScheduler) cancel(nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cancel, ok := s.cancels[nodeID]; ok {
		cancel()
		delete(s.cancels, nodeID)
	}
}

// AudioService handles audio and pump control for nodes.
type AudioService struct {
	mqtt      *mqtt.Client
	nodeRepo  repository.NodeRepository
	scheduler pumpScheduler
}

// NewAudioService creates a new AudioService.
func NewAudioService(mqttClient *mqtt.Client, nodeRepo repository.NodeRepository) *AudioService {
	return &AudioService{
		mqtt:     mqttClient,
		nodeRepo: nodeRepo,
		scheduler: pumpScheduler{
			cancels: make(map[string]context.CancelFunc),
		},
	}
}

// ControlAudio sends an audio command to a node via MQTT.
func (s *AudioService) ControlAudio(ctx context.Context, nodeID string, req *models.AudioControlRequest) error {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return err
	}
	if !node.HasAudio {
		return fmt.Errorf("node %s does not have audio capability", nodeID)
	}
	switch req.Action {
	case "audio_set_lmb", "audio_set_nest", "call_bird":
	default:
		return fmt.Errorf("invalid audio action: %s", req.Action)
	}
	if req.Value != 0 && req.Value != 1 {
		return fmt.Errorf("value must be 0 or 1")
	}
	if err := s.mqtt.PublishAudioCommand(req.Action, req.Value); err != nil {
		logger.Error("Failed to publish audio command: %v", err)
		return fmt.Errorf("failed to send audio command: %w", err)
	}
	state := req.Value == 1
	switch req.Action {
	case "audio_set_lmb":
		if err := s.nodeRepo.UpdateAudioState(ctx, nodeID, &state, nil); err != nil {
			logger.Warn("Failed to update audio LMB state: %v", err)
		}
	case "audio_set_nest":
		if err := s.nodeRepo.UpdateAudioState(ctx, nodeID, nil, &state); err != nil {
			logger.Warn("Failed to update audio nest state: %v", err)
		}
	case "call_bird":
		if err := s.nodeRepo.UpdateAudioState(ctx, nodeID, &state, &state); err != nil {
			logger.Warn("Failed to update audio state: %v", err)
		}
	}
	logger.Info("Audio command sent: node=%s action=%s value=%d", nodeID, req.Action, req.Value)
	return nil
}

// GetAudioState returns the current audio state for a node.
func (s *AudioService) GetAudioState(ctx context.Context, nodeID string) (*models.Node, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	if !node.HasAudio {
		return nil, fmt.Errorf("node %s does not have audio capability", nodeID)
	}
	return node, nil
}

// ControlPump sends a pump command to a node via MQTT and updates DB state.
// It cancels any pending auto-off timer before sending.
// If req.DurationSeconds > 0 and req.Value == 1, a new auto-off timer is scheduled.
//
// This method does NOT modify pump_auto_mode. Callers that want to put the
// node into manual mode must call SetPumpAutoMode(false) separately.
func (s *AudioService) ControlPump(ctx context.Context, nodeID string, req *models.PumpControlRequest) error {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return err
	}
	if !node.HasPump {
		return fmt.Errorf("node %s does not have pump capability", nodeID)
	}
	if req.Value != 0 && req.Value != 1 {
		return fmt.Errorf("value must be 0 or 1")
	}

	// Cancel any existing auto-off timer before issuing the new command.
	s.scheduler.cancel(nodeID)

	if err := s.mqtt.PublishPumpCommand(req.Value); err != nil {
		logger.Error("Failed to publish pump command: %v", err)
		return fmt.Errorf("failed to send pump command: %w", err)
	}

	state := req.Value == 1
	if err := s.nodeRepo.UpdatePumpState(ctx, nodeID, state); err != nil {
		logger.Warn("Failed to update pump state: %v", err)
	}

	// Schedule auto-off timer if a duration was requested and we're turning on.
	if state && req.DurationSeconds != nil && *req.DurationSeconds > 0 {
		duration := time.Duration(*req.DurationSeconds * float64(time.Second))
		s.scheduler.schedule(nodeID, duration, s.mqtt, s.nodeRepo)
		logger.Info("Pump manual timer set: node=%s duration=%.0fs", nodeID, *req.DurationSeconds)
	}

	logger.Info("Pump command sent: node=%s value=%d", nodeID, req.Value)
	return nil
}

// SetPumpAutoMode updates the pump_auto_mode flag for a node.
// Call with false to suspend AI automation; call with true to re-enable it.
func (s *AudioService) SetPumpAutoMode(ctx context.Context, nodeID string, autoMode bool) error {
	if err := s.nodeRepo.UpdatePumpAutoMode(ctx, nodeID, autoMode); err != nil {
		return fmt.Errorf("failed to update pump auto mode: %w", err)
	}
	logger.Info("Pump auto mode set: node=%s auto_mode=%v", nodeID, autoMode)
	return nil
}

// ScheduleAutoOff arranges for the pump on nodeID to be turned off after duration.
// Any previously scheduled auto-off for the same node is cancelled first.
// Calling with duration ≤ 0 is a no-op.
func (s *AudioService) ScheduleAutoOff(nodeID string, duration time.Duration) {
	if duration <= 0 {
		return
	}
	logger.Info("Pump auto-off scheduled: node=%s duration=%s", nodeID, duration)
	s.scheduler.schedule(nodeID, duration, s.mqtt, s.nodeRepo)
}

// SyncAllPumpStates re-publishes the desired pump state for every pump-capable node.
// Called after MQTT reconnect to bring hardware that restarted back in sync with DB.
// NOTE: uses the broadcast topic — in multi-RBW deployments all gateways receive
// every message. Per-gateway isolation requires a firmware subscription update.
func (s *AudioService) SyncAllPumpStates() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	nodes, err := s.nodeRepo.ListAllWithPump(ctx)
	if err != nil {
		logger.Error("Pump state sync failed — could not list pump nodes: %v", err)
		return
	}

	synced := 0
	for _, node := range nodes {
		if node.ESP32UID == nil || strings.TrimSpace(*node.ESP32UID) == "" {
			continue
		}
		value := 0
		if node.StatePump != nil && *node.StatePump {
			value = 1
		}
		if err := s.mqtt.PublishPumpCommand(value); err != nil {
			logger.Warn("Pump state sync publish failed for node %s: %v", node.ID, err)
			continue
		}
		synced++
	}
	logger.Info("Pump state sync complete: %d/%d nodes synced", synced, len(nodes))
}

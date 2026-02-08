package services

import (
	"context"
	"fmt"

	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/mqtt"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
	"github.com/swiftlead/backend-swiftlet/pkg/logger"
)

// AudioService handles audio control for nodes
type AudioService struct {
	mqtt     *mqtt.Client
	nodeRepo repository.NodeRepository
}

// NewAudioService creates a new audio service
func NewAudioService(mqttClient *mqtt.Client, nodeRepo repository.NodeRepository) *AudioService {
	return &AudioService{
		mqtt:     mqttClient,
		nodeRepo: nodeRepo,
	}
}

// ControlAudio sends an audio command to a node via MQTT
func (s *AudioService) ControlAudio(ctx context.Context, nodeID string, req *models.AudioControlRequest) error {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return err
	}

	if !node.HasAudio {
		return fmt.Errorf("node %s does not have audio capability", nodeID)
	}

	// Validate action
	switch req.Action {
	case "audio_set_lmb", "audio_set_nest", "call_bird":
		// valid
	default:
		return fmt.Errorf("invalid audio action: %s", req.Action)
	}

	// Validate value
	if req.Value != 0 && req.Value != 1 {
		return fmt.Errorf("value must be 0 or 1")
	}

	// Publish MQTT command
	if err := s.mqtt.PublishAudioCommand(req.Action, req.Value); err != nil {
		logger.Error("Failed to publish audio command: %v", err)
		return fmt.Errorf("failed to send audio command: %w", err)
	}

	// Update node state in DB
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
		// call_bird is a momentary action, update both states
		if err := s.nodeRepo.UpdateAudioState(ctx, nodeID, &state, &state); err != nil {
			logger.Warn("Failed to update audio state: %v", err)
		}
	}

	logger.Info("Audio command sent: node=%s action=%s value=%d", nodeID, req.Action, req.Value)
	return nil
}

// GetAudioState returns the current audio state for a node
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

// ControlPump sends a pump command to a node via MQTT
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

	if err := s.mqtt.PublishPumpCommand(req.Value); err != nil {
		logger.Error("Failed to publish pump command: %v", err)
		return fmt.Errorf("failed to send pump command: %w", err)
	}

	state := req.Value == 1
	if err := s.nodeRepo.UpdatePumpState(ctx, nodeID, state); err != nil {
		logger.Warn("Failed to update pump state: %v", err)
	}

	logger.Info("Pump command sent: node=%s value=%d", nodeID, req.Value)
	return nil
}

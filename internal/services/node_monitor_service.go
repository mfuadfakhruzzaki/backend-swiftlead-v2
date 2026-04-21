package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/swiftlead/backend-swiftlet/internal/config"
	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
	"github.com/swiftlead/backend-swiftlet/internal/websocket"
	"github.com/swiftlead/backend-swiftlet/pkg/logger"
)

// NodeMonitorService periodically marks stale nodes as offline and fires node_offline alerts.
type NodeMonitorService struct {
	nodeRepo  repository.NodeRepository
	alertRepo repository.AlertRepository
	wsHub     *websocket.Hub
	cfg       *config.Config
	cancel    context.CancelFunc
}

func NewNodeMonitorService(
	nodeRepo repository.NodeRepository,
	alertRepo repository.AlertRepository,
	wsHub *websocket.Hub,
	cfg *config.Config,
) *NodeMonitorService {
	return &NodeMonitorService{
		nodeRepo:  nodeRepo,
		alertRepo: alertRepo,
		wsHub:     wsHub,
		cfg:       cfg,
	}
}

func (s *NodeMonitorService) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.run(ctx)
}

func (s *NodeMonitorService) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *NodeMonitorService) run(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkStaleNodes(ctx)
		}
	}
}

func (s *NodeMonitorService) checkStaleNodes(ctx context.Context) {
	threshold := time.Duration(s.cfg.NodeOfflineThresholdMinutes) * time.Minute
	cutoff := time.Now().Add(-threshold)

	stale, err := s.nodeRepo.ListStale(ctx, cutoff)
	if err != nil {
		logger.Error("NodeMonitor: failed to list stale nodes: %v", err)
		return
	}

	for _, node := range stale {
		if err := s.nodeRepo.UpdateStatus(ctx, node.ID, models.NodeStatusOffline); err != nil {
			logger.Error("NodeMonitor: failed to mark node %s offline: %v", node.ID, err)
			continue
		}

		// Deduplication: skip if there is already an unresolved node_offline alert
		exists, err := s.alertRepo.HasUnresolved(ctx, models.AlertTypeNodeOffline, &node.ID, nil)
		if err != nil {
			logger.Error("NodeMonitor: HasUnresolved check failed for node %s: %v", node.ID, err)
		}
		if exists {
			continue
		}

		msg := fmt.Sprintf("Node has gone offline — no data received for %d minutes", s.cfg.NodeOfflineThresholdMinutes)
		alert := &models.Alert{
			RBWID:     node.RBWID,
			NodeID:    &node.ID,
			AlertType: models.AlertTypeNodeOffline,
			Severity:  4,
			Message:   &msg,
		}
		if err := s.alertRepo.Create(ctx, alert); err != nil {
			logger.Error("NodeMonitor: failed to create node_offline alert for node %s: %v", node.ID, err)
			continue
		}

		logger.Info("NodeMonitor: node %s (%s) marked offline, alert created", node.ID, node.NodeCode)
		s.broadcastAlert(alert)
	}
}

func (s *NodeMonitorService) broadcastAlert(alert *models.Alert) {
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

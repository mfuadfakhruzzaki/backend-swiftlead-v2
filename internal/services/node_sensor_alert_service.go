package services

import (
	"context"

	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
)

// NodeService handles node business logic
type NodeService struct {
	repo repository.NodeRepository
}

func NewNodeService(repo repository.NodeRepository) *NodeService {
	return &NodeService{repo: repo}
}

func (s *NodeService) Create(ctx context.Context, rbwID string, req *models.CreateNodeRequest) (*models.Node, error) {
	node := &models.Node{
		RBWID:    rbwID,
		NodeType: req.NodeType,
		NodeCode: req.NodeCode,
		ESP32UID: req.ESP32UID,
		HasAudio: req.HasAudio,
		HasPump:  req.HasPump,
	}
	if err := s.repo.Create(ctx, node); err != nil {
		return nil, err
	}
	return node, nil
}

func (s *NodeService) GetByID(ctx context.Context, id string) (*models.Node, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *NodeService) Update(ctx context.Context, id string, req *models.UpdateNodeRequest) (*models.Node, error) {
	node, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.NodeCode != nil {
		node.NodeCode = *req.NodeCode
	}
	if req.ESP32UID != nil {
		node.ESP32UID = req.ESP32UID
	}
	if req.HasAudio != nil {
		node.HasAudio = *req.HasAudio
	}
	if req.HasPump != nil {
		node.HasPump = *req.HasPump
	}
	if err := s.repo.Update(ctx, node); err != nil {
		return nil, err
	}
	return node, nil
}

func (s *NodeService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *NodeService) ListByRBW(ctx context.Context, rbwID string, page, limit int) ([]*models.Node, int, error) {
	offset := (page - 1) * limit
	return s.repo.ListByRBW(ctx, rbwID, limit, offset)
}

// SensorService handles sensor business logic
type SensorService struct {
	repo repository.SensorRepository
}

func NewSensorService(repo repository.SensorRepository) *SensorService {
	return &SensorService{repo: repo}
}

func (s *SensorService) Create(ctx context.Context, nodeID string, req *models.CreateSensorRequest) (*models.Sensor, error) {
	sensor := &models.Sensor{
		NodeID:     nodeID,
		SensorType: req.SensorType,
	}
	if req.SensorName != "" {
		sensor.SensorName = &req.SensorName
	}
	if req.Unit != "" {
		sensor.Unit = &req.Unit
	}
	if err := s.repo.Create(ctx, sensor); err != nil {
		return nil, err
	}
	return sensor, nil
}

func (s *SensorService) GetByID(ctx context.Context, id string) (*models.Sensor, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SensorService) Update(ctx context.Context, id string, req *models.UpdateSensorRequest) (*models.Sensor, error) {
	sensor, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.SensorName != nil {
		sensor.SensorName = req.SensorName
	}
	if req.Unit != nil {
		sensor.Unit = req.Unit
	}
	if req.IsActive != nil {
		sensor.IsActive = *req.IsActive
	}
	if err := s.repo.Update(ctx, sensor); err != nil {
		return nil, err
	}
	return sensor, nil
}

func (s *SensorService) ListByNode(ctx context.Context, nodeID string) ([]*models.Sensor, error) {
	return s.repo.ListByNode(ctx, nodeID)
}

// AlertService handles alert business logic
type AlertService struct {
	repo repository.AlertRepository
}

func NewAlertService(repo repository.AlertRepository) *AlertService {
	return &AlertService{repo: repo}
}

func (s *AlertService) Create(ctx context.Context, req *models.CreateAlertRequest) (*models.Alert, error) {
	alert := &models.Alert{
		RBWID:     req.RBWID,
		NodeID:    req.NodeID,
		SensorID:  req.SensorID,
		AlertType: req.AlertType,
		Severity:  req.Severity,
	}
	if req.Message != "" {
		alert.Message = &req.Message
	}
	if err := s.repo.Create(ctx, alert); err != nil {
		return nil, err
	}
	return alert, nil
}

func (s *AlertService) GetByID(ctx context.Context, id string) (*models.Alert, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AlertService) MarkAsRead(ctx context.Context, id string) error {
	return s.repo.MarkAsRead(ctx, id)
}

func (s *AlertService) Resolve(ctx context.Context, id, resolvedBy string) error {
	return s.repo.Resolve(ctx, id, resolvedBy)
}

func (s *AlertService) List(ctx context.Context, rbwID string, isRead, resolved *bool, page, limit int) ([]*models.Alert, int, error) {
	offset := (page - 1) * limit
	return s.repo.List(ctx, rbwID, isRead, resolved, limit, offset)
}

func (s *AlertService) ListByRBW(ctx context.Context, rbwID string, page, limit int) ([]*models.Alert, int, error) {
	offset := (page - 1) * limit
	return s.repo.ListByRBW(ctx, rbwID, limit, offset)
}

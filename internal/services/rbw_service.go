package services

import (
	"context"

	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
)

// RBWService handles RBW business logic
type RBWService struct {
	repo repository.RBWRepository
}

// NewRBWService creates a new RBW service
func NewRBWService(repo repository.RBWRepository) *RBWService {
	return &RBWService{repo: repo}
}

// Create creates a new RBW
func (s *RBWService) Create(ctx context.Context, ownerID string, req *models.CreateRBWRequest) (*models.RBW, error) {
	rbw := &models.RBW{
		OwnerID:     ownerID,
		Code:        req.Code,
		Name:        req.Name,
		TotalFloors: req.TotalFloors,
	}
	if req.Address != "" {
		rbw.Address = &req.Address
	}
	if req.Latitude != nil {
		rbw.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		rbw.Longitude = req.Longitude
	}
	if req.Description != "" {
		rbw.Description = &req.Description
	}

	if err := s.repo.Create(ctx, rbw); err != nil {
		return nil, err
	}

	return rbw, nil
}

// GetByID retrieves an RBW by ID
func (s *RBWService) GetByID(ctx context.Context, id string) (*models.RBW, error) {
	return s.repo.GetByID(ctx, id)
}

// Update updates an RBW
func (s *RBWService) Update(ctx context.Context, id string, req *models.UpdateRBWRequest) (*models.RBW, error) {
	rbw, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		rbw.Name = *req.Name
	}
	if req.Address != nil {
		rbw.Address = req.Address
	}
	if req.Latitude != nil {
		rbw.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		rbw.Longitude = req.Longitude
	}
	if req.TotalFloors != nil {
		rbw.TotalFloors = *req.TotalFloors
	}
	if req.Description != nil {
		rbw.Description = req.Description
	}
	if req.PhotoURL != nil {
		rbw.PhotoURL = req.PhotoURL
	}

	if err := s.repo.Update(ctx, rbw); err != nil {
		return nil, err
	}

	return rbw, nil
}

// Delete removes an RBW
func (s *RBWService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// List retrieves RBWs with pagination
func (s *RBWService) List(ctx context.Context, ownerID string, page, limit int) ([]*models.RBW, int, error) {
	offset := (page - 1) * limit
	return s.repo.List(ctx, ownerID, limit, offset)
}

// CheckOwnership verifies if a user owns an RBW
func (s *RBWService) CheckOwnership(ctx context.Context, rbwID, userID, userRole string) bool {
	if userRole == models.RoleAdmin {
		return true
	}
	rbw, err := s.repo.GetByID(ctx, rbwID)
	if err != nil {
		return false
	}
	return rbw.OwnerID == userID
}

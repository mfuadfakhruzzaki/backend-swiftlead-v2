package services

import (
	"context"
	"fmt"
	"time"

	"github.com/swiftlead/backend-swiftlet/internal/ai"
	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
	"github.com/swiftlead/backend-swiftlet/pkg/logger"
)

// HarvestService handles harvest business logic
type HarvestService struct {
	repo     repository.HarvestRepository
	aiClient *ai.Client
}

func NewHarvestService(repo repository.HarvestRepository, aiClient *ai.Client) *HarvestService {
	return &HarvestService{repo: repo, aiClient: aiClient}
}

func (s *HarvestService) Create(ctx context.Context, createdBy string, req *models.CreateHarvestRequest) (*models.Harvest, error) {
	harvest := &models.Harvest{
		RBWID:       req.RBWID,
		NodeID:      req.NodeID,
		FloorNo:     req.FloorNo,
		HarvestedAt: req.HarvestedAt,
		NestsCount:  req.NestsCount,
		WeightKg:    req.WeightKg,
		Grade:       req.Grade,
		CreatedBy:   &createdBy,
	}
	if req.Notes != "" {
		harvest.Notes = &req.Notes
	}

	// Calculate cycle days from last harvest
	lastHarvest, err := s.repo.GetLastHarvest(ctx, req.RBWID, req.FloorNo)
	if err == nil && lastHarvest != nil {
		days := int(harvest.HarvestedAt.Sub(lastHarvest.HarvestedAt).Hours() / 24)
		harvest.CycleDays = &days
	}

	// AI grade prediction if no grade provided
	if harvest.Grade == nil || *harvest.Grade == "" {
		if s.aiClient != nil && s.aiClient.IsEnabled() {
			gradeResp, err := s.aiClient.PredictGrade(ctx, &ai.GradePredictionRequest{
				RBWID: req.RBWID,
			})
			if err == nil && gradeResp != nil {
				dbGrade := ai.MapGradeToDBEnum(gradeResp.Grade)
				harvest.Grade = &dbGrade
				logger.Info("AI predicted harvest grade: %s (confidence: %.2f)", gradeResp.Grade, gradeResp.Confidence)
			}
		}
	}

	if err := s.repo.Create(ctx, harvest); err != nil {
		return nil, err
	}
	return harvest, nil
}

func (s *HarvestService) GetByID(ctx context.Context, id string) (*models.Harvest, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *HarvestService) Update(ctx context.Context, id string, req *models.UpdateHarvestRequest) (*models.Harvest, error) {
	harvest, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.FloorNo != nil {
		harvest.FloorNo = *req.FloorNo
	}
	if req.NestsCount != nil {
		harvest.NestsCount = *req.NestsCount
	}
	if req.WeightKg != nil {
		harvest.WeightKg = req.WeightKg
	}
	if req.Grade != nil {
		harvest.Grade = req.Grade
	}
	if req.Notes != nil {
		harvest.Notes = req.Notes
	}
	if err := s.repo.Update(ctx, harvest); err != nil {
		return nil, err
	}
	return harvest, nil
}

func (s *HarvestService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *HarvestService) List(ctx context.Context, rbwID string, page, limit int) ([]*models.Harvest, int, error) {
	offset := (page - 1) * limit
	return s.repo.List(ctx, rbwID, limit, offset)
}

func (s *HarvestService) GetStats(ctx context.Context, rbwID string) (*models.HarvestStats, error) {
	return s.repo.GetStats(ctx, rbwID)
}

// ServiceRequestService handles service request business logic
type ServiceRequestService struct {
	repo repository.ServiceRequestRepository
}

func NewServiceRequestService(repo repository.ServiceRequestRepository) *ServiceRequestService {
	return &ServiceRequestService{repo: repo}
}

func (s *ServiceRequestService) Create(ctx context.Context, requestBy string, req *models.CreateServiceRequestRequest) (*models.ServiceRequest, error) {
	sr := &models.ServiceRequest{
		RBWID:     req.RBWID,
		NodeID:    req.NodeID,
		RequestBy: requestBy,
		Type:      req.Type,
		Status:    models.ServiceStatusDraft,
	}
	if req.Issue != "" {
		sr.Issue = &req.Issue
	}
	if err := s.repo.Create(ctx, sr); err != nil {
		return nil, err
	}
	return sr, nil
}

func (s *ServiceRequestService) GetByID(ctx context.Context, id string) (*models.ServiceRequest, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ServiceRequestService) Update(ctx context.Context, id string, req *models.UpdateServiceRequestRequest, userID, userRole string) (*models.ServiceRequest, error) {
	sr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Status != nil {
		sr.Status = *req.Status
		// Set approved_by if status changes to approved
		if *req.Status == models.ServiceStatusApproved {
			sr.ApprovedBy = &userID
		}
	}
	if req.AssignedTo != nil {
		sr.AssignedTo = req.AssignedTo
	}
	if req.ScheduleDate != nil {
		sr.ScheduleDate = req.ScheduleDate
	}
	if req.Resolution != nil {
		sr.Resolution = req.Resolution
	}
	if req.Notes != nil {
		sr.Notes = req.Notes
	}

	if err := s.repo.Update(ctx, sr); err != nil {
		return nil, err
	}
	return sr, nil
}

func (s *ServiceRequestService) List(ctx context.Context, rbwID, requestBy, assignedTo, status string, page, limit int) ([]*models.ServiceRequest, int, error) {
	offset := (page - 1) * limit
	return s.repo.List(ctx, rbwID, requestBy, assignedTo, status, limit, offset)
}

// TransactionService handles transaction business logic
type TransactionService struct {
	repo repository.TransactionRepository
}

func NewTransactionService(repo repository.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) Create(ctx context.Context, createdBy string, req *models.CreateTransactionRequest) (*models.Transaction, error) {
	tx := &models.Transaction{
		RBWID:           req.RBWID,
		CategoryID:      req.CategoryID,
		Amount:          req.Amount,
		Type:            req.Type,
		TransactionDate: req.TransactionDate,
		CreatedBy:       &createdBy,
	}
	if req.Description != "" {
		tx.Description = &req.Description
	}
	if tx.TransactionDate.IsZero() {
		tx.TransactionDate = time.Now()
	}
	if err := s.repo.Create(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *TransactionService) GetByID(ctx context.Context, id string) (*models.Transaction, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TransactionService) Update(ctx context.Context, id string, req *models.UpdateTransactionRequest) (*models.Transaction, error) {
	tx, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.CategoryID != nil {
		tx.CategoryID = *req.CategoryID
	}
	if req.Amount != nil {
		tx.Amount = *req.Amount
	}
	if req.Description != nil {
		tx.Description = req.Description
	}
	if err := s.repo.Update(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *TransactionService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *TransactionService) ListByRBW(ctx context.Context, rbwID string, page, limit int) ([]*models.Transaction, int, error) {
	offset := (page - 1) * limit
	return s.repo.ListByRBW(ctx, rbwID, limit, offset)
}

func (s *TransactionService) ListCategories(ctx context.Context) ([]*models.TransactionCategory, error) {
	return s.repo.ListCategories(ctx)
}

// Category CRUD

func (s *TransactionService) GetCategoryByID(ctx context.Context, id string) (*models.TransactionCategory, error) {
	return s.repo.GetCategoryByID(ctx, id)
}

func (s *TransactionService) CreateCategory(ctx context.Context, req *models.CreateCategoryRequest) (*models.TransactionCategory, error) {
	cat := &models.TransactionCategory{
		Name: req.Name,
		Type: req.Type,
	}
	if req.Description != "" {
		cat.Description = &req.Description
	}
	if err := s.repo.CreateCategory(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *TransactionService) UpdateCategory(ctx context.Context, id string, req *models.UpdateCategoryRequest) (*models.TransactionCategory, error) {
	cat, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		cat.Name = *req.Name
	}
	if req.Description != nil {
		cat.Description = req.Description
	}
	if err := s.repo.UpdateCategory(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *TransactionService) DeleteCategory(ctx context.Context, id string) error {
	return s.repo.DeleteCategory(ctx, id)
}

// Financial Statement

func (s *TransactionService) GenerateStatement(ctx context.Context, rbwID string, startDate, endDate time.Time) (*models.FinancialStatement, error) {
	// Get summary
	summary, err := s.repo.GetFinancialSummary(ctx, rbwID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get financial summary: %w", err)
	}

	// Get income transactions
	incomes, _, err := s.repo.ListByRBWWithDateRange(ctx, rbwID, startDate, endDate, models.TransactionTypeIncome, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get income transactions: %w", err)
	}

	// Get expense transactions
	expenses, _, err := s.repo.ListByRBWWithDateRange(ctx, rbwID, startDate, endDate, models.TransactionTypeExpense, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense transactions: %w", err)
	}

	return &models.FinancialStatement{
		RBWID:        rbwID,
		StartDate:    startDate,
		EndDate:      endDate,
		TotalIncome:  summary.TotalIncome,
		TotalExpense: summary.TotalExpense,
		Balance:      summary.Balance,
		Incomes:      incomes,
		Expenses:     expenses,
	}, nil
}

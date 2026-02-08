package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/swiftlead/backend-swiftlet/internal/auth"
	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
	"github.com/swiftlead/backend-swiftlet/internal/services"
	"github.com/swiftlead/backend-swiftlet/pkg/response"
)

// ServiceRequestHandler handles service request endpoints
type ServiceRequestHandler struct {
	srService *services.ServiceRequestService
}

func NewServiceRequestHandler(srService *services.ServiceRequestService) *ServiceRequestHandler {
	return &ServiceRequestHandler{srService: srService}
}

// List handles GET /service-requests
func (h *ServiceRequestHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())
	page, limit := getPagination(r)

	rbwID := r.URL.Query().Get("rbw_id")
	status := r.URL.Query().Get("status")

	// Filter by role
	requestBy := ""
	assignedTo := ""
	switch claims.Role {
	case models.RoleFarmer:
		requestBy = claims.UserID
	case models.RoleTechnician:
		assignedTo = claims.UserID
	}

	srs, total, err := h.srService.List(r.Context(), rbwID, requestBy, assignedTo, status, page, limit)
	if err != nil {
		response.InternalError(w, "Failed to list service requests")
		return
	}

	response.SuccessWithMeta(w, srs, &response.Meta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + limit - 1) / limit,
	})
}

// Create handles POST /service-requests
func (h *ServiceRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())

	var req models.CreateServiceRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	sr, err := h.srService.Create(r.Context(), claims.UserID, &req)
	if err != nil {
		response.InternalError(w, "Failed to create service request")
		return
	}

	response.Created(w, "Service request created", sr)
}

// Get handles GET /service-requests/{id}
func (h *ServiceRequestHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	sr, err := h.srService.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrServiceRequestNotFound {
			response.NotFound(w, "Service request not found")
		} else {
			response.InternalError(w, "Failed to get service request")
		}
		return
	}

	response.Success(w, "", sr)
}

// Update handles PATCH /service-requests/{id}
func (h *ServiceRequestHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := auth.GetUserFromContext(r.Context())

	var req models.UpdateServiceRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	sr, err := h.srService.Update(r.Context(), id, &req, claims.UserID, claims.Role)
	if err != nil {
		if err == repository.ErrServiceRequestNotFound {
			response.NotFound(w, "Service request not found")
		} else {
			response.InternalError(w, "Failed to update service request")
		}
		return
	}

	response.Success(w, "Service request updated", sr)
}

// TransactionHandler handles transaction endpoints
type TransactionHandler struct {
	txService *services.TransactionService
}

func NewTransactionHandler(txService *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{txService: txService}
}

// ListByRBW handles GET /rbw/{rbw_id}/transactions
func (h *TransactionHandler) ListByRBW(w http.ResponseWriter, r *http.Request) {
	rbwID := chi.URLParam(r, "rbw_id")
	page, limit := getPagination(r)

	txs, total, err := h.txService.ListByRBW(r.Context(), rbwID, page, limit)
	if err != nil {
		response.InternalError(w, "Failed to list transactions")
		return
	}

	response.SuccessWithMeta(w, txs, &response.Meta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + limit - 1) / limit,
	})
}

// Create handles POST /transactions
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())

	var req models.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	tx, err := h.txService.Create(r.Context(), claims.UserID, &req)
	if err != nil {
		response.InternalError(w, "Failed to create transaction")
		return
	}

	response.Created(w, "Transaction created", tx)
}

// Get handles GET /transactions/{id}
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	tx, err := h.txService.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrTransactionNotFound {
			response.NotFound(w, "Transaction not found")
		} else {
			response.InternalError(w, "Failed to get transaction")
		}
		return
	}

	response.Success(w, "", tx)
}

// Update handles PATCH /transactions/{id}
func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req models.UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	tx, err := h.txService.Update(r.Context(), id, &req)
	if err != nil {
		if err == repository.ErrTransactionNotFound {
			response.NotFound(w, "Transaction not found")
		} else {
			response.InternalError(w, "Failed to update transaction")
		}
		return
	}

	response.Success(w, "Transaction updated", tx)
}

// Delete handles DELETE /transactions/{id}
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.txService.Delete(r.Context(), id); err != nil {
		if err == repository.ErrTransactionNotFound {
			response.NotFound(w, "Transaction not found")
		} else {
			response.InternalError(w, "Failed to delete transaction")
		}
		return
	}

	response.Success(w, "Transaction deleted", nil)
}

// ListCategories handles GET /transaction-categories
func (h *TransactionHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.txService.ListCategories(r.Context())
	if err != nil {
		response.InternalError(w, "Failed to list categories")
		return
	}

	response.Success(w, "", categories)
}

// CreateCategory handles POST /transaction-categories (admin)
func (h *TransactionHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	cat, err := h.txService.CreateCategory(r.Context(), &req)
	if err != nil {
		response.InternalError(w, "Failed to create category")
		return
	}

	response.Created(w, "Category created", cat)
}

// UpdateCategory handles PATCH /transaction-categories/{id} (admin)
func (h *TransactionHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req models.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	cat, err := h.txService.UpdateCategory(r.Context(), id, &req)
	if err != nil {
		response.InternalError(w, "Failed to update category")
		return
	}

	response.Success(w, "Category updated", cat)
}

// DeleteCategory handles DELETE /transaction-categories/{id} (admin)
func (h *TransactionHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.txService.DeleteCategory(r.Context(), id); err != nil {
		if err == repository.ErrCategoryInUse {
			response.Conflict(w, "Cannot delete category: still referenced by existing transactions")
		} else {
			response.InternalError(w, "Failed to delete category")
		}
		return
	}

	response.Success(w, "Category deleted", nil)
}

// GenerateStatement handles POST /financial-statements
func (h *TransactionHandler) GenerateStatement(w http.ResponseWriter, r *http.Request) {
	var req models.GenerateStatementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		response.BadRequest(w, "Invalid start_date format, use YYYY-MM-DD")
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		response.BadRequest(w, "Invalid end_date format, use YYYY-MM-DD")
		return
	}

	statement, err := h.txService.GenerateStatement(r.Context(), req.RBWID, startDate, endDate)
	if err != nil {
		response.InternalError(w, "Failed to generate financial statement")
		return
	}

	response.Success(w, "Financial statement generated", statement)
}

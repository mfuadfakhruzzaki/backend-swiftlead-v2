package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/swiftlead/backend-swiftlet/internal/auth"
	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
	"github.com/swiftlead/backend-swiftlet/internal/services"
	"github.com/swiftlead/backend-swiftlet/pkg/response"
)

// HarvestHandler handles harvest endpoints
type HarvestHandler struct {
	harvestService *services.HarvestService
}

func NewHarvestHandler(harvestService *services.HarvestService) *HarvestHandler {
	return &HarvestHandler{harvestService: harvestService}
}

// List handles GET /harvests
func (h *HarvestHandler) List(w http.ResponseWriter, r *http.Request) {
	rbwID := r.URL.Query().Get("rbw_id")
	page, limit := getPagination(r)

	harvests, total, err := h.harvestService.List(r.Context(), rbwID, page, limit)
	if err != nil {
		response.InternalError(w, "Failed to list harvests")
		return
	}

	response.SuccessWithMeta(w, harvests, &response.Meta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + limit - 1) / limit,
	})
}

// ListByRBW handles GET /rbw/{rbw_id}/harvests
func (h *HarvestHandler) ListByRBW(w http.ResponseWriter, r *http.Request) {
	rbwID := chi.URLParam(r, "rbw_id")
	page, limit := getPagination(r)

	harvests, total, err := h.harvestService.List(r.Context(), rbwID, page, limit)
	if err != nil {
		response.InternalError(w, "Failed to list harvests")
		return
	}

	response.SuccessWithMeta(w, harvests, &response.Meta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + limit - 1) / limit,
	})
}

// Create handles POST /harvests
func (h *HarvestHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())

	var req models.CreateHarvestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	harvest, err := h.harvestService.Create(r.Context(), claims.UserID, &req)
	if err != nil {
		response.InternalError(w, "Failed to create harvest")
		return
	}

	response.Created(w, "Harvest created", harvest)
}

// Get handles GET /harvests/{id}
func (h *HarvestHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	harvest, err := h.harvestService.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrHarvestNotFound {
			response.NotFound(w, "Harvest not found")
		} else {
			response.InternalError(w, "Failed to get harvest")
		}
		return
	}

	response.Success(w, "", harvest)
}

// Update handles PATCH /harvests/{id}
func (h *HarvestHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req models.UpdateHarvestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	harvest, err := h.harvestService.Update(r.Context(), id, &req)
	if err != nil {
		if err == repository.ErrHarvestNotFound {
			response.NotFound(w, "Harvest not found")
		} else {
			response.InternalError(w, "Failed to update harvest")
		}
		return
	}

	response.Success(w, "Harvest updated", harvest)
}

// Delete handles DELETE /harvests/{id}
func (h *HarvestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.harvestService.Delete(r.Context(), id); err != nil {
		if err == repository.ErrHarvestNotFound {
			response.NotFound(w, "Harvest not found")
		} else {
			response.InternalError(w, "Failed to delete harvest")
		}
		return
	}

	response.Success(w, "Harvest deleted", nil)
}

// GetStats handles GET /harvests/stats
func (h *HarvestHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	rbwID := r.URL.Query().Get("rbw_id")

	stats, err := h.harvestService.GetStats(r.Context(), rbwID)
	if err != nil {
		response.InternalError(w, "Failed to get harvest stats")
		return
	}

	response.Success(w, "", stats)
}

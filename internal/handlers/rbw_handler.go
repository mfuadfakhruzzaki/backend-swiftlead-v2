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

// RBWHandler handles RBW endpoints
type RBWHandler struct {
	rbwService *services.RBWService
}

func NewRBWHandler(rbwService *services.RBWService) *RBWHandler {
	return &RBWHandler{rbwService: rbwService}
}

// List handles GET /rbw
func (h *RBWHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())
	page, limit := getPagination(r)

	// Admin sees all, others see only their own
	ownerID := ""
	if claims.Role != models.RoleAdmin {
		ownerID = claims.UserID
	} else if r.URL.Query().Get("owner_id") != "" {
		ownerID = r.URL.Query().Get("owner_id")
	}

	rbws, total, err := h.rbwService.List(r.Context(), ownerID, page, limit)
	if err != nil {
		response.InternalError(w, "Failed to list RBW")
		return
	}

	response.SuccessWithMeta(w, rbws, &response.Meta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + limit - 1) / limit,
	})
}

// Create handles POST /rbw
func (h *RBWHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetUserFromContext(r.Context())

	var req models.CreateRBWRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	rbw, err := h.rbwService.Create(r.Context(), claims.UserID, &req)
	if err != nil {
		if err == repository.ErrRBWAlreadyExists {
			response.Conflict(w, "RBW with this code already exists")
		} else {
			response.InternalError(w, "Failed to create RBW")
		}
		return
	}

	response.Created(w, "RBW created", rbw)
}

// Get handles GET /rbw/{rbw_id}
func (h *RBWHandler) Get(w http.ResponseWriter, r *http.Request) {
	rbwID := chi.URLParam(r, "rbw_id")
	claims := auth.GetUserFromContext(r.Context())

	rbw, err := h.rbwService.GetByID(r.Context(), rbwID)
	if err != nil {
		if err == repository.ErrRBWNotFound {
			response.NotFound(w, "RBW not found")
		} else {
			response.InternalError(w, "Failed to get RBW")
		}
		return
	}

	// Check ownership
	if claims.Role != models.RoleAdmin && rbw.OwnerID != claims.UserID {
		response.Forbidden(w, "You don't have access to this RBW")
		return
	}

	response.Success(w, "", rbw)
}

// Update handles PATCH /rbw/{rbw_id}
func (h *RBWHandler) Update(w http.ResponseWriter, r *http.Request) {
	rbwID := chi.URLParam(r, "rbw_id")
	claims := auth.GetUserFromContext(r.Context())

	// Check ownership
	if !h.rbwService.CheckOwnership(r.Context(), rbwID, claims.UserID, claims.Role) {
		response.Forbidden(w, "You don't have access to this RBW")
		return
	}

	var req models.UpdateRBWRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	rbw, err := h.rbwService.Update(r.Context(), rbwID, &req)
	if err != nil {
		if err == repository.ErrRBWNotFound {
			response.NotFound(w, "RBW not found")
		} else {
			response.InternalError(w, "Failed to update RBW")
		}
		return
	}

	response.Success(w, "RBW updated", rbw)
}

// Delete handles DELETE /rbw/{rbw_id}
func (h *RBWHandler) Delete(w http.ResponseWriter, r *http.Request) {
	rbwID := chi.URLParam(r, "rbw_id")

	if err := h.rbwService.Delete(r.Context(), rbwID); err != nil {
		if err == repository.ErrRBWNotFound {
			response.NotFound(w, "RBW not found")
		} else {
			response.InternalError(w, "Failed to delete RBW")
		}
		return
	}

	response.Success(w, "RBW deleted", nil)
}

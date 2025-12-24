package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/swiftlead/backend-swiftlet/internal/auth"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
	"github.com/swiftlead/backend-swiftlet/internal/services"
	"github.com/swiftlead/backend-swiftlet/pkg/response"
)

// AlertHandler handles alert endpoints
type AlertHandler struct {
	alertService *services.AlertService
}

func NewAlertHandler(alertService *services.AlertService) *AlertHandler {
	return &AlertHandler{alertService: alertService}
}

// List handles GET /alerts
func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	rbwID := r.URL.Query().Get("rbw_id")
	page, limit := getPagination(r)

	// Parse filter params
	var isRead, resolved *bool
	if r.URL.Query().Get("is_read") != "" {
		v, _ := strconv.ParseBool(r.URL.Query().Get("is_read"))
		isRead = &v
	}
	if r.URL.Query().Get("resolved") != "" {
		v, _ := strconv.ParseBool(r.URL.Query().Get("resolved"))
		resolved = &v
	}

	alerts, total, err := h.alertService.List(r.Context(), rbwID, isRead, resolved, page, limit)
	if err != nil {
		response.InternalError(w, "Failed to list alerts")
		return
	}

	response.SuccessWithMeta(w, alerts, &response.Meta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + limit - 1) / limit,
	})
}

// ListByRBW handles GET /rbw/{rbw_id}/alerts
func (h *AlertHandler) ListByRBW(w http.ResponseWriter, r *http.Request) {
	rbwID := chi.URLParam(r, "rbw_id")
	page, limit := getPagination(r)

	alerts, total, err := h.alertService.ListByRBW(r.Context(), rbwID, page, limit)
	if err != nil {
		response.InternalError(w, "Failed to list alerts")
		return
	}

	response.SuccessWithMeta(w, alerts, &response.Meta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + limit - 1) / limit,
	})
}

// MarkAsRead handles PATCH /alerts/{alert_id}/read
func (h *AlertHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	alertID := chi.URLParam(r, "alert_id")

	if err := h.alertService.MarkAsRead(r.Context(), alertID); err != nil {
		if err == repository.ErrAlertNotFound {
			response.NotFound(w, "Alert not found")
		} else {
			response.InternalError(w, "Failed to mark alert as read")
		}
		return
	}

	response.Success(w, "Alert marked as read", nil)
}

// Resolve handles PATCH /alerts/{alert_id}/resolve
func (h *AlertHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	alertID := chi.URLParam(r, "alert_id")
	claims := auth.GetUserFromContext(r.Context())

	if err := h.alertService.Resolve(r.Context(), alertID, claims.UserID); err != nil {
		if err == repository.ErrAlertNotFound {
			response.NotFound(w, "Alert not found")
		} else {
			response.InternalError(w, "Failed to resolve alert")
		}
		return
	}

	response.Success(w, "Alert resolved", nil)
}

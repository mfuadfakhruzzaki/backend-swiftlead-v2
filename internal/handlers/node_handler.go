package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
	"github.com/swiftlead/backend-swiftlet/internal/services"
	"github.com/swiftlead/backend-swiftlet/pkg/response"
)

// NodeHandler handles node endpoints
type NodeHandler struct {
	nodeService   *services.NodeService
	sensorService *services.SensorService
	rbwService    *services.RBWService
}

func NewNodeHandler(nodeService *services.NodeService, sensorService *services.SensorService, rbwService *services.RBWService) *NodeHandler {
	return &NodeHandler{
		nodeService:   nodeService,
		sensorService: sensorService,
		rbwService:    rbwService,
	}
}

// ListByRBW handles GET /rbw/{rbw_id}/nodes
func (h *NodeHandler) ListByRBW(w http.ResponseWriter, r *http.Request) {
	rbwID := chi.URLParam(r, "rbw_id")
	page, limit := getPagination(r)

	nodes, total, err := h.nodeService.ListByRBW(r.Context(), rbwID, page, limit)
	if err != nil {
		response.InternalError(w, "Failed to list nodes")
		return
	}

	response.SuccessWithMeta(w, nodes, &response.Meta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + limit - 1) / limit,
	})
}

// Create handles POST /rbw/{rbw_id}/nodes
func (h *NodeHandler) Create(w http.ResponseWriter, r *http.Request) {
	rbwID := chi.URLParam(r, "rbw_id")

	var req models.CreateNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	node, err := h.nodeService.Create(r.Context(), rbwID, &req)
	if err != nil {
		response.InternalError(w, "Failed to create node")
		return
	}

	response.Created(w, "Node created", node)
}

// Get handles GET /nodes/{node_id}
func (h *NodeHandler) Get(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "node_id")

	node, err := h.nodeService.GetByID(r.Context(), nodeID)
	if err != nil {
		if err == repository.ErrNodeNotFound {
			response.NotFound(w, "Node not found")
		} else {
			response.InternalError(w, "Failed to get node")
		}
		return
	}

	response.Success(w, "", node)
}

// Update handles PATCH /nodes/{node_id}
func (h *NodeHandler) Update(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "node_id")

	var req models.UpdateNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	node, err := h.nodeService.Update(r.Context(), nodeID, &req)
	if err != nil {
		if err == repository.ErrNodeNotFound {
			response.NotFound(w, "Node not found")
		} else {
			response.InternalError(w, "Failed to update node")
		}
		return
	}

	response.Success(w, "Node updated", node)
}

// Delete handles DELETE /nodes/{node_id}
func (h *NodeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "node_id")

	if err := h.nodeService.Delete(r.Context(), nodeID); err != nil {
		if err == repository.ErrNodeNotFound {
			response.NotFound(w, "Node not found")
		} else {
			response.InternalError(w, "Failed to delete node")
		}
		return
	}

	response.Success(w, "Node deleted", nil)
}

// ListSensors handles GET /nodes/{node_id}/sensors
func (h *NodeHandler) ListSensors(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "node_id")

	sensors, err := h.sensorService.ListByNode(r.Context(), nodeID)
	if err != nil {
		response.InternalError(w, "Failed to list sensors")
		return
	}

	response.Success(w, "", sensors)
}

// CreateSensor handles POST /nodes/{node_id}/sensors
func (h *NodeHandler) CreateSensor(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "node_id")

	var req models.CreateSensorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	sensor, err := h.sensorService.Create(r.Context(), nodeID, &req)
	if err != nil {
		response.InternalError(w, "Failed to create sensor")
		return
	}

	response.Created(w, "Sensor created", sensor)
}

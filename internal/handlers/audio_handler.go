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

// AudioHandler handles audio control endpoints
type AudioHandler struct {
	audioService *services.AudioService
}

func NewAudioHandler(audioService *services.AudioService) *AudioHandler {
	return &AudioHandler{audioService: audioService}
}

// ControlAudio handles PATCH /nodes/{node_id}/audio
func (h *AudioHandler) ControlAudio(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "node_id")

	var req models.AudioControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := h.audioService.ControlAudio(r.Context(), nodeID, &req); err != nil {
		if err == repository.ErrNodeNotFound {
			response.NotFound(w, "Node not found")
		} else {
			response.InternalError(w, err.Error())
		}
		return
	}

	response.Success(w, "Audio command sent", nil)
}

// GetAudioState handles GET /nodes/{node_id}/audio
func (h *AudioHandler) GetAudioState(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "node_id")

	node, err := h.audioService.GetAudioState(r.Context(), nodeID)
	if err != nil {
		if err == repository.ErrNodeNotFound {
			response.NotFound(w, "Node not found")
		} else {
			response.InternalError(w, err.Error())
		}
		return
	}

	response.Success(w, "", map[string]interface{}{
		"node_id":          node.ID,
		"has_audio":        node.HasAudio,
		"state_audio_lmb":  node.StateAudioLMB,
		"state_audio_nest": node.StateAudioNest,
	})
}

// ControlPump handles PATCH /nodes/{node_id}/pump
func (h *AudioHandler) ControlPump(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "node_id")

	var req models.PumpControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := h.audioService.ControlPump(r.Context(), nodeID, &req); err != nil {
		if err == repository.ErrNodeNotFound {
			response.NotFound(w, "Node not found")
		} else {
			response.InternalError(w, err.Error())
		}
		return
	}

	response.Success(w, "Pump command sent", nil)
}

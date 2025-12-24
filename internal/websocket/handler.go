package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler handles WebSocket connections
type Handler struct {
	Hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{Hub: hub}
}

// ServeWS handles WebSocket requests
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID := ""
	if claims, ok := r.Context().Value("claims").(map[string]interface{}); ok {
		if id, ok := claims["user_id"].(string); ok {
			userID = id
		}
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade error: %v", err)
		return
	}

	// Create client
	client := NewClient(h.Hub, conn, userID)

	// Register client
	h.Hub.register <- client

	// Start goroutines for reading and writing
	go client.WritePump()
	go client.ReadPump()
}

// Routes returns the WebSocket routes
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ServeWS)
	return r
}

// Stats returns WebSocket statistics
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"connected_clients": h.Hub.GetClientCount(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	data, _ := json.Marshal(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
	w.Write(data)
}

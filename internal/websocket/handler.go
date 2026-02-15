package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/swiftlead/backend-swiftlet/internal/config"
)

// Handler handles WebSocket connections
type Handler struct {
	Hub      *Hub
	cfg      *config.Config
	upgrader websocket.Upgrader
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, cfg *config.Config) *Handler {
	return &Handler{
		Hub: hub,
		cfg: cfg,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				// Allow non-browser clients (empty origin)
				if origin == "" {
					return true
				}

				// Check against allowed origins
				for _, allowed := range cfg.CORSAllowedOrigins {
					if allowed == "*" || allowed == origin {
						return true
					}
				}
				return false
			},
		},
	}
}

// ServeWS handles WebSocket requests
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Debug: log incoming headers to diagnose proxy issues
	log.Printf("[WS] Incoming headers: Connection=%q, Upgrade=%q, Sec-WebSocket-Key=%q, Sec-WebSocket-Version=%q",
		r.Header.Get("Connection"), r.Header.Get("Upgrade"),
		r.Header.Get("Sec-WebSocket-Key"), r.Header.Get("Sec-WebSocket-Version"))

	       // Get user ID from context (set by auth middleware)
	       userID := ""
	       if claims := auth.GetUserFromContext(r.Context()); claims != nil {
		       userID = claims.UserID
	       }

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
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

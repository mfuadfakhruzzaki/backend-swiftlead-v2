package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/swiftlead/backend-swiftlet/internal/auth"
	"github.com/swiftlead/backend-swiftlet/internal/config"
)

// New creates a new router with middleware
func New(cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.RequestTimeout))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   cfg.CORSAllowedMethods,
		AllowedHeaders:   cfg.CORSAllowedHeaders,
		AllowCredentials: cfg.CORSAllowCredentials,
		MaxAge:           300,
	}))

	return r
}

// SetupRoutes configures all routes
func SetupRoutes(r *chi.Mux, cfg *config.Config) {
	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Metrics endpoint (placeholder)
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Metrics placeholder\n"))
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			r.Post("/auth/login", loginHandler)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(cfg.JWTSecret))

			// User routes
			r.Get("/users/me", getMeHandler)
			r.Patch("/users/me", updateMeHandler)

			// RBW routes
			r.Route("/rbw", func(r chi.Router) {
				r.Get("/", listRBWHandler)
				r.Post("/", createRBWHandler)
				r.Get("/{rbw_id}", getRBWHandler)
				r.Patch("/{rbw_id}", updateRBWHandler)
				r.Delete("/{rbw_id}", deleteRBWHandler)

				// Nested node routes
				r.Get("/{rbw_id}/nodes", listNodesHandler)
				r.Post("/{rbw_id}/nodes", createNodeHandler)

				// Nested alert routes
				r.Get("/{rbw_id}/alerts", listRBWAlertsHandler)

				// Nested harvest routes
				r.Get("/{rbw_id}/harvests", listRBWHarvestsHandler)

				// Nested transaction routes
				r.Get("/{rbw_id}/transactions", listRBWTransactionsHandler)
			})

			// Node routes
			r.Route("/nodes", func(r chi.Router) {
				r.Get("/{node_id}", getNodeHandler)
				r.Patch("/{node_id}", updateNodeHandler)
				r.Delete("/{node_id}", deleteNodeHandler)
				r.Get("/{node_id}/sensors", listSensorsHandler)
				r.Post("/{node_id}/sensors", createSensorHandler)
				r.Get("/{node_id}/audio", getAudioStateHandler)
				r.Patch("/{node_id}/audio", controlAudioHandler)
			})

			// Sensor routes
			r.Route("/sensors", func(r chi.Router) {
				r.Get("/{sensor_id}", getSensorHandler)
				r.Patch("/{sensor_id}", updateSensorHandler)
				r.Get("/{sensor_id}/readings", getSensorReadingsHandler)
				r.Post("/{sensor_id}/readings", createSensorReadingHandler)
			})

			// Alert routes
			r.Route("/alerts", func(r chi.Router) {
				r.Get("/", listAlertsHandler)
				r.Patch("/{alert_id}/read", markAlertReadHandler)
				r.Patch("/{alert_id}/resolve", resolveAlertHandler)
			})

			// Service request routes
			r.Route("/service-requests", func(r chi.Router) {
				r.Get("/", listServiceRequestsHandler)
				r.Post("/", createServiceRequestHandler)
				r.Get("/{id}", getServiceRequestHandler)
				r.Patch("/{id}", updateServiceRequestHandler)
			})

			// Harvest routes
			r.Route("/harvests", func(r chi.Router) {
				r.Get("/", listHarvestsHandler)
				r.Post("/", createHarvestHandler)
			})

			// Transaction routes
			r.Route("/transactions", func(r chi.Router) {
				r.Post("/", createTransactionHandler)
			})
			r.Get("/transaction-categories", listTransactionCategoriesHandler)

			// Upload routes
			r.Post("/uploads/avatar", uploadAvatarHandler)
			r.Post("/uploads/rbw/{rbw_id}/photo", uploadRBWPhotoHandler)

			// Admin-only routes
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireAdmin())
				r.Get("/users", listUsersHandler)
				r.Post("/users", createUserHandler)
				r.Post("/auth/register", registerHandler)
			})
		})
	})
}

// Placeholder handlers (to be implemented)
func loginHandler(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) }
func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func getMeHandler(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) }
func updateMeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func listRBWHandler(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) }
func createRBWHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func getRBWHandler(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) }
func updateRBWHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func deleteRBWHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func listNodesHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func createNodeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func getNodeHandler(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) }
func updateNodeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func deleteNodeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func listSensorsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func createSensorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func getSensorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func updateSensorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func getSensorReadingsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func createSensorReadingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func getAudioStateHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func controlAudioHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func listAlertsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func listRBWAlertsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func markAlertReadHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func resolveAlertHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func listServiceRequestsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func createServiceRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func getServiceRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func updateServiceRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func listHarvestsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func listRBWHarvestsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func createHarvestHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func createTransactionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func listRBWTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func listTransactionCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func uploadAvatarHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
func uploadRBWPhotoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Suppress unused warning
var _ = time.Now

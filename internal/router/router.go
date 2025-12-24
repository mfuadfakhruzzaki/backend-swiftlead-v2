package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/swiftlead/backend-swiftlet/internal/app"
	"github.com/swiftlead/backend-swiftlet/internal/auth"
	"github.com/swiftlead/backend-swiftlet/internal/config"
	"github.com/swiftlead/backend-swiftlet/internal/metrics"
	"github.com/swiftlead/backend-swiftlet/internal/models"
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
	r.Use(metrics.PrometheusMiddleware)

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

// SetupRoutes configures all routes with handlers from container
func SetupRoutes(r *chi.Mux, cfg *config.Config, c *app.Container) {
	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			r.Post("/auth/login", c.AuthHandler.Login)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(cfg.JWTSecret))

			// User routes
			r.Get("/users/me", c.UserHandler.GetMe)
			r.Patch("/users/me", c.UserHandler.UpdateMe)

			// RBW routes
			r.Route("/rbw", func(r chi.Router) {
				r.Get("/", c.RBWHandler.List)
				r.Post("/", c.RBWHandler.Create)
				r.Get("/{rbw_id}", c.RBWHandler.Get)
				r.Patch("/{rbw_id}", c.RBWHandler.Update)
				r.Delete("/{rbw_id}", c.RBWHandler.Delete)

				// Nested routes
				r.Get("/{rbw_id}/nodes", c.NodeHandler.ListByRBW)
				r.Post("/{rbw_id}/nodes", c.NodeHandler.Create)
				r.Get("/{rbw_id}/alerts", c.AlertHandler.ListByRBW)
				r.Get("/{rbw_id}/harvests", c.HarvestHandler.ListByRBW)
				r.Get("/{rbw_id}/transactions", c.TransactionHandler.ListByRBW)
			})

			// Node routes
			r.Route("/nodes", func(r chi.Router) {
				r.Get("/{node_id}", c.NodeHandler.Get)
				r.Patch("/{node_id}", c.NodeHandler.Update)
				r.Delete("/{node_id}", c.NodeHandler.Delete)
				r.Get("/{node_id}/sensors", c.NodeHandler.ListSensors)
				r.Post("/{node_id}/sensors", c.NodeHandler.CreateSensor)
			})

			// Sensor routes
			r.Route("/sensors", func(r chi.Router) {
				r.Get("/{sensor_id}", c.SensorHandler.Get)
				r.Patch("/{sensor_id}", c.SensorHandler.Update)
				r.Get("/{sensor_id}/readings", c.SensorHandler.GetReadings)
				r.Post("/{sensor_id}/readings", c.SensorHandler.CreateReading)
			})

			// Alert routes
			r.Route("/alerts", func(r chi.Router) {
				r.Get("/", c.AlertHandler.List)
				r.Patch("/{alert_id}/read", c.AlertHandler.MarkAsRead)
				r.Patch("/{alert_id}/resolve", c.AlertHandler.Resolve)
			})

			// Service request routes
			r.Route("/service-requests", func(r chi.Router) {
				r.Get("/", c.ServiceRequestHandler.List)
				r.Post("/", c.ServiceRequestHandler.Create)
				r.Get("/{id}", c.ServiceRequestHandler.Get)
				r.Patch("/{id}", c.ServiceRequestHandler.Update)
			})

			// Harvest routes
			r.Route("/harvests", func(r chi.Router) {
				r.Get("/", c.HarvestHandler.List)
				r.Post("/", c.HarvestHandler.Create)
			})

			// Transaction routes
			r.Post("/transactions", c.TransactionHandler.Create)
			r.Get("/transaction-categories", c.TransactionHandler.ListCategories)

			// Upload routes (only if storage is available)
			r.Route("/uploads", func(r chi.Router) {
				r.Post("/avatar", func(w http.ResponseWriter, r *http.Request) {
					if c.UploadHandler == nil {
						http.Error(w, "Storage not available", http.StatusServiceUnavailable)
						return
					}
					c.UploadHandler.UploadAvatar(w, r)
				})
				r.Post("/rbw/{rbw_id}/photo", func(w http.ResponseWriter, r *http.Request) {
					if c.UploadHandler == nil {
						http.Error(w, "Storage not available", http.StatusServiceUnavailable)
						return
					}
					c.UploadHandler.UploadRBWPhoto(w, r)
				})
			})

			// Admin-only routes
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole(models.RoleAdmin))
				r.Get("/users", c.UserHandler.List)
				r.Post("/users", c.UserHandler.Create)
				r.Post("/auth/register", c.AuthHandler.Register)
			})
		})
	})
}

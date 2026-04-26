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
			r.Post("/auth/register", c.AuthHandler.PublicRegister)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(cfg.JWTSecret))

			// Auth routes
			r.Post("/auth/change-password", c.AuthHandler.ChangePassword)

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

				// Audio control routes
				r.Get("/{node_id}/audio", c.AudioHandler.GetAudioState)
				r.Patch("/{node_id}/audio", c.AudioHandler.ControlAudio)
				r.Patch("/{node_id}/pump", c.AudioHandler.ControlPump)
				r.Patch("/{node_id}/pump/mode", c.AudioHandler.TogglePumpMode)

				// AI endpoints using real-time sensor data from DB (no body needed)
				r.Post("/{node_id}/ai/analyze", c.AIHandler.AnalyzeNode)
				r.Post("/{node_id}/ai/predict-pump", c.AIHandler.PredictPumpNode)
				r.Post("/{node_id}/ai/predict-grade", c.AIHandler.PredictGradeNode)
				r.Post("/{node_id}/ai/anomaly-detect", c.AIHandler.AnomalyDetectNode)
			})

			// Sensor routes
			r.Route("/sensors", func(r chi.Router) {
				r.Get("/{sensor_id}", c.SensorHandler.Get)
				r.Patch("/{sensor_id}", c.SensorHandler.Update)
				r.Get("/{sensor_id}/readings", c.SensorHandler.GetReadings)
				r.Post("/{sensor_id}/readings", c.SensorHandler.CreateReading)
				r.Get("/{sensor_id}/trend", c.SensorHandler.GetTrend)
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

			// Harvest routes (full CRUD)
			r.Route("/harvests", func(r chi.Router) {
				r.Get("/", c.HarvestHandler.List)
				r.Post("/", c.HarvestHandler.Create)
				r.Get("/stats", c.HarvestHandler.GetStats)
				r.Get("/{id}", c.HarvestHandler.Get)
				r.Patch("/{id}", c.HarvestHandler.Update)
				r.Delete("/{id}", c.HarvestHandler.Delete)
			})

			// Transaction routes (full CRUD)
			r.Route("/transactions", func(r chi.Router) {
				r.Post("/", c.TransactionHandler.Create)
				r.Get("/{id}", c.TransactionHandler.Get)
				r.Patch("/{id}", c.TransactionHandler.Update)
				r.Delete("/{id}", c.TransactionHandler.Delete)
			})

			// Transaction category routes
			r.Route("/transaction-categories", func(r chi.Router) {
				r.Get("/", c.TransactionHandler.ListCategories)
			})

			// Financial statement routes
			r.Post("/financial-statements", c.TransactionHandler.GenerateStatement)

			// AI proxy routes
			r.Route("/ai", func(r chi.Router) {
				r.Get("/health", c.AIHandler.HealthCheck)
				r.Post("/predict-grade", c.AIHandler.PredictGrade)
				r.Post("/predict-pump", c.AIHandler.PredictPump)
				r.Post("/analyze", c.AIHandler.Analyze)
				r.Post("/anomaly-detect", c.AIHandler.AnomalyDetect)
			})

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
				r.Post("/auth/admin/register", c.AuthHandler.Register)
				r.Post("/auth/forgot-password", c.AuthHandler.ForgotPassword)

				// Admin category management
				r.Post("/transaction-categories", c.TransactionHandler.CreateCategory)
				r.Patch("/transaction-categories/{id}", c.TransactionHandler.UpdateCategory)
				r.Delete("/transaction-categories/{id}", c.TransactionHandler.DeleteCategory)
			})

			// WebSocket route
			r.Get("/ws", c.WSHandler.ServeWS)
			r.Get("/ws/stats", c.WSHandler.Stats)
		})
	})
}

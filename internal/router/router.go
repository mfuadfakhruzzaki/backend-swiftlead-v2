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
func SetupRoutes(r *chi.Mux, cfg *config.Config, container *app.Container) {
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
			r.Post("/auth/login", container.AuthHandler.Login)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(cfg.JWTSecret))

			// User routes
			r.Get("/users/me", container.UserHandler.GetMe)
			r.Patch("/users/me", container.UserHandler.UpdateMe)

			// RBW routes
			r.Route("/rbw", func(r chi.Router) {
				r.Get("/", container.RBWHandler.List)
				r.Post("/", container.RBWHandler.Create)
				r.Get("/{rbw_id}", container.RBWHandler.Get)
				r.Patch("/{rbw_id}", container.RBWHandler.Update)
				r.Delete("/{rbw_id}", container.RBWHandler.Delete)
			})

			// Admin-only routes
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole(models.RoleAdmin))
				r.Get("/users", container.UserHandler.List)
				r.Post("/users", container.UserHandler.Create)
				r.Post("/auth/register", container.AuthHandler.Register)
			})
		})
	})
}

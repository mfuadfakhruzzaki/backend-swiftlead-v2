package app

import (
	"database/sql"

	"github.com/swiftlead/backend-swiftlet/internal/ai"
	"github.com/swiftlead/backend-swiftlet/internal/config"
	"github.com/swiftlead/backend-swiftlet/internal/handlers"
	"github.com/swiftlead/backend-swiftlet/internal/mqtt"
	"github.com/swiftlead/backend-swiftlet/internal/repository"
	"github.com/swiftlead/backend-swiftlet/internal/services"
	"github.com/swiftlead/backend-swiftlet/internal/storage"
	"github.com/swiftlead/backend-swiftlet/internal/websocket"
)

// Container holds all application dependencies
type Container struct {
	Config *config.Config
	DB     *sql.DB

	// Storage
	Storage *storage.MinIOClient

	// MQTT
	MQTT *mqtt.Client

	// AI
	AI *ai.Client

	// Repositories
	UserRepo           repository.UserRepository
	RBWRepo            repository.RBWRepository
	NodeRepo           repository.NodeRepository
	SensorRepo         repository.SensorRepository
	TelemetryRepo      repository.TelemetryRepository
	AlertRepo          repository.AlertRepository
	HarvestRepo        repository.HarvestRepository
	ServiceRequestRepo repository.ServiceRequestRepository
	TransactionRepo    repository.TransactionRepository

	// Services
	UserService           *services.UserService
	RBWService            *services.RBWService
	NodeService           *services.NodeService
	SensorService         *services.SensorService
	AlertService          *services.AlertService
	TelemetryService      *services.TelemetryService
	HarvestService        *services.HarvestService
	ServiceRequestService *services.ServiceRequestService
	TransactionService    *services.TransactionService

	// Handlers
	AuthHandler           *handlers.AuthHandler
	UserHandler           *handlers.UserHandler
	RBWHandler            *handlers.RBWHandler
	NodeHandler           *handlers.NodeHandler
	SensorHandler         *handlers.SensorHandler
	AlertHandler          *handlers.AlertHandler
	HarvestHandler        *handlers.HarvestHandler
	ServiceRequestHandler *handlers.ServiceRequestHandler
	TransactionHandler    *handlers.TransactionHandler
	UploadHandler         *handlers.UploadHandler

	// WebSocket
	WSHub     *websocket.Hub
	WSHandler *websocket.Handler
}

// NewContainer creates and wires all dependencies
func NewContainer(cfg *config.Config, db *sql.DB) *Container {
	c := &Container{
		Config: cfg,
		DB:     db,
	}

	// Initialize repositories
	c.UserRepo = repository.NewUserRepository(db)
	c.RBWRepo = repository.NewRBWRepository(db)
	c.NodeRepo = repository.NewNodeRepository(db)
	c.SensorRepo = repository.NewSensorRepository(db)
	c.TelemetryRepo = repository.NewTelemetryRepository(db)
	c.AlertRepo = repository.NewAlertRepository(db)
	c.HarvestRepo = repository.NewHarvestRepository(db)
	c.ServiceRequestRepo = repository.NewServiceRequestRepository(db)
	c.TransactionRepo = repository.NewTransactionRepository(db)

	// Initialize services
	c.UserService = services.NewUserService(c.UserRepo, cfg.JWTSecret, cfg.JWTExpirationHours)
	c.RBWService = services.NewRBWService(c.RBWRepo)
	c.NodeService = services.NewNodeService(c.NodeRepo)
	c.SensorService = services.NewSensorService(c.SensorRepo)
	c.AlertService = services.NewAlertService(c.AlertRepo)
	c.TelemetryService = services.NewTelemetryService(c.NodeRepo, c.SensorRepo, c.TelemetryRepo, c.AlertRepo, cfg)
	c.HarvestService = services.NewHarvestService(c.HarvestRepo)
	c.ServiceRequestService = services.NewServiceRequestService(c.ServiceRequestRepo)
	c.TransactionService = services.NewTransactionService(c.TransactionRepo)

	// Initialize handlers
	c.AuthHandler = handlers.NewAuthHandler(c.UserService)
	c.UserHandler = handlers.NewUserHandler(c.UserService)
	c.RBWHandler = handlers.NewRBWHandler(c.RBWService)
	c.NodeHandler = handlers.NewNodeHandler(c.NodeService, c.SensorService, c.RBWService)
	c.SensorHandler = handlers.NewSensorHandler(c.SensorService, c.TelemetryService)
	c.AlertHandler = handlers.NewAlertHandler(c.AlertService)
	c.HarvestHandler = handlers.NewHarvestHandler(c.HarvestService)
	c.ServiceRequestHandler = handlers.NewServiceRequestHandler(c.ServiceRequestService)
	c.TransactionHandler = handlers.NewTransactionHandler(c.TransactionService)

	// Initialize AI client
	c.AI = ai.NewClient(cfg.AIEngineURL, cfg.AIEngineTimeout, cfg.AIEngineEnabled)

	// Initialize MQTT client
	c.MQTT = mqtt.NewClient(cfg)
	c.MQTT.SetHandler(c.TelemetryService.ProcessSensorPayload)

	// Initialize WebSocket hub
	c.WSHub = websocket.NewHub()
	c.WSHandler = websocket.NewHandler(c.WSHub, cfg)
	go c.WSHub.Run()

	return c
}

// InitStorage initializes MinIO storage (optional, may fail)
func (c *Container) InitStorage() error {
	s, err := storage.NewMinIOClient(c.Config)
	if err != nil {
		return err
	}
	c.Storage = s
	// Initialize upload handler after storage is ready
	c.UploadHandler = handlers.NewUploadHandler(s, c.UserService, c.RBWService, c.Config)
	return nil
}

// ConnectMQTT connects to MQTT broker (optional, may fail)
func (c *Container) ConnectMQTT() error {
	return c.MQTT.Connect()
}

// Close cleans up resources
func (c *Container) Close() {
	if c.MQTT != nil {
		c.MQTT.Disconnect()
	}
}

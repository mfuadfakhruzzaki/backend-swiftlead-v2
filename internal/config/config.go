package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port     string
	Env      string
	LogLevel string
	Debug    bool

	// Database
	PostgresDSN    string
	DBMaxOpenConns int
	DBMaxIdleConns int

	// JWT
	JWTSecret          string
	JWTExpirationHours int

	// MinIO
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool
	MinIOPublicURL string

	// MQTT
	MQTTBroker         string
	MQTTUsername       string
	MQTTPassword       string
	MQTTClientID       string
	MQTTTopicPrefix    string
	MQTTKeepAlive      int
	MQTTConnectTimeout int
	MQTTUseTLS         bool
	MQTTCACert         string // Path to CA certificate
	MQTTClientCert     string // Path to client certificate
	MQTTClientKey      string // Path to client private key
	MQTTSkipVerify     bool   // Skip server certificate verification

	// AI Engine
	AIEngineURL     string
	AIEngineTimeout int
	AIEngineEnabled bool

	// CORS
	CORSAllowedOrigins   []string
	CORSAllowedMethods   []string
	CORSAllowedHeaders   []string
	CORSAllowCredentials bool

	// Sensor Thresholds
	TempHighThreshold           float64
	TempLowThreshold            float64
	HumidHighThreshold          float64
	HumidLowThreshold           float64
	AmmoniaHighThreshold        float64
	NodeOfflineThresholdMinutes int
	SensorDataRetentionDays     int

	// File Uploads
	MaxUploadSizeMB     int
	MaxAvatarSizeMB     int
	AllowedImageFormats []string

	// Server Timeouts
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration
	ServerIdleTimeout  time.Duration
	ShutdownTimeout    time.Duration
	RequestTimeout     time.Duration
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		// Server
		Port:     getEnv("PORT", "8080"),
		Env:      getEnv("ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Debug:    getEnvBool("DEBUG", false),

		// Database
		PostgresDSN:    getEnv("POSTGRES_DSN", ""),
		DBMaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),

		// JWT
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTExpirationHours: getEnvInt("JWT_EXPIRATION_HOURS", 24),

		// MinIO
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", ""),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", ""),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", ""),
		MinIOBucket:    getEnv("MINIO_BUCKET", "swiftlet"),
		MinIOUseSSL:    getEnvBool("MINIO_USE_SSL", false),
		MinIOPublicURL: getEnv("MINIO_PUBLIC_URL", ""),

		// MQTT
		MQTTBroker:         getEnv("MQTT_BROKER", ""),
		MQTTUsername:       getEnv("MQTT_USERNAME", ""),
		MQTTPassword:       getEnv("MQTT_PASSWORD", ""),
		MQTTClientID:       getEnv("MQTT_CLIENT_ID", "swiftlet-backend"),
		MQTTTopicPrefix:    getEnv("MQTT_TOPIC_PREFIX", "swiftlead/tel"),
		MQTTKeepAlive:      getEnvInt("MQTT_KEEPALIVE", 30),
		MQTTConnectTimeout: getEnvInt("MQTT_CONNECT_TIMEOUT", 5),
		MQTTUseTLS:         getEnvBool("MQTT_USE_TLS", false),
		MQTTCACert:         getEnv("MQTT_CA_CERT", ""),
		MQTTClientCert:     getEnv("MQTT_CLIENT_CERT", ""),
		MQTTClientKey:      getEnv("MQTT_CLIENT_KEY", ""),
		MQTTSkipVerify:     getEnvBool("MQTT_SKIP_VERIFY", false),

		// AI Engine
		AIEngineURL:     getEnv("AI_ENGINE_URL", "http://localhost:8000"),
		AIEngineTimeout: getEnvInt("AI_ENGINE_TIMEOUT", 5),
		AIEngineEnabled: getEnvBool("AI_ENGINE_ENABLED", false),

		// CORS
		CORSAllowedOrigins:   getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
		CORSAllowedMethods:   getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		CORSAllowedHeaders:   getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}),
		CORSAllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", true),

		// Sensor Thresholds
		TempHighThreshold:           getEnvFloat("TEMP_HIGH_THRESHOLD", 35.0),
		TempLowThreshold:            getEnvFloat("TEMP_LOW_THRESHOLD", 20.0),
		HumidHighThreshold:          getEnvFloat("HUMID_HIGH_THRESHOLD", 85.0),
		HumidLowThreshold:           getEnvFloat("HUMID_LOW_THRESHOLD", 60.0),
		AmmoniaHighThreshold:        getEnvFloat("AMMONIA_HIGH_THRESHOLD", 25.0),
		NodeOfflineThresholdMinutes: getEnvInt("NODE_OFFLINE_THRESHOLD_MINUTES", 15),
		SensorDataRetentionDays:     getEnvInt("SENSOR_DATA_RETENTION_DAYS", 365),

		// File Uploads
		MaxUploadSizeMB:     getEnvInt("MAX_UPLOAD_SIZE_MB", 10),
		MaxAvatarSizeMB:     getEnvInt("MAX_AVATAR_SIZE_MB", 2),
		AllowedImageFormats: getEnvSlice("ALLOWED_IMAGE_FORMATS", []string{"jpg", "jpeg", "png", "webp"}),

		// Server Timeouts
		ServerReadTimeout:  time.Duration(getEnvInt("SERVER_READ_TIMEOUT", 15)) * time.Second,
		ServerWriteTimeout: time.Duration(getEnvInt("SERVER_WRITE_TIMEOUT", 15)) * time.Second,
		ServerIdleTimeout:  time.Duration(getEnvInt("SERVER_IDLE_TIMEOUT", 60)) * time.Second,
		ShutdownTimeout:    time.Duration(getEnvInt("SHUTDOWN_TIMEOUT", 30)) * time.Second,
		RequestTimeout:     time.Duration(getEnvInt("REQUEST_TIMEOUT", 60)) * time.Second,
	}
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// Validate checks that all required configuration variables are set
func (c *Config) Validate() error {
	if c.PostgresDSN == "" {
		return fmt.Errorf("POSTGRES_DSN is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.MinIOEndpoint == "" {
		return fmt.Errorf("MINIO_ENDPOINT is required")
	}
	if c.MinIOAccessKey == "" {
		return fmt.Errorf("MINIO_ACCESS_KEY is required")
	}
	if c.MinIOSecretKey == "" {
		return fmt.Errorf("MINIO_SECRET_KEY is required")
	}
	if c.MQTTBroker == "" {
		return fmt.Errorf("MQTT_BROKER is required")
	}
	return nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

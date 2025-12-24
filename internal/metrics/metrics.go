package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "swiftlet_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "swiftlet_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Sensor metrics
	SensorReadingsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "swiftlet_sensor_readings_total",
			Help: "Total number of sensor readings processed",
		},
		[]string{"sensor_type", "rbw_id"},
	)

	SensorValue = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "swiftlet_sensor_value",
			Help: "Current sensor value",
		},
		[]string{"sensor_id", "sensor_type", "node_id"},
	)

	// Alert metrics
	AlertsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "swiftlet_alerts_total",
			Help: "Total number of alerts generated",
		},
		[]string{"alert_type", "severity"},
	)

	ActiveAlerts = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "swiftlet_active_alerts",
			Help: "Number of unresolved alerts",
		},
	)

	// MQTT metrics
	MQTTMessagesReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "swiftlet_mqtt_messages_received_total",
			Help: "Total MQTT messages received",
		},
	)

	MQTTMessageErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "swiftlet_mqtt_message_errors_total",
			Help: "Total MQTT message processing errors",
		},
	)

	MQTTConnectionStatus = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "swiftlet_mqtt_connection_status",
			Help: "MQTT connection status (1=connected, 0=disconnected)",
		},
	)

	// Node metrics
	NodesOnline = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "swiftlet_nodes_online",
			Help: "Number of online nodes",
		},
	)

	// AI Engine metrics
	AIRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "swiftlet_ai_requests_total",
			Help: "Total AI Engine requests",
		},
		[]string{"endpoint", "status"},
	)

	AIRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "swiftlet_ai_request_duration_seconds",
			Help:    "AI Engine request duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)
)

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(method, path, status string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// RecordSensorReading records sensor reading metrics
func RecordSensorReading(sensorType, rbwID string, value float64, sensorID, nodeID string) {
	SensorReadingsTotal.WithLabelValues(sensorType, rbwID).Inc()
	SensorValue.WithLabelValues(sensorID, sensorType, nodeID).Set(value)
}

// RecordAlert records alert metrics
func RecordAlert(alertType string, severity int) {
	AlertsTotal.WithLabelValues(alertType, severityToString(severity)).Inc()
}

// RecordMQTTMessage records MQTT message metrics
func RecordMQTTMessage(success bool) {
	MQTTMessagesReceived.Inc()
	if !success {
		MQTTMessageErrors.Inc()
	}
}

// SetMQTTConnectionStatus sets MQTT connection status
func SetMQTTConnectionStatus(connected bool) {
	if connected {
		MQTTConnectionStatus.Set(1)
	} else {
		MQTTConnectionStatus.Set(0)
	}
}

func severityToString(severity int) string {
	switch severity {
	case 1:
		return "info"
	case 2:
		return "low"
	case 3:
		return "medium"
	case 4:
		return "high"
	case 5:
		return "critical"
	default:
		return "unknown"
	}
}

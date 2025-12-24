package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/swiftlead/backend-swiftlet/internal/models"
)

var (
	ErrSensorNotFound = errors.New("sensor not found")
)

// SensorRepository interface
type SensorRepository interface {
	Create(ctx context.Context, sensor *models.Sensor) error
	GetByID(ctx context.Context, id string) (*models.Sensor, error)
	Update(ctx context.Context, sensor *models.Sensor) error
	ListByNode(ctx context.Context, nodeID string) ([]*models.Sensor, error)
}

// TelemetryRepository interface
type TelemetryRepository interface {
	CreateReading(ctx context.Context, reading *models.SensorReading) error
	GetReadings(ctx context.Context, sensorID string, from, to time.Time, limit int) ([]*models.SensorReading, error)
	GetLatestReading(ctx context.Context, sensorID string) (*models.SensorReading, error)
}

type sensorRepository struct {
	db *sql.DB
}

func NewSensorRepository(db *sql.DB) SensorRepository {
	return &sensorRepository{db: db}
}

func (r *sensorRepository) Create(ctx context.Context, sensor *models.Sensor) error {
	query := `
		INSERT INTO sensors (node_id, sensor_type, sensor_name, unit)
		VALUES ($1, $2, $3, $4)
		RETURNING id, is_active, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		sensor.NodeID, sensor.SensorType, sensor.SensorName, sensor.Unit,
	).Scan(&sensor.ID, &sensor.IsActive, &sensor.CreatedAt, &sensor.UpdatedAt)
}

func (r *sensorRepository) GetByID(ctx context.Context, id string) (*models.Sensor, error) {
	query := `
		SELECT id, node_id, sensor_type, sensor_name, unit, is_active, created_at, updated_at
		FROM sensors WHERE id = $1
	`
	sensor := &models.Sensor{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sensor.ID, &sensor.NodeID, &sensor.SensorType, &sensor.SensorName,
		&sensor.Unit, &sensor.IsActive, &sensor.CreatedAt, &sensor.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrSensorNotFound
	}
	if err != nil {
		return nil, err
	}
	return sensor, nil
}

func (r *sensorRepository) Update(ctx context.Context, sensor *models.Sensor) error {
	query := `
		UPDATE sensors 
		SET sensor_name = $1, unit = $2, is_active = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		sensor.SensorName, sensor.Unit, sensor.IsActive, sensor.ID,
	).Scan(&sensor.UpdatedAt)
	if err == sql.ErrNoRows {
		return ErrSensorNotFound
	}
	return err
}

func (r *sensorRepository) ListByNode(ctx context.Context, nodeID string) ([]*models.Sensor, error) {
	query := `
		SELECT id, node_id, sensor_type, sensor_name, unit, is_active, created_at, updated_at
		FROM sensors WHERE node_id = $1 ORDER BY sensor_type
	`
	rows, err := r.db.QueryContext(ctx, query, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sensors []*models.Sensor
	for rows.Next() {
		sensor := &models.Sensor{}
		if err := rows.Scan(
			&sensor.ID, &sensor.NodeID, &sensor.SensorType, &sensor.SensorName,
			&sensor.Unit, &sensor.IsActive, &sensor.CreatedAt, &sensor.UpdatedAt,
		); err != nil {
			return nil, err
		}
		sensors = append(sensors, sensor)
	}
	return sensors, rows.Err()
}

// Telemetry repository implementation
type telemetryRepository struct {
	db *sql.DB
}

func NewTelemetryRepository(db *sql.DB) TelemetryRepository {
	return &telemetryRepository{db: db}
}

func (r *telemetryRepository) CreateReading(ctx context.Context, reading *models.SensorReading) error {
	query := `
		INSERT INTO sensor_readings (sensor_id, recorded_at, value, is_anomaly)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	if reading.RecordedAt.IsZero() {
		reading.RecordedAt = time.Now()
	}
	return r.db.QueryRowContext(ctx, query,
		reading.SensorID, reading.RecordedAt, reading.Value, reading.IsAnomaly,
	).Scan(&reading.ID, &reading.CreatedAt)
}

func (r *telemetryRepository) GetReadings(ctx context.Context, sensorID string, from, to time.Time, limit int) ([]*models.SensorReading, error) {
	query := `
		SELECT id, sensor_id, recorded_at, value, is_anomaly, created_at
		FROM sensor_readings 
		WHERE sensor_id = $1 AND recorded_at >= $2 AND recorded_at <= $3
		ORDER BY recorded_at DESC
		LIMIT $4
	`
	rows, err := r.db.QueryContext(ctx, query, sensorID, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []*models.SensorReading
	for rows.Next() {
		reading := &models.SensorReading{}
		if err := rows.Scan(
			&reading.ID, &reading.SensorID, &reading.RecordedAt,
			&reading.Value, &reading.IsAnomaly, &reading.CreatedAt,
		); err != nil {
			return nil, err
		}
		readings = append(readings, reading)
	}
	return readings, rows.Err()
}

func (r *telemetryRepository) GetLatestReading(ctx context.Context, sensorID string) (*models.SensorReading, error) {
	query := `
		SELECT id, sensor_id, recorded_at, value, is_anomaly, created_at
		FROM sensor_readings 
		WHERE sensor_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1
	`
	reading := &models.SensorReading{}
	err := r.db.QueryRowContext(ctx, query, sensorID).Scan(
		&reading.ID, &reading.SensorID, &reading.RecordedAt,
		&reading.Value, &reading.IsAnomaly, &reading.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return reading, nil
}

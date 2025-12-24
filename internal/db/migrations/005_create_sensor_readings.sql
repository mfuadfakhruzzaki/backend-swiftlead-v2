-- 005_create_sensor_readings.sql
-- Sensor readings table (TimescaleDB hypertable for time-series data)

CREATE TABLE IF NOT EXISTS sensor_readings (
    id BIGSERIAL,
    sensor_id UUID NOT NULL REFERENCES sensors(id) ON DELETE CASCADE,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    value DOUBLE PRECISION NOT NULL,
    is_anomaly BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (recorded_at, id)
);

-- Create hypertable for time-series optimization
SELECT create_hypertable('sensor_readings', 'recorded_at', 
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sensor_readings_sensor_id ON sensor_readings(sensor_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_sensor_readings_anomaly ON sensor_readings(is_anomaly) WHERE is_anomaly = TRUE;

-- Compression policy (optional - requires TimescaleDB with compression enabled)
-- To enable: ALTER TABLE sensor_readings SET (timescaledb.compress);
-- SELECT add_compression_policy('sensor_readings', INTERVAL '7 days', if_not_exists => true);

-- Retention policy (optional)
-- SELECT add_retention_policy('sensor_readings', INTERVAL '365 days', if_not_exists => true);

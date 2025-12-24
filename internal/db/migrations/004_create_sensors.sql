-- 004_create_sensors.sql
-- Sensors table

CREATE TABLE IF NOT EXISTS sensors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    sensor_type VARCHAR(50) NOT NULL CHECK (sensor_type IN ('temp', 'humid', 'ammonia')),
    sensor_name VARCHAR(255),
    unit VARCHAR(20),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sensors_node_id ON sensors(node_id);
CREATE INDEX IF NOT EXISTS idx_sensors_type ON sensors(sensor_type);

-- Updated at trigger
DROP TRIGGER IF EXISTS update_sensors_updated_at ON sensors;
CREATE TRIGGER update_sensors_updated_at
    BEFORE UPDATE ON sensors
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 006_create_alerts.sql
-- Alerts table

CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rbw_id UUID NOT NULL REFERENCES rbw(id) ON DELETE CASCADE,
    node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,
    sensor_id UUID REFERENCES sensors(id) ON DELETE SET NULL,
    alert_type VARCHAR(50) NOT NULL CHECK (alert_type IN (
        'temp_high', 'temp_low', 
        'humid_high', 'humid_low', 
        'ammonia_high', 
        'node_offline', 
        'ai_anomaly'
    )),
    severity INT NOT NULL DEFAULT 3 CHECK (severity >= 1 AND severity <= 5),
    message TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMPTZ,
    resolved_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_alerts_rbw_id ON alerts(rbw_id);
CREATE INDEX IF NOT EXISTS idx_alerts_is_read ON alerts(is_read) WHERE is_read = FALSE;
CREATE INDEX IF NOT EXISTS idx_alerts_resolved ON alerts(resolved_at) WHERE resolved_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC);

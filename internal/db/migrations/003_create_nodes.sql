-- 003_create_nodes.sql
-- IoT Nodes (ESP32 devices) table

CREATE TABLE IF NOT EXISTS nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rbw_id UUID NOT NULL REFERENCES rbw(id) ON DELETE CASCADE,
    node_type VARCHAR(50) NOT NULL CHECK (node_type IN ('gateway', 'nest', 'lmb', 'pump')),
    node_code VARCHAR(100) NOT NULL,
    esp32_uid VARCHAR(50),
    status_node VARCHAR(50) DEFAULT 'offline' CHECK (status_node IN ('online', 'offline', 'error')),
    last_seen TIMESTAMPTZ,
    has_audio BOOLEAN DEFAULT FALSE,
    state_audio_lmb BOOLEAN DEFAULT FALSE,
    state_audio_nest BOOLEAN DEFAULT FALSE,
    has_pump BOOLEAN DEFAULT FALSE,
    state_pump BOOLEAN DEFAULT FALSE,
    installed_at TIMESTAMPTZ,
    uninstalled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_nodes_rbw_id ON nodes(rbw_id);
CREATE INDEX IF NOT EXISTS idx_nodes_esp32_uid ON nodes(esp32_uid);
CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status_node);

-- Unique constraint for node_code within an RBW
CREATE UNIQUE INDEX IF NOT EXISTS idx_nodes_rbw_code ON nodes(rbw_id, node_code);

-- Updated at trigger
DROP TRIGGER IF EXISTS update_nodes_updated_at ON nodes;
CREATE TRIGGER update_nodes_updated_at
    BEFORE UPDATE ON nodes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

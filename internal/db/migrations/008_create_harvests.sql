-- 008_create_harvests.sql
-- Harvests table

CREATE TABLE IF NOT EXISTS harvests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rbw_id UUID NOT NULL REFERENCES rbw(id) ON DELETE CASCADE,
    node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,
    floor_no INT NOT NULL CHECK (floor_no >= 1),
    harvested_at TIMESTAMPTZ NOT NULL,
    nests_count INT NOT NULL CHECK (nests_count >= 0),
    weight_kg DOUBLE PRECISION CHECK (weight_kg >= 0),
    grade VARCHAR(20) CHECK (grade IN ('good', 'medium', 'poor')),
    notes TEXT,
    created_by UUID REFERENCES users(id),
    cycle_days INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_harvests_rbw_id ON harvests(rbw_id);
CREATE INDEX IF NOT EXISTS idx_harvests_harvested_at ON harvests(harvested_at DESC);

-- Updated at trigger
DROP TRIGGER IF EXISTS update_harvests_updated_at ON harvests;
CREATE TRIGGER update_harvests_updated_at
    BEFORE UPDATE ON harvests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

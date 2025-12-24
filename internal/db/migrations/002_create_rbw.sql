-- 002_create_rbw.sql
-- RBW (Rumah Burung Walet) table

CREATE TABLE IF NOT EXISTS rbw (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    total_floors INT NOT NULL DEFAULT 1 CHECK (total_floors >= 1),
    description TEXT,
    photo_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_rbw_owner_id ON rbw(owner_id);
CREATE INDEX IF NOT EXISTS idx_rbw_code ON rbw(code);

-- Updated at trigger
DROP TRIGGER IF EXISTS update_rbw_updated_at ON rbw;
CREATE TRIGGER update_rbw_updated_at
    BEFORE UPDATE ON rbw
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

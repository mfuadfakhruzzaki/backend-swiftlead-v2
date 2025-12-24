-- 007_create_service_requests.sql
-- Service requests table

CREATE TABLE IF NOT EXISTS service_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rbw_id UUID NOT NULL REFERENCES rbw(id) ON DELETE CASCADE,
    node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,
    request_by UUID NOT NULL REFERENCES users(id),
    assigned_to UUID REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    type VARCHAR(50) NOT NULL CHECK (type IN ('installation', 'maintenance', 'uninstall')),
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN (
        'draft', 'pending', 'approved', 'rejected', 
        'assigned', 'in_progress', 'resolved', 'cancelled'
    )),
    request_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    schedule_date TIMESTAMPTZ,
    uninstall_date TIMESTAMPTZ,
    issue TEXT,
    resolution TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_service_requests_rbw_id ON service_requests(rbw_id);
CREATE INDEX IF NOT EXISTS idx_service_requests_request_by ON service_requests(request_by);
CREATE INDEX IF NOT EXISTS idx_service_requests_assigned_to ON service_requests(assigned_to);
CREATE INDEX IF NOT EXISTS idx_service_requests_status ON service_requests(status);

-- Updated at trigger
DROP TRIGGER IF EXISTS update_service_requests_updated_at ON service_requests;
CREATE TRIGGER update_service_requests_updated_at
    BEFORE UPDATE ON service_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

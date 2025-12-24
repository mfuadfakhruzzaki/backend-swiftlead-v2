-- 009_create_transactions.sql
-- Transaction categories and transactions tables

CREATE TABLE IF NOT EXISTS transaction_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('income', 'expense')),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default categories
INSERT INTO transaction_categories (name, type, description) VALUES
    ('Penjualan Sarang', 'income', 'Pendapatan dari penjualan sarang walet'),
    ('Listrik', 'expense', 'Biaya listrik bulanan'),
    ('Perawatan', 'expense', 'Biaya perawatan gedung'),
    ('Gaji Karyawan', 'expense', 'Gaji karyawan/penjaga'),
    ('Lain-lain', 'income', 'Pendapatan lain-lain'),
    ('Lain-lain', 'expense', 'Pengeluaran lain-lain')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rbw_id UUID NOT NULL REFERENCES rbw(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES transaction_categories(id),
    amount DOUBLE PRECISION NOT NULL CHECK (amount >= 0),
    type VARCHAR(20) NOT NULL CHECK (type IN ('income', 'expense')),
    description TEXT,
    transaction_date DATE NOT NULL,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_transactions_rbw_id ON transactions(rbw_id);
CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions(category_id);
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(transaction_date DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);

-- Updated at trigger
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;
CREATE TRIGGER update_transactions_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

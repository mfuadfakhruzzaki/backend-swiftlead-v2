-- 010_seed_admin.sql
-- Seed admin user for development
-- Password: admin123

-- This is a valid bcrypt hash for "admin123" with cost 10
INSERT INTO users (name, email, password_hash, role, phone) 
VALUES (
    'Admin', 
    'admin@swiftlead.id', 
    '$2a$10$5kBFmC8HF0BRX8tL4H.7/.mXx.r4YkZpQwBaWzPJKLNq.xD4LnWlq', 
    'admin', 
    '+6281234567890'
)
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;

-- 010_seed_admin.sql
-- Seed admin user for development
-- Password: admin123

-- This is a valid bcrypt hash for "admin123" with cost 10
INSERT INTO users (name, email, password_hash, role, phone) 
VALUES (
    'Admin', 
    'admin@swiftlead.id', 
    '$2a$10$.iMchHPdgsV0Arnssr4/ROMFaU0Se9KxmdSAXRKAZNrDwnNhJd6PK', 
    'admin', 
    '+6281234567890'
)
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;

-- Seed technician user
-- Password: technician123
INSERT INTO users (name, email, password_hash, role, phone) 
VALUES (
    'Technician', 
    'technician@swiftlead.id', 
    '$2a$10$Q.EXEgeaVFh4C3PARPLi4e1p8WrDVvkZf0iT6qOYWF5Y.rNspCPKK', 
    'technician', 
    '+6281234567891'
)
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;

-- Seed farmer user
-- Password: farmer123
INSERT INTO users (name, email, password_hash, role, phone) 
VALUES (
    'Farmer', 
    'farmer@swiftlead.id', 
    '$2a$10$uiJ7fFBHwlnxCHwwLzCbMuav4fDOJH6JLf1PtkMRtxY9U8OiMsywS', 
    'farmer', 
    '+6281234567892'
)
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;

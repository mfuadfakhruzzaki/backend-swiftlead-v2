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
    '$2a$10$eAOS2Hz1tiyVOBsSkhstje2n0pBknCVIGs/H518RPTUx3TyxSCkFG', 
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
    '$2a$10$/538nfqC6gYWHWw2dLymKeeQdMeJkltI1JyIFUdxlhecnpAeC6HPW', 
    'farmer', 
    '+6281234567892'
)
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;

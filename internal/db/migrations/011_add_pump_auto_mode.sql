-- 011_add_pump_auto_mode.sql
-- Adds pump_auto_mode flag to nodes.
-- When TRUE (default), the AI engine drives the pump automatically.
-- When FALSE, the pump is in manual mode and AI automation is suspended.

ALTER TABLE nodes
    ADD COLUMN IF NOT EXISTS pump_auto_mode BOOLEAN NOT NULL DEFAULT TRUE;

COMMENT ON COLUMN nodes.pump_auto_mode IS
    'TRUE = AI controls pump; FALSE = manual override (AI automation suspended)';

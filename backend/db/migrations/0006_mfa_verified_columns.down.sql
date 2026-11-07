-- ============================================================================
-- Migration: 0006_mfa_verified_columns.down.sql
-- ============================================================================
BEGIN;

ALTER TABLE mfa_credentials
    ADD COLUMN IF NOT EXISTS is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ;

UPDATE mfa_credentials
    SET is_enabled = TRUE, verified_at = NOW()
    WHERE is_verified = TRUE;

ALTER TABLE mfa_credentials
    DROP COLUMN IF EXISTS is_verified,
    DROP COLUMN IF EXISTS updated_at;

COMMIT;

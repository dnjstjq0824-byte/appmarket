-- ============================================================================
-- Migration: 0006_mfa_verified_columns.up.sql
-- Description: Align mfa_credentials with the IAM repository contract
--   (is_verified / updated_at). Backfills from the legacy columns, then
--   removes is_enabled / verified_at which no query references.
-- ============================================================================
BEGIN;

ALTER TABLE mfa_credentials
    ADD COLUMN IF NOT EXISTS is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE mfa_credentials
    SET is_verified = TRUE
    WHERE verified_at IS NOT NULL OR is_enabled = TRUE;

ALTER TABLE mfa_credentials
    DROP COLUMN IF EXISTS is_enabled,
    DROP COLUMN IF EXISTS verified_at;

COMMIT;

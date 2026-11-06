-- ============================================================================
-- Migration: 0005_payout_account_last4.down.sql
-- ============================================================================
BEGIN;

ALTER TABLE developer_payout_accounts
    DROP COLUMN IF EXISTS account_last4;

COMMIT;

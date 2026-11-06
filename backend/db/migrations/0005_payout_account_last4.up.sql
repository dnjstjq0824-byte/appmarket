-- ============================================================================
-- Migration: 0005_payout_account_last4.up.sql
-- Description: Masked last-4 display column for encrypted payout accounts
-- ============================================================================
BEGIN;

ALTER TABLE developer_payout_accounts
    ADD COLUMN IF NOT EXISTS account_last4 CHAR(4);

UPDATE developer_payout_accounts
    SET account_last4 = '0000'
    WHERE account_last4 IS NULL;

ALTER TABLE developer_payout_accounts
    ALTER COLUMN account_last4 SET NOT NULL;

COMMIT;

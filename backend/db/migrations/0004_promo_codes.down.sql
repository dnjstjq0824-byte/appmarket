-- ============================================================================
-- Migration: 0004_promo_codes.down.sql
-- ============================================================================
BEGIN;

DROP TABLE IF EXISTS promo_redemptions;
DROP TABLE IF EXISTS promo_codes;

COMMIT;

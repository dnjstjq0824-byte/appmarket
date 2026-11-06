-- ============================================================================
-- Migration: 0004_promo_codes.up.sql
-- Description: Redeemable promo codes per promotion campaign
-- ============================================================================
BEGIN;

CREATE TABLE promo_codes (
    id UUID PRIMARY KEY,
    promotion_id UUID NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
    code VARCHAR(40) NOT NULL UNIQUE,
    max_uses INT NOT NULL DEFAULT 1 CHECK (max_uses > 0),
    used_count INT NOT NULL DEFAULT 0 CHECK (used_count >= 0 AND used_count <= max_uses),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_promo_codes_promotion ON promo_codes (promotion_id);

CREATE TABLE promo_redemptions (
    id UUID PRIMARY KEY,
    promo_code_id UUID NOT NULL REFERENCES promo_codes(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    order_id UUID REFERENCES orders(id) ON DELETE SET NULL,
    redeemed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (promo_code_id, user_id)
);

COMMIT;

-- ============================================================================
-- Migration: 0003_seed_reference_data.up.sql
-- Description: Seed roles, commission tiers, standard categories, and system settings
-- ============================================================================
BEGIN;

INSERT INTO roles (role_code, role_scope, description, permissions) VALUES
('USER', 'PLATFORM', 'Standard authenticated end user', '["catalog:read","review:write","order:create"]'::jsonb),
('ADMIN_SUPER', 'PLATFORM', 'Super Administrator (requires Dual-Control for sensitive ops)', '["*"]'::jsonb),
('ADMIN_KYB_REVIEWER', 'PLATFORM', 'Developer KYB identity verifier', '["admin:kyb:read","admin:kyb:decide"]'::jsonb),
('ADMIN_APP_REVIEWER', 'PLATFORM', 'App metadata and functional reviewer', '["admin:review:read","admin:review:decide"]'::jsonb),
('ADMIN_SECURITY', 'PLATFORM', 'APK malware analyst and emergency kill-switch operator', '["admin:security:read","admin:security:decide","admin:killswitch:execute"]'::jsonb),
('ADMIN_MODERATOR', 'PLATFORM', 'Trust & Safety abuse report and review moderator', '["admin:moderation:read","admin:moderation:decide"]'::jsonb),
('ADMIN_SEARCH_OPS', 'PLATFORM', 'Search ranking, synonyms, and external connector operator', '["admin:search:manage","admin:connector:manage"]'::jsonb),
('ADMIN_FINANCE', 'PLATFORM', 'Refund dispute handler and settlement payout approver', '["admin:finance:read","admin:refund:decide","admin:settlement:approve"]'::jsonb),
('ADMIN_AUDITOR', 'PLATFORM', 'Read-only compliance and security auditor', '["admin:audit:read"]'::jsonb);

INSERT INTO commission_policies (policy_code, display_name, commission_bps, pg_fee_borne_by, hold_period_days, is_active) VALUES
('STANDARD_TIER', 'Standard Commercial Commission (15%)', 1500, 'PLATFORM', 14, TRUE),
('SMALL_BIZ_TIER', 'Small Business & Indie Developer Program (10%)', 1000, 'PLATFORM', 14, TRUE),
('PARTNER_ZERO_TIER', 'Strategic Open-Source Partner Tier (0%)', 0, 'DEVELOPER', 7, TRUE);

INSERT INTO country_policies (country_code, country_name, default_currency, tax_rate_bps, require_adult_verification) VALUES
('KR', 'South Korea', 'KRW', 1000, TRUE),
('US', 'United States', 'USD', 0, FALSE),
('JP', 'Japan', 'JPY', 1000, FALSE),
('DE', 'Germany', 'EUR', 1900, FALSE),
('GB', 'United Kingdom', 'GBP', 2000, FALSE),
('SA', 'Saudi Arabia', 'SAR', 1500, TRUE);

INSERT INTO categories (id, parent_id, category_group, name_i18n, display_order) VALUES
('productivity', NULL, 'APPS', '{"ko-KR":"생산성","en-US":"Productivity","ja-JP":"仕事効率化"}'::jsonb, 1),
('utilities', NULL, 'APPS', '{"ko-KR":"도구 및 유틸리티","en-US":"Tools & Utilities","ja-JP":"ツール"}'::jsonb, 2),
('communication', NULL, 'APPS', '{"ko-KR":"커뮤니케이션","en-US":"Communication","ja-JP":"通信"}'::jsonb, 3),
('finance', NULL, 'APPS', '{"ko-KR":"금융","en-US":"Finance","ja-JP":"ファイナンス"}'::jsonb, 4),
('media_video', NULL, 'APPS', '{"ko-KR":"미디어 및 동영상","en-US":"Media & Video","ja-JP":"動画プレーヤー"}'::jsonb, 5),
('games_action', NULL, 'GAMES', '{"ko-KR":"액션 게임","en-US":"Action Games","ja-JP":"アクション"}'::jsonb, 10),
('games_rpg', NULL, 'GAMES', '{"ko-KR":"롤플레잉","en-US":"Role Playing","ja-JP":"ロールプレ잉"}'::jsonb, 11);

INSERT INTO system_settings (setting_key, setting_value, description) VALUES
('treescan.integration.config', '{"enabled": false, "mode": "STANDALONE_ISOLATED", "rate_limit_rps": 2000}'::jsonb, 'Future TreeScan integration adapter toggle (Phase 10)'),
('upload.limits.config', '{"max_apk_bytes": 2147483648, "max_aab_bytes": 4294967296, "min_signature_scheme": 2}'::jsonb, 'Artifact upload size and signature scheme policy');

COMMIT;

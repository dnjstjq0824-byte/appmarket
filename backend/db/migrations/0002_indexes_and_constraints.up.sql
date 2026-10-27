-- ============================================================================
-- Migration: 0002_indexes_and_constraints.up.sql
-- Description: Performance indexes, partial indexes, GIN indexes & Ledger Trigger
-- ============================================================================
BEGIN;

-- Catalog & Distribution Lookup Indexes
CREATE INDEX idx_catalog_apps_pkg ON catalog_apps (primary_package_name);
CREATE INDEX idx_catalog_apps_category_status ON catalog_apps (primary_category_id, catalog_status, bayesian_rating DESC);
CREATE INDEX idx_app_localizations_title_trgm ON app_localizations USING GIN (title gin_trgm_ops);
CREATE INDEX idx_app_distributions_app_active ON app_distributions (app_id, is_active);
CREATE INDEX idx_app_distributions_pkg_signer ON app_distributions (package_name, signing_cert_sha256);

-- Release & Artifact Lookup Indexes
CREATE INDEX idx_app_releases_dist_track_ver ON app_releases (distribution_id, release_track, version_code DESC);
CREATE INDEX idx_release_artifacts_abis_gin ON release_artifacts USING GIN (native_abis);
CREATE INDEX idx_apk_signatures_cert_fp ON apk_signatures (cert_sha256_fingerprint);

-- Worker & Outbox Queue Polling Partial Indexes
CREATE INDEX idx_outbox_pending ON outbox_events (created_at ASC) WHERE status = 'PENDING';
CREATE INDEX idx_jobs_queued ON background_jobs (queue_name, scheduled_at ASC) WHERE status IN ('QUEUED', 'RETRY_SCHEDULED');

-- Commerce & Audit Indexes
CREATE INDEX idx_orders_user_status ON orders (user_id, order_status, created_at DESC);
CREATE INDEX idx_orders_escrow_release ON orders (escrow_release_at ASC) WHERE order_status = 'PAID_ENTITLED';
CREATE INDEX idx_ledger_entries_tx ON ledger_entries (transaction_id);
CREATE INDEX idx_ledger_entries_acct ON ledger_entries (account_code, created_at DESC);
CREATE INDEX idx_audit_logs_target ON audit_logs (target_type, target_id, created_at DESC);

-- Double-Entry Ledger Zero-Balance Enforcement Trigger Function
CREATE OR REPLACE FUNCTION verify_ledger_transaction_balance()
RETURNS TRIGGER AS $$
DECLARE
    target_tx UUID;
    net_balance BIGINT;
BEGIN
    target_tx := COALESCE(NEW.transaction_id, OLD.transaction_id);
    SELECT COALESCE(SUM(CASE WHEN direction = 'DEBIT' THEN amount_minor ELSE -amount_minor END), 0)
    INTO net_balance
    FROM ledger_entries
    WHERE transaction_id = target_tx;

    IF net_balance <> 0 THEN
        RAISE EXCEPTION 'ERR_LEDGER_UNBALANCED: Transaction % has non-zero balance %', target_tx, net_balance;
    END IF;
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER trg_verify_ledger_balance
AFTER INSERT OR UPDATE OR DELETE ON ledger_entries
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION verify_ledger_transaction_balance();

COMMIT;

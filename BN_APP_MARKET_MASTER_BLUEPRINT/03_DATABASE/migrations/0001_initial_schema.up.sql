-- ============================================================================
-- Migration: 0001_initial_schema.up.sql
-- Project: BN-APP-MARKET-GEMINI4-ARGON-MASTER-BLUEPRINT-001
-- Target DBMS: PostgreSQL 16+
-- Description: Complete commercial-grade schema covering all 19 mandatory domains (34 tables)
-- ============================================================================

BEGIN;

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ----------------------------------------------------------------------------
-- Domain 1: Users, Authentication, MFA & RBAC
-- ----------------------------------------------------------------------------
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    email_blind_index CHAR(64) NOT NULL UNIQUE,
    password_hash TEXT,
    display_name VARCHAR(100) NOT NULL,
    country_code CHAR(2) NOT NULL DEFAULT 'KR',
    preferred_locale VARCHAR(16) NOT NULL DEFAULT 'ko-KR',
    preferred_currency CHAR(3) NOT NULL DEFAULT 'KRW',
    adult_verified_at TIMESTAMPTZ,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING_EMAIL_VERIFICATION'
        CHECK (status IN ('PENDING_EMAIL_VERIFICATION', 'ACTIVE', 'LOCKED_BRUTE_FORCE', 'SUSPENDED_POLICY_VIOLATION', 'DELETED_ANONYMIZED')),
    failed_login_attempts INT NOT NULL DEFAULT 0 CHECK (failed_login_attempts >= 0),
    locked_until TIMESTAMPTZ,
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_identities (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(32) NOT NULL
        CHECK (provider IN ('GOOGLE_OIDC', 'GITHUB_OAUTH', 'APPLE_OIDC', 'TREESCAN_OIDC')),
    provider_subject VARCHAR(255) NOT NULL,
    provider_email VARCHAR(255),
    raw_claims JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_subject)
);

CREATE TABLE auth_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_family_id UUID NOT NULL,
    refresh_token_sha256 CHAR(64) NOT NULL UNIQUE,
    device_fingerprint VARCHAR(128),
    user_agent TEXT,
    client_ip INET,
    mfa_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_reason VARCHAR(64),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rotated_at TIMESTAMPTZ
);

CREATE TABLE mfa_credentials (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mfa_type VARCHAR(24) NOT NULL CHECK (mfa_type IN ('TOTP', 'WEBAUTHN', 'RECOVERY_CODES')),
    secret_ciphertext BYTEA NOT NULL,
    backup_codes_hashes JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, mfa_type)
);

CREATE TABLE roles (
    role_code VARCHAR(48) PRIMARY KEY,
    role_scope VARCHAR(24) NOT NULL CHECK (role_scope IN ('PLATFORM', 'DEVELOPER_ORG')),
    description TEXT NOT NULL,
    permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_code VARCHAR(48) NOT NULL REFERENCES roles(role_code) ON DELETE RESTRICT,
    granted_by UUID REFERENCES users(id) ON DELETE RESTRICT,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_code)
);

-- ----------------------------------------------------------------------------
-- Domain 2: Developer Organizations, KYB Verification, Members & API Keys
-- ----------------------------------------------------------------------------
CREATE TABLE developer_orgs (
    id UUID PRIMARY KEY,
    org_slug VARCHAR(80) NOT NULL UNIQUE,
    legal_name VARCHAR(200) NOT NULL,
    display_name VARCHAR(120) NOT NULL,
    org_type VARCHAR(24) NOT NULL CHECK (org_type IN ('INDIVIDUAL', 'CORPORATION', 'OPEN_SOURCE_COLLECTIVE')),
    country_code CHAR(2) NOT NULL,
    support_email VARCHAR(255) NOT NULL,
    website_url VARCHAR(500),
    verified_domain VARCHAR(255),
    verification_status VARCHAR(32) NOT NULL DEFAULT 'UNVERIFIED'
        CHECK (verification_status IN ('UNVERIFIED', 'PENDING_REVIEW', 'ACTION_REQUIRED', 'VERIFIED', 'REJECTED', 'SUSPENDED')),
    commission_tier_code VARCHAR(32) NOT NULL DEFAULT 'STANDARD_TIER',
    trust_score INT NOT NULL DEFAULT 50 CHECK (trust_score BETWEEN 0 AND 100),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE developer_verifications (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES developer_orgs(id) ON DELETE RESTRICT,
    submitted_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    registration_number_ciphertext BYTEA,
    duns_number VARCHAR(32),
    document_s3_vault_key VARCHAR(512) NOT NULL,
    dns_txt_token VARCHAR(128) NOT NULL,
    dns_verified_at TIMESTAMPTZ,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING_REVIEW'
        CHECK (status IN ('PENDING_REVIEW', 'APPROVED', 'REJECTED', 'MORE_INFO_REQUIRED')),
    reviewer_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    review_notes TEXT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE developer_org_members (
    org_id UUID NOT NULL REFERENCES developer_orgs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_role VARCHAR(32) NOT NULL
        CHECK (org_role IN ('DEV_OWNER', 'DEV_RELEASE_MGR', 'DEV_EDITOR', 'DEV_FINANCE', 'DEV_SUPPORT')),
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (org_id, user_id)
);

CREATE TABLE developer_api_keys (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES developer_orgs(id) ON DELETE CASCADE,
    key_name VARCHAR(100) NOT NULL,
    key_prefix CHAR(12) NOT NULL,
    key_sha256 CHAR(64) NOT NULL UNIQUE,
    scoped_app_id UUID,
    scopes TEXT[] NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE developer_payout_accounts (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL UNIQUE REFERENCES developer_orgs(id) ON DELETE RESTRICT,
    payout_currency CHAR(3) NOT NULL,
    bank_name VARCHAR(120) NOT NULL,
    account_holder_name VARCHAR(160) NOT NULL,
    account_number_ciphertext BYTEA NOT NULL,
    swift_or_routing_code VARCHAR(64),
    tax_id_ciphertext BYTEA,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ----------------------------------------------------------------------------
-- Domain 3: Unified App Catalog, Localization, Taxonomy & Country Policies
-- ----------------------------------------------------------------------------
CREATE TABLE categories (
    id VARCHAR(48) PRIMARY KEY,
    parent_id VARCHAR(48) REFERENCES categories(id) ON DELETE RESTRICT,
    category_group VARCHAR(16) NOT NULL CHECK (category_group IN ('APPS', 'GAMES')),
    name_i18n JSONB NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE tags (
    id VARCHAR(48) PRIMARY KEY,
    name_i18n JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE catalog_apps (
    id UUID PRIMARY KEY,
    canonical_slug VARCHAR(120) NOT NULL UNIQUE,
    primary_package_name VARCHAR(255) NOT NULL,
    owner_org_id UUID REFERENCES developer_orgs(id) ON DELETE RESTRICT,
    primary_category_id VARCHAR(48) NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    default_locale VARCHAR(16) NOT NULL DEFAULT 'ko-KR',
    monetization_type VARCHAR(24) NOT NULL DEFAULT 'FREE'
        CHECK (monetization_type IN ('FREE', 'FREE_WITH_ADS', 'FREE_WITH_IAP', 'PAID_ONCE', 'SUBSCRIPTION_READY')),
    global_age_rating VARCHAR(24) NOT NULL DEFAULT 'ALL'
        CHECK (global_age_rating IN ('ALL', 'AGE_7', 'AGE_12', 'AGE_16', 'AGE_18_ADULT')),
    privacy_policy_url VARCHAR(500) NOT NULL,
    support_url VARCHAR(500),
    source_code_url VARCHAR(500),
    license_spdx VARCHAR(64) NOT NULL DEFAULT 'PROPRIETARY',
    catalog_status VARCHAR(32) NOT NULL DEFAULT 'DRAFT'
        CHECK (catalog_status IN ('DRAFT', 'PENDING_INITIAL_REVIEW', 'PUBLISHED', 'UNLISTED', 'SUSPENDED_BY_ADMIN', 'ARCHIVED')),
    average_rating NUMERIC(3,2) NOT NULL DEFAULT 0.00 CHECK (average_rating BETWEEN 0.00 AND 5.00),
    bayesian_rating NUMERIC(4,3) NOT NULL DEFAULT 0.000,
    rating_count INT NOT NULL DEFAULT 0 CHECK (rating_count >= 0),
    total_installs BIGINT NOT NULL DEFAULT 0 CHECK (total_installs >= 0),
    version INT NOT NULL DEFAULT 1,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE app_localizations (
    id UUID PRIMARY KEY,
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE CASCADE,
    locale VARCHAR(16) NOT NULL,
    title VARCHAR(120) NOT NULL,
    short_description VARCHAR(160) NOT NULL,
    full_description_sanitized TEXT NOT NULL,
    whats_new TEXT,
    icon_s3_key VARCHAR(512) NOT NULL,
    feature_graphic_s3_key VARCHAR(512),
    promo_video_url VARCHAR(500),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (app_id, locale)
);

CREATE TABLE app_media_assets (
    id UUID PRIMARY KEY,
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE CASCADE,
    locale VARCHAR(16) NOT NULL DEFAULT 'ko-KR',
    device_form_factor VARCHAR(24) NOT NULL DEFAULT 'PHONE'
        CHECK (device_form_factor IN ('PHONE', 'TABLET_7', 'TABLET_10', 'TV', 'FOLDABLE')),
    asset_type VARCHAR(24) NOT NULL CHECK (asset_type IN ('SCREENSHOT', 'FEATURE_BANNER')),
    s3_public_key VARCHAR(512) NOT NULL,
    width_px INT NOT NULL CHECK (width_px > 0),
    height_px INT NOT NULL CHECK (height_px > 0),
    sha256_checksum CHAR(64) NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE app_tags (
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE CASCADE,
    tag_id VARCHAR(48) NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (app_id, tag_id)
);

CREATE TABLE country_policies (
    country_code CHAR(2) PRIMARY KEY,
    country_name VARCHAR(100) NOT NULL,
    default_currency CHAR(3) NOT NULL,
    tax_rate_bps INT NOT NULL DEFAULT 1000 CHECK (tax_rate_bps BETWEEN 0 AND 5000),
    require_adult_verification BOOLEAN NOT NULL DEFAULT FALSE,
    prohibited_categories TEXT[] NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE app_country_availability (
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE CASCADE,
    country_code CHAR(2) NOT NULL REFERENCES country_policies(country_code) ON DELETE RESTRICT,
    is_allowed_by_dev BOOLEAN NOT NULL DEFAULT TRUE,
    is_blocked_by_law BOOLEAN NOT NULL DEFAULT FALSE,
    block_reason VARCHAR(255),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (app_id, country_code)
);

-- ----------------------------------------------------------------------------
-- Domain 4, 5, 6, 7: Distributions, Releases, Artifacts, Signatures & External Connectors
-- ----------------------------------------------------------------------------
CREATE TABLE app_distributions (
    id UUID PRIMARY KEY,
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE RESTRICT,
    source_type VARCHAR(32) NOT NULL
        CHECK (source_type IN ('BN_HOSTED', 'FDROID_REPO', 'GOOGLE_PLAY_LINK', 'GALAXY_STORE_LINK', 'APPLE_APPSTORE_LINK', 'OFFICIAL_SITE_LINK')),
    distribution_mode VARCHAR(32) NOT NULL
        CHECK (distribution_mode IN ('DIRECT_HOSTED', 'FEDERATED_REPO_BINARY', 'SEARCH_INFO_ONLY', 'EXTERNAL_LINK_ONLY')),
    rights_level VARCHAR(40) NOT NULL
        CHECK (rights_level IN ('LEVEL_1_DIRECT_HOSTED', 'LEVEL_2_LICENSED_FEDERATED_BINARY', 'LEVEL_3_SEARCH_AND_INFO_ONLY', 'LEVEL_4_OFFICIAL_LINK_ONLY', 'LEVEL_5_RESTRICTED_PROHIBITED')),
    package_name VARCHAR(255) NOT NULL,
    signing_cert_sha256 CHAR(64),
    signature_compatibility_group VARCHAR(80),
    official_store_url VARCHAR(500),
    provenance_label VARCHAR(160) NOT NULL,
    latest_version_code BIGINT NOT NULL DEFAULT 0,
    latest_version_name VARCHAR(64),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source_type, package_name, signing_cert_sha256)
);

CREATE TABLE app_releases (
    id UUID PRIMARY KEY,
    distribution_id UUID NOT NULL REFERENCES app_distributions(id) ON DELETE RESTRICT,
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE RESTRICT,
    release_track VARCHAR(24) NOT NULL DEFAULT 'PRODUCTION'
        CHECK (release_track IN ('INTERNAL', 'BETA', 'PRODUCTION')),
    version_code BIGINT NOT NULL CHECK (version_code > 0),
    version_name VARCHAR(64) NOT NULL,
    release_notes_i18n JSONB NOT NULL DEFAULT '{}'::jsonb,
    release_status VARCHAR(36) NOT NULL DEFAULT 'DRAFT'
        CHECK (release_status IN (
            'DRAFT', 'ARTIFACT_VERIFYING', 'SECURITY_SCANNING', 'SCAN_FAILED_QUARANTINED',
            'READY_FOR_SUBMISSION', 'SUBMITTED_FOR_REVIEW', 'IN_REVIEW', 'CHANGES_REQUESTED',
            'REJECTED', 'APPROVED_PENDING_PUBLISH', 'PUBLISHED_STAGED', 'PUBLISHED_FULL',
            'HALTED', 'SUPERSEDED', 'REVOKED_BY_ADMIN'
        )),
    rollout_percentage INT NOT NULL DEFAULT 100 CHECK (rollout_percentage BETWEEN 0 AND 100),
    previous_release_id UUID REFERENCES app_releases(id) ON DELETE RESTRICT,
    created_by UUID REFERENCES users(id) ON DELETE RESTRICT,
    published_at TIMESTAMPTZ,
    halted_at TIMESTAMPTZ,
    halt_reason TEXT,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (distribution_id, release_track, version_code)
);

CREATE TABLE upload_sessions (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES developer_orgs(id) ON DELETE CASCADE,
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE CASCADE,
    artifact_type VARCHAR(16) NOT NULL CHECK (artifact_type IN ('APK', 'AAB')),
    file_name VARCHAR(255) NOT NULL,
    expected_size_bytes BIGINT NOT NULL CHECK (expected_size_bytes > 0 AND expected_size_bytes <= 4294967296),
    expected_sha256 CHAR(64) NOT NULL,
    s3_quarantine_key VARCHAR(512) NOT NULL UNIQUE,
    s3_upload_id VARCHAR(255),
    status VARCHAR(24) NOT NULL DEFAULT 'INITIATED'
        CHECK (status IN ('INITIATED', 'COMPLETED', 'ABORTED', 'EXPIRED')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE release_artifacts (
    id UUID PRIMARY KEY,
    release_id UUID NOT NULL REFERENCES app_releases(id) ON DELETE RESTRICT,
    upload_session_id UUID REFERENCES upload_sessions(id) ON DELETE RESTRICT,
    artifact_type VARCHAR(16) NOT NULL CHECK (artifact_type IN ('APK', 'AAB', 'APKS_SET')),
    abi_filter VARCHAR(32) NOT NULL DEFAULT 'UNIVERSAL'
        CHECK (abi_filter IN ('UNIVERSAL', 'arm64-v8a', 'armeabi-v7a', 'x86_64', 'x86')),
    dpi_filter VARCHAR(32) NOT NULL DEFAULT 'ANY',
    s3_quarantine_key VARCHAR(512) NOT NULL,
    s3_public_key VARCHAR(512) UNIQUE,
    file_size_bytes BIGINT NOT NULL CHECK (file_size_bytes > 0),
    sha256_checksum CHAR(64) NOT NULL UNIQUE,
    extracted_package_name VARCHAR(255) NOT NULL,
    extracted_version_code BIGINT NOT NULL,
    extracted_version_name VARCHAR(64) NOT NULL,
    min_sdk INT NOT NULL CHECK (min_sdk >= 1),
    target_sdk INT NOT NULL CHECK (target_sdk >= min_sdk),
    native_abis TEXT[] NOT NULL DEFAULT '{}',
    extracted_permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    required_features TEXT[] NOT NULL DEFAULT '{}',
    signature_scheme_version INT NOT NULL CHECK (signature_scheme_version >= 2),
    artifact_status VARCHAR(36) NOT NULL DEFAULT 'UPLOADED_QUARANTINED'
        CHECK (artifact_status IN (
            'UPLOADED_QUARANTINED', 'VERIFYING_INTEGRITY', 'PARSING_MANIFEST',
            'VERIFYING_SIGNATURE', 'SCANNING_MALWARE', 'VERIFIED_READY',
            'PUBLISHED_ON_CDN', 'REJECTED_INVALID_APK', 'REJECTED_SIGNATURE_MISMATCH',
            'QUARANTINED_MALWARE', 'REVOKED_EMERGENCY'
        )),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMPTZ
);

CREATE TABLE apk_signatures (
    id UUID PRIMARY KEY,
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE RESTRICT,
    artifact_id UUID NOT NULL REFERENCES release_artifacts(id) ON DELETE RESTRICT,
    cert_sha256_fingerprint CHAR(64) NOT NULL,
    cert_sha1_fingerprint CHAR(40) NOT NULL,
    public_key_algorithm VARCHAR(32) NOT NULL,
    key_size_bits INT NOT NULL CHECK (key_size_bits >= 2048),
    subject_dn TEXT NOT NULL,
    issuer_dn TEXT NOT NULL,
    serial_number_hex VARCHAR(128) NOT NULL,
    valid_from TIMESTAMPTZ NOT NULL,
    valid_until TIMESTAMPTZ NOT NULL,
    proof_of_rotation JSONB,
    is_primary_lineage BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE external_connectors (
    id UUID PRIMARY KEY,
    provider_code VARCHAR(48) NOT NULL UNIQUE,
    display_name VARCHAR(120) NOT NULL,
    rights_level VARCHAR(40) NOT NULL
        CHECK (rights_level IN ('LEVEL_2_LICENSED_FEDERATED_BINARY', 'LEVEL_3_SEARCH_AND_INFO_ONLY', 'LEVEL_4_OFFICIAL_LINK_ONLY', 'LEVEL_5_RESTRICTED_PROHIBITED')),
    base_url VARCHAR(500) NOT NULL,
    public_key_pem TEXT,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sync_interval_minutes INT NOT NULL DEFAULT 360 CHECK (sync_interval_minutes >= 15),
    last_sync_status VARCHAR(32) NOT NULL DEFAULT 'IDLE',
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE external_sync_jobs (
    id UUID PRIMARY KEY,
    connector_id UUID NOT NULL REFERENCES external_connectors(id) ON DELETE RESTRICT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    status VARCHAR(32) NOT NULL CHECK (status IN ('RUNNING', 'COMPLETED', 'ABORTED_SIGNATURE_INVALID', 'FAILED')),
    upstream_etag_or_hash VARCHAR(128),
    items_processed INT NOT NULL DEFAULT 0,
    items_failed INT NOT NULL DEFAULT 0,
    error_summary TEXT
);

CREATE TABLE external_store_items (
    id UUID PRIMARY KEY,
    connector_id UUID NOT NULL REFERENCES external_connectors(id) ON DELETE RESTRICT,
    external_package_id VARCHAR(255) NOT NULL,
    catalog_app_id UUID REFERENCES catalog_apps(id) ON DELETE SET NULL,
    distribution_id UUID REFERENCES app_distributions(id) ON DELETE SET NULL,
    rights_level VARCHAR(40) NOT NULL,
    official_store_url VARCHAR(500) NOT NULL,
    direct_download_url VARCHAR(500),
    signer_sha256 CHAR(64),
    latest_version_code BIGINT,
    latest_version_name VARCHAR(64),
    raw_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (connector_id, external_package_id)
);

-- ----------------------------------------------------------------------------
-- Domain 8 & 9: Install/Download Telemetry & Search Logs (Partition-Ready)
-- ----------------------------------------------------------------------------
CREATE TABLE install_events (
    id UUID NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    device_id_hash CHAR(64) NOT NULL,
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE RESTRICT,
    distribution_id UUID NOT NULL REFERENCES app_distributions(id) ON DELETE RESTRICT,
    release_id UUID REFERENCES app_releases(id) ON DELETE RESTRICT,
    event_type VARCHAR(32) NOT NULL
        CHECK (event_type IN ('DOWNLOAD_TOKEN_ISSUED', 'DOWNLOAD_COMPLETED', 'INSTALL_STARTED', 'INSTALL_SUCCEEDED', 'INSTALL_FAILED', 'UPDATE_SUCCEEDED', 'EXTERNAL_LINK_CLICKED')),
    failure_code VARCHAR(64),
    device_sdk INT,
    device_abi VARCHAR(32),
    country_code CHAR(2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE install_events_2026_10 PARTITION OF install_events
    FOR VALUES FROM ('2026-10-01 00:00:00+00') TO ('2026-11-01 00:00:00+00');
CREATE TABLE install_events_default PARTITION OF install_events DEFAULT;

CREATE TABLE search_query_logs (
    id UUID NOT NULL,
    session_hash CHAR(64) NOT NULL,
    query_text VARCHAR(120) NOT NULL,
    normalized_query VARCHAR(120) NOT NULL,
    locale VARCHAR(16) NOT NULL,
    country_code CHAR(2) NOT NULL,
    result_count INT NOT NULL DEFAULT 0,
    clicked_app_id UUID REFERENCES catalog_apps(id) ON DELETE SET NULL,
    latency_ms INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE search_query_logs_2026_10 PARTITION OF search_query_logs
    FOR VALUES FROM ('2026-10-01 00:00:00+00') TO ('2026-11-01 00:00:00+00');
CREATE TABLE search_query_logs_default PARTITION OF search_query_logs DEFAULT;

-- ----------------------------------------------------------------------------
-- Domain 10, 11, 12: User Reviews, Moderation Reports, App Reviews & Security Scans
-- ----------------------------------------------------------------------------
CREATE TABLE app_reviews_user (
    id UUID PRIMARY KEY,
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    release_id UUID REFERENCES app_releases(id) ON DELETE SET NULL,
    installed_version_code BIGINT NOT NULL,
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    review_title VARCHAR(120),
    review_body TEXT NOT NULL,
    is_verified_install BOOLEAN NOT NULL DEFAULT TRUE,
    fraud_score NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    status VARCHAR(32) NOT NULL DEFAULT 'PUBLISHED'
        CHECK (status IN ('PUBLISHED', 'QUARANTINED_AUTO_FRAUD', 'PUBLISHED_RESTORED', 'HIDDEN_POLICY_VIOLATION')),
    developer_reply_body TEXT,
    developer_replied_by UUID REFERENCES users(id) ON DELETE SET NULL,
    developer_replied_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (app_id, user_id)
);

CREATE TABLE moderation_reports (
    id UUID PRIMARY KEY,
    reporter_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    reporter_email VARCHAR(255) NOT NULL,
    target_type VARCHAR(24) NOT NULL CHECK (target_type IN ('APP', 'REVIEW', 'DEVELOPER_ORG')),
    target_id UUID NOT NULL,
    report_category VARCHAR(32) NOT NULL
        CHECK (report_category IN ('MALWARE_SECURITY', 'IMPERSONATION', 'COPYRIGHT_DMCA', 'PRIVACY_VIOLATION', 'ILLEGAL_CONTENT', 'BILLING_FRAUD', 'SPAM_ABUSE')),
    description TEXT NOT NULL,
    evidence_urls JSONB NOT NULL DEFAULT '[]'::jsonb,
    sla_priority VARCHAR(16) NOT NULL DEFAULT 'P2' CHECK (sla_priority IN ('P0_CRITICAL', 'P1_HIGH', 'P2_NORMAL')),
    status VARCHAR(32) NOT NULL DEFAULT 'OPEN'
        CHECK (status IN ('OPEN', 'IN_INVESTIGATION', 'ACTION_TAKEN', 'DISMISSED')),
    assigned_moderator_id UUID REFERENCES users(id) ON DELETE SET NULL,
    resolution_notes TEXT,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE release_reviews (
    id UUID PRIMARY KEY,
    release_id UUID NOT NULL REFERENCES app_releases(id) ON DELETE RESTRICT,
    reviewer_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    claimed_until TIMESTAMPTZ,
    checklist_results JSONB NOT NULL DEFAULT '{}'::jsonb,
    decision VARCHAR(32) NOT NULL DEFAULT 'PENDING'
        CHECK (decision IN ('PENDING', 'APPROVED', 'CHANGES_REQUESTED', 'REJECTED')),
    policy_violation_code VARCHAR(48),
    feedback_to_developer TEXT,
    internal_notes TEXT,
    decided_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE security_scan_reports (
    id UUID PRIMARY KEY,
    artifact_id UUID NOT NULL REFERENCES release_artifacts(id) ON DELETE RESTRICT,
    scanner_engine_version VARCHAR(64) NOT NULL,
    clamav_status VARCHAR(24) NOT NULL CHECK (clamav_status IN ('CLEAN', 'INFECTED', 'ERROR')),
    yara_matched_rules TEXT[] NOT NULL DEFAULT '{}',
    dangerous_permissions TEXT[] NOT NULL DEFAULT '{}',
    impersonation_similarity_score NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    risk_score INT NOT NULL CHECK (risk_score BETWEEN 0 AND 100),
    scan_verdict VARCHAR(36) NOT NULL
        CHECK (scan_verdict IN ('CLEAN_LOW_RISK', 'CLEAN_HIGH_RISK_MANUAL_REVIEW', 'INFECTED_QUARANTINED', 'SCAN_ERROR')),
    raw_report JSONB NOT NULL DEFAULT '{}'::jsonb,
    scanned_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE appeals (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES developer_orgs(id) ON DELETE RESTRICT,
    target_type VARCHAR(24) NOT NULL CHECK (target_type IN ('RELEASE_REJECTION', 'APP_SUSPENSION', 'ORG_SUSPENSION')),
    target_id UUID NOT NULL,
    initial_actor_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    appeal_statement TEXT NOT NULL,
    supporting_attachment_s3_key VARCHAR(512),
    assigned_senior_reviewer_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    status VARCHAR(32) NOT NULL DEFAULT 'SUBMITTED'
        CHECK (status IN ('SUBMITTED', 'ASSIGNED_TO_SENIOR_REVIEWER', 'UNDER_REVIEW', 'UPHELD_OVERTURNED', 'DENIED_FINAL')),
    decision_reason TEXT,
    decided_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_appeal_different_reviewer CHECK (assigned_senior_reviewer_id IS NULL OR assigned_senior_reviewer_id <> initial_actor_id)
);

-- ----------------------------------------------------------------------------
-- Domain 13 & 14: Commerce, Payments, Refunds, Double-Entry Ledger & Settlements
-- ----------------------------------------------------------------------------
CREATE TABLE commission_policies (
    policy_code VARCHAR(32) PRIMARY KEY,
    display_name VARCHAR(100) NOT NULL,
    commission_bps INT NOT NULL CHECK (commission_bps BETWEEN 0 AND 5000),
    pg_fee_borne_by VARCHAR(16) NOT NULL DEFAULT 'PLATFORM' CHECK (pg_fee_borne_by IN ('PLATFORM', 'DEVELOPER', 'SHARED')),
    hold_period_days INT NOT NULL DEFAULT 14 CHECK (hold_period_days >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE app_price_tiers (
    id UUID PRIMARY KEY,
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE CASCADE,
    country_code CHAR(2) NOT NULL REFERENCES country_policies(country_code) ON DELETE RESTRICT,
    currency CHAR(3) NOT NULL,
    list_price_minor BIGINT NOT NULL CHECK (list_price_minor >= 0),
    tax_inclusive BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (app_id, country_code)
);

CREATE TABLE promotions (
    id UUID PRIMARY KEY,
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE CASCADE,
    campaign_name VARCHAR(120) NOT NULL,
    discount_type VARCHAR(24) NOT NULL CHECK (discount_type IN ('PERCENTAGE', 'FIXED_MINOR', 'FREE_REDEEM')),
    discount_value INT NOT NULL CHECK (discount_value > 0),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    max_uses INT NOT NULL CHECK (max_uses > 0),
    used_count INT NOT NULL DEFAULT 0 CHECK (used_count >= 0 AND used_count <= max_uses),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (ends_at > starts_at)
);

CREATE TABLE orders (
    id UUID PRIMARY KEY,
    order_number VARCHAR(40) NOT NULL UNIQUE,
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE RESTRICT,
    developer_org_id UUID NOT NULL REFERENCES developer_orgs(id) ON DELETE RESTRICT,
    country_code CHAR(2) NOT NULL,
    currency CHAR(3) NOT NULL,
    list_price_minor BIGINT NOT NULL CHECK (list_price_minor >= 0),
    discount_minor BIGINT NOT NULL DEFAULT 0 CHECK (discount_minor >= 0),
    tax_minor BIGINT NOT NULL DEFAULT 0 CHECK (tax_minor >= 0),
    total_paid_minor BIGINT NOT NULL CHECK (total_paid_minor >= 0),
    net_supply_minor BIGINT NOT NULL CHECK (net_supply_minor >= 0),
    platform_fee_minor BIGINT NOT NULL CHECK (platform_fee_minor >= 0),
    developer_earning_minor BIGINT NOT NULL CHECK (developer_earning_minor >= 0),
    order_status VARCHAR(32) NOT NULL DEFAULT 'CREATED'
        CHECK (order_status IN ('CREATED', 'PAYMENT_PENDING', 'PAID_ENTITLED', 'REFUND_REQUESTED', 'REFUNDED', 'REFUND_REJECTED', 'CHARGEBACK_LOST', 'EXPIRED')),
    paid_at TIMESTAMPTZ,
    escrow_release_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (net_supply_minor + tax_minor = total_paid_minor),
    CHECK (platform_fee_minor + developer_earning_minor = net_supply_minor)
);

CREATE TABLE payments (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    pg_provider VARCHAR(32) NOT NULL CHECK (pg_provider IN ('STRIPE', 'TOSS_PAYMENTS', 'PORTONE', 'SANDBOX_MOCK')),
    pg_transaction_id VARCHAR(255) NOT NULL UNIQUE,
    payment_method_type VARCHAR(32) NOT NULL,
    currency CHAR(3) NOT NULL,
    amount_minor BIGINT NOT NULL CHECK (amount_minor >= 0),
    pg_status VARCHAR(32) NOT NULL CHECK (pg_status IN ('AUTHORIZED', 'CAPTURED', 'FAILED', 'CANCELED', 'REFUNDED')),
    raw_webhook_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    captured_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_entitlements (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE RESTRICT,
    order_id UUID REFERENCES orders(id) ON DELETE RESTRICT,
    entitlement_type VARCHAR(24) NOT NULL CHECK (entitlement_type IN ('PURCHASED', 'PROMO_REDEEMED', 'INTERNAL_TESTER')),
    status VARCHAR(24) NOT NULL DEFAULT 'ACTIVE'
        CHECK (status IN ('ACTIVE', 'REVOKED_REFUNDED', 'REVOKED_CHARGEBACK', 'EXPIRED')),
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    UNIQUE (user_id, app_id)
);

CREATE TABLE refunds (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    requested_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    reason_code VARCHAR(32) NOT NULL CHECK (reason_code IN ('ACCIDENTAL_PURCHASE', 'APP_NOT_WORKING', 'FRAUD_UNAUTHORIZED', 'ADMIN_FORCED')),
    reason_detail TEXT,
    refund_amount_minor BIGINT NOT NULL CHECK (refund_amount_minor > 0),
    had_downloaded_apk BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(24) NOT NULL DEFAULT 'REQUESTED'
        CHECK (status IN ('REQUESTED', 'APPROVED_PROCESSED', 'REJECTED')),
    reviewed_by UUID REFERENCES users(id) ON DELETE RESTRICT,
    pg_refund_id VARCHAR(255),
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ledger_accounts (
    account_code VARCHAR(96) PRIMARY KEY,
    account_type VARCHAR(24) NOT NULL CHECK (account_type IN ('ASSET', 'LIABILITY', 'REVENUE', 'EXPENSE')),
    owner_org_id UUID REFERENCES developer_orgs(id) ON DELETE RESTRICT,
    currency CHAR(3) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ledger_transactions (
    id UUID PRIMARY KEY,
    reference_type VARCHAR(32) NOT NULL CHECK (reference_type IN ('ORDER_PAYMENT', 'ORDER_REFUND', 'ESCROW_RELEASE', 'SETTLEMENT_PAYOUT', 'AD_CHARGE')),
    reference_id UUID NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    currency CHAR(3) NOT NULL,
    description TEXT NOT NULL,
    posted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES ledger_transactions(id) ON DELETE RESTRICT,
    account_code VARCHAR(96) NOT NULL REFERENCES ledger_accounts(account_code) ON DELETE RESTRICT,
    direction VARCHAR(8) NOT NULL CHECK (direction IN ('DEBIT', 'CREDIT')),
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE settlements (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES developer_orgs(id) ON DELETE RESTRICT,
    period_year_month CHAR(7) NOT NULL, -- e.g. '2026-10'
    currency CHAR(3) NOT NULL,
    gross_sales_minor BIGINT NOT NULL DEFAULT 0,
    refunds_deducted_minor BIGINT NOT NULL DEFAULT 0,
    platform_fees_minor BIGINT NOT NULL DEFAULT 0,
    withholding_tax_minor BIGINT NOT NULL DEFAULT 0,
    net_payout_minor BIGINT NOT NULL CHECK (net_payout_minor >= 0),
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT_CALCULATED'
        CHECK (status IN ('DRAFT_CALCULATED', 'PENDING_DUAL_APPROVAL', 'APPROVED_QUEUED', 'PAID_COMPLETED', 'ON_HOLD_COMPLIANCE')),
    approval_request_id UUID,
    bank_transfer_ref VARCHAR(128),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, period_year_month, currency)
);

-- ----------------------------------------------------------------------------
-- Domain 15 & 16: Sponsored Advertising & Aggregated Daily Analytics
-- ----------------------------------------------------------------------------
CREATE TABLE ad_campaigns (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES developer_orgs(id) ON DELETE RESTRICT,
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE RESTRICT,
    campaign_name VARCHAR(120) NOT NULL,
    placement_slot VARCHAR(32) NOT NULL CHECK (placement_slot IN ('SEARCH_TOP', 'CATEGORY_FEATURED', 'HOME_SPOTLIGHT')),
    target_keywords TEXT[] NOT NULL DEFAULT '{}',
    target_countries CHAR(2)[] NOT NULL DEFAULT '{}',
    currency CHAR(3) NOT NULL DEFAULT 'KRW',
    cpc_bid_minor BIGINT NOT NULL CHECK (cpc_bid_minor > 0),
    daily_budget_minor BIGINT NOT NULL CHECK (daily_budget_minor >= cpc_bid_minor),
    total_budget_minor BIGINT NOT NULL CHECK (total_budget_minor >= daily_budget_minor),
    spent_today_minor BIGINT NOT NULL DEFAULT 0 CHECK (spent_today_minor >= 0),
    spent_total_minor BIGINT NOT NULL DEFAULT 0 CHECK (spent_total_minor >= 0 AND spent_total_minor <= total_budget_minor),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    campaign_status VARCHAR(36) NOT NULL DEFAULT 'DRAFT'
        CHECK (campaign_status IN ('DRAFT', 'PENDING_APPROVAL', 'ACTIVE_SERVING', 'PAUSED_DAILY_BUDGET_CAP', 'PAUSED_BY_ADVERTISER', 'COMPLETED_TOTAL_BUDGET_EXHAUSTED', 'REJECTED_POLICY', 'SUSPENDED')),
    reviewed_by UUID REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE app_metrics_daily (
    metric_date DATE NOT NULL,
    app_id UUID NOT NULL REFERENCES catalog_apps(id) ON DELETE CASCADE,
    country_code CHAR(2) NOT NULL,
    impressions_count BIGINT NOT NULL DEFAULT 0,
    detail_views_count BIGINT NOT NULL DEFAULT 0,
    direct_downloads_count BIGINT NOT NULL DEFAULT 0,
    install_success_count BIGINT NOT NULL DEFAULT 0,
    external_link_clicks_count BIGINT NOT NULL DEFAULT 0,
    gross_revenue_minor BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (metric_date, app_id, country_code)
);

-- ----------------------------------------------------------------------------
-- Domain 17, 18, 19: Admin Dual-Control, Audit Chain, Notifications, Settings & Jobs
-- ----------------------------------------------------------------------------
CREATE TABLE dual_approval_requests (
    id UUID PRIMARY KEY,
    action_type VARCHAR(48) NOT NULL
        CHECK (action_type IN ('SETTLEMENT_PAYOUT', 'PERMANENT_APP_BAN', 'GRANT_ADMIN_ROLE', 'COMMISSION_POLICY_UPDATE', 'FORCE_LARGE_REFUND')),
    target_id VARCHAR(128) NOT NULL,
    payload JSONB NOT NULL,
    reason TEXT NOT NULL,
    initiator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    approver_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    status VARCHAR(24) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'APPROVED_EXECUTED', 'REJECTED', 'EXPIRED')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    decided_at TIMESTAMPTZ,
    CONSTRAINT chk_dual_control_distinct_actors CHECK (approver_id IS NULL OR approver_id <> initiator_id)
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    sequence_no BIGSERIAL UNIQUE,
    actor_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    actor_role VARCHAR(48) NOT NULL,
    actor_ip INET,
    action VARCHAR(64) NOT NULL,
    target_type VARCHAR(48) NOT NULL,
    target_id VARCHAR(128) NOT NULL,
    reason TEXT,
    before_state JSONB,
    after_state JSONB,
    prev_hash CHAR(64) NOT NULL,
    record_hash CHAR(64) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(32) NOT NULL CHECK (category IN ('RELEASE_REVIEW', 'SECURITY_ALERT', 'ORDER_RECEIPT', 'SETTLEMENT_NOTICE', 'SYSTEM_NOTICE')),
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    action_url VARCHAR(500),
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE system_settings (
    setting_key VARCHAR(120) PRIMARY KEY,
    setting_value JSONB NOT NULL,
    description TEXT NOT NULL,
    updated_by UUID REFERENCES users(id) ON DELETE RESTRICT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE outbox_events (
    id UUID PRIMARY KEY,
    aggregate_type VARCHAR(64) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(128) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PUBLISHED', 'FAILED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE TABLE background_jobs (
    id UUID PRIMARY KEY,
    queue_name VARCHAR(64) NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    payload JSONB NOT NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'QUEUED'
        CHECK (status IN ('QUEUED', 'RUNNING', 'SUCCEEDED', 'RETRY_SCHEDULED', 'DEAD_LETTER', 'CANCELED')),
    attempt_count INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 5,
    last_error TEXT,
    worker_id VARCHAR(64),
    heartbeat_at TIMESTAMPTZ,
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE idempotency_keys (
    key_hash CHAR(64) PRIMARY KEY,
    actor_id UUID,
    endpoint VARCHAR(128) NOT NULL,
    request_hash CHAR(64) NOT NULL,
    response_status INT NOT NULL,
    response_body JSONB NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;

# WORK_ORDER_DEPENDENCIES.md — 작업지시서 간 비순환 의존성 그래프(DAG) 및 병렬 실행 Wave 설계

- **문서 ID**: `DOC-WO-002`
- **관련 문서**: [MASTER_WORK_ORDER.md](MASTER_WORK_ORDER.md), [12_AGENT_EXECUTION/EXECUTION_SEQUENCE.md](../12_AGENT_EXECUTION/EXECUTION_SEQUENCE.md)

---

## 1. 계약 우선(Contract-First) Wave 기반 병렬 실행 토폴로지

```
[Wave 0: Phase 0 Foundation]
  AM-ARCH-001 (Lead Architect) ──► AM-OPS-001 (DevOps Agent)
       │
       ▼
[Wave 1: Phase 1 Database & Security Foundation]
  AM-DB-001 ──► AM-DB-002 (Database Agent)
       │             │
       ▼             ▼
  AM-API-001 ──► AM-SEC-001 (Security Agent) ──► AM-ARCH-002 (Backend Worker Core)
       │
       ├───────────────────────────────────────────┬────────────────────────────────────────┐
       ▼                                           ▼                                        ▼
[Wave 2: Phase 2~3 Core Catalog, Upload & Consoles (Parallel Fan-Out)]
  AM-API-002 (Catalog API)                 AM-API-003 (Developer/KYB API)          AM-DIST-001 (APK/AAB Parser & Sig)
       │                                           │                                        │
  AM-WEB-001 (User Web Core)               AM-DEV-001 (Dev Console Part 1)         AM-SEC-002 (Malware/YARA Scanner)
                                                   │                                        │
                                           AM-DEV-002 (Dev Console Part 2) ◄────────────────┤
                                           AM-ADMIN-001 ──► AM-ADMIN-002 (Admin Review/KillSwitch)
                                                                                            │
                                                                                            ▼
[Wave 3: Phase 4~5 Native Android Distribution & Unified Search (Parallel Fan-Out)]
  AM-DIST-002 (CDN & Update Engine) ──► AM-ANDROID-001 ──► AM-ANDROID-002 ──► AM-ANDROID-003 (Android Market)
  AM-DIST-003 (AAB bundletool)          AM-SEARCH-001 (OpenSearch) ──► AM-SEARCH-002 (External Connectors)
                                              │
                                              ├──► AM-API-006 (Charts & Recommendations)
                                              └──► AM-WEB-002 (User Web Search & Source Comparison)
                                                                                            │
                                                                                            ▼
[Wave 4: Phase 6~8 Security Hardening, Moderation, Commerce, Ads & Global DR (Parallel Fan-Out)]
  AM-SEC-003 (KMS & RateLimit)          AM-BILLING-001 ──► AM-BILLING-002 (Commerce, Ledger & Payouts)
  AM-MOD-001 (Reviews & Appeals)        AM-ADS-001 (Sponsored Ads & Click-Fraud)
  AM-API-004 (Geo-Fencing)              AM-WEB-003 / AM-DEV-003 / AM-ADMIN-003 (Commerce & Ads UIs)
  AM-API-005 (Notifications/Funnel)     AM-ANDROID-004 / AM-OPS-002 / AM-OPS-003 (Device Matrix & DR)
                                                                                            │
                                                                                            ▼
[Wave 5: Phase 9~10 Full E2E Verification, Independent Audit & TreeScan Readiness]
  AM-TEST-001 (Full E2E & Load Tests) ──► AM-TEST-002 (Independent Audit Gate) ──► AM-TREESCAN-001 (Phase 10 Adapter)
```

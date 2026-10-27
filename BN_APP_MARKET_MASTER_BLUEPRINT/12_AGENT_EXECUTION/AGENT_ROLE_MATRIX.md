# AGENT_ROLE_MATRIX.md — 10대 멀티 에이전트 역할, 책임 및 디렉터리 독점 소유권 매트릭스

- **문서 ID**: `DOC-EXEC-002`
- **관련 조항**: 제16조 (멀티 에이전트 운영 설계)

---

## 1. 10대 에이전트 역할 및 파일 소유권(Write Ownership) 정의

| 번호 | 에이전트 역할명 | 독점 수정 가능 디렉터리 (`Write Owner`) | 읽기 전용 참조 (`Read-Only`) | 할당된 Work Order 목록 |
| :---: | :--- | :--- | :--- | :--- |
| 1 | **Lead Architect Agent** | `backend/cmd/`, `backend/internal/platform/`, `04_API_SPECIFICATION/`, 루트 설정 | 전체 저장소 | `AM-ARCH-001` + 전체 계약 RFC 승인 |
| 2 | **Database Agent** | `backend/db/migrations/`, `backend/db/queries/`, `03_DATABASE/` | `01_A_TO_Z_BLUEPRINT/P_*` | `AM-DB-001`, `AM-DB-002` |
| 3 | **Backend Agent** | `backend/internal/modules/{iam,catalog,developer,distribution,commerce,ads,notification,analytics}/`, `backend/internal/worker/` | `backend/db/`, `04_API_SPECIFICATION/` | `AM-ARCH-002`, `AM-API-001`~`005`, `AM-DIST-001`~`003`, `AM-BILLING-001`~`002`, `AM-ADS-001` |
| 4 | **Search & Connector Agent** | `backend/internal/modules/{search,connector,treescan}/`, `backend/internal/modules/catalog/recommendations/` | `08_SEARCH_CONNECTORS/`, `15_TREESCAN_*` | `AM-SEARCH-001`, `AM-SEARCH-002`, `AM-API-006`, `AM-TREESCAN-001` |
| 5 | **Security Agent** | `backend/internal/modules/{security,moderation}/`, `backend/internal/modules/iam/rbac/`, `backend/internal/platform/{audit,crypto,ratelimit}/` | `09_SECURITY_COMPLIANCE/` | `AM-SEC-001`, `AM-SEC-002`, `AM-SEC-003`, `AM-MOD-001` |
| 6 | **Frontend Agent** | `apps/web-user/`, `apps/web-dev/`, `apps/web-admin/`, `packages/ui-kit/`, `packages/api-client/` | `04_API_SPECIFICATION/`, `05_FRONTEND_UXUI/` | `AM-WEB-001`~`003`, `AM-DEV-001`~`003`, `AM-ADMIN-001`~`003` |
| 7 | **Android Agent** | `android-client/` | `04_API_SPECIFICATION/`, `06_ANDROID/` | `AM-ANDROID-001`~`004` |
| 8 | **DevOps Agent** | `deploy/`, `scripts/`, `.github/workflows/` | `14_INFRASTRUCTURE_DEVOPS/` | `AM-OPS-001`~`003` |
| 9 | **QA / E2E Agent** | `backend/test/integration/`, `e2e/`, `loadtest/` | 전체 저장소 (수정 불가) | `AM-TEST-001` |
| 10 | **Independent Audit Agent** | `artifacts/evidence/AUDIT_REPORTS/` | 전체 저장소 및 모든 증빙 파일 (소스코드 수정 엄격 금지) | `AM-TEST-002` (최종 `VERIFIED` 승인권 독점) |

# AUTHORIZATION_MATRIX.md — 19대 API 그룹별 인증·권한(RBAC)·MFA·감사 요구사항 매트릭스

- **문서 ID**: `DOC-API-001`
- **관련 문서**: [openapi.yaml](openapi.yaml), [01_A_TO_Z_BLUEPRINT/I_IDENTITY_AND_ACCESS.md](../01_A_TO_Z_BLUEPRINT/I_IDENTITY_AND_ACCESS.md)

---

## 1. 엔드포인트별 접근 통제 및 보안 정책 매트릭스 (제9조 준수)

| API 그룹 | 엔드포인트 패턴 | 허용 역할 (Allowed Roles) | 테넌트 소유권 검증 | MFA 필수 | Rate Limit (분당) | 멱등성 키 | 감사 로그 |
| :--- | :--- | :--- | :---: | :---: | :---: | :---: | :---: |
| **01 Auth** | `POST /api/v1/auth/register`, `/login` | `PUBLIC` | N/A | Login시 조건부 | 10 / IP | 선택 | 실패/성공 기록 |
| **02 User** | `GET/PATCH /api/v1/users/me/*` | `USER` (본인) | `user.id == jwt.sub` | X | 60 / User | X | 프로필 변경 기록 |
| **03 Developer** | `POST /api/v1/developer/orgs` | `USER` (인증완료) | N/A | **O** | 5 / User | **필수** | **필수** |
| **03 Developer** | `POST /api/v1/developer/orgs/{orgId}/verification` | `DEV_OWNER` | `org_id` 소속 확인 | **O** | 10 / Org | 선택 | **필수** |
| **03 Developer** | `POST /api/v1/developer/orgs/{orgId}/apps` | `DEV_OWNER`, `DEV_RELEASE_MGR` | `org_id` 소속 확인 | **O** | 20 / Org | **필수** | **필수** |
| **04 Catalog** | `GET /api/v1/catalog/apps*`, `/charts*` | `PUBLIC` | 국가별 차단 필터 자동 적용 | X | 120 / IP | X | X |
| **05 Search** | `GET /api/v1/search/apps`, `/suggest` | `PUBLIC` | 국가별/노출상태 필터 강제 | X | 120 / IP | X | 검색어 익명 로깅 |
| **06 Upload** | `POST /api/v1/developer/apps/{appId}/uploads*` | `DEV_OWNER`, `DEV_RELEASE_MGR`, `CI_API_KEY` | `app.owner_org_id == org_id` | **O** | 30 / Org | **필수** | **필수** |
| **07 Release** | `POST .../releases/{releaseId}/submit` | `DEV_OWNER`, `DEV_RELEASE_MGR` | `app.owner_org_id == org_id` + `org.VERIFIED` | **O (Step-up)** | 20 / Org | 선택 | **필수** |
| **07 Release** | `PATCH .../releases/{releaseId}/rollout` | `DEV_OWNER`, `DEV_RELEASE_MGR` | `app.owner_org_id == org_id` | **O** | 30 / Org | 선택 | **필수** |
| **08 Dist** | `POST .../releases/{releaseId}/download-token` | `PUBLIC` (무료) / `USER` (유료 권한 보유) | 유료 앱은 `user_entitlements.ACTIVE` 검증 | X | 30 / IP·User | X | 다운로드 이벤트 기록 |
| **09 Device** | `POST /api/v1/device/check-updates` | `PUBLIC` / `USER` | N/A | X | 60 / Device | X | X |
| **10 Review** | `POST /api/v1/catalog/apps/{slug}/reviews` | `USER` | `install_events` 설치 이력 필수 검증 | X | 10 / User | 선택 | 어뷰징 점수 기록 |
| **10 Review** | `PUT /api/v1/developer/apps/{appId}/reviews/{id}/reply` | `DEV_OWNER`, `DEV_SUPPORT`, `DEV_EDITOR` | `app.owner_org_id == org_id` | X | 30 / Org | X | **필수** |
| **11 Mod** | `POST /api/v1/moderation/reports` | `PUBLIC` / `USER` | N/A | X | 10 / IP | 선택 | 신고 기록 |
| **11 Mod** | `POST /api/v1/developer/orgs/{orgId}/appeals` | `DEV_OWNER` | `org_id` 소속 확인 | **O** | 5 / Org | 선택 | **필수** |
| **12 Admin** | `POST /api/v1/admin/reviews/{releaseId}/decision` | `ADMIN_APP_REVIEWER`, `ADMIN_SECURITY`, `ADMIN_SUPER` | N/A | **O (Step-up)** | 60 / Admin | 선택 | **필수 (Diff 포함)** |
| **12 Admin** | `POST /api/v1/admin/security/kill-switch/{appId}` | `ADMIN_SECURITY`, `ADMIN_SUPER` | N/A | **O (Step-up)** | 20 / Admin | 선택 | **필수 (최고 등급)** |
| **12 Admin** | `POST /api/v1/admin/approvals/{requestId}/approve` | 대상 도메인 상위 관리자 | `approver_id != initiator_id` 강제 | **O (Step-up)** | 20 / Admin | **필수** | **필수 (Dual-Control)** |
| **13 Payment** | `POST /api/v1/commerce/orders`, `/refund` | `USER` | `order.user_id == jwt.sub` | X | 15 / User | **필수** | **필수 (원장 연동)** |
| **13 Payment** | `POST /api/v1/commerce/webhooks/{provider}` | `PG_WEBHOOK_HMAC` | HMAC-SHA256 서명 검증 | N/A | 500 / Provider | `event_id` | **필수** |
| **14 Settle** | `GET /api/v1/developer/orgs/{orgId}/settlements` | `DEV_OWNER`, `DEV_FINANCE` | `org_id` 소속 확인 | **O** | 30 / Org | X | 열람 기록 |
| **14 Settle** | `POST /api/v1/admin/settlements/{id}/approve-payout` | `ADMIN_FINANCE`, `ADMIN_SUPER` | Dual-Control 2인 승인 필수 | **O (Step-up)** | 10 / Admin | **필수** | **필수** |
| **15 Ads** | `POST /api/v1/developer/orgs/{orgId}/ads/campaigns` | `DEV_OWNER`, `DEV_FINANCE` | 자사 소유 앱(`app.owner_org_id`)만 광고 가능 | **O** | 20 / Org | **필수** | **필수** |
| **18 Conn** | `POST /api/v1/admin/connectors/{id}/trigger-sync` | `ADMIN_SEARCH_OPS`, `ADMIN_SUPER` | N/A | **O** | 10 / Admin | 선택 | **필수** |
| **19 TreeScan** | `GET /api/v1/treescan/*` | `TREESCAN_SERVICE_TOKEN` | Feature Flag 활성 시에만 동작 | N/A | 120,000 / min | X | 서비스 호출 메트릭 |

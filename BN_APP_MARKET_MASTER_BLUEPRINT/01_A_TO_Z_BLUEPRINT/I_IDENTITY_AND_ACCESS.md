# I — Identity & Access — 통합 인증(OIDC/JWT), 소셜 로그인, MFA, 세분화된 RBAC 및 감사 추적 (Identity & Access Management (IAM), OIDC/SSO, MFA, Fine-Grained RBAC & Audit Trail)

- **문서 ID**: `DOC-AZ-I`
- **알파벳 코드**: `I`
- **관련 요구사항 ID**: `REQ-I-001`, `REQ-I-002`, `REQ-I-003`
- **연관 Work Order ID**: `AM-API-001`, `AM-SEC-001`, `AM-TREESCAN-001`
- **연관 테스트 케이스 ID**: `TC-SEC-005`, `TC-SEC-006`, `TC-API-001`
- **참조 상세 설계 문서**: [04_API_SPECIFICATION/AUTHORIZATION_MATRIX.md](../04_API_SPECIFICATION/AUTHORIZATION_MATRIX.md), [15_TREESCAN_FUTURE_INTEGRATION/SSO_STRATEGY.md](../15_TREESCAN_FUTURE_INTEGRATION/SSO_STRATEGY.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 단일 플랫폼 내에서 일반 사용자, 개발자 조직 팀원, 플랫폼 운영 관리자가 공통 계정 체계를 기반으로 안전하게 인증하고, 최소 권한 원칙(Least Privilege)과 직무 분리(Segregation of Duties)를 통해 권한 상승 공격이나 내부자 부정행위를 차단하기 위해 필요하다. 또한 향후 TreeScan 통합 시 계정 마이그레이션 없이 즉시 SSO 연동이 가능하도록 표준 OIDC 기반으로 설계해야 한다.
- **달성 목표**: Argon2id 패스워드 보안, RS256/EdDSA 비대칭키 JWT + Refresh Token Rotation, OAuth2/OIDC 소셜 로그인, TOTP/WebAuthn MFA, 도메인·조직·관리자 3계층 RBAC, 그리고 불변 해시 체인 감사 로그를 구현한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **비로그인 방문자 (`GUEST`)**: 카탈로그 조회, 검색, 무료 공개 앱(성인/국가 제한 제외) 상세 열람.
- **일반 회원 (`USER`)**: 앱 다운로드/설치 이력 동기화, 유료 앱 구매, 리뷰 작성, 신고 접수.
- **개발자 조직 구성원 (`DEV_OWNER`, `DEV_RELEASE_MGR`, `DEV_EDITOR`, `DEV_FINANCE`, `DEV_SUPPORT`)**: 소속 조직 범위 내 앱 관리.
- **플랫폼 관리자 (`ADMIN_SUPER`, `ADMIN_REVIEWER`, `ADMIN_SECURITY`, `ADMIN_MODERATOR`, `ADMIN_FINANCE`, `ADMIN_SEARCH_OPS`, `ADMIN_AUDITOR`)**: 직무별로 분리된 운영 콘솔 권한 수행.

## 3. 필수 기능 (Mandatory Features)
1. **자체 독립 인증 서버 + OIDC 호환 구조**: TreeScan 없이도 자체 이메일/패스워드 가입, 이메일 인증, 비밀번호 재설정, 소셜 로그인(Google, GitHub, Apple OIDC)을 완벽히 처리한다.
2. **필수 MFA (Multi-Factor Authentication)**: 모든 관리자(`ADMIN_*`)와 프로덕션 앱을 출시/정산하는 개발자(`DEV_OWNER`, `DEV_RELEASE_MGR`, `DEV_FINANCE`)는 RFC 6238 TOTP(또는 WebAuthn Passkey) 2단계 인증을 의무 활성화해야 하며, 미설정 시 민감 API 호출이 차단(`403 ERR_MFA_REQUIRED`)된다.
3. **Refresh Token Rotation 및 재사용 탐지(Reuse Detection)**: Refresh Token은 1회 사용 시 즉시 폐기되고 새 토큰이 발급된다. 이미 사용된 이전 Refresh Token이 다시 제출되면 토큰 탈취로 간주하고 해당 세션 패밀리(`token_family_id`) 전체를 즉시 강제 로그아웃시킨다.
4. **민감 관리자 작업 이중 승인 (Dual-Control / Four-Eyes Principle)**:
   - 개발자 정산금 대규모 지급(`> $10,000`), 앱 영구 퇴출(`PERMANENT_BAN`), 관리자 권한 부여(`GRANT_ADMIN_ROLE`), 수수료율 변경은 요청자(`initiator_id`)와 승인자(`approver_id`)가 서로 다른 2명의 관리자여야만 실행된다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **암호화 및 토큰 사양**:
  - 패스워드 해싱: `Argon2id` (`memory=64MB`, `iterations=3`, `parallelism=4`, `salt=16B`, `keyLen=32B`).
  - Access Token: RS256 서명 JWT, 만료 15분(`900s`), 클레임: `sub`(user_id), `roles`(플랫폼 역할), `org_roles`(소속 개발자 조직 역할 맵), `mfa_verified`(boolean), `jti`(고유 토큰 ID).
  - Refresh Token: 256-bit CSPRNG 난수 원문은 클라이언트에 전달하고, DB(`auth_sessions`)에는 `SHA-256` 해시값만 저장(만료 14일).
- **불변 해시 체인 감사 로그 (`audit_logs`)**:
  - 각 감사 로그 레코드는 직전 레코드의 해시(`prev_hash`)와 현재 레코드 페이로드를 결합한 `record_hash = SHA256(prev_hash || actor_id || action || target_id || diff_json || created_at)`를 저장하여 사후 로그 조작 여부를 즉시 탐지할 수 있다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- 로그인/회원가입/MFA 설정/복구코드 발급 화면 (`web-user/pages/auth/*`, `web-dev/src/views/auth/*`, `web-admin/src/views/auth/*`).
- 관리자 콘솔 이중 승인 대기열 화면 (`web-admin/src/views/DualApprovalQueueView.vue`).

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/iam/`:
  - `auth_service.go`, `mfa_service.go`, `rbac_enforcer.go`, `dual_approval_service.go`.
  - [04_API_SPECIFICATION/AUTHORIZATION_MATRIX.md](../04_API_SPECIFICATION/AUTHORIZATION_MATRIX.md)에 정의된 모든 엔드포인트별 권한 규칙을 코드 테이블로 매핑하여 누락된 라우트가 있으면 서버 기동 시 패닉(Fail-Fast)시킨다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**: `users`, `user_identities`(소셜/OIDC 매핑), `auth_sessions`, `mfa_credentials`, `roles`, `user_roles`, `dual_approval_requests`, `audit_logs`.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `POST /api/v1/auth/register`, `POST /api/v1/auth/login`, `POST /api/v1/auth/mfa/verify`, `POST /api/v1/auth/token/refresh`, `POST /api/v1/auth/logout`.
- `GET /api/v1/auth/.well-known/openid-configuration` 및 `GET /api/v1/auth/jwks.json` (OIDC Discovery 및 향후 TreeScan SSO 연동용).
- `POST /api/v1/admin/approvals/{requestId}/approve`: 이중 승인 처리.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **계정 상태 (`user_status`)**: `PENDING_EMAIL_VERIFICATION` → `ACTIVE` → `LOCKED_BRUTE_FORCE` (5회 연속 로그인 실패 시 15분 잠금) / `SUSPENDED_POLICY_VIOLATION` → `DELETED_ANONYMIZED` (개인정보 파기 및 식별자 익명화).

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- OWASP ASVS v4.0 Level 2 V2(Authentication) 및 V4(Access Control) 요구사항 준수.
- 비밀번호 재설정 토큰은 15분 만료 일회용 CSPRNG 토큰으로 발급하며, 비밀번호 변경 시 기존 모든 기기의 `auth_sessions`를 즉시 무효화한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- MFA 활성화 시 10개의 일회용 복구 코드(Backup Recovery Codes)를 화면에 표시하고 사용자가 다운로드/복사 확인 체크박스를 눌러야만 설정이 완료되도록 강제한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **Refresh Token 재사용 탈취 방어 테스트 (`TC-SEC-005`)**: 한 번 사용된 Refresh Token으로 재요청 시 신규 발급된 토큰까지 포함해 해당 세션 패밀리 전체가 즉시 폐기(`401 ERR_TOKEN_REUSE_DETECTED`)되는지 검증.
- **이중 승인 자기 승인 차단 테스트 (`TC-SEC-006`)**: 요청을 생성한 관리자가 본인의 요청에 직접 `approve` 호출 시 `403 ERR_DUAL_CONTROL_SELF_APPROVAL_FORBIDDEN`으로 거부되는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- 플랫폼의 모든 API 모듈(`1`~`19`번 API 그룹)이 본 `I — Identity & Access` 모듈의 미들웨어에 의존한다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0 (최우선 필수)`
- **개발 Phase**: `Phase 1 (데이터베이스, 인증, API 기초)`에서 구현한다 (`AM-API-001`, `AM-SEC-001`).

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 외부 소셜 로그인(Google/GitHub) 장애** → 자체 이메일/패스워드 및 기존 발급된 세션 갱신은 외부 IdP와 무관하게 100% 독립 동작하며, 소셜 로그인 버튼에만 일시 지연 안내를 표시한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- IAM 백엔드 모듈 (`backend/internal/modules/iam/*`).
- 권한 매트릭스 강제 미들웨어 및 해시 체인 감사 로거.

## 17. Acceptance Criteria (수용 기준)
1. MFA 미설정 관리자 또는 개발자가 출시/심사/정산 API 호출 시 100% 차단되어야 한다.
2. `audit_logs` 테이블의 중간 레코드를 임의로 수정했을 때 감사 무결성 검증 커맨드(`worker --verify-audit-chain`)가 즉시 변조 위치를 탐지해야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-SEC-001/token_rotation_test.log`: 토큰 회전 및 탈취 감지 테스트 로그.
- `artifacts/evidence/AM-SEC-001/audit_chain_verify.log`: 해시 체인 무결성 검증 및 변조 탐지 테스트 로그.

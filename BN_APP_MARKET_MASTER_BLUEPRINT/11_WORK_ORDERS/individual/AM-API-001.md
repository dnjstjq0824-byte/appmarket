# AM-API-001 — 통합 인증(IAM), Argon2id, JWT Refresh Rotation, OIDC 및 TOTP MFA API 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-API-001` |
| **작업명 (Title)** | 통합 인증(IAM), Argon2id, JWT Refresh Rotation, OIDC 및 TOTP MFA API 구현 |
| **목표 (Objective)** | `REQ-I-001;REQ-I-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Backend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-ARCH-001;AM-DB-001` |
| **개발 Phase 및 우선순위** | `Phase 1` / `P0` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/I_IDENTITY_AND_ACCESS.md;04_API_SPECIFICATION/openapi.yaml;04_API_SPECIFICATION/AUTHORIZATION_MATRIX.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/iam/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/iam/application/auth_service.go;backend/internal/modules/iam/application/mfa_service.go;backend/internal/modules/iam/presentation/http_handler.go` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/auth/register;POST /api/v1/auth/login;POST /api/v1/auth/token/refresh;GET /api/v1/users/me` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-SEC-005;TC-API-001` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-API-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. Argon2id 패스워드 해싱 및 이메일 블라인드 인덱스 가입/로그인 구현.
2. RS256 Access JWT(15분) 및 Refresh Token Rotation(`token_family_id` 기반 재사용 탈취 감지) 구현.
3. RFC 6238 TOTP MFA 등록, 복구 코드 발급, Step-up MFA 헤더(`X-MFA-TOTP-Code`) 검증 미들웨어 구현.

## 2. Acceptance Criteria (수용 기준)
1. 회전된 구 Refresh Token 재제출 시 해당 세션 패밀리 전체가 즉시 폐기되고 `401 ERR_TOKEN_REUSE_DETECTED` 반환.
2. 5회 연속 로그인 실패 시 15분간 계정 잠금 동작.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- Redis 장애 시에도 DB `auth_sessions`를 폴백 조회하여 토큰 폐기 검증이 우회되지 않도록 Fail-Closed 처리.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-SEC-005;TC-API-001`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-API-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

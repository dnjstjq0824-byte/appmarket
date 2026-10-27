# AM-SEC-003 — OWASP ASVS/MASVS 보안 헤더, WAF Rate Limiter 정책, KMS 필드 암호화 및 PII 블라인드 인덱서 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-SEC-003` |
| **작업명 (Title)** | OWASP ASVS/MASVS 보안 헤더, WAF Rate Limiter 정책, KMS 필드 암호화 및 PII 블라인드 인덱서 구현 |
| **목표 (Objective)** | `REQ-S-001;REQ-I-001` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Security Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-SEC-001` |
| **개발 Phase 및 우선순위** | `Phase 6` / `P0` |
| **입력 문서 (Input Documents)** | `09_SECURITY_COMPLIANCE/SECURITY_ARCHITECTURE.md;02_SYSTEM_ARCHITECTURE/THREAT_MODEL.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/platform/crypto/;backend/internal/platform/ratelimit/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/platform/crypto/envelope_kms.go;backend/internal/platform/ratelimit/sliding_window_redis.go;backend/internal/platform/httpserver/security_headers.go` |
| **API 의존성 (API Dependencies)** | `ALL_APIS` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-API-001;TC-SEC-002` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-SEC-003/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. AES-256-GCM Envelope Encryption으로 정산 계좌번호, 사업자등록번호, TOTP 시크릿 암호화/복호화 구현.
2. Redis Lua 스크립트 기반 Sliding Window Rate Limiter 및 CSP/HSTS/X-Frame-Options 보안 헤더 강제 미들웨어 구현.

## 2. Acceptance Criteria (수용 기준)
1. DB 직접 덤프 시 계좌번호/MFA 시크릿이 평문으로 존재하지 않고 `BYTEA` 암호문으로만 저장됨.
2. 엔드포인트별 분당 호출 한도 초과 시 정확히 `429` 반환.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- KMS 일시 통신 실패 시 민감 정보 조회만 Fail-Closed로 거부하고 일반 카탈로그 조회는 정상 유지.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-API-001;TC-SEC-002`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-SEC-003/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

# AM-SEC-001 — 세분화된 RBAC 인가 가드, 2인 이중 승인(Dual-Control) 및 불변 해시 체인 감사 로거 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-SEC-001` |
| **작업명 (Title)** | 세분화된 RBAC 인가 가드, 2인 이중 승인(Dual-Control) 및 불변 해시 체인 감사 로거 구현 |
| **목표 (Objective)** | `REQ-I-003;REQ-S-001;REQ-W-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Security Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-API-001;AM-DB-002` |
| **개발 Phase 및 우선순위** | `Phase 1` / `P0` |
| **입력 문서 (Input Documents)** | `04_API_SPECIFICATION/AUTHORIZATION_MATRIX.md;09_SECURITY_COMPLIANCE/SECURITY_ARCHITECTURE.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/iam/rbac/;backend/internal/platform/audit/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/iam/rbac/matrix_enforcer.go;backend/internal/modules/iam/rbac/dual_control.go;backend/internal/platform/audit/hash_chain_logger.go` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/admin/approvals/{requestId}/approve` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-SEC-001;TC-SEC-006;TC-API-005` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-SEC-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. `AUTHORIZATION_MATRIX.md`의 모든 라우트별 허용 역할과 테넌트 소유권 검사를 강제하는 `MatrixEnforcer` 구현.
2. 민감 작업(`SETTLEMENT_PAYOUT`, `PERMANENT_APP_BAN`, `GRANT_ADMIN_ROLE` 등)에 대해 `initiator_id != approver_id`를 강제하는 Dual-Control 엔진 구현.
3. 직전 레코드 해시(`prev_hash`)와 연결된 `record_hash`를 계산하여 `audit_logs`에 기록하는 불변 로거 및 체인 검증 CLI 구현.

## 2. Acceptance Criteria (수용 기준)
1. 요청 생성자 본인이 이중 승인 호출 시 `403 ERR_DUAL_CONTROL_SELF_APPROVAL_FORBIDDEN` 반환.
2. 감사 체인 무결성 검증기가 임의의 로그 변조를 100% 탐지.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 감사 로그 DB 기록이 실패하면 해당 민감 비즈니스 트랜잭션 전체를 롤백하여 무감사 실행을 원천 차단.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-SEC-001;TC-SEC-006;TC-API-005`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-SEC-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

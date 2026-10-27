# AM-API-003 — 개발자 조직 온보딩, KYB 신원 검증 제출, 팀원 RBAC 및 CI/CD API Key 백엔드 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-API-003` |
| **작업명 (Title)** | 개발자 조직 온보딩, KYB 신원 검증 제출, 팀원 RBAC 및 CI/CD API Key 백엔드 구현 |
| **목표 (Objective)** | `REQ-D-001;REQ-D-002;REQ-D-003` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Backend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-API-001;AM-SEC-001` |
| **개발 Phase 및 우선순위** | `Phase 3` / `P0` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/D_DEVELOPER_PLATFORM.md;09_SECURITY_COMPLIANCE/DEVELOPER_VERIFICATION.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/developer/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/developer/application/org_service.go;backend/internal/modules/developer/application/kyb_service.go;backend/internal/modules/developer/application/apikey_service.go` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/developer/orgs;POST /api/v1/developer/orgs/{orgId}/verification;POST /api/v1/developer/orgs/{orgId}/apps;POST /api/v1/developer/orgs/{orgId}/api-keys` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-003;TC-SEC-002` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-API-003/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 개발자 조직 생성, KMS Vault 연동 KYB 서류 업로드, 역방향 DNS TXT(`_bn-verify`) 검증 워커 구현.
2. 조직 내부 5대 역할(`DEV_OWNER`, `DEV_RELEASE_MGR`, `DEV_EDITOR`, `DEV_FINANCE`, `DEV_SUPPORT`) 권한 검사 및 테넌트 IDOR 원천 차단.
3. 스코프 제한 CI/CD API 키 발급(SHA-256 해시 저장) 및 즉시 폐기 구현.

## 2. Acceptance Criteria (수용 기준)
1. 타 개발자 조직의 `orgId`/`appId`에 대한 모든 읽기/쓰기 시도가 `403 ERR_FORBIDDEN_TENANT_ACCESS`로 차단.
2. `UNVERIFIED` 조직의 프로덕션 심사 제출이 `422 ERR_DEV_ORG_NOT_VERIFIED`로 차단.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- DNS 조회 타임아웃 시 즉시 실패 처리하지 않고 백그라운드 재시도 큐에 등록 후 상태 안내.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-003;TC-SEC-002`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-API-003/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

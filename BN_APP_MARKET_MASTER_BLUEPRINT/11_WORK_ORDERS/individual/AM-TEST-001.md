# AM-TEST-001 — Testcontainers 기반 전 도메인 통합 테스트, Playwright 웹 E2E 및 k6 성능 부하 테스트 스위트 구축

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-TEST-001` |
| **작업명 (Title)** | Testcontainers 기반 전 도메인 통합 테스트, Playwright 웹 E2E 및 k6 성능 부하 테스트 스위트 구축 |
| **목표 (Objective)** | `REQ-Q-001;REQ-Q-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `QA / E2E Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-WEB-002;AM-DEV-002;AM-ADMIN-003;AM-ANDROID-003;AM-BILLING-002` |
| **개발 Phase 및 우선순위** | `Phase 9` / `P0` |
| **입력 문서 (Input Documents)** | `13_TESTING_QA/MASTER_TEST_MATRIX.csv;13_TESTING_QA/E2E_TEST_PLAN.md;13_TESTING_QA/ACCEPTANCE_CRITERIA.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/test/integration/;e2e/;loadtest/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `backend/internal/modules/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/test/integration/full_lifecycle_test.go;e2e/tests/user_market.spec.ts;e2e/tests/developer_release.spec.ts;e2e/tests/admin_review_killswitch.spec.ts;loadtest/k6_search_and_download.js` |
| **API 의존성 (API Dependencies)** | `ALL_APIS` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-001;TC-UI-001;TC-UI-002;TC-UI-003;TC-COM-004` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-TEST-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. Mock 없이 실제 PostgreSQL, Redis, OpenSearch, MinIO 컨테이너를 띄워 개발자 가입→KYB→APK 업로드→스캔→관리자 승인→사용자 검색→다운로드→설치→리뷰→유료결제→정산 전 과정을 검증하는 종단간 테스트 작성.
2. 76개 전체 화면 라우트 및 접근성 E2E 테스트, k6 부하 테스트 실행.

## 2. Acceptance Criteria (수용 기준)
1. `MASTER_TEST_MATRIX.csv`의 8대 검증군 전 항목 실행 및 통과.
2. 핵심 도메인 커버리지 85% 이상 달성.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 단 1건의 실패 발생 시에도 원인을 기록하고 담당 Work Order를 `FAILED`로 되돌려 수정 후 재검증.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-001;TC-UI-001;TC-UI-002;TC-UI-003;TC-COM-004`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-TEST-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

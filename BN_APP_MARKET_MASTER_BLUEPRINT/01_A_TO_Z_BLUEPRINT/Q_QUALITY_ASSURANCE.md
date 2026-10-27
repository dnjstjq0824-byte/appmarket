# Q — Quality Assurance — 단위·통합·계약·E2E·실기기·보안·성능 테스트 및 독립 감사 체계 (Quality Assurance, Unit/Integration/E2E/Device/Security Testing & Independent Audit Framework)

- **문서 ID**: `DOC-AZ-Q`
- **알파벳 코드**: `Q`
- **관련 요구사항 ID**: `REQ-Q-001`, `REQ-Q-002`
- **연관 Work Order ID**: `AM-TEST-001`, `AM-TEST-002`
- **연관 테스트 케이스 ID**: `TC-FUNC-001` ~ `TC-COM-006` (전체 테스트 스위트)
- **참조 상세 설계 문서**: [13_TESTING_QA/MASTER_TEST_MATRIX.csv](../13_TESTING_QA/MASTER_TEST_MATRIX.csv), [13_TESTING_QA/ACCEPTANCE_CRITERIA.md](../13_TESTING_QA/ACCEPTANCE_CRITERIA.md), [13_TESTING_QA/E2E_TEST_PLAN.md](../13_TESTING_QA/E2E_TEST_PLAN.md), [13_TESTING_QA/INDEPENDENT_AUDIT.md](../13_TESTING_QA/INDEPENDENT_AUDIT.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 제1조 1.2항 및 제18조에 따라, 실행하지 않은 테스트를 PASS로 기재하거나 Mock 응답으로 미구현 기능을 숨기는 행위를 원천 차단하고, 8대 검증군(Functional, API, Database, Android, Security, UI/UX, Operational, Commerce)을 객관적 증빙 기반으로 검증하기 위해 필요하다.
- **달성 목표**: 테스트 피라미드 기준, 도메인별 최소 코드 커버리지 목표, 8대 검증군 테스트 매트릭스, Android API 26~35 실기기/에뮬레이터 검증 프로토콜, 그리고 작성자와 분리된 Independent Audit Agent의 감사 절차를 수립한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **각 도메인 개발 에이전트**: 자신이 구현한 Work Order의 단위 테스트 및 통합 테스트 작성·실행.
- **QA / E2E Agent**: 크로스 모듈 통합 테스트, Playwright 웹 E2E, Android Instrumentation E2E, k6 부하 테스트 수행.
- **Independent Audit Agent**: 개발 에이전트가 제출한 증빙 로그의 진위 확인, 독립 클린 환경에서의 테스트 재실행, 최종 `VERIFIED` 또는 `FAILED` 판정.

## 3. 필수 기능 (Mandatory Features)
1. **8대 검증군 전수 구성 (제18조)**:
   - `Functional` (`TC-FUNC-*`), `API` (`TC-API-*`), `Database` (`TC-DB-*`), `Android` (`TC-AND-*`), `Security` (`TC-SEC-*`), `UI/UX` (`TC-UI-*`), `Operational` (`TC-OPS-*`), `Commerce` (`TC-COM-*`).
2. **Mock 사용 엄격 제한**:
   - 통합 테스트 및 E2E 테스트는 `Testcontainers`(실제 PostgreSQL 16, 실제 Redis 7, 실제 OpenSearch 2, 실제 MinIO S3 컨테이너) 위에서 구동해야 하며, DB나 스토리지를 인메모리 Mock으로 대체한 테스트는 완료 증빙으로 인정하지 않는다.
3. **커버리지 및 성능 정량 기준**:
   - 핵심 도메인(결제/정산 원장, APK 서명 검증, 권한 인가, 릴리스 상태 머신) 라인 커버리지 **85% 이상**, 전체 백엔드 평균 **75% 이상**.
   - 카탈로그/검색 API `p95 < 150ms`, 다운로드 토큰 발급 `p95 < 100ms` (k6 500 VU 부하 테스트 기준).

## 4. 상세 기능 (Detailed Functional Specifications)
- **테스트 아티팩트 표준 경로 (`artifacts/evidence/<WORK_ORDER_ID>/`)**:
  - `junit_report.xml`: 테스트 러너 표준 XML 결과.
  - `coverage.out` & `coverage.html`: Go/Vitest/JaCoCo 커버리지 리포트.
  - `playwright_trace.zip`: 웹 E2E 실행 트레이스 및 실제 렌더링 스크린샷.
  - `adb_logcat.txt`: Android `PackageInstaller` 실행 로그.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- Vitest + Vue Test Utils (컴포넌트/스토어 단위 테스트) + Playwright (크롬/모바일 웹킷/파이어폭스 E2E + `@axe-core` 접근성 검사).

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- Go 표준 `testing` + `testify` + `testcontainers-go` + `go test -v -race` (Data Race 탐지 필수 활성화).

## 7. DB 및 데이터 모델 (Database & Data Model)
- `pgTAP` 또는 Go 통합 테스트를 통해 모든 테이블의 PK/FK, `UNIQUE`, `CHECK`, 트리거 동작을 실제 PostgreSQL 인스턴스에서 검증한다.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `Spectral` 및 `oasdiff`를 통한 `openapi.yaml` 린팅, 그리고 실제 구동 중인 API 서버 응답이 OpenAPI 3.1 JSON Schema와 100% 일치하는지 검증하는 Contract Validation 미들웨어를 테스트 환경에서 강제 활성화한다.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **검증 상태 전이**: `NOT_STARTED` → `PARTIAL` → `IMPLEMENTED_UNVERIFIED` (개발 에이전트 테스트 완료) → `AUDIT_IN_PROGRESS` (감사 에이전트 재검증 중) → `VERIFIED` | `FAILED`.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- OWASP ZAP DAST 스캔, `gosec` SAST 스캔, `trivy` 컨테이너 이미지 스캔, `gitleaks` 시크릿 스캔을 CI 필수 통과 Gate로 설정한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 빈 상태(Empty), 로딩 상태(Skeleton), 에러 상태(Error Retry), 긴 다국어 문자열 오버플로, 다크모드 명도 대비를 E2E 시각적 회귀(Visual Regression)로 점검한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- 전체 개별 테스트 케이스 목록(53개 핵심 검증 항목)과 절차·예상결과·증빙경로는 [13_TESTING_QA/MASTER_TEST_MATRIX.csv](../13_TESTING_QA/MASTER_TEST_MATRIX.csv)에 정의되어 있다.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- 모든 모듈(`A`~`Z`) 및 모든 Work Order(`AM-*`)의 완료 판정 게이트이다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0`
- **개발 Phase**: Phase 0부터 CI 테스트 하네스를 구축하고 각 Phase Exit Gate마다 실행하며, `Phase 9`에서 전체 회귀 및 독립 감사를 완수한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: Flaky Test(간헐적 실패 테스트) 발생** → 임의로 `skip` 처리하는 것을 금지하며, 시간 의존성(`Clock` 인터페이스 주입)이나 포트/데이터 충돌 원인을 격리 수정할 때까지 해당 Work Order는 `VERIFIED`로 승인하지 않는다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 통합 테스트 하네스(`backend/test/integration/`), Playwright E2E 스위트(`e2e/`), Android Instrumentation 스위트(`android-client/app/src/androidTest/`), k6 부하 스크립트(`loadtest/`).

## 17. Acceptance Criteria (수용 기준)
1. [13_TESTING_QA/MASTER_TEST_MATRIX.csv](../13_TESTING_QA/MASTER_TEST_MATRIX.csv)의 모든 필수 테스트가 실제 실행 로그와 함께 통과해야 한다.
2. Independent Audit Agent의 교차 검증 리포트에 미해결 Critical/High 결함이 0건이어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-TEST-001/master_test_execution_summary.json` 및 각 Work Order별 증빙 디렉터리.

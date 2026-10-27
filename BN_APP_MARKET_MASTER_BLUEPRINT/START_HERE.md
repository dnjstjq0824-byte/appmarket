# START_HERE.md — 에이전트 작업 자동 인수 프로토콜 (Autonomous Agent Handoff Protocol)

- **문서 ID**: `DOC-ROOT-START-HERE`
- **패키지 버전**: `v1.0.0` (`BN_APP_MARKET_MASTER_BLUEPRINT_V1_0.zip`)
- **발주 명령**: `BN-APP-MARKET-GEMINI4-ARGON-MASTER-BLUEPRINT-001`
- **작성 주체**: Gemini 4 Argon (Lead System Architect & Development Orchestrator)
- **기준 시각**: `2026-10-27T23:29:45+09:00`
- **현재 프로젝트 상태**: `BLUEPRINT_COMPLETED_CODE_NOT_STARTED` (설계 완료 / 실제 애플리케이션 소스코드 미구현)

> **[경고 / 최상위 준수 원칙]**
> 본 패키지는 상용 수준의 글로벌 독립 통합 앱 마켓 플랫폼(Type 1 Independent App Store + Type 2 Federated App Search Engine + Type 3 Universal App Marketplace)을 실제로 구현하기 위한 **실행형 마스터 블루프린트 및 작업지시서 패키지**입니다.
> 본 문서를 인수한 후속 개발 에이전트(Lead Architect, Backend, Frontend, Android, DB, Search, Security, DevOps, QA, Independent Audit Agent)는 아래 **Step 1 ~ Step 10 인수 프로토콜**과 **제23조 절대 금지/필수 규칙**을 단 하나의 예외도 없이 준수해야 합니다.

---

## 1. 10단계 에이전트 작업 자동 인수 프로토콜 (Step 1 ~ Step 10)

### Step 1 — ZIP 인수 및 패키지 무결성 검증 (ZIP Handoff & Integrity Verification)
1. 전달받은 `BN_APP_MARKET_MASTER_BLUEPRINT_V1_0.zip`의 압축을 해제하고 `BN_APP_MARKET_MASTER_BLUEPRINT/` 루트 디렉터리 이하 전체 파일 및 디렉터리 목록을 `MANIFEST.json`과 대조합니다.
2. 파일 누락, 0바이트 빈 파일, 파일 손상, 문서 내부 상대 경로 링크 오류를 전수 검사합니다.
3. 루트 디렉터리에 위치한 `FILE_CHECKSUMS.sha256`을 실행(`sha256sum -c FILE_CHECKSUMS.sha256`)하여 패키지 내 모든 개별 파일의 SHA-256 해시가 정확히 일치하는지 확인합니다.
4. 만약 단 1개의 파일이라도 누락되거나 체크섬이 불일치할 경우 즉시 작업을 중단(`BLOCKED`)하고 발주자(PO)에게 보고합니다.

### Step 2 — 필수 문서 순서대로 검토 (Mandatory Reading Sequence)
개발 에이전트는 코드를 단 한 줄이라도 작성하기 전에 반드시 다음 순서로 문서를 정독해야 합니다. 필수 문서를 누락한 상태에서 개발을 시작하는 것을 엄격히 금지합니다.
1. `README.md` 및 [00_PROJECT_CHARTER/PROJECT_OVERVIEW.md](00_PROJECT_CHARTER/PROJECT_OVERVIEW.md) — 플랫폼 3대 통합 유형(Type 1/2/3)과 독립성 원칙 파악
2. [02_SYSTEM_ARCHITECTURE/SYSTEM_OVERVIEW.md](02_SYSTEM_ARCHITECTURE/SYSTEM_OVERVIEW.md) 및 [01_A_TO_Z_BLUEPRINT/A_ARCHITECTURE.md](01_A_TO_Z_BLUEPRINT/A_ARCHITECTURE.md) — 전체 시스템 아키텍처, 모듈 경계, 기술 스택 확인
3. [16_FINAL_DELIVERY/REQUIREMENTS_TRACEABILITY.csv](16_FINAL_DELIVERY/REQUIREMENTS_TRACEABILITY.csv) 및 [01_A_TO_Z_BLUEPRINT/](01_A_TO_Z_BLUEPRINT/) (`A`~`Z` 26개 문서) — 도메인별 상세 요구사항, 상태 머신, 예외 처리 확인
4. [03_DATABASE/ERD.md](03_DATABASE/ERD.md), [03_DATABASE/TABLE_DICTIONARY.md](03_DATABASE/TABLE_DICTIONARY.md), [04_API_SPECIFICATION/openapi.yaml](04_API_SPECIFICATION/openapi.yaml) — DB DDL 및 OpenAPI 3.1 계약 확인
5. [11_WORK_ORDERS/MASTER_WORK_ORDER.md](11_WORK_ORDERS/MASTER_WORK_ORDER.md) 및 [11_WORK_ORDERS/WORK_ORDER_INDEX.csv](11_WORK_ORDERS/WORK_ORDER_INDEX.csv) — 개별 작업지시서(`AM-*`)와 선행조건 확인
6. [13_TESTING_QA/ACCEPTANCE_CRITERIA.md](13_TESTING_QA/ACCEPTANCE_CRITERIA.md) 및 [13_TESTING_QA/MASTER_TEST_MATRIX.csv](13_TESTING_QA/MASTER_TEST_MATRIX.csv) — 수용 기준 및 테스트 정책 확인

### Step 3 — 기존 코드베이스 확인 (Target Repository Inspection)
1. 개발 대상 Git 저장소가 이미 존재하는 경우, 디렉터리 구조, 브랜치 상태, 기구현된 모듈과 설정 파일을 먼저 스캔합니다.
2. 기존에 정상 동작하는 구현물이나 인프라 설정을 불필요하게 삭제하거나 초기화(`rm -rf`, `git reset --hard`)하지 않습니다.
3. 기존 코드베이스가 제공되지 않은 경우, [11_WORK_ORDERS/individual/AM-ARCH-001.md](11_WORK_ORDERS/individual/AM-ARCH-001.md)에 정의된 모노레포/멀티모듈 표준 디렉터리 구조(`backend/`, `web-user/`, `web-dev/`, `web-admin/`, `android-client/`, `workers/`, `deploy/`)로 신규 프로젝트를 초기화합니다.

### Step 4 — Gap Analysis (현황 분석 및 상태 분류)
요구사항 추적표([16_FINAL_DELIVERY/REQUIREMENTS_TRACEABILITY.csv](16_FINAL_DELIVERY/REQUIREMENTS_TRACEABILITY.csv))와 Work Order 목록([11_WORK_ORDERS/WORK_ORDER_INDEX.csv](11_WORK_ORDERS/WORK_ORDER_INDEX.csv))의 각 항목에 대해 현재 코드베이스 상태를 다음 6가지 상태값 중 하나로 분류하여 초기 Gap Analysis 리포트를 생성합니다.
- `NOT_STARTED`: 아직 코드 구현이 시작되지 않은 상태 (본 블루프린트 인수 직후 모든 개발 항목의 기본 상태)
- `PARTIAL`: 일부 테이블, API, UI만 작성되었으며 명세 전체를 충족하지 못한 상태
- `IMPLEMENTED_UNVERIFIED`: 코드는 작성되었으나 자동화 테스트 및 독립 에이전트 검증이 완료되지 않은 상태
- `VERIFIED`: 단위/통합/E2E 테스트를 통과하고 실행 로그, DB 레코드, 아티팩트 증빙이 확인된 상태
- `FAILED`: 테스트 실행 또는 보안/감사 검증에서 실패가 확인된 상태
- `BLOCKED`: 선행 작업 미완료, 외부 API 자격증명 미발급, 법률/정책 미확정 등으로 진행이 불가능한 상태
> **엄격 제한**: 실제 실행 로그와 증빙 파일 없이 임의로 `VERIFIED`로 표시하는 행위를 절대 금지합니다.

### Step 5 — 작업 실행 계획 수립 (Execution Planning & Dependency Ordering)
1. [11_WORK_ORDERS/WORK_ORDER_DEPENDENCIES.md](11_WORK_ORDERS/WORK_ORDER_DEPENDENCIES.md)와 [12_AGENT_EXECUTION/EXECUTION_SEQUENCE.md](12_AGENT_EXECUTION/EXECUTION_SEQUENCE.md)에 정의된 DAG(Directed Acyclic Graph)에 따라 작업을 정렬합니다.
2. 반드시 **Database Schema (`AM-DB-*`) → Backend Core & API Contract (`AM-API-*`, `AM-DIST-*`, `AM-SEARCH-*`, `AM-SEC-*`) → Client (`AM-WEB-*`, `AM-DEV-*`, `AM-ADMIN-*`, `AM-ANDROID-*`)** 순서를 준수합니다.
3. 계약(PostgreSQL DDL, `04_API_SPECIFICATION/openapi.yaml`)이 확정된 이후에는 [12_AGENT_EXECUTION/CONFLICT_PREVENTION.md](12_AGENT_EXECUTION/CONFLICT_PREVENTION.md)의 디렉터리 소유권 규칙에 따라 멀티 에이전트 병렬 개발을 수행합니다.

### Step 6 — 개발 수행 (Work Order Execution)
1. 각 에이전트는 할당된 개별 작업지시서(`11_WORK_ORDERS/individual/AM-*.md`)의 `수정 가능 디렉터리` 내에서만 실제 동작하는 상용 코드를 작성합니다.
2. `변경 금지 디렉터리`에 속한 파일은 절대 수정하지 않습니다. 공유 계약 변경이 불가피한 경우 Lead Architect에게 RFC(Request For Change)를 제출하고 승인을 받아야 합니다.
3. **Mock 사용 제한**: Mock 객체는 단위 테스트(Unit Test) 격리 환경에서만 허용됩니다. 실제 API 핸들러, APK 서명 검증기, PackageInstaller 연동, DB 트랜잭션 경로에 `return mockSuccess()`나 하드코딩된 JSON을 넣어 미구현 기능을 숨기는 행위는 즉각 감사 실패(`FAILED`)로 처리합니다.

### Step 7 — 기능 검증 (Functional & Automated Verification)
1. 각 Work Order 구현 직후 해당 문서의 `필수 테스트` 및 `Acceptance Criteria`에 명시된 테스트(Go `go test -v -race`, Playwright E2E, Android Instrumentation Test, OpenAPI 스키마 검증 등)를 실제로 실행합니다.
2. 실행 결과(표준 출력, 에러 로그, 커버리지 리포트, 캡처 증빙)를 `artifacts/evidence/<WORK_ORDER_ID>/` 경로에 저장합니다.
3. 실패 발생 시 원인을 분석하여 코드를 수정하고 재검증합니다.

### Step 8 — 교차 검증 및 독립 감사 (Cross-Verification & Independent Audit)
1. 코드를 작성한 담당 에이전트(예: Backend Agent)가 스스로 `VERIFIED`를 최종 선언할 수 없습니다.
2. 반드시 `QA / E2E Agent`와 `Independent Audit Agent`가 [13_TESTING_QA/INDEPENDENT_AUDIT.md](13_TESTING_QA/INDEPENDENT_AUDIT.md) 절차에 따라 독립된 환경에서 테스트를 재실행하고 증빙 파일을 대조한 뒤에만 최종 `VERIFIED`로 승격됩니다.

### Step 9 — Git 및 배포 관리 (Git Workflow & Safe Deployment)
1. 허가된 Git 저장소의 기능 브랜치(`feat/AM-XXX-001`)에서만 커밋 및 푸시를 수행합니다.
2. 커밋 메시지에는 반드시 Work Order ID와 Requirement ID를 명시합니다(예: `feat(dist): implement APK v2/v3 signature verifier [AM-DIST-001][REQ-N-001]`).
3. `.env`, 개인키(`.pem`, `.jks`, `.keystore`), API Secret을 Git에 커밋하는 것을 절대 금지하며, pre-commit secret scanner를 통과해야 합니다.
4. 프로덕션 배포 및 파괴적 DB 마이그레이션은 반드시 PO 및 Lead Architect의 수동 승인 Gate를 거칩니다.

### Step 10 — 완료 보고 (Completion Reporting)
1. 각 작업 및 Phase 종료 시 [12_AGENT_EXECUTION/REPORTING_TEMPLATE.md](12_AGENT_EXECUTION/REPORTING_TEMPLATE.md) 양식에 맞추어 다음 항목을 누락 없이 보고합니다:
   - 완료된 Work Order ID 목록 및 상태
   - 변경된 파일 목록 (`git diff --stat`)
   - 실제 실행된 테스트 명령어와 전체 로그 경로
   - 실제 Git Commit SHA
   - 빌드/배포 아티팩트 해시 (SHA-256)
   - 미완료(`PARTIAL`, `BLOCKED`, `FAILED`) 항목 및 구체적 차단 사유
2. **전체 완료 선언 금지**: 단 하나의 필수 Work Order나 Acceptance Criteria라도 미검증 상태라면 프로젝트 전체 완료를 선언할 수 없습니다.

---

## 2. 후속 개발 에이전트 절대 준수 규칙 (제23조 명문화)

### 2.1 절대 금지 사항 (Strictly Prohibited Actions)
1. **구현하지 않은 기능 완료 선언 금지**: 스켈레톤 코드나 주석(`// TODO`)만 작성하고 `IMPLEMENTED` 또는 `VERIFIED`로 보고하는 행위.
2. **실행하지 않은 테스트 PASS 보고 금지**: 테스트 도구를 실제로 실행하지 않고 결과 표에 `PASS`를 기재하는 행위.
3. **가짜 스크린샷 생성 금지**: 실제 렌더링된 웹 브라우저/Android 기기 화면이 아닌 합성 이미지나 정적 목업을 완료 증빙으로 제출하는 행위.
4. **가짜 Android 설치 결과 보고 금지**: `android.content.pm.PackageInstaller.Session` API를 실제로 호출하지 않고 토스트 메시지나 UI 상태만 `"설치 완료"`로 바꾸는 행위.
5. **Mock API를 실서버 API로 보고 금지**: DB나 스토리지와 연결되지 않은 인메모리 더미 응답을 상용 API 완료로 보고하는 행위.
6. **존재하지 않는 Git SHA 보고 금지**: `git rev-parse HEAD`로 확인되지 않는 임의의 16진수 문자열을 커밋 해시로 기재하는 행위.
7. **실제 근거 없는 성능 수치 기재 금지**: k6/Locust 등 부하 테스트 도구 실행 로그 없이 `"초당 10,000건 처리 확인"` 등을 보고하는 행위.
8. **명세와 다른 DB 임의 변경 금지**: [03_DATABASE/migrations/](03_DATABASE/migrations/)에 정의된 테이블명, 컬럼명, ENUM 값, 제약조건을 아키텍트 승인 없이 임의 축소·변경하는 행위.
9. **승인 없는 TreeScan 본체 수정 금지**: 본 프로젝트는 독립 플랫폼(Phase 0~9)이며, TreeScan 본체 저장소를 수정하거나 직접 결합하는 행위를 금지함(Phase 10은 어댑터 준비만 수행).
10. **승인 없는 프로덕션 배포 금지**: Staging 검증 및 운영자 승인 없이 프로덕션 클러스터에 배포하는 행위.
11. **다른 개발자(에이전트)의 작업 임의 삭제 금지**: 타 에이전트 소유 디렉터리의 코드를 임의로 덮어쓰거나 삭제하는 행위.
12. **비밀키 및 계정정보 저장소 커밋 금지**: DB 비밀번호, JWT 서명키, S3 자격증명, Android Keystore 파일 등을 Git에 커밋하는 행위.
13. **타사 APK 무단 재배포 금지**: Google Play 등 외부 독점 스토어의 APK 파일을 무단 크롤링·추출하여 자체 CDN에서 재배포하는 행위 ([08_SEARCH_CONNECTORS/PROVENANCE_AND_RIGHTS.md](08_SEARCH_CONNECTORS/PROVENANCE_AND_RIGHTS.md) 참조).

### 2.2 필수 이행 사항 (Mandatory Requirements)
1. **실제 소스코드 생성**: 지정 기술 스택(Go, PostgreSQL, OpenSearch, Redis, Vue 3/Nuxt/TS, Kotlin/Jetpack Compose)으로 컴파일·실행 가능한 상용 코드 작성.
2. **실제 테스트 실행 및 실패 로그 기록**: 성공 로그뿐 아니라 실패 시의 스택트레이스와 재검증 이력을 모두 보존.
3. **API 계약 준수**: [04_API_SPECIFICATION/openapi.yaml](04_API_SPECIFICATION/openapi.yaml) 및 [04_API_SPECIFICATION/API_ERROR_CONTRACT.md](04_API_SPECIFICATION/API_ERROR_CONTRACT.md)와 100% 일치하는 입출력 구현.
4. **마이그레이션 관리**: 모든 스키마 변경은 Up/Down SQL 마이그레이션 파일과 체크섬으로 관리.
5. **실제 Android 설치·업데이트 검증**: Android API 26~35 에뮬레이터/실기기에서 `PackageInstaller` 세션 커밋, 사용자 승인 인텐트, 서명 불일치 차단, 버전 롤백 방지를 실제 검증.
6. **앱 보안 검증**: 업로드된 APK의 ZIP Slip 검사, `apksigner` v2/v3 검증, YARA/ClamAV/정적 권한 분석 파이프라인 실제 구동.
7. **독립 감사(Independent Audit)**: 작성자와 분리된 감사 에이전트에 의한 교차 검증 통과.
8. **문서와 코드 동기화 및 미완료 작업 명시**: 불가피한 차단 요소 발생 시 `BLOCKED` 상태와 구체적 사유를 명시하여 다음 이터레이션으로 투명하게 이월.

---

## 3. 최초 실행 명령어 가이드 (Quick Start for Receiving Agent)

패키지를 인수한 Lead Architect Agent는 작업 시작 전 아래 검증 스크립트를 먼저 실행하여 블루프린트 무결성을 확인하십시오:

```bash
# 1. 체크섬 전체 검증
sha256sum -c FILE_CHECKSUMS.sha256

# 2. 필수 Work Order 착수 순서 확인 (Phase 0 -> Phase 1)
cat 11_WORK_ORDERS/WORK_ORDER_INDEX.csv
```

검증이 완료되면 **Phase 0의 첫 번째 작업지시서인 [11_WORK_ORDERS/individual/AM-ARCH-001.md](11_WORK_ORDERS/individual/AM-ARCH-001.md)**부터 실행을 시작하십시오.

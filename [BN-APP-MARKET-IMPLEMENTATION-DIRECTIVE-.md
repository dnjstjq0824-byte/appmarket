[BN-APP-MARKET-IMPLEMENTATION-DIRECTIVE-001]
수신: Lead Development & Orchestration Agent
첨부 패키지: BN_APP_MARKET_MASTER_BLUEPRINT_V1_0.zip (Expected SHA-256: fd8111b8d631c533c276857a21fcca1a48972f890dbbb5b97fa21759206abe08)
목표: 첨부된 마스터 블루프린트 ZIP 패키지를 인수하여 글로벌 독립 통합 앱 마켓 플랫폼(Type 1 Independent App Store + Type 2 Federated App Search Engine + Type 3 Universal App Marketplace)을 실제 상용 코드로 구현하고 검증하라.

[필수 수행 절차 — 절대 건너뛰지 말 것]
1. Step 1 (ZIP 인수 및 무결성 검증):
   - `BN_APP_MARKET_MASTER_BLUEPRINT_V1_0.zip`의 SHA-256 해시(`fd8111b8d631c533c276857a21fcca1a48972f890dbbb5b97fa21759206abe08`)를 먼저 확인하고 압축을 해제하라.
   - 압축 해제된 `BN_APP_MARKET_MASTER_BLUEPRINT/` 루트에서 `sha256sum -c FILE_CHECKSUMS.sha256`을 실행하여 140개 내부 파일이 단 1개의 오류도 없이 `OK`인지 검증하라.

2. Step 2 (필수 문서 순서대로 정독):
   - 코드를 단 한 줄이라도 작성하기 전에 반드시 다음 순서로 문서를 읽어라:
     (1) `START_HERE.md` 및 `README.md`
     (2) `00_PROJECT_CHARTER/PROJECT_OVERVIEW.md`, `DEVELOPMENT_PRINCIPLES.md`
     (3) `02_SYSTEM_ARCHITECTURE/SYSTEM_OVERVIEW.md`
     (4) `03_DATABASE/ERD.md`, `03_DATABASE/migrations/0001_initial_schema.up.sql` ~ `0003_seed_reference_data.up.sql`
     (5) `04_API_SPECIFICATION/openapi.yaml`, `AUTHORIZATION_MATRIX.md`, `API_ERROR_CONTRACT.md`
     (6) `11_WORK_ORDERS/MASTER_WORK_ORDER.md`, `WORK_ORDER_INDEX.csv`, `WORK_ORDER_DEPENDENCIES.md`
     (7) `12_AGENT_EXECUTION/EXECUTION_SEQUENCE.md`, `CONFLICT_PREVENTION.md`, `REPORTING_TEMPLATE.md`
     (8) `13_TESTING_QA/ACCEPTANCE_CRITERIA.md`, `MASTER_TEST_MATRIX.csv`, `INDEPENDENT_AUDIT.md`

3. Step 3 & Step 4 (기존 저장소 확인 및 초기 Gap Analysis 보고):
   - 현재 작업 저장소 상태를 스캔하고, 전체 41개 Work Order(`AM-*`)와 66개 요구사항(`REQ-*`)의 초기 상태를 분류(`NOT_STARTED`, `PARTIAL`, `IMPLEMENTED_UNVERIFIED`, `VERIFIED`, `FAILED`, `BLOCKED`)하여 보고하라.

4. Step 5 & Step 6 (Wave 기반 단계별 실제 코드 구현):
   - `12_AGENT_EXECUTION/EXECUTION_SEQUENCE.md`에 정의된 순서에 따라 아래 Wave 순서로 개발을 진행하라:
     * Wave 0 (기반 구축): `AM-ARCH-001` (Go 1.23 모듈형 모놀리스 + pnpm 워크스페이스 초기화) → `AM-OPS-001` (Docker 격리 컨테이너 및 CI/OTel 설정)
     * Wave 1 (DB & IAM/보안 코어): `AM-DB-001` → `AM-DB-002` (PostgreSQL 16 34개 테이블, 원장 대차평형 트리거, 파티셔닝) → `AM-API-001` (Argon2id/JWT 회전/TOTP MFA) → `AM-SEC-001` (RBAC, 2인 이중승인, 해시 체인 감사 로그) & `AM-ARCH-002` (Outbox & Redis Streams 6대 워커 큐)
     * Wave 2 (카탈로그, APK 검증/스캔 & 3대 웹 콘솔): `AM-API-002`, `AM-API-003`, `AM-DIST-001`, `AM-SEC-002`, `AM-WEB-001`, `AM-DEV-001`, `AM-DEV-002`, `AM-ADMIN-001`, `AM-ADMIN-002`
     * Wave 3 (CDN 배포, Android 클라이언트 & OpenSearch/외부 커넥터): `AM-DIST-002`, `AM-DIST-003`, `AM-ANDROID-001`~`003`, `AM-SEARCH-001`, `AM-SEARCH-002`, `AM-API-006`, `AM-WEB-002`
     * Wave 4 (상거래 원장, 정산, 스폰서 광고, 글로벌 차단 & DR): `AM-SEC-003`, `AM-MOD-001`, `AM-BILLING-001`~`002`, `AM-ADS-001`, `AM-API-004`~`005`, `AM-WEB-003`, `AM-DEV-003`, `AM-ADMIN-003`, `AM-ANDROID-004`, `AM-OPS-002`~`003`
     * Wave 5 (종단간 E2E 검증, 독립 감사 & Phase 10 TreeScan 어댑터): `AM-TEST-001` → `AM-TEST-002` → `AM-TREESCAN-001`

5. Step 7 ~ Step 10 (실제 테스트 실행 및 증빙 기반 보고):
   - 각 Work Order를 완료할 때마다 `11_WORK_ORDERS/individual/AM-XXX.md`에 명시된 `필수 테스트 ID`(`TC-*`)를 실제 컨테이너(PostgreSQL, Redis, OpenSearch, MinIO) 및 에뮬레이터에서 실행하라.
   - 실행 로그와 커버리지 리포트를 `artifacts/evidence/<WORK_ORDER_ID>/`에 저장하고 `12_AGENT_EXECUTION/REPORTING_TEMPLATE.md` 양식으로 보고하라.

[절대 금지 규칙 — 위반 시 즉시 작업 실패(FAILED) 처리]
- 지정 기술 스택(Go, PostgreSQL 16, OpenSearch 2.x, Redis 7.x, Vue 3/Nuxt 3/TypeScript, Kotlin/Jetpack Compose)을 승인 없이 임의 변경하지 마라.
- 실제 운영 경로(API 핸들러, DB 트랜잭션, APK 서명 검증, Android PackageInstaller, 복식부기 원장)를 Mock이나 하드코딩된 더미 JSON으로 숨기지 마라.
- 실행하지 않은 테스트를 PASS로 기재하거나, 존재하지 않는 Git Commit SHA를 만들어 보고하지 마라.
- 외부 링크 전용(`EXTERNAL_LINK_ONLY`) 앱에 가짜 직접 다운로드 버튼을 만들지 마라.
- 타사 독점 스토어(Google Play 등)의 비공개 API 해킹이나 APK 무단 복제 재배포 코드를 작성하지 마라.
- TreeScan 본체 저장소를 수정하거나 의존성을 연결하지 마라(TreeScan 없이 100% 독립 동작 필수).

지금 즉시 Step 1(패키지 체크섬 검증)과 Step 4(초기 실행 계획 및 Wave 0 `AM-ARCH-001`, `AM-OPS-001` 착수)부터 시작하라.
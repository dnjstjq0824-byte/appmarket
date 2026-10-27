# Booking Network Global Unified App Market Platform — Master Blueprint Package (v1.0.0)

- **문서 식별자**: `BN-APP-MARKET-GEMINI4-ARGON-MASTER-BLUEPRINT-001`
- **패키지 파일명**: `BN_APP_MARKET_MASTER_BLUEPRINT_V1_0.zip`
- **대상 시스템**: 글로벌 독립 통합 앱 마켓 플랫폼 (Independent Store + Federated App Search Engine + Universal App Marketplace)
- **작성 등급**: `MASTER ARCHITECTURE / COMMERCIAL-GRADE DEVELOPMENT BLUEPRINT`
- **현재 구현 상태**: `DESIGN_VERIFIED / IMPLEMENTATION_NOT_STARTED` (제1조 1.2항에 의거, 설계 완료를 개발 완료로 허위 보고하지 않음)

---

## 1. 프로젝트 핵심 정의 (Article 0 & Article 1)

본 설계 패키지는 Booking Network가 발주한 **글로벌 독립 통합 앱 마켓 플랫폼**을 실제 개발 에이전트들이 즉시 구현할 수 있도록 A부터 Z까지의 기술 명세, 데이터베이스 DDL, OpenAPI 3.1 명세, UI/UX 화면 설계, Android 네이티브 클라이언트 설계, APK/AAB 배포 파이프라인, 외부 앱스토어 커넥터 규격, 개별 실행형 작업지시서(Work Orders), 그리고 교차 검증 테스트 매트릭스를 하나로 묶은 마스터 블루프린트입니다.

### 1.1 단일 플랫폼 내 3대 유형 통합 아키텍처
본 플랫폼은 3개의 별도 중복 시스템을 구축하지 않고, **단일 통합 플랫폼(공통 계정 + 공통 앱 카탈로그 + 통합 OpenSearch 검색엔진 + 배포처별 설치 라우팅 엔진)** 위에서 다음 3가지 역할을 동시에 수행합니다:

| 통합 유형 | 명칭 | 핵심 역할 및 배포 권한 |
| :--- | :--- | :--- |
| **Type 1** | **Independent App Store** | 개발자가 직접 APK/AAB를 등록하고, 자동화된 보안 스캐너와 운영자 심사를 거쳐, 자체 스토리지/CDN 및 Android 클라이언트(`PackageInstaller`)를 통해 직접 다운로드·설치·자동 업데이트를 제공하는 독립 앱스토어 (`distribution_mode = DIRECT_HOSTED`) |
| **Type 2** | **Federated App Search Engine** | Google Play, F-Droid, Samsung Galaxy Store 등 외부 앱 배포처의 합법적 메타데이터/피드를 수집·분류·정규화하고, 출처(Provenance)와 동기화 시각을 명시하여 공식 배포처로 안전하게 연결하는 앱 검색엔진 (`distribution_mode = EXTERNAL_LINK_ONLY` 또는 `FEDERATED_REPO`) |
| **Type 3** | **Universal App Marketplace** | 동일 앱(Canonical App Identity)에 대해 자체 스토어와 외부 스토어들의 배포처별 가격, 최신 버전, 요구 Android SDK, 서명 인증서 지문, 라이선스, 권한 정보를 한눈에 비교·탐색하고 최적의 설치 경로를 안내하는 종합 앱 마켓 |

### 1.2 TreeScan 통합 전 완전 독립 운영 원칙 (Article 1.1)
본 플랫폼은 향후 Booking Network의 **TreeScan 검색엔진**에 통합될 예정이나, 현재 단계(Phase 0 ~ Phase 9)에서는 **TreeScan 코드베이스에 대한 어떠한 의존성이나 수정 없이 100% 독립적으로 동작**해야 합니다.
- 독립 실행 가능한 사용자 웹(`web-user`), 개발자 콘솔(`web-dev`), 관리자 콘솔(`web-admin`), Android 앱(`android-client`), 백엔드 API(`backend`), PostgreSQL DB, OpenSearch 클러스터, Redis, S3/CDN 배포망, OIDC 인증/권한 시스템, CI/CD 및 OpenTelemetry 모니터링을 갖춥니다.
- 향후 TreeScan 연동을 위한 검색 어댑터(`Search API Adapter`), 계정 연동(`OIDC/SSO Federation`), 딥링크 규격은 [15_TREESCAN_FUTURE_INTEGRATION/](15_TREESCAN_FUTURE_INTEGRATION/)에 격리 설계되어 있으며, Phase 10에서 별도 PO 승인 후 활성화합니다.

---

## 2. 지정 기술 스택 요약 (Article 2)

본 패키지의 모든 아키텍처, DDL, API, Work Order는 제2조에서 지정한 기술 스택을 엄격히 준수합니다:

| 영역 | 지정 기술 스택 | 적용 버전 및 아키텍처 결정 |
| :--- | :--- | :--- |
| **사용자 웹 (`web-user`)** | Vue 3 + Nuxt 3 + TypeScript | SSR/ISR 하이브리드 렌더링(SEO 및 초기 로딩 최적화), Pinia, TailwindCSS 디자인 토큰 |
| **개발자 콘솔 (`web-dev`)** | Vue 3 + TypeScript (Vite SPA) | 대용량 APK 멀티파트 업로드(TUS/S3 Presigned), 릴리스 상태 관리, 정산 차트 |
| **관리자 콘솔 (`web-admin`)** | Vue 3 + TypeScript (Vite SPA) | RBAC + 민감 작업 이중 승인(Dual Control), 실시간 심사 큐, 감사 로그 뷰어 |
| **Android 클라이언트 (`android-client`)** | Kotlin + Jetpack Compose | `minSdk 26` (Android 8.0) ~ `targetSdk 35` (Android 15), Coroutines/Flow, Hilt, Room, OkHttp, `PackageInstaller.Session` |
| **Backend API (`backend`)** | Go (Golang 1.23+) | 모듈형 모놀리스(Modular Monolith) + 클린 아키텍처, Chi/Echo 라우터, sqlc/pgx, OpenTelemetry |
| **Main Database** | PostgreSQL 16+ | 엄격한 PK/FK, CHECK 제약조건, 파티셔닝(로그/이벤트), 트랜잭셔널 Outbox 패턴 |
| **Search Engine** | OpenSearch 2.17+ | BM25 + 필드 가중치 + 품질/인기도 Function Score, 한국어(Nori)/다국어 분석기, Edge N-gram 자동완성 |
| **Cache & Rate Limit** | Redis 7.4+ | 분산 락(Redlock), API 속도 제한(Sliding Window), 세션/토큰 블랙리스트, 카탈로그 캐시 |
| **Background Jobs (`workers`)** | Message Queue + Isolated Workers | Redis Streams / NATS JetStream + 격리된 컨테이너 워커(APK 파서, 서명 검증기, YARA/ClamAV 악성코드 스캐너, bundletool 변환기, 외부 커넥터 동기화기) |
| **Storage & Delivery** | S3-compatible Object Storage + CDN | 격리 버킷(`quarantine-bucket`)과 공개 배포 버킷(`public-dist-bucket`) 물리적 분리, CloudFront/Cloudflare Signed URL + Immutable Cache(`sha256` 경로) |
| **API Documentation** | OpenAPI 3.1.0 | [04_API_SPECIFICATION/openapi.yaml](04_API_SPECIFICATION/openapi.yaml) |
| **Infra / CI/CD / Observability** | Docker + Git CI/CD + OpenTelemetry | 멀티스테이지 컨테이너, GitHub Actions/GitLab CI, OTel Collector + Prometheus + Grafana + Loki + Tempo |
| **AI/ML (선택적 보조)** | Python 3.12+ (격리 워커) | 허위 리뷰/어뷰징 이상 탐지 및 유사 앱 클러스터링 배치 워커 (핵심 트랜잭션 경로와 분리) |

---

## 3. 패키지 전체 디렉터리 구조 및 문서 네비게이션 (Article 19)

- [START_HERE.md](START_HERE.md) — **후속 개발 에이전트가 가장 먼저 읽어야 할 자동 인수 프로토콜 (Step 1~10)**
- `MANIFEST.json` — 전체 파일 메타데이터, 문서 분류, 크기, 역할 매핑 JSON
- `FILE_CHECKSUMS.sha256` — 패키지 내 전 파일 SHA-256 무결성 검증 목록
- **`00_PROJECT_CHARTER/`**
  - [PROJECT_OVERVIEW.md](00_PROJECT_CHARTER/PROJECT_OVERVIEW.md) — 플랫폼 비전, 3대 마켓 유형 통합 정의, 핵심 비즈니스 목표
  - [DEVELOPMENT_PRINCIPLES.md](00_PROJECT_CHARTER/DEVELOPMENT_PRINCIPLES.md) — 독립 개발 원칙, 상용 수준 정의, 기술 스택 불변 원칙
  - [SCOPE_AND_NON_GOALS.md](00_PROJECT_CHARTER/SCOPE_AND_NON_GOALS.md) — Phase 0~10 포함 범위 및 명시적 제외 범위(Non-Goals)
  - [ASSUMPTIONS_AND_DECISIONS.md](00_PROJECT_CHARTER/ASSUMPTIONS_AND_DECISIONS.md) — 아키텍처 결정 기록(ADR) 및 정책/법률 외부 확인 필요 사항 구분
- **`01_A_TO_Z_BLUEPRINT/`** — **제3조 A~Z 26개 전 항목 초정밀 개발 가이드라인 (각 문서당 18개 필수 섹션 완비)**
  - [A_ARCHITECTURE.md](01_A_TO_Z_BLUEPRINT/A_ARCHITECTURE.md) ~ [Z_ZERO_TRUST_AND_TREESCAN.md](01_A_TO_Z_BLUEPRINT/Z_ZERO_TRUST_AND_TREESCAN.md) (총 26개 파일)
- **`02_SYSTEM_ARCHITECTURE/`**
  - [SYSTEM_OVERVIEW.md](02_SYSTEM_ARCHITECTURE/SYSTEM_OVERVIEW.md), [COMPONENT_DIAGRAMS.md](02_SYSTEM_ARCHITECTURE/COMPONENT_DIAGRAMS.md), [DEPLOYMENT_ARCHITECTURE.md](02_SYSTEM_ARCHITECTURE/DEPLOYMENT_ARCHITECTURE.md), [EVENT_FLOW.md](02_SYSTEM_ARCHITECTURE/EVENT_FLOW.md), [THREAT_MODEL.md](02_SYSTEM_ARCHITECTURE/THREAT_MODEL.md)
- **`03_DATABASE/`**
  - [ERD.md](03_DATABASE/ERD.md), [TABLE_DICTIONARY.md](03_DATABASE/TABLE_DICTIONARY.md), [INDEX_AND_QUERY_DESIGN.md](03_DATABASE/INDEX_AND_QUERY_DESIGN.md), 및 `migrations/` 하위의 실제 실행 가능한 PostgreSQL DDL SQL 파일들
- **`04_API_SPECIFICATION/`**
  - [openapi.yaml](04_API_SPECIFICATION/openapi.yaml) (OpenAPI 3.1), [AUTHORIZATION_MATRIX.md](04_API_SPECIFICATION/AUTHORIZATION_MATRIX.md), [API_ERROR_CONTRACT.md](04_API_SPECIFICATION/API_ERROR_CONTRACT.md), [API_TEST_CASES.md](04_API_SPECIFICATION/API_TEST_CASES.md)
- **`05_FRONTEND_UXUI/`**
  - [USER_WEB.md](05_FRONTEND_UXUI/USER_WEB.md), [DEVELOPER_CONSOLE.md](05_FRONTEND_UXUI/DEVELOPER_CONSOLE.md), [ADMIN_CONSOLE.md](05_FRONTEND_UXUI/ADMIN_CONSOLE.md), [DESIGN_SYSTEM.md](05_FRONTEND_UXUI/DESIGN_SYSTEM.md), [SCREEN_INVENTORY.md](05_FRONTEND_UXUI/SCREEN_INVENTORY.md), [RESPONSIVE_ACCESSIBILITY.md](05_FRONTEND_UXUI/RESPONSIVE_ACCESSIBILITY.md)
- **`06_ANDROID/`**
  - [ANDROID_ARCHITECTURE.md](06_ANDROID/ANDROID_ARCHITECTURE.md), [APK_INSTALL_ENGINE.md](06_ANDROID/APK_INSTALL_ENGINE.md), [APP_UPDATE_ENGINE.md](06_ANDROID/APP_UPDATE_ENGINE.md), [DEVICE_TEST_MATRIX.md](06_ANDROID/DEVICE_TEST_MATRIX.md)
- **`07_DISTRIBUTION/`**
  - [ARTIFACT_PIPELINE.md](07_DISTRIBUTION/ARTIFACT_PIPELINE.md), [APK_AAB_SIGNING.md](07_DISTRIBUTION/APK_AAB_SIGNING.md), [STORAGE_CDN.md](07_DISTRIBUTION/STORAGE_CDN.md)
- **`08_SEARCH_CONNECTORS/`**
  - [OPENSEARCH_SCHEMA.md](08_SEARCH_CONNECTORS/OPENSEARCH_SCHEMA.md), [SEARCH_RANKING.md](08_SEARCH_CONNECTORS/SEARCH_RANKING.md), [EXTERNAL_CONNECTOR_CONTRACT.md](08_SEARCH_CONNECTORS/EXTERNAL_CONNECTOR_CONTRACT.md), [PROVENANCE_AND_RIGHTS.md](08_SEARCH_CONNECTORS/PROVENANCE_AND_RIGHTS.md)
- **`09_SECURITY_COMPLIANCE/`**
  - [SECURITY_ARCHITECTURE.md](09_SECURITY_COMPLIANCE/SECURITY_ARCHITECTURE.md), [APP_REVIEW_POLICY.md](09_SECURITY_COMPLIANCE/APP_REVIEW_POLICY.md), [DEVELOPER_VERIFICATION.md](09_SECURITY_COMPLIANCE/DEVELOPER_VERIFICATION.md), [COMPLIANCE_CHECKLIST.md](09_SECURITY_COMPLIANCE/COMPLIANCE_CHECKLIST.md)
- **`10_COMMERCE_ADS_ANALYTICS/`**
  - [PAYMENT_SETTLEMENT.md](10_COMMERCE_ADS_ANALYTICS/PAYMENT_SETTLEMENT.md), [ADVERTISING_ENGINE.md](10_COMMERCE_ADS_ANALYTICS/ADVERTISING_ENGINE.md), [ANALYTICS_EVENTS.md](10_COMMERCE_ADS_ANALYTICS/ANALYTICS_EVENTS.md)
- **`11_WORK_ORDERS/`**
  - [MASTER_WORK_ORDER.md](11_WORK_ORDERS/MASTER_WORK_ORDER.md), [WORK_ORDER_INDEX.csv](11_WORK_ORDERS/WORK_ORDER_INDEX.csv), [WORK_ORDER_DEPENDENCIES.md](11_WORK_ORDERS/WORK_ORDER_DEPENDENCIES.md), 및 `individual/` 하위의 도메인별 개별 Work Order 파일들
- **`12_AGENT_EXECUTION/`**
  - [AGENT_HANDOFF_PROTOCOL.md](12_AGENT_EXECUTION/AGENT_HANDOFF_PROTOCOL.md), [AGENT_ROLE_MATRIX.md](12_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md), [EXECUTION_SEQUENCE.md](12_AGENT_EXECUTION/EXECUTION_SEQUENCE.md), [CONFLICT_PREVENTION.md](12_AGENT_EXECUTION/CONFLICT_PREVENTION.md), [REPORTING_TEMPLATE.md](12_AGENT_EXECUTION/REPORTING_TEMPLATE.md)
- **`13_TESTING_QA/`**
  - [MASTER_TEST_MATRIX.csv](13_TESTING_QA/MASTER_TEST_MATRIX.csv), [ACCEPTANCE_CRITERIA.md](13_TESTING_QA/ACCEPTANCE_CRITERIA.md), [E2E_TEST_PLAN.md](13_TESTING_QA/E2E_TEST_PLAN.md), [INDEPENDENT_AUDIT.md](13_TESTING_QA/INDEPENDENT_AUDIT.md)
- **`14_INFRASTRUCTURE_DEVOPS/`**
  - [DEPLOYMENT_PLAN.md](14_INFRASTRUCTURE_DEVOPS/DEPLOYMENT_PLAN.md), [CI_CD_REQUIREMENTS.md](14_INFRASTRUCTURE_DEVOPS/CI_CD_REQUIREMENTS.md), [MONITORING.md](14_INFRASTRUCTURE_DEVOPS/MONITORING.md), [BACKUP_RECOVERY.md](14_INFRASTRUCTURE_DEVOPS/BACKUP_RECOVERY.md)
- **`15_TREESCAN_FUTURE_INTEGRATION/`**
  - [INTEGRATION_ARCHITECTURE.md](15_TREESCAN_FUTURE_INTEGRATION/INTEGRATION_ARCHITECTURE.md), [SEARCH_API_ADAPTER.md](15_TREESCAN_FUTURE_INTEGRATION/SEARCH_API_ADAPTER.md), [SSO_STRATEGY.md](15_TREESCAN_FUTURE_INTEGRATION/SSO_STRATEGY.md), [INDEPENDENCE_REQUIREMENTS.md](15_TREESCAN_FUTURE_INTEGRATION/INDEPENDENCE_REQUIREMENTS.md)
- **`16_FINAL_DELIVERY/`**
  - [PHASE_ROADMAP.md](16_FINAL_DELIVERY/PHASE_ROADMAP.md), [REQUIREMENTS_TRACEABILITY.csv](16_FINAL_DELIVERY/REQUIREMENTS_TRACEABILITY.csv), [DELIVERY_CHECKLIST.md](16_FINAL_DELIVERY/DELIVERY_CHECKLIST.md), [KNOWN_LIMITATIONS.md](16_FINAL_DELIVERY/KNOWN_LIMITATIONS.md), [OFFICIAL_REFERENCES.md](16_FINAL_DELIVERY/OFFICIAL_REFERENCES.md), [FINAL_BLUEPRINT_REPORT.md](16_FINAL_DELIVERY/FINAL_BLUEPRINT_REPORT.md)

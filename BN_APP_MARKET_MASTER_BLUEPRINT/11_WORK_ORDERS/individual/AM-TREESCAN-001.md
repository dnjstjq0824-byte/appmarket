# AM-TREESCAN-001 — 미래 TreeScan 통합용 Federated Search 어댑터, 카탈로그 피드 및 RFC 8693 OIDC SSO 토큰 교환기 구현 (Phase 10)

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-TREESCAN-001` |
| **작업명 (Title)** | 미래 TreeScan 통합용 Federated Search 어댑터, 카탈로그 피드 및 RFC 8693 OIDC SSO 토큰 교환기 구현 (Phase 10) |
| **목표 (Objective)** | `REQ-Z-001;REQ-Z-003` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Search & Connector Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-TEST-002` |
| **개발 Phase 및 우선순위** | `Phase 10` / `P2` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/Z_ZERO_TRUST_AND_TREESCAN.md;15_TREESCAN_FUTURE_INTEGRATION/INTEGRATION_ARCHITECTURE.md;15_TREESCAN_FUTURE_INTEGRATION/SEARCH_API_ADAPTER.md;15_TREESCAN_FUTURE_INTEGRATION/SSO_STRATEGY.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/treescan/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/treescan/adapter_handler.go;backend/internal/modules/treescan/sso_token_exchange.go;backend/internal/modules/treescan/adapter_contract_test.go` |
| **API 의존성 (API Dependencies)** | `GET /api/v1/treescan/federated-search;GET /api/v1/treescan/catalog-feed;POST /api/v1/auth/token/exchange` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-API-003;TC-OPS-008` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-TREESCAN-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. TreeScan 본체 코드를 전혀 수정하지 않고, 앱 마켓 내부에 Feature Flag(`TREESCAN_INTEGRATION_ENABLED`, 기본 `false`)로 제어되는 전용 어댑터 모듈 구현.
2. `GET /api/v1/treescan/federated-search`(`p95 < 80ms`), `GET /api/v1/treescan/catalog-feed`(증분 커서 스트림), `POST /api/v1/auth/token/exchange`(RFC 8693 SSO) 구현 및 계약 테스트 완료.

## 2. Acceptance Criteria (수용 기준)
1. 기본값(`false`)에서 앱 마켓이 외부 TreeScan 의존성 없이 100% 독립 동작.
2. 활성화(`true`) 시 3대 연동 엔드포인트가 계약 스키마 및 응답시간 기준을 통과.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- TreeScan 연동 구간 트래픽 폭주 시 전용 Bulkhead Rate Limiter로 차단하여 독립 앱 마켓 본체 서비스 보호.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-API-003;TC-OPS-008`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-TREESCAN-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

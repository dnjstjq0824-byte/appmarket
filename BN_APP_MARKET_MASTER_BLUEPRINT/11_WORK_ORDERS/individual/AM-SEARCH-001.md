# AM-SEARCH-001 — OpenSearch 2.x 다국어(Nori/ICU) 인덱스, BM25+품질 하이브리드 랭킹, 자동완성 및 실시간 동기화 워커 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-SEARCH-001` |
| **작업명 (Title)** | OpenSearch 2.x 다국어(Nori/ICU) 인덱스, BM25+품질 하이브리드 랭킹, 자동완성 및 실시간 동기화 워커 구현 |
| **목표 (Objective)** | `REQ-O-001;REQ-O-002;REQ-O-003;REQ-R-001` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Search & Connector Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-DB-001;AM-API-002` |
| **개발 Phase 및 우선순위** | `Phase 5` / `P0` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/O_OPENSEARCH.md;08_SEARCH_CONNECTORS/OPENSEARCH_SCHEMA.md;08_SEARCH_CONNECTORS/SEARCH_RANKING.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/search/;backend/internal/worker/handlers/indexer/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/search/infrastructure/opensearch_index_manager.go;backend/internal/modules/search/application/search_query_builder.go;backend/internal/modules/search/application/postgres_fallback_search.go` |
| **API 의존성 (API Dependencies)** | `GET /api/v1/search/apps;GET /api/v1/search/suggest;POST /api/v1/admin/search/reindex` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-010;TC-OPS-005` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-SEARCH-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. `bn_apps_v1` 매핑 생성, `bn_apps_read`/`bn_apps_write` Alias 관리 및 Blue-Green 무중단 Reindex 구현.
2. 한국어(Nori) 형태소/초성 분석기, `multi_match` + `function_score`(베이지안 평점, 최근성, 직접설치 부스트) 쿼리 빌더 및 50ms 이내 자동완성 구현.
3. OpenSearch 장애 시 PostgreSQL `pg_trgm` 비상 검색으로 자동 전환하는 Circuit Breaker Fallback 구현.

## 2. Acceptance Criteria (수용 기준)
1. 카탈로그 출시 승인 후 5초 이내에 검색 인덱스 반영.
2. OpenSearch 중단 시 에러 없이 PostgreSQL Fallback 검색 결과 반환.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 색인 실패 문서는 `q:search:indexer` 재시도 큐로 복구하며 부분 실패가 전체 벌크 배치를 중단시키지 않도록 처리.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-010;TC-OPS-005`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-SEARCH-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

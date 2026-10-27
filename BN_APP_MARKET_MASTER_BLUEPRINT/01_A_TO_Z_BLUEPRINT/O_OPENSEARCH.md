# O — OpenSearch — OpenSearch 통합 검색 인덱스, 다국어 형태소 분석, BM25+품질 랭킹 및 자동완성 엔진 (OpenSearch Index Architecture, Multilingual Analyzers, Hybrid Ranking & Autocomplete Engine)

- **문서 ID**: `DOC-AZ-O`
- **알파벳 코드**: `O`
- **관련 요구사항 ID**: `REQ-O-001`, `REQ-O-002`, `REQ-O-003`
- **연관 Work Order ID**: `AM-SEARCH-001`, `AM-SEARCH-002`
- **연관 테스트 케이스 ID**: `TC-FUNC-010`, `TC-OPS-005`
- **참조 상세 설계 문서**: [08_SEARCH_CONNECTORS/OPENSEARCH_SCHEMA.md](../08_SEARCH_CONNECTORS/OPENSEARCH_SCHEMA.md), [08_SEARCH_CONNECTORS/SEARCH_RANKING.md](../08_SEARCH_CONNECTORS/SEARCH_RANKING.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 자체 스토어 앱(Type 1)과 외부 스토어 앱(Type 2) 수십만 건을 대상으로 한국어·영어·일본어 등 다국어 초성/형태소/오타 교정 검색과 50ms 이내의 자동완성을 제공하고, 단순 키워드 매칭을 넘어 앱 품질·평점·최신성·직접 설치 가능 여부를 반영한 정교한 검색 랭킹을 구현하기 위해 필요하다.
- **달성 목표**: OpenSearch 2.17+ 인덱스 매핑(`bn_apps_v1`, `bn_autocomplete_v1`), Nori(한국어)/ICU 다국어 분석기 설정, BM25 + `function_score` 랭킹 수식, Blue-Green Zero-Downtime Reindexing, 그리고 PostgreSQL↔OpenSearch 실시간 동기화 파이프라인을 정의한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **일반 사용자**: 키워드 검색, 실시간 자동완성 제안, 패싯 필터(카테고리, 무료/유료, 직접설치 가능 여부, 기기 호환 여부, 최소 평점) 탐색.
- **Search Ops Admin**: 동의어 사전(`synonyms.txt`), 금칙어(`stop_words.txt`), 검색 결과 핀 고정/부스팅 룰 관리 및 무중단 재색인 실행.

## 3. 필수 기능 (Mandatory Features)
1. **이중 인덱스 및 Alias 무중단 구조**:
   - 읽기/쓰기 별칭 `bn_apps_read` / `bn_apps_write`를 물리적 버전 인덱스(`bn_apps_v1_001`)에 연결하여 매핑 변경 시에도 서비스 중단 없이 Blue-Green Reindex 및 Alias 원자적 스왑을 지원한다.
2. **다국어 텍스트 분석기 파이프라인**:
   - 한국어: `nori_tokenizer` (`decompound_mode: mixed`) + `nori_part_of_speech` + `nori_readingform` + 한글 자모/초성 Edge N-gram(`ㅂㅋ` → `부킹` 자동완성).
   - 글로벌 다국어: `icu_tokenizer` + `icu_folding` + `lowercase` + ` asciifolding` + `shingle`.
3. **복합 랭킹 수식 (BM25 + Quality Function Score)**:
   - 텍스트 관련도 점수(`_score` from Multi-Match on `title^5`, `developer_name^3`, `tags^2.5`, `package_name^4`, `short_description^1.5`, `full_description^0.8`)에 다음 비즈니스 품질 신호를 결합한다:
     - `rating_bayesian_score`: 베이지안 평균 평점 가중치.
     - `downloads_log_boost`: `log1p(install_count_30d)` 스케일 부스트.
     - `freshness_decay`: 최근 업데이트 일자 기준 Gaussian Decay.
     - `direct_install_boost`: 자체 검증 완료된 직접 설치 앱(`BN_HOSTED`, `FDROID_REPO`) 가중치(`1.15x`, 사용자 필터로 조정 가능).
4. **중복 앱 그룹핑 (Collapse / Deduplication)**:
   - 동일한 `catalog_app_id`에 여러 배포처가 존재할 때 검색 결과에서는 대표 카탈로그 앱 1개로 묶어 노출하되(`inner_hits`로 제공 가능한 배포처 배지 목록 함께 반환), 동일 앱이 검색 결과 1~5위를 중복 도배하는 현상을 방지한다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **자동완성(`GET /api/v1/search/suggest`) 아키텍처**:
  - OpenSearch `completion` suggester 및 `search_as_you_type` 필드를 결합하여, 입력 후 `p95 < 40ms` 이내에 ① 추천 검색어(Query Suggestions) 최대 5개와 ② 바로 이동 가능한 대표 앱(Direct App Hits: 아이콘, 이름, 개발사, 평점) 최대 3개를 동시에 반환한다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **`web-user`**: `GlobalSearchBar.vue` (150ms 디바운스 자동완성 드롭다운, 키보드 방향키 탐색), `SearchFacetSidebar.vue` (배포처 유형, 가격, 평점, 라이선스, 요구 SDK 필터).

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/search/`:
  - `opensearch_query_builder.go`: 사용자 로케일, 기기 SDK/ABI 호환성 필터, 국가별 노출 제한(`allowed_countries` term filter)을 모든 검색 쿼리의 `bool.filter` 절에 강제 주입하여 비호환/법적 차단 앱이 검색에 노출되지 않도록 보장한다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- 검색 원본 데이터(Source of Truth)는 PostgreSQL에 있으며, `search_synonyms`, `search_boost_rules`, `search_query_logs` 테이블로 검색 운영 데이터를 관리한다.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `GET /api/v1/search/apps`: 통합 검색 API (`q`, `category`, `source_type`, `price_type`, `min_rating`, `compatible_only`, `sort`, `cursor`, `limit`).
- `GET /api/v1/search/suggest`: 실시간 자동완성 및 추천 검색어 API (`q`, `locale`).
- `POST /api/v1/admin/search/reindex`: 무중단 Blue-Green 재색인 작업 트리거.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **인덱스 동기화 상태**: 카탈로그 앱 출시 승인 또는 메타데이터 변경 시 `outbox_events` → `q:search:indexer`를 통해 3초 이내에 OpenSearch 문서가 Upsert되며, 앱 차단(`SUSPENDED_BY_ADMIN`) 시에는 동기식/최우선 순위로 즉시 인덱스에서 제거(`is_searchable = false`)한다.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 사용자 검색어(`q`)에 포함된 특수 쿼리 구문으로 인한 정규식 서비스 거부(ReDoS)나 무거운 와일드카드 쿼리를 막기 위해 `simple_query_string` 대신 안전한 `multi_match` 쿼리 빌더만 사용하고 검색어 길이를 최대 100자로 제한한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 검색 결과가 0건일 때 오타 교정 제안(`"혹시 '메신저'를 찾으셨나요?"`)과 인기 검색어/카테고리 바로가기를 표시한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **한국어/영어 형태소 및 오타 검색 품질 테스트 (`TC-FUNC-010`)**: 한국어 복합명사, 초성, 영문 오타(`Fuziness: AUTO`) 검색 시 목표 앱이 상위 3위 이내에 반환되는지 골든 데이터셋으로 검증.
- **OpenSearch 장애 시 PostgreSQL Fallback 테스트 (`TC-OPS-005`)**: OpenSearch 연결 차단 시 에러 대신 PostgreSQL `pg_trgm` / Full-Text Search 비상 경로로 자동 전환되어 검색 결과가 반환되는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `C — Catalog Engine`, `E — External Store Connectors`, `R — Recommendation Engine`, `Z — TreeScan Integration`과 연동된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0`
- **개발 Phase**: 기본 검색은 `Phase 3`, 외부 앱 통합 검색 및 고도화 랭킹은 `Phase 5`에서 구현한다 (`AM-SEARCH-001`).

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 대량 외부 피드 유입 시 색인 지연** → OpenSearch `_bulk` API(배치 크기 500건 또는 5MB, 1초 플러시 주기)를 사용하여 단건 색인 대비 처리량을 20배 이상 높인다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- OpenSearch 매핑 JSON(`08_SEARCH_CONNECTORS/OPENSEARCH_SCHEMA.md`), Go 검색 쿼리 빌더 및 인덱서 워커 (`backend/internal/modules/search/*`).

## 17. Acceptance Criteria (수용 기준)
1. 카탈로그 승인 후 5초 이내에 검색 API에서 해당 앱이 조회되어야 한다.
2. 국가 차단된 앱이나 비공개(`DRAFT`/`SUSPENDED`) 앱은 어떤 검색어로도 검색 결과나 자동완성에 절대 노출되지 않아야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-SEARCH-001/opensearch_mapping_and_query.json`: 실제 인덱스 생성 및 복합 랭킹 쿼리 Explain 결과.
- `artifacts/evidence/AM-SEARCH-001/search_fallback_test.log`: OpenSearch 중단 시 PostgreSQL 비상 검색 전환 로그.

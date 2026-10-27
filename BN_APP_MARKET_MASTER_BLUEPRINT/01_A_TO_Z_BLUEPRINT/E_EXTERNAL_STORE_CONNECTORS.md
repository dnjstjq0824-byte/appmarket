# E — External Store Connectors — 외부 앱스토어 커넥터, 데이터 권리 분류 및 출처(Provenance) 관리 엔진 (External Store Connectors, Data Rights Classification & Provenance Engine)

- **문서 ID**: `DOC-AZ-E`
- **알파벳 코드**: `E`
- **관련 요구사항 ID**: `REQ-E-001`, `REQ-E-002`, `REQ-E-003`
- **연관 Work Order ID**: `AM-SEARCH-002`, `AM-ADMIN-002`
- **연관 테스트 케이스 ID**: `TC-FUNC-004`, `TC-SEC-003`
- **참조 상세 설계 문서**: [08_SEARCH_CONNECTORS/EXTERNAL_CONNECTOR_CONTRACT.md](../08_SEARCH_CONNECTORS/EXTERNAL_CONNECTOR_CONTRACT.md), [08_SEARCH_CONNECTORS/PROVENANCE_AND_RIGHTS.md](../08_SEARCH_CONNECTORS/PROVENANCE_AND_RIGHTS.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: Type 2(연합형 외부 앱 검색엔진) 및 Type 3(통합 비교 마켓)를 구현할 때, 외부 배포처마다 데이터 라이선스, API 존재 여부, 바이너리 재배포 허용 여부가 완전히 다르다(제11조). 존재하지 않는 비공식 API를 창작하거나 타사 유료/독점 APK를 무단 복제하는 법적·보안적 위험을 원천 차단하고, 합법적이고 검증된 경로로만 외부 앱 메타데이터와 공식 링크를 수집·동기화하기 위해 필요하다.
- **달성 목표**: 배포 권한 5단계 분류 체계, F-Droid 공식 `index-v2.json` 서명 검증 커넥터, 파트너 스토어 표준 JSON Feed 커넥터, 개발자 공인 외부 스토어 링크 검증기, 동기화 스케줄러 및 중복 제거 파이프라인을 설계한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **일반 사용자**: 외부 스토어에 존재하는 앱을 통합 검색하고, 데이터 출처·동기화 시각·공식 설치 링크를 확인하여 안전하게 이동.
- **Search & Connector Admin**: 커넥터별 활성화 상태(Feature Flag), 동기화 주기(Cron), 에러율 모니터링, 라이선스 정책 위반 소스 즉시 비활성화.

## 3. 필수 기능 (Mandatory Features)
1. **데이터 권리 및 배포 가능 범위 5단계 엄격 구분 (제11조 필수 요건)**:
   - `LEVEL_1_DIRECT_HOSTED`: 개발자가 당사 플랫폼에 직접 업로드하고 배포권을 부여한 앱 (자체 CDN 다운로드 + 직접 설치 가능).
   - `LEVEL_2_LICENSED_FEDERATED_BINARY`: F-Droid 등 오픈소스 라이선스(GPL/Apache/MIT) 및 리포지토리 정책에 따라 메타데이터와 공식 리포지토리 APK 링크/서명 검증이 허용된 앱 (출처·원본 서명 명시 후 공식 리포지토리 미러 또는 연동 설치 가능).
   - `LEVEL_3_SEARCH_AND_INFO_ONLY`: 공식 제휴 피드 또는 공개 허용된 카탈로그를 통해 앱 메타데이터(이름, 아이콘, 버전, 가격) 비교 제공이 가능하나 바이너리 호스팅은 불가한 앱.
   - `LEVEL_4_OFFICIAL_LINK_ONLY`: 상세 설명 전문 복제 없이 앱 식별 정보와 공식 스토어 이동 딥링크(`market://details?id=...`, `https://play.google.com/...`, `samsungapps://...`)만 제공하는 앱.
   - `LEVEL_5_RESTRICTED_PROHIBITED`: 약관상 수집·노출이 금지되었거나 권리자의 삭제 요청(DMCA/권리 침해)이 접수되어 색인 및 노출이 전면 차단된 데이터.
2. **스토어별 실제 기술 제약 반영 (비공식/가짜 API 창작 금지)**:
   - **F-Droid**: 공식 `https://f-droid.org/repo/entry.json` 및 `index-v2.json`(또는 서명된 `index-v1.jar`)을 다운로드하고 JAR/GPG 서명을 암호학적으로 검증한 후 파싱한다.
   - **Google Play / Samsung Galaxy Store / Apple App Store**: 타사 전체 앱을 덤프하는 공개 공식 검색 API가 존재하지 않음을 명시한다. 따라서 ① 당사 개발자가 직접 등록하고 검증받은 공식 외부 스토어 URL, ② 공식 제휴 데이터 피드(`Partner Feed Schema v1`), ③ 법무 승인을 거친 공개 웹 표준 메타데이터(`Schema.org/SoftwareApplication`)만 수집하며, 무단 APK 추출기는 절대 구현하지 않는다.
3. **출처(Provenance) 필수 표기**: 모든 외부 앱 레코드는 `source_name`, `source_license`, `official_url`, `last_synced_at`, `etag_or_hash`, `direct_install_allowed (boolean)`를 필수 저장·표시한다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **F-Droid Index v2 증분 동기화 파이프라인**:
  1. 스케줄러가 6시간 주기로 `entry.json`을 조회하여 `timestamp` 및 `index.sha256` 변경 여부를 확인.
  2. 변경 시에만 `index-v2.json`을 스트리밍 다운로드하고 SHA-256 및 리포지토리 공개키 서명을 검증.
  3. JSON Stream Parser(`encoding/json.Decoder`)를 사용해 메모리 피크를 200MB 이하로 유지하며 패키지 목록을 파싱하고, `external_store_items` 테이블에 Upsert 후 `catalog.synced.v1` 이벤트를 발행하여 OpenSearch 인덱스를 갱신.
- **중복 앱 매칭 및 서명 충돌 방지**:
  - 외부 커넥터가 수집한 앱의 `package_name`이 기존 `catalog_apps.primary_package_name`과 일치하면 동일 카탈로그 앱의 하위 `app_distributions`로 연결하되, 패키지 서명 지문(`signer_sha256`)을 비교하여 일치 여부를 기록한다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **`web-user`**: 검색 결과 및 앱 상세 카드에 `ProvenanceBadge.vue` (예: `출처: F-Droid Official Repo · 동기화: 2시간 전 · 공식 링크 이동`)를 렌더링.
- **`web-admin`**: `ConnectorDashboardView.vue` (커넥터별 ON/OFF 토글, 최근 실행 상태, 수집 건수, 실패 로그, 특정 외부 패키지 즉시 차단/언블록 기능).

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/connector/domain/connector_interface.go`:
  ```go
  type ExternalConnector interface {
      ProviderCode() string
      CheckRightsPolicy() RightsPolicyLevel
      FetchDelta(ctx context.Context, cursor string) (*SyncBatch, error)
      VerifyAuthenticity(ctx context.Context, payload []byte) error
  }
  ```
- 각 커넥터는 독립된 Rate Limiter와 Circuit Breaker를 가지며, 로봇 배제 표준(`robots.txt`) 및 허용된 화이트리스트 도메인 정책을 코드 레벨에서 강제한다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**:
  - `external_connectors` (`id`, `provider_code UNIQUE`, `display_name`, `rights_level`, `base_url`, `public_key_pem`, `is_enabled`, `sync_interval_minutes`, `last_sync_status`, `last_synced_at`).
  - `external_sync_jobs` (`id`, `connector_id`, `started_at`, `finished_at`, `status`, `items_processed`, `items_failed`, `error_summary`).
  - `external_store_items` (`id`, `connector_id`, `external_package_id`, `catalog_app_id`, `rights_level`, `official_store_url`, `raw_metadata JSONB`, `signer_sha256`, `latest_version_code`, `last_synced_at`).

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `GET /api/v1/admin/connectors`: 등록된 외부 스토어 커넥터 목록 및 상태 조회.
- `POST /api/v1/admin/connectors/{connectorId}/trigger-sync`: 수동 동기화 즉시 실행.
- `PATCH /api/v1/admin/connectors/{connectorId}`: 커넥터 활성화/비활성화 및 정책 변경.
- `GET /api/v1/catalog/apps/{slug}/sources`: 특정 앱의 모든 외부/내부 출처, 동기화 시각, 공식 링크 조회.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **커넥터 동기화 상태 머신**: `IDLE` → `FETCHING_MANIFEST` → `VERIFYING_SIGNATURE` → `PARSING_DELTA` → `UPSERTING_CATALOG` → `INDEXING_SEARCH` → `COMPLETED` (또는 서명 오류 시 즉시 `ABORTED_SIGNATURE_INVALID`).

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 외부 리포지토리에서 내려받은 JSON/XML 피드를 파싱할 때 XXE(XML External Entity) 방지, JSON 최대 깊이(Max Depth 32) 및 단일 문자열 길이 제한을 적용하여 악성 피드를 통한 DoS 공격을 차단한다.
- 외부 스토어 딥링크(`official_store_url`)는 허용된 스킴/도메인 정규식(`^https://play\.google\.com/`, `^https://f-droid\.org/`, `^https://galaxystore\.samsung\.com/`)만 통과시켜 Open Redirect 및 피싱 링크 삽입을 차단한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- `LEVEL_3_SEARCH_AND_INFO_ONLY` 및 `LEVEL_4_OFFICIAL_LINK_ONLY` 등급의 앱은 버튼 레이블을 명확히 `Google Play에서 보기 ↗` 또는 `공식 배포처로 이동 ↗`으로 표시하며, `설치` 또는 `다운로드`라는 표현을 사용하지 않는다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **F-Droid 서명 변조 차단 테스트 (`TC-SEC-003`)**: 테스트용 `index-v2.json`의 바이트를 1바이트 변조한 뒤 동기화 실행 시 `ABORTED_SIGNATURE_INVALID`로 즉시 중단되고 DB에 단 1건도 반영되지 않음을 검증.
- **권리 등급별 UI/API 응답 검증 (`TC-FUNC-004`)**: `LEVEL_4_OFFICIAL_LINK_ONLY` 앱에 대해 직접 다운로드 URL 발급 API 호출 시 `403 ERR_DIRECT_DOWNLOAD_NOT_PERMITTED`가 반환되는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `C — Catalog Engine`, `O — OpenSearch`, `K — Kotlin Android Market`(외부 스토어 인텐트 호출)과 연동된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P1 (Type 2 & Type 3 마켓 핵심)`
- **개발 Phase**: `Phase 5 (외부 앱 검색 및 통합 검색)`에서 집중 구현한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 외부 리포지토리 서버 다운 또는 응답 지연** → 커넥터 워커 타임아웃(30초) 및 지수 백오프(Exponential Backoff) 적용. 기존 수집된 메타데이터는 유지하되 UI에 `마지막 확인: YYYY-MM-DD HH:MM UTC`를 투명하게 표시한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 외부 커넥터 프레임워크 및 F-Droid/Partner Feed 구현체 (`backend/internal/modules/connector/*`).
- 관리자 커넥터 관제 화면 및 사용자 출처 배지 컴포넌트.

## 17. Acceptance Criteria (수용 기준)
1. 존재하지 않는 외부 스토어 비공식 API나 무단 APK 스크래퍼 코드가 저장소에 없어야 한다.
2. F-Droid 테스트 피드 동기화 시 서명 검증, 카탈로그 매핑, 출처/동기화 시각 표시가 정확히 동작해야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-SEARCH-002/fdroid_sync_verification.log`: 서명 검증 성공 및 변조 차단 테스트 로그.
- `artifacts/evidence/AM-SEARCH-002/provenance_api_output.json`: 출처·권리등급·공식링크가 포함된 API 응답 스냅샷.

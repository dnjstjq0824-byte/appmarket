# C — Catalog Engine — 통합 앱 카탈로그, 다국어 메타데이터, 카테고리 및 배포처 매핑 엔진 (Unified App Catalog, Localized Metadata, Taxonomy & Multi-Source Distribution Mapping)

- **문서 ID**: `DOC-AZ-C`
- **알파벳 코드**: `C`
- **관련 요구사항 ID**: `REQ-C-001`, `REQ-C-002`, `REQ-C-003`
- **연관 Work Order ID**: `AM-DB-001`, `AM-API-002`, `AM-WEB-001`
- **연관 테스트 케이스 ID**: `TC-FUNC-002`, `TC-DB-001`
- **참조 상세 설계 문서**: [03_DATABASE/ERD.md](../03_DATABASE/ERD.md), [03_DATABASE/TABLE_DICTIONARY.md](../03_DATABASE/TABLE_DICTIONARY.md), [08_SEARCH_CONNECTORS/PROVENANCE_AND_RIGHTS.md](../08_SEARCH_CONNECTORS/PROVENANCE_AND_RIGHTS.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 자체 등록 앱(Type 1)과 외부 스토어에서 수집된 앱(Type 2)을 하나의 통합 뷰(Type 3)에서 비교·탐색하려면, 논리적 앱 상품(`catalog_apps`), 다국어 설명/미디어(`app_localizations`), 계층형 카테고리/태그, 그리고 복수의 배포처(`app_distributions`)를 일관되게 관리하는 카탈로그 엔진이 필수적이다.
- **달성 목표**: 앱 카탈로그의 동일성(Catalog Identity), 외부 배포처의 동일성(Distribution Identity), 앱 서명의 동일성(Signature Identity)을 엄격히 분리하면서도 단일 상세 페이지에서 통합 비교할 수 있는 데이터 구조와 API를 구축한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **일반 사용자**: 카테고리별 탐색, 다국어 앱 설명/스크린샷/영상 조회, 배포처별 버전·권한·서명 비교.
- **개발자 조직**: 자사 앱의 기본 메타데이터, 다국어 번역본, 아이콘/그래픽 에셋, 연령 등급 설문, 개인정보처리방침 URL 관리.
- **카탈로그 관리자 / Search Ops**: 카테고리 체계 관리, 에디터 추천 컬렉션(Editorial Curations) 편성, 중복 카탈로그 병합/분리.

## 3. 필수 기능 (Mandatory Features)
1. **3대 동일성 분리 스키마**:
   - `catalog_apps`: 논리적 앱 엔티티 (`canonical_slug`, `primary_package_name`, `developer_org_id`, `primary_category_id`, `content_rating`).
   - `app_distributions`: 배포처별 엔티티 (`app_id`, `source_type` [`BN_HOSTED`, `FDROID_REPO`, `GOOGLE_PLAY_LINK`, `GALAXY_STORE_LINK`, `OFFICIAL_SITE_LINK`], `distribution_mode` [`DIRECT_APK`, `EXTERNAL_LINK_ONLY`], `package_name`, `signing_cert_sha256`, `official_store_url`, `last_synced_at`).
   - `app_releases`: 각 배포처 하위의 버전 이력 (`version_code`, `version_name`, `min_sdk`, `target_sdk`, `supported_abis`, `release_notes`).
2. **다국어 메타데이터 폴백(Locale Fallback) 체인**: 사용자의 요청 로케일(예: `ko-KR`) → 언어 기본값(`ko`) → 앱의 기본 로케일(`default_locale`, 예: `en-US`) 순으로 자동 폴백.
3. **계층형 카테고리 및 태그**: 대분류(Apps / Games) → 중분류(생산성, 금융, 유틸리티, RPG 등 36개 표준 카테고리) → 시맨틱 태그(오프라인 지원, 광고 없음, 오픈소스 등).
4. **미디어 에셋 검증 및 최적화**: 앱 아이콘(512×512 PNG/WebP), 피처 그래픽(1024×500), 기기별 스크린샷(Phone/Tablet, 최대 8장), 소개 영상 링크의 해상도·비율·용량 검증.

## 4. 상세 기능 (Detailed Functional Specifications)
- **패키지명 충돌 및 카탈로그 매칭 규칙**:
  - 동일한 `package_name`(예: `org.mozilla.firefox`)을 가진 자체 등록 앱과 F-Droid/Google Play 외부 항목은 `catalog_apps` 하위에서 복수의 `app_distributions` 레코드로 연결하여 가격·버전·서명 비교 표를 구성한다.
  - 단, `BN_HOSTED`(직접 배포) 배포처의 소유권은 오직：① 최초 등록 후 서명 키 검증을 통과한 개발자 조직, 또는 ② 공식 도메인 DNS TXT 레코드(`_bn-verify.<domain>`) 및 기존 서명 키 소유 증명을 통과한 공식 개발자에게만 부여된다.
  - 만약 서명 인증서(`signing_cert_sha256`)가 다른 두 배포처(예: 공식 개발자 서명 vs F-Droid 재빌드 서명)가 동일 카탈로그 앱에 연결될 경우, DB의 `app_distributions.signature_compatibility_group`을 분리하고 UI에 `"배포처 간 서명이 달라 교차 업데이트 시 기존 앱 삭제 후 재설치가 필요합니다"`라는 경고를 자동으로 표시한다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **`web-user`**:
  - `CatalogCard.vue`: 앱 아이콘, 이름, 개발사, 평균 평점, 가격, 제공 배포처 배지(`Direct APK`, `F-Droid`, `Official Link`) 표시.
  - `DistributionComparisonTable.vue`: 배포처별 최신 버전, 파일 크기, 요구 Android 버전, 서명 지문(축약형 + 복사 버튼), 라이선스, 설치/이동 버튼 렌더링.
- **`web-dev`**:
  - `AppMetadataForm.vue`: 로케일 탭 전환형 다국어 입력 폼, 드래그앤드롭 스크린샷 업로더(해상도 실시간 검사).

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/catalog/application/catalog_query_service.go`: Redis Read-Through 캐시(TTL 5분, 배포 승인 이벤트 시 즉시 무효화)와 PostgreSQL JSON 집계를 결합하여 단일 쿼리로 앱 상세 + 다국어 + 배포처 비교 목록 + 최신 릴리스 정보를 조합.
- 이미지 업로드 시 격리 워커가 `libvips`/Go `image` 패키지로 EXIF 위치정보를 제거(Strip EXIF)하고 WebP 썸네일을 생성한다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**: `catalog_apps`, `app_localizations`, `app_media_assets`, `categories`, `tags`, `app_tags`, `app_distributions`, `editorial_collections`, `editorial_collection_items`.
- **제약조건**: `UNIQUE(app_id, locale)` on `app_localizations`; `UNIQUE(source_type, package_name, signing_cert_sha256)` on `app_distributions`.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `GET /api/v1/catalog/apps`: 카테고리, 태그, 배포 방식(`direct_only`), 최소 평점, 무료/유료 필터 및 커서 기반 페이지네이션(`cursor`, `limit`).
- `GET /api/v1/catalog/apps/{slug}`: 앱 상세 정보, 다국어 메타데이터, 모든 배포처 비교 리스트, 권한 목록, 서명 인증서 정보 조회.
- `GET /api/v1/catalog/categories`: 전체 카테고리 트리 및 앱 개수 조회.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **카탈로그 노출 상태 (`catalog_status`)**:
  - `DRAFT` (개발자 작성 중, 비공개) → `PENDING_INITIAL_REVIEW` (최초 심사 대기) → `PUBLISHED` (공개 노출) → `UNLISTED` (직접 링크로만 접근 가능) → `SUSPENDED_BY_ADMIN` (정책 위반 차단, 기존 설치자에게 경고 노출) → `ARCHIVED` (영구 종료).

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 앱 설명(`full_description`) 및 릴리스 노트에 포함된 마크다운/HTML은 서버 측에서 `bluemonday` Strict Policy로 살균(Sanitize)하여 Stored XSS를 원천 차단한다.
- 업로드된 이미지 파일에 숨겨진 폴리글랏(Polyglot) 스크립트 실행을 막기 위해 이미지 재인코딩 및 `Content-Disposition: inline; filename=...`, `X-Content-Type-Options: nosniff` 헤더를 강제한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 모든 앱 상세 화면은 **데이터 출처(Provenance)**와 **설치 유형**을 오해의 소지 없이 구분해야 한다:
  - 자체 호스팅 앱: `BN Verified Direct APK` 녹색 배지 + 파일 크기 및 SHA-256 표시.
  - 외부 스토어 링크 전용 앱: `외부 공식 스토어 이동 (Google Play)` 아웃링크 아이콘 표시 — 절대 자체 서버에서 직접 APK가 내려받아지는 것처럼 속이는 가짜 다운로드 버튼을 렌더링하지 않는다(제4조 준수).

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **다국어 폴백 테스트 (`TC-FUNC-002`)**: `Accept-Language: fr-FR` 요청 시 프랑스어 번역이 없으면 기본 로케일(`en-US`)이 반환되고, `ko-KR` 요청 시 한국어 번역이 정확히 반환되는지 검증.
- **3대 동일성 분리 무결성 테스트 (`TC-DB-001`)**: 동일 패키지명·상이한 서명 인증서를 가진 2개 배포처 등록 시 서명 그룹이 분리되고 덮어쓰기 호환성 플래그가 `false`로 계산되는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `D — Developer Platform`(메타데이터 입력), `E — External Store Connectors`(외부 카탈로그 입수), `O — OpenSearch`(색인 동기화), `N — Native App Distribution`(릴리스 연결)의 중심 허브로 동작한다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0 (플랫폼 핵심 도메인)`
- **개발 Phase**: `Phase 1 (DB 스키마)` 및 `Phase 3 (자체 카탈로그)`에서 우선 구현하고 `Phase 5`에서 외부 배포처 비교를 통합한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: Redis 캐시 스탬피드(Cache Stampede) 발생** → 인기 앱 상세 조회 시 Go `singleflight.Group`을 적용하여 동시에 수천 건의 캐시 미스가 발생해도 PostgreSQL로는 단 1건의 쿼리만 전달되도록 보호한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 카탈로그 도메인 모델, 리포지토리, 캐시 레이어 (`backend/internal/modules/catalog/*`).
- 사용자 웹 카탈로그 탐색/상세/비교 컴포넌트 (`web-user/pages/apps/[slug].vue`).

## 17. Acceptance Criteria (수용 기준)
1. 단일 앱 상세 API 응답에 자체 배포처와 외부 배포처의 버전, 가격, 서명 지문, 동기화 시각이 구조화되어 반환되어야 한다.
2. 허용되지 않은 HTML 태그(`<script>`, `onerror=`)가 포함된 앱 설명 저장 시 서버에서 즉시 제거되어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-API-002/catalog_detail_response.json`: 다중 배포처 및 서명 지문이 포함된 실제 API 응답 JSON.
- `artifacts/evidence/AM-API-002/xss_sanitization_test.log`: 악성 스크립트 태그 필터링 단위/통합 테스트 로그.

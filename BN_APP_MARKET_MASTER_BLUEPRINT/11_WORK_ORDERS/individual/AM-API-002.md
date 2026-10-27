# AM-API-002 — 통합 앱 카탈로그, 다국어 폴백, 카테고리/태그 및 다중 배포처 비교 API 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-API-002` |
| **작업명 (Title)** | 통합 앱 카탈로그, 다국어 폴백, 카테고리/태그 및 다중 배포처 비교 API 구현 |
| **목표 (Objective)** | `REQ-C-001;REQ-C-002;REQ-C-003` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Backend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-DB-001;AM-API-001` |
| **개발 Phase 및 우선순위** | `Phase 3` / `P0` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/C_CATALOG.md;04_API_SPECIFICATION/openapi.yaml` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/catalog/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/catalog/domain/catalog_app.go;backend/internal/modules/catalog/application/catalog_service.go;backend/internal/modules/catalog/presentation/http_handler.go` |
| **API 의존성 (API Dependencies)** | `GET /api/v1/catalog/apps;GET /api/v1/catalog/apps/{slug};GET /api/v1/catalog/categories;GET /api/v1/catalog/charts/{chartType}` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-002;TC-DB-001` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-API-002/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 카탈로그 목록/상세 조회, BCP 47 로케일 폴백 체인(`ko-KR -> ko -> default_locale`), `bluemonday` HTML 살균 구현.
2. 단일 앱 하위의 복수 배포처(`BN_HOSTED`, `FDROID_REPO`, `GOOGLE_PLAY_LINK` 등) 가격·버전·서명 호환성 그룹 비교 응답 조립.
3. `singleflight` + Redis Read-Through 캐시 적용.

## 2. Acceptance Criteria (수용 기준)
1. 상세 조회 시 모든 활성 배포처의 버전·서명지문·출처·공식링크·직접설치가능여부가 정확히 반환.
2. 악성 `<script>` 태그가 포함된 설명 저장 시 100% 살균 제거.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- Redis 캐시 장애 시 `singleflight`로 DB 직접 조회를 보호하며 에러 없이 정상 응답 반환.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-002;TC-DB-001`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-API-002/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

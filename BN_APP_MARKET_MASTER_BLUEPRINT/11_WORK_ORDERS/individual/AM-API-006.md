# AM-API-006 — 에디터 추천 컬렉션, 실시간/일간 4대 인기 차트 산출 배치 및 유사 앱 추천 API 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-API-006` |
| **작업명 (Title)** | 에디터 추천 컬렉션, 실시간/일간 4대 인기 차트 산출 배치 및 유사 앱 추천 API 구현 |
| **목표 (Objective)** | `REQ-R-001` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Search & Connector Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-API-002;AM-SEARCH-001` |
| **개발 Phase 및 우선순위** | `Phase 5` / `P1` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/R_RECOMMENDATION_ENGINE.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/catalog/recommendations/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/catalog/recommendations/chart_calculator.go;backend/internal/modules/catalog/recommendations/similar_apps_service.go` |
| **API 의존성 (API Dependencies)** | `GET /api/v1/catalog/charts/{chartType};GET /api/v1/catalog/apps/{slug}/recommendations` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-011` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-API-006/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. `TOP_FREE`, `TOP_PAID`, `TRENDING_RISING`, `NEW_NOTEWORTHY` 4대 차트 시간별 스냅샷 계산기 구현.
2. OpenSearch `more_like_this` 및 설치 상관관계 기반 유사 앱 추천 서비스 구현.

## 2. Acceptance Criteria (수용 기준)
1. 신규 출시 14일 이내 고품질 앱이 `NEW_NOTEWORTHY` 차트에 정상 노출.
2. 비호환/국가차단 앱이 추천 결과에서 제외됨.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 배치 지연 시 직전 시간대 차트 스냅샷을 무중단 서빙.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-011`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-API-006/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

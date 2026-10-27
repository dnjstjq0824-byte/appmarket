# AM-API-004 — 글로벌 국가별 노출 통제(Geo-Fencing), 연령 등급 강제 및 다통화 정책 엔진 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-API-004` |
| **작업명 (Title)** | 글로벌 국가별 노출 통제(Geo-Fencing), 연령 등급 강제 및 다통화 정책 엔진 구현 |
| **목표 (Objective)** | `REQ-G-001;REQ-G-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Backend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-API-002` |
| **개발 Phase 및 우선순위** | `Phase 8` / `P1` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/G_GLOBALIZATION.md;09_SECURITY_COMPLIANCE/COMPLIANCE_CHECKLIST.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/catalog/geo/;backend/internal/platform/i18n/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/catalog/geo/geo_policy_evaluator.go;backend/internal/platform/i18n/locale_resolver.go` |
| **API 의존성 (API Dependencies)** | `PUT /api/v1/developer/apps/{appId}/countries;PUT /api/v1/admin/policies/countries/{countryCode}` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-005;TC-SEC-004` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-API-004/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. `app_country_availability`의 개발자 설정(`is_allowed_by_dev`)과 법적 규제(`is_blocked_by_law`)를 통합 평가하는 `GeoPolicyEvaluator` 구현.
2. 성인 등급(`AGE_18_ADULT`) 앱에 대한 국가별 성인인증(`adult_verified_at`) 확인 가드 구현.
3. 법적 차단 국가 요청 시 HTTP `451 Unavailable For Legal Reasons` 반환.

## 2. Acceptance Criteria (수용 기준)
1. 법적 차단 설정된 국가 IP/계정에서 상세 조회 및 다운로드 토큰 요청 시 100% `451` 차단.
2. 법적 차단된 국가는 개발자가 임의로 허용 상태로 되돌릴 수 없음.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- GeoIP 헤더 누락 시 보수적 기본 프로필을 적용하여 성인/규제 앱이 미확인 지역에 노출되지 않도록 보호.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-005;TC-SEC-004`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-API-004/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

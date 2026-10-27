# AM-DEV-003 — 개발자 콘솔 유료 앱 국가별 가격표, 프로모션 코드 생성기 및 스폰서 광고 캠페인 매니저 UI 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-DEV-003` |
| **작업명 (Title)** | 개발자 콘솔 유료 앱 국가별 가격표, 프로모션 코드 생성기 및 스폰서 광고 캠페인 매니저 UI 구현 |
| **목표 (Objective)** | `REQ-B-001;REQ-B-002;REQ-Y-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Frontend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-DEV-002;AM-BILLING-001;AM-ADS-001` |
| **개발 Phase 및 우선순위** | `Phase 7` / `P1` |
| **입력 문서 (Input Documents)** | `05_FRONTEND_UXUI/DEVELOPER_CONSOLE.md;10_COMMERCE_ADS_ANALYTICS/ADVERTISING_ENGINE.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `apps/web-dev/src/views/commerce/;apps/web-dev/src/views/ads/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `backend/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `apps/web-dev/src/views/commerce/AppPricingEditorView.vue;apps/web-dev/src/views/commerce/PromoCodeManagerView.vue;apps/web-dev/src/views/ads/AdCampaignManagerView.vue` |
| **API 의존성 (API Dependencies)** | `PUT /api/v1/developer/apps/{appId}/pricing;POST /api/v1/developer/apps/{appId}/promotions;POST /api/v1/developer/orgs/{orgId}/ads/campaigns` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-COM-001;TC-COM-003` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-DEV-003/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 국가별 판매가·부가세·예상 수수료·정산 예정액 실시간 시뮬레이터 및 가격 일괄 저장 화면 구현.
2. 할인/리딤 프로모션 배치 생성기 및 스폰서 검색/카테고리 광고 캠페인(CPC 입찰, 일일 예산, ROAS 차트) 관리 화면 구현.

## 2. Acceptance Criteria (수용 기준)
1. 수수료 티어(15%/10%)에 따른 예상 수령액이 서버 계산식과 정확히 일치.
2. 광고 캠페인 일일 예산 소진 상태가 실시간으로 대시보드에 표시됨.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 최소 입찰가 미달 또는 일일 예산이 입찰가보다 작을 때 저장 전 폼 검증 차단.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-COM-001;TC-COM-003`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-DEV-003/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

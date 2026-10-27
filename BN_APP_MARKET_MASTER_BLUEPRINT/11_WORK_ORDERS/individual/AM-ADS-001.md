# AM-ADS-001 — 스폰서 앱 광고 캠페인, Generalized Second-Price 경매, 노출 토큰 및 부정클릭(Click Fraud) 방지 엔진 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-ADS-001` |
| **작업명 (Title)** | 스폰서 앱 광고 캠페인, Generalized Second-Price 경매, 노출 토큰 및 부정클릭(Click Fraud) 방지 엔진 구현 |
| **목표 (Objective)** | `REQ-B-002;REQ-Y-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Backend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-BILLING-001;AM-SEARCH-001` |
| **개발 Phase 및 우선순위** | `Phase 7` / `P1` |
| **입력 문서 (Input Documents)** | `10_COMMERCE_ADS_ANALYTICS/ADVERTISING_ENGINE.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/ads/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/ads/application/auction_engine.go;backend/internal/modules/ads/application/click_fraud_filter.go;backend/internal/modules/ads/presentation/http_handler.go` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/developer/orgs/{orgId}/ads/campaigns;GET /api/v1/ads/serve;POST /api/v1/ads/click` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-COM-003` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-ADS-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 자사 소유 앱에 한정한 스폰서 광고 캠페인 생성 및 관리자 승인 워크플로 구현.
2. GSP 경매, 서명된 1회용 `impression_token` 발급, 3중 부정클릭 필터(만료/속도/1시간 동일기기 중복 제거) 및 일일 예산 상한 자동 일시정지(`PAUSED_DAILY_BUDGET_CAP`) 구현.

## 2. Acceptance Criteria (수용 기준)
1. 동일 기기 50회 연속 클릭 시 최초 1회만 과금되고 49회는 무료 중복 처리.
2. 일일 예산 도달 즉시 광고 노출이 중단되고 예산 초과 과금이 0원이어야 함.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 광고 경매 엔진 지연(> 30ms) 시 검색 결과 지연을 막기 위해 광고 슬롯을 즉시 비우고 오가닉 검색 결과만 반환.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-COM-003`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-ADS-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

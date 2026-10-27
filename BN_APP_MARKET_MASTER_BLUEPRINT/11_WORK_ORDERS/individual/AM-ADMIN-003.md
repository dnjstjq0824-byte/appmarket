# AM-ADMIN-003 — 관리자 콘솔 광고 심사, 결제·환불 분쟁, 개발자 정산 집행 및 통합 통계 화면(ADM-12~15) 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-ADMIN-003` |
| **작업명 (Title)** | 관리자 콘솔 광고 심사, 결제·환불 분쟁, 개발자 정산 집행 및 통합 통계 화면(ADM-12~15) 구현 |
| **목표 (Objective)** | `REQ-W-001;REQ-M-003;REQ-Y-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Frontend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-ADMIN-001;AM-BILLING-001;AM-ADS-001` |
| **개발 Phase 및 우선순위** | `Phase 7` / `P1` |
| **입력 문서 (Input Documents)** | `05_FRONTEND_UXUI/ADMIN_CONSOLE.md;10_COMMERCE_ADS_ANALYTICS/PAYMENT_SETTLEMENT.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `apps/web-admin/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `android-client/;apps/web-user/` |
| **예상 생성/수정 파일 (Target Files)** | `apps/web-admin/src/views/AdCampaignReviewView.vue;apps/web-admin/src/views/OrdersRefundsAdminView.vue;apps/web-admin/src/views/SettlementPayoutAdminView.vue;apps/web-admin/src/views/PlatformAnalyticsAdminView.vue` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/admin/settlements/{settlementId}/approve-payout` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-COM-004;TC-COM-006` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-ADMIN-003/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. `ADM-12`(스폰서 광고 소재 심사 및 부정클릭 통계), `ADM-13`(결제/환불 분쟁 강제 환불 및 원장 대조), `ADM-14`(월간 개발자 정산 2인 승인 집행), `ADM-15`(플랫폼 전체 전환율 및 매출 통계 센터) 구현으로 20개 관리자 화면 전체 완성.

## 2. Acceptance Criteria (수용 기준)
1. 제7조의 20개 관리자 운영 영역(`ADM-01`~`ADM-20`)이 100% 구현 및 검증 완료.
2. 정산 지급 승인은 반드시 서로 다른 2명의 관리자 승인을 거쳐야만 최종 실행됨.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 원장 대차평형 검증 경보가 존재하는 정산 월은 지급 승인 버튼을 자동 잠금(`ON_HOLD_COMPLIANCE`).

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-COM-004;TC-COM-006`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-ADMIN-003/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

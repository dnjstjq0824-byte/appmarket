# AM-WEB-003 — 사용자 마켓 웹사이트 유료 앱 결제 체크아웃, 프로모션 리딤 적용 및 영수증/환불 신청 UI 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-WEB-003` |
| **작업명 (Title)** | 사용자 마켓 웹사이트 유료 앱 결제 체크아웃, 프로모션 리딤 적용 및 영수증/환불 신청 UI 구현 |
| **목표 (Objective)** | `REQ-B-001;REQ-M-001;REQ-M-003` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Frontend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-WEB-002;AM-BILLING-001` |
| **개발 Phase 및 우선순위** | `Phase 7` / `P1` |
| **입력 문서 (Input Documents)** | `05_FRONTEND_UXUI/USER_WEB.md;10_COMMERCE_ADS_ANALYTICS/PAYMENT_SETTLEMENT.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `apps/web-user/pages/checkout/;apps/web-user/pages/account/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `backend/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `apps/web-user/pages/checkout/[appSlug].vue;apps/web-user/components/PromoRedeemInput.vue;apps/web-user/components/RefundRequestModal.vue` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/commerce/quotes;POST /api/v1/commerce/orders;POST /api/v1/commerce/orders/{orderId}/refund` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-COM-001;TC-COM-006` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-WEB-003/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 유료 앱 결제 전 공급가액·부가세·프로모션 할인이 반영된 실시간 견적(`OrderQuote`) 확인 화면 구현.
2. PG 결제창 연동, 결제 완료 후 즉시 다운로드 활성화, 구매 내역 영수증 조회 및 환불 신청 모달 구현.

## 2. Acceptance Criteria (수용 기준)
1. 프로모션 코드 적용 시 할인 금액과 최종 결제액이 서버 견적과 100% 일치.
2. 환불 완료 즉시 라이브러리에서 상태가 `환불됨(다운로드 불가)`로 전환.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 결제 도중 창 닫힘 발생 시 주문 상태 폴링으로 중복 결제 없이 기존 미결 주문을 이어서 처리.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-COM-001;TC-COM-006`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-WEB-003/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

# AM-BILLING-002 — 환불 및 구매 권한 회수 연동, 14일 에스크로 해제 배치 및 월간 개발자 정산(Dual-Control) 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-BILLING-002` |
| **작업명 (Title)** | 환불 및 구매 권한 회수 연동, 14일 에스크로 해제 배치 및 월간 개발자 정산(Dual-Control) 구현 |
| **목표 (Objective)** | `REQ-M-003` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Backend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-BILLING-001;AM-SEC-001` |
| **개발 Phase 및 우선순위** | `Phase 7` / `P1` |
| **입력 문서 (Input Documents)** | `10_COMMERCE_ADS_ANALYTICS/PAYMENT_SETTLEMENT.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/commerce/settlement/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/commerce/settlement/refund_service.go;backend/internal/modules/commerce/settlement/escrow_release_worker.go;backend/internal/modules/commerce/settlement/monthly_payout_service.go` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/commerce/orders/{orderId}/refund;GET /api/v1/developer/orgs/{orgId}/settlements;POST /api/v1/admin/settlements/{settlementId}/approve-payout` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-COM-006` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-BILLING-002/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 미다운로드 48시간 이내 자동 환불 및 다운로드 후 심층 환불 처리, 환불 완료 즉시 `user_entitlements` 회수(`REVOKED_REFUNDED`) 및 역분개 기장.
2. 구매 확정 14일 경과 건을 `LIAB:DEV_ESCROW`에서 `LIAB:DEV_SETTLEABLE`로 전환하는 일일 워커 구현.
3. 월간 정산 명세서 생성 및 관리자 2인 이중 승인(`Dual-Control`) 지급 집행 구현.

## 2. Acceptance Criteria (수용 기준)
1. 환불 승인 즉시 해당 사용자의 유료 앱 재다운로드 및 업데이트 토큰 발급 차단.
2. 2인 이중 승인 없이는 정산 상태가 `PAID_COMPLETED`로 전환될 수 없음.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 차지백 발생 시 해당 개발자 조직의 차기 정산 예정액에서 역분개 차감하고 음수 잔액 발생 시 관리자 경보 발령.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-COM-006`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-BILLING-002/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

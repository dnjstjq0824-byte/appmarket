# AM-BILLING-001 — 다통화 가격/세금/수수료 계산기, 유료 앱 주문, PG 웹훅 검증 및 복식부기 원장 엔진 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-BILLING-001` |
| **작업명 (Title)** | 다통화 가격/세금/수수료 계산기, 유료 앱 주문, PG 웹훅 검증 및 복식부기 원장 엔진 구현 |
| **목표 (Objective)** | `REQ-B-001;REQ-M-001;REQ-M-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Backend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-DB-002;AM-API-002` |
| **개발 Phase 및 우선순위** | `Phase 7` / `P1` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/B_BUSINESS_MODEL.md;01_A_TO_Z_BLUEPRINT/M_MARKETPLACE_COMMERCE.md;10_COMMERCE_ADS_ANALYTICS/PAYMENT_SETTLEMENT.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/commerce/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/commerce/domain/pricing_calculator.go;backend/internal/modules/commerce/application/order_service.go;backend/internal/modules/commerce/application/ledger_service.go;backend/internal/modules/commerce/infrastructure/pg_webhook_verifier.go` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/commerce/quotes;POST /api/v1/commerce/orders;POST /api/v1/commerce/webhooks/{provider};GET /api/v1/users/me/entitlements` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-COM-001;TC-COM-002;TC-COM-004;TC-COM-005` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-BILLING-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 정수형 Minor Unit 기반 공급가액·세액·수수료(15%/10%/0%)·개발자 수익 계산기 및 프로모션 리딤 코드 원자적 차감 구현.
2. 주문 생성(`Idempotency-Key`), PG 웹훅 `HMAC-SHA256` 검증, 구매 권한(`user_entitlements`) 부여 및 복식부기 원장(`ledger_transactions`, `ledger_entries`) 분개 구현.

## 2. Acceptance Criteria (수용 기준)
1. 동일 결제 완료 웹훅 20회 중복 전송 시 단 1건만 원장에 기장되고 대차평형(`Debit == Credit`) 100% 유지.
2. 유료 앱 미구매 계정의 다운로드 토큰 발급 요청 시 `403 ERR_ENTITLEMENT_REQUIRED` 반환.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 웹훅 처리 실패 시 PG사에 `500`을 반환하여 PG 재전송을 유도하고 내부 DLQ에도 동시 보존.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-COM-001;TC-COM-002;TC-COM-004;TC-COM-005`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-BILLING-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

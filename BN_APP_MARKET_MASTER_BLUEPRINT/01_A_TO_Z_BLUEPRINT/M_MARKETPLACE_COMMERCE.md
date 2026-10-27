# M — Marketplace Commerce — 주문, 결제 검증, 구매 권한(Entitlement), 환불 및 복식부기 정산 원장 시스템 (Orders, Payment Verification, Entitlements, Refunds & Double-Entry Settlement Ledger)

- **문서 ID**: `DOC-AZ-M`
- **알파벳 코드**: `M`
- **관련 요구사항 ID**: `REQ-M-001`, `REQ-M-002`, `REQ-M-003`
- **연관 Work Order ID**: `AM-BILLING-001`, `AM-BILLING-002`
- **연관 테스트 케이스 ID**: `TC-COM-004`, `TC-COM-005`, `TC-COM-006`
- **참조 상세 설계 문서**: [10_COMMERCE_ADS_ANALYTICS/PAYMENT_SETTLEMENT.md](../10_COMMERCE_ADS_ANALYTICS/PAYMENT_SETTLEMENT.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 유료 앱 구매와 개발자 수익 정산은 금전이 오가는 핵심 상거래 영역이다(제12조). 중복 결제, 웹훅 위조, 환불 악용(결제 후 APK만 다운받고 즉시 환불), 원장 불일치를 원천 차단하기 위해 은행 수준의 멱등성 제어와 복식부기 원장(Double-Entry Ledger) 기반 상거래 엔진이 필요하다.
- **달성 목표**: 주문 생성부터 PG 결제 승인 검증, 구매 권한(`user_entitlements`) 부여, 환불 정책 자동/수동 심사, 구매 후 14일 정산 보류(Escrow Hold), 월간 개발자 정산 배치 및 세무 감사 체계를 명세한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **일반 사용자**: 유료 앱 결제, 영수증 조회, 구매한 앱의 다중 기기 재다운로드, 환불 신청.
- **개발자 Finance Manager**: 실시간 매출 및 주문 내역 조회, 정산 계좌 관리, 월별 정산서 확인.
- **플랫폼 Finance Admin**: 이상 결제/차지백 모니터링, 분쟁 환불 강제 승인, 월간 정산금 지급(Payout) 이중 승인.

## 3. 필수 기능 (Mandatory Features)
1. **결제 어댑터 추상화 (`PaymentGatewayAdapter`)**: Stripe(글로벌 카드/Apple Pay/Google Pay), Toss Payments / PortOne(한국 간편결제/카드/계좌이체) 등 복수 PG사를 플러그인 구조로 연동하며, PG사 웹훅 서명(`HMAC-SHA256`) 검증을 필수로 수행한다.
2. **구매 권한(`user_entitlements`) 연동 다운로드 통제**:
   - 유료 앱(`monetization_type = PAID_ONCE`)의 다운로드 토큰 발급 API는 반드시 `user_entitlements` 테이블에서 해당 사용자의 `status = 'ACTIVE'` 권한을 확인한 뒤에만 CDN Signed URL을 발급한다.
3. **불변 복식부기 원장 (`ledger_transactions` & `ledger_entries`)**:
   - 모든 금전 이벤트(결제 완료, 플랫폼 수수료 인식, 세금 예수금, 환불 차감, 개발자 정산 지급)는 반드시 차변(`DEBIT`) 합계와 대변(`CREDIT`) 합계가 정확히 일치하는 원장 엔트리로만 기록되며, 기존 엔트리의 `UPDATE`나 `DELETE`는 금지되고 오직 역분개(Reversing Entry)로만 수정된다.
4. **환불 및 권한 회수 연동**:
   - 구매 후 48시간 이내 미다운로드 건은 즉시 자동 환불 승인 가능하며, 이미 APK를 다운로드한 건은 결함 증빙 심사 또는 정책 기준에 따라 처리한다.
   - 환불 또는 차지백(`CHARGEBACK`) 완료 즉시 `user_entitlements.status`를 `REVOKED_REFUNDED`로 전환하여 향후 업데이트 및 재다운로드를 차단한다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **결제 완료 시 복식부기 분개 예시 (₩11,000 유료 앱 구매: 공급가 ₩10,000, 부가세 ₩1,000, 수수료 15% = ₩1,500)**:
  - `DEBIT` `ACCT_PG_RECEIVABLE` (PG 미수금 자산): `+11,000`
  - `CREDIT` `ACCT_TAX_PAYABLE` (부가세 예수금 부채): `+1,000`
  - `CREDIT` `ACCT_PLATFORM_REVENUE` (플랫폼 수수료 수익): `+1,500`
  - `CREDIT` `ACCT_DEV_ESCROW_PAYABLE:{org_id}` (개발자 정산보류 예수금 부채): `+8,500`
- **14일 정산 보류 후 정산 가능 전환 (Settlement Release)**:
  - 구매 확정 후 14일(`Hold Period`) 동안 환불/차지백이 발생하지 않은 건은 일일 배치 워커가 `ACCT_DEV_ESCROW_PAYABLE:{org_id}`에서 `ACCT_DEV_SETTLEABLE_PAYABLE:{org_id}`(정산 가능 잔액)로 대체 분개한다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **`web-user`**: `CheckoutView.vue`, `PurchaseHistoryView.vue`, `RefundRequestModal.vue`.
- **`web-dev`**: `OrdersTableView.vue`, `SettlementStatementsView.vue`.
- **`web-admin`**: `FinanceLedgerAuditView.vue`, `PayoutApprovalView.vue`.

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/commerce/`:
  - `order_service.go`, `ledger_service.go`, `entitlement_service.go`, `settlement_batch_worker.go`.
  - 주문 생성 및 결제 승인 API는 `Idempotency-Key` 헤더와 DB `UNIQUE(user_id, app_id) WHERE status = 'ACTIVE'` 제약조건으로 중복 결제를 원천 방지한다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**: `orders`, `payments`, `refunds`, `user_entitlements`, `ledger_accounts`, `ledger_transactions`, `ledger_entries`, `settlements`, `settlement_line_items`.
- **무결성 제약**: `ledger_entries` 삽입 트랜잭션 커밋 시 `SUM(CASE WHEN direction='DEBIT' THEN amount_minor ELSE -amount_minor END) = 0`을 검증하는 `CONSTRAINT TRIGGER` 적용.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `POST /api/v1/commerce/orders` (`Idempotency-Key` 필수): 주문 생성 및 PG 결제 세션 반환.
- `POST /api/v1/commerce/webhooks/{provider}`: PG사 결제/환불 웹훅 수신 (HMAC 서명 검증).
- `POST /api/v1/commerce/orders/{orderId}/refund`: 환불 요청.
- `GET /api/v1/users/me/entitlements`: 내 구매 앱 및 다운로드 권한 목록 조회.
- `POST /api/v1/admin/settlements/{settlementId}/approve`: 정산 지급 승인 (Dual-Control 필수).

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **주문 상태 (`order_status`)**: `CREATED` → `PAYMENT_PENDING` → `PAID_ENTITLED` → (`REFUND_REQUESTED` → `REFUNDED` | `REFUND_REJECTED`) | `CHARGEBACK_LOST` | `EXPIRED`.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- PCI-DSS 준수를 위해 카드번호(PAN), CVV 등 민감 결제정보는 당사 서버를 절대 거치지 않고 PG사 Hosted Checkout / Tokenization SDK로만 처리한다.
- 웹훅 재전송 공격(Replay Attack) 방지를 위해 웹훅 타임스탬프 허용 오차(5분)와 `provider_event_id UNIQUE` 테이블 기록을 강제한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 유료 앱 결제 전 화면에 환불 가능 조건, 지원 기기 호환성 확인 여부, 부가세 포함 최종 결제 금액, 개발자 상호 및 연락처를 전자상거래법 기준에 맞게 명시한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **중복 결제 및 웹훅 중복 수신 방어 테스트 (`TC-COM-004`)**: 동일 주문에 대해 PG 결제 완료 웹훅을 동시에 20번 전송했을 때 `ledger_transactions`와 `user_entitlements`가 정확히 1건만 생성되는지 검증.
- **복식부기 대차평형 무결성 테스트 (`TC-COM-005`)**: 차변과 대변이 1원이라도 다른 분개 삽입 시도 시 PostgreSQL 트랜잭션이 즉시 롤백되는지 검증.
- **미구매자 유료 앱 다운로드 차단 테스트 (`TC-COM-006`)**: 유료 앱을 구매하지 않았거나 이미 환불(`REVOKED_REFUNDED`)된 계정이 다운로드 토큰 발급 API를 호출할 때 `403 ERR_ENTITLEMENT_REQUIRED`로 차단되는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `B — Business Model`, `I — Identity & Access`, `N — Native App Distribution`, `Y — Yield & Analytics`와 연동된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P1`
- **개발 Phase**: `Phase 7 (결제, 정산, 광고)`에서 집중 구현한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 사용자는 PG 결제를 완료했으나 브라우저 종료로 리다이렉트가 누락된 경우** → PG 서버 간 비동기 웹훅(Server-to-Server Webhook)과 5분 주기 미결 주문 상태 대사(Reconciliation Poller)가 동작하여 누락 없이 구매 권한을 자동 부여한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 상거래·결제·복식부기 원장·정산 모듈 (`backend/internal/modules/commerce/*`).

## 17. Acceptance Criteria (수용 기준)
1. 모든 결제·환불·정산 시나리오 실행 후 `SELECT SUM(CASE WHEN direction='DEBIT' THEN amount_minor ELSE -amount_minor END) FROM ledger_entries;`의 결과가 항상 `0`이어야 한다.
2. 유료 앱 환불 완료 즉시 해당 사용자의 다운로드 토큰 발급이 차단되어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-BILLING-001/ledger_zero_balance_audit.log`: 1,000건 시뮬레이션 거래 후 복식부기 대차평형 검증 로그.
- `artifacts/evidence/AM-BILLING-002/entitlement_revocation_test.log`: 환불 후 유료 APK 다운로드 차단 통합 테스트 로그.

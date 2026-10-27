# B — Business Model — 글로벌 비즈니스 모델, 수수료, 구독, 프로모션 및 수익 배분 체계 (Global Business Model, Commission Tiers, Subscriptions, Promotions & Revenue Share)

- **문서 ID**: `DOC-AZ-B`
- **알파벳 코드**: `B`
- **관련 요구사항 ID**: `REQ-B-001`, `REQ-B-002`
- **연관 Work Order ID**: `AM-BILLING-001`, `AM-BILLING-002`, `AM-ADS-001`
- **연관 테스트 케이스 ID**: `TC-COM-001`, `TC-COM-002`, `TC-COM-003`
- **참조 상세 설계 문서**: [10_COMMERCE_ADS_ANALYTICS/PAYMENT_SETTLEMENT.md](../10_COMMERCE_ADS_ANALYTICS/PAYMENT_SETTLEMENT.md), [10_COMMERCE_ADS_ANALYTICS/ADVERTISING_ENGINE.md](../10_COMMERCE_ADS_ANALYTICS/ADVERTISING_ENGINE.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 독립 앱스토어가 지속 가능한 상용 플랫폼으로 운영되기 위해서는 무료 앱 배포뿐만 아니라 유료 앱 판매, 인앱 구독 확장성, 개발자 친화적 차등 수수료 정책, 프로모션/할인 쿠폰, 투명한 스폰서 검색/배너 광고 수익 모델을 정밀한 수학적·회계적 규칙으로 정의해야 한다.
- **달성 목표**: 글로벌 다통화 가격 책정(Pricing Matrix), 수수료 계산 공식, 세금(VAT/GST) 공제 규칙, 프로모션 코드 적용 우선순위, 개발자 정산 배분율을 코드와 DB 스키마로 명세한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **일반 사용자 (Buyer)**: 무료 앱 즉시 다운로드, 유료 앱 일회성 구매(Paid App Purchase), 프로모션 코드 적용, 환불 요청 수행.
- **개발자 조직 Finance Manager**: 앱의 유료/무료 전환, 국가별 기준 가격 설정, 프로모션 캠페인 생성, 수수료 티어 확인, 월별 정산 명세서(Payout Statement) 조회.
- **플랫폼 Finance Admin**: 수수료 티어 정책 관리, 국가별 세율(`tax_rules`) 업데이트, 대량 환불 및 월간 정산 배치 최종 승인.

## 3. 필수 기능 (Mandatory Features)
1. **앱 과금 모델 분류 (`monetization_type`)**: `FREE`(완전 무료), `FREE_WITH_ADS`(무료+광고 포함 표시), `FREE_WITH_IAP`(무료+인앱결제 안내), `PAID_ONCE`(유료 다운로드), `SUBSCRIPTION_READY`(정기 구독 상품 연결).
2. **차등 수수료 엔진 (Tiered Commission Engine)**:
   - 기본 플랫폼 수수료(`STANDARD_TIER`): 결제 총액(세금 제외 공급가액 기준)의 **15.00%**.
   - 중소/인디 개발자 우대 티어(`SMALL_BIZ_TIER`, 연간 누적 매출 USD $1,000,000 이하 승인 조직): **10.00%**.
   - 오픈소스/전략 파트너 티어(`PARTNER_ZERO_TIER`): **0.00% ~ 5.00%** (PG 실비만 차감).
3. **프로모션 및 리딤 코드 (Promo & Redeem Codes)**: 기간 한정 할인율(10%~90%) 및 1회용 유료 앱 무료 리딤 코드 발급·검증.
4. **스폰서 앱 광고 수익 모델**: CPC(Cost-Per-Click) 및 CPT(고정 구좌 기간제) 방식의 검색 상단/카테고리 추천 스폰서 슬롯 과금(반드시 `Sponsored` 배지 명시).

## 4. 상세 기능 (Detailed Functional Specifications)
- **정산 금액 산출 공식 (결정론적 정수/고정소수점 연산)**:
  - 모든 금액 연산은 부동소수점(`float64`) 사용을 엄격히 금지하며, 최소 통화 단위(Minor Unit, 예: KRW는 1원, USD는 1센트 `BIGINT` 또는 `NUMERIC(18,4)`)와 `ROUND_HALF_EVEN`(Banker's Rounding)을 사용한다.
  - `GrossAmount` (사용자 결제 금액) - `DiscountAmount` (프로모션 할인액, 개발자 부담 또는 플랫폼 부담 구분) = `PaidAmount` (실결제액).
  - `NetSupplyAmount` (공급가액) = `PaidAmount` - `TaxAmount` (국가별 VAT/GST).
  - `PlatformFee` = `ROUND(NetSupplyAmount * CommissionRate)`.
  - `PgProcessingFee` = PG사 수수료 (기본 플랫폼 수수료 내 포함 또는 별도 패스스루 정책을 `commission_policies` 테이블의 `pg_fee_borne_by` 컬럼으로 제어).
  - `DeveloperPayoutAmount` = `NetSupplyAmount` - `PlatformFee`.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **`web-user`**: 앱 상세 페이지의 가격 배지(원가, 할인가, VAT 포함 여부 명시), 프로모션 코드 입력 모달, 외부 스토어와의 가격 비교 테이블(`PriceComparisonMatrix.vue`).
- **`web-dev`**: 가격 및 배포 국가 설정 탭(`AppPricingEditor.vue`), 수수료 티어 현황 및 예상 수령액 시뮬레이터(`PayoutSimulator.vue`), 프로모션 코드 생성기(`PromoCodeManager.vue`).
- **`web-admin`**: 수수료 정책 관리(`CommissionPolicyAdmin.vue`), 세율 관리(`TaxRuleAdmin.vue`).

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/commerce/domain/pricing_calculator.go`: 통화별 소수점 자리수(ISO 4217: KRW/JPY 0자리, USD/EUR 2자리), 세율, 수수료 티어를 입력받아 `OrderQuote` 구조체를 생성하는 순수 함수 도메인 엔진.
- `internal/modules/commerce/application/promo_service.go`: 리딤 코드 원자적 차감(`UPDATE promo_codes SET used_count = used_count + 1 WHERE code = $1 AND used_count < max_uses`) 처리.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**:
  - `commission_policies` (`id`, `policy_code`, `commission_bps INT CHECK (commission_bps BETWEEN 0 AND 5000)`, `pg_fee_borne_by`, `is_active`).
  - `app_price_tiers` (`id`, `app_id`, `country_code CHAR(2)`, `currency CHAR(3)`, `list_price_minor BIGINT`, `tax_inclusive BOOLEAN`).
  - `promotions` 및 `promo_codes` (`id`, `app_id`, `code_hash`, `discount_type`, `discount_value`, `starts_at`, `ends_at`, `max_uses`, `used_count`).

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `POST /api/v1/commerce/quotes`: 앱 ID, 국가, 통화, 프로모션 코드를 전달받아 세금·할인이 반영된 최종 결제 견적(`OrderQuote`) 반환.
- `PUT /api/v1/developer/apps/{appId}/pricing`: 국가별 가격표 일괄 설정.
- `POST /api/v1/developer/apps/{appId}/promotions`: 프로모션 및 리딤 코드 배치 생성.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **유료→무료 / 무료→유료 전환 규칙**:
  - 유료 앱을 무료로 전환하거나 가격을 변경할 경우 진행 중인 결제 견적(`OrderQuote`)은 15분간 기존 가격을 보장하며, 이미 유료로 구매한 사용자의 영구 다운로드 권한(`user_entitlements`)은 앱이 유료/무료로 어떻게 바뀌든 영구 유지된다.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 클라이언트(웹/Android)에서 전달한 가격(`price`) 파라미터는 절대 신뢰하지 않으며, 서버가 DB의 `app_price_tiers`와 유효한 `promo_codes`를 직접 조회하여 결제 금액을 산출한다(Price Tampering 원천 차단).
- 리딤 코드 무차별 대입 공격(Brute-force) 방지를 위해 사용자당 분당 5회 실패 시 1시간 동안 코드 입력을 차단한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 유료 앱 구매 버튼에는 반드시 통화 기호와 세금 포함 여부(예: `₩5,500 (부가세 포함)`)를 명확히 표시하고, 인앱결제 또는 광고 포함 시 구매 버튼 하단에 `인앱 구매 포함` / `광고 포함` 레이블을 의무 표시한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **수수료 및 세금 계산 정밀도 단위 테스트 (`TC-COM-001`)**: KRW, USD, EUR, JPY 4개 통화와 0%, 10%, 15% 수수료 티어, 부분 환불 케이스에 대해 단 1센트/1원의 반올림 오차도 발생하지 않음을 검증.
- **리딤 코드 동시성 테스트 (`TC-COM-002`)**: `max_uses = 1`인 리딤 코드에 50개의 동시 결제 요청을 전송했을 때 정확히 1건만 성공하고 49건은 `ERR_PROMO_CODE_EXHAUSTED`로 실패하는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `C — Catalog Engine`(앱 상품 정보), `M — Marketplace Commerce`(결제·원장), `G — Globalization`(다통화·국가별 세법), `Y — Yield & Analytics`(매출 분석)와 연동된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P1 (상거래 및 수익화 핵심)`
- **개발 Phase**: 무료 앱 모델은 `Phase 3`, 유료 결제·수수료·정산·광고는 `Phase 7`에서 구현한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 환율 API 연동 실패 시 글로벌 기준가 변환 오류** → 실시간 외부 환율에 의존해 결제 금액을 매번 바꾸지 않고, 개발자가 승인한 국가별 고정 가격표 스냅샷(`app_price_tiers`)을 기준으로 결제하여 외부 환율 장애의 영향을 받지 않도록 설계한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 가격/수수료/세금 계산 도메인 패키지 (`backend/internal/modules/commerce/domain/pricing*`).
- 개발자 가격/프로모션 관리 UI 및 견적 API 핸들러.

## 17. Acceptance Criteria (수용 기준)
1. 국가별·통화별 가격 설정 후 견적 API 호출 시 정확한 공급가액, 세액, 플랫폼 수수료, 개발자 정산 예정액이 계산되어야 한다.
2. 만료되었거나 사용 횟수를 초과한 프로모션 코드는 즉시 거부되어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-BILLING-001/pricing_matrix_test.log`: 다통화/수수료 티어별 정산 수식 검증 로그.
- `artifacts/evidence/AM-BILLING-001/promo_concurrency_test.log`: 동시 요청 50건 중 단일 성공 검증 결과.

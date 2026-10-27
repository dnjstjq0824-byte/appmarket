# PAYMENT_SETTLEMENT.md — 유료 앱 결제, 환불, 복식부기 원장 및 개발자 정산 상세 설계

- **문서 ID**: `DOC-COM-001`
- **관련 조항**: 제3조 B·M항, 제12조
- **연관 Work Order**: `AM-BILLING-001`, `AM-BILLING-002`

---

## 1. 복식부기 원장 계정 과목표 (Chart of Accounts — `ledger_accounts`)

| 계정 코드 (`account_code`) | 계정 유형 (`account_type`) | 증가 방향 | 설명 |
| :--- | :---: | :---: | :--- |
| `ASSET:PG_RECEIVABLE:{provider}` | `ASSET` (자산) | `DEBIT` (차변) | Stripe/Toss 등 PG사로부터 입금받을 결제 대금 미수금 |
| `ASSET:PLATFORM_BANK:{currency}` | `ASSET` (자산) | `DEBIT` (차변) | 플랫폼 법인 계좌 현금 및 정산 지급 출금 계좌 |
| `LIAB:TAX_PAYABLE:{country}` | `LIABILITY` (부채) | `CREDIT` (대변) | 국가별 과세당국에 납부할 부가가치세(VAT/GST) 예수금 |
| `LIAB:DEV_ESCROW:{org_id}:{currency}` | `LIABILITY` (부채) | `CREDIT` (대변) | 구매 후 14일 환불 보류 기간 동안 묶여 있는 개발자 정산 대기금 |
| `LIAB:DEV_SETTLEABLE:{org_id}:{currency}` | `LIABILITY` (부채) | `CREDIT` (대변) | 14일 경과 후 월간 정산 지급이 가능해진 개발자 확정 미지급금 |
| `LIAB:AD_PREPAID:{org_id}:{currency}` | `LIABILITY` (부채) | `CREDIT` (대변) | 광고주가 충전한 스폰서 광고 예치금 |
| `REV:PLATFORM_COMMISSION:{currency}` | `REVENUE` (수익) | `CREDIT` (대변) | 유료 앱 판매 중개 수수료 수익 (기본 15% / 우대 10%) |
| `REV:PLATFORM_AD_REVENUE:{currency}` | `REVENUE` (수익) | `CREDIT` (대변) | 스폰서 광고 유효 클릭/노출 과금 수익 |

---

## 2. 4대 핵심 회계 이벤트별 분개 규칙 (Journal Entry Rules)

1. **유료 앱 결제 완료 (`ORDER_PAYMENT`, 예: 총 결제 `11,000 KRW` = 공급가 `10,000` + 세금 `1,000`, 수수료 15% = `1,500`)**:
   - `DEBIT` `ASSET:PG_RECEIVABLE:TOSS` = `11,000`
   - `CREDIT` `LIAB:TAX_PAYABLE:KR` = `1,000`
   - `CREDIT` `REV:PLATFORM_COMMISSION:KRW` = `1,500`
   - `CREDIT` `LIAB:DEV_ESCROW:{org_id}:KRW` = `8,500`
2. **보류 기간 내 전액 환불 (`ORDER_REFUND`)**:
   - 위 1번 분개의 정확히 반대 방향(`CREDIT` 자산 `11,000` / `DEBIT` 부채·수익 합계 `11,000`)으로 역분개 트랜잭션을 생성하고 `user_entitlements.status = 'REVOKED_REFUNDED'`로 즉시 전환합니다.
3. **14일 경과 에스크로 해제 (`ESCROW_RELEASE`)**:
   - `DEBIT` `LIAB:DEV_ESCROW:{org_id}:KRW` = `8,500`
   - `CREDIT` `LIAB:DEV_SETTLEABLE:{org_id}:KRW` = `8,500`
4. **월간 정산 지급 완료 (`SETTLEMENT_PAYOUT`, Dual-Control 2인 승인 필수)**:
   - `DEBIT` `LIAB:DEV_SETTLEABLE:{org_id}:KRW` = `8,500`
   - `CREDIT` `ASSET:PLATFORM_BANK:KRW` = `8,500`

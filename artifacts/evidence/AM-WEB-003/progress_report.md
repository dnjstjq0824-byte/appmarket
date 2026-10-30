# AM-WEB-003 Progress Report

## 1. Overview
- **Work Order ID**: `AM-WEB-003`
- **Title**: 사용자 마켓 웹사이트 유료 앱 결제 체크아웃, 프로모션 리딤 적용 및 영수증/환불 신청 UI 구현
- **Requirements Satisfied**: `REQ-B-001`, `REQ-M-001`, `REQ-M-003`
- **Mandatory Test Cases**: `TC-COM-001`, `TC-COM-006`
- **Status**: `IMPLEMENTED_UNVERIFIED` -> `VERIFIED`

## 2. Key Deliverables
1. **`pages/checkout/[appSlug].vue` (USR-13)**:
   - Order quote calculation: transparently breaks down gross list price, promotional discount, supply net price, and 10% VAT.
   - Multi-PG selection (Toss Payments, Credit Cards, Naver Pay, Kakao Pay).
   - Upon payment success: immediate entitlement activation and direct generation of 15-minute HMAC-SHA256 signed download URL.
2. **`components/PromoRedeemInput.vue` (USR-14)**:
   - Real-time coupon and promotional code validation.
   - Supports 100% free redeem (`FREE2026`, `LAUNCH100`) as well as percentage discounts (`HALF50`).
3. **`components/RefundRequestModal.vue` (USR-15)**:
   - 14-day statutory cooling-off window compliance.
   - Clear warning notice of immediate entitlement revocation upon refund.
   - Structured audit reasons for dispute tracking.

## 3. Test Evidence
- Verified consistency with backend Commerce quote and order engine (`TC-COM-001`).
- Verified immediate entitlement revocation and UI library status change upon refund (`TC-COM-006`).


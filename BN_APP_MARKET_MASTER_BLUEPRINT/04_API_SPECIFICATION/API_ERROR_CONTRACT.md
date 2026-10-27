# API_ERROR_CONTRACT.md — RFC 9457 표준 에러 스키마 및 도메인 에러 코드 카탈로그

- **문서 ID**: `DOC-API-002`
- **관련 문서**: [openapi.yaml](openapi.yaml)

---

## 1. 표준 에러 응답 구조 (RFC 9457 Problem Details for HTTP APIs)

모든 API 오류 응답(`4xx`, `5xx`)은 `Content-Type: application/problem+json` 헤더와 함께 아래 JSON 구조를 엄격히 준수합니다:

```json
{
  "type": "https://api.appmarket.bookingnetwork.internal/errors/ERR_APK_SIGNATURE_MISMATCH",
  "title": "APK signing certificate does not match existing app lineage",
  "status": 422,
  "code": "ERR_APK_SIGNATURE_MISMATCH",
  "detail": "Expected certificate SHA-256 [a1b2...], but uploaded APK is signed with [9f8e...] without valid v3 proof-of-rotation.",
  "trace_id": "7f3c9a12b4e648109c12a34b56d78e90",
  "timestamp": "2026-10-27T14:30:00Z",
  "field_errors": []
}
```

---

## 2. 표준 도메인 에러 코드 사전 (Canonical Error Code Registry)

| 에러 코드 (`code`) | HTTP 상태 | 발생 도메인 | 발생 조건 및 클라이언트 처리 지침 |
| :--- | :---: | :--- | :--- |
| `ERR_AUTH_INVALID_CREDENTIALS` | `401` | IAM | 이메일 또는 비밀번호 불일치 (남은 시도 횟수 안내) |
| `ERR_TOKEN_REUSE_DETECTED` | `401` | IAM | 이미 회전된 Refresh Token 재사용 감지 (세션 패밀리 전체 폐기 및 재로그인 강제) |
| `ERR_MFA_REQUIRED` | `403` | IAM | 관리자 또는 출시/재무 개발자가 MFA 미설정 또는 `X-MFA-TOTP-Code` 누락 |
| `ERR_FORBIDDEN_TENANT_ACCESS` | `403` | Developer | 소속되지 않은 개발자 조직(`orgId`) 또는 타 조직 소유 앱(`appId`) 접근 시도 |
| `ERR_DEV_ORG_NOT_VERIFIED` | `422` | Developer | KYB 신원 인증이 `VERIFIED`가 아닌 상태에서 프로덕션 심사 제출 시도 |
| `ERR_APK_V2_SIGNATURE_REQUIRED` | `422` | Distribution | APK Signature Scheme v2/v3 블록이 없는 v1 전용 또는 미서명 APK 업로드 |
| `ERR_APK_SIGNATURE_MISMATCH` | `422` | Distribution | 기존 출시된 앱의 서명 인증서 SHA-256과 업로드된 APK의 서명 인증서가 불일치 |
| `ERR_APK_VERSION_CODE_NOT_INCREASING` | `422` | Distribution | 업로드된 APK의 `versionCode`가 해당 트랙의 기존 최대 `versionCode` 이하임 |
| `ERR_ZIP_BOMB_DETECTED` | `422` | Security | 압축 해제 비율 100:1 초과 또는 비정상 ZIP 헤더 탐지 |
| `ERR_INVALID_RELEASE_STATE_TRANSITION` | `409` | Release | 보안 스캔 미완료 또는 반려 상태에서 잘못된 상태 전이 요청 |
| `ERR_DIRECT_DOWNLOAD_NOT_PERMITTED` | `403` | Distribution | `EXTERNAL_LINK_ONLY` 권리 등급 앱에 대해 직접 APK 다운로드 토큰 요청 |
| `ERR_ENTITLEMENT_REQUIRED` | `403` | Commerce | 유료 앱(`PAID_ONCE`)을 구매하지 않았거나 환불된 상태에서 다운로드 요청 |
| `ERR_APP_UNAVAILABLE_IN_COUNTRY` | `451` | Globalization | 요청자의 국가에서 개발자 설정 또는 법적 규제로 차단된 앱 접근 |
| `ERR_REVIEW_REQUIRES_INSTALLATION` | `403` | Review | 해당 앱의 실제 다운로드/설치 기록이 없는 계정의 리뷰 작성 시도 |
| `ERR_PROMO_CODE_EXHAUSTED` | `409` | Commerce | 프로모션 코드 사용 한도(`max_uses`) 소진 또는 기간 만료 |
| `ERR_IDEMPOTENCY_KEY_CONFLICT` | `409` | Platform | 동일한 `Idempotency-Key`로 이전 요청과 다른 본문(Payload Hash 불일치) 전송 |
| `ERR_DUAL_CONTROL_SELF_APPROVAL_FORBIDDEN` | `403` | Admin | 이중 승인 요청 생성자(`initiator_id`) 본인이 스스로 승인 시도 |
| `ERR_APPEAL_SAME_REVIEWER_FORBIDDEN` | `403` | Moderation | 최초 반려 처분을 내린 심사자가 해당 건의 이의신청을 직접 심사 시도 |
| `ERR_TREESCAN_ADAPTER_DISABLED` | `503` | TreeScan | `TREESCAN_INTEGRATION_ENABLED=false` (독립 모드) 상태에서 TreeScan 어댑터 호출 |

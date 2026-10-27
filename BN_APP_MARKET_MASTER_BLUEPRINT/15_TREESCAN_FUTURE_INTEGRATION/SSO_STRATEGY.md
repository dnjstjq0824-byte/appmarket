# SSO_STRATEGY.md — RFC 8693 OAuth 2.0 Token Exchange 기반 미래 TreeScan SSO 연동 설계

- **문서 ID**: `DOC-TS-003`
- **관련 문서**: [01_A_TO_Z_BLUEPRINT/I_IDENTITY_AND_ACCESS.md](../01_A_TO_Z_BLUEPRINT/I_IDENTITY_AND_ACCESS.md)

---

## 1. 독립 계정 유지 + OIDC 연합(Federation) 아키텍처

1. 앱 마켓은 자체 `users` 및 `auth_sessions` 테이블을 영구적으로 유지하므로 TreeScan 인증 서버가 점검 중이거나 분리되더라도 자체 로그인과 기존 세션이 100% 정상 동작합니다.
2. 향후 TreeScan 통합이 활성화되면 `user_identities` 테이블의 `provider = 'TREESCAN_OIDC'` 슬롯을 사용하여:
   - 클라이언트가 TreeScan에서 발급받은 ID Token(`subject_token`)을 `POST /api/v1/auth/token/exchange`로 제출합니다.
   - 앱 마켓 백엔드는 TreeScan JWKS 공개키로 서명과 `aud`, `exp`, `iss`를 검증한 뒤, 매핑된 로컬 `users` 레코드를 찾거나 자동 프로비저닝(Just-In-Time Provisioning)하여 앱 마켓 자체 Access/Refresh 토큰을 발급합니다.

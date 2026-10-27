# SECURITY_ARCHITECTURE.md — OWASP ASVS / MASVS 기반 플랫폼 보안 아키텍처

- **문서 ID**: `DOC-SEC-001`
- **관련 조항**: 제13조 (보안, 운영 정책 및 법률 대응)
- **참조 표준**: [OWASP ASVS v4.0 Level 2](https://owasp.org/projects/asvs), [OWASP MASVS v2.0](https://mas.owasp.org/MASVS/)

---

## 1. 심층 방어 (Defense-in-Depth) 5계층 통제 체계

1. **Edge & Network Layer**: TLS 1.3 전용(TLS 1.0/1.1 및 약한 암호화 스위트 비활성화), HSTS(`max-age=63072000; includeSubDomains; preload`), WAF(SQLi/XSS/Bot 차단), 격리 워커 서브넷 외부 Egress 전면 차단.
2. **Identity & Access Layer**: Argon2id 패스워드 해싱, RS256 단기 JWT(15분) + Refresh Token Rotation(재사용 즉시 패밀리 폐기), 관리자 및 출시 개발자 TOTP MFA 의무화, 민감 운영 액션 2인 이중 승인(`Dual-Control`).
3. **Artifact & Supply Chain Layer**: 업로드 바이너리 격리 저장(`bn-quarantine-bucket`), SHA-256 검증, APK Signature Scheme v2/v3 강제, 서명 인증서 연속성 고정, ClamAV + YARA + 권한 위험도 정적 분석, 불변 CDN 경로(`/{sha256}.apk`).
4. **Data & Ledger Layer**: PII 필드(계좌번호, 사업자번호, MFA 시크릿) KMS `AES-256-GCM` 암호화, `audit_logs` 및 `ledger_entries`에 대한 애플리케이션 DB 계정의 `UPDATE/DELETE` 권한 회수, 암호학적 해시 체인 감사 로그.
5. **Android Client Layer (OWASP MASVS)**: `EncryptedSharedPreferences`/Keystore 토큰 보관, 설치 직전 로컬 `SHA-256` 및 `PackageManager.GET_SIGNING_CERTIFICATES` 서명 대조, 스토어 자체 업데이트 서명 키 핀닝.

---

## 2. 보안 검사 한계 고지 원칙 (제13조 의무 사항)

플랫폼의 어떠한 화면, API 응답, 알림 메시지에서도 `"보안 검사를 통과했으므로 100% 안전을 보장합니다"`라는 문구를 사용하지 않습니다. 대신 다음과 같이 검사 범위와 시점을 사실대로 고지합니다:
> `"본 APK 파일은 [YYYY-MM-DD HH:MM UTC] 기준 당사의 자동화된 서명 무결성 검증(APK Signature v2/v3) 및 정적 악성코드 시그니처 검사(Engine vX.Y)에서 알려진 위협이 탐지되지 않았습니다. 단, 자동화된 검사가 모든 잠재적 보안 위험을 보증하지는 않으므로 요구 권한을 반드시 확인하시기 바랍니다."`

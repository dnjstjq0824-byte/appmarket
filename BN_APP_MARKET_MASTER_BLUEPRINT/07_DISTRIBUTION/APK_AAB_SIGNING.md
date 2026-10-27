# APK_AAB_SIGNING.md — APK Signature Scheme v2/v3/v4 검증 및 AAB bundletool 서명 아키텍처

- **문서 ID**: `DOC-DIST-002`
- **관련 조항**: 제10조, 제21조
- **참조 공식 문서**: [Android Open Source Project - APK Signature Scheme v2/v3/v4](https://source.android.com/docs/security/features/apksigning)

---

## 1. APK 서명 검증 규격 (Developer-Signed APK Mode — 기본 운영 경로)

1. **지원 서명 스킴**: APK Signature Scheme **v2** (필수), **v3 / v3.1** (키 회전 지원), **v4** (선택적 `.apk.idsig` 증분 설치 해시 트리).
2. **JAR 서명(v1 단독) 거부 근거**: v1 서명은 ZIP 엔트리 개별 파일만 서명하고 APK 구조 전체를 보호하지 않아 Janus 취약점(악성 DEX를 APK 앞단에 주입해도 서명이 통과되는 공격)에 노출되므로, 본 플랫폼은 **v2 또는 v3 서명 블록이 없는 APK를 100% 거부**합니다.
3. **X.509 인증서 제약조건**:
   - 최소 공개키 길이: RSA `2048-bit` 이상 또는 ECDSA `P-256` 이상.
   - 디버그 인증서(`CN=Android Debug,O=Android,C=US`) 거부.

---

## 2. AAB(Android App Bundle) 처리 및 서명 아키텍처 (Seamless Extension Interface)

제10조 요구사항에 따라 초기에는 개발자가 서명한 APK를 중심으로 운영하되, 구조 변경 없이 AAB를 처리할 수 있도록 다음 2가지 서명 소유권 모드를 인터페이스로 정의합니다:

| 구분 | 모드 A: 개발자 업로드 키 + 플랫폼 관리형 앱 서명키 (`PLATFORM_MANAGED_KEY`) | 모드 B: 사전 생성된 서명 APK 세트 업로드 (`DEVELOPER_HELD_KEY`) |
| :--- | :--- | :--- |
| **App Signing Key 소유권** | 플랫폼 KMS/HSM 보안 볼트(`bn-signing-vault`)에 앱별 전용 배포 서명키 보관 (절대 외부 반출 불가) | 개발자가 배포 서명키를 직접 보유 (자사 CI에서 `bundletool`로 빌드·서명하거나 Universal APK 업로드) |
| **업로드 바이너리** | 개발자 Upload Key로 서명된 `.aab` 파일 | 개발자 Release Key로 서명된 `.apk` (Universal 또는 ABI별 Multi-APK) |
| **`bundletool` 실행** | 격리 워커(`worker-bundletool`)가 `bundletool build-apks --bundle=in.aab --output=out.apks --ks=kms://...` 실행하여 Universal APK 및 기기 스펙별 Split APK 생성 | 불필요 (업로드된 APK의 서명만 검증하여 그대로 배포) |

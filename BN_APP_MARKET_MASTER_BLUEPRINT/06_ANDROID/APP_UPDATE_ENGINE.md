# APP_UPDATE_ENGINE.md — 기설치 앱 스캔, 업데이트 호환성 판별 및 마켓 자체 업데이트 전략

- **문서 ID**: `DOC-AND-003`
- **관련 조항**: 제5조 (항목 11, 14, 15, 17, 19)
- **연관 Work Order**: `AM-ANDROID-003`

---

## 1. 기기 설치 앱 스캔 및 서명 지문 추출 (`InstalledPackagesScanner`)

1. Android 11+(API 30) 패키지 가시성 정책에 따라 `AndroidManifest.xml`에 `<uses-permission android:name="android.permission.QUERY_ALL_PACKAGES" />`를 선언합니다(앱스토어 클라이언트는 공식 허용 대상이며, 자체 배포 APK이므로 Google Play 정책 제약 없이 정상 동작).
2. 각 설치된 패키지에 대해 `PackageManager.GET_SIGNING_CERTIFICATES` (API 28+) 또는 `GET_SIGNATURES` (API 26~27) 플래그로 서명 인증서 바이트 배열을 읽고 `MessageDigest.getInstance("SHA-256")` 지문을 계산합니다.
3. `InstallSourceInfo`를 통해 최초 설치 스토어(`installerPackageName`)를 함께 추출하여 `POST /api/v1/device/check-updates`로 전송합니다.

---

## 2. 마켓 앱 자체 업데이트 전략 (Store Self-Update Bootstrap)

1. 마켓 앱 기동 시 및 12시간 주기로 `GET /api/v1/system/capabilities`의 `client_version_policy` (`latest_version_code`, `min_supported_version_code`, `self_update_apk_url`, `self_update_sha256`, `pinned_store_signer_sha256`)를 조회합니다.
2. 현재 마켓 앱의 `BuildConfig.VERSION_CODE < min_supported_version_code`이면 강제 업데이트 모달(Force Update)을 띄우고, 최신 버전이 존재하면 설정 화면 및 상단 배너에 원클릭 업데이트를 제공합니다.
3. **자체 업데이트 보안 핀닝(Hardcoded Signer Pinning)**: 마켓 앱 자체 업데이트 APK는 서버 응답값뿐만 아니라 마켓 앱 바이너리 내부에 하드코딩된 공식 Booking Network 스토어 서명 인증서 지문(`PINNED_STORE_CERT_SHA256`)과 일치할 때만 `PackageInstaller` 세션을 커밋하여, API 서버가 침해되더라도 악성 마켓 앱으로 교체되는 것을 원천 차단합니다.

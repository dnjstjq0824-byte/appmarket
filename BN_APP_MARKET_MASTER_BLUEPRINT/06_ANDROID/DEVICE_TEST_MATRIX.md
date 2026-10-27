# DEVICE_TEST_MATRIX.md — Android 버전별(API 26~35) 및 ABI별 실기기·에뮬레이터 테스트 매트릭스

- **문서 ID**: `DOC-AND-004`
- **관련 조항**: 제5조 (항목 20), 제18조
- **연관 테스트 ID**: `TC-AND-001` ~ `TC-AND-004`

---

## 1. Android OS 버전별 핵심 동작 차이 및 필수 검증 매트릭스

| Android 버전 | API Level | OS 패키지 설치/권한 주요 변경점 | 필수 자동화/실기기 검증 시나리오 |
| :--- | :---: | :--- | :--- |
| **Android 8.0 / 8.1** | `26 / 27` | 앱별 `REQUEST_INSTALL_PACKAGES` 도입, `GET_SIGNATURES` 사용 | 알 수 없는 출처 설정 화면 이동 및 복귀 후 정상 설치 검증 |
| **Android 9.0** | `28` | APK Signature Scheme v3(키 회전) 및 `GET_SIGNING_CERTIFICATES` 도입 | v3 `SigningInfo` 추출 및 서명 일치 검증 |
| **Android 10** | `29` | Scoped Storage 도입 | 앱 전용 외부 저장소(`getExternalFilesDir`) 스테이징 다운로드 및 정리 검증 |
| **Android 11** | `30` | 패키지 가시성 제한(`QUERY_ALL_PACKAGES`), `getInstallSourceInfo` 도입 | 전체 설치 앱 스캔 및 설치 소유자(`installerPackageName`) 판별 검증 |
| **Android 12 / 12L** | `31 / 32` | `PendingIntent.FLAG_MUTABLE` 명시 의무화, `USER_ACTION_NOT_REQUIRED` 자동 업데이트 도입 | 최초 설치(사용자 승인 1회) 후 버전 2 업데이트 시 무음 업데이트 동작 검증 |
| **Android 13** | `33` | `POST_NOTIFICATIONS` 런타임 권한 분리, `PackageManager.PackageInfoFlags.of(Long)` 도입 | 알림 권한 거부 상태에서도 포그라운드/인앱 설치가 크래시 없이 완료되는지 검증 |
| **Android 14** | `34` | `InstallConstraints` API, `targetSdk < 23` 앱 설치 OS 차단,Foreground Service 타입(`dataSync`) 강제 | `foregroundServiceType="dataSync"` 대용량 백그라운드 다운로드 및 제약조건 처리 검증 |
| **Android 15** | `35` | `targetSdk < 24` 설치 차단 강화, 아카이브 앱(App Archiving) 상태 처리 | 최신 타깃 SDK 35 환경에서 전체 설치/업데이트/롤백 차단 시나리오 통과 검증 |

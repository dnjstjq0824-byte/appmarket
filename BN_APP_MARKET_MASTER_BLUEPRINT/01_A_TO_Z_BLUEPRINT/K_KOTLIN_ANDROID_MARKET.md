# K — Kotlin Android Market — Kotlin + Jetpack Compose 네이티브 Android 앱 마켓 클라이언트 및 설치 엔진 (Kotlin + Jetpack Compose Native Android Market Client & PackageInstaller Engine)

- **문서 ID**: `DOC-AZ-K`
- **알파벳 코드**: `K`
- **관련 요구사항 ID**: `REQ-K-001`, `REQ-K-002`, `REQ-K-003`
- **연관 Work Order ID**: `AM-ANDROID-001`, `AM-ANDROID-002`, `AM-ANDROID-003`
- **연관 테스트 케이스 ID**: `TC-AND-001`, `TC-AND-002`, `TC-AND-003`, `TC-AND-004`
- **참조 상세 설계 문서**: [06_ANDROID/ANDROID_ARCHITECTURE.md](../06_ANDROID/ANDROID_ARCHITECTURE.md), [06_ANDROID/APK_INSTALL_ENGINE.md](../06_ANDROID/APK_INSTALL_ENGINE.md), [06_ANDROID/APP_UPDATE_ENGINE.md](../06_ANDROID/APP_UPDATE_ENGINE.md), [06_ANDROID/DEVICE_TEST_MATRIX.md](../06_ANDROID/DEVICE_TEST_MATRIX.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 웹 브라우저만으로는 Android 기기에 설치된 앱들의 현재 버전·서명 인증서를 일괄 스캔하여 원클릭 업데이트를 제공하거나, 대용량 APK 백그라운드 이어받기 및 OS 수준의 `PackageInstaller` 세션 제어를 수행할 수 없다(제5조). 따라서 실제 Android OS API와 직결된 상용 Kotlin 네이티브 마켓 앱이 필수적이다.
- **달성 목표**: 제5조의 20개 필수 설계 항목을 100% 충족하는 멀티모듈 Kotlin + Jetpack Compose 아키텍처, 이어받기 다운로드 매니저, SHA-256 및 서명 인증서 검증기, `PackageInstaller.Session` 설치/업데이트 엔진, 기기 호환성 판별기 및 스토어 자체 업데이트 메커니즘을 명세한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **Android 기기 사용자**: 네이티브 마켓 앱을 통해 기기 맞춤형 호환 앱 검색, 원클릭 다운로드·설치, 기설치 앱 자동/수동 업데이트, 외부 스토어 딥링크 이동 수행.

## 3. 필수 기능 (Mandatory Features)
1. **제5조 20대 필수 설계 항목 전수 반영**:
   - (1) 멀티모듈 Kotlin 구조 (`:app`, `:core:network`, `:core:database`, `:core:installer`, `:core:downloader`, `:core:device`, `:feature:home`, `:feature:search`, `:feature:detail`, `:feature:updates`, `:feature:settings`)
   - (2) Jetpack Compose Material 3 + MVI 단방향 상태 아키텍처
   - (3) Room 로컬 DB + DataStore 상태 영속화
   - (4) Retrofit2/OkHttp3 + Kotlinx Serialization + TLS 1.3 Certificate Pinning(선택적 구성)
   - (5~6) HTTP `Range: bytes=start-` 기반 다운로드 일시정지·재개·취소·지수 백오프 재시도 (`WorkManager` + Foreground Service)
   - (7) 스트림 완료 직후 로컬 파일 `SHA-256` 해시 재계산 및 서버 매니페스트 대조
   - (8~10) `android.content.pm.PackageInstaller.Session` 스트림 기록(`openWrite` + `fsync`), `BroadcastReceiver`를 통한 `STATUS_PENDING_USER_ACTION` 사용자 승인 인텐트 실행 및 7대 `STATUS_FAILURE_*` 세분화 복구
   - (11, 15) 설치/업데이트 전 `PackageManager.getPackageArchiveInfo(..., GET_SIGNING_CERTIFICATES)`로 다운로드된 APK의 서명 인증서 SHA-256을 추출하여 서버 기댓값 및 기설치 앱 서명과 대조
   - (12) `StorageManager.getAllocatableBytes()`로 `APK 크기 * 2.2배 + 100MB` 여유 공간 사전 확인 및 `allocateBytes` 확보
   - (13) `ConnectivityManager.NetworkCallback` 기반 오프라인/Wi-Fi 전용 다운로드 전환
   - (14) `Build.VERSION.SDK_INT`, `Build.SUPPORTED_ABIS`, `DisplayMetrics.densityDpi`, `PackageManager.getSystemAvailableFeatures()` 기반 기기 호환성 사전 판별
   - (16) 다운로드 상태(`DownloadState`)와 설치 상태(`InstallState`)의 엄격한 분리 관리
   - (17) 외부 스토어(`market://`, `samsungapps://`, `https://`) 안전한 `Intent.ACTION_VIEW` 딥링크 라우팅
   - (18) Android 13+(API 33) `POST_NOTIFICATIONS` 런타임 권한 처리 및 다운로드/업데이트/보안 경고 알림 채널 분리
   - (19) 마켓 앱 자체 업데이트(Self-Update) 전용 부트스트랩 및 서명 고정 검증
   - (20) Android 8.0(API 26) ~ Android 15(API 35) 버전별 호환성 및 실기기 테스트 매트릭스.
2. **보안 정책 준수**: 일반 사용자 기기에 대한 무단 무음 설치 우회(루팅, 접근성 악용 등)를 절대 구현하지 않으며 공식 Android 보안 정책을 100% 준수한다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **`PackageInstaller.Session` 정밀 실행 흐름**:
  1. 사전 권한 확인: `context.packageManager.canRequestPackageInstalls()`가 `false`이면 `Settings.ACTION_MANAGE_UNKNOWN_APP_SOURCES` (`package:$packageName`) 설정 화면으로 안내하고, 권한 획득 시 자동으로 설치 플로우 재개.
  2. 세션 파라미터 생성: `SessionParams(MODE_FULL_INSTALL)`에 `setAppPackageName(pkg)`, `setSize(bytes)`, `setInstallLocation(...)` 설정. Android 12+(API 31+)에서 본 마켓이 설치한 앱을 업데이트할 때만 `setRequireUserAction(USER_ACTION_NOT_REQUIRED)` 설정.
  3. 바이트 기록 및 동기화: `session.openWrite("base.apk", 0, fileSize)`에 버퍼 스트리밍 후 반드시 `session.fsync(outStream)` 호출.
  4. 커밋 및 결과 수신: `PendingIntent.FLAG_MUTABLE` (Android 12+ 필수 요구사항, 단 명시적 내부 `ComponentName` 지정으로 하이재킹 방지)을 전달하여 `session.commit(intentSender)` 실행.
  5. `InstallResultReceiver`에서 `PackageInstaller.EXTRA_STATUS`를 분석하여:
     - `STATUS_PENDING_USER_ACTION`: 동봉된 `Intent`에 `FLAG_ACTIVITY_NEW_TASK`를 붙여 시스템 설치 승인 다이얼로그 호출.
     - `STATUS_SUCCESS`: 설치 완료 상태 갱신, 임시 APK 파일 삭제(옵션), 서버에 설치 성공 이벤트 보고.
     - `STATUS_FAILURE_CONFLICT` (`INSTALL_FAILED_UPDATE_INCOMPATIBLE`): 기존 설치된 앱과 서명 키가 다름을 설명하는 모달 표시.
     - `STATUS_FAILURE_STORAGE`: 부족한 용량을 MB 단위로 표시하고 캐시 정리 유도.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- Jetpack Compose 기반 화면 구성: `HomeScreen`, `SearchScreen`, `AppDetailScreen`, `DownloadManagerScreen`, `UpdatesScreen`, `SettingsScreen`, `PermissionOnboardingSheet`.

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- Android 클라이언트는 기기 프로필 헤더(`X-Device-Sdk: 34`, `X-Device-Abis: arm64-v8a,armeabi-v7a`, `X-Device-Dpi: 480`)를 전달하며, 백엔드 Distribution API는 해당 기기에 가장 적합한 APK/Split APK 아티팩트와 서명 해시를 내려준다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **Android 로컬 Room DB (`MarketDatabase`)**:
  - `download_tasks` (`task_id PK`, `app_id`, `release_id`, `package_name`, `version_code`, `download_url`, `expected_sha256`, `expected_signer_sha256`, `total_bytes`, `downloaded_bytes`, `etag`, `download_state`, `install_state`, `install_session_id`, `failure_code`, `updated_at`).

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `POST /api/v1/device/check-updates`: 기기에 설치된 앱 목록(`package_name`, `version_code`, `signer_sha256`, `installer_package`)을 전송하여 업데이트 가능 목록 및 서명 호환성 판정 결과를 일괄 수신.
- `POST /api/v1/catalog/apps/{appId}/releases/{releaseId}/download-token`: 서명된 다운로드 URL 및 검증용 `sha256`/`signer_sha256` 발급.
- `POST /api/v1/events/install`: 다운로드/설치/업데이트 성공·실패 결과 텔레메트리 보고.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **다운로드 상태(`DownloadState`)와 설치 상태(`InstallState`)의 이중 상태 머신**:
  - `DownloadState`: `IDLE` → `QUEUED` → `DOWNLOADING(progress%)` ⇄ `PAUSED_BY_USER` / `PAUSED_WAITING_WIFI` / `PAUSED_NETWORK_ERROR` → `VERIFYING_CHECKSUM` → `COMPLETED_VERIFIED` | `FAILED_CHECKSUM_MISMATCH` | `FAILED_STORAGE_FULL` | `CANCELED`.
  - `InstallState`: `NOT_STARTED` → `PRECHECKING_COMPATIBILITY` → `WAITING_UNKNOWN_SOURCES_PERMISSION` → `WRITING_SESSION` → `AWAITING_USER_CONFIRMATION` → `INSTALLING_OS` → `INSTALLED` | `FAILED_SIGNATURE_CONFLICT` | `FAILED_DOWNGRADE` | `FAILED_ABORTED_BY_USER` | `FAILED_OS_ERROR`.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- OWASP MASVS-STORAGE, MASVS-CRYPTO, MASVS-NETWORK, MASVS-PLATFORM 준수.
- 다운로드 중인 APK 파일은 외부 앱이 접근할 수 없는 앱 전용 내부 디렉터리(`context.getExternalFilesDir("apks_staging")` 또는 `context.filesDir`)에 저장하며, `SHA-256` 검증과 `apksigner`/인증서 지문 검증이 모두 통과하기 전에는 절대 `PackageInstaller.Session`에 바이트를 전달하지 않는다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 기기와 호환되지 않는 앱(예: 기기는 `armeabi-v7a` 32비트인데 앱은 `arm64-v8a` 전용이거나, 기기 SDK가 28인데 앱 `minSdk`가 31인 경우)은 설치 버튼을 비활성화(`Disabled`)하고 `"현재 기기(Android 9 / 32-bit ARM)와 호환되지 않습니다 (요구사양: Android 12 이상)"`라고 명확히 안내한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **APK 해시 변조 설치 차단 테스트 (`TC-AND-001`)**: 다운로드된 파일 1바이트 수정 시 `FAILED_CHECKSUM_MISMATCH`로 전이되고 임시 파일이 즉시 삭제되며 `PackageInstaller`가 호출되지 않음을 검증.
- **이어받기(HTTP Range) 테스트 (`TC-AND-002`)**: 50% 다운로드 시점 네트워크 차단 후 재연결 시 `206 Partial Content`로 50% 지점부터 이어받아 설치까지 완료되는지 검증.
- **서명 불일치 업데이트 차단 테스트 (`TC-AND-003`)**: 동일 패키지명·다른 서명 키로 서명된 APK 업데이트 시도 시 다운로드 전 경고 및 설치 충돌 처리가 정상 동작하는지 검증.
- **API 26~35 버전별 설치 테스트 (`TC-AND-004`)**: 에뮬레이터 매트릭스에서 실제 설치 승인 인텐트 및 업데이트 동작 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `N — Native App Distribution`, `V — Versioning & Update Engine`, `C — Catalog Engine`과 직접 연동된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0 (네이티브 앱 마켓 핵심)`
- **개발 Phase**: `Phase 4 (APK 배포 및 Android 설치)`에서 집중 구현한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 설치 세션 진행 중 앱 프로세스 종료** → `PackageInstaller.getMySessions()`를 앱 기동 시 스캔하여 진행 중이던 세션을 복원하거나 고아 세션을 안전하게 `abandonSession()`으로 정리한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- Android 프로젝트 전체 소스코드 (`android-client/`) 및 서명된 릴리스 APK 빌드 파이프라인.

## 17. Acceptance Criteria (수용 기준)
1. Android 기기/에뮬레이터에서 앱 검색 → 다운로드 → 무결성 검증 → 사용자 승인 → 실제 OS 앱 설치 및 실행 → 버전 업데이트까지 전 과정이 실제 동작해야 한다.
2. 저장공간 부족, 네트워크 끊김, 서명 충돌, 사용자 취소 4대 예외 상황에서 앱 크래시 없이 복구 UI가 동작해야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-ANDROID-002/adb_logcat_package_installer.log`: `PackageInstaller` 세션 생성, 커밋, `STATUS_SUCCESS` 수신 `adb logcat` 실증 로그.
- `artifacts/evidence/AM-ANDROID-003/instrumentation_test_report.xml`: Android Connected Test 결과 리포트.

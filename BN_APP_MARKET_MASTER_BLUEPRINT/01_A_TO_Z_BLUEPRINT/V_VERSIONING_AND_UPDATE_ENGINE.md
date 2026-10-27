# V — Versioning & Update Engine — 버전 비교, 서명 연속성 검증, 설치 소유권, 업데이트 알림 및 호환 기기 판별 엔진 (Version Comparison, Signature Continuity, Installer Ownership, Update Alerts & Device Compatibility Engine)

- **문서 ID**: `DOC-AZ-V`
- **알파벳 코드**: `V`
- **관련 요구사항 ID**: `REQ-V-001`, `REQ-V-002`
- **연관 Work Order ID**: `AM-DIST-002`, `AM-ANDROID-003`
- **연관 테스트 케이스 ID**: `TC-AND-003`, `TC-FUNC-012`
- **참조 상세 설계 문서**: [06_ANDROID/APP_UPDATE_ENGINE.md](../06_ANDROID/APP_UPDATE_ENGINE.md), [07_DISTRIBUTION/APK_AAB_SIGNING.md](../07_DISTRIBUTION/APK_AAB_SIGNING.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 통합 앱 마켓(Type 3)에서는 사용자가 특정 앱을 당사 스토어, F-Droid, 또는 타 스토어에서 설치했을 수 있다. 업데이트를 제공할 때 단순히 `versionName` 문자열을 비교하면 안 되며, Android OS의 엄격한 규칙인 **`versionCode` 정수 단조 증가**, **서명 인증서 일치(또는 v3 Key Rotation Lineage 일치)**, **설치 소유권(`installerPackageName`)**, 그리고 **기기 SDK/ABI 호환성**을 동시에 판별해야 설치 실패(`INSTALL_FAILED_UPDATE_INCOMPATIBLE`, `INSTALL_FAILED_OLDER_SDK`)를 예방할 수 있다.
- **달성 목표**: 서버·클라이언트 양방향 업데이트 판별 알고리즘, 서명 충돌 감지 및 교차 배포처 마이그레이션 안내, 설치 소유권 판별, 실패 복구 체계를 명세한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **Android 기기 사용자**: 내 기기에 설치된 앱들의 업데이트 가능 목록을 확인하고, 서명 호환성이 보장된 업데이트를 안전하게 적용.
- **Release Manager (개발자)**: 기존 사용자 기기에서 업데이트가 깨지지 않도록 서명 키 연속성과 `minSdk`/`versionCode` 규칙을 검증받음.

## 3. 필수 기능 (Mandatory Features)
1. **`versionCode` (Long/INT8) 기준 결정론적 버전 비교**:
   - 표시용 문자열인 `versionName`(예: `"2.4.1-beta"`)은 UI 표시 전용이며, 모든 업데이트 판별은 오직 정수형 `versionCode`(`new_version_code > installed_version_code`)로만 수행한다.
2. **서명 호환성 3단계 판정 (`SignatureCompatibilityStatus`)**:
   - `COMPATIBLE_SAME_SIGNER`: 기설치 앱의 인증서 SHA-256 지문과 신규 APK의 인증서 SHA-256 지문이 100% 일치 → **즉시 덮어쓰기 업데이트 가능(데이터 유지)**.
   - `COMPATIBLE_KEY_ROTATION`: 인증서 지문은 다르나 신규 APK의 Signing Block v3 `Proof-of-Rotation` 체인에 기설치 앱의 인증서가 이전 키로 포함되어 있고 암호학적 서명이 유효함 → **덮어쓰기 업데이트 가능**.
   - `INCOMPATIBLE_DIFFERENT_SIGNER`: 동일 `packageName`이나 서명 키가 다름(예: Google Play 설치본 vs F-Droid 자체 빌드본) → **자동 업데이트 대상에서 제외**하고, 수동 선택 시 `"배포처 서명이 달라 기존 앱을 삭제(앱 데이터 초기화)한 후에만 설치할 수 있습니다"` 경고를 표시.
3. **설치 소유권(`InstallSourceInfo`) 확인**:
   - Android 11+(API 30) `PackageManager.getInstallSourceInfo(pkg).initiatingPackageName` / `installingPackageName`을 조회하여 본 마켓(`com.bookingnetwork.appmarket`)이 설치한 앱인지 타 스토어가 설치한 앱인지 구분한다.
   - 타 스토어가 설치한 앱은 사용자가 앱별 설정에서 `'이 앱의 업데이트를 BN App Market에서 관리'`를 명시적으로 켠 경우에만 업데이트 알림을 보낸다(원치 않는 스토어 하이재킹 방지).
4. **기기 호환성 4차원 매칭 (Device Compatibility Matrix)**:
   - ① `device.sdk_int >= artifact.min_sdk`
   - ② `artifact.native_abis`가 비어있거나(순수 Java/Kotlin 앱) `device.supported_abis`와 최소 1개 이상 교집합 존재
   - ③ `artifact.required_features` ⊆ `device.available_features` (예: `android.hardware.nfc`, `android.hardware.camera.ar`)
   - ④ `artifact.dpi_filter` 호환.

## 4. 상세 기능 (Detailed Functional Specifications)
- **배치 업데이트 확인 API (`POST /api/v1/device/check-updates`) 최적화**:
  - 클라이언트가 기기에 설치된 서드파티 앱 패키지 목록(최대 300개)의 `(package_name, version_code, cert_sha256, installer_pkg)`와 기기 사양 `(sdk_int, abis, dpi, features_hash)`를 전송하면, 서버는 단일 배치 쿼리와 단계적 배포 버킷 평가를 수행하여 각 패키지별로 `NO_UPDATE`, `UPDATE_AVAILABLE_COMPATIBLE`, `UPDATE_AVAILABLE_SIGNER_CONFLICT`, `INCOMPATIBLE_DEVICE_SDK_ABI` 상태를 반환한다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **`android-client`**: `UpdatesScreen.kt` (섹션 1: `즉시 업데이트 가능 (서명 일치)` — `'모두 업데이트'` 원클릭 지원 / 섹션 2: `다른 배포처에서 설치된 앱 (서명 주의)` — 개별 확인 후 수동 전환 지원).
- **`web-user`**: `VersionHistoryTab.vue` (이전 릴리스 아카이브 목록 및 각 버전별 `versionCode`, `minSdk`, `SHA-256`, 서명 지문 표시).

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/distribution/domain/compatibility_engine.go` 및 `update_resolver.go`.

## 7. DB 및 데이터 모델 (Database & Data Model)
- `app_releases` 테이블의 `(distribution_id, version_code DESC)` 복합 인덱스와 `release_artifacts`의 GIN 인덱스(`native_abis`)를 통해 수백 개 패키지의 호환 최신 버전을 20ms 이내에 조회한다.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `POST /api/v1/device/check-updates`: 기기 설치 앱 일괄 업데이트 및 호환성 판정.
- `GET /api/v1/catalog/apps/{slug}/versions`: 버전 히스토리 및 이전 버전 다운로드 정보 조회.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **업데이트 실패 자동 복구**: 업데이트 다운로드 또는 `PackageInstaller` 커밋 중 실패(`STATUS_FAILURE_STORAGE`, `STATUS_FAILURE_IO`) 시 기존 설치된 구버전 앱은 Android OS의 트랜잭션 세션 특성상 그대로 보존되며, 클라이언트는 손상된 스테이징 파일만 정리하고 재시도 큐에 등록한다.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- **버전 다운그레이드 공격(Rollback Attack) 차단**: 서버나 네트워크 중간자가 기기에 이미 설치된 `versionCode`보다 낮거나 같은 버전을 업데이트로 내려주더라도 클라이언트의 `UpdateValidator`가 `newVersionCode <= installedVersionCode`를 즉시 감지하여 설치를 거부한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 업데이트 목록에서 각 앱의 `'변경 사항(Release Notes)'`을 펼침 메뉴로 바로 읽을 수 있게 제공하고, 업데이트로 인해 **새로 추가되는 민감 권한**(예: 구버전에는 없던 `마이크` 권한이 신버전에 추가된 경우)이 있으면 `'업데이트 전 신규 권한 확인 필요'` 배지를 띄워 자동 업데이트 대신 사용자 동의를 받는다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **4차원 기기 호환성 및 서명 연속성 판별 테스트 (`TC-FUNC-012`)**: SDK 부족, ABI 불일치(`x86_64` 전용 APK vs `armeabi-v7a` 기기), 동일 서명, 상이 서명, v3 키 회전 5가지 시나리오에 대해 배치 업데이트 API가 100% 정확한 상태값을 반환하는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `K — Kotlin Android Market`, `L — Lifecycle & Release Management`, `N — Native App Distribution`과 직결된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0`
- **개발 Phase**: `Phase 4 (APK 배포 및 Android 설치)`에서 구현한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 수십만 대의 기기가 정각(00:00)에 동시에 업데이트 체크를 요청하는 Thundering Herd 현상** → Android `WorkManager` 주기적 작업에 `flexInterval`과 기기별 랜덤 지터(0~120분)를 부여하고, 대규모 앱 출시 푸시 알림도 배치 단위로 분산 발송한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 서버 측 호환성/업데이트 판별 엔진 (`backend/internal/modules/distribution/domain/update_resolver.go`) 및 Android 클라이언트 업데이트 매니저 (`android-client/core/installer/UpdateManager.kt`).

## 17. Acceptance Criteria (수용 기준)
1. 서명 인증서가 다른 동일 패키지명 앱은 절대 자동 업데이트 큐에 포함되지 않아야 한다.
2. 신규 버전에서 고위험 권한이 추가된 경우 사용자에게 변경된 권한이 고지되어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-DIST-002/update_compatibility_matrix_test.log`: 5대 호환성·서명 시나리오 검증 로그.

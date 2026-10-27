# AM-ANDROID-002 — Android HTTP Range 이어받기 다운로더, SHA-256/서명 인증서 검증기 및 PackageInstaller.Session 설치 엔진 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-ANDROID-002` |
| **작업명 (Title)** | Android HTTP Range 이어받기 다운로더, SHA-256/서명 인증서 검증기 및 PackageInstaller.Session 설치 엔진 구현 |
| **목표 (Objective)** | `REQ-K-002;REQ-N-003` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Android Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-ANDROID-001;AM-DIST-002` |
| **개발 Phase 및 우선순위** | `Phase 4` / `P0` |
| **입력 문서 (Input Documents)** | `06_ANDROID/APK_INSTALL_ENGINE.md;06_ANDROID/DEVICE_TEST_MATRIX.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `android-client/core/downloader/;android-client/core/installer/;android-client/feature/updates/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `backend/;apps/` |
| **예상 생성/수정 파일 (Target Files)** | `android-client/core/downloader/src/main/kotlin/HttpRangeDownloader.kt;android-client/core/installer/src/main/kotlin/ApkInstallCoordinator.kt;android-client/core/installer/src/main/kotlin/InstallResultReceiver.kt;android-client/core/installer/src/main/kotlin/ApkCertVerifier.kt` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/catalog/apps/{appId}/releases/{releaseId}/download-token;POST /api/v1/events/install` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-AND-001;TC-AND-002` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-ANDROID-002/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. `StorageManager.getAllocatableBytes()` 사전 용량 확인 및 HTTP `Range` 기반 일시정지/재개/취소/재시도 다운로더 구현.
2. 다운로드 완료 직후 로컬 `SHA-256` 해시 계산 및 `PackageManager.getPackageArchiveInfo(..., GET_SIGNING_CERTIFICATES)` 서명 인증서 SHA-256 대조.
3. `REQUEST_INSTALL_PACKAGES` 권한 온보딩, `PackageInstaller.Session` 스트림 쓰기(`fsync`), `STATUS_PENDING_USER_ACTION` 인텐트 실행 및 7대 실패 코드 복구 UI 구현.

## 2. Acceptance Criteria (수용 기준)
1. 다운로드된 APK 1바이트 변조 시 `FAILED_CHECKSUM_MISMATCH`로 즉시 폐기되고 `PackageInstaller`가 절대 호출되지 않음.
2. 실제 Android 에뮬레이터/기기에서 사용자 승인 후 앱이 정상 설치되고 `STATUS_SUCCESS`가 수신됨.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 설치 도중 앱 재시작 시 `PackageInstaller.getMySessions()`로 세션 상태를 복원하거나 고아 세션을 정리.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-AND-001;TC-AND-002`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-ANDROID-002/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

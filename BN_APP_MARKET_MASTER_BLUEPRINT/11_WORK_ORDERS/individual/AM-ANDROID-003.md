# AM-ANDROID-003 — Android 기설치 앱 스캔, 서명 연속성 검증 업데이트 엔진 및 마켓 앱 자체 업데이트(Self-Update) 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-ANDROID-003` |
| **작업명 (Title)** | Android 기설치 앱 스캔, 서명 연속성 검증 업데이트 엔진 및 마켓 앱 자체 업데이트(Self-Update) 구현 |
| **목표 (Objective)** | `REQ-K-003;REQ-V-001;REQ-V-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Android Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-ANDROID-002` |
| **개발 Phase 및 우선순위** | `Phase 4` / `P0` |
| **입력 문서 (Input Documents)** | `06_ANDROID/APP_UPDATE_ENGINE.md;06_ANDROID/DEVICE_TEST_MATRIX.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `android-client/core/installer/;android-client/feature/updates/;android-client/feature/settings/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `backend/;apps/` |
| **예상 생성/수정 파일 (Target Files)** | `android-client/core/installer/src/main/kotlin/InstalledPackagesScanner.kt;android-client/core/installer/src/main/kotlin/StoreSelfUpdater.kt;android-client/feature/updates/src/main/kotlin/UpdatesCenterScreen.kt` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/device/check-updates;GET /api/v1/system/capabilities` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-AND-003;TC-AND-004` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-ANDROID-003/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 기설치 앱의 `versionCode`, `signingInfo` 인증서 지문, `InstallSourceInfo`를 스캔하여 서버 `/api/v1/device/check-updates`와 동기화.
2. 서명 일치 앱의 원클릭 업데이트(API 31+ 자체 설치 앱 `USER_ACTION_NOT_REQUIRED` 지원)와 서명 불일치 앱의 충돌 경고 모달(`SignatureConflictDialog`) 구현.
3. 하드코딩된 스토어 서명 인증서 핀닝(`PINNED_STORE_CERT_SHA256`) 기반 마켓 앱 자체 업데이트 구현.

## 2. Acceptance Criteria (수용 기준)
1. 서명이 다른 동일 패키지명 업데이트는 자동 업데이트에서 제외되고 경고 다이얼로그가 표시됨.
2. 버전 다운그레이드(`newVersionCode <= installedVersionCode`) 업데이트 시도가 클라이언트에서 100% 차단됨.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 자체 업데이트 APK의 서명이 핀닝된 인증서와 다를 경우 즉시 파일을 삭제하고 보안 경고를 표시.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-AND-003;TC-AND-004`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-ANDROID-003/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

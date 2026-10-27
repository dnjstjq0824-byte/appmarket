# AM-ANDROID-004 — Android 클라이언트バージョン별(API 26~35) Connected Instrumentation 테스트 및 예외 복구 검증 스위트 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-ANDROID-004` |
| **작업명 (Title)** | Android 클라이언트バージョン별(API 26~35) Connected Instrumentation 테스트 및 예외 복구 검증 스위트 구현 |
| **목표 (Objective)** | `REQ-K-002;REQ-K-003;REQ-Q-001` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Android Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-ANDROID-003` |
| **개발 Phase 및 우선순위** | `Phase 8` / `P1` |
| **입력 문서 (Input Documents)** | `06_ANDROID/DEVICE_TEST_MATRIX.md;13_TESTING_QA/E2E_TEST_PLAN.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `android-client/app/src/androidTest/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `backend/;apps/` |
| **예상 생성/수정 파일 (Target Files)** | `android-client/app/src/androidTest/kotlin/PackageInstallerE2ETest.kt;android-client/app/src/androidTest/kotlin/HttpRangeResumeTest.kt;android-client/app/src/androidTest/kotlin/SignatureConflictTest.kt` |
| **API 의존성 (API Dependencies)** | `NONE` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-AND-001;TC-AND-002;TC-AND-003;TC-AND-004` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-ANDROID-004/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. Android API 26, 29, 31, 33, 34, 35 에뮬레이터 매트릭스에서 다운로드 이어받기, SHA-256 변조 차단, `PackageInstaller` 세션 커밋, 서명 불일치 차단, 저장공간 부족 복구를 자동 검증하는 Instrumentation 테스트 구현.

## 2. Acceptance Criteria (수용 기준)
1. 6개 대상 API 레벨에서 모든 Connected Test 통과 및 `adb logcat` 증빙 수집 완료.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 특정 API 레벨에서 OS 동작 차이로 인한 실패 발견 시 SDK 분기 처리 수정 후 전체 매트릭스 재실행.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-AND-001;TC-AND-002;TC-AND-003;TC-AND-004`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-ANDROID-004/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

# AM-ANDROID-001 — Android 클라이언트(Kotlin + Jetpack Compose) 멀티모듈 프로젝트, 네트워크/Room/기기호환성 코어 및 탐색 UI 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-ANDROID-001` |
| **작업명 (Title)** | Android 클라이언트(Kotlin + Jetpack Compose) 멀티모듈 프로젝트, 네트워크/Room/기기호환성 코어 및 탐색 UI 구현 |
| **목표 (Objective)** | `REQ-K-001` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Android Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-ARCH-001;AM-API-002` |
| **개발 Phase 및 우선순위** | `Phase 4` / `P0` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/K_KOTLIN_ANDROID_MARKET.md;06_ANDROID/ANDROID_ARCHITECTURE.md;05_FRONTEND_UXUI/SCREEN_INVENTORY.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `android-client/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `backend/;apps/` |
| **예상 생성/수정 파일 (Target Files)** | `android-client/settings.gradle.kts;android-client/app/build.gradle.kts;android-client/core/device/src/main/kotlin/DeviceSpecProvider.kt;android-client/feature/home/src/main/kotlin/HomeScreen.kt;android-client/feature/detail/src/main/kotlin/AppDetailScreen.kt` |
| **API 의존성 (API Dependencies)** | `GET /api/v1/catalog/apps;GET /api/v1/catalog/apps/{slug};GET /api/v1/search/apps` |
| **DB 변경 여부 (DB Schema Changes)** | `NO (Android Room Local DB)` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-AND-004` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-ANDROID-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. `minSdk = 26`, `targetSdk = 35` 기준 멀티모듈 Kotlin + Jetpack Compose + Hilt + Room + Retrofit2/OkHttp3 프로젝트 구성.
2. 기기의 `SDK_INT`, `SUPPORTED_ABIS`, `densityDpi`, 시스템 하드웨어 피처를 수집하는 `DeviceSpecProvider` 및 호환성 배지 UI 구현.
3. `AND-01`~`AND-05`, `AND-11` 화면 및 외부 스토어(`market://`, `https://`) 안전한 딥링크 인텐트 라우터 구현.

## 2. Acceptance Criteria (수용 기준)
1. `./gradlew assembleRelease testDebugUnitTest` 에러 0건 빌드 성공.
2. 비호환 ABI/SDK 앱 상세 진입 시 설치 버튼이 비활성화되고 사유가 정확히 표시됨.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 네트워크 단절 시 Room 로컬 캐시 카탈로그 표시 및 재연결 시 자동 리프레시.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-AND-004`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-ANDROID-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

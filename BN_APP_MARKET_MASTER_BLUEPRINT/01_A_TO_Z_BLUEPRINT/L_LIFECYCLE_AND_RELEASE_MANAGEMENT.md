# L — Lifecycle & Release Management — 앱 등록, 심사, 트랙별 출시, 단계적 배포, 롤백 및 라이프사이클 관리 (App Registration, Review Pipeline, Release Tracks, Staged Rollout, Rollback & Lifecycle Management)

- **문서 ID**: `DOC-AZ-L`
- **알파벳 코드**: `L`
- **관련 요구사항 ID**: `REQ-L-001`, `REQ-L-002`
- **연관 Work Order ID**: `AM-DIST-001`, `AM-DEV-002`, `AM-ADMIN-002`
- **연관 테스트 케이스 ID**: `TC-FUNC-007`, `TC-FUNC-008`
- **참조 상세 설계 문서**: [07_DISTRIBUTION/ARTIFACT_PIPELINE.md](../07_DISTRIBUTION/ARTIFACT_PIPELINE.md), [09_SECURITY_COMPLIANCE/APP_REVIEW_POLICY.md](../09_SECURITY_COMPLIANCE/APP_REVIEW_POLICY.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 앱이 최초 생성되어 바이너리가 업로드되고, 자동 보안 검사와 운영자 심사를 통과해 내부 테스트·베타·프로덕션 트랙으로 출시되며, 장애 발생 시 단계적 배포(Staged Rollout)를 일시 중단(Halt)하거나 이전 안정 버전으로 롤백(Rollback)하고, 최종적으로 서비스 종료(Sunset/Archive) 또는 정책 위반 강제 삭제에 이르는 전 과정을 결정론적 상태 머신으로 통제하기 위해 필요하다.
- **달성 목표**: 12단계 릴리스 상태 머신, 3대 배포 트랙(`INTERNAL`, `BETA`, `PRODUCTION`), 결정론적 해시 버킷 기반 단계적 배포(1%~100%), 긴급 롤백 및 단종(Deprecation) 절차를 정의한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **Release Manager (개발자)**: 릴리스 생성, 아티팩트 연결, 트랙 선택, 심사 제출, 단계적 배포 비율(1%, 5%, 20%, 50%, 100%) 조정, 배포 일시중단/재개, 이전 버전 롤백.
- **App Reviewer / Security Analyst (관리자)**: 심사 큐에서 자동 스캔 리포트 검토, 실제 테스트 기기 설치 점검, 승인(`APPROVED`)/반려(`REJECTED`)/보완요청(`CHANGES_REQUESTED`) 판정, 긴급 차단(`SUSPENDED`).

## 3. 필수 기능 (Mandatory Features)
1. **3대 배포 트랙 (`release_track`)**:
   - `INTERNAL`: 개발자 조직 내부 테스터(최대 100명)에게만 자동 보안 스캔 통과 즉시(운영자 수동 심사 없이) 5분 내 배포.
   - `BETA`: 공개/비공개 베타 참여 사용자에게 배포 (경량 정책 심사 적용).
   - `PRODUCTION`: 전체 일반 사용자 대상 정식 배포 (자동 보안 스캔 + 운영자 심사 필수).
2. **단계적 배포 (Deterministic Staged Rollout)**:
   - `rollout_percentage` (`1` ~ `100`) 설정 시, `CRC32(device_id_or_user_id || ":" || app_id) % 100 < rollout_percentage` 공식을 사용하여 동일한 사용자는 항상 일관되게 신규 버전 대상에 포함되거나 유지되도록 보장한다.
3. **긴급 배포 중단(Halt) 및 롤백(Rollback)**:
   - 신규 릴리스에서 치명적 크래시가 발견되면 개발자 또는 관리자가 1클릭으로 `HALTED` 상태로 전환하여 추가 확산을 즉시 멈춘다.
   - 이미 신규 버전을 설치한 기기는 Android OS 특성상 버전 코드를 낮추는 다운그레이드 덮어쓰기가 불가능하므로(`INSTALL_FAILED_VERSION_DOWNGRADE`),① 아직 업데이트하지 않은 기기에는 직전 안정 버전(`PREVIOUS_ACTIVE_RELEASE`)을 계속 서빙하고, ② 이미 업데이트한 기기를 위해서는 이전 안정 코드베이스의 `versionCode`만 높여 재서명한 긴급 롤백 릴리스(Hotfix Rollback Release) 또는 빠른 재검증 패스를 지원한다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **릴리스 상태 머신 (12 States)**:
  - `DRAFT` → `ARTIFACT_VERIFYING` → `SECURITY_SCANNING` → `SCAN_FAILED_QUARANTINED` (종단/보안검역)
  - `SECURITY_SCANNING` → `READY_FOR_SUBMISSION` → `SUBMITTED_FOR_REVIEW` → `IN_REVIEW`
  - `IN_REVIEW` → `CHANGES_REQUESTED` (개발자 수정 후 재제출 가능) | `REJECTED`
  - `IN_REVIEW` → `APPROVED_PENDING_PUBLISH` (개발자의 지정 출시 시각 대기 또는 수동 출시 버튼 대기)
  - `APPROVED_PENDING_PUBLISH` → `PUBLISHED_STAGED` (`rollout_percentage < 100`) → `PUBLISHED_FULL` (`rollout_percentage = 100`)
  - `PUBLISHED_STAGED` / `PUBLISHED_FULL` → `HALTED` (배포 중단) | `SUPERSEDED` (신규 버전 100% 출시로 대체됨) | `REVOKED_BY_ADMIN` (보안/정책 위반 긴급 회수).

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **`web-dev`**: `ReleaseTrackBoard.vue`, `StagedRolloutSlider.vue`, `EmergencyHaltModal.vue`.
- **`web-admin`**: `ReleaseReviewWorkspace.vue` (이전 릴리스 대비 추가된 위험 권한 Diff, APK 크기 변화, 서명 키 일치 여부, 스캔 결과 한눈에 비교).

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/distribution/application/release_lifecycle_service.go`: 상태 전이 유효성 검사(`CanTransition(from, to, actorRole)`), Outbox 이벤트 발행, CDN 캐시 워밍/무효화 수행.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**: `app_releases`, `release_artifacts`, `release_reviews`, `release_status_history`.
- **제약조건**: `UNIQUE(distribution_id, version_code)` — 동일 배포처에서 동일한 `version_code`는 절대 재사용할 수 없으며, 신규 릴리스의 `version_code`는 기존 출시된 최대 `version_code`보다 strictly 커야 한다(`version_code > max_previous_version_code`).

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `POST /api/v1/developer/apps/{appId}/releases`: 신규 릴리스 초안 생성.
- `POST /api/v1/developer/apps/{appId}/releases/{releaseId}/submit`: 심사 제출.
- `PATCH /api/v1/developer/apps/{appId}/releases/{releaseId}/rollout`: 단계적 배포 비율 변경 또는 `HALT` 전환.
- `POST /api/v1/admin/reviews/{releaseId}/decision`: 관리자 심사 승인/반려 판정.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- 불법 상태 전이 차단: 예를 들어 `SECURITY_SCANNING`이 완료되지 않았거나 `SCAN_FAILED_QUARANTINED` 상태인 릴리스를 관리자라도 `PUBLISHED_FULL`로 강제 전이할 수 없도록 DB 트리거 및 도메인 가드로 이중 차단한다.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 심사 승인(`APPROVED_PENDING_PUBLISH`) 시점의 아티팩트 `sha256` 해시와 실제 배포 시점의 아티팩트 `sha256` 해시가 일치하는지 재검증하여 심사 후 파일 바꿔치기(TOCTOU 공격)를 원천 차단한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 개발자 콘솔에서 새 릴리스를 제출할 때 직전 버전 대비 변경 사항(예: `versionCode: 104 → 105`, `신규 추가 권한: android.permission.CAMERA`, `minSdk: 26 유지`) 요약 카드를 보여주어 실수를 방지한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **단계적 배포 결정론 테스트 (`TC-FUNC-007`)**: 10,000개의 가상 `device_id`로 `rollout_percentage = 20`인 앱의 업데이트 확인 API 호출 시 약 20%(±1.5%)만 신규 릴리스를 받고, 이후 `rollout_percentage = 50`으로 상향 시 기존 20% 기기가 모두 유지된 채 30%가 추가되는지(단조 증가성) 검증.
- **불법 상태 전이 차단 테스트 (`TC-FUNC-008`)**: 스캔 미완료 상태에서 심사 제출 또는 출시 API 호출 시 `409 ERR_INVALID_RELEASE_STATE_TRANSITION`이 반환되는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `D — Developer Platform`, `N — Native App Distribution`, `S — Security & Safety`, `V — Versioning & Update Engine`, `W — Workflow & Administration`과 연결된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0`
- **개발 Phase**: `Phase 3 (앱 등록 및 심사)` 및 `Phase 4 (배포)`에서 구현한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 유료 앱 판매 종료(Sunset) 시 기존 구매자 피해** → 앱이 신규 판매 종료(`ARCHIVED_FOR_NEW_USERS`)되더라도, 기존에 구매 권한(`user_entitlements`)을 보유한 사용자는 아카이브 스토리지에서 마지막 안정 버전을 계속 재다운로드할 수 있도록 보장한다(단, 악성코드로 인한 `REVOKED_BY_ADMIN`은 제외).

## 16. 최종 산출물 (Target Implementation Deliverables)
- 릴리스 상태 머신 및 단계적 배포 라우터 (`backend/internal/modules/distribution/*`).

## 17. Acceptance Criteria (수용 기준)
1. 12단계 릴리스 상태 전이가 모든 정상/예외 경로에서 명세대로 동작하고 `release_status_history`에 빠짐없이 기록되어야 한다.
2. `HALTED` 전환 즉시 1초 이내에 신규 기기에 대한 해당 버전 배포가 중단되고 이전 안정 버전이 반환되어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-DIST-001/release_state_machine_test.log`: 전체 상태 전이 및 불법 전이 차단 테스트 로그.
- `artifacts/evidence/AM-DIST-001/staged_rollout_monotonicity_test.log`: 단계적 배포 단조 증가성 검증 로그.

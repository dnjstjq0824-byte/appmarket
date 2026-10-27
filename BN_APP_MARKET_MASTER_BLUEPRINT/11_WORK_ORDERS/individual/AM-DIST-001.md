# AM-DIST-001 — APK/AAB 격리 업로드, ZIP/SHA-256 무결성 검증, AXML 파서, apksigner v2/v3 서명 검증 및 릴리스 상태 머신 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-DIST-001` |
| **작업명 (Title)** | APK/AAB 격리 업로드, ZIP/SHA-256 무결성 검증, AXML 파서, apksigner v2/v3 서명 검증 및 릴리스 상태 머신 구현 |
| **목표 (Objective)** | `REQ-N-001;REQ-N-002;REQ-L-001` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Backend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-DB-001;AM-API-001` |
| **개발 Phase 및 우선순위** | `Phase 3` / `P0` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/N_NATIVE_APP_DISTRIBUTION.md;07_DISTRIBUTION/ARTIFACT_PIPELINE.md;07_DISTRIBUTION/APK_AAB_SIGNING.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/distribution/;backend/internal/worker/handlers/artifact/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/distribution/infrastructure/axml_parser.go;backend/internal/modules/distribution/infrastructure/apksigner_verifier.go;backend/internal/modules/distribution/infrastructure/bundletool_runner.go;backend/internal/modules/distribution/domain/release_state_machine.go` |
| **API 의존성 (API Dependencies)** | `POST .../uploads;POST .../complete;POST .../releases;POST .../submit` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-007;TC-FUNC-008;TC-FUNC-009;TC-SEC-007` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-DIST-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. S3 격리 버킷(`bn-quarantine-bucket`) 멀티파트 업로드 세션 관리.
2. 격리 워커에서 Zip Bomb/Zip Slip 방어, SHA-256 대조, Binary `AndroidManifest.xml` 파싱(`package`, `versionCode`, `minSdk`, `targetSdk`, 권한, ABI 추출) 구현.
3. `apksigner` v2/v3 서명 블록 검증, X.509 인증서 SHA-256 지문 추출 및 기존 릴리스와의 서명 연속성(Key Lineage) 강제.
4. AAB 업로드를 위한 `bundletool build-apks` 프로세서 인터페이스 및 12단계 릴리스 상태 머신 구현.

## 2. Acceptance Criteria (수용 기준)
1. v1 서명만 있는 APK나 변조된 APK 업로드 시 `ERR_APK_V2_SIGNATURE_REQUIRED` / `REJECTED_INVALID_APK`로 100% 거부.
2. 기존 출시 버전과 다른 서명 키로 서명된 업데이트 APK 업로드 시 `REJECTED_SIGNATURE_MISMATCH`로 자동 차단.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 파싱 중 타임아웃(60초) 또는 비정상 크래시 발생 시 해당 아티팩트를 즉시 격리 종료하고 API 서버에는 영향이 없도록 보장.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-007;TC-FUNC-008;TC-FUNC-009;TC-SEC-007`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-DIST-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

# AM-DIST-003 — AAB(Android App Bundle) bundletool APK-Set 변환 워커 및 관리형 서명 키 볼트 연동 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-DIST-003` |
| **작업명 (Title)** | AAB(Android App Bundle) bundletool APK-Set 변환 워커 및 관리형 서명 키 볼트 연동 구현 |
| **목표 (Objective)** | `REQ-N-001;REQ-N-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Backend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-DIST-001` |
| **개발 Phase 및 우선순위** | `Phase 4` / `P1` |
| **입력 문서 (Input Documents)** | `07_DISTRIBUTION/APK_AAB_SIGNING.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/distribution/aab/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/distribution/aab/bundletool_executor.go;backend/internal/modules/distribution/aab/signing_vault_client.go` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/developer/apps/{appId}/uploads` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-009` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-DIST-003/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 격리 샌드박스 워커에서 `bundletool build-apks`를 호출하여 `.aab`로부터 Universal APK 및 기기 스펙별 Split APK 생성.
2. KMS/Vault에 보관된 앱별 서명키로 생성된 APK 세트를 자동 서명하고 `apksigner`로 검증.

## 2. Acceptance Criteria (수용 기준)
1. 샘플 `.aab` 업로드 시 Universal APK 추출, 서명, SHA-256 기록이 자동 완료됨.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 유효하지 않은 AAB 구조 시 `REJECTED_INVALID_APK` 및 상세 bundletool 표준에러 리포트 반환.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-009`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-DIST-003/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

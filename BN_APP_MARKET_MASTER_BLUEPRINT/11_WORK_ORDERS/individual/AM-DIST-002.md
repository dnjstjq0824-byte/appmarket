# AM-DIST-002 — 불변 공개 CDN 배포 워커, 서명된 다운로드 토큰 발급기, 기기 호환성 및 일괄 업데이트 판별 엔진 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-DIST-002` |
| **작업명 (Title)** | 불변 공개 CDN 배포 워커, 서명된 다운로드 토큰 발급기, 기기 호환성 및 일괄 업데이트 판별 엔진 구현 |
| **목표 (Objective)** | `REQ-N-003;REQ-V-001;REQ-V-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Backend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-DIST-001;AM-SEC-002` |
| **개발 Phase 및 우선순위** | `Phase 4` / `P0` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/V_VERSIONING_AND_UPDATE_ENGINE.md;07_DISTRIBUTION/STORAGE_CDN.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/distribution/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/distribution/application/cdn_publisher.go;backend/internal/modules/distribution/application/download_token_service.go;backend/internal/modules/distribution/domain/update_resolver.go` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/catalog/apps/{appId}/releases/{releaseId}/download-token;POST /api/v1/device/check-updates;PATCH .../rollout` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-012;TC-OPS-003` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-DIST-002/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 출시 승인된 아티팩트를 `bn-quarantine-bucket`에서 `bn-public-cdn-bucket`의 불변 경로(`/packages/{app_id}/{release_id}/{sha256}.apk`)로 복사하는 워커 구현.
2. 국가 차단, 유료 구매 권한(`user_entitlements`), 권리 등급(`rights_level`) 검증 후 15분 만료 HMAC Signed CDN URL 발급.
3. 기기 SDK/ABI/Feature 4차원 호환성 판별, 단계적 배포(`CRC32` 버킷) 평가, 서명 호환성(`COMPATIBLE_SAME_SIGNER` vs `INCOMPATIBLE_DIFFERENT_SIGNER`) 일괄 업데이트 리졸버 구현.

## 2. Acceptance Criteria (수용 기준)
1. 만료된 Signed URL 또는 `EXTERNAL_LINK_ONLY` 앱에 대한 다운로드 토큰 요청이 100% 거부.
2. 서명 인증서가 다른 패키지는 `UPDATE_AVAILABLE_SIGNER_CONFLICT`로 정확히 분류.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 긴급 차단(Kill Switch) 시 Redis 블랙리스트와 CDN Purge를 즉시 실행하여 기존 발급된 URL 접근도 Edge에서 차단.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-012;TC-OPS-003`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-DIST-002/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

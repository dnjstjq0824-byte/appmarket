# AM-MOD-001 — 설치 검증 사용자 리뷰, 어뷰징/조작 탐지 필터, 유해 앱 신고(DMCA) 및 개발자 이의신청 워크플로 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-MOD-001` |
| **작업명 (Title)** | 설치 검증 사용자 리뷰, 어뷰징/조작 탐지 필터, 유해 앱 신고(DMCA) 및 개발자 이의신청 워크플로 구현 |
| **목표 (Objective)** | `REQ-R-002;REQ-S-003;REQ-W-003` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Security Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-API-002;AM-DIST-002` |
| **개발 Phase 및 우선순위** | `Phase 6` / `P1` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/R_RECOMMENDATION_ENGINE.md;01_A_TO_Z_BLUEPRINT/W_WORKFLOW_AND_ADMINISTRATION.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/moderation/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/moderation/application/review_service.go;backend/internal/modules/moderation/application/anti_fraud_detector.go;backend/internal/modules/moderation/application/appeal_service.go` |
| **API 의존성 (API Dependencies)** | `GET/POST /api/v1/catalog/apps/{slug}/reviews;PUT .../reply;POST /api/v1/moderation/reports;POST /api/v1/developer/orgs/{orgId}/appeals` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-011;TC-SEC-008;TC-FUNC-013` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-MOD-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. `install_events` 또는 `user_entitlements`가 확인된 사용자만 리뷰를 작성할 수 있도록 검증하고 베이지안 가중 평점을 갱신.
2. 동일 기기/IP 대역 반복 다운로드 및 비정상 별점 폭격을 필터링하는 `AntiFraudDetector` 구현.
3. 신고 접수 SLA 큐 및 최초 처분 심사자와 다른 시니어 심사자에게만 배정되는 이의신청(`appeals`) 워크플로 구현.

## 2. Acceptance Criteria (수용 기준)
1. 미설치 계정의 리뷰 작성 시 `403 ERR_REVIEW_REQUIRES_INSTALLATION` 차단.
2. 최초 반려 심사자가 본인이 반려한 건의 이의신청 심사 시도 시 `403 ERR_APPEAL_SAME_REVIEWER_FORBIDDEN` 차단.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 어뷰징 점수 급증 시 해당 앱의 평점/차트 반영을 일시 동결하고 모더레이터 경보 발령.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-011;TC-SEC-008;TC-FUNC-013`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-MOD-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

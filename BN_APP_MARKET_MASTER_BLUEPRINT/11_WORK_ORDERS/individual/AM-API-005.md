# AM-API-005 — 인앱/이메일/푸시 알림 센터 및 전환율 퍼널 분석 이벤트 수집 API 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-API-005` |
| **작업명 (Title)** | 인앱/이메일/푸시 알림 센터 및 전환율 퍼널 분석 이벤트 수집 API 구현 |
| **목표 (Objective)** | `REQ-T-001;REQ-Y-001` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Backend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-API-001;AM-DB-002` |
| **개발 Phase 및 우선순위** | `Phase 6` / `P1` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/T_THIRD_PARTY_INTEGRATION.md;01_A_TO_Z_BLUEPRINT/Y_YIELD_ANALYTICS_AND_MONETIZATION.md;10_COMMERCE_ADS_ANALYTICS/ANALYTICS_EVENTS.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/notification/;backend/internal/modules/analytics/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/notification/service.go;backend/internal/modules/analytics/funnel_ingestor.go;backend/internal/modules/analytics/rollup_worker.go` |
| **API 의존성 (API Dependencies)** | `GET /api/v1/users/me/notifications;POST /api/v1/events/batch;POST /api/v1/events/install;GET /api/v1/developer/apps/{appId}/analytics` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-014;TC-OPS-006` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-API-005/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 인앱 알림 저장(`notifications`) 및 외부 이메일/FCM 푸시 발송 어댑터(서킷 브레이커 내장) 구현.
2. 노출→상세→다운로드→설치→결제 6단계 퍼널 이벤트 배치 수집 및 `app_metrics_daily` 롤업 워커 구현.

## 2. Acceptance Criteria (수용 기준)
1. 외부 이메일/푸시 서버 장애 시에도 인앱 알림은 100% 정상 저장 및 조회.
2. 주입된 퍼널 이벤트와 일별 집계 테이블 수치가 정확히 일치.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 외부 알림 발송 실패 시 서킷 브레이커 개방 및 백그라운드 큐 지수 백오프 재시도.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-014;TC-OPS-006`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-API-005/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

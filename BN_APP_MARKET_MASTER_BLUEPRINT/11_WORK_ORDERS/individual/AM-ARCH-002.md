# AM-ARCH-002 — Transactional Outbox Relay 워커, Redis Streams Consumer Group 및 DLQ Reaper 엔진 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-ARCH-002` |
| **작업명 (Title)** | Transactional Outbox Relay 워커, Redis Streams Consumer Group 및 DLQ Reaper 엔진 구현 |
| **목표 (Objective)** | `REQ-J-001;REQ-J-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Backend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-ARCH-001;AM-DB-001` |
| **개발 Phase 및 우선순위** | `Phase 1` / `P0` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/J_JOB_PROCESSING.md;02_SYSTEM_ARCHITECTURE/EVENT_FLOW.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/worker/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/worker/outbox_relay.go;backend/internal/worker/dispatcher.go;backend/internal/worker/reaper.go` |
| **API 의존성 (API Dependencies)** | `GET /api/v1/admin/jobs;POST /api/v1/admin/jobs/{jobId}/retry` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-006;TC-OPS-004` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-ARCH-002/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. `outbox_events` 폴링 및 Redis Streams 6대 큐 발행 Relay 구현.
2. `background_jobs` 멱등성 잠금(`FOR UPDATE SKIP LOCKED`), 지수 백오프 재시도, 60초 하트비트 Zombie Job Reaper, DLQ 수동 재처리 API 구현.

## 2. Acceptance Criteria (수용 기준)
1. 동일 멱등성 키 작업 10회 발행 시 정확히 1회만 실행.
2. 워커 `SIGKILL` 시 60초 내 타 워커가 작업 인계 완료.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 최대 재시도 횟수(5회) 초과 시 즉시 `DEAD_LETTER`로 격리하고 관리자 알림 트리거.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-006;TC-OPS-004`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-ARCH-002/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

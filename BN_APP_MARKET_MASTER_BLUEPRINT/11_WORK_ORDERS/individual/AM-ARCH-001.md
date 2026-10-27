# AM-ARCH-001 — 모노레포 초기화, Go 모듈형 모놀리스 코어 프레임워크 및 미들웨어 구축

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-ARCH-001` |
| **작업명 (Title)** | 모노레포 초기화, Go 모듈형 모놀리스 코어 프레임워크 및 미들웨어 구축 |
| **목표 (Objective)** | `REQ-A-001;REQ-A-002;REQ-A-003` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Lead Architect` |
| **선행 작업 ID (Prerequisites)** | `NONE` |
| **개발 Phase 및 우선순위** | `Phase 0` / `P0` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/A_ARCHITECTURE.md;02_SYSTEM_ARCHITECTURE/SYSTEM_OVERVIEW.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/cmd/;backend/internal/platform/;packages/;pnpm-workspace.yaml;go.work` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `android-client/;03_DATABASE/migrations/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/cmd/api/main.go;backend/cmd/worker/main.go;backend/internal/platform/httpserver/server.go;backend/internal/platform/errors/problem.go;backend/internal/architecture_test.go` |
| **API 의존성 (API Dependencies)** | `GET /healthz/live;GET /healthz/ready;GET /api/v1/system/capabilities` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-001;TC-OPS-001` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-ARCH-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. Go 1.23 모듈(`backend/go.mod`) 및 `pnpm-workspace.yaml` 초기화.
2. `RequestID`, `OTelTrace`, `RealIP`, `PanicRecovery`, `CORS`, `RateLimiter`, `IdempotencyGuard`, `ProblemDetails(RFC 9457)` 미들웨어 체인 구현.
3. 도메인 모듈 간 불법 import를 차단하는 `internal/architecture_test.go` 작성.
4. Transactional Outbox Relay 및 Redis Streams Worker Dispatcher 기본 루프 구현.

## 2. Acceptance Criteria (수용 기준)
1. `go build ./...` 및 `go test -v -race ./internal/...` 에러 0건 통과.
2. `/healthz/ready`가 DB/Redis/OpenSearch/S3 연결 상태를 정확히 반환.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 기동 시 필수 환경변수나 DB/Redis 연결 실패 시 즉시 구조화 에러 로그 출력 후 Non-Zero 코드로 Fail-Fast 종료.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-001;TC-OPS-001`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-ARCH-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

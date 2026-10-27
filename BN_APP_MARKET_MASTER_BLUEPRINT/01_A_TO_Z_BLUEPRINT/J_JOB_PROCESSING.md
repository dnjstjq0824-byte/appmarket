# J — Job Processing — 비동기 작업 큐, 격리 워커, 멱등성 보장 및 실패 복구(DLQ) 아키텍처 (Asynchronous Job Queue, Isolated Workers, Idempotency & Dead-Letter Recovery)

- **문서 ID**: `DOC-AZ-J`
- **알파벳 코드**: `J`
- **관련 요구사항 ID**: `REQ-J-001`, `REQ-J-002`
- **연관 Work Order ID**: `AM-DIST-001`, `AM-SEARCH-001`, `AM-SEC-002`
- **연관 테스트 케이스 ID**: `TC-OPS-004`, `TC-FUNC-006`
- **참조 상세 설계 문서**: [02_SYSTEM_ARCHITECTURE/EVENT_FLOW.md](../02_SYSTEM_ARCHITECTURE/EVENT_FLOW.md), [07_DISTRIBUTION/ARTIFACT_PIPELINE.md](../07_DISTRIBUTION/ARTIFACT_PIPELINE.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 대용량 APK/AAB 파일의 무결성 검증, AXML 매니페스트 추출, 서명 검증, 악성코드 정적/동적 스캔, `bundletool` APK 세트 빌드, OpenSearch 색인 동기화, 외부 스토어 피드 수집, 월간 정산 원장 집계는 수 초~수 분이 소요되는 작업이므로 동기식 HTTP 요청 스레드에서 처리하면 서버가 마비된다. 이를 안전하고 확장 가능하게 처리할 비동기 작업 엔진이 필요하다.
- **달성 목표**: Transactional Outbox + Redis Streams Consumer Group 기반의 6대 워커 큐, 분산 락 기반 중복 실행 방지, 지수 백오프 재시도, Dead Letter Queue(DLQ) 격리 및 운영자 수동 재처리 체계를 확립한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **시스템 워커 프로세스 (`worker-*`)**: 할당된 큐에서 작업을 가져와 격리 실행하고 결과를 DB에 기록.
- **Platform Admin / SRE**: 관리자 콘솔에서 큐 대기열 깊이, 워커 처리량, DLQ 실패 작업 목록을 조회하고 원인 해결 후 재시도(`Retry Job`) 트리거.

## 3. 필수 기능 (Mandatory Features)
1. **6대 전용 작업 큐 분리 (Priority & Isolation Queues)**:
   - `q:artifact:verify` (P0): 업로드된 APK/AAB의 SHA-256 검증, ZIP 구조 검사, AXML 메타데이터 추출, `apksigner` 서명 검증.
   - `q:artifact:malware_scan` (P0): ClamAV, YARA 룰셋, 권한 위험도 점수 산출, DEX 문자열 정적 분석.
   - `q:artifact:publish_cdn` (P0): 심사 승인된 바이너리를 격리 버킷에서 공개 CDN 버킷으로 복사 및 `bundletool` 변환.
   - `q:search:indexer` (P1): 카탈로그/리뷰/평점 변경 시 OpenSearch 문서 실시간/마이크로배치 갱신.
   - `q:connector:sync` (P2): F-Droid 및 외부 제휴 피드 주기적 수집·정규화.
   - `q:commerce:settlement` (P1): 결제 웹훅 후처리, 정산 배치 계산, 이메일/푸시 알림 발송.
2. **정확히 한 번 처리 효과 (At-Least-Once Delivery + Idempotent Consumer)**:
   - 모든 Job은 고유 `job_id`와 도메인 멱등성 키(`idempotency_key`, 예: `scan:{artifact_id}:{sha256}`)를 가지며, 워커는 실행 전 `background_jobs` 테이블의 상태를 원자적으로 확인(`SELECT ... FOR UPDATE SKIP LOCKED`)하여 중복 실행을 완벽히 방지한다.
3. **재시도 및 Dead Letter Queue (DLQ)**:
   - 일시적 네트워크/스토리지 오류(`TransientError`)는 최대 5회까지 지수 백오프(`5s, 25s, 125s, 625s, 3125s` + Jitter)로 자동 재시도한다.
   - 영구적 오류(`PermanentError`, 예: 손상된 ZIP 헤더) 또는 5회 재시도 초과 작업은 즉시 `DEAD_LETTER` 상태로 전환하고 관리자 알림을 발송한다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **Zombie Job 회수 (Stale Lock Reaper)**:
  - 워커가 작업 도중 OOM이나 강제 종료(`SIGKILL`)로 죽으면, 워커는 실행 중 15초마다 `heartbeat_at`을 갱신하며, `Reaper` 고루틴이 60초 이상 하트비트가 끊긴 `RUNNING` 작업을 찾아 안전하게 재큐잉(`QUEUED`)하거나 크래시 횟수 초과 시 `DEAD_LETTER`로 격리한다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **`web-admin`**: `JobQueueMonitorView.vue` (큐별 Pending/Running/Completed/DLQ 카운트, 실패 스택트레이스 조회, 단일/일괄 재시도 버튼).

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/worker/dispatcher.go`, `internal/worker/reaper.go`, `internal/worker/handlers/*`.
- 각 핸들러는 `context.WithTimeout`을 통해 작업 종류별 엄격한 최대 실행 시간(예: APK 파싱 60초, 악성코드 스캔 300초)을 강제한다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**:
  - `background_jobs` (`id UUID PK`, `queue_name VARCHAR(64)`, `idempotency_key VARCHAR(128) UNIQUE`, `payload JSONB`, `status VARCHAR(24)`, `attempt_count INT`, `max_attempts INT`, `last_error TEXT`, `worker_id VARCHAR(64)`, `heartbeat_at TIMESTAMPTZ`, `scheduled_at TIMESTAMPTZ`, `completed_at TIMESTAMPTZ`).

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `GET /api/v1/admin/jobs`: 큐별 작업 상태 및 DLQ 목록 조회.
- `POST /api/v1/admin/jobs/{jobId}/retry`: DLQ에 빠진 작업을 초기 상태로 재실행.
- `POST /api/v1/admin/jobs/{jobId}/cancel`: 대기 중인 작업 취소.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **작업 상태 머신**: `QUEUED` → `RUNNING` → `SUCCEEDED` | (`RETRY_SCHEDULED` → `RUNNING`) → `DEAD_LETTER` | `CANCELED`.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- Job 페이로드(`payload JSONB`)에는 평문 비밀번호나 개인키를 절대 저장하지 않으며, 참조 ID(`artifact_id`, `order_id`)만 전달한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 개발자가 APK를 업로드한 직후 콘솔 화면에서 `파일 무결성 검사 중 (1/3) → 서명 및 매니페스트 분석 중 (2/3) → 보안 스캔 중 (3/3)` 단계별 진행 상태를 실시간(폴링/SSE)으로 확인할 수 있도록 표시한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **중복 메시지 멱등성 테스트 (`TC-FUNC-006`)**: 동일한 `idempotency_key`를 가진 APK 분석 이벤트를 큐에 10번 연속 발행했을 때 실제 분석 및 결과 레코드 생성이 정확히 1회만 수행되는지 검증.
- **워커 강제 종료 복구 테스트 (`TC-OPS-004`)**: 실행 중인 워커에 `SIGKILL` 전송 후 Reaper가 60초 뒤 작업을 회수하여 다른 워커가 정상 완료하는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `N — Native App Distribution`, `S — Security & Safety`, `O — OpenSearch`, `E — External Store Connectors`의 백그라운드 실행 엔진이다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0`
- **개발 Phase**: 프레임워크는 `Phase 1`, APK/보안 워커는 `Phase 3~4`, 검색/커넥터 워커는 `Phase 5`에서 구현한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 대량 외부 피드 동기화로 인해 긴급한 APK 출시 배포 작업이 밀리는 현상(Head-of-Line Blocking)** → 큐와 워커 풀을 물리적으로 분리(`q:artifact:*` 워커 풀과 `q:connector:sync` 워커 풀 분리)하여 외부 동기화 부하가 자체 앱 심사/배포 속도에 전혀 영향을 주지 않도록 격리한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 비동기 작업 큐 디스패처, Reaper, 6종 워커 핸들러 (`backend/internal/worker/*`).

## 17. Acceptance Criteria (수용 기준)
1. 모든 백그라운드 작업이 DB `background_jobs` 테이블에서 상태 추적 가능해야 한다.
2. 의도적으로 발생시킨 영구 오류 작업이 지정된 횟수 후 정확히 `DEAD_LETTER` 상태로 전이되고 관리자 API로 조회·재시도 가능해야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-DIST-001/worker_idempotency_dlq.log`: 중복 방지 및 DLQ 전이·재시도 통합 테스트 로그.

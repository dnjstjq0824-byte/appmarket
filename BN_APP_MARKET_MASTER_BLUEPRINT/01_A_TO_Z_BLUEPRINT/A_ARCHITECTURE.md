# A — Architecture — 독립 서비스 전체 아키텍처 및 모듈 경계 설계 (Master System Architecture, Module Boundaries & Fault Isolation)

- **문서 ID**: `DOC-AZ-A`
- **알파벳 코드**: `A`
- **관련 요구사항 ID**: `REQ-A-001`, `REQ-A-002`, `REQ-A-003`
- **연관 Work Order ID**: `AM-ARCH-001`, `AM-OPS-001`
- **연관 테스트 케이스 ID**: `TC-FUNC-001`, `TC-OPS-001`, `TC-SEC-001`
- **참조 상세 설계 문서**: [02_SYSTEM_ARCHITECTURE/SYSTEM_OVERVIEW.md](../02_SYSTEM_ARCHITECTURE/SYSTEM_OVERVIEW.md), [02_SYSTEM_ARCHITECTURE/COMPONENT_DIAGRAMS.md](../02_SYSTEM_ARCHITECTURE/COMPONENT_DIAGRAMS.md), [02_SYSTEM_ARCHITECTURE/DEPLOYMENT_ARCHITECTURE.md](../02_SYSTEM_ARCHITECTURE/DEPLOYMENT_ARCHITECTURE.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: Type 1(독립 앱스토어), Type 2(연합형 외부 앱 검색엔진), Type 3(통합 비교 마켓플레이스)를 중복 개발 없이 단일 플랫폼으로 통합하면서도, TreeScan 본체와 100% 분리된 독립 실행 능력과 악성 바이너리 처리 시의 완벽한 장애·보안 격리를 보장하기 위해 필요하다.
- **달성 목표**: Go 기반 모듈형 모놀리스(Modular Monolith) 코어 API와 물리적으로 격리된 비동기 샌드박스 워커(`workers`), 3개의 독립 웹 프론트엔드(`web-user`, `web-dev`, `web-admin`), 네이티브 Android 클라이언트(`android-client`), PostgreSQL 16, OpenSearch 2.x, Redis 7.x, S3 이중 버킷(Quarantine/Public) 및 글로벌 CDN의 전체 토폴로지를 정의한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **일반 사용자 (End User)**: `web-user`(SSR 웹) 및 `android-client`를 통해 Edge CDN/WAF를 거쳐 Public Read API 및 다운로드 스트림을 소비한다.
- **개발자 조직 (Developer Org)**: `web-dev` 콘솔 및 CI/CD API Key를 통해 Developer API 및 대용량 바이너리 업로드 게이트웨이에 접근한다.
- **플랫폼 관리자 (Platform Admin)**: VPN/Zero-Trust Access 및 MFA가 강제된 `web-admin` 콘솔을 통해 Admin API 및 심사·관제 워크플로를 제어한다.
- **시스템 워커 (Isolated Worker Service Account)**: 내부 전용 프라이빗 서브넷에서 메시지 큐를 구독하고 S3 격리 버킷의 바이너리를 검사한다.

## 3. 필수 기능 (Mandatory Features)
1. **10대 독립 컴포넌트 자립 실행**: TreeScan 없이도 단독 도메인·단독 DB·단독 인증으로 100% 구동되는 아키텍처.
2. **Go 모듈형 모놀리스 도메인 경계 강제**: `internal/modules/{iam,catalog,developer,distribution,search,commerce,ads,moderation,admin,treescan}` 간 순환 참조 금지 및 인터페이스 기반 통신.
3. **고위험 작업 프로세스 격리**: APK/AAB 압축 해제, AXML 파싱, `apksigner` 검증, YARA/ClamAV 악성코드 스캔, `bundletool` 실행, 외부 스토어 피드 파싱은 API 서버와 분리된 무권한 컨테이너(`workers`)에서 실행.
4. **이중 스토리지 아키텍처**: 검증 전 파일이 저장되는 `bn-quarantine-bucket`(외부 접근 완전 차단)과 출시 승인된 불변 아티팩트만 저장되는 `bn-public-cdn-bucket` 분리.
5. **Transactional Outbox 기반 이벤트 일관성**: PostgreSQL 트랜잭션 커밋과 동시에 `outbox_events`에 기록하고, Relay Worker가 Redis Streams/OpenSearch로 비동기 전파하여 데이터 유실 방지.

## 4. 상세 기능 (Detailed Functional Specifications)
- **네트워크 3-Tier 분리**:
  - **Public Edge Tier**: Cloudflare/AWS WAF → L7 Load Balancer (Envoy/Nginx Ingress) → CDN Edge.
  - **Application Tier**: `web-user` (Nuxt SSR), `web-dev` (SPA 정적 서빙 + API), `web-admin` (IP/mTLS 제한 + SPA), `backend-api` (Go 1.23 Stateless Pod).
  - **Worker & Data Tier (Private Subnet, No Inbound Internet)**: `worker-apk-analyzer`, `worker-malware-scanner` (Egress 차단), `worker-connector` (허용된 도메인만 Egress 허용), PostgreSQL 16 Primary/Replica, OpenSearch 3-Node 클러스터, Redis Sentinel/Cluster.
- **모듈 간 통신 규칙**:
  - Go 모놀리스 내부의 타 도메인 데이터 변경은 반드시 대상 모듈의 Application Service 인터페이스를 통해서만 호출하며, 타 모듈 소유 DB 테이블에 직접 `INSERT/UPDATE` 쿼리를 실행하는 것을 `golangci-lint` (`depguard`) 규칙으로 차단한다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **모노레포 워크스페이스 구조 (`pnpm workspaces`)**:
  - `packages/design-system/`: 공통 디자인 토큰, 접근성 준수 UI 컴포넌트(`BnButton`, `BnModal`, `BnDataTable`, `BnBadge`), 다국어/통화 포맷터.
  - `apps/web-user/`: Nuxt 3 + Vue 3 + TypeScript. SSR 및 Route Rules 기반 ISR 캐싱 적용.
  - `apps/web-dev/`: Vue 3 + Vite + TypeScript + Pinia + Vue Router. 개발자 전용 대시보드 및 TUS/Multipart 업로더 포함.
  - `apps/web-admin/`: Vue 3 + Vite + TypeScript. RBAC 라우트 가드, 심사 비교 뷰어(Diff Viewer), 이중 승인 모달 포함.

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- **Go 디렉터리 표준 (`backend/`)**:
  - `cmd/api/main.go`: HTTP API 서버 엔트리포인트 (Graceful Shutdown, OTel 초기화, 미들웨어 체인 구성).
  - `cmd/worker/main.go`: 백그라운드 큐 컨슈머 엔트리포인트 (`--role=scanner|analyzer|indexer|connector|settlement`).
  - `internal/platform/`: DB 커넥션 풀(`pgxpool`), Redis 클라이언트, S3 Presigner, OTel Tracer, 구조화 로거(`slog`), RFC 9457 에러 렌더러.
  - `internal/modules/<domain>/{domain,application,infrastructure,presentation}`: 헥사고날/클린 아키텍처 4계층 구조.
- **미들웨어 체인 순서**: `RequestID` → `OTelTrace` → `RealIP` → `PanicRecovery` → `CORS` → `RateLimiter(Redis)` → `AuthJWT/APIKey` → `RBACGuard` → `IdempotencyGuard` → `AuditLogger` → `Handler`.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 인프라 테이블**:
  - `outbox_events` (`id UUID PK`, `aggregate_type VARCHAR(64)`, `aggregate_id UUID`, `event_type VARCHAR(128)`, `payload JSONB`, `status VARCHAR(24)`, `created_at TIMESTAMPTZ`, `processed_at TIMESTAMPTZ`).
  - `idempotency_keys` (`key_hash CHAR(64) PK`, `actor_id UUID`, `endpoint VARCHAR(128)`, `request_hash CHAR(64)`, `response_status INT`, `response_body JSONB`, `expires_at TIMESTAMPTZ`).
  - 전체 32개 테이블 명세는 [03_DATABASE/TABLE_DICTIONARY.md](../03_DATABASE/TABLE_DICTIONARY.md) 및 `03_DATABASE/migrations/0001_initial_schema.up.sql`과 100% 일치한다.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- **헬스체크 및 메타 엔드포인트**:
  - `GET /healthz/live`: 프로세스 생존 확인 (`200 OK`).
  - `GET /healthz/ready`: PostgreSQL, Redis, OpenSearch, S3 연결 상태 종합 점검 (`200 OK` 또는 `503 Service Unavailable`).
  - `GET /api/v1/system/capabilities`: 지원되는 배포 모드, 최소 Android 클라이언트 버전, 활성 기능 플래그 반환.
- **내부 이벤트 계약 (Outbox → Redis Streams)**:
  - `artifact.uploaded.v1`, `artifact.scanned.v1`, `release.approved.v1`, `catalog.updated.v1`, `connector.synced.v1`, `order.paid.v1`.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **시스템 가용성 상태 전이**:
  - `NORMAL` → (OpenSearch 장애 감지 시) → `DEGRADED_SEARCH_FALLBACK` (PostgreSQL `pg_trgm` 비상 검색으로 자동 전환) → (복구 확인 후) → `NORMAL`.
  - `NORMAL` → (보안 침해 또는 스토리지 이상 감지 시) → `UPLOAD_FROZEN` (신규 APK 업로드 및 배포 승인 일시 정지, 기존 조회만 허용).
- **예외 처리 원칙**: 모든 패닉은 미들웨어에서 포착하여 스택트레이스를 내부 로그에만 남기고, 클라이언트에는 표준 에러 코드(`ERR_INTERNAL_SERVER_ERROR`, HTTP 500)와 `trace_id`만 반환한다.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- **네트워크 및 컨테이너 보안**:
  - 모든 컨테이너는 `non-root` 사용자(UID 10001), `readOnlyRootFilesystem: true`, `allowPrivilegeEscalation: false`, Linux Capabilities `DROP ALL`로 실행한다.
  - 악성코드 분석 워커는 외부 인터넷 통신이 완전히 차단된 격리 네트워크(`network_mode: internal-quarantine`)에서 동작하여 악성 APK의 C2 통신이나 SSRF를 원천 차단한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 시스템 장애 또는 부분 점검(`DEGRADED`) 시 사용자 웹(`web-user`), 개발자 콘솔(`web-dev`), Android 클라이언트 상단에 실시간 상태 배너(`SystemStatusBanner`)를 노출하고, 읽기 전용 모드 전환 여부를 명확히 고지한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **아키텍처 경계 테스트 (`TC-FUNC-001`)**: Go `go test ./internal/architecture_test.go`를 통해 모듈 간 불법 패키지 import 여부를 정적 검사.
- **장애 격리 테스트 (`TC-OPS-001`)**: Chaos 테스트로 OpenSearch 컨테이너 또는 Malware Scanner 워커에 OOM을 유발했을 때 `backend-api`가 중단되지 않고 정상 응답하는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- 본 `A — Architecture` 모듈은 모든 후속 모듈(`B`~`Z`)의 기반 레이어이며, `H — Hosting & Infrastructure`, `I — Identity & Access`, `P — PostgreSQL`, `Z — Zero-Trust & TreeScan`과 직접 결합된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0 (최우선 선행 필수)`
- **개발 Phase**: `Phase 0 (요구사항, 저장소, 아키텍처 기반 구축)` — 모든 Work Order의 최상위 루트인 `AM-ARCH-001`에서 구현한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오 1: 메시지 큐(Redis Streams) 지연 또는 중단** → Transactional Outbox 테이블(`outbox_events`)에 이벤트가 안전하게 영속화되어 있으므로 큐 복구 즉시 유실 없이 재발행(Replay) 수행.
- **장애 시나리오 2: 악성 조작된 초대형 APK(Zip Bomb)로 인한 워커 메모리 고갈** → API 서버가 아닌 격리 워커만 재시작되며, 해당 Job의 `retry_count`가 임계치(3회) 도달 시 즉시 `QUARANTINED_CRASH`로 마킹하고 DLQ로 이동.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 모노레포 루트 설정 (`go.work` 또는 `backend/go.mod`, `pnpm-workspace.yaml`, `docker-compose.yml`).
- Go 모듈형 모놀리스 코어 프레임워크 및 미들웨어 패키지 (`backend/internal/platform/*`).
- Outbox Relay 및 Worker Dispatcher 프레임워크 (`backend/internal/worker/*`).

## 17. Acceptance Criteria (수용 기준)
1. `docker compose up --build` 실행 시 PostgreSQL, Redis, OpenSearch, MinIO(S3 호환), Backend API, Worker, 3개 Web App이 단일 명령으로 기동되고 `/healthz/ready`가 `200 OK`를 반환해야 한다.
2. TreeScan 관련 외부 서버나 환경변수가 전혀 없는 상태에서도 모든 서비스가 100% 정상 기동되어야 한다.
3. 모듈 간 순환 참조 및 금지된 패키지 참조가 0건이어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-ARCH-001/healthz_ready_response.json`: 전체 컴포넌트 `UP` 상태 응답 캡처.
- `artifacts/evidence/AM-ARCH-001/go_test_arch.log`: 아키텍처 의존성 규칙 테스트 통과 로그.
- `artifacts/evidence/AM-ARCH-001/docker_ps_inspect.json`: 비루트(UID 10001) 및 격리 네트워크 설정 검증 덤프.

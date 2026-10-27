# AM-OPS-001 — Docker 멀티스테이지 컨테이너, 3중 분리 스토리지, OpenTelemetry 관제 및 CI/CD 파이프라인 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-OPS-001` |
| **작업명 (Title)** | Docker 멀티스테이지 컨테이너, 3중 분리 스토리지, OpenTelemetry 관제 및 CI/CD 파이프라인 구현 |
| **목표 (Objective)** | `REQ-H-001;REQ-H-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `DevOps Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-ARCH-001` |
| **개발 Phase 및 우선순위** | `Phase 0` / `P0` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/H_HOSTING_AND_INFRASTRUCTURE.md;14_INFRASTRUCTURE_DEVOPS/DEPLOYMENT_PLAN.md;14_INFRASTRUCTURE_DEVOPS/CI_CD_REQUIREMENTS.md;14_INFRASTRUCTURE_DEVOPS/MONITORING.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `deploy/;scripts/;.github/workflows/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `backend/internal/modules/;apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `deploy/docker-compose.yml;deploy/Dockerfile.backend;deploy/Dockerfile.worker;deploy/otel-collector-config.yaml;deploy/prometheus/alerts.yml;.github/workflows/ci-verify.yml` |
| **API 의존성 (API Dependencies)** | `GET /metrics` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-OPS-001;TC-OPS-004` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-OPS-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 비루트(UID 10001) 멀티스테이지 Dockerfile 및 격리 서브넷(`internal-quarantine`)이 구성된 `docker-compose.yml` 작성.
2. OpenTelemetry Collector, Prometheus, Grafana 대시보드 및 큐 적체/5xx 에러율 알람 룰 구성.
3. SAST(`gosec`), 컨테이너 스캔(`trivy`), 시크릿 스캔(`gitleaks`), 계약 검증을 수행하는 CI 파이프라인 구축.

## 2. Acceptance Criteria (수용 기준)
1. 모든 컨테이너가 비루트 권한으로 기동되고 스캐너 워커의 외부 인터넷 Egress가 차단됨.
2. CI 파이프라인에서 시크릿/취약점 검출 시 자동 빌드 차단.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 롤링 배포 중 헬스체크 실패 시 이전 컨테이너 이미지로 자동 롤백.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-OPS-001;TC-OPS-004`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-OPS-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

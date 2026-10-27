# X — eXecution & Agent Orchestration — 멀티 에이전트 작업 분담, 실행 순서(DAG), 파일 소유권 격리 및 완료 보고 체계 (Multi-Agent Execution Sequence, Role Matrix, File Ownership Isolation & Reporting Protocol)

- **문서 ID**: `DOC-AZ-X`
- **알파벳 코드**: `X`
- **관련 요구사항 ID**: `REQ-X-001`, `REQ-X-002`
- **연관 Work Order ID**: `AM-ARCH-001` ~ `AM-TREESCAN-001` (전체 41개 Work Order 오케스트레이션)
- **연관 테스트 케이스 ID**: `TC-OPS-007`
- **참조 상세 설계 문서**: [START_HERE.md](../START_HERE.md), [11_WORK_ORDERS/MASTER_WORK_ORDER.md](../11_WORK_ORDERS/MASTER_WORK_ORDER.md), [12_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md](../12_AGENT_EXECUTION/AGENT_ROLE_MATRIX.md), [12_AGENT_EXECUTION/EXECUTION_SEQUENCE.md](../12_AGENT_EXECUTION/EXECUTION_SEQUENCE.md), [12_AGENT_EXECUTION/CONFLICT_PREVENTION.md](../12_AGENT_EXECUTION/CONFLICT_PREVENTION.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 제14조, 제15조, 제16조에 따라 복수의 자율 개발 에이전트(10개 전문 역할)가 이 ZIP 패키지를 전달받아 동시에 개발을 진행할 때, 작업 순서가 뒤엉키거나 동일한 파일을 동시에 수정하여 Git 충돌(Merge Conflict) 및 인터페이스 불일치가 발생하는 것을 원천 차단하기 위해 필요하다.
- **달성 목표**: 10대 에이전트 역할과 디렉터리 독점 소유권(Write Ownership), 계약 우선(Contract-First) 실행 순서(Wave 0 ~ Wave 10), 공유 계약 변경 RFC 절차, 그리고 독립 감사 기반의 완료 보고 체계를 정의한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **10대 전문 개발 에이전트 (제16조 필수 역할)**:
  1. `Lead Architect Agent`: 전체 아키텍처, 공유 계약(`openapi.yaml`, DDL) 변경 승인, 에이전트 간 의존성 조율.
  2. `Database Agent`: `03_DATABASE/migrations/`, `sqlc` 쿼리 정의, 인덱스/파티션 튜닝 전담.
  3. `Backend Agent`: `backend/internal/modules/{iam,catalog,developer,commerce,ads,moderation,admin}` API 및 비즈니스 로직 구현.
  4. `Search & Connector Agent`: `backend/internal/modules/{search,connector,treescan}` 및 OpenSearch 매핑/파이프라인 구현.
  5. `Security Agent`: `backend/internal/modules/security`, APK 서명/악성코드 스캔 워커, RBAC/MFA/감사 체인 검증.
  6. `Frontend Agent`: `apps/web-user`, `apps/web-dev`, `apps/web-admin`, `packages/ui-kit` 구현.
  7. `Android Agent`: `android-client/` 네이티브 Kotlin + Jetpack Compose 앱 및 `PackageInstaller` 엔진 구현.
  8. `DevOps Agent`: `deploy/`, `Dockerfile`, CI/CD 파이프라인, S3/CDN 설정, OpenTelemetry 모니터링 및 백업 스크립트 구현.
  9. `QA / E2E Agent`: `e2e/`, `loadtest/`, `backend/test/integration/` 자동화 검증 스위트 작성 및 실행.
  10. `Independent Audit Agent`: 코드 작성에 일절 참여하지 않고, 제3자 입장에서 모든 Work Order의 Acceptance Criteria와 증빙 파일의 진위를 독립 검증하여 최종 `VERIFIED` 승인권 행사.

## 3. 필수 기능 (Mandatory Features)
1. **파일 및 디렉터리 독점 소유권 (Single-Writer Directory Ownership)**:
   - 특정 디렉터리는 한 시점에 지정된 담당 에이전트만 수정(`Write`)할 수 있으며, 타 에이전트는 읽기(`Read-Only`)만 허용된다.
2. **계약 우선 병렬화 (Contract-First Wave Execution)**:
   - `Wave 0 (Foundation & Contracts)`: `AM-ARCH-001`, `AM-DB-001`, `AM-DB-002`를 먼저 완료하여 PostgreSQL 스키마와 OpenAPI 3.1 명세를 동결(Freeze)한다.
   - 계약이 동결되면 Backend Agent, Frontend Agent, Android Agent, Search Agent, Security Agent가 동일한 `openapi.yaml`과 DB 스키마를 바라보고 충돌 없이 병렬로 작업한다.
3. **자가 승인 금지 (No Self-Certification)**:
   - 코드를 구현한 에이전트는 상태를 최대 `IMPLEMENTED_UNVERIFIED`까지만 변경할 수 있으며, `VERIFIED`로의 전환은 오직 `Independent Audit Agent`만 수행할 수 있다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **공유 계약 변경 프로토콜 (Shared Contract RFC Workflow)**:
  - 개발 도중 `openapi.yaml`이나 DB 마이그레이션 수정이 불가피할 경우:
    1. 요청 에이전트가 `docs/rfc/RFC-XXXX.md`에 변경 이유, 영향받는 Work Order 목록, 하위 호환성 유지 방안을 작성.
    2. `Lead Architect Agent`가 검토 후 승인하면 `Database Agent` 또는 `Lead Architect`가 계약을 갱신하고 버전을 올림.
    3. 연관된 클라이언트/백엔드 Work Order에 재검증 플래그를 자동 부여.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- `Frontend Agent`는 `apps/web-user/`, `apps/web-dev/`, `apps/web-admin/`, `packages/ui-kit/` 디렉터리에 대한 독점 수정 권한을 가지며, `backend/`나 `android-client/` 파일은 절대 수정하지 않는다.

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- 백엔드 모듈별 디렉터리 소유권 분리:
  - 코어/배포/상거래: `Backend Agent` (`internal/modules/{iam,catalog,developer,distribution,commerce,ads,admin}`)
  - 검색/외부커넥터/TreeScan: `Search & Connector Agent` (`internal/modules/{search,connector,treescan}`)
  - 보안/스캐너/모더레이션: `Security Agent` (`internal/modules/{security,moderation}`).

## 7. DB 및 데이터 모델 (Database & Data Model)
- `Database Agent`가 `backend/db/migrations/` 및 `backend/db/queries/`를 독점 관리한다.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `04_API_SPECIFICATION/openapi.yaml`은 Single Source of Truth 계약 파일로서 `Lead Architect Agent`의 승인 하에만 수정된다.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **Work Order 실행 상태 머신**: `NOT_STARTED` → `IN_PROGRESS` → `IMPLEMENTED_UNVERIFIED` → `AUDIT_REVIEW` → `VERIFIED` | (`FAILED` → `IN_PROGRESS`) | `BLOCKED`.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 각 에이전트의 Git 커밋은 브랜치 보호 규칙(`CODEOWNERS`)에 의해 자신이 소유하지 않은 디렉터리 변경이 포함될 경우 CI에서 자동 거부(`ERR_OWNERSHIP_VIOLATION`)된다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- N/A (에이전트 오케스트레이션 명세).

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **CODEOWNERS 및 디렉터리 경계 준수 검증 (`TC-OPS-007`)**: 각 Work Order의 `수정 가능 디렉터리`와 `변경 금지 디렉터리`가 서로 충돌하지 않고 전체 코드베이스를 빠짐없이 커버하는지 스크립트로 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `11_WORK_ORDERS/` 및 `12_AGENT_EXECUTION/`의 모든 문서와 직결된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0 (개발 착수 전 필수 숙지)`
- **개발 Phase**: 전 Phase(`Phase 0` ~ `Phase 10`)에 걸쳐 적용되는 실행 거버넌스이다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 선행 Work Order 지연으로 인한 후행 에이전트 대기(Blocking)** → `openapi.yaml` 기반의 OpenAPI Mock Server(`Prism` / 계약 기반 테스트 스텁 — 단, 개발 중 클라이언트 테스트 용도로만 사용하며 운영 코드나 최종 검증에서는 실서버 필수)를 사용하여 프론트엔드/Android UI 개발은 병렬 진행하되, 최종 통합 검증은 실서버 백엔드 완료 후 수행한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- `11_WORK_ORDERS/MASTER_WORK_ORDER.md`, `WORK_ORDER_INDEX.csv`, `WORK_ORDER_DEPENDENCIES.md`, `11_WORK_ORDERS/individual/AM-*.md`(41개), `12_AGENT_EXECUTION/*`(5개).

## 17. Acceptance Criteria (수용 기준)
1. 모든 41개 Work Order에 담당 에이전트, 선행 작업 ID, 수정 가능/금지 디렉터리, 예상 생성 파일, 필수 테스트, 완료 증빙 경로가 명시되어야 한다.
2. 작업 의존성 그래프(DAG)에 순환 의존성(Cycle)이 0건이어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/BLUEPRINT_VERIFY/work_order_dag_check.json`: 41개 Work Order ID 유일성 및 비순환 DAG 검증 결과.

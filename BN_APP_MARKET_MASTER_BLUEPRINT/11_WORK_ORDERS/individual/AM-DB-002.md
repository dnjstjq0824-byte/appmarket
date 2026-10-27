# AM-DB-002 — 상거래 복식부기 원장 트리거, 월별 로그 파티셔닝 및 고성능 인덱스 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-DB-002` |
| **작업명 (Title)** | 상거래 복식부기 원장 트리거, 월별 로그 파티셔닝 및 고성능 인덱스 구현 |
| **목표 (Objective)** | `REQ-P-003;REQ-M-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Database Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-DB-001` |
| **개발 Phase 및 우선순위** | `Phase 1` / `P0` |
| **입력 문서 (Input Documents)** | `03_DATABASE/INDEX_AND_QUERY_DESIGN.md;03_DATABASE/migrations/0002_indexes_and_constraints.up.sql` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/db/migrations/;backend/db/queries/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/db/migrations/0002_indexes_and_constraints.up.sql;backend/db/migrations/0003_seed_reference_data.up.sql;backend/db/partition_manager.go` |
| **API 의존성 (API Dependencies)** | `NONE` |
| **DB 변경 여부 (DB Schema Changes)** | `YES (Indexes, Partitions, Triggers, Seeds)` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-DB-003;TC-COM-005` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-DB-002/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. `verify_ledger_transaction_balance()` DEFERRABLE CONSTRAINT TRIGGER 생성하여 대차평형(`SUM(DEBIT) = SUM(CREDIT)`) 강제.
2. `install_events` 및 `search_query_logs` 월별 Range Partition 자동 생성기 구현.
3. `bn_app_rw` 계정에 대해 `audit_logs` 및 `ledger_entries`의 `UPDATE/DELETE` 권한 회수.

## 2. Acceptance Criteria (수용 기준)
1. 차변과 대변이 불일치하는 원장 트랜잭션 커밋 시 `ERR_LEDGER_UNBALANCED` 예외 발생 및 롤백 확인.
2. `audit_logs` 레코드 수정/삭제 시도 시 DB 권한 거부 확인.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 원장 불균형 발생 시 전체 결제/정산 트랜잭션 즉시 어보트 및 Critical 경보 발령.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-DB-003;TC-COM-005`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-DB-002/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

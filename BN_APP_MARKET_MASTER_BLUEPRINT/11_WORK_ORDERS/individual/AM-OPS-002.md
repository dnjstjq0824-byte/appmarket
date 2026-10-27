# AM-OPS-002 — PostgreSQL PITR / OpenSearch 무결성 백업·복원 자동화 및 장애 복구(DR) 드릴 검증

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-OPS-002` |
| **작업명 (Title)** | PostgreSQL PITR / OpenSearch 무결성 백업·복원 자동화 및 장애 복구(DR) 드릴 검증 |
| **목표 (Objective)** | `REQ-H-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `DevOps Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-OPS-001;AM-DB-002` |
| **개발 Phase 및 우선순위** | `Phase 8` / `P1` |
| **입력 문서 (Input Documents)** | `14_INFRASTRUCTURE_DEVOPS/BACKUP_RECOVERY.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `scripts/backup/;deploy/backup/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `scripts/backup/pg_backup_verify.sh;scripts/backup/opensearch_snapshot.sh;scripts/backup/dr_restore_drill.sh` |
| **API 의존성 (API Dependencies)** | `NONE` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-OPS-002` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-OPS-002/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. PostgreSQL 암호화 백업 + WAL 아카이빙 및 OpenSearch 스냅샷 자동화 스크립트 구현.
2. 백업본을 별도 격리 컨테이너에 실제로 복원하여 테이블 행 수, 복식부기 대차평형(`SUM=0`), 감사 로그 해시 체인을 자동 대조하는 `dr_restore_drill.sh` 구현.

## 2. Acceptance Criteria (수용 기준)
1. DR 복원 드릴 실행 시 RPO < 5분, RTO < 30분 기준 충족 및 복원 데이터 무결성 100% 일치.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 백업 또는 복원 검증 실패 시 P0 긴급 운영 알람 즉시 발송.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-OPS-002`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-OPS-002/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

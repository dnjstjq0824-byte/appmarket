# AM-DB-001 — PostgreSQL 16 코어 스키마(회원·개발자·카탈로그·배포·보안) 마이그레이션 및 제약조건 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-DB-001` |
| **작업명 (Title)** | PostgreSQL 16 코어 스키마(회원·개발자·카탈로그·배포·보안) 마이그레이션 및 제약조건 구현 |
| **목표 (Objective)** | `REQ-P-001;REQ-P-002;REQ-C-001` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Database Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-ARCH-001` |
| **개발 Phase 및 우선순위** | `Phase 1` / `P0` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/P_POSTGRESQL_AND_DATA_ARCHITECTURE.md;03_DATABASE/ERD.md;03_DATABASE/TABLE_DICTIONARY.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/db/migrations/;backend/db/queries/;03_DATABASE/migrations/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/db/migrations/0001_initial_schema.up.sql;backend/db/migrations/0001_initial_schema.down.sql;backend/db/sqlc.yaml` |
| **API 의존성 (API Dependencies)** | `NONE` |
| **DB 변경 여부 (DB Schema Changes)** | `YES (Core Tables & Constraints)` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-DB-001;TC-DB-002` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-DB-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. `03_DATABASE/migrations/0001_initial_schema.up.sql` 및 `down.sql`을 `golang-migrate`로 적용.
2. 카탈로그(`catalog_apps`), 배포처(`app_distributions`), 서명(`apk_signatures`) 3대 동일성 분리 제약조건 및 FK `ON DELETE RESTRICT` 검증.
3. `sqlc` 타입 안전 Go 코드 생성 설정 구성.

## 2. Acceptance Criteria (수용 기준)
1. Up → Seed → Down → Up 왕복 마이그레이션 무결성 통과.
2. 34개 전체 테이블 PK/FK/CHECK 제약조건 정상 작동.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 마이그레이션 도중 구문/제약 오류 발생 시 단일 트랜잭션(`BEGIN...COMMIT`) 내에서 즉시 자동 롤백.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-DB-001;TC-DB-002`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-DB-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

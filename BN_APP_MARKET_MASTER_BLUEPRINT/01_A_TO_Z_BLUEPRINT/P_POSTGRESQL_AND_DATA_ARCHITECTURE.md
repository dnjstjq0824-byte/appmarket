# P — PostgreSQL & Data Architecture — PostgreSQL 16 전체 ERD, 실제 DDL, 제약조건, 마이그레이션 및 데이터 보존 아키텍처 (PostgreSQL 16 Full ERD, Executable DDL, Constraints, Migrations & Data Governance)

- **문서 ID**: `DOC-AZ-P`
- **알파벳 코드**: `P`
- **관련 요구사항 ID**: `REQ-P-001`, `REQ-P-002`, `REQ-P-003`
- **연관 Work Order ID**: `AM-DB-001`, `AM-DB-002`
- **연관 테스트 케이스 ID**: `TC-DB-001`, `TC-DB-002`, `TC-DB-003`
- **참조 상세 설계 문서**: [03_DATABASE/ERD.md](../03_DATABASE/ERD.md), [03_DATABASE/TABLE_DICTIONARY.md](../03_DATABASE/TABLE_DICTIONARY.md), [03_DATABASE/INDEX_AND_QUERY_DESIGN.md](../03_DATABASE/INDEX_AND_QUERY_DESIGN.md), [03_DATABASE/migrations/](../03_DATABASE/migrations/)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 제8조에서 요구한 바와 같이 개념적 다이어그램에 그치지 않고, 회원·인증·개발자·카탈로그·버전·배포파일·서명·외부출처·설치기록·검색로그·리뷰·신고·심사·보안검사·결제·환불·정산·광고·통계·관리자활동·알림·시스템설정의 **19대 필수 도메인 전체를 포괄하는 실제 실행 가능한 PostgreSQL DDL과 무결성 제약조건**을 확립하기 위해 필요하다.
- **달성 목표**: 32개 핵심 테이블의 `CREATE TABLE` DDL, PK/FK, `UNIQUE`/`CHECK` 제약조건, `ON DELETE` 정책, 복합/부분/GIN 인덱스, 트랜잭션 격리 수준, 낙관적/비관적 동시성 제어, DB 역할 분리 및 데이터 보존 정책을 정의한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **Database Agent**: 마이그레이션 스크립트 작성, 인덱스 튜닝, 파티션 관리, 제약조건 검증 테스트 수행.
- **Backend Agent**: `sqlc` / `pgx/v5`를 통해 타입 안전 쿼리 및 트랜잭션 경계 구현.

## 3. 필수 기능 (Mandatory Features)
1. **제8조 19대 필수 도메인 32개 테이블 완비**:
   - (1) 회원 및 인증: `users`, `user_identities`, `auth_sessions`, `mfa_credentials`, `roles`, `user_roles`
   - (2) 개발자·법인·조직: `developer_orgs`, `developer_verifications`, `developer_org_members`, `developer_api_keys`, `developer_payout_accounts`
   - (3) 앱 카탈로그: `categories`, `tags`, `catalog_apps`, `app_localizations`, `app_media_assets`, `app_tags`, `app_country_availability`
   - (4~7) 배포처·버전·배포파일·서명·외부출처: `app_distributions`, `app_releases`, `upload_sessions`, `release_artifacts`, `apk_signatures`, `external_connectors`, `external_sync_jobs`, `external_store_items`
   - (8~9) 설치·다운로드 기록 및 검색 로그: `install_events` (월별 Range Partition), `search_query_logs` (월별 Range Partition)
   - (10~12) 리뷰·신고·심사·보안검사: `app_reviews_user` (사용자 리뷰), `moderation_reports` (신고), `release_reviews` (관리자 출시 심사), `security_scan_reports` (APK 악성코드/권한 검사)
   - (13~14) 결제·환불·정산: `commission_policies`, `app_price_tiers`, `promotions`, `promo_codes`, `orders`, `payments`, `refunds`, `user_entitlements`, `ledger_accounts`, `ledger_transactions`, `ledger_entries`, `settlements`
   - (15~16) 광고 및 통계: `ad_campaigns`, `ad_creatives`, `ad_events_daily`, `app_metrics_daily`
   - (17~19) 관리자 활동·알림·시스템설정·인프라: `dual_approval_requests`, `audit_logs`, `notifications`, `system_settings`, `outbox_events`, `background_jobs`, `idempotency_keys`.
2. **엄격한 제약조건(Constraints) 및 삭제 정책**:
   - 금전(`orders`, `payments`, `ledger_*`, `settlements`), 감사(`audit_logs`), 릴리스 이력(`app_releases`, `release_artifacts`) 테이블은 `ON DELETE RESTRICT`를 강제하여 부모 레코드 실수 삭제 시 연쇄 증발(Cascade Loss)을 원천 차단한다.
   - 모든 상태값 컬럼은 허용된 문자열 집합만 들어오도록 `CHECK (status IN (...))` 제약조건을 부여한다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **트랜잭션 및 동시성 제어 원칙**:
  - 일반 CRUD: `READ COMMITTED` + 낙관적 잠금(`WHERE id = $1 AND version = $2`, 성공 시 `version = version + 1`).
  - 주문/환불/정산 원장 및 리딤 코드 차감: `SELECT ... FOR UPDATE` 비관적 행 잠금 또는 `SERIALIZABLE` 트랜잭션 + `Idempotency-Key` 결합.
  - 비동기 워커 작업 할당: `SELECT ... FROM background_jobs WHERE status = 'QUEUED' ORDER BY scheduled_at ASC FOR UPDATE SKIP LOCKED LIMIT 1`.
- **DB 계정 최소 권한 분리**:
  - `bn_migrator`: DDL 실행 권한 보유 (배포 시에만 사용).
  - `bn_app_rw`: 일반 테이블 `SELECT, INSERT, UPDATE, DELETE` 보유. 단, `audit_logs`와 `ledger_entries`에 대해서는 `UPDATE, DELETE, TRUNCATE` 권한을 회수(`REVOKE`)하여 애플리케이션이 해킹당하더라도 기존 감사 로그와 원장을 변조할 수 없도록 차단한다.
  - `bn_readonly_analytics`: 리플리카 조회 전용(`SELECT` only).

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- N/A (데이터베이스 아키텍처 문서이나, 모든 화면의 입력 검증 규칙은 본 DB의 `CHECK` 제약조건 및 컬럼 길이와 1:1 일치한다).

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `03_DATABASE/migrations/` 하위의 SQL 마이그레이션 파일(`0001_initial_schema.up.sql`, `0001_initial_schema.down.sql`, `0002_indexes_and_partitions.up.sql`, `0003_seed_reference_data.up.sql`)을 `golang-migrate`로 관리한다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- 전체 테이블 컬럼 사전과 타입 선정 근거는 [03_DATABASE/TABLE_DICTIONARY.md](../03_DATABASE/TABLE_DICTIONARY.md)에, 실행 가능한 전체 SQL 코드는 [03_DATABASE/migrations/0001_initial_schema.up.sql](../03_DATABASE/migrations/0001_initial_schema.up.sql)에 수록되어 있다.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- DB 스키마의 모든 UUID 기본키는 시간순 정렬과 B-Tree 인덱스 지역성(Locality)이 우수한 **UUIDv7 (RFC 9562)** 생성기를 Go 애플리케이션 레이어에서 사용하여 페이지 분할(Page Split) 부하를 최소화한다.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- 마이그레이션 상태: 모든 스키마 변경은 Forward-Compatible(Expand-and-Contract 패턴)로 작성하여 무중단 롤링 배포 중 구버전 API 서버와 신버전 DB 스키마가 동시에 동작할 수 있도록 보장한다.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 개인정보 컬럼(전화번호, 정산 계좌번호, 세무 식별번호)은 애플리케이션 레벨 `AES-256-GCM`(Envelope Encryption with KMS)으로 암호화한 후 `BYTEA`/`TEXT`로 저장하고, 검색이 필요한 이메일은 정규화된 `HMAC-SHA256` 블라인드 인덱스(`email_blind_index`)를 병행 사용한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- N/A (백엔드 데이터 계층).

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **마이그레이션 Up/Down/Up 왕복 테스트 (`TC-DB-002`)**: 빈 DB에 전체 Up 마이그레이션 실행 → 시드 데이터 삽입 → Down 롤백 → 재실행이 에러 0건으로 수행되는지 검증.
- **제약조건 및 참조 무결성 강제 테스트 (`TC-DB-003`)**: 음수 가격, 허용되지 않은 상태값, 고아 외래키, `audit_logs` 삭제 시도가 모두 DB 엔진 레벨에서 즉시 거부되는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- 모든 백엔드 도메인 모듈의 영속성 기반이다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0 (최우선 선행 작업)`
- **개발 Phase**: `Phase 1`의 첫 번째 작업인 `AM-DB-001`에서 구현한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 대용량 로그 테이블(`install_events`, `search_query_logs`) 비대화로 인한 쿼리 성능 저하** → 월별 선언적 파티셔닝(`PARTITION BY RANGE (created_at)`)을 기본 적용하고, 보존 기간(핫 스토리지 6개월, 콜드 S3 Parquet 아카이브 5년) 경과 파티션을 `DETACH PARTITION`으로 락 없이 아카이빙한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- `03_DATABASE/ERD.md`, `03_DATABASE/TABLE_DICTIONARY.md`, `03_DATABASE/INDEX_AND_QUERY_DESIGN.md`, `03_DATABASE/migrations/*.sql`.

## 17. Acceptance Criteria (수용 기준)
1. `03_DATABASE/migrations/*.up.sql`이 문법 오류 없이 파싱·실행되어야 하며 19개 필수 도메인 테이블이 모두 생성되어야 한다.
2. 모든 테이블에 PK, 타임스탬프, 적절한 FK 및 인덱스가 정의되어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-DB-001/sql_schema_validation.log`: SQL DDL 구문 및 제약조건 실행 검증 로그.

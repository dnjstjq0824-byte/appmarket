# BACKUP_RECOVERY.md — 데이터베이스·검색엔진·스토리지 백업 및 재해 복구(DR) 절차서

- **문서 ID**: `DOC-OPS-004`
- **연관 Work Order**: `AM-OPS-002`

---

## 1. 컴포넌트별 RPO / RTO 목표 및 백업 방식

| 데이터 저장소 | 백업 주기 및 방식 | 목표 RPO (최대 데이터 유실) | 목표 RTO (최대 복구 시간) | 보존 기간 |
| :--- | :--- | :---: | :---: | :--- |
| **PostgreSQL 16** | 일 1회 전체 물리 백업(`pg_basebackup`) + 연속 WAL 아카이빙(S3 암호화 전송) | **< 5분** (동기 Replica 페일오버 시 **0초**) | **< 30분** | 일간 백업 35일, 월말 스냅샷 5년 |
| **OpenSearch 2.x** | 일 1회 S3 스냅샷 리포지토리 증분 백업 (유실 시에도 PostgreSQL 원본으로부터 전량 재색인 가능) | **< 24시간** (재색인 시 **0초**) | **< 45분** | 14일 |
| **`bn-public-cdn-bucket`** | S3 Object Versioning + 교차 리전 복제(Cross-Region Replication) | **0초** (불변 객체) | **< 15분** | 영구 보존 (법적 퇴출 앱 제외) |
| **`bn-kyb-vault-bucket`** | KMS 암호화 스냅샷 + 버전 관리 | **0초** | **< 30분** | 법적 보존 기간 준수 |

---

## 2. 월간 정기 복원 훈련 (`dr_restore_drill.sh`) 검증 절차

백업 파일이 실제로 복원 가능한지 확인하기 위해 매월 자동화 스크립트가 다음 절차를 수행하고 증빙 로그를 남깁니다:
1. 최신 `pg_basebackup` + WAL 아카이브를 격리된 임시 PostgreSQL 컨테이너에 PITR 복원합니다.
2. 복원된 DB에서 ① 전체 테이블 레코드 수, ② `SELECT SUM(CASE WHEN direction='DEBIT' THEN amount_minor ELSE -amount_minor END) FROM ledger_entries;` 결과가 `0`인지, ③ `audit_logs` 해시 체인 검증기가 처음부터 끝까지 통과하는지 확인합니다.

# INDEX_AND_QUERY_DESIGN.md — 인덱스 전략, 핵심 쿼리 실행 계획 및 데이터 보존 정책

- **문서 ID**: `DOC-DB-003`
- **관련 문서**: [0002_indexes_and_constraints.up.sql](migrations/0002_indexes_and_constraints.up.sql)

---

## 1. 핵심 고빈도 쿼리 및 인덱스 매핑 설계

### 1.1 Android 기기 일괄 업데이트 및 호환성 조회 쿼리 (`POST /api/v1/device/check-updates`)
- **호출 빈도**: 전체 API 트래픽의 약 35% 차지 (수십만 기기의 백그라운드 업데이트 체크).
- **최적화 인덱스**:
  - `idx_app_distributions_pkg_signer ON app_distributions (package_name, signing_cert_sha256)`
  - `idx_app_releases_dist_track_ver ON app_releases (distribution_id, release_track, version_code DESC)`
  - `idx_release_artifacts_abis_gin ON release_artifacts USING GIN (native_abis)`
- **쿼리 패턴**:
  ```sql
  SELECT d.package_name, d.signing_cert_sha256, r.id AS release_id, r.version_code,
         r.version_name, r.rollout_percentage, a.min_sdk, a.native_abis, a.sha256_checksum
  FROM app_distributions d
  JOIN LATERAL (
      SELECT id, version_code, version_name, rollout_percentage
      FROM app_releases
      WHERE distribution_id = d.id
        AND release_track = 'PRODUCTION'
        AND release_status IN ('PUBLISHED_STAGED', 'PUBLISHED_FULL')
      ORDER BY version_code DESC
      LIMIT 1
  ) r ON TRUE
  JOIN release_artifacts a ON a.release_id = r.id AND a.artifact_status = 'PUBLISHED_ON_CDN'
  WHERE d.package_name = ANY($1::varchar[])
    AND d.is_active = TRUE;
  ```
- **예상 실행 계획**: `Index Scan using idx_app_distributions_pkg_signer` + `Nested Loop Lateral Index Scan using idx_app_releases_dist_track_ver` → 100개 패키지 기준 실행 시간 **< 8ms**.

### 1.2 비동기 워커 작업 큐 및 Outbox 폴링 최적화 (Partial Index + `SKIP LOCKED`)
- 완료된 수백만 건의 `SUCCEEDED`/`PUBLISHED` 행을 스캔하지 않도록 **부분 인덱스(Partial Index)**를 적용합니다:
  - `CREATE INDEX idx_outbox_pending ON outbox_events (created_at ASC) WHERE status = 'PENDING';`
  - `CREATE INDEX idx_jobs_queued ON background_jobs (queue_name, scheduled_at ASC) WHERE status IN ('QUEUED', 'RETRY_SCHEDULED');`
- 워커 폴링 시 `FOR UPDATE SKIP LOCKED`를 사용하여 다수의 워커 인스턴스가 경합(Lock Wait) 없이 병렬로 작업을 가져갑니다.

---

## 2. 데이터 보존 및 파기 정책 (Data Retention & Lifecycle Policy)

| 테이블 그룹 | 핫 스토리지 보존(PostgreSQL) | 콜드 아카이브 보존(S3 Parquet/Glacier) | 파기/익명화 방식 | 법적·운영적 근거 |
| :--- | :---: | :---: | :--- | :--- |
| `orders`, `payments`, `refunds`, `ledger_*`, `settlements` | **5년** | **10년** | 물리적 삭제 금지 (`RESTRICT`) | 전자상거래법 및 세법상 계약·대금결제 기록 5년 보존 의무 |
| `audit_logs`, `dual_approval_requests` | **3년** | **7년** | 해시 체인 검증 후 연 단위 읽기전용 스냅샷 이관 | 정보보호 및 보안 감사 추적성 확보 |
| `install_events`, `search_query_logs` | **6개월** (월별 파티션) | **2년** (익명 집계본) | `DETACH PARTITION` 후 S3 이관 및 구 파티션 `DROP` | 대용량 텔레메트리 테이블 IOPS 유지 |
| `idempotency_keys` | **7일** | 미보존 | 일일 배치로 `DELETE WHERE expires_at < NOW()` | 단기 재시도 중복 방지 목적 달성 후 정리 |
| `upload_sessions` (미완료/만료) | **7일** | 미보존 | S3 Quarantine Multipart Abort 후 행 정리 | 고아 스토리지 비용 방지 |
| 탈퇴 회원 (`users`) | 즉시 익명화 | N/A | `email`을 `deleted_{uuid}@anon.invalid`로 치환, `password_hash=NULL`, `status='DELETED_ANONYMIZED'` | 개인정보보호법/GDPR 파기 의무와 주문/리뷰 외래키 무결성 동시 충족 |

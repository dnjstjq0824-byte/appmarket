# EVENT_FLOW.md — 비동기 도메인 이벤트 계약 및 Transactional Outbox 흐름

- **문서 ID**: `DOC-ARCH-004`
- **관련 문서**: [01_A_TO_Z_BLUEPRINT/J_JOB_PROCESSING.md](../01_A_TO_Z_BLUEPRINT/J_JOB_PROCESSING.md)

---

## 1. Transactional Outbox 패턴 보장 메커니즘

DB 상태 변경과 메시지 큐 발행 사이의 이중 쓰기(Dual-Write) 불일치를 막기 위해 모든 도메인 이벤트는 아래 3단계로 처리합니다:
1. 비즈니스 테이블(`app_releases`, `orders` 등) 변경과 `outbox_events` 테이블 `INSERT`를 **단일 PostgreSQL 로컬 트랜잭션**으로 커밋합니다.
2. `Outbox Relay Worker`가 `SELECT * FROM outbox_events WHERE status = 'PENDING' ORDER BY created_at ASC FOR UPDATE SKIP LOCKED LIMIT 100`으로 폴링하여 Redis Streams 토픽에 발행합니다.
3. 발행 성공 시 `outbox_events.status = 'PUBLISHED'`, `processed_at = NOW()`로 마킹합니다.

---

## 2. 표준 도메인 이벤트 카탈로그 (Event Catalog)

| 이벤트 타입 (`event_type`) | 발행 모듈 | 소비 큐 (`queue_name`) | 페이로드 핵심 필드 | 후속 처리 동작 |
| :--- | :--- | :--- | :--- | :--- |
| `artifact.uploaded.v1` | `distribution` | `q:artifact:verify` | `artifact_id`, `app_id`, `s3_quarantine_key`, `expected_sha256` | ZIP 무결성, AXML 매니페스트, `apksigner` 서명 검증 |
| `artifact.verified.v1` | `distribution` | `q:artifact:malware_scan` | `artifact_id`, `app_id`, `extracted_permissions`, `cert_sha256` | ClamAV, YARA, 권한 위험도 점수 및 사칭 검사 |
| `release.submitted.v1` | `distribution` | `q:moderation:review` | `release_id`, `app_id`, `risk_score`, `track` | `INTERNAL` 트랙은 즉시 승인, `PRODUCTION`은 관리자 심사 큐 배정 |
| `release.approved.v1` | `admin` | `q:artifact:publish_cdn` | `release_id`, `artifact_id`, `sha256_checksum` | 공개 CDN 버킷 복사, OpenSearch 색인 갱신, 개발자 알림 |
| `release.halted.v1` | `distribution` / `admin` | `q:search:indexer` | `release_id`, `app_id`, `reason`, `is_emergency` | 즉시 배포 차단, CDN 퍼지, 긴급 시 기설치자 보안 푸시 발송 |
| `catalog.updated.v1` | `catalog` | `q:search:indexer` | `app_id`, `canonical_slug`, `updated_locales` | OpenSearch `bn_apps_write` 문서 부분/전체 갱신 |
| `connector.synced.v1` | `connector` | `q:search:indexer` | `connector_id`, `sync_job_id`, `changed_app_ids[]` | 외부 스토어 메타데이터 변경분 OpenSearch 벌크 색인 |
| `order.paid.v1` | `commerce` | `q:commerce:settlement` | `order_id`, `user_id`, `app_id`, `amount_minor`, `currency` | 복식부기 원장 검증, 구매 영수증 이메일/인앱 알림 발송 |
| `order.refunded.v1` | `commerce` | `q:commerce:settlement` | `order_id`, `refund_id`, `user_id`, `app_id` | `user_entitlements` 즉시 회수(`REVOKED_REFUNDED`), 역분개 기록 |

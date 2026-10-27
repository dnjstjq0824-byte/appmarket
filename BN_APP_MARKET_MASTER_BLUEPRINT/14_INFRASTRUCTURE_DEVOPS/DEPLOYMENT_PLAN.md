# DEPLOYMENT_PLAN.md — 환경별(Local/Staging/Production) 컨테이너 배포 및 무중단 롤아웃 계획

- **문서 ID**: `DOC-OPS-001`
- **관련 문서**: [01_A_TO_Z_BLUEPRINT/H_HOSTING_AND_INFRASTRUCTURE.md](../01_A_TO_Z_BLUEPRINT/H_HOSTING_AND_INFRASTRUCTURE.md), [02_SYSTEM_ARCHITECTURE/DEPLOYMENT_ARCHITECTURE.md](../02_SYSTEM_ARCHITECTURE/DEPLOYMENT_ARCHITECTURE.md)

---

## 1. 3단계 배포 환경 분리 원칙

| 환경 구분 | 목적 | 스토리지 및 외부 연동 구성 | 배포 승인 조건 |
| :--- | :--- | :--- | :--- |
| **Local / CI (`docker-compose`)** | 에이전트 개발 및 통합/E2E 자동 검증 | PostgreSQL 16, Redis 7, OpenSearch 2, MinIO(3개 버킷 자동 초기화), ClamAV/YARA 로컬 컨테이너 | PR 생성 시 자동 기동 및 테스트 |
| **Staging (`staging.appmarket...`)** | 프로덕션 동일 토폴로지 리허설, 실기기 테스트, DR 드릴 | Managed PostgreSQL(HA), OpenSearch 클러스터, S3, CDN, PG 샌드박스 웹훅 | `main` 브랜치 병합 및 CI 전체 통과 시 자동 카나리 배포 |
| **Production (`appmarket...`)** | 실제 글로벌 상용 서비스 운영 | Multi-AZ PostgreSQL + Read Replica, KMS HSM 볼트, 글로벌 CDN + WAF, 실 PG 연동 | Staging E2E/독립감사 `VERIFIED` + **PO 및 Lead Architect 수동 승인 Gate 필수** |

---

## 2. 무중단 배포(Zero-Downtime Rollout) 및 자동 롤백 절차

1. **DB 마이그레이션 선행 (Expand-and-Contract)**: 파괴적 컬럼 삭제/이름 변경을 단일 배포에서 수행하지 않고, 하위 호환 가능한 컬럼/테이블 추가(`Expand`) 마이그레이션을 먼저 적용합니다.
2. **카나리 트래픽 전환**: 신규 `backend-api` 및 `web-user` 파드에 트래픽을 `10% (5분 관찰) → 50% (5분 관찰) → 100%` 순서로 점진 전환합니다.
3. **자동 롤백 트리거**: 전환 중 Prometheus에서 `http_requests_5xx_ratio > 0.01` 또는 `http_request_duration_seconds_p99 > 2.0`이 2분 연속 감지되면 즉시 카나리를 중단하고 이전 이미지 다이제스트로 10초 내 트래픽을 복귀시킵니다.

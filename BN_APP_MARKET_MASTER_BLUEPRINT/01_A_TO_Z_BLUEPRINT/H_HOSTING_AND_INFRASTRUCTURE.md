# H — Hosting & Infrastructure — 컨테이너 인프라, 이중 스토리지, CDN, WAF, 백업 및 무중단 배포 아키텍처 (Container Infrastructure, Dual Object Storage, CDN, WAF, Backup & High-Availability Operations)

- **문서 ID**: `DOC-AZ-H`
- **알파벳 코드**: `H`
- **관련 요구사항 ID**: `REQ-H-001`, `REQ-H-002`
- **연관 Work Order ID**: `AM-OPS-001`, `AM-OPS-002`, `AM-DIST-002`
- **연관 테스트 케이스 ID**: `TC-OPS-002`, `TC-OPS-003`
- **참조 상세 설계 문서**: [02_SYSTEM_ARCHITECTURE/DEPLOYMENT_ARCHITECTURE.md](../02_SYSTEM_ARCHITECTURE/DEPLOYMENT_ARCHITECTURE.md), [07_DISTRIBUTION/STORAGE_CDN.md](../07_DISTRIBUTION/STORAGE_CDN.md), [14_INFRASTRUCTURE_DEVOPS/DEPLOYMENT_PLAN.md](../14_INFRASTRUCTURE_DEVOPS/DEPLOYMENT_PLAN.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 수백 MB~수 GB에 달하는 APK/AAB 바이너리 업로드와 글로벌 동시 다발적 다운로드를 안정적으로 처리하고, 데이터베이스 유실 없이 무중단 배포와 장애 복구(DR)를 수행할 수 있는 상용 인프라 설계가 필요하다.
- **달성 목표**: Docker 기반 멀티스테이지 컨테이너 구성, S3 호환 오브젝트 스토리지(Quarantine/Public/KYB 3중 분리), CDN Signed URL 배포, WAF 방어 규칙, PostgreSQL PITR(특정 시점 복구) 및 OpenTelemetry 관제 체계를 수립한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **DevOps Agent / SRE**: 컨테이너 이미지 빌드, CI/CD 파이프라인 운영, 리소스 오토스케일링, 백업 복원 훈련, 모니터링 알람 대응.
- **Security Analyst**: WAF 차단 룰셋 튜닝, 격리 스토리지 접근 제어 감사.

## 3. 필수 기능 (Mandatory Features)
1. **Docker 기반 표준 컨테이너화**: 모든 서비스(`backend-api`, `worker-*`, `web-user`, `web-dev`, `web-admin`)는 재현 가능한 멀티스테이지 `Dockerfile`과 로컬/스테이징 통합 실행용 `docker-compose.yml` 및 프로덕션 K8s/컨테이너 매니페스트를 갖는다.
2. **3중 분리 오브젝트 스토리지 버킷**:
   - `bn-quarantine-bucket`: 업로드 직후 미검증 바이너리 저장 (외부 퍼블릭 액세스 완전 차단, 7일 후 미승인 파일 자동 만료 Lifecycle Rule).
   - `bn-public-cdn-bucket`: 심사 및 서명 검증을 통과한 출시 아티팩트 저장 (`Object Lock / Versioning` 활성화로 기존 버전 덮어쓰기 원천 차단).
   - `bn-kyb-vault-bucket`: 개발자 신원 인증 서류 및 정산 증빙 암호화 저장.
3. **글로벌 CDN 및 변조 방지 배포**: 모든 APK 다운로드 경로는 `/packages/{app_id}/{release_id}/{sha256}.apk` 형태의 콘텐츠 주소 지정(Content-Addressable) 불변 경로를 사용하며, HMAC 서명과 만료 시각이 포함된 Signed URL로만 접근을 허용한다.
4. **자동 백업 및 복원 검증**: PostgreSQL 일일 전체 백업 + WAL 아카이빙(RPO < 5분, RTO < 30분), OpenSearch 일일 스냅샷.

## 4. 상세 기능 (Detailed Functional Specifications)
- **WAF 및 Edge Rate Limiting 정책**:
  - 로그인/회원가입/인증번호 요청: IP당 10 req/min.
  - 검색 및 카탈로그 조회: IP당 120 req/min.
  - APK 다운로드 URL 발급: 사용자/기기당 30 req/min (대역폭 고갈 공격 방어).
  - SQLi, XSS, Log4Shell, 비정상 User-Agent 봇 트래픽 Edge 차단.
- **관제 및 알림 (OpenTelemetry + Prometheus + Grafana + Alertmanager)**:
  - Golden Signals(Latency p50/p95/p99, Traffic, Error Rate 5xx, Saturation) 대시보드 및 큐 적체(`queue_depth > 1000`), 디스크 사용량(`> 80%`), 인증서 만료 임박 알림 구성.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- Nuxt SSR 서버(`web-user`)와 정적 콘솔(`web-dev`, `web-admin`)의 정적 에셋(`/_nuxt/*`, `/assets/*`)은 파일명 해시 기반 `Cache-Control: public, max-age=31536000, immutable`로 CDN에서 서빙한다.

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- Go 서버는 `SIGTERM` 수신 시 로드밸런서 헬스체크(`readyz`)를 먼저 실패 상태로 전환하고, 진행 중인 HTTP 요청 및 파일 업로드 스트림을 최대 30초간 안전하게 완료한 뒤 종료(Graceful Drain)한다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- PostgreSQL 16 Primary + Read Replica 구성. `pgBouncer` 커넥션 풀러를 통해 수백 개의 API/워커 인스턴스가 DB 연결을 고갈시키지 않도록 보호한다.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `GET /metrics`: Prometheus 스크래핑용 내부 전용 포트(`:9090/metrics`) 노출 (외부 Ingress에서는 접근 차단).

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **배포 롤아웃 및 자동 롤백 상태**:
  - 카나리/롤링 배포 중 5분간 HTTP 5xx 에러율이 1.0%를 초과하거나 `p99` 지연시간이 2,000ms를 초과하면 CI/CD 파이프라인이 자동으로 이전 안정 이미지 태그로 롤백한다.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 컨테이너 이미지 빌드 시 `Trivy` 및 `govulncheck`, `pnpm audit`를 실행하여 Critical/High CVE가 발견되면 빌드를 실패 처리한다.
- 모든 시크릿은 환경변수 또는 Secret Manager/Vault를 통해 런타임 주입하며 이미지 레이어에 절대 포함하지 않는다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- N/A (백엔드 및 인프라 계층 명세이나, 관리자 콘솔의 `SystemHealthView.vue`에 인프라 상태 지표를 연동하여 시각화한다).

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **백업 및 복원 드릴 테스트 (`TC-OPS-002`)**: 테스트 DB에 샘플 카탈로그/원장 데이터를 넣은 후 `pg_dump` / PITR 백업본으로부터 신규 DB 인스턴스로 복원하여 레코드 수와 원장 해시가 100% 일치하는지 검증.
- **불변 아티팩트 덮어쓰기 차단 테스트 (`TC-OPS-003`)**: 이미 배포된 `{sha256}.apk` 경로에 다른 바이트의 파일 업로드를 시도했을 때 스토리지 정책에 의해 거부되는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `A — Architecture`, `N — Native App Distribution`, `P — PostgreSQL`, `Z — Zero-Trust Verification`과 연동된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0 (기반 인프라)` / `P1 (상용 고가용성 및 DR)`
- **개발 Phase**: 로컬/CI 컨테이너 환경은 `Phase 0`, CDN 배포망은 `Phase 4`, 고가용성/백업 복원 검증은 `Phase 8`에서 수행한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 주 CDN 사업자 장애** → 다운로드 URL 발급기가 상태 체크에 따라 Secondary Origin/Mirror 도메인으로 Signed URL 호스트를 즉시 전환(Multi-CDN Failover)할 수 있도록 URL 생성 로직을 설정화한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- `deploy/docker-compose.yml`, `deploy/Dockerfile.*`, `deploy/otel-collector-config.yaml`, `deploy/prometheus/alerts.yml`, `scripts/backup_restore_verify.sh`.

## 17. Acceptance Criteria (수용 기준)
1. 모든 컨테이너가 헬스체크를 통과하고 OpenTelemetry 트레이스가 수집되어야 한다.
2. 백업 및 복원 스크립트 실행 시 데이터 유실이 0건이어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-OPS-001/container_vuln_scan.json`: 컨테이너 보안 스캔 리포트.
- `artifacts/evidence/AM-OPS-002/backup_restore_drill.log`: DB 백업 및 복원 무결성 대조 로그.

# Work Order Progress Report: AM-OPS-001

- **담당**: DevOps
- **상태**: `PARTIAL`
- **Git 브랜치**: `feat/AM-ARCH-001` (Wave 0 초기 착수 브랜치)
- **Git Commit SHA**: 없음. 신규 로컬 저장소이며 아직 커밋하지 않았다.

## 변경 파일

현재 파일은 모두 신규 미추적 파일이므로 `git diff --stat`에는 나타나지 않는다. 작성 파일은 `deploy/Dockerfile.backend`, `deploy/docker-compose.yml`, `deploy/otel-collector-config.yaml`, `deploy/prometheus/prometheus.yml`, `deploy/prometheus/alerts.yml`, `.github/workflows/ci-verify.yml`, `.env.example`, `.gitignore`이다.

## 실행 결과

- `docker` 명령을 찾지 못해 `docker compose config`, 컨테이너 기동, 비루트 UID 및 격리 네트워크 검증을 실행하지 못했다.
- `TC-OPS-001` 스캐너 워커 강제 장애 테스트: **미실행**. 스캐너 워커 기능도 아직 구현되지 않았다.
- `TC-OPS-004` Zombie Job Reaper 테스트: **미실행**. 선행 `AM-ARCH-002`가 아직 구현되지 않았다.
- CI 워크플로는 작성했으나 원격 CI 실행 결과가 없으므로 PASS로 기록하지 않는다.

## 미완료 및 제한

- Compose에는 PostgreSQL 16, Redis 7.4, OpenSearch 2.17, MinIO 3개 비공개 버킷, Go API, OTel Collector, Prometheus, Grafana 정의가 있다. 실제 이미지 풀 및 기동은 확인되지 않았다.
- 스캐너 워커 컨테이너 및 크래시 격리, Tempo/Loki 연동, 취약점·시크릿·계약 검사 CI 게이트, 전체 서비스 비루트 UID 검증이 남아 있다.
- 독립 감사가 없으므로 `VERIFIED`로 승격하지 않는다.

## 2026-10-28 후속 작업

- CI에 Go 정적 보안 분석(gosec), OpenAPI 번들 검증, 시크릿 스캔(gitleaks), 취약점·설정 스캔(Trivy) 게이트를 추가했다.
- 현재 환경에서 `docker version` 실행은 명령을 찾지 못해 실패했다. 컨테이너 기동과 `TC-OPS-001`/`TC-OPS-004`는 계속 미실행이다. 원격 CI 실행 결과도 아직 없다.
- 상태는 `PARTIAL`이며 CI의 성공 여부를 증빙 없이 PASS로 표시하지 않는다.

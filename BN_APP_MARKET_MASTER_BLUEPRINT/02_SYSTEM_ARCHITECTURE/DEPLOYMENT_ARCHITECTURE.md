# DEPLOYMENT_ARCHITECTURE.md — 컨테이너 배포 토폴로지 및 네트워크 격리 아키텍처

- **문서 ID**: `DOC-ARCH-003`
- **관련 문서**: [01_A_TO_Z_BLUEPRINT/H_HOSTING_AND_INFRASTRUCTURE.md](../01_A_TO_Z_BLUEPRINT/H_HOSTING_AND_INFRASTRUCTURE.md), [14_INFRASTRUCTURE_DEVOPS/DEPLOYMENT_PLAN.md](../14_INFRASTRUCTURE_DEVOPS/DEPLOYMENT_PLAN.md)

---

## 1. 네트워크 서브넷 및 방화벽 격리 구조 (4-Zone Network Segmentation)

| 네트워크 존 | 배치 워크로드 | 인바운드 허용(Ingress) | 아웃바운드 허용(Egress) |
| :--- | :--- | :--- | :--- |
| **Zone 1: Public DMZ** | L7 Ingress Controller, CDN Edge Proxy | `443/tcp` (전 세계 HTTPS) | Zone 2 (`web-*`, `backend-api`)로만 허용 |
| **Zone 2: App Private Zone** | `web-user`, `web-dev`, `web-admin`, `backend-api`, `worker-indexer`, `worker-settlement` | Zone 1로부터의 HTTP/gRPC만 허용 | Zone 3 Data Zone 및 허용된 외부 PG/OIDC/SES HTTPS만 허용 |
| **Zone 3: Quarantine Sandbox Zone** | `worker-apk-analyzer`, `worker-malware-scanner`, `worker-bundletool` | **인바운드 포트 0개** (큐 Pull 전용) | 오직 내부 Redis, PostgreSQL, S3 Quarantine 엔드포인트로만 통신 허용 (**외부 인터넷 Egress 100% 차단**) |
| **Zone 4: Data Persistence Zone** | PostgreSQL 16(Primary/Replica), OpenSearch, Redis, S3 Storage | Zone 2 및 Zone 3의 지정된 서비스 시큐리티 그룹에서만 허용 | 외부 인터넷 통신 완전 차단 (백업 S3 VPC Endpoint만 허용) |

---

## 2. 컨테이너 리소스 규격 및 오토스케일링 기준

| 서비스명 | 기본 레플리카 | CPU Req / Limit | Mem Req / Limit | HPA 스케일아웃 조건 |
| :--- | :---: | :---: | :---: | :--- |
| `backend-api` | 3 (Min 3, Max 20) | `500m / 2000m` | `512Mi / 2Gi` | CPU > 65% 또는 RPS > 500/pod |
| `web-user` | 3 (Min 3, Max 15) | `500m / 1500m` | `512Mi / 1.5Gi` | CPU > 65% |
| `web-dev` / `web-admin` | 각 2 (Nginx Static) | `100m / 500m` | `128Mi / 256Mi` | CPU > 75% |
| `worker-apk-analyzer` | 2 (Min 2, Max 10) | `1000m / 2000m` | `1Gi / 4Gi` | `q:artifact:verify` lag > 10 |
| `worker-malware-scanner` | 2 (Min 2, Max 10) | `1000m / 4000m` | `2Gi / 6Gi` | `q:artifact:malware_scan` lag > 5 |
| `worker-connector` | 1 (Singleton/Leader) | `500m / 1000m` | `512Mi / 2Gi` | 스케줄 기반 실행 |

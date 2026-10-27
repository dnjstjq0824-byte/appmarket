# SYSTEM_OVERVIEW.md — 글로벌 독립 통합 앱 마켓 시스템 아키텍처 개요

- **문서 ID**: `DOC-ARCH-001`
- **관련 문서**: [01_A_TO_Z_BLUEPRINT/A_ARCHITECTURE.md](../01_A_TO_Z_BLUEPRINT/A_ARCHITECTURE.md), [00_PROJECT_CHARTER/PROJECT_OVERVIEW.md](../00_PROJECT_CHARTER/PROJECT_OVERVIEW.md)
- **연관 Work Order**: `AM-ARCH-001`, `AM-OPS-001`

---

## 1. 아키텍처 핵심 비전 및 3대 플랫폼 유형 통합 원리

본 시스템은 중복된 3개의 플랫폼을 따로 구축하지 않고, **단일 통합 코어 플랫폼** 위에서 다음 3가지 마켓 유형을 유기적으로 제공합니다:

1. **Type 1 — Independent App Store (`BN_HOSTED`)**: 개발자 직접 등록, APK/AAB 보안 검증, 운영자 심사, 자체 S3/CDN 배포 및 Android `PackageInstaller` 직접 설치·업데이트.
2. **Type 2 — Federated App Search Engine (`FDROID_REPO`, `PARTNER_FEED`, `OFFICIAL_SITE_LINK`)**: 외부 허용 리포지토리 및 공식 배포처 메타데이터 수집·정규화·출처 표시 및 공식 배포처 연결.
3. **Type 3 — Universal App Marketplace**: 단일 논리적 앱(`catalog_apps`) 아래에 복수의 배포처(`app_distributions`)를 매핑하여 가격·버전·요구 SDK·서명 인증서 지문을 비교 탐색.

---

## 2. 기술 스택 확정 매트릭스 (제2조 준수)

| 계층 | 채택 기술 | 버전 | 아키텍처 역할 및 격리 방식 |
| :--- | :--- | :--- | :--- |
| **User Web (`web-user`)** | Vue 3 + Nuxt 3 + TypeScript | Vue 3.5 / Nuxt 3.13 | SSR + ISR 하이브리드 렌더링, SEO(JSON-LD, Sitemap), 다국어/다통화 UI |
| **Developer Console (`web-dev`)** | Vue 3 + Vite + TypeScript | Vue 3.5 / TS 5.6 | 개발자 온보딩(KYB), 멀티파트 바이너리 업로드, 트랙별 출시 및 정산 관리 |
| **Admin Console (`web-admin`)** | Vue 3 + Vite + TypeScript | Vue 3.5 / TS 5.6 | 20개 운영 도메인 관제, 앱/보안 심사, 이중 승인(Dual-Control), 감사 로그 |
| **Android Client (`android-client`)** | Kotlin + Jetpack Compose | Kotlin 2.0 / API 26~35 | 기기 호환성 판별, 이어받기 다운로더, SHA-256/서명 검증, `PackageInstaller` |
| **Core Backend (`backend-api`)** | Go (Modular Monolith) | Go 1.23+ | 10개 도메인 모듈, REST/OpenAPI 3.1, OIDC/JWT/RBAC, Transactional Outbox |
| **Isolated Workers (`workers`)** | Go + Sandboxed CLI Tools | Go 1.23 / ClamAV / YARA / bundletool | API 서버와 물리적 분리된 비동기 워커(APK 파서, 서명 검증, 악성코드 스캔, 검색 색인, 외부 커넥터) |
| **Main Database** | PostgreSQL | 16+ | 32개 핵심 테이블, 엄격한 FK/CHECK 제약조건, 복식부기 원장, 월별 파티셔닝 |
| **Search Engine** | OpenSearch | 2.17+ | 한국어(Nori)/ICU 다국어 분석기, BM25 + Quality Function Score, 자동완성 |
| **Cache & Queue** | Redis | 7.4+ | 분산 락, Rate Limiting, 캐시, Redis Streams 기반 비동기 작업 큐 |
| **Object Storage & CDN** | S3-Compatible + Global CDN | AWS S3 / MinIO + CloudFront/Cloudflare | `bn-quarantine-bucket`(격리), `bn-public-cdn-bucket`(불변 공개), `bn-kyb-vault-bucket`(암호화) |
| **Observability** | OpenTelemetry + Prometheus + Grafana | OTel 1.30+ | 분산 트레이싱(Tempo), 메트릭(Prometheus), 중앙 로그(Loki), 알림(Alertmanager) |

---

## 3. Go 모듈형 모놀리스 내부 도메인 경계

```text
backend/
├── cmd/
│   ├── api/main.go               # Stateless HTTP API Server
│   └── worker/main.go            # Isolated Queue Worker Process
├── internal/
│   ├── platform/                 # DB(pgx), Redis, S3, OTel, Auth Middleware, Error RFC9457
│   └── modules/
│       ├── iam/                  # 회원, OIDC, JWT, MFA, RBAC, Dual-Control, Audit Chain
│       ├── developer/            # 개발자 조직, KYB 검증, 팀원, Developer API Key
│       ├── catalog/              # 통합 카탈로그, 다국어, 미디어, 카테고리, 국가별 노출 정책
│       ├── distribution/         # 업로드 세션, APK/AAB 파서, 서명 검증, 릴리스 상태 머신, CDN URL
│       ├── security/             # ClamAV/YARA 스캐너, 권한 위험도 분석, 사칭 탐지, Kill Switch
│       ├── search/               # OpenSearch 인덱서, 다국어 쿼리 빌더, 자동완성, 차트/추천
│       ├── connector/            # F-Droid Index v2, 파트너 피드, 출처/권리 등급 관리
│       ├── commerce/             # 가격/프로모션, 주문, PG 웹훅, 구매 권한, 복식부기 원장, 정산
│       ├── ads/                  # 스폰서 캠페인, 경매, 노출 토큰, 부정클릭 필터링
│       ├── moderation/           # 사용자 리뷰, 신고(Abuse/DMCA), 이의신청(Appeals)
│       ├── admin/                # 20개 관리자 운영 도메인 통합 파사드 및 시스템 설정
│       └── treescan/             # 향후 TreeScan 통합 전용 검색/카탈로그/SSO 어댑터 (기본 비활성)
```

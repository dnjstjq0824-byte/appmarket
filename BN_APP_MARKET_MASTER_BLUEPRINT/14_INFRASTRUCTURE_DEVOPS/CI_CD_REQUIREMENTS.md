# CI_CD_REQUIREMENTS.md — Git 기반 CI/CD 자동화 파이프라인 및 보안 품질 게이트 명세

- **문서 ID**: `DOC-OPS-002`
- **연관 Work Order**: `AM-OPS-001`

---

## 1. PR 및 Main 브랜치 필수 통과 8단계 CI 파이프라인

```
[Stage 1: Ownership & Secret Check]  (gitleaks + Work Order writable_dirs 검증)
         │
         ▼
[Stage 2: Contract & Schema Lint]    (openapi.yaml Spectral 린트 + SQL 마이그레이션 문법 검증)
         │
         ▼
[Stage 3: Static Analysis & SAST]    (golangci-lint, gosec, vue-tsc, ktlint/detekt)
         │
         ▼
[Stage 4: Unit & Domain Tests]       (go test -race, vitest, ./gradlew testDebugUnitTest)
         │
         ▼
[Stage 5: Testcontainers Integration] -- (PostgreSQL + Redis + OpenSearch + MinIO 실컨테이너 통합 테스트)
         │
         ▼
[Stage 6: E2E & Accessibility]       (Playwright 3개 웹 앱 E2E + axe-core 접근성 + Android Emulator Test)
         │
         ▼
[Stage 7: Container Build & Scan]    (Multi-stage Docker Build + Trivy Critical/High CVE = 0 검증)
         │
         ▼
[Stage 8: Artifact Signing & Report] (Cosign 이미지 서명 + SHA-256 체크섬 및 테스트 증빙 아카이빙)
```

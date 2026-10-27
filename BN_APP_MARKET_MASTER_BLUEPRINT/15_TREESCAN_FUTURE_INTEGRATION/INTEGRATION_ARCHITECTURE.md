# INTEGRATION_ARCHITECTURE.md — 미래 TreeScan 검색엔진 통합 아키텍처 (Phase 10)

- **문서 ID**: `DOC-TS-001`
- **관련 조항**: 제1조 1.1항, 제3조 Z항, 제17조 (Phase 10)
- **연관 Work Order**: `AM-TREESCAN-001`

---

## 1. 독립 우선 및 느슨한 결합 원칙 (Standalone First, Loosely-Coupled Federation)

1. **현재 단계(Phase 0~9) 완전 분리**: 본 앱 마켓 플랫폼은 TreeScan의 어떠한 라이브러리, DB, 인증 서버, 프론트엔드 번들에도 의존하지 않고 100% 독립 서비스로 먼저 구축·검증·상용화됩니다.
2. **Phase 10 통합 시에도 TreeScan 본체 무수정 준비**: 앱 마켓 내부에 격리된 어댑터 패키지(`backend/internal/modules/treescan`)와 환경변수 플래그(`TREESCAN_INTEGRATION_ENABLED=false`)만 사전 구현합니다.
3. **향후 PO 승인 후 연동 토폴로지**:
   - TreeScan 사용자가 웹 통합 검색을 수행할 때, TreeScan 검색 오케스트레이터가 앱 마켓의 `GET /api/v1/treescan/federated-search`를 호출하여 검증된 앱 리치 카드(Rich App Snippet)를 실시간 결합하거나,
   - `GET /api/v1/treescan/catalog-feed` 증분 피드를 통해 앱 메타데이터를 색인하고, 클릭 시 `POST /api/v1/auth/token/exchange` SSO를 통해 앱 마켓 상세/설치 화면으로 끊김 없이 연결합니다.

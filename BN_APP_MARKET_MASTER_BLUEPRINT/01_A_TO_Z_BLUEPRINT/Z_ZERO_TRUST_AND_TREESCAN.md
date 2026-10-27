# Z — Zero-Trust Verification & TreeScan Integration Readiness — 제로 트러스트 검증, 증거 기반 완료 판정 및 미래 TreeScan 통합 준비 규격 (Zero-Trust Security Posture, Evidence-Based Completion Gate & Future TreeScan Integration Specification)

- **문서 ID**: `DOC-AZ-Z`
- **알파벳 코드**: `Z`
- **관련 요구사항 ID**: `REQ-Z-001`, `REQ-Z-002`, `REQ-Z-003`
- **연관 Work Order ID**: `AM-TREESCAN-001`, `AM-TEST-002`
- **연관 테스트 케이스 ID**: `TC-API-003`, `TC-OPS-008`
- **참조 상세 설계 문서**: [15_TREESCAN_FUTURE_INTEGRATION/INTEGRATION_ARCHITECTURE.md](../15_TREESCAN_FUTURE_INTEGRATION/INTEGRATION_ARCHITECTURE.md), [15_TREESCAN_FUTURE_INTEGRATION/SEARCH_API_ADAPTER.md](../15_TREESCAN_FUTURE_INTEGRATION/SEARCH_API_ADAPTER.md), [15_TREESCAN_FUTURE_INTEGRATION/SSO_STRATEGY.md](../15_TREESCAN_FUTURE_INTEGRATION/SSO_STRATEGY.md), [15_TREESCAN_FUTURE_INTEGRATION/INDEPENDENCE_REQUIREMENTS.md](../15_TREESCAN_FUTURE_INTEGRATION/INDEPENDENCE_REQUIREMENTS.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 제1조 1.1항, 제17조(Phase 10), 제23조에 따라 ① 본 앱 마켓이 TreeScan과 완전히 독립된 상태로 상용 출시 준비를 마쳤음을 **제로 트러스트(Zero-Trust) 증거 기반 감사**로 최종 증명하고, ② 향후 PO 승인 시 본체 재설계 없이 즉시 Booking Network의 **TreeScan 검색엔진** 내부에 앱 검색·계정·설치 경험을 통합할 수 있는 표준 어댑터 규격을 마련하기 위해 필요하다.
- **달성 목표**: 서비스 간 최소 권한(mTLS/Service Token) 제로 트러스트 원칙, 증거 기반 최종 인수 테스트(UAT) 게이트, 그리고 TreeScan 통합용 3대 어댑터(Federated Search API, Catalog Delta Export Stream, RFC 8693 OAuth2 Token Exchange SSO)를 명세한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **Independent Audit Agent & Lead Architect**: 전체 Phase 0~9 산출물에 대한 Zero-Trust 증거 감사 수행 및 독립 실행 검증.
- **TreeScan Integration Engineer (Phase 10)**: 본 문서와 `15_TREESCAN_FUTURE_INTEGRATION/` 규격에 따라 TreeScan과 앱 마켓 어댑터를 연결.

## 3. 필수 기능 (Mandatory Features)
1. **완전한 독립성 보장 (Standalone First, Adapter Ready)**:
   - 환경변수 `TREESCAN_INTEGRATION_ENABLED=false`(기본값)일 때 플랫폼은 외부 TreeScan 엔드포인트에 단 1건의 네트워크 요청도 보내지 않으며 100% 자체 기능으로 동작한다.
   - 현재 단계(Phase 0~10)에서는 TreeScan 본체 저장소를 절대 수정하지 않으며, 오직 앱 마켓 측의 **수신/제공 어댑터 모듈(`internal/modules/treescan`)**만 구현하고 계약 테스트로 검증한다.
2. **미래 TreeScan 통합 3대 표준 인터페이스**:
   - **인터페이스 1 — TreeScan Federated App Search API (`GET /api/v1/treescan/federated-search`)**: TreeScan 통합 검색창에서 사용자가 앱 관련 쿼리(예: `"카카오톡 다운로드"`, `"오프라인 지도 앱"`)를 입력했을 때, TreeScan 백엔드가 호출하여 80ms 이내에 상위 매칭 앱 카드(아이콘, 평점, 가격, 배포처 배지, 딥링크 URL, 직접 설치 가능 여부)를 받아갈 수 있는 초저지연 전용 엔드포인트.
   - **인터페이스 2 — TreeScan Catalog Index Sync Feed (`GET /api/v1/treescan/catalog-feed`)**: TreeScan 자체 웹 인덱서가 앱 카탈로그 메타데이터를 증분 색인할 수 있도록 커서 기반 변경 스트림(`updated_after`, `cursor`)을 Schema.org 호환 JSON 포맷으로 제공.
   - **인터페이스 3 — OIDC / RFC 8693 Token Exchange SSO (`POST /api/v1/auth/token/exchange`)**: 향후 TreeScan 로그인 사용자가 앱 마켓으로 넘어올 때 별도 재로그인 없이 TreeScan ID Token을 앱 마켓 세션 토큰으로 안전하게 교환(또는 계정 연결)하는 표준 인터페이스.
3. **제로 트러스트 증거 기반 완료 판정 (Evidence-Based Zero-Trust Verification)**:
   - 모든 요구사항(`REQ-A-001` ~ `REQ-Z-003`)은 `Design Section → API/DB Component → Work Order → Test Case → Verification Evidence File`의 6단계 추적성 사슬([16_FINAL_DELIVERY/REQUIREMENTS_TRACEABILITY.csv](../16_FINAL_DELIVERY/REQUIREMENTS_TRACEABILITY.csv))이 완비되어야만 완료로 인정된다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **TreeScan 어댑터 인증 및 속도 제한**:
  - `/api/v1/treescan/*` 엔드포인트는 일반 사용자 토큰이 아닌 mTLS 클라이언트 인증서 또는 전용 서비스 간 HMAC/JWT(`X-TreeScan-Service-Token`)로만 접근을 허용하며, 전용 Redis Rate Limit 버킷(기본 2,000 req/sec)을 할당하여 일반 사용자 트래픽과 격리한다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- 사용자 웹(`web-user`)의 앱 상세 페이지는 `?utm_source=treescan&ts_ref=...` 파라미터로 유입될 경우 상단에 `'TreeScan 검색 결과로 돌아가기'` 컨텍스트 브레드크럼(선택적 활성화)을 렌더링할 수 있도록 준비한다.

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/treescan/adapter_handler.go`: Feature Flag(`TREESCAN_INTEGRATION_ENABLED`) 제어, 서비스 토큰 검증, 초저지연 OpenSearch 프로젝션 응답 생성.

## 7. DB 및 데이터 모델 (Database & Data Model)
- `user_identities` 테이블의 `provider = 'TREESCAN_OIDC'` 매핑 슬롯 및 `system_settings`의 `treescan.integration.config` JSONB 설정을 기본 스키마에 포함하여 향후 무중단 활성화를 보장한다.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `GET /api/v1/treescan/federated-search` (`q`, `locale`, `country`, `device_platform`, `limit`).
- `GET /api/v1/treescan/catalog-feed` (`cursor`, `updated_after`, `limit`).
- `POST /api/v1/auth/token/exchange` (RFC 8693 OAuth 2.0 Token Exchange).

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **TreeScan 연동 모드 상태**: `STANDALONE_ISOLATED` (기본값, Phase 0~9) → `SHADOW_TEST_MODE` (Phase 10 어댑터 계약 검증) → `FEDERATED_ACTIVE` (향후 PO 별도 승인 후 활성화).

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 제로 트러스트 원칙에 따라 향후 TreeScan과 연동되더라도 TreeScan 서비스 토큰으로는 **카탈로그 검색 및 공개 메타데이터 조회만 허용**하며, 관리자 심사 승인·APK 변조·정산 원장 접근 권한은 절대 부여하지 않는다(Blast Radius 격리).

## 11. UX/UI 요구사항 (UX/UI Requirements)
- TreeScan 검색 결과 카드에 삽입될 앱 리치 스니펫(App Rich Card)의 JSON 응답 규격에는 반드시 데이터 출처(`provenance_label`), 직접 설치 가능 여부(`can_direct_install`), 공식 딥링크(`canonical_market_url`)가 포함된다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **독립 기동 및 무의존성 검증 테스트 (`TC-OPS-008`)**: TreeScan 관련 모든 네트워크/환경변수가 차단된 상태에서 전체 기능 E2E 스위트가 100% 통과하는지 검증.
- **TreeScan 어댑터 계약 테스트 (`TC-API-003`)**: 샌드박스에서 `TREESCAN_INTEGRATION_ENABLED=true`로 설정 후 `/api/v1/treescan/federated-search` 및 `/catalog-feed`, `/token/exchange` 호출 시 명세된 스키마와 응답시간(`p95 < 80ms`)을 충족하는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `A — Architecture`, `I — Identity & Access`, `O — OpenSearch`, `C — Catalog Engine`과 연결된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: 독립성 원칙은 `P0 (전 기간 강제)`, TreeScan 연동 어댑터 구현은 `P2 (Phase 10)`
- **개발 Phase**: `Phase 10 (TreeScan 연동 준비 및 별도 통합 계획)`에서 `AM-TREESCAN-001`로 구현한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 향후 연동 후 TreeScan 측 트래픽 급증 또는 장애 발생** → 전용 Rate Limiter와 Bulkhead(고루틴/커넥션 풀 분리)를 통해 TreeScan 어댑터 부하가 앱 마켓 자체 사용자 웹(`web-user`)이나 Android 앱 설치 서비스에 전이되지 않도록 완벽히 격리한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- TreeScan 연동 어댑터 모듈 (`backend/internal/modules/treescan/*`) 및 통합 설계 문서 4종 (`15_TREESCAN_FUTURE_INTEGRATION/*`).

## 17. Acceptance Criteria (수용 기준)
1. Phase 0~9 동안 TreeScan 본체 코드 수정이 0건이어야 하며, 앱 마켓이 단독으로 100% 동작해야 한다.
2. Phase 10 어댑터 계약 테스트가 통과하고, Feature Flag 하나로 즉시 활성화/비활성화가 가능해야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-TREESCAN-001/standalone_independence_audit.log`: 외부 TreeScan 의존성 0건 검증 로그.
- `artifacts/evidence/AM-TREESCAN-001/treescan_adapter_contract_test.log`: 어댑터 API 3종 계약 및 응답시간 검증 로그.

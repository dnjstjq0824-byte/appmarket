# T — Third-Party Integration — 외부 PG, OIDC 인증, 이메일/푸시 알림, 외부 스토어 및 TreeScan 연동 어댑터 아키텍처 (Third-Party Integrations: Payment Gateways, OIDC, Messaging, External Stores & TreeScan Adapter)

- **문서 ID**: `DOC-AZ-T`
- **알파벳 코드**: `T`
- **관련 요구사항 ID**: `REQ-T-001`, `REQ-T-002`
- **연관 Work Order ID**: `AM-BILLING-001`, `AM-OPS-002`, `AM-TREESCAN-001`
- **연관 테스트 케이스 ID**: `TC-API-002`, `TC-OPS-006`
- **참조 상세 설계 문서**: [15_TREESCAN_FUTURE_INTEGRATION/INTEGRATION_ARCHITECTURE.md](../15_TREESCAN_FUTURE_INTEGRATION/INTEGRATION_ARCHITECTURE.md), [08_SEARCH_CONNECTORS/EXTERNAL_CONNECTOR_CONTRACT.md](../08_SEARCH_CONNECTORS/EXTERNAL_CONNECTOR_CONTRACT.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 플랫폼은 독립적으로 운영되면서도 결제 사업자(PG), 소셜 로그인(OIDC), 이메일 발송(SES/SMTP), Android 푸시 알림(FCM/WebPush), 외부 앱 리포지토리(F-Droid 등), 그리고 장래의 TreeScan 검색엔진과 안전하게 통신해야 한다. 특정 외부 사업자의 장애나 API 변경이 코어 비즈니스 로직을 무너뜨리지 않도록 부패 방지 계층(Anti-Corruption Layer, ACL)과 서킷 브레이커를 설계해야 한다.
- **달성 목표**: 모든 외부 연동 지점을 Go 인터페이스로 추상화하고, 웹훅 서명 검증, 타임아웃/서킷 브레이커, 로컬/CI용 샌드박스 에뮬레이터, TreeScan 전용 어댑터 규격을 명세한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **Backend / Integration Agent**: 각 외부 연동 어댑터 및 서킷 브레이커 구현.
- **DevOps / Admin**: 외부 연동 API 자격증명 로테이션 및 연동 상태 헬스체크 모니터링.

## 3. 필수 기능 (Mandatory Features)
1. **포트 및 어댑터(Hexagonal Architecture) 기반 외부 의존성 격리**:
   - `PaymentGatewayPort` (Stripe / Toss Payments / LocalSandboxAdapter)
   - `IdentityProviderPort` (Google OIDC / GitHub OAuth / Apple OIDC / Future TreeScan OIDC)
   - `NotificationSenderPort` (AWS SES/SMTP Email / FCM Push / Webhook Notifier)
   - `ExternalCatalogConnectorPort` (F-Droid Index v2 / Partner JSON Feed)
   - `TreeScanFederationPort` (카탈로그 피드 노출, 실시간 통합 검색 위임 응답, SSO 토큰 교환).
2. **외부 호출 보호 장치 (Resilience4Go 패턴)**:
   - 모든 외부 HTTP 호출은 명시적 `context.WithTimeout` (기본 5초, 대용량 피드 다운로드 30초), 지수 백오프 재시도, 서킷 브레이커(최근 20회 요청 중 실패율 50% 초과 시 30초간 Open)를 필수 적용한다.
3. **수신 웹훅 보안 표준**:
   - 외부에서 당사 서버로 들어오는 모든 웹훅(`/api/v1/commerce/webhooks/*`)은 Raw Body에 대한 `HMAC-SHA256` 서명 검증, 타임스탬프 유효기간(±300초) 검증, `webhook_events` 테이블 기반 멱등성 중복 제거를 거친다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **알림(Notification) 발송 아키텍처**:
  - 개발자 심사 승인/반려 안내, 정산 완료 안내, 사용자 구매 영수증, 긴급 보안 삭제 권고는 DB `notifications` 테이블에 먼저 저장(인앱 알림 센터 즉시 조회 가능)한 후, 비동기 워커가 이메일 및 푸시 알림을 발송한다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- 사용자 웹 및 개발자 콘솔 상단의 인앱 알림 벨(`NotificationCenterDropdown.vue`) 및 알림 수신 설정 화면(`NotificationPreferencesView.vue`).

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/platform/httpclient/resilient_client.go`: OTel 분산 트레이싱 헤더(`traceparent`), 타임아웃, 서킷 브레이커가 내장된 표준 외부 통신 클라이언트.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**: `webhook_events` (`id`, `provider`, `external_event_id UNIQUE`, `payload JSONB`, `status`, `processed_at`), `notifications` (`id`, `user_id`, `category`, `title`, `body`, `action_url`, `is_read`, `created_at`).

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `GET /api/v1/users/me/notifications`: 내 인앱 알림 목록 조회.
- `PATCH /api/v1/users/me/notifications/{notificationId}/read`: 알림 읽음 처리.
- TreeScan 연동 어댑터 API(`/api/v1/treescan/*`)는 `Z — Zero-Trust & TreeScan` 및 [15_TREESCAN_FUTURE_INTEGRATION/SEARCH_API_ADAPTER.md](../15_TREESCAN_FUTURE_INTEGRATION/SEARCH_API_ADAPTER.md)에 상세 정의한다.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **서킷 브레이커 상태**: `CLOSED` (정상) → `OPEN` (장애 감지, 즉시 빠른 실패 또는 Fallback 반환) → `HALF_OPEN` (시험 요청 3건 통과 시 `CLOSED` 복귀).

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 외부 연동용 API Secret 및 Webhook Signing Secret은 최소 90일 주기로 무중단 교체(Dual-Key Grace Period: 구 키와 신 키를 동시에 24시간 허용 후 구 키 폐기)할 수 있도록 배열 형태의 키 검증을 지원한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 외부 결제창 호출이나 외부 소셜 로그인 이동 시 현재 진행 단계를 오버레이 스피너로 표시하고, 사용자가 창을 닫거나 취소했을 때 즉시 안내 메시지와 함께 이전 상태를 복구한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **외부 API 서킷 브레이커 동작 테스트 (`TC-OPS-006`)**: 외부 PG/알림 서버가 500 에러 또는 지연(10초)을 반환하도록 시뮬레이션했을 때 5초 타임아웃 후 서킷이 열리고 코어 API 스레드 풀이 고갈되지 않음을 검증.
- **웹훅 서명 위조 차단 테스트 (`TC-API-002`)**: 잘못된 HMAC 서명이나 10분 지난 타임스탬프를 가진 웹훅 요청 시 `401 ERR_WEBHOOK_SIGNATURE_INVALID` 반환 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `I — Identity`, `M — Marketplace Commerce`, `E — External Store Connectors`, `Z — TreeScan Integration`과 연동된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P1`
- **개발 Phase**: 알림/인증 연동은 `Phase 1~3`, PG 연동은 `Phase 7`, TreeScan 어댑터는 `Phase 10`에서 구현한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 이메일(SMTP/SES) 발송 한도 초과 또는 일시 장애** → 인앱 알림(`notifications` 테이블)은 이미 커밋되어 있으므로 사용자는 웹/앱 내 알림 센터에서 심사 결과와 영수증을 즉시 확인할 수 있으며, 이메일만 큐에서 지연 재발송된다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 외부 연동 어댑터 패키지 (`backend/internal/platform/integrations/*`, `backend/internal/modules/notification/*`).

## 17. Acceptance Criteria (수용 기준)
1. 외부 연동 서비스가 모두 중단된 상태에서도 카탈로그 조회, 검색, 무료 앱 다운로드 등 핵심 자체 기능이 정상 동작해야 한다.
2. 위조된 웹훅 요청은 100% 차단되어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-OPS-002/circuit_breaker_and_webhook_test.log`: 서킷 브레이커 개폐 및 웹훅 서명 검증 테스트 로그.

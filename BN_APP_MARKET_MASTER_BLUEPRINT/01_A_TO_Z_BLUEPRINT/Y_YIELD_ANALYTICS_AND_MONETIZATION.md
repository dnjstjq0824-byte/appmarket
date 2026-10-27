# Y — Yield, Analytics & Monetization — 검색·노출·설치 전환율 퍼널, 스폰서 광고 엔진 및 운영 KPI 분석 시스템 (Conversion Funnel Analytics, Sponsored Advertising Engine & Operational KPI Telemetry)

- **문서 ID**: `DOC-AZ-Y`
- **알파벳 코드**: `Y`
- **관련 요구사항 ID**: `REQ-Y-001`, `REQ-Y-002`
- **연관 Work Order ID**: `AM-ADS-001`, `AM-API-005`
- **연관 테스트 케이스 ID**: `TC-COM-003`, `TC-FUNC-014`
- **참조 상세 설계 문서**: [10_COMMERCE_ADS_ANALYTICS/ADVERTISING_ENGINE.md](../10_COMMERCE_ADS_ANALYTICS/ADVERTISING_ENGINE.md), [10_COMMERCE_ADS_ANALYTICS/ANALYTICS_EVENTS.md](../10_COMMERCE_ADS_ANALYTICS/ANALYTICS_EVENTS.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 제12조 및 제3조 Y항에 따라, 사용자가 검색에서 상세 페이지 조회, 다운로드 시작, 실제 Android 설치 완료(또는 외부 스토어 이동), 유료 구매에 이르는 **전체 전환율 퍼널(Conversion Funnel)**을 정밀하게 측정하고, 개발자가 자사 앱을 홍보할 수 있는 투명하고 부정클릭이 방지된 **자체 스폰서 광고 엔진**을 운영하기 위해 필요하다.
- **달성 목표**: 표준 분석 이벤트 스키마(`AnalyticsEventEnvelope`), 전환 퍼널 집계 파이프라인, 스폰서 광고 경매/예산/노출·클릭 과금 엔진, 광고 부정행위(Click Fraud) 필터링 및 KPI 대시보드를 설계한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **개발자 / 광고주 (Advertiser / Developer Org)**: 앱의 유입 경로별 전환율 분석 조회, 스폰서 검색/카테고리 광고 캠페인 생성, 일일/총 예산 설정, ROAS(광고비 대비 수익률) 확인.
- **Ad Reviewer / Finance Admin**: 광고 캠페인 소재 및 타기팅 적정성 심사, 광고비 선불 예치금/후불 청구 정산 관리.
- **일반 사용자**: 명확하게 `'Sponsored (광고)'` 표시가 된 추천 앱을 확인하고 선택적으로 탐색.

## 3. 필수 기능 (Mandatory Features)
1. **6단계 핵심 전환 퍼널 추적 (End-to-End Funnel Telemetry)**:
   - `IMPRESSION` (검색/차트/추천 카드 노출) → `DETAIL_VIEW` (앱 상세 페이지 진입) → `INSTALL_CLICK` (직접 다운로드 또는 외부 스토어 링크 클릭) → `DOWNLOAD_COMPLETE` (APK 바이트 수신 및 해시 검증 완료) → `INSTALL_SUCCESS` (`PackageInstaller` OS 설치 완료) → `PURCHASE_PAID` (유료 앱/인앱 결제 완료).
2. **광고 시스템 전 주기 구현 (제12조 필수 항목)**:
   - 광고주 자격 확인(검증된 개발자 조직만 자사 보유 앱 광고 가능 — 제3자 피싱 앱 광고 원천 차단) → 캠페인 생성(키워드/카테고리/국가/기기 타기팅, 일일 예산, CPC 입찰가) → 관리자 광고 승인 → 실시간 경매(`eCPM = Bid_CPC * Predicted_CTR * QualityScore`) 및 노출 → `'Sponsored'` 배지 의무 표시 → 클릭/노출 측정 → 부정클릭 필터링 후 일일 예산 차감 및 광고 정산.
3. **광고 부정클릭(Click Fraud) 방지 엔진**:
   - 동일 사용자/IP/기기 지문에서 1시간 내 동일 광고 캠페인 반복 클릭 시 최초 1회만 유효 과금 클릭(`BILLABLE_CLICK`)으로 인정하고 나머지는 `INVALID_DUPLICATE_CLICK`으로 무료 처리하여 광고주 예산을 보호한다.
   - 서명된 1회용 노출 토큰(`impression_token`) 없이 클릭 URL만 직접 호출하는 봇 트래픽은 `INVALID_MISSING_IMPRESSION`으로 과금에서 제외한다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **오가닉 검색 품질 보호 규칙 (Organic Integrity Rule)**:
  - 검색 결과 상단 스폰서 광고 슬롯은 페이지당 **최대 2개**로 엄격히 제한하며, 광고 입찰가가 아무리 높아도 검색어와의 최소 관련도 임계치(`BM25 Relevance Threshold`)를 넘지 못하거나 사용자 기기와 호환되지 않는 앱은 광고로 노출하지 않는다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **`web-user`**: `SponsoredAppCard.vue` (WCAG 명도 대비를 준수하는 명확한 `광고 / Sponsored` 배지 및 `'왜 이 광고가 표시되나요?'` 툴팁 포함).
- **`web-dev`**: `AnalyticsFunnelView.vue` (유입 채널별 노출→상세→다운로드→설치 전환율 차트), `AdCampaignManagerView.vue` (캠페인 생성/예산/입찰가 관리).
- **`web-admin`**: `AdApprovalQueueView.vue`, `PlatformKpiDashboardView.vue`.

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/ads/` (광고 경매, 노출 토큰 발급, 부정클릭 필터, 예산 차감기) 및 `internal/modules/analytics/` (배치 이벤트 수집 및 일별 롤업 집계기).

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**: `ad_campaigns`, `ad_creatives`, `ad_impressions_log`, `ad_clicks_log`, `ad_events_daily`, `app_metrics_daily`, `install_events`.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `POST /api/v1/events/batch`: 클라이언트(웹/Android) 분석 이벤트 배치 수집.
- `GET /api/v1/ads/serve`: 슬롯/키워드/국가/기기 기준 스폰서 광고 조회 및 `impression_token` 발급.
- `POST /api/v1/ads/click`: `impression_token` 검증, 부정클릭 판별 후 과금 처리.
- `POST /api/v1/developer/orgs/{orgId}/ads/campaigns`: 광고 캠페인 생성.
- `GET /api/v1/developer/apps/{appId}/analytics`: 기간별 전환율 퍼널 및 버전별 설치 유지율 조회.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **광고 캠페인 상태 (`campaign_status`)**: `DRAFT` → `PENDING_APPROVAL` → `ACTIVE_SERVING` ⇄ `PAUSED_DAILY_BUDGET_CAP` (일 예산 소진 시 자정까지 자동 일시정지) / `PAUSED_BY_ADVERTISER` → `COMPLETED_TOTAL_BUDGET_EXHAUSTED` | `REJECTED_POLICY` | `SUSPENDED`.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 개인정보 보호(Privacy-by-Design): 분석 이벤트 수집 시 사용자의 실명이나 이메일을 로그에 남기지 않으며, 일일 회전 솔트 기반 익명 세션 해시와 집계 지표만 사용하고 사용자의 맞춤형 광고 거부(`opt_out_personalized_ads`) 설정을 존중한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 다크 패턴 금지: 스폰서 광고 카드를 일반 검색 결과와 구별할 수 없게 숨기거나 닫기/구별 표시를 흐리게 만드는 다크 패턴을 절대 사용하지 않는다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **광고 예산 초과 방지 및 부정클릭 무효화 테스트 (`TC-COM-003`)**: 일일 예산이 10,000원(클릭당 1,000원)인 캠페인에 동일 기기에서 50회 연속 클릭 및 서로 다른 기기에서 20회 클릭을 발생시켰을 때, 동일 기기 중복 클릭은 1회만 과금되고 총 과금액이 정확히 10,000원에서 멈추며 캠페인이 `PAUSED_DAILY_BUDGET_CAP`으로 전환되는지 검증.
- **전환 퍼널 집계 정합성 테스트 (`TC-FUNC-014`)**: 노출→상세→다운로드→설치 이벤트 주입 후 일별 롤업 테이블(`app_metrics_daily`)의 수치와 전환율이 정확히 계산되는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `B — Business Model`, `M — Marketplace Commerce`, `O — OpenSearch`, `W — Workflow & Administration`과 연동된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P1`
- **개발 Phase**: 기본 설치/검색 이벤트 로깅은 `Phase 4~5`, 광고 시스템 및 고급 수익 분석은 `Phase 7`에서 구현한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 분석 이벤트 수집 트래픽 폭주** → 이벤트 수집 엔드포인트(`/api/v1/events/batch`)는 메모리 링버퍼 + Redis Streams 비동기 적재로 동작하며, 큐 포화 시 코어 결제/다운로드 트랜잭션에 영향을 주지 않도록 비핵심 노출 로그부터 샘플링(Load Shedding)한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 광고 경매/부정클릭 방지 엔진 (`backend/internal/modules/ads/*`) 및 퍼널 분석 모듈 (`backend/internal/modules/analytics/*`).

## 17. Acceptance Criteria (수용 기준)
1. 모든 노출된 광고에 `is_sponsored: true` 플래그와 시각적 `'Sponsored'` 배지가 포함되어야 한다.
2. 유효한 `impression_token`이 없거나 동일 기기에서 반복 발생한 클릭은 광고비를 차감하지 않아야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-ADS-001/click_fraud_and_budget_cap_test.log`: 부정클릭 필터링 및 예산 상한선 즉시 정지 테스트 로그.

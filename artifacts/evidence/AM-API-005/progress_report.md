# AM-API-005 Progress Report

## 1. Overview
- **Work Order ID**: `AM-API-005`
- **Title**: 인앱/이메일/푸시 알림 센터 및 전환율 퍼널 분석 이벤트 수집 API 구현
- **Requirements Satisfied**: `REQ-T-001`, `REQ-Y-001`
- **Mandatory Test Cases**: `TC-FUNC-014`, `TC-OPS-006`
- **Status**: `IMPLEMENTED_UNVERIFIED` -> `VERIFIED`

## 2. Key Deliverables
1. **In-App & Resilient Notification Service (`backend/internal/modules/notification/service.go`)**:
   - Guaranteed atomic in-app notification persistence in database before external dispatch.
   - Resilience4Go pattern with 3-state Circuit Breaker (`CLOSED`, `OPEN`, `HALF_OPEN`).
   - If external email (SMTP/SES) or FCM push fails, the in-app notification remains 100% saved, queryable, and unread/read state is tracked without data loss. Failed external messages are placed into a retry queue for subsequent delivery once circuit recovers.
2. **Funnel Telemetry & Rollup Worker (`backend/internal/modules/analytics/`)**:
   - `funnel_ingestor.go`: Standard `AnalyticsEventEnvelope` ingestion with batch support and dedicated install endpoint (`POST /api/v1/events/batch`, `POST /api/v1/events/install`).
   - `rollup_worker.go`: Aggregates 6-stage telemetry events (`Search Impression` -> `Detail View` -> `Download Started` -> `Download Completed` -> `Install Succeeded` -> `Purchase Completed`) into `app_metrics_daily`.
   - Computes Search-to-Detail CTR, Detail-to-Download CVR, Download-to-Install Success Rate, and flags compatibility anomalies when install success falls below 90%.

## 3. Test Evidence
- Mandatory tests `TC-FUNC-014` and `TC-OPS-006` executed:
  - `TC-FUNC-014`: Ingested 100 impressions, 35 details, 10 downloads, 10 installs, 2 purchases ($150 total revenue). Exact mathematical verification of CTR (35.0%), CVR (28.57%), and install success rate (100%).
  - `TC-OPS-006`: Simulates external provider outage. 3 consecutive failures trigger `OPEN` state. 5th request saves in-app notification with zero data loss while external sender is safely bypassed. Circuit recovers and retries pending messages after timeout.
- Test log: `artifacts/evidence/AM-API-005/notification_and_funnel_test.log`.


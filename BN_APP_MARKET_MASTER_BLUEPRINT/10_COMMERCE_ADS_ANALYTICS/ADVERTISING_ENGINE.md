# ADVERTISING_ENGINE.md — 스폰서 앱 광고 경매, 노출 토큰 및 부정클릭 방지 엔진 상세 설계

- **문서 ID**: `DOC-COM-002`
- **관련 조항**: 제12조 (광고 및 수익화 시스템), 제3조 Y항
- **연관 Work Order**: `AM-ADS-001`

---

## 1. 스폰서 광고 경매 및 랭킹 알고리즘 (Second-Price Generalized Auction)

1. **후보 필터링**: 요청된 슬롯(`SEARCH_TOP`, `CATEGORY_FEATURED`, `HOME_SPOTLIGHT`), 사용자 국가(`target_countries`), 키워드(`target_keywords`), 기기 호환성을 만족하며 `campaign_status = 'ACTIVE_SERVING'`이고 `spent_today_minor + cpc_bid_minor <= daily_budget_minor`인 캠페인만 후보로 추출합니다.
2. **광고 랭크 점수 (`AdRank`)**:
   `AdRank = cpc_bid_minor * predicted_ctr * app_quality_score`
3. **실제 과금 단가 (Generalized Second-Price)**:
   1위 낙찰 광고주가 지불하는 실제 클릭 단가 `ActualCPC`는 차순위 광고주의 `AdRank`를 이기는 데 필요한 최소 금액(`ceil((AdRank_runner_up / (predicted_ctr_winner * app_quality_score_winner)) + 1)`)으로 결정되어 과열 입찰을 방지합니다.

---

## 2. 노출 토큰(`impression_token`) 및 3중 부정클릭(Click Fraud) 필터링

1. 광고 서빙 시(`GET /api/v1/ads/serve`), 서버는 `impression_id`, `campaign_id`, `actual_cpc_minor`, `client_ip_subnet`, `device_hash`, `issued_at`을 묶어 서버 비밀키로 서명한 `impression_token` (유효기간 10분)을 발급합니다.
2. 클릭 발생 시(`POST /api/v1/ads/click`):
   - **Rule 1 (서명 및 만료 검증)**: 유효한 `impression_token` 서명이 없거나 10분이 지난 경우 → `INVALID_EXPIRED_TOKEN` (과금 0원).
   - **Rule 2 (노출-클릭 최소 인지 시간)**: 노출 후 `150ms` 미만에 발생한 초고속 자동 클릭(봇 스크립트) → `INVALID_BOT_SPEED` (과금 0원).
   - **Rule 3 (중복 클릭 디듀플리케이션)**: Redis `SETNX ad:click:dedup:{campaign_id}:{device_or_ip_hash}` (TTL 3600초)가 이미 존재하면 → `INVALID_DUPLICATE_CLICK` (과금 0원, 앱 상세 페이지 이동은 정상 수행).
   - 3가지 규칙을 모두 통과한 경우에만 원자적 SQL 업데이트로 예산을 차감하고 일일 예산 도달 시 즉시 `PAUSED_DAILY_BUDGET_CAP`으로 전환합니다.

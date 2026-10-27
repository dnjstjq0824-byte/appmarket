# SEARCH_API_ADAPTER.md — TreeScan 연동용 초저지연 검색 어댑터 및 카탈로그 피드 API 규격

- **문서 ID**: `DOC-TS-002`
- **관련 문서**: [04_API_SPECIFICATION/openapi.yaml](../04_API_SPECIFICATION/openapi.yaml)

---

## 1. 실시간 연합 검색 어댑터 (`GET /api/v1/treescan/federated-search`)

- **성능 목표**: `p95 < 80ms`, `p99 < 120ms` (전용 Redis 캐시 + OpenSearch 경량 프로젝션 쿼리).
- **인증 헤더**: `X-TreeScan-Service-Token: <HMAC_OR_MTLS_TOKEN>`
- **응답 페이로드 예시**:
  ```json
  {
    "query": "오프라인 지도",
    "took_ms": 14,
    "items": [
      {
        "app_id": "0192ce10-4b00-7000-8000-000000000301",
        "canonical_slug": "osmand-maps-navigation",
        "primary_package_name": "net.osmand.plus",
        "title": "OsmAnd — 오프라인 지도 및 내비게이션",
        "developer_name": "OsmAnd B.V.",
        "icon_url": "https://cdn.appmarket.bookingnetwork.internal/icons/osmand.webp",
        "bayesian_rating": 4.72,
        "rating_count": 1840,
        "monetization_type": "FREE",
        "can_direct_install": true,
        "available_sources": ["BN_HOSTED", "FDROID_REPO", "GOOGLE_PLAY_LINK"],
        "provenance_summary": "BN Verified Direct APK & F-Droid Official",
        "canonical_market_url": "https://appmarket.bookingnetwork.internal/apps/osmand-maps-navigation?utm_source=treescan"
      }
    ]
  }
  ```

---

## 2. 카탈로그 증분 색인 피드 (`GET /api/v1/treescan/catalog-feed`)

- TreeScan 인덱서가 주기적으로 변경된 앱 카탈로그(`PUBLISHED`, `SUSPENDED_BY_ADMIN`, `ARCHIVED`)만 가져갈 수 있도록 `(updated_at ASC, id ASC)` 복합 커서 페이지네이션을 제공합니다.
- 삭제되거나 차단된 앱은 `"catalog_status": "SUSPENDED_BY_ADMIN"`, `"action": "REMOVE_FROM_INDEX"` 지시자를 명시하여 TreeScan 검색 결과에서도 즉시 제거되도록 동기화합니다.

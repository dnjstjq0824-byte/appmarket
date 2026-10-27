# ANALYTICS_EVENTS.md — 전환율 퍼널 이벤트 스키마 및 운영 KPI 지표 정의

- **문서 ID**: `DOC-COM-003`
- **관련 조항**: 제3조 Y항
- **연관 Work Order**: `AM-API-005`, `AM-ADS-001`

---

## 1. 표준 텔레메트리 이벤트 봉투 스키마 (`AnalyticsEventEnvelope`)

```json
{
  "event_id": "0192ce10-7788-7000-8000-abcdef123456",
  "event_name": "app.install.succeeded",
  "occurred_at": "2026-10-27T14:25:10Z",
  "session_hash": "a8f5f167f44f4964e6c998dee827110c...",
  "platform": "ANDROID_CLIENT",
  "country_code": "KR",
  "locale": "ko-KR",
  "device": {
    "sdk_int": 34,
    "abi": "arm64-v8a",
    "form_factor": "PHONE"
  },
  "properties": {
    "app_id": "0192ce10-4b00-7000-8000-000000000101",
    "distribution_id": "0192ce10-4b00-7000-8000-000000000501",
    "release_id": "0192ce10-4b00-7000-8000-000000000901",
    "source_type": "BN_HOSTED",
    "duration_ms": 4820
  }
}
```

---

## 2. 핵심 운영 및 비즈니스 KPI 산출 공식

| KPI 지표명 | 산출 공식 | 목표치 / 알람 임계치 |
| :--- | :--- | :--- |
| **Search-to-Detail CTR** | `검색 결과에서 앱 상세 클릭 수 / 총 유효 검색 횟수` | `>= 28%` |
| **Detail-to-Download CVR** | `(직접 다운로드 시작 + 외부 스토어 이동 클릭) / 앱 상세 조회 수` | `>= 18%` |
| **Download-to-Install Success Rate** | `INSTALL_SUCCEEDED 이벤트 수 / DOWNLOAD_COMPLETED 이벤트 수` | `>= 94%` (`< 90%` 하락 시 `PackageInstaller` 호환성 경보 발령) |
| **Malware Escape Rate** | `출시 후 사후 발견된 악성 앱 수 / 총 출시 승인 앱 수` | `0.00%` (Zero-Tolerance) |
| **Review SLA Compliance** | `24시간 이내 1차 심사 완료된 프로덕션 릴리스 수 / 전체 제출 수` | `>= 95%` |

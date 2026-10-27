# MONITORING.md — OpenTelemetry 기반 분산 추적, 메트릭, 로그 및 경보(Alerting) 설계

- **문서 ID**: `DOC-OPS-003`
- **연관 Work Order**: `AM-OPS-001`

---

## 1. 3대 관제 축 (Metrics, Traces, Logs) 표준 규격

1. **Distributed Tracing (OpenTelemetry → Tempo)**:
   - 모든 HTTP 요청과 비동기 워커 Job(`outbox_events` → Redis Streams)에 W3C `traceparent` 컨텍스트를 전파하여, 개발자의 APK 업로드 요청부터 격리 워커의 서명 검증·악성코드 스캔·CDN 배포까지 단일 Trace ID로 전 구간 지연시간을 추적합니다.
2. **Metrics (Prometheus → Grafana)**:
   - API 지연시간 히스토그램(`bn_http_request_duration_seconds_bucket`), 다운로드 토큰 발급 수(`bn_download_tokens_issued_total`), 스캐너 검역 건수(`bn_malware_quarantined_total`), 큐 적체 깊이(`bn_worker_queue_lag`), 원장 불균형 카운터(`bn_ledger_imbalance_errors_total`).
3. **Structured Logging (Go `slog` JSON → Loki)**:
   - 모든 로그는 `timestamp`, `level`, `trace_id`, `span_id`, `module`, `actor_id`, `msg` 필드를 가진 JSON으로 출력하며, 비밀번호/토큰/계좌번호는 로거 미들웨어에서 자동 마스킹(`[REDACTED]`)합니다.

---

## 2. 핵심 경보(Alertmanager) 임계치 규칙

| 경보명 | 조건식 (PromQL) | 심각도 | 자동 대응 조치 |
| :--- | :--- | :---: | :--- |
| `LedgerImbalanceDetected` | `increase(bn_ledger_imbalance_errors_total[1m]) > 0` | `P0_CRITICAL` | 정산 승인 자동 동결 및 재무/보안팀 즉시 호출 |
| `AuditChainIntegrityBroken` | `bn_audit_chain_verify_status == 0` | `P0_CRITICAL` | 보안 침해 대응 프로토콜 가동 |
| `MalwareScanQueueBacklog` | `bn_worker_queue_lag{queue="q:artifact:malware_scan"} > 50` | `P1_HIGH` | `worker-malware-scanner` HPA 스케일아웃 트리거 |
| `HighInstallFailureRate` | `rate(bn_install_failed_total[15m]) / rate(bn_install_started_total[15m]) > 0.10` | `P1_HIGH` | 해당 릴리스 단계적 배포 자동 일시정지 권고 |

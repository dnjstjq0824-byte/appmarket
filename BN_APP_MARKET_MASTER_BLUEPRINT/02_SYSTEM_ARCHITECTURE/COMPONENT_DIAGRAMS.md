# COMPONENT_DIAGRAMS.md — 컴포넌트 구조 및 모듈 상호작용 다이어그램

- **문서 ID**: `DOC-ARCH-002`
- **관련 문서**: [SYSTEM_OVERVIEW.md](SYSTEM_OVERVIEW.md), [EVENT_FLOW.md](EVENT_FLOW.md)

---

## 1. 전체 시스템 컨텍스트 및 트래픽 흐름도 (System Context & Traffic Topology)

```
[End Users (Web / Android)]     [Developers (Web / CI)]      [Platform Admins (MFA/VPN)]
            │                              │                              │
            ▼                              ▼                              ▼
   ┌───────────────────────────────────────────────────────────────────────────────┐
   │                     Global Edge WAF / DDoS Protection / CDN                   │
   └────────────────┬──────────────────────┬──────────────────────┬────────────────┘
                    │                      │                      │
        ┌───────────▼──────────┐ ┌─────────▼──────────┐ ┌─────────▼──────────┐
        │   web-user (Nuxt 3)  │ │  web-dev (Vue 3)   │ │ web-admin (Vue 3)  │
        └───────────┬──────────┘ └─────────┬──────────┘ └─────────┬──────────┘
                    └──────────────────────┼──────────────────────┘
                                           ▼
                    ┌─────────────────────────────────────────────┐
                    │       backend-api (Go Modular Monolith)     │
                    │  [IAM | Catalog | Dist | Search | Commerce] │
                    └───┬──────────────┬──────────────┬───────────┘
                        │              │              │
         ┌──────────────▼───┐   ┌──────▼──────┐   ┌───▼──────────────┐
         │  PostgreSQL 16   │   │  Redis 7.4  │   │ OpenSearch 2.17  │
         │ (Primary+Replica)│   │(Cache+Queue)│   │ (Search Cluster) │
         └──────────────┬───┘   └──────┬──────┘   └───▲──────────────┘
                        │              │              │
                        │     ┌────────▼────────┐     │
                        │     │ Isolated Workers│─────┘
                        │     │ - apk-analyzer  │
                        │     │ - malware-scan  │──────► [S3 Quarantine Bucket]
                        │     │ - cdn-publisher │──────► [S3 Public CDN Bucket]
                        │     │ - ext-connector │──────► [F-Droid / Partner Feeds]
                        └─────┴─────────────────┘
```

---

## 2. APK 업로드·검증·심사·배포 컴포넌트 상호작용

1. **Upload Initiation**: `web-dev` → `backend-api (distribution)` → `upload_sessions` 생성 및 `bn-quarantine-bucket` Presigned Multipart URL 반환.
2. **Quarantine Storage**: `web-dev`가 바이너리를 `bn-quarantine-bucket`에 직접 업로드 완료 후 `/complete` 호출.
3. **Outbox Dispatch**: `backend-api`가 `release_artifacts`(`UPLOADED_QUARANTINED`) 및 `outbox_events` 기록 → `worker-apk-analyzer`가 `q:artifact:verify` 소비.
4. **Manifest & Signature Check**: `worker-apk-analyzer`가 SHA-256, AXML 매니페스트, `apksigner` v2/v3 서명 지문 및 기존 릴리스와의 서명 연속성을 검증 후 `q:artifact:malware_scan` 발행.
5. **Security Sandbox Scan**: 외부 네트워크가 차단된 `worker-malware-scanner`가 ClamAV, YARA, 고위험 권한 점수를 계산하여 `security_scan_reports`에 저장.
6. **Admin Review & Approval**: `web-admin`에서 심사자가 검토 후 승인(`APPROVED_PENDING_PUBLISH`) → 출시 트리거 시 `worker-cdn-publisher`가 `bn-quarantine-bucket`에서 `bn-public-cdn-bucket`의 불변 경로(`/packages/{app_id}/{release_id}/{sha256}.apk`)로 복사하고 OpenSearch 인덱스를 갱신.

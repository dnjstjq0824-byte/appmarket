# EXECUTION_SEQUENCE.md — Wave 0 ~ Wave 5 단계별 상세 실행 순서 및 병렬화 가이드

- **문서 ID**: `DOC-EXEC-003`
- **관련 문서**: [11_WORK_ORDERS/WORK_ORDER_DEPENDENCIES.md](../11_WORK_ORDERS/WORK_ORDER_DEPENDENCIES.md)

---

## 1. Wave별 실행 순서 및 동기화 배리어 (Synchronization Barriers)

| 실행 단계 | 포함 Phase | 동시 실행 가능한 Work Order 묶음 | 다음 Wave 진입 조건 (Gate Barrier) |
| :--- | :---: | :--- | :--- |
| **Wave 0** | Phase 0 | `AM-ARCH-001` → `AM-OPS-001` | 모노레포 빌드 및 로컬 컨테이너 인프라(`/healthz/ready`) 기동 확인 |
| **Wave 1** | Phase 1 | `AM-DB-001` → `AM-DB-002` → `AM-API-001` → (`AM-SEC-001` ∥ `AM-ARCH-002`) | 34개 테이블 마이그레이션 완료, JWT/MFA/RBAC/감사체인 및 워커 큐 동작 확인 |
| **Wave 2** | Phase 2 ~ 3 | (`AM-API-002` ∥ `AM-API-003` ∥ `AM-DIST-001`) → (`AM-WEB-001` ∥ `AM-DEV-001` ∥ `AM-ADMIN-001` ∥ `AM-SEC-002`) → (`AM-DEV-002` ∥ `AM-ADMIN-002`) | 자체 앱 등록 → APK 업로드 → 서명/악성코드 검사 → 관리자 심사 승인 루프 완결 |
| **Wave 3** | Phase 4 ~ 5 | (`AM-DIST-002` ∥ `AM-DIST-003` ∥ `AM-SEARCH-001`) → (`AM-ANDROID-001`→`003` ∥ `AM-SEARCH-002` ∥ `AM-API-006` ∥ `AM-WEB-002`) | Android 실제 기기 설치/업데이트 완료 및 OpenSearch + F-Droid 통합 검색 완료 |
| **Wave 4** | Phase 6 ~ 8 | `AM-SEC-003` ∥ `AM-MOD-001` ∥ (`AM-BILLING-001`→`002`) ∥ `AM-ADS-001` ∥ `AM-API-004` ∥ `AM-API-005` ∥ (`AM-WEB-003` ∥ `AM-DEV-003` ∥ `AM-ADMIN-003`) ∥ (`AM-ANDROID-004` ∥ `AM-OPS-002` ∥ `AM-OPS-003`) | 유료 결제·복식부기 원장·정산·스폰서 광고·글로벌 차단·DR 백업 복원 드릴 모두 통과 |
| **Wave 5** | Phase 9 ~ 10 | `AM-TEST-001` → `AM-TEST-002` (독립 감사 최종 승인) → `AM-TREESCAN-001` (Phase 10 어댑터) | 8대 검증군 전수 통과, 독립 감사 `VERIFIED` 획득, TreeScan 어댑터 계약 테스트 통과 |

# MASTER_WORK_ORDER.md — 전체 41개 실행형 작업지시서(Work Orders) 마스터 총람

- **문서 ID**: `DOC-WO-001`
- **관련 조항**: 제14조 (개발 작업의 세부 분해)
- **기계 판독 인덱스**: [WORK_ORDER_INDEX.csv](WORK_ORDER_INDEX.csv)
- **의존성 그래프**: [WORK_ORDER_DEPENDENCIES.md](WORK_ORDER_DEPENDENCIES.md)

---

## 1. Work Order 분해 원칙 및 현황 요약

- **총 Work Order 수**: **41개** (개별 파일: `11_WORK_ORDERS/individual/AM-*.md` 41개 전수 생성 완료)
- **우선순위별 분포**: `P0` (25개 — 코어 아키텍처, DB, 인증, 카탈로그, APK 배포, 보안, 콘솔, Android, QA), `P1` (15개 — 글로벌 정책, 외부 커넥터, 결제/정산, 광고, DR), `P2` (1개 — Phase 10 TreeScan 연동 어댑터)
- **초기 상태**: 제1조 1.2항에 따라 모든 Work Order의 코드 구현 상태는 `NOT_STARTED`이며, 설계 상태는 `DESIGN_VERIFIED`입니다.

---

## 2. 전체 41개 Work Order 마스터 목록

| Work Order ID | Phase | 우선순위 | 담당 에이전트 | 선행 작업 ID | 필수 검증 테스트 | 현재 구현 상태 |
| :--- | :---: | :---: | :--- | :--- | :--- | :---: |
| [AM-ARCH-001](individual/AM-ARCH-001.md) | Phase 0 | `P0` | `Lead Architect` | `NONE` | `TC-FUNC-001;TC-OPS-001` | `NOT_STARTED` |
| [AM-OPS-001](individual/AM-OPS-001.md) | Phase 0 | `P0` | `DevOps Agent` | `AM-ARCH-001` | `TC-OPS-001;TC-OPS-004` | `NOT_STARTED` |
| [AM-API-001](individual/AM-API-001.md) | Phase 1 | `P0` | `Backend Agent` | `AM-ARCH-001;AM-DB-001` | `TC-SEC-005;TC-API-001` | `NOT_STARTED` |
| [AM-ARCH-002](individual/AM-ARCH-002.md) | Phase 1 | `P0` | `Backend Agent` | `AM-ARCH-001;AM-DB-001` | `TC-FUNC-006;TC-OPS-004` | `NOT_STARTED` |
| [AM-DB-001](individual/AM-DB-001.md) | Phase 1 | `P0` | `Database Agent` | `AM-ARCH-001` | `TC-DB-001;TC-DB-002` | `NOT_STARTED` |
| [AM-DB-002](individual/AM-DB-002.md) | Phase 1 | `P0` | `Database Agent` | `AM-DB-001` | `TC-DB-003;TC-COM-005` | `NOT_STARTED` |
| [AM-SEC-001](individual/AM-SEC-001.md) | Phase 1 | `P0` | `Security Agent` | `AM-API-001;AM-DB-002` | `TC-SEC-001;TC-SEC-006;TC-API-005` | `NOT_STARTED` |
| [AM-ADMIN-001](individual/AM-ADMIN-001.md) | Phase 2 | `P0` | `Frontend Agent` | `AM-ARCH-001;AM-SEC-001` | `TC-FUNC-013;TC-SEC-006` | `NOT_STARTED` |
| [AM-DEV-001](individual/AM-DEV-001.md) | Phase 2 | `P0` | `Frontend Agent` | `AM-ARCH-001;AM-API-003` | `TC-FUNC-003;TC-UI-001` | `NOT_STARTED` |
| [AM-WEB-001](individual/AM-WEB-001.md) | Phase 2 | `P0` | `Frontend Agent` | `AM-ARCH-001;AM-API-002` | `TC-UI-001;TC-UI-002` | `NOT_STARTED` |
| [AM-ADMIN-002](individual/AM-ADMIN-002.md) | Phase 3 | `P0` | `Frontend Agent` | `AM-ADMIN-001;AM-DIST-001;AM-SEC-002` | `TC-FUNC-013;TC-SEC-009` | `NOT_STARTED` |
| [AM-API-002](individual/AM-API-002.md) | Phase 3 | `P0` | `Backend Agent` | `AM-DB-001;AM-API-001` | `TC-FUNC-002;TC-DB-001` | `NOT_STARTED` |
| [AM-API-003](individual/AM-API-003.md) | Phase 3 | `P0` | `Backend Agent` | `AM-API-001;AM-SEC-001` | `TC-FUNC-003;TC-SEC-002` | `NOT_STARTED` |
| [AM-DEV-002](individual/AM-DEV-002.md) | Phase 3 | `P0` | `Frontend Agent` | `AM-DEV-001;AM-DIST-001` | `TC-FUNC-007;TC-FUNC-008` | `NOT_STARTED` |
| [AM-DIST-001](individual/AM-DIST-001.md) | Phase 3 | `P0` | `Backend Agent` | `AM-DB-001;AM-API-001` | `TC-FUNC-007;TC-FUNC-008;TC-FUNC-009;TC-SEC-007` | `NOT_STARTED` |
| [AM-SEC-002](individual/AM-SEC-002.md) | Phase 3 | `P0` | `Security Agent` | `AM-DIST-001` | `TC-SEC-009;TC-SEC-010` | `NOT_STARTED` |
| [AM-ANDROID-001](individual/AM-ANDROID-001.md) | Phase 4 | `P0` | `Android Agent` | `AM-ARCH-001;AM-API-002` | `TC-AND-004` | `NOT_STARTED` |
| [AM-ANDROID-002](individual/AM-ANDROID-002.md) | Phase 4 | `P0` | `Android Agent` | `AM-ANDROID-001;AM-DIST-002` | `TC-AND-001;TC-AND-002` | `NOT_STARTED` |
| [AM-ANDROID-003](individual/AM-ANDROID-003.md) | Phase 4 | `P0` | `Android Agent` | `AM-ANDROID-002` | `TC-AND-003;TC-AND-004` | `NOT_STARTED` |
| [AM-DIST-002](individual/AM-DIST-002.md) | Phase 4 | `P0` | `Backend Agent` | `AM-DIST-001;AM-SEC-002` | `TC-FUNC-012;TC-OPS-003` | `NOT_STARTED` |
| [AM-DIST-003](individual/AM-DIST-003.md) | Phase 4 | `P1` | `Backend Agent` | `AM-DIST-001` | `TC-FUNC-009` | `NOT_STARTED` |
| [AM-API-006](individual/AM-API-006.md) | Phase 5 | `P1` | `Search & Connector Agent` | `AM-API-002;AM-SEARCH-001` | `TC-FUNC-011` | `NOT_STARTED` |
| [AM-SEARCH-001](individual/AM-SEARCH-001.md) | Phase 5 | `P0` | `Search & Connector Agent` | `AM-DB-001;AM-API-002` | `TC-FUNC-010;TC-OPS-005` | `NOT_STARTED` |
| [AM-SEARCH-002](individual/AM-SEARCH-002.md) | Phase 5 | `P1` | `Search & Connector Agent` | `AM-SEARCH-001` | `TC-FUNC-004;TC-SEC-003` | `NOT_STARTED` |
| [AM-WEB-002](individual/AM-WEB-002.md) | Phase 5 | `P0` | `Frontend Agent` | `AM-WEB-001;AM-SEARCH-001;AM-DIST-002` | `TC-UI-003;TC-FUNC-005` | `NOT_STARTED` |
| [AM-API-005](individual/AM-API-005.md) | Phase 6 | `P1` | `Backend Agent` | `AM-API-001;AM-DB-002` | `TC-FUNC-014;TC-OPS-006` | `NOT_STARTED` |
| [AM-MOD-001](individual/AM-MOD-001.md) | Phase 6 | `P1` | `Security Agent` | `AM-API-002;AM-DIST-002` | `TC-FUNC-011;TC-SEC-008;TC-FUNC-013` | `NOT_STARTED` |
| [AM-SEC-003](individual/AM-SEC-003.md) | Phase 6 | `P0` | `Security Agent` | `AM-SEC-001` | `TC-API-001;TC-SEC-002` | `NOT_STARTED` |
| [AM-ADMIN-003](individual/AM-ADMIN-003.md) | Phase 7 | `P1` | `Frontend Agent` | `AM-ADMIN-001;AM-BILLING-001;AM-ADS-001` | `TC-COM-004;TC-COM-006` | `NOT_STARTED` |
| [AM-ADS-001](individual/AM-ADS-001.md) | Phase 7 | `P1` | `Backend Agent` | `AM-BILLING-001;AM-SEARCH-001` | `TC-COM-003` | `NOT_STARTED` |
| [AM-BILLING-001](individual/AM-BILLING-001.md) | Phase 7 | `P1` | `Backend Agent` | `AM-DB-002;AM-API-002` | `TC-COM-001;TC-COM-002;TC-COM-004;TC-COM-005` | `NOT_STARTED` |
| [AM-BILLING-002](individual/AM-BILLING-002.md) | Phase 7 | `P1` | `Backend Agent` | `AM-BILLING-001;AM-SEC-001` | `TC-COM-006` | `NOT_STARTED` |
| [AM-DEV-003](individual/AM-DEV-003.md) | Phase 7 | `P1` | `Frontend Agent` | `AM-DEV-002;AM-BILLING-001;AM-ADS-001` | `TC-COM-001;TC-COM-003` | `NOT_STARTED` |
| [AM-WEB-003](individual/AM-WEB-003.md) | Phase 7 | `P1` | `Frontend Agent` | `AM-WEB-002;AM-BILLING-001` | `TC-COM-001;TC-COM-006` | `NOT_STARTED` |
| [AM-ANDROID-004](individual/AM-ANDROID-004.md) | Phase 8 | `P1` | `Android Agent` | `AM-ANDROID-003` | `TC-AND-001;TC-AND-002;TC-AND-003;TC-AND-004` | `NOT_STARTED` |
| [AM-API-004](individual/AM-API-004.md) | Phase 8 | `P1` | `Backend Agent` | `AM-API-002` | `TC-FUNC-005;TC-SEC-004` | `NOT_STARTED` |
| [AM-OPS-002](individual/AM-OPS-002.md) | Phase 8 | `P1` | `DevOps Agent` | `AM-OPS-001;AM-DB-002` | `TC-OPS-002` | `NOT_STARTED` |
| [AM-OPS-003](individual/AM-OPS-003.md) | Phase 8 | `P1` | `DevOps Agent` | `AM-OPS-001;AM-DIST-002` | `TC-OPS-003` | `NOT_STARTED` |
| [AM-TEST-001](individual/AM-TEST-001.md) | Phase 9 | `P0` | `QA / E2E Agent` | `AM-WEB-002;AM-DEV-002;AM-ADMIN-003;AM-ANDROID-003;AM-BILLING-002` | `TC-FUNC-001;TC-UI-001;TC-UI-002;TC-UI-003;TC-COM-004` | `NOT_STARTED` |
| [AM-TEST-002](individual/AM-TEST-002.md) | Phase 9 | `P0` | `Independent Audit Agent` | `AM-TEST-001;AM-OPS-002` | `TC-OPS-007;TC-OPS-008` | `NOT_STARTED` |
| [AM-TREESCAN-001](individual/AM-TREESCAN-001.md) | Phase 10 | `P2` | `Search & Connector Agent` | `AM-TEST-002` | `TC-API-003;TC-OPS-008` | `NOT_STARTED` |

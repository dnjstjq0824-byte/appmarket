# PHASE_ROADMAP.md — 개발 Phase 0 ~ Phase 10 전체 실행 로드맵 및 Exit Gate 명세

- **문서 ID**: `DOC-DELIV-001`
- **관련 조항**: 제17조 (개발 Phase 전체 계획)
- **연관 문서**: [11_WORK_ORDERS/MASTER_WORK_ORDER.md](../11_WORK_ORDERS/MASTER_WORK_ORDER.md), [12_AGENT_EXECUTION/EXECUTION_SEQUENCE.md](../12_AGENT_EXECUTION/EXECUTION_SEQUENCE.md)

---

## 1. Phase 0 ~ Phase 10 단계별 상세 계획 (제17조 10대 필수 정보 전수 포함)

### Phase 0 — 요구사항, 저장소 및 코어 아키텍처 기반 구축
- **선행조건**: `BN_APP_MARKET_MASTER_BLUEPRINT_V1_0.zip` 인수 및 `FILE_CHECKSUMS.sha256` 무결성 검증 통과.
- **작업 목록**: `AM-ARCH-001`, `AM-OPS-001`.
- **작업 순서**: `AM-ARCH-001` (모노레포 및 Go/pnpm 코어 초기화) → `AM-OPS-001` (Docker 멀티스테이지 컨테이너 및 CI/OTel 구성).
- **병렬 수행 가능성**: `AM-ARCH-001`의 디렉터리 골격 생성 직후 DevOps Agent가 `AM-OPS-001`을 병렬 작성 가능.
- **필요한 인프라**: Docker Engine, GitHub Actions/GitLab CI Runner, 로컬 PostgreSQL/Redis/OpenSearch/MinIO 컨테이너.
- **기능 완료 조건**: Go API 서버 및 Worker 엔트리포인트, 공통 미들웨어, `/healthz/live`, `/healthz/ready` 정상 응답.
- **테스트 완료 조건**: `TC-FUNC-001`(아키텍처 경계 정적 테스트), `TC-OPS-001`(컨테이너 격리 기동 테스트) 통과.
- **위험 요인**: 도메인 간 순환 참조 발생 → `golangci-lint` `depguard` 규칙으로 사전 차단.
- **최종 산출물**: `backend/cmd/`, `backend/internal/platform/`, `deploy/docker-compose.yml`, `.github/workflows/ci-verify.yml`.
- **Phase Exit Gate**: 단일 `docker compose up --build` 명령으로 전 인프라 기동 및 헬스체크 `200 OK` 확인.

### Phase 1 — 데이터베이스, 통합 인증(IAM), RBAC 및 비동기 큐 기초
- **선행조건**: Phase 0 Exit Gate 통과.
- **작업 목록**: `AM-DB-001`, `AM-DB-002`, `AM-API-001`, `AM-SEC-001`, `AM-ARCH-002`.
- **작업 순서**: `AM-DB-001` → `AM-DB-002` → `AM-API-001` → (`AM-SEC-001` ∥ `AM-ARCH-002`).
- **병렬 수행 가능성**: `AM-API-001` 완료 후 Security Agent(`AM-SEC-001`)와 Backend Agent(`AM-ARCH-002`) 병렬 수행.
- **필요한 인프라**: PostgreSQL 16, Redis 7.4, KMS/로컬 암호화 키 볼트.
- **기능 완료 조건**: 34개 테이블 마이그레이션, 복식부기 원장 트리거, Argon2id/JWT 회전/TOTP MFA, RBAC 가드, Dual-Control, 해시 체인 감사 로그, Outbox 워커 큐 동작.
- **테스트 완료 조건**: `TC-DB-001`~`003`, `TC-SEC-001`, `TC-SEC-005`, `TC-SEC-006`, `TC-API-001`, `TC-FUNC-006`, `TC-OPS-004` 통과.
- **위험 요인**: 토큰 재사용 경합 조건 → DB 행 잠금 및 `token_family_id` 일괄 폐기로 해결.
- **최종 산출물**: `03_DATABASE/migrations/*.sql`, `backend/internal/modules/iam/`, `backend/internal/worker/`.
- **Phase Exit Gate**: 34개 테이블 무결성 왕복 테스트 및 MFA/Dual-Control/감사 체인 검증 통과.

### Phase 2 — 사용자 마켓 웹, 개발자 콘솔, 관리자 콘솔 기반 구축
- **선행조건**: Phase 1 Exit Gate 및 OpenAPI 3.1 계약 동결.
- **작업 목록**: `AM-WEB-001`, `AM-DEV-001`, `AM-ADMIN-001`.
- **작업 순서**: `packages/ui-kit` 및 `packages/api-client` 생성 후 3개 웹 앱 병렬 개발.
- **병렬 수행 가능성**: 3개 프론트엔드 앱(`web-user`, `web-dev`, `web-admin`) 디렉터리가 분리되어 있어 100% 병렬 진행 가능.
- **필요한 인프라**: Node.js 22+, pnpm, Nuxt 3 SSR 서버, Playwright 테스트 러너.
- **기능 완료 조건**: 공통 디자인 시스템(다크/라이트 모드), 사용자 탐색/상세 뷰(`USR-01`~`12`), 개발자 온보딩/메타데이터 뷰, 관리자 대시보드/KYB/감사 뷰 구동.
- **테스트 완료 조건**: `TC-UI-001`, `TC-UI-002` (반응형 및 axe-core 접근성 검사) 통과.
- **위험 요인**: SSR 하이드레이션 불일치 → 테마/로케일 쿠키 동기화로 방지.
- **최종 산출물**: `packages/ui-kit/`, `apps/web-user/`, `apps/web-dev/`, `apps/web-admin/`.
- **Phase Exit Gate**: 3개 웹 앱 `pnpm typecheck && pnpm build` 에러 0건 및 SSR JSON-LD 렌더링 확인.

### Phase 3 — 앱 등록, APK/AAB 보안 검증, 심사 및 자체 카탈로그 완성 (Type 1 Store Core)
- **선행조건**: Phase 1 및 Phase 2 완료.
- **작업 목록**: `AM-API-002`, `AM-API-003`, `AM-DIST-001`, `AM-SEC-002`, `AM-DEV-002`, `AM-ADMIN-002`.
- **작업 순서**: (`AM-API-002` ∥ `AM-API-003` ∥ `AM-DIST-001`) → `AM-SEC-002` → (`AM-DEV-002` ∥ `AM-ADMIN-002`).
- **병렬 수행 가능성**: 카탈로그 API, 개발자 API, APK 분석 워커 3개 트랙 병렬 개발 가능.
- **필요한 인프라**: S3 격리 버킷(`bn-quarantine-bucket`), ClamAV/YARA 샌드박스 컨테이너(`internal-quarantine` 서브넷).
- **기능 완료 조건**: 개발자 KYB 승인 → 앱 생성 → APK 멀티파트 업로드 → SHA-256/AXML/`apksigner` v2/v3 검증 → 악성코드/위험권한 스캔 → 관리자 심사 승인/반려 전체 워크플로 동작.
- **테스트 완료 조건**: `TC-FUNC-002`, `TC-FUNC-003`, `TC-FUNC-007`~`009`, `TC-SEC-002`, `TC-SEC-007`, `TC-SEC-009` 통과.
- **위험 요인**: 악성 조작된 APK로 인한 파서 크래시 → 격리 워커 타임아웃 및 메모리 제한으로 코어 API 완벽 보호.
- **최종 산출물**: `backend/internal/modules/{catalog,developer,distribution,security}/` 및 개발자/관리자 심사 UI.
- **Phase Exit Gate**: 실제 서명된 APK 업로드부터 심사 승인까지 End-to-End 통과 및 변조/미서명/악성 APK 100% 차단 확인.

### Phase 4 — APK/AAB 글로벌 CDN 배포 및 Android 네이티브 클라이언트 설치·업데이트 완성
- **선행조건**: Phase 3 Exit Gate 통과.
- **작업 목록**: `AM-DIST-002`, `AM-DIST-003`, `AM-ANDROID-001`, `AM-ANDROID-002`, `AM-ANDROID-003`.
- **작업 순서**: (`AM-DIST-002` ∥ `AM-DIST-003` ∥ `AM-ANDROID-001`) → `AM-ANDROID-002` → `AM-ANDROID-003`.
- **병렬 수행 가능성**: 백엔드 CDN 배포/업데이트 리졸버와 Android UI 골격 병렬 개발 가능.
- **필요한 인프라**: S3 공개 버킷(`bn-public-cdn-bucket`), CDN Signed URL 발급기, `bundletool`, Android SDK 35 및 에뮬레이터.
- **기능 완료 조건**: 불변 경로(`/{sha256}.apk`) 배포, 15분 만료 Signed URL, Android 이어받기 다운로더, 로컬 SHA-256/서명 대조, `PackageInstaller.Session` 설치 및 일괄 업데이트, 마켓 자체 업데이트 동작.
- **테스트 완료 조건**: `TC-FUNC-012`, `TC-AND-001` ~ `TC-AND-003` 통과.
- **위험 요인**: Android 버전별 `PackageInstaller` 인텐트 제약 차이 → API 26~35 분기 처리로 해결.
- **최종 산출물**: `android-client/` 전체 모듈 및 `release.apk`, 백엔드 CDN 배포 엔진.
- **Phase Exit Gate**: Android 에뮬레이터에서 실제 앱 다운로드 → 설치 승인 → 앱 실행 → 상위 버전 업데이트 성공 `adb logcat` 증빙 확보.

### Phase 5 — 외부 앱 검색 커넥터 및 통합 비교 검색엔진 완성 (Type 2 & Type 3 Market)
- **선행조건**: Phase 3 카탈로그 엔진 완료.
- **작업 목록**: `AM-SEARCH-001`, `AM-SEARCH-002`, `AM-API-006`, `AM-WEB-002`.
- **작업 순서**: `AM-SEARCH-001` → (`AM-SEARCH-002` ∥ `AM-API-006` ∥ `AM-WEB-002`).
- **병렬 수행 가능성**: OpenSearch 인덱서 완료 후 외부 커넥터, 추천/차트 API, 웹 비교 UI 병렬 개발 가능.
- **필요한 인프라**: OpenSearch 2.17 클러스터(Nori/ICU 플러그인 포함).
- **기능 완료 조건**: 다국어 통합 검색, 50ms 자동완성, F-Droid `index-v2.json` 서명 검증 동기화, 5단계 권리 등급 강제, 배포처·가격·버전·서명 비교 매트릭스 UI 완성.
- **테스트 완료 조건**: `TC-FUNC-004`, `TC-FUNC-010`, `TC-SEC-003`, `TC-UI-003`, `TC-OPS-005` 통과.
- **위험 요인**: 외부 스토어 약관 위반 또는 변조 피드 유입 → 무단 APK 스크래핑 금지 및 GPG/JAR 서명 검증 강제.
- **최종 산출물**: `backend/internal/modules/{search,connector}/`, 사용자 웹 비교/검색 화면(`USR-02,03,14~17`).
- **Phase Exit Gate**: 자체 앱과 외부 앱이 단일 상세 페이지에서 비교되고, 외부 링크 전용 앱에 가짜 설치 버튼이 0개임을 검증.

### Phase 6 — 보안 심화, 리뷰 어뷰징 방어, 신고(DMCA), 이의신청 및 운영 감사 완성
- **선행조건**: Phase 4 및 Phase 5 완료.
- **작업 목록**: `AM-SEC-003`, `AM-MOD-001`, `AM-API-005`.
- **작업 순서**: 3개 Work Order 병렬 진행 가능.
- **병렬 수행 가능성**: KMS 암호화/Rate Limit(`AM-SEC-003`), 모더레이션/이의신청(`AM-MOD-001`), 알림/퍼널 수집(`AM-API-005`) 100% 병렬 수행.
- **필요한 인프라**: KMS Envelope Encryption, Redis Rate Limit 클러스터, SMTP/FCM 어댑터.
- **기능 완료 조건**: 설치 검증 사용자만 리뷰 작성, 봇 다운로드/평점 조작 차단, 신고 SLA 큐, 시니어 심사자 재배정 이의신청, 긴급 Kill Switch 기기 알림 연동.
- **테스트 완료 조건**: `TC-FUNC-011`, `TC-FUNC-013`, `TC-FUNC-014`, `TC-SEC-008`, `TC-SEC-010` 통과.
- **위험 요인**: 허위 신고를 통한 경쟁 앱 테러 → 자동 블라인드 금지 및 관리자 증빙 확인 후 조치.
- **최종 산출물**: `backend/internal/modules/moderation/`, `backend/internal/platform/crypto/`.
- **Phase Exit Gate**: 침투 테스트(IDOR, 토큰 탈취, 감사 체인 변조, 블랙리스트 인증서 업로드) 전 항목 방어 통과.

### Phase 7 — 유료 앱 결제, 복식부기 정산 원장 및 스폰서 광고 엔진 완성
- **선행조건**: Phase 1 원장 스키마 및 Phase 6 보안 모듈 완료.
- **작업 목록**: `AM-BILLING-001`, `AM-BILLING-002`, `AM-ADS-001`, `AM-WEB-003`, `AM-DEV-003`, `AM-ADMIN-003`.
- **작업 순서**: (`AM-BILLING-001` ∥ `AM-ADS-001`) → `AM-BILLING-002` → (`AM-WEB-003` ∥ `AM-DEV-003` ∥ `AM-ADMIN-003`).
- **병렬 수행 가능성**: 결제 원장 백엔드와 광고 경매 백엔드 병렬 개발 후 3개 프론트엔드 UI 병렬 연결.
- **필요한 인프라**: PG 샌드박스(Stripe/Toss), 복식부기 원장 검증 트리거.
- **기능 완료 조건**: 다통화 견적, 프로모션 리딤, PG 웹훅 검증, 구매 권한(`user_entitlements`) 연동 다운로드, 환불 및 권한 즉시 회수, 14일 에스크로 해제, 월간 정산(Dual-Control), 스폰서 광고 경매 및 부정클릭 필터링.
- **테스트 완료 조건**: `TC-COM-001` ~ `TC-COM-006` 전수 통과.
- **위험 요인**: 부동소수점 반올림 오차 또는 중복 웹훅 기장 → 정수 Minor Unit 연산 및 `Idempotency-Key` + DB 대차평형 트리거로 원천 차단.
- **최종 산출물**: `backend/internal/modules/{commerce,ads}/` 및 결제/정산/광고 UI.
- **Phase Exit Gate**: 1,000건 결제·환불·정산 시뮬레이션 후 `ledger_entries` 차변-대변 오차 `0` 확인.

### Phase 8 — 글로벌 운영 정책(Geo-Fencing), 성능 부하 검증, 다중 기기 매트릭스 및 DR 복구 검증
- **선행조건**: Phase 4 ~ Phase 7 완료.
- **작업 목록**: `AM-API-004`, `AM-ANDROID-004`, `AM-OPS-002`, `AM-OPS-003`.
- **작업 순서**: 4개 Work Order 완전 병렬 수행 가능.
- **병렬 수행 가능성**: Backend, Android, DevOps 에이전트가 독립 디렉터리에서 동시 수행.
- **필요한 인프라**: Android API 26~35 에뮬레이터 팜, DR 백업 복원용 격리 DB 인스턴스, CDN Edge Worker.
- **기능 완료 조건**: 국가별 법적 차단(`451`), 아랍어 RTL, API 26~35 전 버전 호환성 검증, 카나리 자동 롤백, PostgreSQL PITR 및 OpenSearch 백업·복원 드릴 완료.
- **테스트 완료 조건**: `TC-FUNC-005`, `TC-SEC-004`, `TC-AND-004`, `TC-OPS-002`, `TC-OPS-003` 통과.
- **위험 요인**: 특정 구형 Android API 레벨에서의 서명 추출 API 차이 → API 28 기준 `GET_SIGNING_CERTIFICATES` / `GET_SIGNATURES` 호환 래퍼로 해결.
- **최종 산출물**: `scripts/backup/dr_restore_drill.sh`, Android Connected Test 스위트, CDN/WAF 설정.
- **Phase Exit Gate**: DR 백업 복원 드릴 RTO < 30분 통과 및 Android API 26~35 전 버전 설치 성공.

### Phase 9 — 독립 상용화 최종 통합 검증 및 Independent Audit Agent 감사
- **선행조건**: Phase 0 ~ Phase 8의 모든 38개 Work Order `IMPLEMENTED_UNVERIFIED` 도달.
- **작업 목록**: `AM-TEST-001`, `AM-TEST-002`.
- **작업 순서**: `AM-TEST-001` (QA/E2E 에이전트 전체 회귀 및 k6 부하 테스트) → `AM-TEST-002` (Independent Audit Agent 독립 교차 감사).
- **병렬 수행 가능성**: 순차 실행 필수 (QA 완료 후 독립 감사가 최종 판정).
- **필요한 인프라**: Staging 클린 클러스터, k6 부하 생성기.
- **기능 완료 조건**: 신규 기능 추가 금지(Code Freeze), 발견된 결함 수정 및 전 요구사항 추적성 증빙 완비.
- **테스트 완료 조건**: `MASTER_TEST_MATRIX.csv`의 53개 전체 테스트 케이스 100% `PASS`.
- **위험 요인**: 미구현 기능을 Mock으로 숨긴 코드 적발 → 독립 감사 에이전트가 즉시 `AUDIT_REJECTED` 판정 후 재구현 명령.
- **최종 산출물**: `artifacts/evidence/AUDIT_REPORTS/independent_audit_verdict.json`.
- **Phase Exit Gate**: Independent Audit Agent의 최종 `VERIFIED` 서명 획득 (독립 상용 출시 가능 상태 달성).

### Phase 10 — TreeScan 연동 어댑터 준비 및 별도 통합 계획 검증 (TreeScan 본체 수정 없음)
- **선행조건**: Phase 9 독립 상용화 감사 통과.
- **작업 목록**: `AM-TREESCAN-001`.
- **작업 순서**: `AM-TREESCAN-001` 단독 실행 및 계약 검증.
- **병렬 수행 가능성**: 단일 모듈 작업.
- **필요한 인프라**: 앱 마켓 Staging 서버 + TreeScan 계약 테스트 스텁.
- **기능 완료 조건**: `GET /api/v1/treescan/federated-search`, `GET /api/v1/treescan/catalog-feed`, `POST /api/v1/auth/token/exchange` 구현 및 Feature Flag(`TREESCAN_INTEGRATION_ENABLED=false` 기본값) 검증.
- **테스트 완료 조건**: `TC-API-003`, `TC-OPS-008` 통과.
- **위험 요인**: TreeScan 본체 코드를 건드리는 실수 → 본 단계는 앱 마켓 측 어댑터만 구현하며 TreeScan 본체 배포는 향후 별도 PO 승인 후 진행함을 엄격히 강제.
- **최종 산출물**: `backend/internal/modules/treescan/` 및 통합 준비 완료 리포트.
- **Phase Exit Gate**: 기본 모드에서 100% 독립성 유지 확인 및 활성화 모드에서 어댑터 응답시간 `p95 < 80ms` 통과.

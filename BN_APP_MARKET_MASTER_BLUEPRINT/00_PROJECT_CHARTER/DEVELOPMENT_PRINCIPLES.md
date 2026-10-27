# DEVELOPMENT_PRINCIPLES.md — 절대 개발 원칙 및 상용 품질 기준

- **문서 ID**: `DOC-CHARTER-002`
- **관련 조항**: 제1조, 제2조, 제20조, 제23조

---

## 1. TreeScan 통합 이전 100% 독립 개발 원칙 (제1조 1.1항)

본 플랫폼은 장기적으로 Booking Network의 TreeScan 검색엔진 생태계와 결합될 예정이지만, **Phase 0부터 Phase 9까지의 모든 개발·검증·상용화 과정은 TreeScan과 완전히 분리된 독립 플랫폼으로 수행**합니다.

### 1.1 10대 독립성 충족 요건
1. **독립 실행 가능한 사용자 웹사이트 (`web-user`)**: 자체 도메인, 자체 SSR 렌더러, 자체 라우팅으로 동작하며 TreeScan 웹 프론트엔드 코드를 참조하지 않음.
2. **독립 실행 가능한 Android 앱 마켓 (`android-client`)**: 자체 패키지명(`com.bookingnetwork.appmarket`)과 서명키를 가진 독립 APK로 배포 및 자체 업데이트 수행.
3. **독립 개발자 콘솔 (`web-dev`)**: 개발자 온보딩, APK 업로드, 출시 관리, 수익 정산을 단독 수행.
4. **독립 관리자 콘솔 (`web-admin`)**: 앱 심사, 보안 관제, 신고 처리, 정산 승인을 단독 수행.
5. **독립 백엔드 API (`backend`)**: Go 기반 독자 API 서버로 동작하며 TreeScan 내부 RPC에 의존하지 않음.
6. **독립 데이터베이스 (`PostgreSQL 16`)**: 독자적인 데이터베이스 클러스터 및 마이그레이션 라이프사이클 유지.
7. **독립 검색엔진 (`OpenSearch 2.x`)**: 자체 인덱스(`bn_apps_v1`, `bn_autocomplete_v1`)와 자체 랭킹 파이프라인으로 검색 서비스 제공.
8. **독립 앱 등록 및 배포 시스템**: 자체 S3 격리/공개 버킷, 자체 서명/악성코드 분석 워커, 자체 CDN Signed URL 발급기로 동작.
9. **독립 인증·권한 시스템**: 자체 OIDC/JWT 발급 및 RBAC/MFA 검증 체계를 보유.
10. **독립 CI/CD 및 운영 모니터링**: 자체 Git 파이프라인, 자체 Docker 레지스트리, 자체 OpenTelemetry/Prometheus/Grafana 스택으로 관제.

### 1.2 향후 TreeScan 연동을 위한 경계 설계
- TreeScan 본체 코드는 건드리지 않으며, 오직 `/api/v1/treescan/*` 어댑터 엔드포인트([04_API_SPECIFICATION/openapi.yaml](../04_API_SPECIFICATION/openapi.yaml) 및 [15_TREESCAN_FUTURE_INTEGRATION/](../15_TREESCAN_FUTURE_INTEGRATION/))와 OIDC Token Exchange 인터페이스만 표준 계약으로 준비합니다.

---

## 2. 설계(Documentation)와 실제 구현(Implementation)의 엄격한 구분 (제1조 1.2항)

1. 본 패키지(`BN_APP_MARKET_MASTER_BLUEPRINT_V1_0.zip`)는 **마스터 설계 패키지**입니다.
2. 설계 문서가 작성되었다고 해서 실제 기능이 구현(`IMPLEMENTED`)되었거나 테스트가 통과(`PASS`)된 것으로 표기하는 행위는 엄격히 금지됩니다.
3. 본 패키지 내 모든 요구사항과 Work Order의 초기 구현 상태는 `NOT_STARTED`이며, 설계 검증 상태만 `DESIGN_VERIFIED`로 기록됩니다.
4. 후속 개발 에이전트가 실제로 Go/Vue/Kotlin 코드를 작성하고 테스트를 통과시킨 후에만 해당 Work Order의 상태를 `IMPLEMENTED_UNVERIFIED` → `VERIFIED`로 갱신할 수 있습니다.

---

## 3. 상용 수준(Commercial-Grade)의 정의 및 14대 필수 요건 (제1조 1.3항)

단순히 화면 레이아웃이 렌더링되거나 더미 JSON을 반환하는 프로토타입은 상용 수준으로 인정하지 않습니다. 모든 기능은 다음 14대 요건의 **구현 방법**과 **검증 방법**을 동시에 갖추어야 합니다:

| 번호 | 상용 기준 항목 | 구체적 구현 기준 | 완료 검증(Evidence) 기준 |
| :---: | :--- | :--- | :--- |
| 1 | **실제 데이터 저장** | PostgreSQL 트랜잭션, 외래키 제약조건, 낙관적 잠금(`version`), 마이그레이션 적용 | DB 쿼리 결과 스냅샷, 트랜잭션 롤백 테스트 로그 |
| 2 | **실제 인증 및 접근 통제** | Argon2id 패스워드 해싱, RS256 JWT + Refresh Token Rotation, TOTP MFA, 세분화된 RBAC | 권한 없는 역할의 API 호출 시 `403 FORBIDDEN` 통합 테스트 로그 |
| 3 | **실제 앱 업로드** | 대용량 바이너리 스트리밍 업로드, 격리 스토리지 저장, 파일 크기/확장자/MIME 검증 | 실제 APK 업로드 후 S3 격리 버킷 오브젝트 메타데이터 확인 |
| 4 | **실제 파일 검증** | APK AXML 파싱, `apksigner` v2/v3 서명 블록 검증, SHA-256 체크섬 대조, 악성코드 스캔 | 변조된 APK 업로드 시 `ERR_APK_SIGNATURE_INVALID` 차단 로그 |
| 5 | **실제 앱 심사** | 자동 보안 검사 통과 후 관리자 콘솔에서 심사자 할당, 체크리스트 판정, 승인/반려 상태 전이 | `app_reviews` 및 `audit_logs` 테이블 상태 전이 기록 |
| 6 | **실제 앱 다운로드** | 출시 승인된 파일만 공개 배포 버킷으로 복사 후 만료 시간 및 IP/서명이 포함된 CDN URL 발급 | 만료된 서명 URL 접근 시 `403` 차단 및 정상 URL 바이트 스트리밍 검증 |
| 7 | **실제 Android 설치** | Android `PackageInstaller.Session` 생성, 스트림 기록, `fsync`, 사용자 승인 `PendingIntent` 처리 | Android API 26~35 기기/에뮬레이터에서 `STATUS_SUCCESS` 브로드캐스트 수신 로그 |
| 8 | **실제 업데이트 호환성 확인** | 기설치 앱의 `packageName`, `versionCode`, `signingInfo`(인증서 SHA-256)와 서버 메타데이터 비교 | 서명 불일치 업데이트 시도 시 다운로드 전 UI 경고 및 설치 차단 검증 |
| 9 | **실제 검색 결과** | OpenSearch 인덱스 실시간/배치 동기화, 한국어/다국어 형태소 분석, 필터링 및 정렬 | OpenSearch 쿼리 DSL 응답 및 색인 반영 지연시간(< 5초) 측정 로그 |
| 10 | **실제 관리자 운영** | 긴급 배포 중단(Kill Switch), 신고 처리, 개발자 정산 승인, 커넥터 제어 | 관리자 액션 즉시 사용자 웹/앱 노출 차단 확인 E2E 테스트 |
| 11 | **장애 및 보안 대응** | 서킷 브레이커, Dead Letter Queue(DLQ) 재처리, Rate Limiting, WAF 룰 | Redis/OpenSearch 일시 중단 시 Graceful Degradation 동작 로그 |
| 12 | **운영 데이터 감사** | 모든 민감 작업에 대해 변경 전/후 JSON Diff, 수행자 IP, 타임스탬프, 해시 체인 기록 | `audit_logs` 테이블 불변성(UPDATE/DELETE 권한 차단) 검증 |
| 13 | **권한 분리** | 개발자 조직 간 테넌트 격리(IDOR 원천 차단), 관리자 직무 분리(심사자 vs 정산 승인자) | 타 개발자 `app_id` 수정 시도 시 `403/404` 차단 침투 테스트 |
| 14 | **확장 가능한 시스템 구성** | 무상태(Stateless) API 서버, 격리된 비동기 워커 오토스케일링, 읽기 복제본 분리 | 컨테이너 스케일아웃 및 무중단 롤링 업데이트 검증 |

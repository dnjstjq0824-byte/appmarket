# ACCEPTANCE_CRITERIA.md — 플랫폼 상용화 최종 수용 기준 (Master Acceptance Criteria)

- **문서 ID**: `DOC-QA-001`
- **관련 조항**: 제1조 1.3항, 제18조, 제22조, 제23조
- **테스트 매트릭스**: [MASTER_TEST_MATRIX.csv](MASTER_TEST_MATRIX.csv)

---

## 1. 10대 상용 수용 게이트 (Commercial Acceptance Gates)

후속 개발 에이전트와 `Independent Audit Agent`는 아래 10대 수용 기준이 **실제 실행 증빙과 함께 100% 충족되기 전까지 프로젝트 전체 완료를 선언할 수 없습니다**:

1. **Gate 1 (TreeScan 무의존 100% 독립 운영)**: TreeScan 관련 코드·서버·환경변수가 전혀 없는 상태에서 사용자 웹, 개발자 콘솔, 관리자 콘솔, Android 앱, 백엔드 API, DB, OpenSearch, S3/CDN이 단독으로 완벽히 구동되어야 한다 (`TC-OPS-008`).
2. **Gate 2 (3대 동일성 분리 및 정직한 출처/설치 UI)**: 카탈로그 동일성, 배포처 동일성, 서명 동일성이 DB와 API에서 분리되어 동작하며, 외부 링크 전용(`EXTERNAL_LINK_ONLY`) 앱에는 가짜 설치 버튼이 단 1개도 존재하지 않아야 한다 (`TC-DB-001`, `TC-UI-003`).
3. **Gate 3 (실제 APK/AAB 보안 검증 파이프라인)**: 업로드된 APK가 격리 버킷에 저장되고, SHA-256, AXML 매니페스트, `apksigner` v2/v3 서명 연속성, ClamAV/YARA 악성코드 및 위험 권한 스캔을 모두 통과해야만 공개 CDN 버킷(`/{sha256}.apk`)으로 배포되어야 한다 (`TC-FUNC-009`, `TC-SEC-007`, `TC-SEC-009`).
4. **Gate 4 (실제 Android OS `PackageInstaller` 설치 및 업데이트)**: Android API 26~35 환경에서 HTTP Range 이어받기, 로컬 SHA-256 및 서명 인증서 대조, `PackageInstaller.Session` 커밋, 사용자 승인 인텐트, 서명 불일치 업데이트 차단이 실제로 동작해야 한다 (`TC-AND-001` ~ `TC-AND-004`).
5. **Gate 5 (통합 다국어 검색 및 외부 커넥터 서명 검증)**: 한국어(Nori)/다국어 OpenSearch 검색과 50ms 이내 자동완성이 동작하며, F-Droid 등 외부 피드 동기화 시 서명 변조 피드가 100% 거부되어야 한다 (`TC-FUNC-010`, `TC-SEC-003`).
6. **Gate 6 (복식부기 상거래 원장 무결성)**: 모든 유료 앱 결제·환불·에스크로 해제·월간 정산 트랜잭션 후 `ledger_entries`의 차변과 대변 합계가 단 1원의 오차 없이 일치(`SUM = 0`)해야 하며, 환불 즉시 다운로드 권한이 회수되어야 한다 (`TC-COM-001` ~ `TC-COM-006`).
7. **Gate 7 (스폰서 광고 투명성 및 부정클릭 방어)**: 모든 광고는 `'Sponsored'` 배지를 포함하며, 노출 토큰이 없거나 동일 기기에서 반복된 부정클릭은 과금되지 않고 일일 예산 상한을 절대 초과하지 않아야 한다 (`TC-COM-003`).
8. **Gate 8 (최소 권한 RBAC, MFA, 이중 승인 및 불변 감사 로그)**: 개발자 조직 간 IDOR가 100% 차단되고, 민감 관리자 작업은 2인 이중 승인(`Dual-Control`)이 강제되며, 해시 체인 감사 로그 변조가 즉시 탐지되어야 한다 (`TC-SEC-001`, `TC-SEC-002`, `TC-SEC-006`).
9. **Gate 9 (76개 전체 화면 반응형 및 접근성)**: 사용자 웹(20), 개발자 콘솔(24), 관리자 콘솔(20), Android(12) 전 화면이 5대 뷰 상태와 WCAG 2.1 AA 기준을 충족해야 한다 (`TC-UI-001`, `TC-UI-002`).
10. **Gate 10 (고가용성, 장애 격리 및 DR 복원 검증)**: 워커 크래시나 OpenSearch 중단 시에도 코어 서비스가 중단되지 않으며, DB 백업 복원 드릴이 데이터 유실 0건으로 통과해야 한다 (`TC-OPS-001`, `TC-OPS-002`, `TC-OPS-005`).

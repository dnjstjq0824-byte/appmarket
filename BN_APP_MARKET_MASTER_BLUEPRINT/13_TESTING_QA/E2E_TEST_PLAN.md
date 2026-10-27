# E2E_TEST_PLAN.md — 종단간(End-to-End) 시나리오 테스트 및 실기기 검증 계획

- **문서 ID**: `DOC-QA-002`
- **관련 문서**: [MASTER_TEST_MATRIX.csv](MASTER_TEST_MATRIX.csv)

---

## 1. 5대 핵심 골든 패스(Golden Path) E2E 시나리오

### E2E-01: 자체 앱 등록부터 Android 실기기 설치·업데이트까지 전체 라이프사이클
1. 개발자가 `web-dev`에서 가입 → TOTP MFA 설정 → 조직 생성 → KYB 서류 제출.
2. 관리자가 `web-admin`(`ADM-03`)에서 KYB 승인.
3. 개발자가 신규 앱(`com.example.goldenapp`) 생성 → v1(`versionCode=100`, `Key A` 서명) APK 업로드.
4. `worker-apk-analyzer` 및 `worker-malware-scanner`가 자동 검증 완료(`VERIFIED_READY`) → 개발자가 프로덕션 심사 제출.
5. 관리자가 `ADM-04`/`ADM-05`/`ADM-06`에서 체크리스트 확인 후 승인 → 공개 CDN 배포(`PUBLISHED_FULL`) 및 OpenSearch 색인 반영.
6. 사용자가 `android-client`에서 `"goldenapp"` 검색 → 다운로드 → SHA-256/서명 검증 → `PackageInstaller` 승인 및 설치 성공(`INSTALL_SUCCEEDED`) → 설치 검증 리뷰 작성.
7. 개발자가 v2(`versionCode=101`, 동일 `Key A` 서명) 출시 → Android 클라이언트의 `UpdatesCenterScreen`에서 업데이트 감지 및 덮어쓰기 업데이트 성공.

### E2E-02: 외부 스토어(F-Droid + 공식 링크) 통합 비교 및 서명 충돌 방어
1. 관리자가 `ADM-10`에서 F-Droid 커넥터 동기화 실행 → 서명 검증 통과 및 카탈로그 매핑.
2. 사용자가 `web-user` 상세 화면에서 `[BN 직접 다운로드]` 탭과 `[F-Droid 빌드]` 탭, `[Google Play 공식 링크]` 탭의 버전·가격·서명 지문 비교.
3. 서명이 다른 배포처로 교차 업데이트 시도 시 `android-client`가 `SignatureConflictDialog`를 띄워 데이터 손실 없이 사전 경고.

### E2E-03: 유료 앱 구매, 복식부기 원장 기장, 환불 및 권한 회수
1. 사용자가 유료 앱(₩11,000)에 10% 프로모션 코드를 적용해 결제 완료 → PG 웹훅 수신 → 복식부기 원장 4줄 분개(`SUM=0`) 및 `user_entitlements`(`ACTIVE`) 생성 → APK 다운로드 성공.
2. 환불 요청 및 승인 → 역분개 기장 및 `user_entitlements`(`REVOKED_REFUNDED`) 즉시 전환 → 다운로드 토큰 재요청 시 `403 ERR_ENTITLEMENT_REQUIRED` 차단.

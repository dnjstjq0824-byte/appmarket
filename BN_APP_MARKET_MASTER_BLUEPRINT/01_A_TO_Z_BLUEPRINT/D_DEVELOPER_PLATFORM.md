# D — Developer Platform — 개발자 온보딩, 법인/개인 KYB 인증, 조직 RBAC 및 셀프서비스 출시 플랫폼 (Developer Onboarding, KYB Verification, Organization RBAC & Self-Service Publishing Console)

- **문서 ID**: `DOC-AZ-D`
- **알파벳 코드**: `D`
- **관련 요구사항 ID**: `REQ-D-001`, `REQ-D-002`, `REQ-D-003`
- **연관 Work Order ID**: `AM-DEV-001`, `AM-DEV-002`, `AM-API-003`
- **연관 테스트 케이스 ID**: `TC-FUNC-003`, `TC-SEC-002`
- **참조 상세 설계 문서**: [05_FRONTEND_UXUI/DEVELOPER_CONSOLE.md](../05_FRONTEND_UXUI/DEVELOPER_CONSOLE.md), [09_SECURITY_COMPLIANCE/DEVELOPER_VERIFICATION.md](../09_SECURITY_COMPLIANCE/DEVELOPER_VERIFICATION.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 개발자가 관리자의 수작업 개입 없이도 스스로 계정을 생성하고, 신원(개인/법인) 검증을 완료하며, 팀원을 초대해 역할을 나누고, 앱 등록부터 APK 업로드·심사 제출·단계적 출시·리뷰 응대·정산 조회까지 완결할 수 있는 셀프서비스 개발자 플랫폼(제6조)을 제공하기 위해 필요하다.
- **달성 목표**: 테넌트 격리가 보장된 개발자 조직(`developer_orgs`) 구조, 개발자 신원 검증(KYC/KYB) 워크플로, 패키지명 소유권 증명, CI/CD 연동용 스코프 제한 API Key 관리 체계를 완성한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **Organization Owner (조직 소유자)**: 법인/개인 신원 인증 제출, 정산 계좌 등록, 앱 삭제/소유권 이전, 전체 팀원 관리(MFA 필수).
- **Release Manager (출시 관리자)**: APK/AAB 업로드, 트랙별(`INTERNAL`, `BETA`, `PRODUCTION`) 릴리스 생성, 심사 제출, 단계적 배포 비율 조정, 배포 중단.
- **Store Listing Editor (스토어 편집자)**: 앱 설명, 다국어 번역, 스크린샷/아이콘 편집(바이너리 배포 권한 없음).
- **Finance Manager (재무 담당자)**: 유료 앱 가격 설정, 주문/환불 내역 조회, 월별 정산 명세서 및 세금계산서 다운로드.
- **Support Agent (고객지원 담당자)**: 사용자 리뷰 조회 및 공식 답변 작성/수정.

## 3. 필수 기능 (Mandatory Features)
1. **개발자 신원 검증 (KYC/KYB)**: 개인 개발자(정부 발급 신분증 + 본인 명의 연락처 + 도메인/이메일 인증) 및 법인 개발자(사업자등록증/D-U-N-S 번호 + 법인 대표자 확인 + 법인 계좌 인증) 심사 프로세스.
2. **조직 및 팀원 RBAC**: 하나의 사용자 계정이 여러 개발자 조직에 서로 다른 역할로 소속될 수 있는 멀티테넌트 구조.
3. **셀프서비스 앱 라이프사이클 관리**: 앱 생성 → 패키지명 예약 → APK 업로드 → 서명 인증서 등록 → 메타데이터 입력 → 심사 제출 → 반려 사유 확인 및 재제출 → 출시/단계적 배포.
4. **CI/CD 자동화용 Developer API Key**: 특정 앱과 특정 권한(`release:upload`, `release:submit`)으로만 범위가 제한된 장기 토큰(SHA-256 해시 저장, 만료일 강제, 즉시 폐기 지원).
5. **리뷰 응대 및 분석 대시보드**: 일별 다운로드/설치/유지율/크래시 지표 및 리뷰 답변 이력 관리.

## 4. 상세 기능 (Detailed Functional Specifications)
- **패키지명 소유권 선점 방지 및 도메인 역방향 DNS 검증**:
  - `com.company.app` 형태의 패키지명 등록 시, 유명 브랜드 사칭을 막기 위해 예약어/보호 브랜드 목록(`protected_package_prefixes`)과 대조한다.
  - 엔터프라이즈/검증 배지(`Verified Developer`)를 원하는 개발자는 역방향 도메인(`company.com`)의 DNS에 `TXT _bn-appmarket-verify="org_id=..."` 레코드를 등록하면 백엔드 워커가 주기적으로 DNS 조회로 소유권을 자동 검증한다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **`web-dev` (24개 주요 뷰 컴포넌트)**:
  - 온보딩/KYB 위저드(`KybWizardView.vue`), 조직/팀원 관리(`OrgMembersView.vue`), 앱 목록(`AppListView.vue`), 앱 생성 모달(`CreateAppModal.vue`), APK/AAB 드래그앤드롭 업로더(`ArtifactUploaderView.vue`), 서명 인증서 뷰어(`SigningLineageView.vue`), 심사 상태 및 반려 피드백 뷰(`ReviewStatusView.vue`), 단계적 배포 슬라이더(`RolloutController.vue`), 리뷰 관리(`ReviewReplyView.vue`), 수익/정산 대시보드(`FinanceDashboardView.vue`), API 키 관리(`ApiKeyManagerView.vue`). 상세 명세는 [05_FRONTEND_UXUI/DEVELOPER_CONSOLE.md](../05_FRONTEND_UXUI/DEVELOPER_CONSOLE.md) 참조.

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/developer/`:
  - 모든 Developer API 요청은 URL 경로의 `{orgId}` 또는 `{appId}`에 대해 현재 로그인한 사용자의 `developer_org_members` 소속 여부와 역할 권한을 검증하는 `RequireOrgPermission(perm)` 미들웨어를 통과해야 한다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**: `developer_orgs`, `developer_verifications`, `developer_org_members`, `developer_api_keys`, `developer_payout_accounts`.
- `developer_orgs.verification_status` ENUM: `'UNVERIFIED'`, `'PENDING_REVIEW'`, `'ACTION_REQUIRED'`, `'VERIFIED'`, `'REJECTED'`, `'SUSPENDED'`.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `POST /api/v1/developer/orgs`: 개발자 조직 생성.
- `POST /api/v1/developer/orgs/{orgId}/verification`: KYB 인증 서류 및 도메인 검증 요청 제출.
- `POST /api/v1/developer/orgs/{orgId}/members`: 팀원 초대 및 역할 부여.
- `POST /api/v1/developer/orgs/{orgId}/apps`: 신규 앱 슬롯 생성 및 패키지명 예약.
- `POST /api/v1/developer/orgs/{orgId}/api-keys`: CI/CD API 키 발급(원문은 발급 시 단 1회만 반환).

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **개발자 조직 상태 전이**:
  - `UNVERIFIED` 상태에서는 앱 초안 작성 및 내부 테스트 트랙(`INTERNAL`) 업로드만 가능하며, 프로덕션 심사 제출(`SUBMITTED_FOR_REVIEW`)은 반드시 조직 상태가 `VERIFIED`이고 2단계 인증(MFA)이 활성화된 경우에만 허용된다.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- **테넌트 간 완벽한 격리 (Anti-IDOR)**: 조직 A의 Release Manager가 조직 B의 `appId`로 API를 호출할 경우 리소스 존재 여부조차 노출하지 않기 위해 `404 Not Found` 또는 `403 Forbidden`을 일관되게 반환하고 보안 감사 로그에 기록한다.
- KYB 신분증/사업자등록증 첨부파일은 일반 S3 버킷이 아닌 서버 측 KMS 암호화(`SSE-KMS`)가 적용된 `bn-kyb-vault-bucket`에 저장하며, 인가된 `KYB_REVIEWER` 관리자만 5분 만료 일회성 뷰어로 열람할 수 있다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 앱 심사 반려 시 단순 `"반려됨"` 표시가 아니라, 위반한 정책 조항 코드(예: `POL-SEC-004: 과도한 백그라운드 위치 권한`), 문제 화면/바이너리 위치, 수정 가이드라인, 그리고 재제출 버튼을 한 화면에서 제공한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **조직 간 IDOR 침투 테스트 (`TC-SEC-002`)**: 조직 A의 JWT 및 API Key로 조직 B의 앱 메타데이터 수정, APK 업로드, 정산 조회 시도 시 100% 차단되는지 자동화 검증.
- **KYB 미인증 출시 차단 테스트 (`TC-FUNC-003`)**: `UNVERIFIED` 조직이 프로덕션 심사 제출 API 호출 시 `422 ERR_DEV_ORG_NOT_VERIFIED`가 반환되는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `I — Identity & Access`(인증/MFA), `L — Lifecycle & Release Management`(출시 트랙), `N — Native App Distribution`(바이너리 업로드), `M — Marketplace Commerce`(정산)와 긴밀히 연동된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0 (독립 앱스토어 공급자 핵심)`
- **개발 Phase**: `Phase 2 (개발자 콘솔 UI)` 및 `Phase 3 (앱 등록 및 심사 워크플로)`에서 구현한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: CI/CD API Key 유출 사고** → 개발자 콘솔 및 관리자 콘솔에서 즉시 키 폐기(`Revoke`) 버튼을 제공하며, 폐기 즉시 Redis 블랙리스트에 반영되어 1초 이내에 해당 키로 들어오는 모든 업로드 요청을 차단한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 개발자 조직/인증/API 키 백엔드 모듈 (`backend/internal/modules/developer/*`).
- 개발자 콘솔 프론트엔드 애플리케이션 (`web-dev/`).

## 17. Acceptance Criteria (수용 기준)
1. 개발자가 관리자 DB 수동 조작 없이 웹 콘솔만으로 조직 생성 → KYB 제출 →(관리자 승인 후)→ 앱 등록 → APK 업로드 → 심사 제출의 전 과정을 수행할 수 있어야 한다.
2. 역할별 권한 매트릭스(`Store Listing Editor`는 APK 배포 불가, `Support Agent`는 가격 변경 불가)가 API 레벨에서 100% 강제되어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-DEV-001/rbac_matrix_test.log`: 5개 개발자 역할별 허용/차단 API 통합 테스트 결과.
- `artifacts/evidence/AM-DEV-002/e2e_developer_onboarding.mp4` (또는 Playwright trace): 개발자 가입부터 심사 제출까지의 E2E 실행 증빙.

# W — Workflow & Administration — 관리자 심사 콘솔(20개 운영 영역), 승인 워크플로, 이의신청 및 운영 거버넌스 (Admin Operations Console (20 Operational Domains), Review Workflows, Appeals & Governance)

- **문서 ID**: `DOC-AZ-W`
- **알파벳 코드**: `W`
- **관련 요구사항 ID**: `REQ-W-001`, `REQ-W-002`, `REQ-W-003`
- **연관 Work Order ID**: `AM-ADMIN-001`, `AM-ADMIN-002`, `AM-ADMIN-003`
- **연관 테스트 케이스 ID**: `TC-FUNC-013`, `TC-SEC-006`
- **참조 상세 설계 문서**: [05_FRONTEND_UXUI/ADMIN_CONSOLE.md](../05_FRONTEND_UXUI/ADMIN_CONSOLE.md), [04_API_SPECIFICATION/AUTHORIZATION_MATRIX.md](../04_API_SPECIFICATION/AUTHORIZATION_MATRIX.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 제7조에서 명령한 바와 같이, 관리자 웹(`web-admin`)은 단순 게시글 수정용 CMS가 아니라 **개발자 신원 검증, 앱/APK 보안 심사, 긴급 배포 중단, 신고/이의신청 처리, 외부 커넥터 제어, 검색 랭킹 운영, 광고 심사, 결제·환불·정산 승인, 감사 로그 및 백업 점검**까지 총 20개 핵심 운영 영역을 최소 권한 분리와 이중 승인(Dual-Control) 기반으로 통제하는 상용 앱스토어 운영 시스템이어야 한다.
- **달성 목표**: 제7조의 20개 관리자 운영 영역 각각의 화면·API·최소 권한 매핑, SLA 기반 심사 할당 큐, 개발자 이의신청(Appeal) 재심 워크플로, 민감 행동 감사 및 이중 승인 절차를 명세한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **`ADMIN_SUPER`**: 시스템 설정, 관리자 역할 부여(Dual-Control 필수), 긴급 복구.
- **`ADMIN_KYB_REVIEWER`**: 개발자 개인/법인 신원 검증 서류 심사.
- **`ADMIN_APP_REVIEWER`**: 신규 앱 및 업데이트 메타데이터/콘텐츠/기능 심사.
- **`ADMIN_SECURITY`**: 고위험 APK 정적 분석 리포트 검토, 악성코드 검역 해제/영구 차단, 긴급 Kill Switch 발동.
- **`ADMIN_MODERATOR`**: 사용자 신고, 불법 리뷰 블라인드, 개발자 이의신청 1차 검토.
- **`ADMIN_SEARCH_OPS`**: 외부 스토어 커넥터 운영, 검색 동의어/부스팅/에디터 컬렉션 관리.
- **`ADMIN_FINANCE`**: 이상 결제 조사, 분쟁 환불 강제 집행, 개발자 정산 배치 승인(Dual-Control 필수), 광고 정산.
- **`ADMIN_AUDITOR` (읽기 전용 감사역)**: 모든 운영 영역 및 해시 체인 감사 로그 열람 전용(수정 권한 0개).

## 3. 필수 기능 (Mandatory Features)
1. **제7조 필수 20대 관리자 운영 영역 전수 분리 구현**:
   - `ADM-01` 전체 운영 대시보드 (실시간 트래픽, 심사 대기열 SLA, 보안 경보, 일간 매출)
   - `ADM-02` 회원 및 계정 관리 (사용자 조회, 비정상 계정 잠금/해제, 세션 강제 종료)
   - `ADM-03` 개발자 신원 검증 (개인/법인 KYB 서류 열람 및 승인/반려)
   - `ADM-04` 앱 등록 및 메타데이터 심사 (설명, 스크린샷, 연령 등급, 정책 체크리스트 심사)
   - `ADM-05` APK 보안 심사 (YARA/ClamAV 결과, 위험 권한 Diff, 서명 지문 검토)
   - `ADM-06` 앱 출시 최종 승인 (심사 완료 릴리스의 배포 승인 및 보류 해제)
   - `ADM-07` 앱 배포 중단 및 긴급 차단 (Kill Switch, 버전 회수, 인증서 블랙리스트)
   - `ADM-08` 신고 및 사용자 보호 (악성/사칭/저작권 침해 신고 SLA 큐 및 제재)
   - `ADM-09` 리뷰·평점 운영 (어뷰징 의심 리뷰 일괄 격리/복원, 개발자 악성 답글 제재)
   - `ADM-10` 외부 앱 검색 커넥터 관리 (F-Droid 등 커넥터 스케줄, 수동 동기화, 에러 로그)
   - `ADM-11` 검색 랭킹 및 큐레이션 운영 (동의어 사전, 금칙어, 에디터 추천 컬렉션)
   - `ADM-12` 광고 심사 및 캠페인 관제 (스폰서 앱 광고 소재 심사, 부정클릭 탐지 현황)
   - `ADM-13` 결제·환불 분쟁 관리 (주문/결제 원장 조회, 강제 환불, 차지백 대응)
   - `ADM-14` 개발자 정산 관리 (월간 정산 배치 검토, 지급 보류/해제, 이중 승인 집행)
   - `ADM-15` 통계·분석 센터 (전환율 퍼널, 코호트 유지율, 배포처별 클릭/설치 지표)
   - `ADM-16` 다국어·국가별 정책 관리 (국가별 법적 차단 앱, 부가세율, 성인인증 의무 설정)
   - `ADM-17` 관리자 권한 및 직무 분리 관리 (RBAC 할당, MFA 상태, 이중 승인 대기열)
   - `ADM-18` 시스템 설정 및 Feature Flag (수수료 기본값, 업로드 용량 제한, 점검 모드 토글)
   - `ADM-19` 모니터링 및 불변 감사 로그 (워커 큐/DLQ 관제, 해시 체인 감사 로그 검색/검증)
   - `ADM-20` 백업·복구 및 운영 점검 (DB/OpenSearch 스냅샷 상태 조회, 무결성 점검 실행).
2. **이의신청(Appeals) 공식 워크플로**:
   - 앱 반려, 배포 중단, 계정 정지 처분을 받은 개발자는 14일 이내에 증빙 자료와 소명서를 첨부하여 이의신청(`POST /api/v1/developer/appeals`)을 제출할 수 있으며, 공정성을 위해 **최초 반려/제재를 수행한 심사자와 다른 상위 심사자(`reviewer_id != initial_actor_id`)**에게 자동 배정된다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **심사 체크리스트 강제화 (`ReviewChecklistSchema`)**:
  - 심사자가 `APPROVE` 버튼을 누르려면 반드시 ① 메타데이터-실제 앱 기능 일치 확인, ② 개인정보처리방침 유효 URL 확인, ③ 고위험 권한 타당성 확인, ④ 연령 등급 적정성 확인 항목을 모두 체크해야만 API 요청이 활성화된다.
  - `REJECT` 또는 `CHANGES_REQUESTED` 선택 시에는 표준 정책 위반 코드(`policy_violation_code`)와 구체적 수정 가이드 텍스트(최소 20자 이상) 입력을 강제한다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **`web-admin` (20개 전용 뷰 컴포넌트)**: 상세 레이아웃 및 컴포넌트 구조는 [05_FRONTEND_UXUI/ADMIN_CONSOLE.md](../05_FRONTEND_UXUI/ADMIN_CONSOLE.md)에 전수 정의되어 있다.

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/admin/`: 모든 관리자 변이(Mutation) API는 사유(`reason` 필드, 최소 10자) 입력을 필수로 요구하며, 실행 전후 스냅샷을 `audit_logs`에 자동 기록한다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**: `release_reviews`, `moderation_reports`, `moderation_actions`, `appeals`, `dual_approval_requests`, `audit_logs`, `system_settings`.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `/api/v1/admin/*` 하위 35개 엔드포인트 ([04_API_SPECIFICATION/openapi.yaml](../04_API_SPECIFICATION/openapi.yaml) 및 [04_API_SPECIFICATION/AUTHORIZATION_MATRIX.md](../04_API_SPECIFICATION/AUTHORIZATION_MATRIX.md) 참조).

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **이의신청 상태 (`appeal_status`)**: `SUBMITTED` → `ASSIGNED_TO_SENIOR_REVIEWER` → `UNDER_REVIEW` → `UPHELD_OVERTURNED` (원처분 취소 및 앱/계정 복구) | `DENIED_FINAL` (기각 확정).

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 관리자 콘솔은 일반 사용자 웹과 도메인/쿠키 스코프를 완전히 분리(`admin.internal.appmarket...`)하고, 세션 유휴 타임아웃(Idle Timeout)을 15분으로 제한하며, 모든 민감 액션 시 TOTP 재확인(`Step-up MFA`)을 요구한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 고밀도 운영 효율을 위해 단축키(`j`/`k` 행 이동, `a` 승인 모달, `r` 반려 모달, `/` 전역 검색)와 변경점 하이라이트 Diff 뷰어를 기본 제공한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **관리자 최소 권한 및 이의신청 재배정 테스트 (`TC-FUNC-013`)**: `ADMIN_SEARCH_OPS`가 정산 승인 API 호출 시 `403` 차단 확인, 그리고 심사자 A가 반려한 앱의 이의신청을 심사자 A 본인이 승인 시도할 때 `403 ERR_APPEAL_SAME_REVIEWER_FORBIDDEN`으로 차단되는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- 모든 도메인 모듈의 운영/관제 평면(Control Plane)으로 동작한다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0`
- **개발 Phase**: 기본 콘솔 및 심사 기능은 `Phase 2~3`, 보안·신고·이의신청·감사는 `Phase 6`, 정산·광고 운영은 `Phase 7`에서 완성한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 특정 관리자가 심사 대기열 아이템을 열어둔 채 자리를 비워 심사가 지연되는 현상(Stale Claim Lock)** → 심사 할당 잠금(`claimed_until`)을 30분으로 제한하고 만료 시 자동으로 대기열 공용 풀로 반환한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 관리자 백엔드 모듈(`backend/internal/modules/admin/*`) 및 관리자 콘솔 프론트엔드(`web-admin/`).

## 17. Acceptance Criteria (수용 기준)
1. 제7조의 20개 운영 영역이 모두 독립된 라우트와 최소 권한 가드로 동작해야 한다.
2. 사유(`reason`)가 누락된 관리자 상태 변경 요청은 `400 ERR_AUDIT_REASON_REQUIRED`로 100% 거부되어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-ADMIN-001/admin_20_modules_e2e.json`: 20개 관리자 화면 및 권한 분리 통합 검증 리포트.

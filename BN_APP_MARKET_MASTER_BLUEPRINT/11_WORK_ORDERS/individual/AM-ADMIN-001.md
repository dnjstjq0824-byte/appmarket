# AM-ADMIN-001 — 관리자 콘솔(Vue 3) 프레임워크, 대시보드, 회원/KYB 심사, 권한/이중승인 및 감사 로그 화면(ADM-01~03,16~20) 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-ADMIN-001` |
| **작업명 (Title)** | 관리자 콘솔(Vue 3) 프레임워크, 대시보드, 회원/KYB 심사, 권한/이중승인 및 감사 로그 화면(ADM-01~03,16~20) 구현 |
| **목표 (Objective)** | `REQ-W-001;REQ-W-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Frontend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-ARCH-001;AM-SEC-001` |
| **개발 Phase 및 우선순위** | `Phase 2` / `P0` |
| **입력 문서 (Input Documents)** | `05_FRONTEND_UXUI/ADMIN_CONSOLE.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `apps/web-admin/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `backend/;android-client/;apps/web-user/` |
| **예상 생성/수정 파일 (Target Files)** | `apps/web-admin/src/views/AdminOverviewDashboard.vue;apps/web-admin/src/views/DeveloperKybReviewView.vue;apps/web-admin/src/views/AdminRbacDualControlView.vue;apps/web-admin/src/views/AuditLogChainExplorerView.vue` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/admin/approvals/{requestId}/approve;PUT /api/v1/admin/policies/countries/{countryCode}` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-013;TC-SEC-006` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-ADMIN-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 관리자 전용 RBAC 라우트 가드, 15분 유휴 세션 잠금, Step-up MFA 재인증 모달 구현.
2. `ADM-01`(전체 대시보드), `ADM-02`(회원 관리), `ADM-03`(개발자 KYB 서류 보안 뷰어 심사), `ADM-16`(국가별 정책), `ADM-17`(관리자 권한 및 이중 승인 큐), `ADM-18`(시스템 설정), `ADM-19`(워커 큐 및 해시 체인 감사 로그 탐색기), `ADM-20`(백업/복구 점검) 구현.

## 2. Acceptance Criteria (수용 기준)
1. 모든 관리자 변이 액션 실행 시 사유(`reason`, 최소 10자)와 TOTP 코드 입력이 강제되어야 함.
2. 이중 승인 대기열에서 본인이 발의한 요청의 승인 버튼은 비활성화(`Disabled`)되어야 함.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 감사 체인 불일치 감지 시 화면 상단에 즉시 적색 보안 경보 배너 고정 노출.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-013;TC-SEC-006`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-ADMIN-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

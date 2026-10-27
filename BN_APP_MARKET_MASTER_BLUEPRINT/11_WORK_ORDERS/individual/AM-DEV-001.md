# AM-DEV-001 — 개발자 콘솔(Vue 3) 온보딩, KYB 인증, 조직/팀원 RBAC 및 앱 메타데이터 관리 화면(DEV-01~04,08,09,22~24) 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-DEV-001` |
| **작업명 (Title)** | 개발자 콘솔(Vue 3) 온보딩, KYB 인증, 조직/팀원 RBAC 및 앱 메타데이터 관리 화면(DEV-01~04,08,09,22~24) 구현 |
| **목표 (Objective)** | `REQ-D-001;REQ-D-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Frontend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-ARCH-001;AM-API-003` |
| **개발 Phase 및 우선순위** | `Phase 2` / `P0` |
| **입력 문서 (Input Documents)** | `05_FRONTEND_UXUI/DEVELOPER_CONSOLE.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `apps/web-dev/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `backend/;android-client/;apps/web-user/` |
| **예상 생성/수정 파일 (Target Files)** | `apps/web-dev/src/views/KybVerificationView.vue;apps/web-dev/src/views/CreateAppView.vue;apps/web-dev/src/views/AppMetadataEditorView.vue;apps/web-dev/src/views/TeamMembersRbacView.vue;apps/web-dev/src/views/DeveloperApiKeysView.vue` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/developer/orgs;POST .../verification;POST .../apps;PUT .../metadata;POST .../api-keys` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-003;TC-UI-001` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-DEV-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 개발자 조직 생성 및 KYB 서류 업로드/DNS 검증 위저드 구현.
2. 다국어 앱 메타데이터 편집기, 해상도 자동 검증 스크린샷 업로더, 국가별 배포 허용/차단 매트릭스 구현.
3. 팀원 역할 관리, CI/CD API 키 발급기 및 이의신청 제출 화면 구현.

## 2. Acceptance Criteria (수용 기준)
1. 역할 권한이 없는 메뉴/버튼은 비활성화되고 우회 호출 시 에러 안내 표시.
2. 이미지 해상도 규격 미달 시 업로드 전 클라이언트에서 즉시 안내.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 폼 편집 중 세션 만료 시 작성 중인 초안을 세션 스토리지에 임시 암호화 백업 후 재로그인 시 복원.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-003;TC-UI-001`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-DEV-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

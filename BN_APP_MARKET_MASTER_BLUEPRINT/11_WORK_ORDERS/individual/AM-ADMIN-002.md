# AM-ADMIN-002 — 관리자 콘솔 앱/APK 보안 심사, 출시 승인, 긴급 차단(Kill Switch), 신고 및 커넥터/검색 운영 화면(ADM-04~11) 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-ADMIN-002` |
| **작업명 (Title)** | 관리자 콘솔 앱/APK 보안 심사, 출시 승인, 긴급 차단(Kill Switch), 신고 및 커넥터/검색 운영 화면(ADM-04~11) 구현 |
| **목표 (Objective)** | `REQ-W-001;REQ-W-003;REQ-S-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Frontend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-ADMIN-001;AM-DIST-001;AM-SEC-002` |
| **개발 Phase 및 우선순위** | `Phase 3` / `P0` |
| **입력 문서 (Input Documents)** | `05_FRONTEND_UXUI/ADMIN_CONSOLE.md;09_SECURITY_COMPLIANCE/APP_REVIEW_POLICY.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `apps/web-admin/;backend/internal/modules/admin/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `android-client/;apps/web-user/` |
| **예상 생성/수정 파일 (Target Files)** | `apps/web-admin/src/views/AppMetadataReviewView.vue;apps/web-admin/src/views/ApkSecurityReviewView.vue;apps/web-admin/src/views/EmergencyTakedownView.vue;apps/web-admin/src/views/ModerationReportsQueueView.vue;apps/web-admin/src/views/ExternalConnectorsAdminView.vue` |
| **API 의존성 (API Dependencies)** | `POST /api/v1/admin/reviews/{releaseId}/decision;POST /api/v1/admin/security/kill-switch/{appId};POST /api/v1/admin/connectors/{connectorId}/trigger-sync` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-013;TC-SEC-009` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-ADMIN-002/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. `ADM-04`(앱 메타데이터 심사 체크리스트), `ADM-05`(ClamAV/YARA/위험권한 Diff 보안 심사), `ADM-06`(출시 승인), `ADM-07`(긴급 배포 중단 Kill Switch), `ADM-08`(신고/DMCA 처리 및 이의신청 재심 배정), `ADM-09`(리뷰 어뷰징 관리), `ADM-10`(외부 커넥터 관제), `ADM-11`(검색 랭킹/동의어 운영) 구현.

## 2. Acceptance Criteria (수용 기준)
1. 필수 심사 체크리스트 4항목을 모두 체크하지 않으면 승인 버튼이 활성화되지 않아야 함.
2. Kill Switch 발동 즉시 앱 상태가 `REVOKED_BY_ADMIN`으로 변경되고 다운로드가 차단됨.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 심사자 할당 잠금(`claimed_until`) 30분 경과 시 자동으로 대기열 공용 풀로 반환.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-013;TC-SEC-009`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-ADMIN-002/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

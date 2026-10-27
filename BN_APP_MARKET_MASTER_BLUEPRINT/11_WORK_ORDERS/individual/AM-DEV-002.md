# AM-DEV-002 — 개발자 콘솔 대용량 APK/AAB 업로더, 서명 검증 뷰어, 릴리스 트랙·단계적 배포 및 정산/분석 화면(DEV-05~07,10~21) 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-DEV-002` |
| **작업명 (Title)** | 개발자 콘솔 대용량 APK/AAB 업로더, 서명 검증 뷰어, 릴리스 트랙·단계적 배포 및 정산/분석 화면(DEV-05~07,10~21) 구현 |
| **목표 (Objective)** | `REQ-D-003;REQ-L-001;REQ-L-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Frontend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-DEV-001;AM-DIST-001` |
| **개발 Phase 및 우선순위** | `Phase 3` / `P0` |
| **입력 문서 (Input Documents)** | `05_FRONTEND_UXUI/DEVELOPER_CONSOLE.md;07_DISTRIBUTION/ARTIFACT_PIPELINE.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `apps/web-dev/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `backend/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `apps/web-dev/src/views/ApkUploaderView.vue;apps/web-dev/src/workers/sha256.worker.ts;apps/web-dev/src/views/SigningLineageView.vue;apps/web-dev/src/views/StagedRolloutView.vue;apps/web-dev/src/views/SettlementsDashboardView.vue` |
| **API 의존성 (API Dependencies)** | `POST .../uploads;POST .../complete;POST .../releases;POST .../submit;PATCH .../rollout;GET .../settlements` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-007;TC-FUNC-008` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-DEV-002/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. Web Worker 기반 청크 SHA-256 사전 계산 및 S3 Multipart 이어올리기 업로더 구현.
2. 추출된 매니페스트/권한/서명 지문 확인 화면, 심사 제출(Step-up MFA 모달), 반려 사유 피드백 뷰, 단계적 배포 슬라이더(1%~100%), 긴급 중단(`HALT`) 버튼 구현.
3. 리뷰 공식 답변, 전환율 퍼널 차트, 주문/환불 및 월별 정산 대시보드 구현.

## 2. Acceptance Criteria (수용 기준)
1. 24개 개발자 콘솔 전체 화면이 실제 백엔드 API와 연결되어 동작.
2. 업로드 중 네트워크 단절 후 재연결 시 실패한 파트부터 이어올리기 성공.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 바이너리 검증 실패 시 구체적 에러 코드(`ERR_APK_V2_SIGNATURE_REQUIRED` 등)와 해결 가이드를 화면에 즉시 표시.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-007;TC-FUNC-008`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-DEV-002/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

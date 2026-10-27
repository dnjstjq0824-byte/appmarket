# AM-SEARCH-002 — F-Droid Index v2 서명 검증 커넥터, 파트너 피드 수집기 및 5단계 권리 등급·출처(Provenance) 엔진 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-SEARCH-002` |
| **작업명 (Title)** | F-Droid Index v2 서명 검증 커넥터, 파트너 피드 수집기 및 5단계 권리 등급·출처(Provenance) 엔진 구현 |
| **목표 (Objective)** | `REQ-E-001;REQ-E-002;REQ-E-003` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Search & Connector Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-SEARCH-001` |
| **개발 Phase 및 우선순위** | `Phase 5` / `P1` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/E_EXTERNAL_STORE_CONNECTORS.md;08_SEARCH_CONNECTORS/EXTERNAL_CONNECTOR_CONTRACT.md;08_SEARCH_CONNECTORS/PROVENANCE_AND_RIGHTS.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/connector/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/connector/fdroid/fdroid_v2_syncer.go;backend/internal/modules/connector/partner/partner_feed_syncer.go;backend/internal/modules/connector/domain/rights_guard.go` |
| **API 의존성 (API Dependencies)** | `GET /api/v1/admin/connectors;POST /api/v1/admin/connectors/{connectorId}/trigger-sync;GET /api/v1/catalog/apps/{slug}/sources` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-FUNC-004;TC-SEC-003` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-SEARCH-002/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. F-Droid `entry.json` 및 `index-v2.json` JAR/공개키 암호학적 서명 검증 및 JSON 스트림 파서 구현.
2. 5단계 데이터 권리 등급(`LEVEL_1`~`LEVEL_5`) 강제 가드 및 출처/동기화시각/공식링크 정규화 매퍼 구현.
3. 동일 패키지명 매칭 시 서명 인증서 지문 일치 여부에 따른 호환성 그룹 분리.

## 2. Acceptance Criteria (수용 기준)
1. 변조된 외부 인덱스 피드 주입 시 `ABORTED_SIGNATURE_INVALID`로 즉시 폐기되고 DB 오염 0건.
2. `LEVEL_3`/`LEVEL_4` 외부 앱에 대한 직접 다운로드 요청 100% 차단.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 외부 리포지토리 응답 불능 시 기존 동기화 데이터를 유지하고 관리자 대시보드에 커넥터 상태만 `FAILED`로 표시.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-FUNC-004;TC-SEC-003`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-SEARCH-002/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

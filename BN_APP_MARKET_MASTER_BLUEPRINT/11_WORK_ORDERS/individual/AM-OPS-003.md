# AM-OPS-003 — 글로벌 CDN Signed URL Edge Worker, WAF 방어 룰셋 및 무중단 카나리 배포 자동 롤백 구성

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-OPS-003` |
| **작업명 (Title)** | 글로벌 CDN Signed URL Edge Worker, WAF 방어 룰셋 및 무중단 카나리 배포 자동 롤백 구성 |
| **목표 (Objective)** | `REQ-H-001;REQ-N-003` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `DevOps Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-OPS-001;AM-DIST-002` |
| **개발 Phase 및 우선순위** | `Phase 8` / `P1` |
| **입력 문서 (Input Documents)** | `07_DISTRIBUTION/STORAGE_CDN.md;14_INFRASTRUCTURE_DEVOPS/DEPLOYMENT_PLAN.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `deploy/cdn/;deploy/waf/;deploy/rollout/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `backend/internal/;apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `deploy/cdn/signed_url_verifier.js;deploy/waf/rules.yaml;deploy/rollout/canary_analysis.yaml` |
| **API 의존성 (API Dependencies)** | `NONE` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-OPS-003` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-OPS-003/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. CDN Edge에서 HMAC 서명과 만료 시각(`exp`)을 검증하고 불변 캐시 헤더를 주입하는 설정 구성.
2. 카나리 배포 중 5xx 에러율 > 1% 또는 p99 > 2s 감지 시 자동 롤백하는 배포 파이프라인 구현.

## 2. Acceptance Criteria (수용 기준)
1. 서명 위조/만료 다운로드 URL이 Origin 서버에 도달하기 전 CDN Edge에서 `403` 차단.
2. 불량 이미지 배포 시뮬레이션 시 2분 이내 자동 롤백 완료.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 주 CDN 장애 시 Secondary Origin 도메인으로 다운로드 토큰 발급 호스트를 즉시 Failover.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-OPS-003`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-OPS-003/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

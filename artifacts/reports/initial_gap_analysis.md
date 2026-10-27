# BN App Market 초기 Gap Analysis

- 기준: 블루프린트 v1.0.0 ZIP SHA-256 `fd8111b8d631c533c276857a21fcca1a48972f890dbbb5b97fa21759206abe08`
- 패키지 검증: `sha256sum -c FILE_CHECKSUMS.sha256` 종료 코드 0, 140/140 `OK`
- 저장소 상태: 작업 루트에 Git 저장소 및 애플리케이션 소스 없음. 블루프린트 패키지만 존재.
- 설계 상태와 구현 상태는 별개로 취급한다.

## 초기 상태

| 대상 | NOT_STARTED | PARTIAL | IMPLEMENTED_UNVERIFIED | VERIFIED | FAILED | BLOCKED |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Work Order (41개) | 41 | 0 | 0 | 0 | 0 | 0 |
| 요구사항 (66개) | 66 | 0 | 0 | 0 | 0 | 0 |

전체 항목별 분류는 [initial_work_order_status.csv](initial_work_order_status.csv)와 [initial_requirement_status.csv](initial_requirement_status.csv)에 기록했다. 두 CSV의 `implementation_status`는 모두 `NOT_STARTED`이다. 이 표는 코딩 착수 *전* 기준이며, 이후 변경은 별도 진행 보고로 기록한다.

## 실행 계획

1. Wave 0: `AM-ARCH-001`의 Go 1.23 모듈형 모놀리스 기반, 실제 미들웨어 및 헬스체크를 구현하고 빌드/테스트를 실행한다.
2. 이어서 `AM-OPS-001`의 격리된 Docker 서비스, 세 버킷, OTel 관제와 CI 게이트를 구성하고 실제 컨테이너 검증을 실행한다.
3. Wave 0 게이트가 충족되면 `AM-DB-001` → `AM-DB-002` → IAM/보안/워커 순으로 Wave 1에 진입한다. 이후 Wave 2~5는 `EXECUTION_SEQUENCE.md`의 선행관계와 각 Work Order의 테스트 게이트를 따른다.
4. 각 Work Order 결과는 `artifacts/evidence/<ID>/`에 실제 실행 명령, 전체 출력, 커버리지 및 미완료 사유를 남긴다. 독립 감사 전에는 `VERIFIED`로 표시하지 않는다.

## 착수 전 명세 불일치

- `02_SYSTEM_ARCHITECTURE/SYSTEM_OVERVIEW.md`는 32개 핵심 테이블을 기술하고, Wave 1 게이트는 34개를 요구하지만 `0001_initial_schema.up.sql`에는 부모/일반 `CREATE TABLE` 54개와 파티션 테이블 4개가 있다. 실행 가능한 DDL을 기준으로 실제 개수를 검증하고 문서 정합성을 별도로 정리해야 한다.
- `0002_indexes_and_constraints.up.sql`의 원장 제약 트리거는 `DELETE` 이벤트를 포함하면서 함수에서 `NEW.transaction_id`를 참조한다. Wave 1의 `TC-COM-005` 및 삭제 경로 테스트 전에 수정이 필요하다.
- `TC-OPS-001`과 `TC-OPS-004`는 각각 실제 격리 워커 강제 장애와 실행 중 워커 재수거를 요구한다. Wave 0 단계에서 필요한 워커 기능이 준비되지 않으면 테스트를 통과했다고 보고할 수 없다.

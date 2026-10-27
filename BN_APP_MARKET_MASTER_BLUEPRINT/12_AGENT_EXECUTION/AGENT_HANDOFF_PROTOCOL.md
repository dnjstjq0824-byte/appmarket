# AGENT_HANDOFF_PROTOCOL.md — 에이전트 간 작업 인계 및 상태 전이 프로토콜

- **문서 ID**: `DOC-EXEC-001`
- **관련 조항**: 제15조, 제16조, 제23조
- **루트 진입점**: [START_HERE.md](../START_HERE.md)

---

## 1. 작업 착수 전-중-후 표준 인계 절차

1. **착수 전 (Pre-Flight Check)**:
   - 선행 Work Order(`prerequisites`)의 상태가 `IMPLEMENTED_UNVERIFIED` 이상인지 확인합니다.
   - 할당된 Work Order의 `입력 문서`와 `04_API_SPECIFICATION/openapi.yaml`, `03_DATABASE/migrations/`를 읽고 인터페이스 계약을 확인합니다.
2. **작업 수행 중 (In-Flight Isolation)**:
   - 지정된 `수정 가능 디렉터리` 외부의 파일은 절대 건드리지 않습니다.
   - 공유 계약 수정이 필요한 경우 임의로 수정하지 않고 `Lead Architect Agent`에게 RFC를 요청합니다.
3. **인계 및 감사 요청 (Handoff to QA & Independent Audit)**:
   - 단위/통합 테스트 통과 후 `artifacts/evidence/<WORK_ORDER_ID>/`에 실행 로그를 저장하고 [REPORTING_TEMPLATE.md](REPORTING_TEMPLATE.md) 양식으로 인계서를 작성합니다.
   - `Independent Audit Agent`가 클린 컨테이너에서 재검증하여 통과했을 때만 최종 `VERIFIED`로 승인됩니다.

# INDEPENDENT_AUDIT.md — 독립 감사 에이전트(Independent Audit Agent) 교차 검증 프로토콜

- **문서 ID**: `DOC-QA-003`
- **관련 조항**: 제15조 (Step 8), 제16조, 제18조, 제23조
- **연관 Work Order**: `AM-TEST-002`

---

## 1. 독립 감사 에이전트의 권한 및 7단계 감사 절차

`Independent Audit Agent`는 애플리케이션 코드를 절대 작성하지 않으며, 개발 에이전트들이 제출한 Work Order 완료 보고서를 다음 7단계로 엄격히 감사합니디:

1. **Step 1 (Git 커밋 및 디렉터리 소유권 검증)**: 보고된 Commit SHA가 실제 Git 히스토리에 존재하는지(`git cat-file -t <sha>`), 그리고 해당 Work Order의 `writable_dirs` 외부 파일이 무단 수정되지 않았는지 확인합니다.
2. **Step 2 (Mock 및 하드코딩 우회 탐지 스캔)**: 프로덕션 코드(`backend/internal/`, `apps/*/src/`, `android-client/*/src/main/`) 내에 `TODO`, `FIXME`, `mockSuccess`, 하드코딩된 가짜 서명 결과, 가짜 설치 성공 코드가 존재하는지 AST/정규식 스캔으로 전수 조사합니다.
3. **Step 3 (시크릿 및 무단 재배포 코드 스캔)**: `gitleaks`로 개인키/비밀번호 포함 여부를 검사하고, 타사 독점 스토어 APK 무단 스크래핑 코드가 없는지 확인합니다.
4. **Step 4 (독립 클린 컨테이너 테스트 재실행)**: 개발 에이전트가 제출한 로그 파일만 믿지 않고, 격리된 새 테스트 러너에서 해당 Work Order의 `mandatory_tests`를 직접 재실행하여 동일한 `PASS`가 나오는지 확인합니다.
5. **Step 5 (DB 제약조건 및 원장/감사체인 무결성 쿼리 실행)**: PostgreSQL에 직접 접속하여 `ledger_entries` 대차평형(`SUM=0`)과 `audit_logs` 해시 체인 무결성을 검증합니다.
6. **Step 6 (문서-코드-계약 일치성 대조)**: 구현된 엔드포인트와 테이블 컬럼이 `04_API_SPECIFICATION/openapi.yaml` 및 `03_DATABASE/migrations/`와 100% 일치하는지 확인합니다.
7. **Step 7 (최종 판정 서명)**: 1~6단계를 모두 통과한 경우에만 상태를 `VERIFIED`로 변경하고, 단 하나라도 실패하면 `FAILED`와 구체적 결함 사유를 기록하여 반려합니다.

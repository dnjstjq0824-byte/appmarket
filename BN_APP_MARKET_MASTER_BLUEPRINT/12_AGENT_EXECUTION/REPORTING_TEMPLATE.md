# REPORTING_TEMPLATE.md — Work Order 및 Phase 완료 보고서 표준 양식

- **문서 ID**: `DOC-EXEC-005`
- **관련 조항**: 제15조 (Step 10), 제23조

---

## 1. 개별 Work Order 완료 보고 템플릿 (에이전트 제출용)

```markdown
# Work Order Completion Report: [AM-XXX-000]

- **Work Order ID**: `AM-XXX-000`
- **담당 에이전트**: `[Agent Role Name]`
- **보고 상태**: `IMPLEMENTED_UNVERIFIED` | `PARTIAL` | `BLOCKED` | `FAILED`
  *(주의: 담당 에이전트는 스스로 VERIFIED를 선언할 수 없음 — Independent Audit Agent만 선언 가능)*
- **Git Branch & Commit SHA**: `feat/AM-XXX-000` / `<40-char git rev-parse HEAD>`

## 1. 변경된 파일 목록 (`git diff --stat`)
- `path/to/file1.go` (+240, -0)
- `path/to/file1_test.go` (+180, -0)

## 2. 실제 실행한 테스트 명령어 및 결과
- 실행 명령어: `go test -v -race -coverprofile=coverage.out ./internal/modules/...`
- 테스트 결과: `PASS (14 passed, 0 failed, 0 skipped)` — 라인 커버리지: `88.4%`
- 증빙 파일 경로:
  - `artifacts/evidence/AM-XXX-000/test_output.log` (SHA-256: `<hash>`)
  - `artifacts/evidence/AM-XXX-000/coverage.out` (SHA-256: `<hash>`)

## 3. 미완료 또는 차단 사항 (Unfinished / Blocked Items)
- 없음 (또는 구체적 차단 사유와 필요한 외부 조치 기재)
```

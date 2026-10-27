# CONFLICT_PREVENTION.md — 멀티 에이전트 파일 충돌 방지, 브랜치 전략 및 공유 계약 변경 절차

- **문서 ID**: `DOC-EXEC-004`
- **관련 조항**: 제16조

---

## 1. Git 브랜치 격리 및 CODEOWNERS 강제 규칙

1. **Work Order 단위 전용 브랜치**: 모든 에이전트는 `feat/<WORK_ORDER_ID>` (예: `feat/AM-DIST-001`) 브랜치에서 작업하며, PR 병합 시 CI 스크립트가 해당 Work Order의 `writable_dirs` 외부 파일이 변경되었는지 자동 검사합니다.
2. **자동 생성 코드 커밋 충돌 방지**:
   - `packages/api-client/src/generated/schema.d.ts` 및 `backend/db/sqlc/` 코드는 수동 편집을 절대 금지하며, 오직 계약 변경 PR에서만 재생성합니다.
3. **공유 계약 변경 4단계 프로토콜**:
   - Step 1: 변경 필요성 발생 시 `RFC` 이슈 생성 (변경 필드, 하위 호환성 여부 명시).
   - Step 2: `Lead Architect Agent` 승인 후 `04_API_SPECIFICATION/openapi.yaml` 또는 신규 마이그레이션 파일(`0004_*.up.sql`) 추가(기존 마이그레이션 파일 사후 변조 금지).
   - Step 3: 타입 클라이언트 재생성 후 영향받는 에이전트에게 알림.
   - Step 4: 계약 테스트(`oasdiff` breaking change 검사) 통과 후 병합.

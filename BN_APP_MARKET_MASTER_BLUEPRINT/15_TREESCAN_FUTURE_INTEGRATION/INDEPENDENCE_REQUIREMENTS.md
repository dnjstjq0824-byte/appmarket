# INDEPENDENCE_REQUIREMENTS.md — TreeScan 통합 전 독립성 보장을 위한 10대 아키텍처 가드레일

- **문서 ID**: `DOC-TS-004`
- **관련 조항**: 제1조 1.1항

---

## 1. 독립성 위반 자동 검출 체크리스트 (`TC-OPS-008`)

후속 개발 에이전트는 개발 전 과정에서 아래 5가지 금지 규칙을 준수해야 하며, CI 파이프라인이 이를 매 커밋마다 자동 검사합니다:
1. **코드 레벨 외부 import 금지**: `backend/`, `apps/`, `android-client/` 내에 외부 TreeScan 내부 패키지나 비공개 SDK 참조가 0건이어야 합니다.
2. **DB 스키마 공유 금지**: 앱 마켓 PostgreSQL DB는 오직 앱 마켓 서비스만 접근하며, TreeScan DB에 대한 Foreign Data Wrapper(FDW)나 크로스 DB 조인을 절대 생성하지 않습니다.
3. **환경변수 기본값 격리**: `TREESCAN_INTEGRATION_ENABLED`의 기본값은 항상 `false`여야 하며, 이 상태에서 모든 헬스체크(`/healthz/ready`)와 E2E 테스트가 100% 통과해야 합니다.
4. **네트워크 및 스레드 풀 Bulkhead**: Phase 10에서 어댑터를 켜더라도 TreeScan 인바운드 요청이 앱 마켓 고유 사용자의 다운로드/결제 API 리소스를 잠식하지 못하도록 별도 Rate Limiter로 격리합니다.
5. **TreeScan 본체 수정 금지**: 별도 PO 서면 승인이 떨어지기 전까지 TreeScan 저장소에 대한 어떠한 PR이나 커밋도 생성하지 않습니다.

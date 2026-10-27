# KNOWN_LIMITATIONS.md — 현 단계 미검증/미구현 구분 및 상용 운영 전 외부 확인 필수 사항

- **문서 ID**: `DOC-DELIV-004`
- **관련 조항**: 제1조 1.2항, 제13조, 제22조, 제24조 I항, 제25조 10항

---

## 1. 설계 완료(`DESIGN_VERIFIED`)와 실제 개발 미착수(`NOT_STARTED`)의 명확한 구분

1. **애플리케이션 소스코드 미구현 명시**:
   - 본 프로젝트 명령(`BN-APP-MARKET-GEMINI4-ARGON-MASTER-BLUEPRINT-001`)의 작업 유형은 **"전체 시스템 설계 + A~Z 기술 명세 + 에이전트 작업지시서 + ZIP 패키지 제작"**입니다.
   - 따라서 제1조 1.2항에 의거하여, 현재 시점에는 Go 백엔드 코드, Vue/Nuxt 프론트엔드 코드, Kotlin Android 코드가 아직 작성되지 않았으므로 모든 Work Order(`AM-*`)의 구현 상태는 정직하게 `NOT_STARTED`로, 테스트 케이스(`TC-*`)의 실행 상태는 `NOT_EXECUTED`로 기록되어 있습니다.
2. **샌드박스 환경 도구 제약에 따른 검증 범위 투명 공개**:
   - 본 설계 패키지를 생성한 VM 환경에는 JSON, YAML, CSV, SHA-256, ZIP 검증 도구가 존재하여 해당 항목은 100% 실제 도구로 실행 검증(`PASS`)하였습니다.
   - 반면 PostgreSQL 16 서버 데몬(`postgres`/`psql`)과 실제 Android 에뮬레이터(`adb`)는 후속 개발 에이전트가 `AM-DB-001` 및 `AM-ANDROID-004`를 수행할 CI/개발 클러스터에서 구동되므로, SQL DDL은 구문/참조 정적 검증(`STATIC_VERIFIED`)까지 수행하고 실제 DB 엔진 실행은 `AM-DB-001` 단계의 검증 과제로 명시하였습니다.

---

## 2. 상용 오픈 전 발주자(PO) 및 법무·재무·외부 사업자 확정이 필요한 사항

| 구분 코드 | 미확정 / 외부 의존 항목 | 현재 블루프린트 반영 상태 | 상용 오픈 전 해결 방안 및 의사결정 주체 |
| :---: | :--- | :--- | :--- |
| `LIM-01` | **국가별 판매대행(MoR) 지위 및 전자금융업/PG 계약** | `country_policies` 및 `commission_policies` 테이블에서 설정값으로 즉시 변경 가능하도록 유연하게 설계됨 | Booking Network 법무·재무팀이 서비스 출시 국가별 PG사(Stripe, Toss Payments 등) 본계약 체결 및 부가세 원천징수 주체 확정 |
| `LIM-02` | **외부 독점 스토어(Google Play, Galaxy Store) 공식 데이터 제휴** | 무단 APK 복제 및 비공식 API 크롤링은 전면 금지(`LEVEL_4_OFFICIAL_LINK_ONLY` 기본 적용) | 제휴 전까지는 개발자가 직접 등록한 공식 스토어 링크와 F-Droid 오픈소스 리포지토리(`LEVEL_2`)만 활성화 운영 |
| `LIM-03` | **상용 코드 서명 인증서(EV/OV Code Signing) 및 Android 마켓 서명키 발급** | KMS/HSM 볼트 경로 및 `PINNED_STORE_CERT_SHA256` 슬롯 설계 완료 | DevOps 및 Security 팀이 프로덕션 HSM에서 마켓 전용 루트 서명키를 생성하고 오프라인 백업 보관 |
| `LIM-04` | **TreeScan 본체와의 실제 통합 시점** | Phase 10 어댑터(`AM-TREESCAN-001`) 및 `TREESCAN_INTEGRATION_ENABLED=false` 격리 설계 완료 | 독립 앱 마켓 상용 안정화(Phase 9 완료) 이후 PO의 별도 통합 프로젝트 승인을 거쳐 활성화 |

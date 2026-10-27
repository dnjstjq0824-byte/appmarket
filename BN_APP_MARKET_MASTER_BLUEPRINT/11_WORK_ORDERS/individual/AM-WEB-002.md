# AM-WEB-002 — 사용자 마켓 웹사이트 통합 검색, 배포처·가격 비교, 정직한 설치 CTA 및 계정/신고 화면(USR-02,03,13~20) 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-WEB-002` |
| **작업명 (Title)** | 사용자 마켓 웹사이트 통합 검색, 배포처·가격 비교, 정직한 설치 CTA 및 계정/신고 화면(USR-02,03,13~20) 구현 |
| **목표 (Objective)** | `REQ-F-002;REQ-U-003;REQ-E-003` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Frontend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-WEB-001;AM-SEARCH-001;AM-DIST-002` |
| **개발 Phase 및 우선순위** | `Phase 5` / `P0` |
| **입력 문서 (Input Documents)** | `05_FRONTEND_UXUI/USER_WEB.md;05_FRONTEND_UXUI/RESPONSIVE_ACCESSIBILITY.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `apps/web-user/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `backend/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `apps/web-user/pages/search.vue;apps/web-user/components/DistributionComparisonMatrix.vue;apps/web-user/components/HonestInstallActionBox.vue;apps/web-user/pages/account/library.vue;apps/web-user/pages/support/report.vue` |
| **API 의존성 (API Dependencies)** | `GET /api/v1/search/apps;GET /api/v1/search/suggest;POST .../download-token;GET/POST .../reviews;POST /api/v1/moderation/reports` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-UI-003;TC-FUNC-005` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-WEB-002/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 실시간 자동완성(`USR-03`), 패싯 검색 결과(`USR-02`), 다중 배포처 비교(`USR-14`), 가격 비교(`USR-15`) 구현.
2. 권리 등급(`rights_level`)에 따라 직접 APK 다운로드(`USR-17`)와 외부 스토어 공식 링크 이동(`USR-16`)을 엄격히 분리하는 `HonestInstallActionBox` 구현(가짜 설치 버튼 0건 보장).
3. 리뷰(`USR-13`), 구매/다운로드 라이브러리(`USR-18`), 계정/MFA(`USR-19`), 신고 센터(`USR-20`) 및 아랍어 RTL 대응 완성.

## 2. Acceptance Criteria (수용 기준)
1. `EXTERNAL_LINK_ONLY` 앱에서 가짜 다운로드 버튼이 절대 렌더링되지 않고 공식 링크 이동 안내와 동기화 시각이 표시되어야 함.
2. `@axe-core/playwright` 접근성 Critical/Serious 위반 0건.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 다운로드 토큰 만료 시 사용자에게 안내 후 원클릭으로 새 서명 URL을 재발급.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-UI-003;TC-FUNC-005`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-WEB-002/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

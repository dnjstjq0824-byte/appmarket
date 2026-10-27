# AM-WEB-001 — 사용자 마켓 웹사이트(Nuxt 3 SSR) 코어 레이아웃, 디자인 시스템 및 탐색/상세 화면(USR-01~USR-12) 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-WEB-001` |
| **작업명 (Title)** | 사용자 마켓 웹사이트(Nuxt 3 SSR) 코어 레이아웃, 디자인 시스템 및 탐색/상세 화면(USR-01~USR-12) 구현 |
| **목표 (Objective)** | `REQ-F-001;REQ-U-001;REQ-U-002` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Frontend Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-ARCH-001;AM-API-002` |
| **개발 Phase 및 우선순위** | `Phase 2` / `P0` |
| **입력 문서 (Input Documents)** | `05_FRONTEND_UXUI/USER_WEB.md;05_FRONTEND_UXUI/DESIGN_SYSTEM.md;05_FRONTEND_UXUI/SCREEN_INVENTORY.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `apps/web-user/;packages/ui-kit/;packages/api-client/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `backend/;android-client/;apps/web-admin/` |
| **예상 생성/수정 파일 (Target Files)** | `packages/ui-kit/src/tokens.css;apps/web-user/nuxt.config.ts;apps/web-user/pages/index.vue;apps/web-user/pages/apps/[slug].vue;apps/web-user/components/TrustTransparencyCard.vue` |
| **API 의존성 (API Dependencies)** | `GET /api/v1/ui/home;GET /api/v1/catalog/apps;GET /api/v1/catalog/apps/{slug};GET /api/v1/catalog/categories` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-UI-001;TC-UI-002` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-WEB-001/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. `packages/ui-kit` 디자인 토큰(라이트/다크 모드) 및 `openapi-typescript` 자동 생성 클라이언트 구성.
2. `USR-01`(메인), `USR-04~07`(카테고리/인기/신규/추천), `USR-08~12`(앱 상세, 미디어 라이트박스, 버전 기록, 권한/보안 분석, 개인정보/개발자 정보) 구현.
3. SSR `Schema.org/SoftwareApplication` JSON-LD 및 동적 메타태그 렌더링.

## 2. Acceptance Criteria (수용 기준)
1. `USR-01` ~ `USR-12` 모든 화면에서 5대 상태(Loading Skeleton, Data, Empty, Error Retry) 동작.
2. SSR HTML 응답에 JSON-LD 및 OpenGraph 태그 포함.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 부가 추천 API 지연 시 `<Suspense>` 폴백으로 핵심 앱 상세 정보와 설치 액션은 즉시 렌더링.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-UI-001;TC-UI-002`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-WEB-001/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

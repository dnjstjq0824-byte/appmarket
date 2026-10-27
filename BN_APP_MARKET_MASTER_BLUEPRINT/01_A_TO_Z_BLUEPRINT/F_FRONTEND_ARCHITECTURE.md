# F — Frontend Architecture — 사용자 웹(Nuxt 3 SSR), 개발자·관리자 콘솔(Vue 3) 및 공통 디자인 시스템 아키텍처 (User Web (Nuxt 3 SSR), Developer/Admin Consoles (Vue 3) & Design System Architecture)

- **문서 ID**: `DOC-AZ-F`
- **알파벳 코드**: `F`
- **관련 요구사항 ID**: `REQ-F-001`, `REQ-F-002`, `REQ-U-001`
- **연관 Work Order ID**: `AM-WEB-001`, `AM-WEB-002`, `AM-DEV-001`, `AM-ADMIN-001`
- **연관 테스트 케이스 ID**: `TC-UI-001`, `TC-UI-002`
- **참조 상세 설계 문서**: [05_FRONTEND_UXUI/USER_WEB.md](../05_FRONTEND_UXUI/USER_WEB.md), [05_FRONTEND_UXUI/DESIGN_SYSTEM.md](../05_FRONTEND_UXUI/DESIGN_SYSTEM.md), [05_FRONTEND_UXUI/RESPONSIVE_ACCESSIBILITY.md](../05_FRONTEND_UXUI/RESPONSIVE_ACCESSIBILITY.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 제2조에서 지정한 Vue 3 + Nuxt + TypeScript 기술 스택을 바탕으로, 검색엔진 유입(SEO)과 빠른 초기 렌더링이 필수적인 사용자 마켓 웹(`web-user`), 복잡한 파일 업로드와 상태 제어가 필요한 개발자 콘솔(`web-dev`), 고밀도 데이터 테이블과 감사 워크플로가 필요한 관리자 콘솔(`web-admin`)을 일관된 디자인 시스템과 높은 접근성·성능 기준으로 구축하기 위해 필요하다.
- **달성 목표**: 3개 웹 애플리케이션의 렌더링 전략, 상태 관리(Pinia), OpenAPI 기반 타입 안전 API 클라이언트 자동 생성, 반응형 브레이크포인트, WCAG 2.1 AA 접근성, Core Web Vitals 목표치를 정의한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **일반 사용자 (PC/모바일 웹 방문자)**: 검색엔진(Google, TreeScan 등) 또는 직접 유입을 통해 앱을 탐색하고 상세 정보를 비교.
- **개발자 사용자**: 데스크톱/태블릿 환경에서 앱 등록 및 바이너리 배포 관리.
- **관리자 사용자**: 고해상도 데스크톱 환경에서 다중 심사 큐 및 관제 지표 운영.

## 3. 필수 기능 (Mandatory Features)
1. **3개 프론트엔드 앱 분리 및 공통 패키지 공유**:
   - `apps/web-user`: Vue 3.5 + Nuxt 3.13 + TypeScript 5.6 (SSR + Route-level SWR 캐싱).
   - `apps/web-dev`: Vue 3.5 + Vite 5 + TypeScript 5.6 (Strict CSR SPA + 인증 보호).
   - `apps/web-admin`: Vue 3.5 + Vite 5 + TypeScript 5.6 (Strict CSR SPA + RBAC/MFA/Dual-Approval 가드).
   - `packages/ui-kit` 및 `packages/api-client`: `openapi.yaml`로부터 `openapi-typescript`로 자동 생성된 100% 타입 안전 클라이언트 공유.
2. **가짜 UI(Fake UI) 절대 금지**: 백엔드 API와 연결되지 않은 더미 버튼이나, 직접 다운로드가 불가능한 외부 앱에 대해 클릭 가능한 가짜 '직접 설치' 버튼을 렌더링하는 행위를 금지한다.
3. **성능 및 SEO 기준**: `web-user`는 동적 `sitemap.xml`, `Schema.org/SoftwareApplication` JSON-LD 구조화 데이터, OpenGraph/Twitter 메타태그를 서버 사이드에서 렌더링하며, LCP < 2.0초, CLS < 0.05, INP < 150ms를 달성한다.
4. **접근성(Accessibility)**: 모든 인터랙티브 요소의 키보드 포커스 이동(`Tab`/`Shift+Tab`/`Enter`/`Escape`), ARIA 라이브 리전, 명도 대비 4.5:1 이상(다크모드/라이트모드 모두)을 보장한다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **Nuxt 3 하이브리드 렌더링 정책 (`nuxt.config.ts` `routeRules`)**:
  - `/`, `/categories/**`, `/charts/**`: `isr: 60` (60초 Stale-While-Revalidate SSR 캐싱).
  - `/apps/[slug]`: `isr: 30` + 배포 승인 웹훅 수신 시 즉시 캐시 퍼지.
  - `/search`: 실시간 SSR + 클라이언트 필터 전환 시 CSR Fetch.
  - `/account/**`, `/checkout/**`: `ssr: false` (개인화/결제 영역은 캐싱 완전 비활성화 및 `Cache-Control: private, no-store`).
- **대용량 APK 업로드 프론트엔드 아키텍처 (`web-dev`)**:
  - Web Worker 내부에서 업로드할 APK 파일의 `SHA-256` 해시를 4MB 청크 단위로 먼저 계산(`crypto.subtle.digest`)하여 UI 스레드 블로킹을 방지하고, S3 Multipart Upload(Presigned Part URLs)를 통해 네트워크 단절 시 실패한 청크부터 이어올리기(Resume)를 수행한다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **상태 관리 구조 (Pinia Stores)**:
  - `useAuthStore`: 액세스 토큰(메모리 보관) + HttpOnly Secure Refresh Cookie 갱신 인터셉터.
  - `useCatalogFilterStore`: URL 쿼리 파라미터(`?q=...&category=...&source=direct&sort=rating`)와 양방향 동기화되는 검색/필터 상태.
  - `useUploadSessionStore` (`web-dev`): 멀티파트 업로드 진행률, 청크 재시도, 서버 측 검증 폴링/SSE 상태 관리.

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- BFF(Backend-For-Frontend) 역할은 Nuxt Server Engine(`Nitro`)과 Go API의 조합으로 처리하며, SSR 렌더링 시에는 클러스터 내부 Private DNS(`http://backend-api.internal:8080`)를 통해 저지연(< 5ms)으로 데이터를 조회한다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- 프론트엔드 전용 영속 DB는 두지 않으며(무상태 원칙), 사용자 테마/언어/최근 검색어는 `localStorage` 및 계정 설정 테이블(`user_preferences`)에 동기화한다.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- 프론트엔드가 호출하는 모든 엔드포인트는 [04_API_SPECIFICATION/openapi.yaml](../04_API_SPECIFICATION/openapi.yaml)에 정의되어 있으며, CI 파이프라인에서 `openapi-typescript` 코드젠 결과와 Git 작업 트리가 일치하는지 검증한다.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **화면 공통 5대 뷰 상태 (5-State View Pattern)**:
  - 모든 데이터 패칭 화면은 반드시 ① `IDLE`, ② `LOADING` (레이아웃 시프트가 없는 Skeleton UI), ③ `SUCCESS_DATA` (정상 렌더링), ④ `SUCCESS_EMPTY` (조건에 맞는 앱이 없을 때 필터 초기화 버튼을 제공하는 빈 상태 화면), ⑤ `ERROR` (에러 코드, `trace_id`, 재시도 버튼을 제공하는 에러 바운더리) 5가지 상태를 명시적으로 구현해야 한다.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- **CSP(Content-Security-Policy) 강제**: `default-src 'self'; script-src 'self' 'nonce-{random}'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'` 헤더를 적용하여 XSS 및 Clickjacking을 방어한다.
- JWT Access Token은 `localStorage`에 저장하지 않고 Pinia 인메모리에만 유지하며, Refresh Token은 `HttpOnly; Secure; SameSite=Strict` 쿠키로만 전송한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 반응형 브레이크포인트: Mobile (`320px~767px`), Tablet (`768px~1023px`), Desktop (`1024px~1439px`), Wide (`1440px+`).
- 디자인 토큰(컬러, 타이포그래피, 간격, 그림자, 다크모드 변수)은 [05_FRONTEND_UXUI/DESIGN_SYSTEM.md](../05_FRONTEND_UXUI/DESIGN_SYSTEM.md)에 정의된 CSS Custom Properties를 사용한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **E2E 및 접근성 자동화 테스트 (`TC-UI-001`, `TC-UI-002`)**: Playwright + `@axe-core/playwright`를 사용해 20개 사용자 웹 화면, 24개 개발자 콘솔 화면, 20개 관리자 콘솔 화면에서 심각한 접근성 위반(Critical/Serious) 0건 및 반응형 뷰포트 깨짐 0건을 검증한다.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `U — User Experience & UI`, `I — Identity & Access`, `C — Catalog Engine`, `D — Developer Platform`, `W — Workflow & Administration`과 직접 연결된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0 (핵심 접점)`
- **개발 Phase**: `Phase 2`에서 기반 구조 및 콘솔을 구축하고 `Phase 3~7`에 걸쳐 도메인 기능을 완성한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: SSR 렌더링 중 백엔드 특정 부가 API(예: 추천 위젯) 지연** → Nuxt `<Suspense>` 및 `lazy: true` 비동기 컴포넌트 분리로 핵심 앱 상세 정보와 설치 버튼은 즉시 SSR로 내려주고, 부가 추천 영역만 클라이언트에서 점진적 로딩(Graceful Degradation) 처리한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- `apps/web-user/`, `apps/web-dev/`, `apps/web-admin/`, `packages/ui-kit/`, `packages/api-client/`.
- Playwright E2E 테스트 스위트 (`e2e/`).

## 17. Acceptance Criteria (수용 기준)
1. `pnpm build` 및 `pnpm typecheck` 실행 시 3개 웹 앱 모두 TypeScript 컴파일 에러 0건으로 빌드되어야 한다.
2. 앱 상세 페이지의 SSR HTML 응답 원문에 `<title>`, `<meta property="og:title">`, `<script type="application/ld+json">` 구조화 데이터가 실제로 포함되어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-WEB-001/ssr_curl_snapshot.html`: 실제 SSR 렌더링된 HTML 원문 및 JSON-LD 검증 파일.
- `artifacts/evidence/AM-WEB-001/playwright_axe_report.json`: Playwright + axe-core 접근성 및 반응형 E2E 통과 리포트.

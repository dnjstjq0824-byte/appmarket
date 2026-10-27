# U — User Experience & UI — 전체 사용자 여정, 화면별 와이어프레임, 디자인 토큰, 다크모드 및 오류·빈 상태 설계 (User Journeys, Screen Wireframes, Design Tokens, Dark Mode & Empty/Error States)

- **문서 ID**: `DOC-AZ-U`
- **알파벳 코드**: `U`
- **관련 요구사항 ID**: `REQ-U-001`, `REQ-U-002`, `REQ-U-003`
- **연관 Work Order ID**: `AM-WEB-001`, `AM-WEB-002`, `AM-DEV-001`, `AM-ADMIN-001`, `AM-ANDROID-001`
- **연관 테스트 케이스 ID**: `TC-UI-001`, `TC-UI-002`, `TC-UI-003`
- **참조 상세 설계 문서**: [05_FRONTEND_UXUI/USER_WEB.md](../05_FRONTEND_UXUI/USER_WEB.md), [05_FRONTEND_UXUI/DEVELOPER_CONSOLE.md](../05_FRONTEND_UXUI/DEVELOPER_CONSOLE.md), [05_FRONTEND_UXUI/ADMIN_CONSOLE.md](../05_FRONTEND_UXUI/ADMIN_CONSOLE.md), [05_FRONTEND_UXUI/DESIGN_SYSTEM.md](../05_FRONTEND_UXUI/DESIGN_SYSTEM.md), [05_FRONTEND_UXUI/SCREEN_INVENTORY.md](../05_FRONTEND_UXUI/SCREEN_INVENTORY.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: Google Play(풍부한 미디어와 추천, 그러나 서명 지문·구버전·타 배포처 비교 불가), F-Droid(투명한 서명·버전·오픈소스 정보, 그러나 일반 사용자에게 다소 어려운 탐색 UX), 외부 앱 검색 서비스의 장단점을 종합 분석하여, **일반 사용자에게는 직관적이고 빠르며 파워유저에게는 배포처·서명·권한·버전 비교의 투명성을 제공하는 독창적인 통합 앱 마켓 UX/UI**(제4조)를 구축하기 위해 필요하다.
- **달성 목표**: 제4조의 필수 20개 사용자 웹 화면 + 개발자 콘솔 24개 화면 + 관리자 콘솔 20개 화면 + Android 12개 화면(총 76개 화면 인벤토리)의 레이아웃, 디자인 토큰, 다크모드, 반응형 구조 및 5대 화면 상태를 명세한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **모든 플랫폼 사용자 (End User, Developer, Admin)**: 일관된 디자인 언어와 예측 가능한 피드백을 제공하는 UI를 통해 작업 수행.

## 3. 필수 기능 (Mandatory Features)
1. **기존 스토어 장단점 분석 기반 3대 고유 UX 차별화 (제4조)**:
   - **차별점 1 — 통합 배포처·가격·서명 비교 매트릭스 (`Universal Source Switcher`)**: 앱 상세 페이지 상단에서 `[BN 직접 다운로드 (무료, v2.4.1)]`, `[F-Droid 오픈소스 빌드 (무료, v2.4.0)]`, `[Google Play 공식 링크 (₩3,300)]` 탭을 전환하며 가격, 버전, 요구 SDK, 서명 호환성을 즉시 비교.
   - **차별점 2 — 투명한 보안·권한·서명 카드 (`Trust & Transparency Card`)**: 숨겨진 팝업이 아닌 본문 핵심 영역에 요구 권한의 위험도 배지, APK SHA-256 체크섬, 개발자 X.509 인증서 지문, 마지막 보안 스캔 일시를 명확히 시각화.
   - **차별점 3 — 정직한 설치 액션 분리 (`Honest Install CTA`)**: 자체 호스팅 APK는 `직접 APK 다운로드 (42.8 MB)` 또는 Android 앱 내 `즉시 설치` 버튼을 제공하고, 외부 스토어 링크 전용 앱은 `Google Play 공식 페이지로 이동 ↗` 버튼을 제공하여 가짜 설치 버튼을 완전히 배제.
2. **제4조 필수 20개 사용자 웹 화면 전수 설계**:
   - `USR-01` 앱스토어 메인, `USR-02` 앱 검색 결과, `USR-03` 검색 자동완성 오버레이, `USR-04` 카테고리 허브, `USR-05` 카테고리별 인기 앱, `USR-06` 신규 출시 앱, `USR-07` 에디터 추천 앱, `USR-08` 앱 상세정보(종합), `USR-09` 스크린샷 라이트박스 및 영상 플레이어, `USR-10` 앱 버전 기록 및 아카이브, `USR-11` 앱 권한 및 보안 분석 상세, `USR-12` 앱 개인정보처리방침 및 개발자 신원 정보, `USR-13` 앱 리뷰 및 평점 목록/작성, `USR-14` 다중 배포처 비교 매트릭스, `USR-15` 글로벌/배포처 가격 비교, `USR-16` 공식 설치 링크 이동 안내(인터스티셜/딥링크), `USR-17` 직접 APK 다운로드 및 설치 가이드, `USR-18` 사용자 다운로드 및 구매 기록, `USR-19` 계정·보안(MFA)·기기 관리, `USR-20` 유해 앱 신고 및 고객지원 센터.

## 4. 상세 기능 (Detailed Functional Specifications)
- **디자인 토큰 체계 (`packages/ui-kit/tokens.css`)**:
  - **Brand & Semantic Colors (Light / Dark Mode)**:
    - Primary: `#2563EB` (Light) / `#60A5FA` (Dark) — 주요 액션 및 검증된 직접 설치 CTA
    - Trust Verified Green: `#059669` / `#34D399` — 서명 검증 및 보안 스캔 통과 배지
    - Warning Amber: `#D97706` / `#FBBF24` — 고위험 권한 포함 또는 서명 불일치 주의
    - Danger Red: `#DC2626` / `#F87171` — 악성/차단/실패 경고
    - Surface Background: `#F8FAFC` (Light) / `#0F172A` (Dark)
    - Elevated Card: `#FFFFFF` (Light) / `#1E293B` (Dark)
  - **Typography**: 본문 `Pretendard Variable`, `'Inter'`, `'Noto Sans KR'`, 코드/해시/패키지명 전용 고정폭 폰트 `'JetBrains Mono'` (SHA-256 지문 및 `versionCode` 가독성 극대화).
  - **8px Spatial Grid**: `4px (xs)`, `8px (sm)`, `16px (md)`, `24px (lg)`, `32px (xl)`, `48px (2xl)`.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- 상세 화면별 컴포넌트 트리, Props, 상태, API 호출 시퀀스는 [05_FRONTEND_UXUI/USER_WEB.md](../05_FRONTEND_UXUI/USER_WEB.md) 및 [05_FRONTEND_UXUI/SCREEN_INVENTORY.md](../05_FRONTEND_UXUI/SCREEN_INVENTORY.md)에 전수 정의되어 있다.

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- 화면 렌더링에 필요한 데이터는 과도한 Waterfall 요청이 발생하지 않도록 페이지 단위 Composite DTO(`AppDetailCompositeResponse`)로 묶어 1회의 SSR API 호출로 제공한다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- 사용자의 다크모드 설정(`SYSTEM`, `LIGHT`, `DARK`), 선호 언어, 접근성 고대비 모드 설정은 `users.preferences JSONB` 및 클라이언트 쿠키(`bn_theme`)에 저장하여 SSR 첫 프레임부터 깜빡임(FOUC) 없이 렌더링한다.

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `GET /api/v1/ui/home`: 메인 화면 배너, 인기 차트, 신규 출시, 에디터 추천 컬렉션 통합 응답.
- `GET /api/v1/catalog/apps/{slug}`: 앱 상세 화면의 모든 탭(개요, 미디어, 버전 기록, 권한/보안, 배포처/가격 비교, 개발자 정보) 통합 응답.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **모든 화면의 오류 및 빈 상태 표준화**:
  - 네트워크 단절 시: 오프라인 감지 토스트 + 캐시된 마지막 조회 데이터 표시 + `다시 시도` 버튼.
  - 검색/필터 결과 0건 시: 적용된 필터 칩 목록을 보여주고 `필터 1개씩 해제` 또는 `전체 필터 초기화` 원클릭 액션 제공.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 외부 공식 설치 링크(`USR-16`) 이동 시, 피싱 방지를 위해 이동 대상 도메인(예: `play.google.com`)과 외부 스토어 정책 적용 안내를 명확히 표시하고 `rel="noopener noreferrer nofollow"` 속성으로 새 탭 열기를 수행한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- PC(12-column Grid, 최대 너비 `1280px` + 고정 우측 설치/비교 요약 스티키 패널)와 Mobile(단일 컬럼 + 하단 고정 `Sticky Install Bar`) 레이아웃을 분리 최적화한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **가짜 설치 버튼 부재 감사 테스트 (`TC-UI-003`)**: `distribution_mode = EXTERNAL_LINK_ONLY`인 앱 상세 페이지 렌더링 시 DOM 내에 직접 다운로드 API를 호출하는 버튼이 0개이고 외부 링크 이동 안내만 표시되는지 E2E 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `F — Frontend Architecture`, `K — Kotlin Android Market`, `C — Catalog Engine`과 직접 연결된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0`
- **개발 Phase**: `Phase 2`에서 전체 화면 골격 및 디자인 시스템을 구축하고 `Phase 3~7`에서 도메인 기능을 연결한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 앱 아이콘 또는 스크린샷 CDN 이미지 로드 실패(404)** → 깨진 이미지 아이콘 대신 앱 이름 첫 글자와 카테고리 컬러를 조합한 SVG Fallback Avatar(`BnAppIconFallback.vue`)를 즉시 렌더링한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 76개 전체 화면 컴포넌트(`web-user`, `web-dev`, `web-admin`, `android-client`) 및 스토리북/디자인 시스템 토큰 패키지.

## 17. Acceptance Criteria (수용 기준)
1. 제4조의 20개 사용자 웹 화면, 제6조의 개발자 콘솔 화면, 제7조의 관리자 콘솔 화면이 단 하나도 누락 없이 실제 라우트와 API에 연결되어야 한다.
2. 라이트모드와 다크모드 전환 시 FOUC(깜빡임)가 없어야 하며 모든 텍스트가 WCAG AA 명도 대비(4.5:1)를 충족해야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-WEB-001/screen_inventory_e2e_report.json`: 20개 사용자 웹 화면 전체 라우트 렌더링 및 상태 전환 E2E 테스트 리포트.

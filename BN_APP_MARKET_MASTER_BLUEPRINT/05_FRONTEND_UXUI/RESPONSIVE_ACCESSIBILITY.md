# RESPONSIVE_ACCESSIBILITY.md — 반응형 레이아웃, WCAG 2.1 AA 접근성 및 SEO 구조 설계

- **문서 ID**: `DOC-UX-006`
- **관련 문서**: [USER_WEB.md](USER_WEB.md), [DESIGN_SYSTEM.md](DESIGN_SYSTEM.md)

---

## 1. 반응형 브레이크포인트 및 그리드 규칙

| 브레이크포인트 | 해상도 범위 | 그리드 컬럼 | 레이아웃 전환 규칙 |
| :--- | :---: | :---: | :--- |
| **Mobile (`sm`)** | `320px ~ 767px` | 4 Columns (`16px` gutter) | 하단 고정 네비게이션 바 + 앱 상세 하단 고정 `Sticky Install CTA Bar`, 비교 표는 가로 스크롤 카드 스와이프로 전환 |
| **Tablet (`md`)** | `768px ~ 1023px` | 8 Columns (`20px` gutter) | 좌측 접이식 드로어 메뉴 + 2열/3열 앱 카드 그리드 |
| **Desktop (`lg`)** | `1024px ~ 1439px` | 12 Columns (`24px` gutter) | 상단 GNB + 앱 상세 좌측 본문(8 col) / 우측 고정 설치·배포처 비교 패널(4 col) |
| **Wide (`xl`)** | `1440px+` | 12 Columns (Max `1360px`) | 관리자 콘솔 다중 패널(심사 목록 + 바이너리 분석 리포트 + 미리보기 3단 분할 뷰) |

---

## 2. 웹 접근성 (WCAG 2.1 Level AA) 필수 체크리스트

1. **키보드 내비게이션 100% 보장**: 마우스 없이 `Tab`, `Shift+Tab`, `Arrow Keys`, `Enter`, `Space`, `Escape`만으로 검색 자동완성 선택, 모달 닫기, 배포처 비교 탭 전환, 스크린샷 갤러리 탐색이 가능해야 하며 포커스 링(`outline: 2px solid var(--bn-primary)`)을 절대 제거하지 않는다.
2. **스크린 리더(ARIA) 지원**: 다운로드 진행률 바에 `role="progressbar" aria-valuenow="45" aria-valuemin="0" aria-valuemax="100"` 및 상태 전환 알림용 `aria-live="polite"` 리전을 제공한다.
3. **명도 대비 및 색약 배려**: 정상/경고/위험 상태를 색상만으로 구분하지 않고 반드시 아이콘(`✓`, `⚠`, `✕`)과 텍스트 레이블을 함께 표기한다.

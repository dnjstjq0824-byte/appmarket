# USER_WEB.md — 사용자 마켓 웹사이트(Nuxt 3 + Vue 3 + TypeScript) 전체 UX/UI 상세 설계

- **문서 ID**: `DOC-UX-001`
- **관련 조항**: 제4조 (사용자 웹사이트 전체 UX/UI 설계)
- **연관 Work Order**: `AM-WEB-001`, `AM-WEB-002`

---

## 1. 기존 앱스토어 장단점 비교 및 당사 고유 UX 설계 원칙 (제4조 준수)

| 분석 대상 | 주요 장점 | 주요 한계점 | BN App Market 고유 UX 해법 |
| :--- | :--- | :--- | :--- |
| **Google Play 웹** | 고해상도 미디어, 직관적인 평점/리뷰, 친숙한 카드 레이아웃 | APK 직접 다운로드 불가, 서명 인증서 지문 비공개, 이전 버전 롤백 다운로드 불가, 타 배포처 가격 비교 불가 | 미디어/카드 사용성은 계승하되 **배포처·가격·버전 비교 탭(`Universal Source Switcher`)**과 **서명/해시 투명성 카드**를 본문에 통합 |
| **F-Droid 웹** | 서명 인증서 SHA-256, 요구 권한, 오픈소스 라이선스, 버전별 APK 아카이브 투명 공개 | 텍스트 위주의 투박한 레이아웃, 다국어 검색 자동완성 및 상용 유료 앱 결제 UX 부재 | F-Droid의 암호학적 투명성 정보를 모던한 아코디언 배지와 원클릭 해시 복사 컴포넌트로 시각화 |
| **외부 APK 미러 사이트** | 버전별 파일 접근 용이 | 출처 불분명, 가짜 다운로드 버튼 광고(Dark Pattern), 라이선스 위반 재배포 | **가짜 설치 버튼 전면 금지**. 권리 등급(`rights_level`)에 따라 `직접 APK 다운로드`와 `외부 공식 스토어 이동 ↗`을 시각적·기능적으로 완벽히 분리 |

---

## 2. 제4조 필수 20개 화면 상세 컴포넌트·상태·API 흐름 명세

| 화면 ID | 화면명 | Nuxt 라우트 경로 | 핵심 하위 컴포넌트 구조 | 연동 API 엔드포인트 | 로딩/빈 상태/실패 처리 및 SEO |
| :---: | :--- | :--- | :--- | :--- | :--- |
| `USR-01` | **앱스토어 메인** | `/` | `HeroSpotlightCarousel`, `TopChartsPreview`, `NewReleasesRow`, `EditorialCollectionsGrid`, `AndroidClientPromoBanner` | `GET /api/v1/ui/home` | SSR + ISR(60s), 부분 위젯 실패 시 섹션 단위 폴백, `WebSite` SearchAction JSON-LD |
| `USR-02` | **앱 검색 결과** | `/search?q=...` | `SearchFacetSidebar`, `SponsoredAppCard`(최대 2개), `SearchResultAppCard`, `SpellingSuggestionBanner` | `GET /api/v1/search/apps`, `GET /api/v1/ads/serve` | 0건 시 오타 교정 및 필터 해제 칩 제공, `noindex, follow` (파라미터 조합 폭발 방지) |
| `USR-03` | **검색 자동완성** | Global Header Overlay | `GlobalSearchBar`, `InstantQuerySuggestions`, `InstantAppHitList` | `GET /api/v1/search/suggest` | 150ms 디바운스, 방향키/Enter/Esc 키보드 완전 제어, `p95 < 50ms` |
| `USR-04` | **카테고리 허브** | `/categories` | `CategoryGroupTabs`(`APPS`/`GAMES`), `CategoryTileGrid` | `GET /api/v1/catalog/categories` | SSR + ISR(300s), 각 카테고리별 등록 앱 수 표시 |
| `USR-05` | **카테고리별 인기 앱** | `/categories/[id]/charts` | `ChartTypeSwitcher`(`TOP_FREE`/`TOP_PAID`/`TRENDING`), `RankedAppList` | `GET /api/v1/catalog/charts/{type}?category={id}` | SSR + ISR(60s), 순위 변동폭(`▲2`, `NEW`) 표시 |
| `USR-06` | **신규 출시 앱** | `/new-releases` | `NewReleaseFilterBar`, `NewNoteworthyGrid` | `GET /api/v1/catalog/charts/NEW_NOTEWORTHY` | 출시 30일 이내 검증 완료 앱 노출 |
| `USR-07` | **추천 앱 (에디터 컬렉션)** | `/collections/[slug]` | `CollectionHeroBanner`, `CuratedAppCardList` | `GET /api/v1/catalog/collections/{slug}` | 테마별 에디터 추천 사유 표시 |
| `USR-08` | **앱 상세정보 (종합)** | `/apps/[slug]` | `AppHeaderSummary`, `HonestInstallActionBox`, `TrustTransparencyBadge`, `AppDetailTabs`, `SimilarAppsCarousel` | `GET /api/v1/catalog/apps/{slug}` | SSR + `SoftwareApplication` JSON-LD, 국가 차단 시 `451` 전용 안내 뷰 |
| `USR-09` | **앱 스크린샷 및 소개 영상** | `/apps/[slug]#media` (Modal) | `MediaGalleryStrip`, `LightboxZoomViewer`, `PrivacySafeVideoPlayer` | 상세 API 포함 | 키보드 좌우 화살표 탐색, 이미지 `alt` 텍스트 필수 |
| `USR-10` | **앱 버전 기록** | `/apps/[slug]/versions` | `VersionTimelineTable`, `VersionDiffBadge`, `ArchiveApkDownloadBtn` | `GET /api/v1/catalog/apps/{slug}/versions` | 각 버전별 `versionCode`, `minSdk`, `SHA-256`, 서명 지문, 릴리스 노트 표시 |
| `USR-11` | **앱 권한 및 보안 정보** | `/apps/[slug]/permissions` | `SecurityScanSummaryCard`, `PermissionRiskGroupList`(`HIGH`/`NORMAL`), `SignerCertificateCard` | 상세 API 포함 | 고위험 권한에 황색 주의 배지 및 쉬운 설명 제공, `"100% 안전"` 과장 문구 금지 |
| `USR-12` | **개인정보 및 개발자 정보** | `/apps/[slug]/developer` | `DeveloperKybVerifiedBadge`, `PrivacyPolicyLinkCard`, `SameDeveloperAppsGrid` | 상세 API 포함 | 법인/도메인 검증 여부(`verified_domain`), 고객지원 이메일, 개인정보처리방침 외부 링크 표시 |
| `USR-13` | **앱 리뷰 및 평점** | `/apps/[slug]/reviews` | `RatingDistributionBars`, `VerifiedReviewFilter`, `ReviewWriteForm`, `DeveloperReplyBox` | `GET/POST /api/v1/catalog/apps/{slug}/reviews` | 미설치 계정은 작성 폼 비활성화 및 `'설치한 사용자만 작성 가능'` 안내 |
| `USR-14` | **배포처 비교** | `/apps/[slug]/sources` | `DistributionComparisonMatrix`, `SignatureCompatibilityAlert` | `GET /api/v1/catalog/apps/{slug}/sources` | `BN_HOSTED`, `FDROID_REPO`, `GOOGLE_PLAY_LINK` 간 최신 버전, SDK, 서명 호환성 비교표 |
| `USR-15` | **가격 비교** | `/apps/[slug]/pricing` | `SourcePriceComparisonTable`, `PromoCodeRedeemBox` | 상세 API + `POST /api/v1/commerce/quotes` | 배포처별 가격, 부가세 포함 여부, 인앱결제 유무 비교 |
| `USR-16` | **공식 설치 링크 이동** | `/apps/[slug]/outlink` | `ExternalStoreInterstitialCard`, `ProvenanceTimestampNotice` | 외부 `official_store_url` 연결 | 외부 스토어 도메인, 마지막 동기화 시각 명시 후 안전하게 새 탭/딥링크 이동 |
| `USR-17` | **직접 다운로드 및 설치 안내** | `/apps/[slug]/download` | `SignedDownloadProgressCard`, `Sha256VerifyGuide`, `OpenInAndroidAppButton` | `POST .../download-token` | 15분 유효 서명 URL 발급, Android 웹 접속 시 `'BN 마켓 앱으로 바로 설치'` `bnmarket://` 인텐트 병행 제공 |
| `USR-18` | **사용자 다운로드/구매 기록** | `/account/library` | `PurchasedEntitlementsTab`, `DownloadHistoryTable`, `RefundRequestAction` | `GET /api/v1/users/me/entitlements` | `ssr: false`, 유료 구매 앱 영구 재다운로드 및 영수증 출력 |
| `USR-19` | **계정 및 보안 관리** | `/account/settings` | `ProfileForm`, `MfaTotpSetupCard`, `ActiveSessionsTable`, `ConnectedOidcList` | `GET/PATCH /api/v1/users/me` | TOTP QR 설정, 복구 코드 발급, 타 기기 세션 강제 로그아웃 |
| `USR-20` | **신고 및 고객지원** | `/support/report` | `AbuseReportForm`, `DmcaNoticeGuide`, `MySubmittedReportsList` | `POST /api/v1/moderation/reports` | 악성앱/사칭/저작권/결제분쟁 신고 접수 및 접수 번호 발급 |

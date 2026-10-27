# DEVELOPER_CONSOLE.md — 개발자 콘솔(Vue 3 + TypeScript) 24개 핵심 영역 상세 설계

- **문서 ID**: `DOC-UX-002`
- **관련 조항**: 제6조 (개발자 콘솔 상세 설계)
- **연관 Work Order**: `AM-DEV-001`, `AM-DEV-002`

---

## 1. 제6조 필수 24대 개발자 콘솔 영역별 화면·DB·API·권한·검증 매트릭스

| 번호 | 개발자 콘솔 영역 | Vue 뷰 컴포넌트 (`web-dev/src/views/`) | 연동 DB 테이블 | 연동 API 엔드포인트 | 허용 조직 역할 (`org_role`) | 상태 처리 및 서버 측 검증 기준 |
| :---: | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | **개발자 인증 (KYB)** | `KybVerificationView.vue` | `developer_orgs`, `developer_verifications` | `POST /api/v1/developer/orgs/{orgId}/verification` | `DEV_OWNER` | 사업자등록증/신분증 Vault 업로드 + DNS TXT 토큰 확인 후 `PENDING_REVIEW` 전이 |
| 2 | **조직 및 팀 관리** | `OrgSettingsView.vue` | `developer_orgs` | `POST/PATCH /api/v1/developer/orgs` | `DEV_OWNER` | 조직 슬러그 중복 검사, 지원 이메일 필수 검증 |
| 3 | **앱 신규 등록** | `CreateAppView.vue` | `catalog_apps`, `app_distributions` | `POST /api/v1/developer/orgs/{orgId}/apps` | `DEV_OWNER`, `DEV_RELEASE_MGR` | 패키지명 형식(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`) 및 보호 브랜드 선점 충돌 검사 |
| 4 | **앱 정보 편집** | `AppMetadataEditorView.vue` | `catalog_apps`, `app_localizations` | `PUT /api/v1/developer/apps/{appId}/metadata` | `DEV_OWNER`, `DEV_RELEASE_MGR`, `DEV_EDITOR` | 설명 HTML 살균(`bluemonday`), 개인정보처리방침 HTTPS URL 유효성 검사 |
| 5 | **APK 업로드** | `ApkUploaderView.vue` | `upload_sessions`, `release_artifacts` | `POST .../uploads`, `POST .../complete` | `DEV_OWNER`, `DEV_RELEASE_MGR` | Web Worker 사전 SHA-256 계산, 최대 2GB 제한, 격리 버킷 멀티파트 업로드 |
| 6 | **AAB 등록 및 변환** | `AabBundleManagerView.vue` | `upload_sessions`, `release_artifacts` | `POST .../uploads` (`artifact_type=AAB`) | `DEV_OWNER`, `DEV_RELEASE_MGR` | 최대 4GB 제한, `bundletool` APK 세트 빌드 상태 실시간 표시 |
| 7 | **서명 인증서 확인** | `SigningLineageView.vue` | `apk_signatures` | `GET /api/v1/developer/apps/{appId}/signatures` | 전체 조직원 조회 가능 | 최초 등록된 X.509 인증서 SHA-256 지문 고정 표시 및 v3 키 회전 증명 검증 |
| 8 | **메타데이터·스크린샷** | `MediaAssetsManagerView.vue` | `app_media_assets` | `POST/DELETE /api/v1/developer/apps/{appId}/media` | `DEV_OWNER`, `DEV_RELEASE_MGR`, `DEV_EDITOR` | 아이콘 512×512, 스크린샷 최소 2장~최대 8장, EXIF 제거 및 WebP 변환 |
| 9 | **국가별 배포 설정** | `CountryAvailabilityView.vue` | `app_country_availability` | `PUT /api/v1/developer/apps/{appId}/countries` | `DEV_OWNER`, `DEV_RELEASE_MGR` | 법적 차단(`is_blocked_by_law=true`) 국가는 체크박스 잠금(`DISABLED`) 처리 |
| 10 | **앱 심사 제출** | `SubmitReviewView.vue` | `app_releases`, `release_reviews` | `POST .../releases/{releaseId}/submit` | `DEV_OWNER`, `DEV_RELEASE_MGR` | 조직 `VERIFIED` 상태 + 아티팩트 `VERIFIED_READY` + Step-up MFA 코드 필수 |
| 11 | **심사 결과 확인** | `ReviewStatusTimelineView.vue` | `release_reviews`, `release_status_history` | `GET /api/v1/developer/apps/{appId}/releases/{id}` | 전체 조직원 | 단계별 진행 상태(`IN_REVIEW`, `APPROVED`, `REJECTED`) 타임라인 표시 |
| 12 | **반려 사유 및 재제출** | `RejectionFeedbackView.vue` | `release_reviews`, `appeals` | `GET .../reviews`, `POST .../appeals` | `DEV_OWNER`, `DEV_RELEASE_MGR` | 위반 정책 코드, 심사자 피드백 열람, 수정 후 재제출 또는 이의신청 연동 |
| 13 | **앱 출시 (Publish)** | `ReleasePublishView.vue` | `app_releases` | `POST .../releases/{releaseId}/publish` | `DEV_OWNER`, `DEV_RELEASE_MGR` | `APPROVED_PENDING_PUBLISH` 릴리스를 즉시 또는 예약 시각에 공개 CDN 배포 |
| 14 | **버전 업데이트** | `NewVersionWizardView.vue` | `app_releases`, `release_artifacts` | `POST /api/v1/developer/apps/{appId}/releases` | `DEV_OWNER`, `DEV_RELEASE_MGR` | `versionCode > previous_max_version_code` 및 동일 서명 키 강제 검증 |
| 15 | **단계적 배포 제어** | `StagedRolloutView.vue` | `app_releases` | `PATCH .../releases/{releaseId}/rollout` | `DEV_OWNER`, `DEV_RELEASE_MGR` | 1% → 5% → 20% → 50% → 100% 슬라이더 조절 및 크래시율 모니터링 |
| 16 | **배포 중단 및 롤백** | `EmergencyHaltRollbackView.vue` | `app_releases` | `PATCH .../releases/{releaseId}/rollout` (`HALT`) | `DEV_OWNER`, `DEV_RELEASE_MGR` | 1클릭 `HALTED` 전환으로 신규 기기 확산 즉시 중단 및 직전 안정 버전 폴백 |
| 17 | **사용자 리뷰 확인** | `UserReviewsListView.vue` | `app_reviews_user` | `GET /api/v1/catalog/apps/{slug}/reviews` | 전체 조직원 | 버전별·별점별·미답변 필터링 조회 |
| 18 | **리뷰 공식 답변** | `ReviewReplyModal.vue` | `app_reviews_user` | `PUT /api/v1/developer/apps/{appId}/reviews/{id}/reply` | `DEV_OWNER`, `DEV_SUPPORT`, `DEV_EDITOR` | 답변 작성/수정 시 작성자에게 인앱 알림 발송 및 감사 로그 기록 |
| 19 | **설치·다운로드 분석** | `InstallAnalyticsView.vue` | `app_metrics_daily`, `install_events` | `GET /api/v1/developer/apps/{appId}/analytics` | 전체 조직원 | 노출→상세조회→다운로드→설치성공 전환율 퍼널 및 국가/버전별 차트 |
| 20 | **구매 및 환불 내역** | `OrdersAndRefundsView.vue` | `orders`, `payments`, `refunds` | `GET /api/v1/developer/orgs/{orgId}/orders` | `DEV_OWNER`, `DEV_FINANCE` | 주문 상태, 부가세, 수수료, 환불 사유 조회 |
| 21 | **수익·정산 대시보드** | `SettlementsDashboardView.vue` | `settlements`, `ledger_entries`, `developer_payout_accounts` | `GET /api/v1/developer/orgs/{orgId}/settlements` | `DEV_OWNER`, `DEV_FINANCE` | 에스크로 보류액, 정산 가능 잔액, 월별 정산 명세서 PDF/CSV 다운로드 |
| 22 | **API 키 및 CI/CD 연동** | `DeveloperApiKeysView.vue` | `developer_api_keys` | `POST/DELETE /api/v1/developer/orgs/{orgId}/api-keys` | `DEV_OWNER` | 앱별·권한별 스코프 제한 키 발급(원문 1회 표시), 만료일 설정, 즉시 폐기 |
| 23 | **팀원 권한 관리** | `TeamMembersRbacView.vue` | `developer_org_members` | `POST/PATCH/DELETE .../orgs/{orgId}/members` | `DEV_OWNER` | 이메일 초대, 5개 역할 변경, 퇴사자 즉시 접근 차단 |
| 24 | **개발자 고객지원/이의신청** | `DeveloperSupportAppealsView.vue` | `appeals` | `POST /api/v1/developer/orgs/{orgId}/appeals` | `DEV_OWNER`, `DEV_SUPPORT` | 정책 문의 및 반려/제재 처분에 대한 공식 재심 청구 |

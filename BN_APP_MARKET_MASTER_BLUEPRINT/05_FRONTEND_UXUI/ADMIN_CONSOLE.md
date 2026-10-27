# ADMIN_CONSOLE.md — 관리자 콘솔(Vue 3 + TypeScript) 20개 운영 도메인 상세 설계

- **문서 ID**: `DOC-UX-003`
- **관련 조항**: 제7조 (관리자 콘솔 상세 설계)
- **연관 Work Order**: `AM-ADMIN-001`, `AM-ADMIN-002`, `AM-ADMIN-003`

---

## 1. 제7조 필수 20대 관리자 운영 영역 및 최소 권한·이중 승인 매트릭스

| 영역 번호 | 운영 영역 명칭 | Vue 뷰 컴포넌트 (`web-admin/src/views/`) | 최소 요구 권한 (`role_code`) | 감사 로그 | 이중 승인 (`Dual-Control`) | 핵심 운영 기능 및 검증 기준 |
| :---: | :--- | :--- | :--- | :---: | :---: | :--- |
| `ADM-01` | **전체 운영 대시보드** | `AdminOverviewDashboard.vue` | 전 관리자 (`ADMIN_*`) | X | X | 실시간 RPS, 심사 대기열 건수/SLA, 악성코드 격리 경보, 일일 거래액 종합 관제 |
| `ADM-02` | **회원 및 계정 관리** | `UserAccountsAdminView.vue` | `ADMIN_MODERATOR`, `ADMIN_SUPER` | **필수** | X | 비정상 계정 잠금/해제, 활성 세션 강제 종료, 성인인증 상태 확인 |
| `ADM-03` | **개발자 신원 검증 (KYB)** | `DeveloperKybReviewView.vue` | `ADMIN_KYB_REVIEWER`, `ADMIN_SUPER` | **필수** | X | KMS Vault 서류 일회성 보안 뷰어 열람, D-U-N-S 및 DNS TXT 검증 확인 후 승인/반려 |
| `ADM-04` | **앱 등록 심사** | `AppMetadataReviewView.vue` | `ADMIN_APP_REVIEWER`, `ADMIN_SUPER` | **필수** | X | 다국어 설명, 스크린샷, 개인정보처리방침, 연령 등급 적정성 체크리스트 심사 |
| `ADM-05` | **APK 보안 심사** | `ApkSecurityReviewView.vue` | `ADMIN_SECURITY`, `ADMIN_SUPER` | **필수** | X | ClamAV/YARA 매칭 룰, 고위험 권한 Diff, 사칭 유사도 점수(`risk_score`) 심층 분석 |
| `ADM-06` | **앱 출시 승인** | `ReleaseApprovalQueueView.vue` | `ADMIN_APP_REVIEWER`, `ADMIN_SECURITY`, `ADMIN_SUPER` | **필수** | 고위험앱 O | 심사 통과 릴리스를 `APPROVED_PENDING_PUBLISH`로 승인 및 Step-up MFA 검증 |
| `ADM-07` | **앱 배포 중단 (Kill Switch)** | `EmergencyTakedownView.vue` | `ADMIN_SECURITY`, `ADMIN_SUPER` | **필수** | 영구차단시 O | 1클릭 배포 즉시 중단, 공개 CDN 버킷 객체 차단/퍼지, 기설치자 보안 삭제 알림 발송 |
| `ADM-08` | **신고 및 사용자 보호** | `ModerationReportsQueueView.vue` | `ADMIN_MODERATOR`, `ADMIN_SECURITY` | **필수** | X | 악성/사칭/저작권(DMCA) 신고 우선순위(`P0`~`P2`) 처리 및 권리자 회신 |
| `ADM-09` | **리뷰·평점 운영** | `ReviewsModerationView.vue` | `ADMIN_MODERATOR` | **필수** | X | 어뷰징 탐지 보류(`QUARANTINED_AUTO_FRAUD`) 리뷰 일괄 블라인드 또는 정상 복원 |
| `ADM-10` | **외부 스토어 커넥터 관리** | `ExternalConnectorsAdminView.vue` | `ADMIN_SEARCH_OPS`, `ADMIN_SUPER` | **필수** | X | F-Droid 등 외부 커넥터 활성화 토글, 동기화 스케줄 조정, 수동 동기화 트리거 |
| `ADM-11` | **검색 랭킹 운영** | `SearchRankingOpsView.vue` | `ADMIN_SEARCH_OPS` | **필수** | X | 동의어 사전(`synonyms`), 금칙어, 에디터 추천 컬렉션 편성, Blue-Green 재색인 실행 |
| `ADM-12` | **광고 심사 및 관제** | `AdCampaignReviewView.vue` | `ADMIN_MODERATOR`, `ADMIN_FINANCE` | **필수** | X | 스폰서 앱 캠페인 타기팅/소재 승인 및 부정클릭(Click Fraud) 차단 통계 조회 |
| `ADM-13` | **결제·환불 관리** | `OrdersRefundsAdminView.vue` | `ADMIN_FINANCE`, `ADMIN_SUPER` | **필수** | 고액환불시 O | 주문/PG 웹훅 상태 대사, 분쟁 환불 강제 승인 및 구매 권한(`user_entitlements`) 회수 |
| `ADM-14` | **개발자 정산 관리** | `SettlementPayoutAdminView.vue` | `ADMIN_FINANCE`, `ADMIN_SUPER` | **필수** | **필수 (2인 승인)** | 월간 정산 원장 검증, 컴플라이언스 지급 보류/해제, 최종 은행 이체 배치 승인 |
| `ADM-15` | **통계·분석 센터** | `PlatformAnalyticsAdminView.vue` | 전 관리자 조회 가능 | X | X | 전체 검색 전환율, 배포처별 클릭 비중, 카테고리별 매출 및 리텐션 리포트 |
| `ADM-16` | **다국어·국가별 정책** | `CountryPoliciesAdminView.vue` | `ADMIN_SUPER`, `ADMIN_MODERATOR` | **필수** | X | 국가별 부가세율(`tax_rate_bps`), 성인인증 의무화, 특정 앱 국가별 법적 차단 설정 |
| `ADM-17` | **관리자 권한 관리** | `AdminRbacDualControlView.vue` | `ADMIN_SUPER` | **필수** | **필수 (2인 승인)** | 관리자 역할 부여/회수, MFA 강제 초기화, 이중 승인 대기열(`dual_approval_requests`) 처리 |
| `ADM-18` | **시스템 설정** | `SystemSettingsAdminView.vue` | `ADMIN_SUPER` | **필수** | 수수료변경 O | 수수료 티어 정책, 업로드 크기 제한, TreeScan 어댑터 Feature Flag 제어 |
| `ADM-19` | **모니터링 및 감사 로그** | `AuditLogChainExplorerView.vue` | `ADMIN_AUDITOR`, `ADMIN_SECURITY`, `ADMIN_SUPER` | 열람기록 | X | 비동기 워커/DLQ 관제, 불변 감사 로그 검색 및 해시 체인(`prev_hash -> record_hash`) 무결성 검증 |
| `ADM-20` | **백업·복구 및 운영 점검** | `BackupRecoveryDrillView.vue` | `ADMIN_SUPER` | **필수** | X | PostgreSQL WAL/스냅샷 및 OpenSearch 백업 상태 점검, 읽기 전용 점검 모드 전환 |

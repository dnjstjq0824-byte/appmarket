# TABLE_DICTIONARY.md — 전체 34개 테이블 및 컬럼 상세 데이터 사전

- **문서 ID**: `DOC-DB-002`
- **관련 문서**: [ERD.md](ERD.md), [0001_initial_schema.up.sql](migrations/0001_initial_schema.up.sql)

---

## 1. 도메인별 전체 테이블 목록 및 설계 근거 (34 Tables Across 19 Domains)

| 번호 | 도메인 영역 | 테이블명 | PK 타입 | 핵심 역할 및 데이터 타입 선정 근거 |
| :---: | :--- | :--- | :--- | :--- |
| 1 | 회원·인증 | `users` | `UUID (v7)` | 통합 사용자 계정. 이메일 평문 외에 `email_blind_index CHAR(64)`(HMAC-SHA256)를 두어 인덱스 검색 성능을 유지한다. |
| 2 | 회원·인증 | `user_identities` | `UUID` | Google, GitHub, Apple 및 미래 TreeScan OIDC(`TREESCAN_OIDC`) 외부 IdP 주체(`provider_subject`) 매핑. |
| 3 | 회원·인증 | `auth_sessions` | `UUID` | Refresh Token 원문 대신 `refresh_token_sha256 CHAR(64)`만 저장하며 `token_family_id`로 탈취 재사용을 감지한다. |
| 4 | 회원·인증 | `mfa_credentials` | `UUID` | TOTP 시크릿은 KMS로 암호화된 `secret_ciphertext BYTEA`로 저장한다. |
| 5 | 회원·인증 | `roles` | `VARCHAR(48)` | 플랫폼 및 조직 역할 코드(`USER`, `ADMIN_SUPER`, `ADMIN_SECURITY` 등)와 JSONB 권한 배열 정의. |
| 6 | 회원·인증 | `user_roles` | `(user_id, role_code)` | 사용자-역할 다대다 매핑 및 권한 부여자(`granted_by`) 감사 추적. |
| 7 | 개발자·조직 | `developer_orgs` | `UUID` | 개인/법인/오픈소스 개발자 조직. `verification_status`와 `commission_tier_code`, `trust_score(0~100)` 보유. |
| 8 | 개발자·조직 | `developer_verifications` | `UUID` | 사업자등록번호(`BYTEA` 암호화), D-U-N-S, 신원 서류 S3 Vault 키, 역방향 DNS TXT 검증 토큰 관리. |
| 9 | 개발자·조직 | `developer_org_members` | `(org_id, user_id)` | 조직 내 5대 역할(`DEV_OWNER`, `DEV_RELEASE_MGR`, `DEV_EDITOR`, `DEV_FINANCE`, `DEV_SUPPORT`) RBAC 매핑. |
| 10 | 개발자·조직 | `developer_api_keys` | `UUID` | CI/CD 자동 배포용 API 키. 원문은 저장하지 않고 `key_prefix CHAR(12)`와 `key_sha256 CHAR(64)`만 저장. |
| 11 | 개발자·조직 | `developer_payout_accounts` | `UUID` | 정산 계좌번호(`account_number_ciphertext BYTEA`) 및 세무 식별번호 암호화 저장. |
| 12 | 앱 카탈로그 | `categories` | `VARCHAR(48)` | `APPS` / `GAMES` 대분류 및 부모-자식 계층형 카테고리, `name_i18n JSONB`. |
| 13 | 앱 카탈로그 | `tags` | `VARCHAR(48)` | 시맨틱 검색 및 필터용 다국어 태그 사전. |
| 14 | 앱 카탈로그 | `catalog_apps` | `UUID` | 논리적 앱 상품(Canonical App). `canonical_slug`, `monetization_type`, `global_age_rating`, `bayesian_rating NUMERIC(4,3)` 저장. |
| 15 | 앱 카탈로그 | `app_localizations` | `UUID` | `(app_id, locale)` 유일 제약 기반 다국어 제목, 요약, 살균된 상세 설명(`full_description_sanitized`), 아이콘 경로. |
| 16 | 앱 카탈로그 | `app_media_assets` | `UUID` | 폼팩터별(`PHONE`, `TABLET_10` 등) 스크린샷 해상도(`width_px`, `height_px`) 및 `sha256_checksum` 관리. |
| 17 | 앱 카탈로그 | `app_tags` | `(app_id, tag_id)` | 앱-태그 다대다 연결 테이블. |
| 18 | 글로벌 정책 | `country_policies` | `CHAR(2)` | ISO 3166-1 alpha-2 국가별 기본 통화, 부가세율(`tax_rate_bps` 정수형 Basis Points), 성인인증 의무 여부. |
| 19 | 글로벌 정책 | `app_country_availability` | `(app_id, country_code)` | 개발자 허용 여부(`is_allowed_by_dev`)와 법적 차단 여부(`is_blocked_by_law`)를 분리 기록. |
| 20 | 배포처·버전 | `app_distributions` | `UUID` | 카탈로그 앱 하위의 배포처(`BN_HOSTED`, `FDROID_REPO`, `GOOGLE_PLAY_LINK` 등), 5단계 권리 등급(`rights_level`), 대표 서명 지문. |
| 21 | 배포처·버전 | `app_releases` | `UUID` | 트랙별(`INTERNAL`, `BETA`, `PRODUCTION`) 버전 릴리스. `version_code BIGINT`, 15개 상태값, `rollout_percentage(0~100)`. |
| 22 | 배포 파일 | `upload_sessions` | `UUID` | S3 격리 버킷 멀티파트 업로드 세션. 최대 4GB(`4294967296` bytes) `CHECK` 제약 및 사전 `expected_sha256` 고정. |
| 23 | 배포 파일 | `release_artifacts` | `UUID` | 실제 APK/AAB 바이너리 메타데이터. `min_sdk`, `target_sdk`, `native_abis TEXT[]`, `signature_scheme_version >= 2`. |
| 24 | 서명 정보 | `apk_signatures` | `UUID` | APK에서 추출한 X.509 인증서 `cert_sha256_fingerprint CHAR(64)`, Subject/Issuer DN, v3 `proof_of_rotation JSONB`. |
| 25 | 외부 앱 출처 | `external_connectors` | `UUID` | 외부 스토어 커넥터 설정, 공개키(`public_key_pem`), 동기화 주기 및 권리 등급. |
| 26 | 외부 앱 출처 | `external_sync_jobs` | `UUID` | 커넥터 동기화 실행 이력, 처리 건수, 서명 검증 실패 사유 기록. |
| 27 | 외부 앱 출처 | `external_store_items` | `UUID` | 외부 배포처 원본 패키지 메타데이터, 공식 링크(`official_store_url`), 마지막 동기화 시각. |
| 28 | 설치·검색로그 | `install_events` | `(id, created_at)` | 월별 Range Partitioning 적용. 다운로드 토큰 발급, 다운로드 완료, OS 설치 성공/실패 코드 기록. |
| 29 | 설치·검색로그 | `search_query_logs` | `(id, created_at)` | 월별 Range Partitioning 적용. 검색어, 결과 수, 클릭된 앱, 응답 지연시간(`latency_ms`) 기록. |
| 30 | 리뷰·신고·심사 | `app_reviews_user` | `UUID` | 실제 설치 검증(`is_verified_install`) 사용자 리뷰, 어뷰징 점수(`fraud_score`), 개발자 공식 답글. |
| 31 | 리뷰·신고·심사 | `moderation_reports` | `UUID` | 악성코드, 사칭, 저작권(DMCA) 등 7대 카테고리 사용자 신고 및 SLA 우선순위 큐. |
| 32 | 리뷰·신고·심사 | `release_reviews` | `UUID` | 관리자의 릴리스 심사 체크리스트 결과, 승인/반려 결정, 정책 위반 코드 기록. |
| 33 | 보안 검사 | `security_scan_reports` | `UUID` | ClamAV 결과, YARA 매칭 룰 배열, 위험 권한 목록, 사칭 유사도 점수, 종합 `risk_score(0~100)`. |
| 34 | 이의신청·상거래·운영 | `appeals`, `orders`, `payments`, `user_entitlements`, `refunds`, `ledger_*`, `settlements`, `ad_campaigns`, `app_metrics_daily`, `dual_approval_requests`, `audit_logs`, `notifications`, `system_settings`, `outbox_events`, `background_jobs`, `idempotency_keys` | `UUID` / `VARCHAR` | 모든 금액은 부동소수점 오차 방지를 위해 `BIGINT` Minor Unit(원/센트)으로 저장하며, `orders` 테이블은 `net_supply_minor + tax_minor = total_paid_minor` 및 `platform_fee_minor + developer_earning_minor = net_supply_minor`를 DB `CHECK` 제약조건으로 강제한다. |

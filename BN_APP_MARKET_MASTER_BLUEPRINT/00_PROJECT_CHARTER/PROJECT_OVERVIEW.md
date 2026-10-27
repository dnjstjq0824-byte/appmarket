# PROJECT_OVERVIEW.md — 글로벌 독립 통합 앱 마켓 플랫폼 프로젝트 헌장

- **문서 ID**: `DOC-CHARTER-001`
- **관련 요구사항 ID**: `REQ-A-001`, `REQ-B-001`, `REQ-C-001`, `REQ-E-001`, `REQ-Z-001`
- **연관 Work Order**: `AM-ARCH-001`, `AM-DB-001`, `AM-API-001`

---

## 1. 프로젝트 배경 및 목적

모바일 생태계에서 사용자는 원하는 앱을 찾기 위해 Google Play, 오픈소스 저장소(F-Droid), 제조사 스토어(Galaxy Store), 또는 개발자 공식 웹사이트를 개별적으로 방문해야 하며, 각 스토어마다 버전·가격·지원 기기·설치 방식이 파편화되어 있습니다. 동시에 독립 앱 개발자들은 단일 거대 플랫폼에 종속되지 않고 합리적인 수수료와 투명한 심사 절차를 통해 글로벌 사용자에게 직접 앱을 배포할 수 있는 상용 대안 마켓을 필요로 합니다.

**Booking Network 글로벌 독립 통합 앱 마켓 플랫폼(BN App Market)**은 이 문제를 해결하기 위해 다음 **3가지 플랫폼 유형을 단일 아키텍처로 통합**하여 구축합니다:

1. **Type 1 — Independent App Store (자체 독립 앱스토어)**
   - 개발자가 직접 개발자 콘솔(`web-dev`)에 가입하고 신원(개인/법인 KYB) 인증을 거친 뒤 APK 또는 AAB 바이너리를 업로드합니다.
   - 플랫폼은 격리된 보안 파이프라인에서 파일 무결성(SHA-256), Android 서명 인증서(v2/v3/v4), 정적 권한 및 악성코드 검사를 수행하고, 운영자 콘솔(`web-admin`)의 심사를 거쳐 승인된 버전을 자체 S3 호환 스토리지 및 글로벌 CDN을 통해 배포합니다.
   - 사용자는 웹(`web-user`) 또는 전용 Android 클라이언트 앱(`android-client`)을 통해 앱을 검색·다운로드·설치(`PackageInstaller`)하고 서명 호환성이 보장된 자동/수동 업데이트를 제공받습니다.

2. **Type 2 — Federated App Search Engine (연합형 외부 앱 검색엔진)**
   - F-Droid(공식 서명된 `index-v2.json` 오픈소스 리포지토리), 파트너십이 체결된 스토어 피드, 개발자 공식 등록 웹페이지, 그리고 Google Play 등 외부 공식 배포처의 공개 메타데이터/공식 링크 정보를 합법적 범위 내에서 수집·정규화합니다.
   - 데이터 출처(Provenance), 라이선스, 마지막 동기화 시각, 직접 다운로드 가능 여부와 외부 공식 링크 이동(`EXTERNAL_LINK_ONLY`) 여부를 엄격히 구분하여 사용자에게 투명하게 안내합니다.

3. **Type 3 — Universal App Marketplace (통합 비교 앱 마켓)**
   - 하나의 논리적 앱 엔티티(`catalog_apps`) 아래에 여러 배포처(`app_distributions`: 자체 호스팅, F-Droid, Google Play 공식 링크, Galaxy Store 공식 링크 등)를 연결합니다.
   - 사용자는 단일 상세 화면에서 배포처별 최신 버전(`version_code`), 가격/인앱결제 여부, 요구 Android 버전(`min_sdk`), APK 서명 인증서 지문(`cert_sha256`), 요구 권한을 비교하고 원하는 설치 경로를 선택할 수 있습니다.

---

## 2. 핵심 도메인 모델의 3대 동일성 분리 원칙 (제8조 준수)

본 플랫폼의 모든 데이터베이스 설계, 백엔드 API, 검색 인덱스, Android 클라이언트는 다음 **3가지 동일성(Identity)을 절대 혼동하지 않고 분리 관리**합니다:

| 구분 | 개념 정의 | 핵심 식별자 및 테이블 | 주요 규칙 |
| :--- | :--- | :--- | :--- |
| **1. 앱 카탈로그 동일성 (Catalog Identity)** | 사용자에게 보여지는 논리적 앱 상품 단위 | `catalog_apps.id` (UUID), `canonical_slug`, `primary_package_name` | 하나의 카탈로그 앱은 여러 배포처(자체 스토어, F-Droid, 외부 스토어 링크)를 가질 수 있음 |
| **2. 배포처별 패키지 동일성 (Distribution Source Identity)** | 특정 배포처에서 제공되는 패키지 및 버전 단위 | `app_distributions.id`, `source_type` (`BN_HOSTED`, `FDROID_REPO`, `GOOGLE_PLAY_LINK`, `GALAXY_STORE_LINK`, `OFFICIAL_SITE_LINK`) | 배포처마다 최신 버전(`version_code`), 가격, 다운로드 방식(직접 APK vs 외부 딥링크)이 다를 수 있음 |
| **3. 앱 서명 동일성 (Signing Certificate Identity)** | 실제 Android OS가 업데이트 호환성을 판별하는 암호학적 서명 단위 | `apk_signatures.cert_sha256_fingerprint` + `package_name` | 동일한 `package_name`이라도 배포처(예: 자체 빌드 vs F-Droid 재빌드)에 따라 서명 인증서가 다르면 Android OS에서 덮어쓰기 업데이트가 불가능함(`INSTALL_FAILED_UPDATE_INCOMPATIBLE`). 플랫폼은 이를 사전에 판별하여 사용자에게 경고함 |

---

## 3. 이해관계자 및 주요 사용자 역할 (Actors & Roles)

1. **일반 사용자 (End User - Guest / Authenticated Member)**
   - 웹 브라우저 및 Android 네이티브 마켓 앱을 통해 앱 검색, 배포처/가격/버전 비교, 무료/유료 앱 구매, APK 다운로드 및 기기 설치, 자동 업데이트, 리뷰 작성, 유해 앱 신고 수행.
2. **개발자 및 퍼블리셔 (Developer Organization - Owner / Release Manager / Finance Manager / Support Agent)**
   - 개인 또는 법인(KYB) 개발자 등록, 팀원 초대 및 RBAC 권한 부여, 앱 메타데이터 다국어 등록, APK/AAB 업로드, 단계적 출시(Staged Rollout), 롤백, 크래시/설치 통계 조회, 유료 앱 판매 수익 및 정산 내역 관리, 사용자 리뷰 답변.
3. **플랫폼 운영자 및 관리자 (Platform Admin - Super Admin / App Reviewer / Security Analyst / Trust & Safety Moderator / Finance Admin / Search Ops)**
   - 개발자 KYB 신원 심사, 앱 등록 및 업데이트 심사, APK 악성코드 검역 해제/영구 차단, 긴급 배포 중단(Kill Switch), 신고 처리, 외부 스토어 커넥터 스케줄 및 상태 관리, 검색 랭킹 부스팅/금칙어 관리, 광고 캠페인 심사, 환불 및 개발자 정산 집행(민감 작업 이중 승인 필수).
4. **외부 연동 시스템 (External Systems & Future TreeScan)**
   - OIDC 인증 제공자, PG 결제 게이트웨이(Stripe, Toss Payments 등), 외부 앱 리포지토리(F-Droid Index v2 등), 향후 통합될 Booking Network **TreeScan 검색엔진**(Phase 10 어댑터 연동).

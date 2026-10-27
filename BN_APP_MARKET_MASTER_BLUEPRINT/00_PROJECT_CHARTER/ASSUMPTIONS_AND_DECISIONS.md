# ASSUMPTIONS_AND_DECISIONS.md — 아키텍처 결정 기록(ADR) 및 외부 정책 검증 구분

- **문서 ID**: `DOC-CHARTER-004`
- **관련 조항**: 제2조, 제5조, 제10조, 제11조, 제12조, 제13조, 제21조, 제26조

---

## 1. 핵심 아키텍처 결정 기록 (Architecture Decision Records - ADR)

### ADR-001: Go 모듈형 모놀리스(Modular Monolith) + 격리 워커(Isolated Workers) 채택
- **결정**: 초기 백엔드는 단일 Go 코드베이스 내부에 도메인 패키지(`iam`, `catalog`, `developer`, `distribution`, `search`, `commerce`, `ads`, `moderation`, `admin`, `treescan`)가 엄격한 인터페이스로 분리된 **모듈형 모놀리스**로 구성하되, CPU/메모리 집약적이거나 보안 위험이 있는 **APK 파싱, 서명 검증, 악성코드 스캔, `bundletool` AAB 변환, 외부 스토어 크롤러/동기화기**는 별도 컨테이너 프로세스(`workers`)로 물리적 격리한다.
- **근거**: 제2조 지침을 준수하며, 초기 마이크로서비스 도입 시 발생하는 분산 트랜잭션 복잡도와 네트워크 오버헤드를 방지하면서도, 악성 APK 분석 중 발생할 수 있는 파서 취약점(Zip Bomb, AXML 버퍼 오버플로)으로부터 핵심 API 서버를 완벽히 보호할 수 있다.

### ADR-002: Android 클라이언트 최소 지원 버전 `minSdk = 26` (Android 8.0 Oreo), `targetSdk = 35` (Android 15)
- **결정**: Android 네이티브 마켓 클라이언트의 `minSdk`를 API Level 26 (Android 8.0)으로, `targetSdk` 및 `compileSdk`를 API Level 35 (Android 15)로 확정한다.
- **근거**:
  1. Android 8.0(API 26)부터 알 수 없는 출처 앱 설치 권한이 전역 설정에서 **앱별 `REQUEST_INSTALL_PACKAGES` 권한(`Settings.ACTION_MANAGE_UNKNOWN_APP_SOURCES`)**으로 변경되었으므로, 보안 정책 일관성을 위해 API 26 이상을 기준으로 설계한다.
  2. APK Signature Scheme v2(Android 7.0+) 및 v3(Android 9.0+) 검증과 최신 TLS 1.3 보안 통신을 안정적으로 강제할 수 있으며, 글로벌 활성 Android 기기의 97% 이상을 커버한다.
  3. 단, 마켓에 **등록되는 서드파티 앱(카탈로그 앱)** 자체는 `minSdk < 26`인 레거시 앱도 등록 가능하며, 기기 호환성 엔진이 사용자의 실제 `Build.VERSION.SDK_INT`와 비교하여 설치 가능 여부를 판별한다.

### ADR-003: APK 우선 지원 및 AAB(Android App Bundle) 무중단 확장 인터페이스
- **결정**: Phase 3~4 초기 출시에서는 개발자가 직접 서명한 **Universal APK 또는 Multi-APK(ABI/DPI별 APK)** 업로드를 기본 운영 경로로 지원하되, DB 스키마(`release_artifacts.artifact_type ENUM ('APK', 'AAB', 'APKS_SET')`)와 배포 파이프라인 인터페이스는 처음부터 `bundletool build-apks` 및 플랫폼 관리형 서명키(Managed App Signing Vault)를 수용하도록 설계한다.
- **근거**: 제10조 요구사항에 따라 AAB 도입 시 DB나 API를 재설계하지 않고 워커 플러그인만 활성화하여 즉시 기기 맞춤형 Split APK 배포를 지원할 수 있다.

### ADR-004: 복식부기(Double-Entry Ledger) 기반 결제·정산 원장 설계
- **결정**: 유료 앱 결제, 환불, 플랫폼 수수료(기본 15%, 파트너/초기 프로모션 정책에 따라 설정 가능), 세금(VAT/GST), 개발자 정산 가능 잔액은 단순 숫자 컬럼(`balance = balance + x`) 업데이트가 아닌 **불변 복식부기 원장 테이블(`ledger_entries`, `ledger_transactions`)**로 기록한다.
- **근거**: 동시 환불 요청, 웹훅 중복 수신, 부분 환불, 차지백(Chargeback) 발생 시에도 단 1원의 오차 없이 원장 합계(`SUM(debit) = SUM(credit)`)를 감사(Audit)할 수 있다.

---

## 2. 공식 문서 확인 사실(Verified Facts) vs 설계상 가정(Design Assumptions) vs 외부 정책/법률 확인 대상(Pending Legal/Business Sign-off)

제13조 및 제21조에 따라 확인된 사실과 가정, 그리고 운영 전 외부 승인이 필요한 항목을 명확히 분리합니다:

| 구분 | 항목 | 상세 내용 및 근거 | 후속 조치 주체 |
| :--- | :--- | :--- | :--- |
| **확인된 사실 (Verified Technical Fact)** | Android `PackageInstaller` 및 사용자 승인 제약 | 일반 서드파티 앱 마켓은 최초 앱 설치 시 반드시 `STATUS_PENDING_USER_ACTION`을 통해 시스템 설치 확인 창을 띄워야 함. 단, API 31+(Android 12+)에서 자신이 설치한 앱(`installerPackageName` 일치)을 업데이트할 때만 `SessionParams.setRequireUserAction(USER_ACTION_NOT_REQUIRED)` 및 `UPDATE_PACKAGES_WITHOUT_USER_ACTION` 권한으로 무음 자동 업데이트 가능 ([Android 공식 문서](https://developer.android.com/reference/android/content/pm/PackageInstaller.SessionParams)) | Android Agent가 API 레벨별 분기 코드로 구현 |
| **확인된 사실 (Verified Technical Fact)** | Google Play 배포 정책과 서드파티 앱스토어 제한 | Google Play 개발자 프로그램 정책(Device and Network Abuse)상 타 앱스토어 기능을 하는 앱은 Google Play 스토어 내에 등록할 수 없으며, `REQUEST_INSTALL_PACKAGES` 권한 또한 엄격히 제한됨. 따라서 BN App Market Android 클라이언트는 **사용자 웹(`web-user`)을 통한 직접 공식 APK 다운로드(Sideloading Bootstrap)** 및 제휴 OEM 프리로드 경로로 배포해야 함 | 배포/마케팅 및 사용자 웹 가이드 반영 완료 |
| **확인된 사실 (Verified Technical Fact)** | F-Droid 메타데이터 연동 방식 | F-Droid는 별도 REST 검색 API가 아닌 서명된 JAR(`index-v1.jar`, `index-v2.json` 및 `entry.json`) 정적 피드 구조로 전체 카탈로그와 APK 해시/서명 정보를 공개함 ([F-Droid 공식 문서](https://gitlab.com/fdroid/wiki/-/wikis/Index-V2)) | Search & Connector Agent가 주기적 Poller 및 GPG/JAR 서명 검증기로 구현 |
| **설계상 가정 (Design Assumption)** | 기본 플랫폼 수수료율 및 정산 주기 | 유료 앱 기본 판매 수수료는 15%(연 매출 100만 달러 이하 개발자 우대 티어 10% 설정 가능), 정산 보류 기간(Refund Hold Period)은 구매 확정 후 14일, 정기 정산일은 매월 15일로 가정하여 스키마 및 정산 배치를 파라미터화함 (`system_settings` 테이블에서 무중단 변경 가능) | PO가 상용 오픈 전 최종 수치 확정 |
| **외부 정책/법률 확인 대상 (Pending Legal/Compliance)** | 국가별 PG 라이선스, 전자금융업 등록 및 세무(MoR 여부) | 플랫폼이 **Merchant of Record(MoR, 판매대행 책임자)**로서 각국 부가가치세(VAT/GST)를 원천 징수할지, 아니면 중개자 모델로 운영할지는 국가별 세법 및 PG사(Stripe Connect, Toss Payments 등) 계약에 따라 달라짐. 현재 설계는 MoR과 중개자 설정을 모두 수용하도록 세금 테이블(`tax_rules`)을 분리해 두었으나, 실제 법적 효력은 법무/재무팀의 국가별 승인 전까지 `PENDING_LEGAL_REVIEW` 상태로 관리함 | Booking Network 법무·재무팀 및 PO |
| **외부 정책/법률 확인 대상 (Pending Legal/Compliance)** | Google Play / Galaxy Store 메타데이터 수집 범위 | Google Play Android Developer API는 **자신이 소유한 앱**을 관리하는 API이며, 타사 전체 앱을 검색하는 공개 API를 제공하지 않음. 따라서 Google Play/Galaxy Store 앱 정보는 ① 개발자가 직접 자사 앱의 외부 스토어 링크를 등록하거나, ② 공식 제휴/데이터 라이선스 피드, 또는 ③ 법무팀이 허용한 공개 웹 메타데이터(Schema.org `SoftwareApplication` JSON-LD) 범위 내에서만 수집하도록 커넥터 Feature Flag를 분리함 | 법무팀 승인 전까지 무단 스크래핑 비활성화(`DISABLED_BY_POLICY`) 기본값 유지 |

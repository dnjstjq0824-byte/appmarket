# SCOPE_AND_NON_GOALS.md — 프로젝트 범위 및 명시적 비목표(Non-Goals)

- **문서 ID**: `DOC-CHARTER-003`
- **관련 조항**: 제0조, 제1조, 제5조, 제11조, 제13조, 제17조

---

## 1. 포함 범위 (In-Scope)

1. **사용자 마켓 웹사이트 (`web-user`)**: 앱 탐색, 통합 검색 및 자동완성, 카테고리/인기/신규/추천 앱, 배포처 및 가격 비교, 버전/권한/서명/개인정보 상세 표시, 리뷰/평점, 직접 APK 다운로드 및 외부 공식 스토어 딥링크 연결, 구매/다운로드 내역, 신고/고객지원.
2. **Android 네이티브 클라이언트 (`android-client`)**: 기기 사양(SDK, ABI, 화면 밀도) 자동 감지, 호환성 필터링, 백그라운드 멀티스레드 다운로드(일시정지/재개/취소), SHA-256 및 서명 검증, `PackageInstaller` 기반 설치 및 일괄 업데이트, 서명 불일치 경고, 자체 클라이언트 인앱 업데이트.
3. **개발자 콘솔 (`web-dev`)**: 개인/법인(KYB) 인증, 조직 및 팀원 RBAC, 신규 앱 생성, APK 업로드(및 AAB `bundletool` 파이프라인 지원), 서명 인증서 등록 및 지문 검증, 국가별 가격/노출 설정, 단계적 출시(1%~100%), 버전 롤백, 리뷰 답변, 설치/매출 분석 및 정산 신청.
4. **관리자 콘솔 (`web-admin`)**: 통합 관제 대시보드, 회원/개발자 KYB 심사, 앱 메타데이터 및 APK 보안 심사, 출시 승인/반려, 긴급 배포 중단(Kill Switch), 신고/이의신청 처리, 외부 스토어 커넥터 관리, 검색 랭킹/스폰서 광고 운영, 환불 및 정산 이중 승인, 감사 로그 및 백업 관리.
5. **백엔드 및 배포/검색/상거래 엔진 (`backend`, `workers`)**: Go 모듈형 모놀리스 API(19개 API 그룹), PostgreSQL 16 스키마(32개 핵심 테이블), OpenSearch 통합 검색, S3 격리/공개 버킷 및 CDN 배포 파이프라인, 유료 앱 결제/환불/수수료 원장 및 스폰서 광고 엔진, TreeScan 연동 어댑터 API.

---

## 2. 명시적 비목표 및 금지 범위 (Explicit Non-Goals & Prohibited Scope)

후속 개발 에이전트는 아래 항목을 임의로 구현해서는 안 됩니다:

1. **일반 Android 기기에서의 무단 무음 설치(Silent Install) 우회 금지**
   - 루팅(Root) 권한 탈취, 비공개 시스템 API 호출, 접근성 서비스(AccessibilityService) 악용을 통한 자동 클릭 설치 등 Android 공식 보안 정책을 우회하는 기법은 절대 구현하지 않습니다.
   - 일반 사용자 기기에서는 반드시 Android 공식 `PackageInstaller` 및 `REQUEST_INSTALL_PACKAGES` 권한과 사용자 명시적 승인 인텐트(`STATUS_PENDING_USER_ACTION`)를 따릅니다. (단, Android 12+ API 31 이상의 `setRequireUserAction(USER_ACTION_NOT_REQUIRED)`를 통한 동일 스토어 소유 앱 자동 업데이트는 공식 API 스펙 범위 내에서만 지원합니다.)
2. **타사 독점 앱스토어(Google Play 등)의 비공개 API 해킹 및 APK 무단 미러링 금지**
   - Google Play 내부 비공식 프로토콜을 리버스 엔지니어링하여 타사 APK 바이너리를 무단 추출·복제·재배포하는 기능을 기본 구현으로 설계하지 않습니다.
   - 외부 앱은 ① 오픈소스/라이선스 허용 리포지토리(F-Droid 등), ② 공식 제휴 피드, ③ 개발자 직접 위임 등록, ④ 검색 안내 및 공식 배포처 링크(`EXTERNAL_LINK_ONLY`)로만 처리합니다.
3. **TreeScan 본체 코드 수정 금지**
   - 현재 프로젝트 범위(Phase 0~9)에서 TreeScan 본체 저장소의 코드를 직접 수정하거나 DB를 합치지 않습니다.
4. **iOS 네이티브 서드파티 앱 직접 바이너리 설치(IPA 사이드로딩) 제외**
   - iOS는 Apple의 OS 정책 및 지역별 규제(EU DMA Alternative App Marketplace 자격 등)에 따라 별도의 Apple 승인 엔타이틀먼트 없이는 웹/외부 앱에서 임의의 IPA 설치가 불가능합니다 ([16_FINAL_DELIVERY/OFFICIAL_REFERENCES.md](../16_FINAL_DELIVERY/OFFICIAL_REFERENCES.md) 참조).
   - 따라서 본 플랫폼에서 iOS 플랫폼 정보는 Universal App Marketplace(Type 3)의 **크로스 플랫폼 메타데이터 비교 및 Apple App Store 공식 링크 연결(`EXTERNAL_LINK_ONLY`)**로만 제공하며, IPA 직접 바이너리 배포는 범위에서 제외합니다.

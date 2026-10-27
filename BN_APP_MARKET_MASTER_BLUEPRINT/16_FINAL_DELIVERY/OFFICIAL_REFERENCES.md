# OFFICIAL_REFERENCES.md — 공식 기술·정책 레퍼런스 검증 기록 (제21조 준수)

- **문서 ID**: `DOC-DELIV-002`
- **관련 조항**: 제21조 (공식 레퍼런스 검증 의무)
- **검증 기준일**: `2026-10-27`

---

## 1. 공식 기술 및 정책 레퍼런스 검증 테이블 (제21조 7대 필수 기록 항목 완비)

| 번호 | 문서명 | 공식 URL | 확인일 | 관련 기술 / 정책 | 적용되는 요구사항 및 설계 섹션 | 버전 의존성 | 미확인 사항 / 외부 확인 필요 사항 |
| :---: | :--- | :--- | :---: | :--- | :--- | :--- | :--- |
| 1 | **Android Developers — PackageInstaller API Reference** | https://developer.android.com/reference/android/content/pm/PackageInstaller | `2026-10-27` | Android 네이티브 APK 세션 설치(`Session`, `SessionParams`, `STATUS_PENDING_USER_ACTION`) | `REQ-K-002`, `01_A_TO_Z_BLUEPRINT/K_KOTLIN_ANDROID_MARKET.md`, `06_ANDROID/APK_INSTALL_ENGINE.md` | Android API 21+ (무음 자동 업데이트 `setRequireUserAction`은 API 31+ 및 설치 소유권 일치 필수) | 없음 (공식 OS 동작 검증 완료) |
| 2 | **Android Open Source Project — APK Signature Scheme v2 / v3 / v4** | https://source.android.com/docs/security/features/apksigning | `2026-10-27` | APK 전체 파일 서명 블록(v2), 키 회전 증명(v3 `Proof-of-Rotation`), 스트리밍 해시 트리(v4) | `REQ-N-002`, `07_DISTRIBUTION/APK_AAB_SIGNING.md` | Android 7.0+(API 24)부터 v2 지원, Android 9.0+(API 28)부터 v3 지원 (`minSdk 26` 정책으로 v2/v3 100% 강제 가능) | 없음 |
| 3 | **Android Developers — bundletool Command-Line Reference** | https://developer.android.com/tools/bundletool | `2026-10-27` | Android App Bundle(`.aab`)로부터 `.apks` 세트 및 기기별 Split/Universal APK 생성 | `REQ-N-001`, `07_DISTRIBUTION/APK_AAB_SIGNING.md`, `AM-DIST-003` | `bundletool` 1.15+ / Java 17+ 런타임 필요 | 플랫폼 관리형 서명키(Managed Key) 사용 시 개발자 약관 동의 문안은 법무팀 확인 필요 |
| 4 | **Google Play Console Help — Use of REQUEST_INSTALL_PACKAGES Permission** | https://support.google.com/googleplay/android-developer/answer/12085295 | `2026-10-27` | 서드파티 패키지 설치 권한 정책 및 Google Play 내 타 앱스토어 배포 제한 | `REQ-K-001`, `00_PROJECT_CHARTER/ASSUMPTIONS_AND_DECISIONS.md` | Android 8.0+(API 26) `ACTION_MANAGE_UNKNOWN_APP_SOURCES` | BN App Market Android 앱 자체는 Google Play가 아닌 자체 웹(`web-user`) 직접 다운로드 및 OEM 프리로드로 배포함을 확정 |
| 5 | **Google Play Android Developer API Reference** | https://developers.google.com/android-publisher/api-ref/rest | `2026-10-27` | 외부 앱스토어 API 범위 확인 (자사 소유 앱 출시 관리 전용이며 타사 전체 카탈로그 검색 API 아님) | `REQ-E-001`, `08_SEARCH_CONNECTORS/EXTERNAL_CONNECTOR_CONTRACT.md` | REST v3 | 타사 Google Play 앱은 무단 APK 추출을 금지하고 공식 링크(`LEVEL_4_OFFICIAL_LINK_ONLY`) 또는 개발자 직접 등록으로만 처리 |
| 6 | **F-Droid Wiki & Documentation — Index V2 Metadata Format** | https://gitlab.com/fdroid/wiki/-/wikis/Index-V2 | `2026-10-27` | F-Droid 오픈소스 리포지토리 서명 인덱스(`entry.json`, `index-v2.json`) 구조 및 검증 | `REQ-E-002`, `01_A_TO_Z_BLUEPRINT/E_EXTERNAL_STORE_CONNECTORS.md`, `AM-SEARCH-002` | F-Droid Server Index V2 Schema | 미러 대역폭 정책 준수를 위해 대량 바이너리 트래픽 발생 시 자체 공식 미러 캐시 운영 여부 협의 대상 |
| 7 | **OpenSearch Documentation — Mappings & Okapi BM25 Similarity** | https://docs.opensearch.org/latest/search-plugins/keyword-search/ | `2026-10-27` | OpenSearch BM25 랭킹, 다국어 형태소 분석(`analysis-nori`, `analysis-icu`), 하이브리드 검색 | `REQ-O-001`, `REQ-O-002`, `08_SEARCH_CONNECTORS/OPENSEARCH_SCHEMA.md` | OpenSearch `2.17+` | 없음 |
| 8 | **OWASP Application Security Verification Standard (ASVS) & MASVS** | https://owasp.org/projects/asvs / https://mas.owasp.org/MASVS/ | `2026-10-27` | 웹/API 보안(ASVS v4.0 L2) 및 모바일 앱 보안(MASVS v2.0) 검증 기준 | `REQ-S-001`, `REQ-I-001`, `09_SECURITY_COMPLIANCE/SECURITY_ARCHITECTURE.md` | ASVS 4.0.3 / MASVS 2.0 | 없음 |
| 9 | **OpenAPI Initiative — OpenAPI Specification v3.1.0 / v3.1.1** | https://spec.openapis.org/oas/v3.1.0.html | `2026-10-27` | JSON Schema Draft 2020-12 100% 호환 API 명세 표준 | `04_API_SPECIFICATION/openapi.yaml` | OpenAPI `3.1.0` | 없음 |

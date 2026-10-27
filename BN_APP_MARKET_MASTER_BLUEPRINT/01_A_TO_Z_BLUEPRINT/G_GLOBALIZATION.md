# G — Globalization — 다국어(i18n), 다통화, 시간대, 국가별 노출 제어 및 법률 컴플라이언스 (Internationalization (i18n), Multi-Currency, Timezones, Geo-Fencing & Regional Compliance)

- **문서 ID**: `DOC-AZ-G`
- **알파벳 코드**: `G`
- **관련 요구사항 ID**: `REQ-G-001`, `REQ-G-002`
- **연관 Work Order ID**: `AM-API-004`, `AM-WEB-002`, `AM-ADMIN-003`
- **연관 테스트 케이스 ID**: `TC-FUNC-005`, `TC-SEC-004`
- **참조 상세 설계 문서**: [09_SECURITY_COMPLIANCE/COMPLIANCE_CHECKLIST.md](../09_SECURITY_COMPLIANCE/COMPLIANCE_CHECKLIST.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 글로벌 앱 마켓은 사용자마다 사용하는 언어(한국어, 영어, 일본어, 아랍어 RTL 등), 결제 통화, 시간대가 다르며, 특히 국가별 법적 규제(연령 등급제, 도박/암호화폐 앱 금지, 수출 통제, GDPR/개인정보보호법)에 따라 특정 국가에서 앱 노출과 다운로드를 물리적으로 차단(Geo-fencing)할 수 있어야 한다.
- **달성 목표**: BCP 47 로케일 표준, ISO 4217 다통화 처리, UTC 저장 및 로컬 시간대 렌더링, 국가별 배포 허용/차단 매트릭스(`app_country_availability`), 국가별 연령 등급 매핑 체계를 정의한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **글로벌 사용자**: 자신의 언어·통화·국가에 맞는 카탈로그와 가격을 조회하고 연령 제한 정책에 따라 보호받음.
- **개발자 조직**: 앱 배포 대상 국가(전체 국가 또는 특정 국가 화이트리스트/블랙리스트) 및 국가별 현지화된 설명/스크린샷/가격 설정.
- **Policy & Compliance Admin**: 국가별 규제 정책(예: 한국 GRAC 연령등급, 유럽 PEGI, 북미 ESRB, 글로벌 IARC 기준 매핑) 및 국가 단위의 특정 앱 강제 차단(`REGULATORY_TAKEDOWN`) 설정.

## 3. 필수 기능 (Mandatory Features)
1. **다국어 지원 (BCP 47 Locale & RTL 지원)**:
   - 기본 지원 UI 로케일: `ko-KR`, `en-US`, `ja-JP`, `zh-CN`, `zh-TW`, `es-ES`, `de-DE`, `fr-FR`, `pt-BR`, `ar-SA`(RTL 레이아웃 지원).
   - 앱 카탈로그 메타데이터는 개발자가 원하는 모든 BCP 47 로케일로 추가 가능.
2. **국가별 앱 노출 및 다운로드 통제 (`Geo-Availability & Regulatory Block`)**:
   - 개발자가 설정한 배포 허용 국가(`developer_allowed_countries`)와 관리자가 설정한 법적 차단 국가(`regulatory_blocked_countries`)의 교집합을 평가하여, 허용되지 않은 국가에서는 검색 노출 및 다운로드 URL 발급을 모두 차단한다.
3. **국가 판별 우선순위 (Geo-Resolution Hierarchy)**:
   - ① 유료 결제/인증 사용자의 청구지 국가(Verified Billing Country) → ② 기기 SIM/계정 설정 국가 → ③ Edge CDN GeoIP(`CF-IPCountry` / MaxMind GeoLite2) 순으로 판별하며, 법적 차단(`REGULATORY_TAKEDOWN`)은 GeoIP와 계정 국가 중 하나라도 해당되면 즉시 차단한다.
4. **모든 타임스탬프 UTC 표준화**: DB(`TIMESTAMPTZ`)와 API 응답(ISO 8601 UTC `Z`)은 100% UTC로 처리하고, 클라이언트 UI에서만 사용자의 시간대(`Intl.DateTimeFormat`)로 변환 표시한다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **연령 등급(Age Rating) 통합 매핑**:
  - 글로벌 표준 설문(폭력성, 선정성, 사행성, 사용자 간 상호작용, 확률형 아이템 유무) 결과를 바탕으로 `GLOBAL_AGE_RATING` (`ALL`, `AGE_7`, `AGE_12`, `AGE_16`, `AGE_18_ADULT`)을 산출하고 국가별 등급 체계(한국 전체/12세/15세/청소년이용불가 등)로 매핑한다.
  - `AGE_18_ADULT` 등급 앱은 해당 국가 법률이 요구하는 성인 인증(본인인증/연령확인 토큰)이 완료된 계정에서만 상세 미디어 열람 및 다운로드가 허용된다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- CSS 논리적 속성(`margin-inline-start`, `padding-inline-end`, `text-align: start`)을 전면 사용하여 `dir="rtl"`(아랍어/히브리어) 전환 시 레이아웃이 자동으로 좌우 반전되도록 구현한다.
- `LocaleCurrencySelector.vue`: 헤더/푸터에서 언어 및 표시 통화를 즉시 전환할 수 있는 선택기 제공.

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/catalog/domain/geo_policy.go`:
  - `CanAccessAppInCountry(app, userCountry, ipCountry) (bool, BlockReason)` 도메인 함수가 검색 쿼리 필터와 다운로드 토큰 발급기 양쪽에서 동일하게 실행되어 우회를 방지한다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**:
  - `supported_locales` (`locale_code VARCHAR(16) PK`, `display_name`, `is_rtl BOOLEAN`, `is_active BOOLEAN`).
  - `country_policies` (`country_code CHAR(2) PK`, `default_currency CHAR(3)`, `require_adult_verification BOOLEAN`, `prohibited_categories TEXT[]`, `tax_rate_bps INT`).
  - `app_country_availability` (`app_id UUID`, `country_code CHAR(2)`, `is_allowed_by_dev BOOLEAN`, `is_blocked_by_law BOOLEAN`, `block_reason VARCHAR(255)`, `PRIMARY KEY (app_id, country_code)`).

## 8. API 및 이벤트 계약 (API & Event Contracts)
- 모든 Public API는 `Accept-Language` 헤더와 `X-Country-Code` (또는 서버 GeoIP 감지값)를 수용하며, 응답 헤더에 `Content-Language` 및 `Vary: Accept-Language, CF-IPCountry`를 포함한다.
- `PUT /api/v1/admin/policies/countries/{countryCode}`: 국가별 규제/세율/성인인증 정책 수정.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- 특정 국가에서 법원 명령 또는 규제 기관 요청으로 앱이 차단(`is_blocked_by_law = true`)되면, 개발자가 콘솔에서 해당 국가를 다시 활성화하려 해도 `LOCKED_BY_REGULATION` 상태로 잠겨 변경할 수 없다.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 사용자가 HTTP 헤더(`X-Forwarded-For` 또는 `X-Country-Code`)를 위조하여 국가 제한을 우회하지 못하도록, 신뢰할 수 있는 리버스 프록시/CDN이 주입한 서명된 GeoIP 헤더만 인정하고 클라이언트 임의 헤더는 덮어쓴다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 국가 제한으로 접근이 불가능한 앱 URL에 직접 접속한 경우 빈 화면이나 `404`가 아닌, `"이 앱은 개발자 설정 또는 거주 국가(KR)의 법적 정책에 따라 현재 지역에서 제공되지 않습니다"`라는 명확한 안내 화면을 제공한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **국가별 차단 검증 테스트 (`TC-SEC-004`)**: `KR`에서 법적 차단 설정된 앱에 대해 한국 GeoIP로 검색 API, 상세 API, 다운로드 API 호출 시 다운로드가 100% 거부(`451 Unavailable For Legal Reasons`)되는지 검증.
- **RTL 및 다국어 렌더링 테스트 (`TC-FUNC-005`)**: `ar-SA` 로케일에서 `dir="rtl"` 속성 적용 및 날짜/통화 포맷팅 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `B — Business Model`, `C — Catalog Engine`, `O — OpenSearch`(국가별 필터링), `S — Security & Safety`와 연동된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P1`
- **개발 Phase**: 기본 로케일/국가 필드는 `Phase 1~3`에 반영하고, 고도화된 글로벌 정책/RTL/규제 차단은 `Phase 8 (글로벌 운영)`에서 완성한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: GeoIP 데이터베이스 조회 실패** → 기본 국가 코드를 보수적 기준(`DEFAULT_RESTRICTED_PROFILE`)으로 폴백하여, 성인용 앱이나 국가 제한 앱이 미확인 IP에 잘못 노출되지 않도록 Fail-Safe 처리한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 다국어/국가 정책 미들웨어 및 도메인 평가기 (`backend/internal/platform/i18n/`, `backend/internal/modules/catalog/domain/geo_policy.go`).
- 프론트엔드 `vue-i18n` 메시지 번들 및 RTL 스타일시트.

## 17. Acceptance Criteria (수용 기준)
1. 법적 차단 국가에서 다운로드 요청 시 HTTP `451 Unavailable For Legal Reasons`와 표준 에러 코드 `ERR_APP_UNAVAILABLE_IN_COUNTRY`가 반환되어야 한다.
2. 최소 3개 이상 언어(`ko-KR`, `en-US`, `ar-SA`)에서 UI 텍스트 하드코딩 없이 번역 리소스와 RTL 전환이 동작해야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-API-004/geo_block_451_test.log`: 국가별 허용/법적 차단 통합 테스트 로그.
- `artifacts/evidence/AM-WEB-002/rtl_arabic_screenshot.png`: 아랍어 RTL 레이아웃 실제 렌더링 캡처.

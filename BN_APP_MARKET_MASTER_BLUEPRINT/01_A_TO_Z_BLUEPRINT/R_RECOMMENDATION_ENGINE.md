# R — Recommendation Engine — 앱 추천, 실시간/일간 인기 차트, 신규 앱 발견 및 허위 리뷰·조작 방지 엔진 (App Recommendation, Top Charts, New Release Discovery & Anti-Manipulation Engine)

- **문서 ID**: `DOC-AZ-R`
- **알파벳 코드**: `R`
- **관련 요구사항 ID**: `REQ-R-001`, `REQ-R-002`
- **연관 Work Order ID**: `AM-SEARCH-002`, `AM-MOD-001`
- **연관 테스트 케이스 ID**: `TC-FUNC-011`, `TC-SEC-008`
- **참조 상세 설계 문서**: [08_SEARCH_CONNECTORS/SEARCH_RANKING.md](../08_SEARCH_CONNECTORS/SEARCH_RANKING.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 사용자가 검색어를 직접 입력하지 않더라도 홈 화면, 카테고리 화면, 앱 상세 하단에서 신뢰할 수 있는 인기 앱, 유망한 신규 출시 앱, 유사 앱을 발견할 수 있어야 하며, 동시에 봇 농장(Bot Farm)에 의한 다운로드 수 부풀리기나 허위 별점 테러/조작이 차트와 추천 결과를 오염시키지 못하도록 방어해야 한다.
- **달성 목표**: 베이지안 가중 평점(Bayesian Rating), 시간 감쇠(Time-Decay) 인기 차트 알고리즘, 콘텐츠/협업 필터링 하이브리드 유사 앱 추천, 그리고 어뷰징 탐지 및 이상치 제외(Anti-Manipulation) 규칙을 설계한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **일반 사용자**: 홈/카테고리 인기 차트(급상승, 인기 무료, 인기 유료, 신규 출시), 앱 상세 페이지의 `'이 앱과 비슷한 앱'`, `'같은 개발자의 다른 앱'` 탐색.
- **Trust & Safety Moderator / Search Ops**: 비정상 평점 급증 알림 확인, 조작 의심 리뷰 일괄 블라인드(Quarantine) 및 차트 가중치 페널티 적용.

## 3. 필수 기능 (Mandatory Features)
1. **4대 공식 차트 산출 엔진 (매시간 갱신)**:
   - `TOP_FREE` (인기 무료 앱), `TOP_PAID` (인기 유료 앱), `TRENDING_RISING` (최근 48시간 다운로드 가속도 급상승 앱), `NEW_NOTEWORTHY` (출시 30일 이내 고품질 신규 앱).
2. **베이지안 평균 평점 (Bayesian Weighted Rating)**:
   - 리뷰 1개짜리 5.0점 앱이 리뷰 10,000개짜리 4.8점 앱보다 상위에 노출되는 왜곡을 막기 위해 `WeightedRating = (v / (v + m)) * R + (m / (v + m)) * C` (`R`: 앱 평균 평점, `v`: 유효 리뷰 수, `m`: 최소 신뢰 리뷰 임계치=25, `C`: 전체 카탈로그 평균 평점=4.1) 공식을 적용한다.
3. **신규 앱 콜드스타트 부스트 (New App Discovery Slot)**:
   - 출시 후 14일 이내이며 보안 스캔 무결점 및 설명/스크린샷 품질 점수가 기준치 이상인 신규 앱을 카테고리 상단 `'주목할 신규 앱'` 구좌에 순환 노출하여 초기 발견 기회를 보장한다.
4. **허위 리뷰 및 다운로드 조작 방지 (Anti-Manipulation Guard)**:
   - 실제 다운로드 및 설치 완료 이벤트(`install_events` 또는 `user_entitlements`)가 확인된 계정만 리뷰를 작성할 수 있다(`Verified Installer Only Review`).
   - 동일 IP 대역/기기 지문(Device Fingerprint)에서의 반복 다운로드, 가입 직후 단시간 대량 1점/5점 투척은 이상 탐지 규칙에 의해 `FLAGGED_SUSPICIOUS`로 분류되어 공개 평점 및 차트 집계에서 자동 제외된다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **유사 앱 추천 (`Similar Apps`) 산출 방식**:
  - 1차(기본 결정론적 엔진): OpenSearch `more_like_this` 쿼리(카테고리, 태그, 설명 핵심 키워드 TF-IDF 유사도) + 기기 호환성 필터 + 베이지안 평점 재정렬.
  - 2차(Python 선택적 배치 워커): 공동 설치 행렬(Co-installation Matrix: 앱 A를 설치한 사용자가 함께 설치한 앱 B의 Pointwise Mutual Information 점수)을 일 1회 계산하여 `app_recommendations` 테이블에 캐싱한다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **`web-user`**: `TopChartsSection.vue`, `SimilarAppsCarousel.vue`, `VerifiedInstallBadge.vue` (리뷰 작성자가 실제 설치 검증된 사용자임을 표시하고 리뷰 작성 시점의 앱 버전 코드 예: `v2.4.1 사용 후 작성`을 함께 표기).

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/catalog/application/chart_service.go` 및 `internal/modules/moderation/application/anti_fraud_service.go`.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**: `app_charts_snapshot`, `app_recommendations`, `app_reviews_user` (`is_verified_install BOOLEAN NOT NULL`, `fraud_score NUMERIC(5,2)`, `status`).

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `GET /api/v1/catalog/charts/{chartType}`: 국가·카테고리별 차트 조회.
- `GET /api/v1/catalog/apps/{slug}/recommendations`: 유사 앱 및 동일 개발자 앱 목록 조회.
- `POST /api/v1/catalog/apps/{slug}/reviews`: 설치 검증된 사용자의 리뷰/평점 등록.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **리뷰 상태 (`review_status`)**: `PUBLISHED` / `QUARANTINED_AUTO_FRAUD` (자동 어뷰징 탐지 보류) → (모더레이터 검토 후) → `PUBLISHED_RESTORED` | `HIDDEN_POLICY_VIOLATION`.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 미설치 사용자의 리뷰 작성 API 호출은 서버에서 `install_events` 및 `user_entitlements` 기록을 대조하여 `403 ERR_REVIEW_REQUIRES_INSTALLATION`으로 차단한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 앱 평점 표시 옆에 별점 분포 막대그래프(5점~1점 비율)와 `'실제 설치가 확인된 사용자의 리뷰만 반영됩니다'` 안내 문구를 표시한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **미설치자 리뷰 차단 및 베이지안 평점 계산 테스트 (`TC-FUNC-011`)**: 설치 기록이 없는 유저의 리뷰 등록 시 `403` 거부 확인 및 리뷰 수에 따른 베이지안 가중 평점 정확도 검증.
- **동일 IP/기기 어뷰징 다운로드 차트 반영 차단 테스트 (`TC-SEC-008`)**: 동일 기기 ID로 100회 반복 다운로드 발생 시 차트 유효 설치 카운트는 1회만 증가하는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `C — Catalog Engine`, `O — OpenSearch`, `S — Security & Safety`, `Y — Yield & Analytics`와 연동된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P1`
- **개발 Phase**: 기본 차트와 설치 검증 리뷰는 `Phase 3~5`, 어뷰징 탐지 고도화는 `Phase 6`에서 구현한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 시간별 차트 집계 배치 지연** → 이전 시간대 스냅샷(`app_charts_snapshot`)을 그대로 서빙하여 사용자 홈 화면에는 빈 차트가 노출되지 않도록 보장한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 차트/추천 집계 서비스 및 리뷰 어뷰징 방어 모듈 (`backend/internal/modules/catalog/*`, `backend/internal/modules/moderation/*`).

## 17. Acceptance Criteria (수용 기준)
1. 실제 앱 다운로드/설치 이력이 없는 계정은 리뷰를 등록할 수 없어야 한다.
2. 동일 사용자/기기의 반복 다운로드는 인기 차트 점수를 부풀리지 못해야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-MOD-001/verified_review_and_fraud_test.log`: 미설치 리뷰 차단 및 중복 다운로드 디듀플리케이션 검증 로그.

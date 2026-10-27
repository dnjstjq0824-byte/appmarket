# SEARCH_RANKING.md — 검색 품질, BM25 + 품질/인기도 하이브리드 랭킹 수식 및 중복 제거 설계

- **문서 ID**: `DOC-SRCH-002`
- **관련 문서**: [OPENSEARCH_SCHEMA.md](OPENSEARCH_SCHEMA.md), [01_A_TO_Z_BLUEPRINT/O_OPENSEARCH.md](../01_A_TO_Z_BLUEPRINT/O_OPENSEARCH.md)

---

## 1. 하이브리드 검색 랭킹 수식 (Hybrid Relevance & Quality Score)

최종 검색 점수 `FinalScore(q, app)`는 텍스트 유사도(`BM25`)와 앱 품질·신뢰도 신호를 결합하여 산출합니다:

```
FinalScore(q, d) = BM25_MultiMatch(q, d)
                   * (1.0 + 0.25 * (d.bayesian_rating / 5.0))
                   * (1.0 + 0.15 * log10(1 + d.installs_30d))
                   * (1.0 + 0.15 * DirectInstallBoost(d.has_direct_install))
                   * GaussianDecay(d.updated_at, scale="180d", decay=0.85)
```

### 1.1 필드별 가중치 (`multi_match` `best_fields` + `phrase_prefix`)
- 정확한 패키지명 일치 (`primary_package_name.raw`): `Boost ^10.0`
- 앱 제목 정확/구문 일치 (`title_i18n.{locale}`): `Boost ^5.0`
- 개발자/제작사 명칭 (`developer_name`): `Boost ^3.0`
- 시맨틱 태그 (`tags`): `Boost ^2.5`
- 짧은 설명 (`short_description_i18n`): `Boost ^1.5`
- 상세 설명 (`full_description_i18n`): `Boost ^0.8`

### 1.2 필수 보안/국가/호환성 하드 필터 (`bool.filter`)
점수가 아무리 높아도 아래 조건에 위배되면 검색 결과에서 즉시 제외됩니다:
1. `term: { "allowed_countries": user_country }` AND `must_not: { "term": { "blocked_countries": user_country } }`
2. 성인 미인증 사용자의 경우 `must_not: { "term": { "global_age_rating": "AGE_18_ADULT" } }`
3. 사용자가 `'내 기기 호환 앱만 보기'`를 활성화한 경우 `range: { "min_sdk_lowest": { "lte": device_sdk } }` AND `terms: { "supported_abis_union": [device_abis, "UNIVERSAL"] }`.

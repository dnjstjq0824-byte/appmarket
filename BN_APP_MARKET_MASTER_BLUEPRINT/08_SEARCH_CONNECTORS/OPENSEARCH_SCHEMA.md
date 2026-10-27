# OPENSEARCH_SCHEMA.md — OpenSearch 2.x 통합 앱 인덱스 및 자동완성 인덱스 매핑 JSON

- **문서 ID**: `DOC-SRCH-001`
- **관련 조항**: 제3조 O항, 제11조
- **연관 Work Order**: `AM-SEARCH-001`

---

## 1. 통합 앱 검색 인덱스 매핑 (`bn_apps_v1`)

```json
{
  "settings": {
    "number_of_shards": 3,
    "number_of_replicas": 1,
    "refresh_interval": "2s",
    "analysis": {
      "tokenizer": {
        "korean_nori_mixed": {
          "type": "nori_tokenizer",
          "decompound_mode": "mixed"
        },
        "edge_ngram_tokenizer": {
          "type": "edge_ngram",
          "min_gram": 1,
          "max_gram": 15,
          "token_chars": ["letter", "digit"]
        }
      },
      "analyzer": {
        "ko_search_analyzer": {
          "type": "custom",
          "tokenizer": "korean_nori_mixed",
          "filter": ["lowercase", "nori_Part_of_speech", "nori_readingform", "icu_folding"]
        },
        "autocomplete_index_analyzer": {
          "type": "custom",
          "tokenizer": "edge_ngram_tokenizer",
          "filter": ["lowercase", "icu_folding"]
        }
      }
    }
  },
  "mappings": {
    "dynamic": "strict",
    "properties": {
      "app_id": { "type": "keyword" },
      "canonical_slug": { "type": "keyword" },
      "primary_package_name": {
        "type": "text",
        "analyzer": "standard",
        "fields": { "raw": { "type": "keyword" } }
      },
      "title_i18n": {
        "type": "object",
        "properties": {
          "ko": { "type": "text", "analyzer": "ko_search_analyzer", "fields": { "prefix": { "type": "text", "analyzer": "autocomplete_index_analyzer" } } },
          "en": { "type": "text", "analyzer": "english", "fields": { "prefix": { "type": "text", "analyzer": "autocomplete_index_analyzer" } } },
          "default": { "type": "text", "analyzer": "ko_search_analyzer" }
        }
      },
      "developer_name": { "type": "text", "analyzer": "ko_search_analyzer", "fields": { "raw": { "type": "keyword" } } },
      "short_description_i18n": { "type": "object", "enabled": true },
      "full_description_i18n": { "type": "object", "enabled": true },
      "primary_category_id": { "type": "keyword" },
      "tags": { "type": "keyword" },
      "monetization_type": { "type": "keyword" },
      "global_age_rating": { "type": "keyword" },
      "allowed_countries": { "type": "keyword" },
      "blocked_countries": { "type": "keyword" },
      "has_direct_install": { "type": "boolean" },
      "source_types": { "type": "keyword" },
      "min_sdk_lowest": { "type": "integer" },
      "supported_abis_union": { "type": "keyword" },
      "bayesian_rating": { "type": "scaled_float", "scaling_factor": 1000 },
      "rating_count": { "type": "integer" },
      "installs_30d": { "type": "long" },
      "quality_score": { "type": "float" },
      "published_at": { "type": "date" },
      "updated_at": { "type": "date" },
      "distributions": {
        "type": "nested",
        "properties": {
          "distribution_id": { "type": "keyword" },
          "source_type": { "type": "keyword" },
          "distribution_mode": { "type": "keyword" },
          "rights_level": { "type": "keyword" },
          "latest_version_code": { "type": "long" },
          "latest_version_name": { "type": "keyword" },
          "price_krw": { "type": "long" },
          "provenance_label": { "type": "keyword" },
          "official_store_url": { "type": "keyword", "index": false },
          "last_synced_at": { "type": "date" }
        }
      }
    }
  }
}
```

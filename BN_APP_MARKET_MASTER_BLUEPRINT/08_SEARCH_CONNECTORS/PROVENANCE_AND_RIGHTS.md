# PROVENANCE_AND_RIGHTS.md — 외부 앱 데이터 권리 5단계 분류 및 출처(Provenance) 표시 의무 규격

- **문서 ID**: `DOC-SRCH-004`
- **관련 조항**: 제11조, 제13조
- **연관 문서**: [EXTERNAL_CONNECTOR_CONTRACT.md](EXTERNAL_CONNECTOR_CONTRACT.md)

---

## 1. 5단계 데이터 권리 등급 (`rights_level`) 및 허용 동작 매트릭스

| 권리 등급 코드 | 메타데이터 검색 노출 | 상세 설명/스크린샷 캐싱 | 자체 CDN APK 다운로드 | 외부 공식 링크 연결 | UI 버튼 표시 문구 강제 규칙 |
| :--- | :---: | :---: | :---: | :---: | :--- |
| `LEVEL_1_DIRECT_HOSTED` | 허용 | 허용 (자체 원본) | **허용 (기본)** | N/A | `직접 APK 다운로드` / `즉시 설치` |
| `LEVEL_2_LICENSED_FEDERATED_BINARY` | 허용 | 허용 (라이선스/출처 명시 필수) | 허용 (검증된 공식 리포지토리/미러) | 허용 | `오픈소스 패키지 설치 (출처: F-Droid)` |
| `LEVEL_3_SEARCH_AND_INFO_ONLY` | 허용 | 요약 및 제휴 범위 내 허용 | **엄격 금지 (`403`)** | **필수** | `[스토어명]에서 확인 및 설치 ↗` |
| `LEVEL_4_OFFICIAL_LINK_ONLY` | 허용 (식별정보 중심) | 원본 이미지 미러링 금지, 공식 링크만 유지 | **엄격 금지 (`403`)** | **필수** | `공식 배포처(Google Play)로 이동 ↗` |
| `LEVEL_5_RESTRICTED_PROHIBITED` | **전면 차단** | **전면 차단** | **엄격 금지** | **차단** | 검색 및 상세 노출 제외 (`404`/`451`) |

---

## 2. 검색 결과 및 상세 화면의 출처(Provenance) 4대 필수 표시 항목

외부 배포처(`source_type != 'BN_HOSTED'`)가 포함된 모든 API 응답과 UI 카드에는 다음 4가지 필드가 반드시 노출되어야 합니다:
1. **`provenance_label`**: 데이터 출처 명칭 (예: `"F-Droid Official FOSS Repository"`, `"Developer Verified Official Link"`).
2. **`last_synced_at`**: 외부 출처와 마지막으로 동기화된 정확한 UTC 시각 (예: `"2026-10-27 12:00 UTC 동기화됨"`).
3. **`official_store_url`**: 클릭 시 이동할 검증된 공식 배포처 원본 URL.
4. **`can_direct_install`**: 당사 앱/웹에서 직접 APK 다운로드·설치가 가능한지(`true`/`false`) 여부.

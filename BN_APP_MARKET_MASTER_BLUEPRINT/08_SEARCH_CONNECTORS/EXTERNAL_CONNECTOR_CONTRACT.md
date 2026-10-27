# EXTERNAL_CONNECTOR_CONTRACT.md — 외부 앱스토어 커넥터 기술 계약 및 스토어별 연동 규격

- **문서 ID**: `DOC-SRCH-003`
- **관련 조항**: 제11조 (외부 앱 검색엔진 상세 명세)
- **연관 Work Order**: `AM-SEARCH-002`

---

## 1. 외부 배포처별 연동 가능성·조건·제약 조사 결과 (제11조 준수)

| 대상 외부 배포처 | 공개 전체 카탈로그 API 유무 | 바이너리 직접 재배포 가능 여부 | 당사 플랫폼 표준 연동 방식 및 기술 제약 | 기본 권리 등급 (`rights_level`) |
| :--- | :---: | :---: | :--- | :--- |
| **F-Droid (공식 오픈소스 리포지토리)** | **있음** (`entry.json` + `index-v2.json` 정적 서명 피드) | **가능** (FOSS 라이선스 및 F-Droid 미러 정책 준수 시) | `entry.json` JAR 서명 검증 → `index-v2.json` 스트리밍 파싱 → 패키지 메타데이터·버전·SHA256·서명 지문 수집 및 공식 미러/링크 연결 | `LEVEL_2_LICENSED_FEDERATED_BINARY` |
| **Google Play Store** | **없음** (Android Developer API는 본인 소유 앱 관리 전용) | **불가** (무단 APK 추출·미러링 엄격 금지) | ① 당사 가입 개발자가 직접 자사 앱의 Google Play URL 등록, ② 공식 제휴 데이터 피드, ③ 허용된 공개 `SoftwareApplication` 메타데이터만 제한적 색인 후 **공식 링크(`https://play.google.com/store/apps/details?id=...`)로만 연결** | `LEVEL_4_OFFICIAL_LINK_ONLY` |
| **Samsung Galaxy Store** | **없음** (Seller Portal API는 본인 앱 관리 전용) | **불가** | 개발자 직접 등록 링크 또는 공식 제휴 피드 기반으로만 색인 후 **공식 딥링크(`samsungapps://ProductDetail/<pkg>`) 및 웹 링크로만 연결** | `LEVEL_4_OFFICIAL_LINK_ONLY` |
| **개발자 공식 웹사이트 / GitHub Releases** | **있음** (개발자 승인 시 GitHub Releases API / 표준 `appmarket-feed.json`) | **계약/설정에 따름** | 개발자가 콘솔에서 공식 GitHub 리포지토리 또는 자체 호스팅 피드를 연결하고 서명 인증서를 검증한 경우 비교 표에 노출 | `LEVEL_3_SEARCH_AND_INFO_ONLY` 또는 `LEVEL_1_DIRECT_HOSTED` |

> **핵심 준수 원칙**: 존재하지 않는 외부 스토어 API를 창작하거나 비공식 스크래퍼로 타사 APK 파일을 복제해 자체 CDN에서 서빙하는 코드는 절대 작성하지 않습니다.

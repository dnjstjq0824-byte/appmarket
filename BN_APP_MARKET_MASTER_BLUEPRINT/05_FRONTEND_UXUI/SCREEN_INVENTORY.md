# SCREEN_INVENTORY.md — 전체 76개 화면(사용자 웹 20 + 개발자 24 + 관리자 20 + Android 12) 마스터 인벤토리

- **문서 ID**: `DOC-UX-005`
- **관련 문서**: [USER_WEB.md](USER_WEB.md), [DEVELOPER_CONSOLE.md](DEVELOPER_CONSOLE.md), [ADMIN_CONSOLE.md](ADMIN_CONSOLE.md), [06_ANDROID/ANDROID_ARCHITECTURE.md](../06_ANDROID/ANDROID_ARCHITECTURE.md)

---

## 1. 플랫폼별 화면 인벤토리 총괄 요약 (Total: 76 Screens)

| 클라이언트 플랫폼 | 화면 ID 범위 | 총 화면 수 | 상세 명세 참조 위치 |
| :--- | :---: | :---: | :--- |
| **사용자 마켓 웹사이트 (`web-user`)** | `USR-01` ~ `USR-20` | **20개** | [USER_WEB.md](USER_WEB.md) 섹션 2 (`USR-01`~`USR-20` 전수 매핑) |
| **개발자 콘솔 (`web-dev`)** | `DEV-01` ~ `DEV-24` | **24개** | [DEVELOPER_CONSOLE.md](DEVELOPER_CONSOLE.md) 섹션 1 (`1`~`24`번 전수 매핑) |
| **관리자 콘솔 (`web-admin`)** | `ADM-01` ~ `ADM-20` | **20개** | [ADMIN_CONSOLE.md](ADMIN_CONSOLE.md) 섹션 1 (`ADM-01`~`ADM-20` 전수 매핑) |
| **Android 네이티브 앱 (`android-client`)** | `AND-01` ~ `AND-12` | **12개** | 아래 섹션 2 및 [06_ANDROID/ANDROID_ARCHITECTURE.md](../06_ANDROID/ANDROID_ARCHITECTURE.md) |

---

## 2. Android 네이티브 클라이언트 12개 화면 인벤토리 (`AND-01` ~ `AND-12`)

| 화면 ID | Compose Screen 명칭 | Navigation Route | 핵심 기능 및 연동 OS API |
| :---: | :--- | :--- | :--- |
| `AND-01` | `HomeScreen` | `home` | 기기 호환 추천 앱, 인기 차트, 상단 업데이트 알림 배너 |
| `AND-02` | `SearchScreen` | `search?q={q}` | 실시간 자동완성, 기기 호환 필터(`호환 앱만 보기`), 배포처 배지 표시 |
| `AND-03` | `CategoryBrowseScreen` | `categories/{id}` | 앱/게임 카테고리별 탐색 및 차트 조회 |
| `AND-04` | `AppDetailScreen` | `apps/{slug}` | 진행률 연동 설치/업데이트 버튼, 배포처 전환기, 스크린샷, 리뷰, 권한/서명 카드 |
| `AND-05` | `PermissionSecuritySheet` | `apps/{slug}/security` | 요구 권한 위험도 및 로컬 vs 서버 서명 인증서 SHA-256 일치 검증 결과 표시 |
| `AND-06` | `DownloadManagerScreen` | `downloads` | 진행 중/일시정지/대기 중 다운로드 목록, 속도/남은 시간, 일시정지·재개·취소 제어 |
| `AND-07` | `UpdatesCenterScreen` | `updates` | 기설치 앱 스캔 결과: `즉시 업데이트 가능(서명 일치)` vs `타 배포처 설치 앱(서명 불일치 경고)` 분리 |
| `AND-08` | `UnknownSourcesGuideSheet` | `onboarding/install-perm` | `canRequestPackageInstalls() == false`일 때 `ACTION_MANAGE_UNKNOWN_APP_SOURCES` 안내 및 자동 복귀 |
| `AND-09` | `SignatureConflictDialog` | `dialog/sig-conflict` | `INSTALL_FAILED_UPDATE_INCOMPATIBLE` 예방을 위해 기존 앱 삭제 후 재설치 여부 확인 |
| `AND-10` | `StorageCleanupDialog` | `dialog/storage-low` | 저장공간 부족 시 필요 용량(MB) 안내 및 다운로드 스테이징 캐시 정리 |
| `AND-11` | `MyLibraryScreen` | `library` | 유료 구매한 앱 목록(`user_entitlements`), 기기 설치 이력, 리뷰 작성 |
| `AND-12` | `SettingsAndSelfUpdateScreen` | `settings` | Wi-Fi 전용 다운로드, 자동 업데이트 설정, 다크모드, BN 마켓 앱 자체 업데이트 확인 |

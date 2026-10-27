# ANDROID_ARCHITECTURE.md — Kotlin + Jetpack Compose 네이티브 Android 마켓 아키텍처

- **문서 ID**: `DOC-AND-001`
- **관련 조항**: 제5조 (Android 사용자 앱 전체 개발 명세)
- **연관 Work Order**: `AM-ANDROID-001`, `AM-ANDROID-002`, `AM-ANDROID-003`

---

## 1. Gradle 멀티모듈 아키텍처 (`android-client/`)

```text
android-client/
├── app/                          # Application Entrypoint, Hilt DI Graph, MainActivity, Self-Update
├── core/
│   ├── model/                    # Pure Kotlin Domain Models (CatalogApp, ReleaseArtifact, InstallTask)
│   ├── network/                  # OkHttp3, Retrofit2, Kotlinx.Serialization, AuthAuthenticator
│   ├── database/                 # Room Database (DownloadTaskDao, InstalledSnapshotDao), DataStore
│   ├── device/                   # DeviceSpecProvider (SDK, ABIs, DPI, Features, StorageManager)
│   ├── downloader/               # Resumable HttpRangeDownloader, Sha256Verifier, ForegroundService
│   └── installer/                # PackageInstallerSessionManager, ApkCertVerifier, InstallResultReceiver
└── feature/
    ├── home/                     # HomeScreen, TopCharts, Spotlight Compose UI + ViewModel
    ├── search/                   # SearchScreen, Autocomplete, Facet Filter Compose UI + ViewModel
    ├── detail/                   # AppDetailScreen, SourceComparison, PermissionSheet
    ├── updates/                  # UpdatesCenterScreen, DownloadManagerScreen
    └── settings/                 # SettingsScreen, UnknownSourcesOnboarding
```

---

## 2. 단방향 데이터 흐름 (MVI + Kotlin Coroutines `StateFlow`)

- 모든 화면 ViewModel은 불변 `UiState` 데이터 클래스를 `StateFlow<UiState>`로 노출하고, 사용자 액션은 `sealed interface UiIntent`로 수신하며, 일회성 이벤트(인텐트 실행, 스낵바)는 `Channel<UiEffect>`로 전달합니다.
- 다운로드 및 설치 상태는 `DownloadRepository`와 `PackageInstallerRepository`의 Room `Flow<List<DownloadTaskEntity>>`를 단일 진실 공급원(Single Source of Truth)으로 구독하여 앱을 종료했다가 다시 열어도 진행률이 즉시 동기화됩니다.

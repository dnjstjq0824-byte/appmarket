# APK_INSTALL_ENGINE.md — 이어받기 다운로드 매니저, 무결성 검증 및 PackageInstaller 설치 엔진

- **문서 ID**: `DOC-AND-002`
- **관련 조항**: 제5조 (항목 5~16)
- **연관 Work Order**: `AM-ANDROID-002`

---

## 1. 다운로드 → 무결성/서명 검증 → PackageInstaller 설치 파이프라인

```kotlin
// 핵심 실행 순서 (core/installer/ApkInstallCoordinator.kt 설계 규격)
// 1. 여유 저장공간 사전 확인 (StorageManager)
val allocatable = storageManager.getAllocatableBytes(storageUuid)
val requiredBytes = (artifact.fileSizeBytes * 22) / 10 + (100L * 1024 * 1024)
if (allocatable < requiredBytes) return InstallResult.InsufficientStorage(requiredBytes - allocatable)

// 2. HTTP Range 이어받기 완료 후 SHA-256 해시 검증
val actualSha256 = sha256Verifier.computeFileSha256(stagedApkFile)
if (!actualSha256.equals(artifact.expectedSha256, ignoreCase = true)) {
    stagedApkFile.delete()
    return InstallResult.ChecksumMismatch(expected = artifact.expectedSha256, actual = actualSha256)
}

// 3. 로컬 APK 서명 인증서 추출 및 서버 기대 지문 / 기설치 앱 서명 대조
val archiveCertSha256 = apkCertVerifier.extractPrimarySignerSha256(stagedApkFile)
if (!archiveCertSha256.equals(artifact.expectedSignerSha256, ignoreCase = true)) {
    stagedApkFile.delete()
    return InstallResult.SignerVerificationFailed
}

// 4. REQUEST_INSTALL_PACKAGES 권한 확인 (API 26+)
if (!packageManager.canRequestPackageInstalls()) {
    return InstallResult.RequiresUnknownSourcesPermission
}

// 5. PackageInstaller.Session 생성, 스트림 복사, fsync, commit
val params = PackageInstaller.SessionParams(PackageInstaller.SessionParams.MODE_FULL_INSTALL).apply {
    setAppPackageName(artifact.packageName)
    setSize(stagedApkFile.length())
    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S && isSelfOwnedUpdate) {
        setRequireUserAction(PackageInstaller.SessionParams.USER_ACTION_NOT_REQUIRED)
    }
}
val sessionId = packageInstaller.createSession(params)
packageInstaller.openSession(sessionId).use { session ->
    session.openWrite("base.apk", 0, stagedApkFile.length()).use { out ->
        stagedApkFile.inputStream().use { input -> input.copyTo(out, bufferSize = 64 * 1024) }
        session.fsync(out)
    }
    session.commit(createInstallIntentSender(context, sessionId, artifact.packageName))
}
```

---

## 2. `PackageInstaller` 상태 코드별 상세 예외 복구 매트릭스

| `PackageInstaller` 상태 상수 | 세부 원인 (`EXTRA_STATUS_MESSAGE`) | 클라이언트 자동 복구 및 사용자 안내 동작 |
| :--- | :--- | :--- |
| `STATUS_PENDING_USER_ACTION` | 시스템 설치 승인 다이얼로그 필요 | 브로드캐스트에서 추출한 `Intent.EXTRA_INTENT`를 즉시 실행하고 상태를 `AWAITING_USER_CONFIRMATION`으로 전환 |
| `STATUS_SUCCESS` | 설치 또는 업데이트 완료 | 상태를 `INSTALLED`로 갱신, 스테이징 APK 삭제, 서버에 `INSTALL_SUCCEEDED` 보고, `'열기'` 버튼 활성화 |
| `STATUS_FAILURE_ABORTED` | 사용자가 설치 확인 창에서 `'취소'` 클릭 | 에러가 아닌 `CANCELED_BY_USER`로 전환하고 다운로드된 APK는 24시간 캐시 유지(재시도 시 재다운로드 불필요) |
| `STATUS_FAILURE_CONFLICT` | `INSTALL_FAILED_UPDATE_INCOMPATIBLE` (서명 불일치) | `SignatureConflictDialog` 호출: 기존 앱과 서명 키가 다름을 안내하고 기존 앱 제거 인텐트(`ACTION_DELETE`) 후 재설치 옵션 제공 |
| `STATUS_FAILURE_INVALID` | `INSTALL_FAILED_VERSION_DOWNGRADE` 또는 손상된 APK | 스테이징 파일 즉시 삭제 후 서버에 오류 리포트 전송 |
| `STATUS_FAILURE_INCOMPATIBLE` | `INSTALL_FAILED_NO_MATCHING_ABIS` / `OLDER_SDK` | 기기 CPU ABI 또는 Android 버전 비호환 상세 사유 표시 |
| `STATUS_FAILURE_STORAGE` | 기기 내부 저장공간 부족 | `StorageCleanupDialog` 호출 및 부족 용량(MB) 안내 |

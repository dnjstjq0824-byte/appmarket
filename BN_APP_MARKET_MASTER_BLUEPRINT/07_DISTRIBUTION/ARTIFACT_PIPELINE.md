# ARTIFACT_PIPELINE.md — APK/AAB 업로드·격리·분석·보안검증·출시 11단계 파이프라인 명세

- **문서 ID**: `DOC-DIST-001`
- **관련 조항**: 제10조 (APK/AAB 및 앱 배포 엔진)
- **연관 Work Order**: `AM-DIST-001`, `AM-DIST-002`, `AM-SEC-002`

---

## 1. 11단계 아티팩트 처리 파이프라인 상세 규격

```
[Step 1: Upload Session] ──► [Step 2: Quarantine S3] ──► [Step 3: ZIP & SHA-256 Check]
                                                                    │
[Step 6: Lineage Check] ◄── [Step 5: apksigner v2/v3] ◄── [Step 4: AXML Manifest Parse]
         │
         ▼
[Step 7: ClamAV + YARA + Permission Risk Scan] ──► [Step 8: Admin Review & Approval]
                                                                    │
[Step 11: Android OS Install] ◄── [Step 10: Signed CDN URL] ◄── [Step 9: Public Immutable Copy]
```

### 1.1 각 단계별 실패 코드 및 격리 규칙
1. **Step 3 (무결성 및 압축폭탄 검사)**: 업로드 시 선언한 `expected_sha256`과 실제 스트림 `SHA-256` 불일치 시 즉시 삭제 후 `REJECTED_INVALID_APK`. 압축 해제 비율 `> 100:1` 또는 경로 내 `../` 포함 시 `ERR_ZIP_BOMB_DETECTED`.
2. **Step 4 (AXML 바이너리 매니페스트 추출)**: `AndroidManifest.xml`에서 `package`, `versionCode`, `versionName`, `minSdkVersion`, `targetSdkVersion`, `uses-permission`, `uses-feature`, 그리고 `lib/` 디렉터리 하위 네이티브 `.so` 아키텍처(`arm64-v8a`, `armeabi-v7a`, `x86_64`, `x86`) 목록을 추출합니다.
3. **Step 5 & 6 (서명 검증 및 키 연속성)**: `signature_scheme_version < 2`이면 거부합니다. 동일 배포처의 이전 릴리스가 존재할 경우 `apk_signatures.cert_sha256_fingerprint`가 일치하거나 유효한 v3 `proof_of_rotation` 체인이 확인되어야만 통과합니다.

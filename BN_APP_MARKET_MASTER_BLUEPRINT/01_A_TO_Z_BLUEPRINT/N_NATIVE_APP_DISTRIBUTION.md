# N — Native App Distribution — APK/AAB 처리, 서명 검증, 무결성 보장 및 글로벌 배포 엔진 (APK/AAB Processing Pipeline, Signature Verification, Integrity & CDN Distribution Engine)

- **문서 ID**: `DOC-AZ-N`
- **알파벳 코드**: `N`
- **관련 요구사항 ID**: `REQ-N-001`, `REQ-N-002`, `REQ-N-003`
- **연관 Work Order ID**: `AM-DIST-001`, `AM-DIST-002`
- **연관 테스트 케이스 ID**: `TC-FUNC-009`, `TC-SEC-007`, `TC-AND-001`
- **참조 상세 설계 문서**: [07_DISTRIBUTION/ARTIFACT_PIPELINE.md](../07_DISTRIBUTION/ARTIFACT_PIPELINE.md), [07_DISTRIBUTION/APK_AAB_SIGNING.md](../07_DISTRIBUTION/APK_AAB_SIGNING.md), [07_DISTRIBUTION/STORAGE_CDN.md](../07_DISTRIBUTION/STORAGE_CDN.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 제10조에서 명령한 본 플랫폼의 핵심 엔진으로서, 개발자가 업로드한 APK/AAB 파일을 **업로드 → 격리 저장 → 무결성 검증 → 메타데이터 분석 → 서명 검증 → 악성코드 검사 → 관리자 심사 → 출시 승인 → 스토리지 배포 → CDN 전달 → 실제 Android 설치**까지 단 하나의 보안 허점이나 파일 변조 없이 전달하기 위해 필요하다.
- **달성 목표**: APK/AAB 바이너리 파싱 규격, APK Signature Scheme v2/v3/v4 검증, 업데이트 서명 연속성(Key Lineage) 강제, AAB `bundletool` 확장 구조, 불변 CDN 배포 및 긴급 회수(Kill Switch) 파이프라인을 설계한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **Release Manager (개발자)**: 서명된 APK(또는 AAB) 바이너리 업로드 및 아티팩트 분석 리포트 확인.
- **Isolated Artifact Worker (`worker-apk-analyzer`)**: 격리 버킷의 바이너리를 스트리밍 분석하여 매니페스트와 서명 정보를 추출.
- **일반 사용자 / Android Client**: 서명된 CDN URL을 통해 무결성이 검증된 APK를 다운로드 및 설치.

## 3. 필수 기능 (Mandatory Features)
1. **11단계 종단간 배포 파이프라인 강제 (제10조)**:
   - Step 1: 업로드 세션 발급 (`POST /api/v1/developer/uploads/sessions`)
   - Step 2: 격리 스토리지(`bn-quarantine-bucket`) 직접 멀티파트 업로드
   - Step 3: SHA-256 해시 일치 및 ZIP 구조 무결성(Zip Slip, Zip Bomb 방어) 검증
   - Step 4: Binary XML(`AndroidManifest.xml`) 및 `resources.arsc` 파싱 (`package`, `versionCode`, `versionName`, `minSdkVersion`, `targetSdkVersion`, `uses-permission`, `uses-feature`, `lib/<abi>/*.so` 추출)
   - Step 5: Android `apksigner verify --print-certs` 기반 v2/v3 서명 블록 검증 및 X.509 인증서 SHA-256 지문 추출
   - Step 6: 기존 출시 버전과의 서명 인증서 일치성(또는 v3 Proof-of-Rotation) 검증
   - Step 7: 악성코드 정적/YARA/ClamAV 스캔 (`S — Security & Safety`)
   - Step 8: 관리자 심사 및 출시 승인 (`L — Lifecycle`)
   - Step 9: 공개 불변 버킷(`bn-public-cdn-bucket`)으로 원자적 복사 (`packages/{app_id}/{release_id}/{sha256}.apk`)
   - Step 10: HMAC-SHA256 기반 만료 시간 포함 CDN Signed URL 발급
   - Step 11: Android 클라이언트 `SHA-256` 재검증 및 `PackageInstaller` 설치.
2. **AAB(Android App Bundle) 무중단 확장 설계**:
   - 초기에는 개발자 직접 서명 APK(Universal APK 또는 ABI별 Multi-APK)를 기본으로 처리하되, `ArtifactProcessor` 인터페이스의 `AABBundletoolProcessor` 구현체를 통해 `bundletool build-apks --bundle=input.aab --output=output.apks --ks=...` 명령으로 기기 스펙별 Split APK 또는 Universal APK(`--mode=universal`)를 생성하는 파이프라인을 기본 탑재한다.

## 4. 상세 기능 (Detailed Functional Specifications)
- **업로드 제한 및 Zip Bomb 방어**:
  - 단일 APK 최대 크기: `2,048 MB (2GB)`, AAB 최대 크기: `4,096 MB (4GB)`.
  - ZIP 엔트리 최대 개수: `65,535개`, 압축 해제 시 총 크기 비율(Compression Ratio): `100:1` 초과 시 즉시 `ERR_ZIP_BOMB_DETECTED`로 스트림을 중단하고 검역 처리한다.
- **CDN 캐시 및 파일 교체 방지 (Immutable Content-Addressable Path)**:
  - 파일 경로에 파일 자체의 `sha256` 해시를 포함하므로(`/packages/{app_id}/{release_id}/{sha256}.apk`), 동일한 URL에서 파일 내용이 바뀌는 일(Cache Poisoning)이 원천적으로 불가능하다.

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **`web-dev`**: 업로드 완료 후 추출된 `packageName`, `versionCode`, `minSdk`, `targetSdk`, 지원 ABI(`arm64-v8a` 등), 요구 권한 목록, 서명 인증서 SHA-256 지문을 화면에 표로 보여주어 개발자가 최종 확인 후 릴리스에 연결하도록 구현한다.

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/distribution/infrastructure/apk_analyzer.go`:
  - AXML 파서와 `apksigner` 래퍼를 통해 매니페스트 및 서명 블록(v1/v2/v3/v4)을 검증하고 결과를 `release_artifacts` 및 `apk_signatures` 테이블에 트랜잭션으로 저장한다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**:
  - `upload_sessions` (`id`, `org_id`, `app_id`, `expected_sha256`, `expected_size_bytes`, `s3_quarantine_key`, `status`, `expires_at`).
  - `release_artifacts` (`id`, `release_id`, `artifact_type`, `abi_filter`, `dpi_filter`, `s3_quarantine_key`, `s3_public_key`, `file_size_bytes`, `sha256_checksum UNIQUE`, `min_sdk`, `target_sdk`, `extracted_permissions JSONB`, `native_abis TEXT[]`, `signature_scheme_version INT`, `status`).
  - `apk_signatures` (`id`, `app_id`, `artifact_id`, `cert_sha256_fingerprint CHAR(64)`, `subject_dn`, `issuer_dn`, `valid_from`, `valid_until`, `proof_of_rotation JSONB`).

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `POST /api/v1/developer/apps/{appId}/uploads`: 업로드 세션 및 Presigned URL 생성.
- `POST /api/v1/developer/apps/{appId}/uploads/{sessionId}/complete`: 업로드 완료 통지 및 검증 Job 트리거.
- `POST /api/v1/catalog/apps/{appId}/releases/{releaseId}/download-token`: 클라이언트 기기 호환성 및 구매 권한 확인 후 15분 유효 CDN Signed URL 반환.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **아티팩트 상태 (`artifact_status`)**: `UPLOADING` → `UPLOADED_QUARANTINED` → `VERIFYING_INTEGRITY` → `PARSING_MANIFEST` → `VERIFYING_SIGNATURE` → `SCANNING_MALWARE` → `VERIFIED_READY` → `PUBLISHED_ON_CDN` | `REJECTED_INVALID_APK` | `REJECTED_SIGNATURE_MISMATCH` | `QUARANTINED_MALWARE` | `REVOKED_EMERGENCY`.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- **APK Signature Scheme v2 이상 필수 강제**: `minSdk >= 24`인 모든 APK는 반드시 APK Signature Scheme v2 또는 v3로 서명되어야 하며, v1(JAR 서명)만 존재하는 APK는 Janus 취약점(CVE-2017-13156) 방어를 위해 업로드를 즉시 거부(`ERR_APK_V2_SIGNATURE_REQUIRED`)한다.
- **디버그 서명 및 만료 인증서 거부**: `CN=Android Debug` 인증서로 서명된 APK나 잔여 유효기간이 부족한 인증서는 프로덕션 트랙 등록을 자동 차단한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 사용자 웹 및 Android 상세 화면에 **"보안 및 무결성 검증 정보"** 아코디언을 두어 파일 `SHA-256` 전문, 개발자 서명 인증서 `SHA-256` 지문, 서명 스킴(`v2 + v3`), 보안 스캔 통과 일시를 투명하게 공개한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **서명 불일치 업로드 거부 테스트 (`TC-SEC-007`)**: 버전 1(`Key A`)이 출시된 앱에 버전 2를 다른 키(`Key B`, 회전 증명 없음)로 서명해 업로드했을 때 `REJECTED_SIGNATURE_MISMATCH`로 자동 차단되는지 검증.
- **손상된 매니페스트/v1 전용 APK 거부 테스트 (`TC-FUNC-009`)**: v1 서명만 있는 APK 업로드 시 `ERR_APK_V2_SIGNATURE_REQUIRED` 반환 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `J — Job Processing`, `K — Kotlin Android Market`, `L — Lifecycle & Release Management`, `S — Security & Safety`, `V — Versioning & Update Engine`과 직결된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0 (플랫폼 코어 엔진)`
- **개발 Phase**: `Phase 3` 및 `Phase 4`에서 최우선 구현한다 (`AM-DIST-001`, `AM-DIST-002`).

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: 출시된 앱에서 사후 제로데이 악성 행위 발견 (Emergency Takedown)** → 관리자가 `Emergency Kill Switch`를 발동하면 ① 즉시 `bn-public-cdn-bucket`의 객체 ACL 차단 및 CDN Edge 캐시 Purge API 호출, ② DB 상태를 `REVOKED_EMERGENCY`로 변경하여 신규 다운로드 토큰 발급 중단, ③ 해당 버전을 설치한 모든 활성 Android 클라이언트에 보안 경고 푸시 알림(`SECURITY_UNINSTALL_ALERT`)을 발송한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- APK/AAB 파서, 서명 검증기, `bundletool` 래퍼, CDN Signed URL 발급기 (`backend/internal/modules/distribution/*`).

## 17. Acceptance Criteria (수용 기준)
1. 실제 서명된 샘플 APK 업로드 시 패키지명, 버전코드, SDK 레벨, 권한 목록, 인증서 SHA-256이 100% 정확히 DB에 추출되어야 한다.
2. 변조된 APK나 다른 키로 서명된 업데이트 APK가 100% 자동 거부되어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-DIST-001/apk_parser_signature_test.log`: 정상/변조/키불일치 APK 5종에 대한 파이프라인 검증 로그.
- `artifacts/evidence/AM-DIST-002/cdn_signed_url_expiry_test.log`: 서명된 다운로드 URL 정상 다운로드 및 만료 후 `403` 차단 로그.

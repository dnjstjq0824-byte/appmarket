# API_TEST_CASES.md — 19대 API 그룹 실제 호출 예시 및 계약·보안 검증 명세

- **문서 ID**: `DOC-API-003`
- **관련 문서**: [openapi.yaml](openapi.yaml), [AUTHORIZATION_MATRIX.md](AUTHORIZATION_MATRIX.md), [API_ERROR_CONTRACT.md](API_ERROR_CONTRACT.md)

---

## 1. 핵심 API 실제 호출 예시 및 서버 측 검증 시나리오 (제9조 준수)

### 1.1 APK 격리 업로드 세션 생성 및 완료 (`06_Upload`)
- **호출 예시**:
  ```http
  POST /api/v1/developer/apps/0192ce10-4b00-7000-8000-000000000101/uploads HTTP/1.1
  Host: api.appmarket.bookingnetwork.internal
  Authorization: Bearer <DEV_RELEASE_MGR_JWT>
  Idempotency-Key: 0192ce10-9911-7a00-b000-112233445566
  Content-Type: application/json

  {
    "artifact_type": "APK",
    "file_name": "com.example.superapp-v105-release.apk",
    "expected_size_bytes": 44879872,
    "expected_sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
  }
  ```
- **서버 측 검증 단계**:
  1. JWT 서명 검증 및 `developer_org_members`에서 호출자가 해당 `appId` 소유 조직의 `DEV_OWNER` 또는 `DEV_RELEASE_MGR`인지 확인.
  2. `expected_size_bytes <= 2147483648` (2GB) 확인 및 `expected_sha256` 64자리 16진수 형식 검증.
  3. `bn-quarantine-bucket`에 `CheckSumSHA256` 조건이 고정된 S3 Presigned Multipart URL 생성 후 `201 Created` 반환.

### 1.2 Android 기기 일괄 업데이트 및 서명 호환성 확인 (`09_InstallUpdateEvent`)
- **호출 예시**:
  ```http
  POST /api/v1/device/check-updates HTTP/1.1
  Host: api.appmarket.bookingnetwork.internal
  Content-Type: application/json

  {
    "device_sdk": 34,
    "supported_abis": ["arm64-v8a", "armeabi-v7a"],
    "density_dpi": 480,
    "available_features": ["android.hardware.camera", "android.hardware.nfc"],
    "installed_packages": [
      {
        "package_name": "org.videolan.vlc",
        "version_code": 13050200,
        "signing_cert_sha256": "9c3a8b12f0e14d22a567890123456789abcdef0123456789abcdef0123456789",
        "installer_package": "com.bookingnetwork.appmarket"
      }
    ]
  }
  ```
- **응답 예시 (`200 OK`)**:
  ```json
  {
    "updates": [
      {
        "package_name": "org.videolan.vlc",
        "app_id": "0192ce10-4b00-7000-8000-000000000202",
        "release_id": "0192ce10-4b00-7000-8000-000000000999",
        "installed_version_code": 13050200,
        "latest_version_code": 13050400,
        "latest_version_name": "3.5.4",
        "update_status": "UPDATE_AVAILABLE_COMPATIBLE",
        "signature_compatibility": "COMPATIBLE_SAME_SIGNER",
        "new_dangerous_permissions": [],
        "file_size_bytes": 38912000,
        "expected_sha256": "7d865e959b2466918c9863afca942d0fb89d7c9ac0c99bafc3749504ded97730"
      }
    ]
  }
  ```

---

## 2. API 보안 및 엣지 케이스 필수 테스트 매트릭스

| 테스트 ID | 대상 API | 검증 시나리오 | 요청 조건 | 기대 HTTP 응답 및 에러 코드 |
| :---: | :--- | :--- | :--- | :--- |
| `TC-API-001` | `POST .../download-token` | 속도 제한(Rate Limit) 초과 방어 | 동일 IP/User로 1분 내 31번째 호출 | `429 Too Many Requests` (`ERR_RATE_LIMIT_EXCEEDED`, `Retry-After` 헤더 포함) |
| `TC-API-002` | `POST /api/v1/commerce/webhooks/stripe` | 위조된 PG 웹훅 서명 차단 | 잘못된 `Stripe-Signature` 헤더 전송 | `401 Unauthorized` (`ERR_WEBHOOK_SIGNATURE_INVALID`) |
| `TC-API-003` | `GET /api/v1/treescan/federated-search` | 독립 모드 및 활성화 모드 전환 검증 | 기본 모드에서 호출 vs 활성화 후 서비스 토큰 호출 | 기본 모드 `503` (`ERR_TREESCAN_ADAPTER_DISABLED`), 활성화 시 `200 OK` (`< 80ms`) |
| `TC-API-004` | `POST /api/v1/commerce/orders` | 멱등성 키(Idempotency-Key) 재전송 및 충돌 검증 | 동일 키+동일 본문 재전송 vs 동일 키+다른 `app_id` 전송 | 동일 본문은 최초 `201` 응답 그대로 반환(중복 생성 없음), 다른 본문은 `409 ERR_IDEMPOTENCY_KEY_CONFLICT` |
| `TC-API-005` | `POST /api/v1/admin/reviews/{id}/decision` | 권한 없는 일반 유저/타 역할 관리자 접근 차단 | `USER` 또는 `ADMIN_SEARCH_OPS` JWT로 호출 | `403 Forbidden` (`ERR_INSUFFICIENT_ROLE_PERMISSION`) |

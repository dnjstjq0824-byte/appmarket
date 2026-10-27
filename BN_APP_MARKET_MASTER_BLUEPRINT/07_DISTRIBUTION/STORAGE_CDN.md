# STORAGE_CDN.md — 격리/공개 오브젝트 스토리지 분리, 불변 CDN 캐싱 및 긴급 차단 설계

- **문서 ID**: `DOC-DIST-003`
- **관련 조항**: 제10조
- **연관 Work Order**: `AM-DIST-002`, `AM-OPS-001`

---

## 1. 버킷 분리 및 콘텐츠 주소 지정(Content-Addressable) 불변 경로

1. **격리 버킷 (`bn-quarantine-bucket`)**:
   - 경로: `quarantine/{org_id}/{app_id}/{upload_session_id}.tmp`
   - 정책: 외부 네트워크/CDN에서 접근 완전 차단(`BlockPublicAccess = true`). 오직 `worker-apk-analyzer`, `worker-malware-scanner`, `worker-cdn-publisher` IAM 역할만 읽기 가능.
2. **공개 배포 버킷 (`bn-public-cdn-bucket`)**:
   - 경로: `packages/{app_id}/{release_id}/{sha256_checksum}.apk`
   - 정책: 파일명 자체가 64자리 `sha256_checksum`이므로 내용 변조나 덮어쓰기가 불가능합니다. CDN 응답 헤더에 `Cache-Control: public, max-age=31536000, immutable` 및 `Content-Disposition: attachment; filename="{package_name}-{version_code}.apk"`를 설정합니다.

---

## 2. HMAC 서명 다운로드 URL 및 긴급 배포 중단 (Emergency Kill Switch)

- **Signed URL 구조**:
  `https://cdn.appmarket.bookingnetwork.internal/packages/{app_id}/{release_id}/{sha256}.apk?exp={unix_ts}&uid={user_or_device_hash}&sig={HMAC_SHA256}`
- **긴급 차단(Kill Switch) 실행 시 3단계 즉시 차단**:
  1. PostgreSQL `app_releases.release_status = 'REVOKED_BY_ADMIN'` 및 Redis 블랙리스트(`revoked:release:{release_id}`) 등록 → 0.1초 내 신규 다운로드 토큰 발급 차단.
  2. `bn-public-cdn-bucket`에서 해당 객체 태그를 `status=quarantined_revoked`로 변경하여 Origin 접근 차단 + CDN Purge API(`POST /purge`) 호출로 전 세계 Edge 캐시 즉시 파기.
  3. `install_events`에서 해당 `release_id`를 설치한 기기 목록을 조회하여 보안 삭제 권고 알림 발송.

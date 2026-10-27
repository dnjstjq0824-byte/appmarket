# THREAT_MODEL.md — STRIDE 위협 모델링 및 심층 방어(Defense-in-Depth) 매트릭스

- **문서 ID**: `DOC-ARCH-005`
- **관련 문서**: [01_A_TO_Z_BLUEPRINT/S_SECURITY_AND_SAFETY.md](../01_A_TO_Z_BLUEPRINT/S_SECURITY_AND_SAFETY.md), [09_SECURITY_COMPLIANCE/SECURITY_ARCHITECTURE.md](../09_SECURITY_COMPLIANCE/SECURITY_ARCHITECTURE.md)

---

## 1. STRIDE 기반 12대 핵심 위협 및 통제 설계

| 위협 ID | STRIDE 분류 | 위협 시나리오 | 공격 대상 | 기술적 방어 및 완화 통제 (Mitigation Controls) | 검증 테스트 ID |
| :---: | :--- | :--- | :--- | :--- | :---: |
| `THR-01` | **Spoofing** | 유명 은행/메신저 앱의 패키지명이나 아이콘을 도용한 피싱 앱 등록 | `catalog`, `distribution` | 보호 브랜드 패키지 접두사(`protected_brands`) 차단, 역방향 DNS TXT 소유권 검증, 아이콘/이름 유사도 사칭 탐지 큐 강제 라우팅 | `TC-SEC-009` |
| `THR-02` | **Spoofing** | 타 개발자 조직의 앱에 악성 업데이트 APK를 업로드하여 기존 사용자 감염 | `distribution`, `apk_signatures` | 최초 등록된 서명 인증서 SHA-256 지문(`apk_signatures`) 고정 강제. 서명 불일치 시 업로드 즉시 `REJECTED_SIGNATURE_MISMATCH` 차단 | `TC-SEC-007` |
| `THR-03` | **Tampering** | 심사 승인된 정상 APK를 배포 직전 또는 CDN 구간에서 악성 APK로 바꿔치기 | Storage, CDN, Android Client | 콘텐츠 주소 지정 불변 경로(`/packages/{app_id}/{release_id}/{sha256}.apk`), S3 Object Lock, Android 클라이언트 설치 직전 로컬 `SHA-256` 재검증 | `TC-AND-001`, `TC-OPS-003` |
| `THR-04` | **Tampering** | APK 파서 워커를 노린 Zip Bomb, Zip Slip(`../../`), 기형 AXML 공격 | `worker-apk-analyzer` | 압축 해제 비율 100:1 제한, 엔트리 경로 정규화(`filepath.Clean` 및 상위 디렉터리 탈출 차단), 외부 통신이 차단된 비루트 임시 컨테이너 실행 | `TC-SEC-009` |
| `THR-05` | **Repudiation** | 관리자가 악성 앱을 부정 승인하거나 정산금을 임의 조작 후 로그 삭제 | `admin`, `audit_logs` | DB 레벨에서 `bn_app_rw` 계정의 `audit_logs` `UPDATE/DELETE` 권한 회수 + 이전 레코드 해시를 연결한 암호학적 해시 체인(`prev_hash -> record_hash`) | `TC-SEC-001` |
| `THR-06` | **Information Disclosure** | 개발자 조직 A가 URL의 `appId`를 조작해 조직 B의 비공개 빌드 및 수익 조회 (IDOR) | `developer`, `commerce` | 모든 Developer API에서 JWT의 소속 `org_id`와 리소스 소유 `org_id`를 교차 검증하는 미들웨어 적용 | `TC-SEC-002` |
| `THR-07` | **Information Disclosure** | 개발자 KYB 신분증/사업자등록증 스토리지 버킷 외부 노출 | `bn-kyb-vault-bucket` | 퍼블릭 액세스 완전 차단, KMS 서버측 암호화(`SSE-KMS`), `ADMIN_KYB_REVIEWER` 역할에만 5분 유효 일회용 열람 URL 발급 및 열람 감사 기록 | `TC-SEC-002` |
| `THR-08` | **Denial of Service** | 대용량 APK 다운로드 URL을 봇넷으로 무한 호출하여 CDN 대역폭 비용 폭탄 유발 | Distribution API, CDN | 다운로드 토큰 발급 Rate Limit(기기/IP당 분당 30회), 유료 앱 구매 권한 선검증, 15분 단기 만료 및 IP 바인딩 Signed URL 적용 | `TC-API-001` |
| `THR-09` | **Elevation of Privilege** | 일반 사용자 또는 단일 관리자가 단독으로 대규모 정산 지급이나 슈퍼관리자 권한 탈취 | `iam`, `admin` | 전 관리자 MFA 강제, 민감 작업(권한 부여, 정산 집행, 앱 영구 차단) 2인 이중 승인(`Dual-Control`) 강제 | `TC-SEC-006` |
| `THR-10` | **Tampering (Commerce)** | 결제 요청 파라미터의 가격 변조 또는 가짜 PG 결제 완료 웹훅 전송 | `commerce` | 클라이언트 전달 가격 무시(서버 DB 정가표 조회), PG 웹훅 `HMAC-SHA256` 서명 검증, 복식부기 대차평형 트리거 강제 | `TC-COM-004`, `TC-API-002` |
| `THR-11` | **Tampering (Android)** | 중간자 공격을 통한 구버전 취약 APK 다운그레이드 설치 유도 | `android-client` | 클라이언트에서 `newVersionCode > installedVersionCode` 및 서명 인증서 일치 여부를 `PackageInstaller` 호출 전 검증 | `TC-AND-003` |
| `THR-12` | **Supply Chain (External)** | 외부 커넥터(F-Droid 등) 동기화 시 변조된 미러 피드 주입 | `connector` | `index-v2.json` / `entry.json`에 대한 리포지토리 공개키 암호학적 서명 검증 실패 시 전체 배치 즉시 폐기 | `TC-SEC-003` |

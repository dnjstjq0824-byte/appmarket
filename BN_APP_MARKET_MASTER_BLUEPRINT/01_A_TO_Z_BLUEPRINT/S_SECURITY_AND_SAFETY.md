# S — Security & Safety — 악성코드 스캔, 서명/공급망 보안, 사칭 방지, 신고·제재 및 위협 모델 대응 (Malware Scanning, Supply-Chain Security, Anti-Impersonation, Abuse Reporting & Threat Mitigation)

- **문서 ID**: `DOC-AZ-S`
- **알파벳 코드**: `S`
- **관련 요구사항 ID**: `REQ-S-001`, `REQ-S-002`, `REQ-S-003`
- **연관 Work Order ID**: `AM-SEC-001`, `AM-SEC-002`, `AM-MOD-001`
- **연관 테스트 케이스 ID**: `TC-SEC-009`, `TC-SEC-010`
- **참조 상세 설계 문서**: [02_SYSTEM_ARCHITECTURE/THREAT_MODEL.md](../02_SYSTEM_ARCHITECTURE/THREAT_MODEL.md), [09_SECURITY_COMPLIANCE/SECURITY_ARCHITECTURE.md](../09_SECURITY_COMPLIANCE/SECURITY_ARCHITECTURE.md), [09_SECURITY_COMPLIANCE/APP_REVIEW_POLICY.md](../09_SECURITY_COMPLIANCE/APP_REVIEW_POLICY.md)
- **현재 구현 상태**: `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`)

---

## 1. 개발 목적 (Development Purpose)
- **왜 필요한가**: 제13조에서 명령한 바와 같이, 상용 앱 마켓은 트로이목마·랜섬웨어·스파이웨어 등 악성 APK 배포, 유명 은행/메신저 앱 사칭(Impersonation), 다운로드 스트림 변조, 권한 상승, 개인정보 유출로부터 사용자와 개발자를 보호해야 한다.
- **달성 목표**: OWASP MASVS 및 OWASP ASVS Level 2 기준을 충족하는 다중 계층 APK 보안 스캐너, 개발자 신뢰도 점수(Trust Score), 패키지/아이콘 사칭 탐지기, 사용자 신고 및 긴급 차단(Kill Switch) 대응 체계를 명세한다.

## 2. 주요 사용자와 역할 (Target Users & Roles)
- **Security Analyst (관리자)**: 악성코드 탐지 리포트 및 고위험 권한 요청 앱 정밀 분석, YARA 룰셋 업데이트, 침해 사고 시 긴급 배포 중단 및 인증서 블랙리스트 등록.
- **Trust & Safety Moderator**: 사용자의 유해 앱/저작권 침해/사칭 신고 접수 및 심사 처리.
- **일반 사용자**: 앱 상세 화면에서 보안 스캔 결과 및 요구 권한의 위험도를 확인하고, 의심스러운 앱을 신고.

## 3. 필수 기능 (Mandatory Features)
1. **4단계 자동화 APK 보안 검사 파이프라인 (`worker-malware-scanner`)**:
   - Stage 1 (구조/무결성): ZIP 중앙 디렉터리 변조, 이중 매니페스트, 비정상 ELF/SO 헤더 검사.
   - Stage 2 (시그니처/YARA 스캔): ClamAV 엔진 + Android 악성코드 전용 YARA 룰셋(Joker, FluBot, Anatsa/TeaBot, 뱅킹 트로이목마, 난독화된 동적 DEX 로더 `DexClassLoader` 패턴 등) 검사.
   - Stage 3 (권한 위험도 정적 분석): `BIND_ACCESSIBILITY_SERVICE`, `REQUEST_INSTALL_PACKAGES`, `RECEIVE_SMS`, `READ_CALL_LOG`, `SYSTEM_ALERT_WINDOW`, `BIND_DEVICE_ADMIN` 등 고위험 권한 조합을 분석하여 `risk_score (0~100)` 산출. `risk_score >= 70`인 앱은 자동 승인을 금지하고 `ADMIN_SECURITY` 심층 심사 큐로 강제 라우팅.
   - Stage 4 (사칭 및 브랜드 보호 검사): 유명 앱의 패키지명 유사도(Levenshtein Distance / Homograph), 앱 이름, 서명 인증서 지문 불일치 여부를 대조하여 사칭 의심 경고 생성.
2. **보안 문구 정직성 원칙 (제13조 준수)**:
   - UI나 마케팅 문구에 `"100% 안전 보장"` 또는 `"바이러스 완전 무결"`이라는 과장 표현을 절대 사용하지 않으며, `"자동 정적 분석 및 악성코드 시그니처 검사 통과 (검사일: YYYY-MM-DD, 엔진 버전: vX.Y)"`와 같이 확인된 사실만 정확히 표시한다.
3. **신고(Report) 및 권리 침해(DMCA/Trademark) 대응 워크플로**:
   - 악성행위, 사칭, 저작권 침해, 개인정보 침해, 결제 사기 카테고리별 신고 접수 및 SLA(악성코드 신고는 4시간 이내 1차 검토) 관리.

## 4. 상세 기능 (Detailed Functional Specifications)
- **제13조 16대 보안 검토 항목 전수 매핑**:
  - (1) 인증/접근권한 (`I`), (2) 악성코드 (`worker-malware-scanner`), (3) APK 서명 (`N`), (4) 비밀키 보호 (KMS/Vault), (5) 데이터 암호화 (TLS 1.3 + AES-256-GCM), (6) API 남용 (Redis Rate Limiter + WAF), (7) 공급망 공격 (의존성 체크섬 고정 + SBOM + 컨테이너 서명), (8) 관리자 권한 탈취 (MFA + IP 제한 + Dual-Control), (9) 파일 다운로드 변조 (불변 SHA-256 경로 + 클라이언트 재검증), (10) 개인정보 노출 (PII 마스킹/암호화), (11) 결제 부정행위 (PG 웹훅 검증 + 14일 정산 보류), (12) 앱 위장/사칭 (패키지명/서명/이름 유사도 대조), (13) 저작권/배포권 (5단계 권리 등급 및 DMCA 워크플로), (14) 개발자 신원 (KYB + DNS 검증), (15) Android 설치 제약 준수 (`K`), (16) 국가별 배포 제한 (`G`).

## 5. 프론트엔드 구현 구조 (Frontend Implementation Structure)
- **`web-user`**: `SecurityScanReportCard.vue` (검사 일시, 서명 검증 상태, 고위험 권한 설명 표시), `ReportAbuseModal.vue` (증빙 캡처 첨부 신고 모달).
- **`web-admin`**: `SecurityScannerQueueView.vue`, `AbuseReportsModerationView.vue`, `CertBlacklistManagerView.vue`.

## 6. 백엔드 처리 구조 (Backend Processing Structure)
- `internal/modules/security/` 및 `internal/modules/moderation/`:
  - 블랙리스트에 등록된 서명 인증서(`blacklisted_cert_sha256`)로 서명된 APK는 개발자 계정을 바꿔 재업로드하더라도 업로드 단계에서 즉시 영구 차단하고 해당 계정을 자동 동결(`SUSPENDED_SECURITY`)한다.

## 7. DB 및 데이터 모델 (Database & Data Model)
- **핵심 테이블**: `security_scan_reports`, `blacklisted_certificates`, `protected_brands`, `moderation_reports`, `moderation_actions`, `appeals`(개발자 이의신청).

## 8. API 및 이벤트 계약 (API & Event Contracts)
- `POST /api/v1/moderation/reports`: 사용자/권리자의 앱·리뷰 신고 접수.
- `GET /api/v1/admin/security/scans/{artifactId}`: 상세 정적 분석 및 YARA 매칭 결과 조회.
- `POST /api/v1/admin/security/kill-switch/{appId}`: 긴급 배포 중단 및 기존 설치자 보안 경고 발송.

## 9. 상태 전이와 예외 처리 (State Transitions & Exception Handling)
- **보안 스캔 상태**: `QUEUED` → `SCANNING` → `CLEAN_LOW_RISK` | `CLEAN_HIGH_RISK_MANUAL_REVIEW` | `INFECTED_QUARANTINED` | `SCAN_ERROR`.

## 10. 보안·권한·개인정보 요건 (Security, Authorization & Privacy Requirements)
- 스캐너 워커 자체가 악성 APK에 의해 침해되는 것을 막기 위해 네트워크 Egress가 완전히 차단된 샌드박스 컨테이너에서 최대 60초 후 컨테이너 컨텍스트를 초기화하며, 파일시스템은 메모리 마운트(`tmpfs`)만 사용한다.

## 11. UX/UI 요구사항 (UX/UI Requirements)
- 고위험 권한(예: ` 접근성 서비스`, `문자 메시지 읽기`, `다른 앱 위에 그리기`)을 요구하는 앱은 상세 페이지 상단에 황색 주의 배너로 해당 권한이 왜 민감한지 사용자 눈높이 언어로 안내한다.

## 12. 테스트 및 검증 기준 (Testing & Verification Criteria)
- **EICAR 및 테스트 악성 YARA 패턴 격리 테스트 (`TC-SEC-009`)**: EICAR 테스트 문자열 및 고위험 YARA 테스트 시그니처가 포함된 APK 업로드 시 `INFECTED_QUARANTINED`로 분류되고 공개 버킷 복사가 100% 차단되는지 검증.
- **블랙리스트 인증서 재업로드 차단 테스트 (`TC-SEC-010`)**: 차단된 인증서 지문으로 서명된 신규 앱 업로드 시 즉시 거부되는지 검증.

## 13. 다른 모듈과의 의존성 (Cross-Module Dependencies)
- `N — Native App Distribution`, `L — Lifecycle`, `W — Workflow & Administration`과 연동된다.

## 14. 개발 우선순위 (Development Priority & Phase Mapping)
- **우선순위**: `P0 (보안 필수)`
- **개발 Phase**: 기본 서명/무결성 및 정적 권한 분석은 `Phase 3~4`, 고도화된 스캔·신고·제재·감사는 `Phase 6`에서 완성한다.

## 15. 예상 장애와 대응책 (Failure Modes & Mitigation Strategies)
- **장애 시나리오: ClamAV/YARA 스캐너 데몬 응답 불능** → Fail-Closed 원칙을 적용하여 보안 스캔이 완료되지 않은 바이너리는 절대 `READY_FOR_SUBMISSION`이나 `PUBLISHED`로 넘어가지 못하고 `SCAN_PENDING_RETRY` 상태에서 대기한다.

## 16. 최종 산출물 (Target Implementation Deliverables)
- 보안 스캐너 워커, 권한 위험도 분석기, 사칭 탐지기, 신고/제재 백엔드 모듈 (`backend/internal/modules/security/*`, `backend/internal/modules/moderation/*`).

## 17. Acceptance Criteria (수용 기준)
1. 보안 스캔을 통과하지 않은 단 하나의 APK도 공개 스토리지(`bn-public-cdn-bucket`)로 이동할 수 없어야 한다.
2. UI 및 API 어디에도 `"100% 안전 보장"` 문구가 없어야 하며 실제 스캔 시각과 스캔 범위가 표시되어야 한다.

## 18. 해당 기능의 실제 구현 여부를 확인할 증빙 방법 (Verification Evidence & Audit Artifacts)
- `artifacts/evidence/AM-SEC-002/malware_quarantine_test.log`: 악성 시그니처 샘플 APK 격리 및 배포 차단 테스트 로그.

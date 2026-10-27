# AM-SEC-002 — 격리 악성코드 스캐너(ClamAV + YARA), 고위험 권한 점수 산출기, 브랜드 사칭 탐지기 및 Kill Switch 구현

## Work Order 필수 명세 (제14조 준수)

| 항목 | 상세 내용 |
| :--- | :--- |
| **작업 ID (Work Order ID)** | `AM-SEC-002` |
| **작업명 (Title)** | 격리 악성코드 스캐너(ClamAV + YARA), 고위험 권한 점수 산출기, 브랜드 사칭 탐지기 및 Kill Switch 구현 |
| **목표 (Objective)** | `REQ-S-001;REQ-S-002;REQ-S-003` 요구사항을 충족하는 상용 수준 코드 및 자동화 검증 구현 |
| **담당 권장 에이전트 (Assigned Agent)** | `Security Agent` |
| **선행 작업 ID (Prerequisites)** | `AM-DIST-001` |
| **개발 Phase 및 우선순위** | `Phase 3` / `P0` |
| **입력 문서 (Input Documents)** | `01_A_TO_Z_BLUEPRINT/S_SECURITY_AND_SAFETY.md;09_SECURITY_COMPLIANCE/SECURITY_ARCHITECTURE.md;09_SECURITY_COMPLIANCE/APP_REVIEW_POLICY.md` |
| **수정 가능 디렉터리 (Writable Directories)** | `backend/internal/modules/security/;backend/internal/worker/handlers/scanner/` |
| **변경 금지 디렉터리 (Forbidden Directories)** | `apps/;android-client/` |
| **예상 생성/수정 파일 (Target Files)** | `backend/internal/modules/security/scanner/clamav_yara_scanner.go;backend/internal/modules/security/scanner/permission_risk_scorer.go;backend/internal/modules/security/scanner/impersonation_detector.go;backend/internal/modules/security/rules/android_malware.yar` |
| **API 의존성 (API Dependencies)** | `GET /api/v1/admin/security/scans/{artifactId};POST /api/v1/admin/security/kill-switch/{appId}` |
| **DB 변경 여부 (DB Schema Changes)** | `NO` |
| **필수 테스트 ID (Mandatory Tests)** | `TC-SEC-009;TC-SEC-010` |
| **증빙 파일 저장 경로 (Evidence Path)** | `artifacts/evidence/AM-SEC-002/` |
| **현재 구현 상태 (Initial Status)** | `NOT_STARTED` (설계 검증 완료: `DESIGN_VERIFIED`) |

---

## 1. 구현 상세 (Implementation Details)
1. 외부 네트워크 Egress가 차단된 워커에서 ClamAV 및 Android 악성코드 전용 YARA 룰 스캔 수행.
2. `BIND_ACCESSIBILITY_SERVICE`, `READ_SMS` 등 고위험 권한 조합을 분석해 `risk_score (0~100)`를 산출하고 70점 이상 시 보안 심층 심사 큐로 강제 배정.
3. 보호 브랜드 패키지명/이름 유사도 사칭 탐지 및 블랙리스트 서명 인증서(`blacklisted_certificates`) 즉시 차단 구현.

## 2. Acceptance Criteria (수용 기준)
1. EICAR 및 테스트 악성 YARA 패턴이 포함된 APK가 `INFECTED_QUARANTINED`로 격리되고 공개 버킷 복사가 원천 차단.
2. 블랙리스트 인증서로 서명된 APK 재업로드 즉시 차단.

## 3. 실패 처리 및 롤백 지침 (Failure Handling & Recovery)
- 스캐너 엔진 오류(`SCAN_ERROR`) 시 절대 자동 통과시키지 않는 Fail-Closed 정책 강제.

## 4. 완료 정의 (Definition of Done)
1. `수정 가능 디렉터리` 내에 실제 동작하는 상용 코드가 작성되고 Mock/더미 우회 코드가 0건일 것.
2. 필수 테스트(`TC-SEC-009;TC-SEC-010`)를 실제로 실행하여 통과 로그와 커버리지 리포트를 `artifacts/evidence/AM-SEC-002/`에 생성할 것.
3. 담당 에이전트가 `IMPLEMENTED_UNVERIFIED`로 보고한 뒤, `Independent Audit Agent`의 교차 검증을 통과하여 `VERIFIED` 승인을 획득할 것.

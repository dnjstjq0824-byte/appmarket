# APP_REVIEW_POLICY.md — 앱 등록·업데이트 심사 정책, 고위험 권한 기준 및 제재 규정

- **문서 ID**: `DOC-SEC-002`
- **관련 조항**: 제7조, 제10조, 제13조

---

## 1. 고위험 Android 권한 (`Dangerous & High-Risk Permissions`) 심사 기준

업로드된 APK의 `AndroidManifest.xml`에서 아래 권한이 추출될 경우 자동 승인이 차단되며, 개발자는 심사 제출 시 **권한 사용 목적 소명서 및 시연 영상**을 필수로 첨부해야 합니다:

| 권한 명칭 (`android.permission.*`) | 위험 점수 (`risk_score`) | 허용되는 핵심 기능 범위 (Permitted Use Cases) | 위반 시 조치 |
| :--- | :---: | :--- | :--- |
| `BIND_ACCESSIBILITY_SERVICE` | `+45점` | 장애인 보조 기술 또는 사용자가 명시적으로 인지·동의한 UI 자동화 유틸리티 | 무단 백그라운드 클릭/정보 수집 시 즉시 `REJECTED` 및 계정 제재 (`POL-SEC-001`) |
| `REQUEST_INSTALL_PACKAGES` | `+35점` | 파일 관리자, 백업/복원 도구, 기업용 MDM, 연동 패키지 관리자 | 본래 목적과 무관한 드로퍼(Dropper) 행위 시 영구 차단 (`POL-SEC-002`) |
| `READ_SMS` / `RECEIVE_SMS` / `READ_CALL_LOG` | `+35점` | 기본 전화/메시지 핸들러 앱 또는 법령상 허용된 금융 본인인증 보안 모듈 | OTP 탈취 위험으로 인해 일반 유틸리티/게임 앱에서는 무조건 반려 (`POL-SEC-003`) |
| `BIND_DEVICE_ADMIN` | `+40점` | 엔터프라이즈 기기 관리(EMM) 및 분실 방지 보안 앱 | 삭제 방해(Ransomware 패턴) 악용 시 즉시 검역 (`POL-SEC-004`) |
| `SYSTEM_ALERT_WINDOW` | `+20점` | 화면 위 플로팅 번역기, 녹화 컨트롤러, 통화 알림창 | 타 앱 로그인 창을 덮어쓰는 오버레이 피싱(Cloak & Dagger) 여부 집중 심사 (`POL-SEC-005`) |
| `QUERY_ALL_PACKAGES` | `+15점` | 런처, 보안 스캐너, 기기 간 마이그레이션, 파일/앱 관리 도구 | 광고 타기팅 목적의 설치 앱 목록 무단 수집 금지 (`POL-PRIV-001`) |

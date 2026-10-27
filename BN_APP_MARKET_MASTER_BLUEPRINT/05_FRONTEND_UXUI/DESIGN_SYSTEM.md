# DESIGN_SYSTEM.md — 통합 디자인 시스템 토큰, 타이포그래피, 컬러 및 공통 컴포넌트 규격

- **문서 ID**: `DOC-UX-004`
- **관련 문서**: [01_A_TO_Z_BLUEPRINT/U_USER_EXPERIENCE_AND_UI.md](../01_A_TO_Z_BLUEPRINT/U_USER_EXPERIENCE_AND_UI.md)

---

## 1. CSS Custom Properties 디자인 토큰 (`packages/ui-kit/src/tokens.css`)

```css
:root, [data-theme="light"] {
  /* Surface & Background */
  --bn-bg-canvas: #F8FAFC;
  --bn-bg-surface: #FFFFFF;
  --bn-bg-elevated: #F1F5F9;
  --bn-border-subtle: #E2E8F0;

  /* Text Hierarchy (WCAG AA >= 4.5:1 guaranteed) */
  --bn-text-primary: #0F172A;
  --bn-text-secondary: #334155;
  --bn-text-muted: #64748B;

  /* Brand & Semantic Actions */
  --bn-primary: #2563EB;
  --bn-primary-hover: #1D4ED8;
  --bn-trust-verified: #059669;   /* Direct Verified APK & Clean Signature */
  --bn-external-link: #4F46E5;    /* External Official Store Outlink */
  --bn-warning-amber: #D97706;    /* High-Risk Permission / Signer Conflict */
  --bn-danger-red: #DC2626;       /* Malware Quarantined / Revoked */
}

[data-theme="dark"] {
  --bn-bg-canvas: #0B1120;
  --bn-bg-surface: #1E293B;
  --bn-bg-elevated: #334155;
  --bn-border-subtle: #334155;
  --bn-text-primary: #F8FAFC;
  --bn-text-secondary: #CBD5E1;
  --bn-text-muted: #94A3B8;
  --bn-primary: #60A5FA;
  --bn-primary-hover: #3B82F6;
  --bn-trust-verified: #34D399;
  --bn-external-link: #818CF8;
  --bn-warning-amber: #FBBF24;
  --bn-danger-red: #F87171;
}
```

---

## 2. 핵심 공통 컴포넌트 규격 (`packages/ui-kit`)

1. **`BnInstallActionButton`**: `distribution_mode`와 `rights_level`을 Props로 받아:
   - `DIRECT_HOSTED` / `FEDERATED_REPO_BINARY`: 녹색/파란색 다운로드 아이콘 + 파일 용량 + `직접 APK 다운로드` 렌더링.
   - `EXTERNAL_LINK_ONLY`: 외부 링크 화살표(`↗`) + 스토어명(`Google Play 공식 페이지로 이동`) 렌더링.
2. **`BnHashFingerprintBadge`**: 64자리 SHA-256 파일 체크섬과 서명 인증서 지문을 고정폭 폰트(`JetBrains Mono`)로 축약(`9c3a:8b12...4567`) 표시하고 클릭 시 전문 펼침 및 클립보드 복사 수행.
3. **`BnSponsoredBadge`**: 광고 법규 준수를 위해 배경 대비 7:1 이상의 명도를 가진 고정 배지(`광고 · Sponsored`)를 렌더링하며 CSS로 숨길 수 없도록 강제.

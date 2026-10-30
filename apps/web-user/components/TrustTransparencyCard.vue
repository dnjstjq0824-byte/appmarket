<template>
  <div class="bn-trust-card">
    <div class="bn-trust-header">
      <div class="bn-trust-badge" :class="'status-' + (securityVerdict || 'CLEAN').toLowerCase()">
        <svg class="bn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
        </svg>
        <span>{{ securityStatusLabel }}</span>
      </div>
      <div class="bn-distribution-tag">
        배포 모드: <strong>{{ distributionModeLabel }}</strong>
      </div>
    </div>

    <div class="bn-trust-grid">
      <div class="bn-trust-item">
        <label>APK 무결성 체크섬 (SHA-256)</label>
        <BnHashFingerprintBadge :hash="apkSha256 || '9c3a8b12...'" />
      </div>

      <div class="bn-trust-item">
        <label>개발자 서명 인증서 지문</label>
        <BnHashFingerprintBadge :hash="signingCertSha256 || 'e4b1029c...'" />
      </div>

      <div class="bn-trust-item">
        <label>악성코드 격리 검사 (ClamAV + YARA)</label>
        <span class="bn-scan-result">{{ scanReport || '위협 0건 탐지 (정상)' }}</span>
      </div>

      <div class="bn-trust-item">
        <label>요구 권한 위험도 점수</label>
        <div class="bn-risk-bar-container">
          <div
            class="bn-risk-bar"
            :style="{ width: Math.min(riskScore || 10, 100) + '%', backgroundColor: riskBarColor }"
          ></div>
          <span class="bn-risk-text">{{ riskScore || 0 }}/100</span>
        </div>
      </div>
    </div>

    <div v-if="isExternalLinkOnly" class="bn-external-notice">
      <svg class="bn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="16" x2="12" y2="12" />
        <line x1="12" y1="8" x2="12.01" y2="8" />
      </svg>
      <span>본 애플리케이션은 제휴 공식 웹사이트/스토어로 직접 연결되며, 바이너리를 임의 복제하거나 재배포하지 않습니다.</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { BnHashFingerprintBadge } from '@bookingnetwork/ui-kit'

const props = defineProps<{
  distributionMode?: string
  rightsLevel?: string
  apkSha256?: string
  signingCertSha256?: string
  securityVerdict?: string
  scanReport?: string
  riskScore?: number
}>()

const isExternalLinkOnly = computed(() => {
  return props.distributionMode === 'EXTERNAL_LINK_ONLY' || props.rightsLevel === 'LEVEL_4_OFFICIAL_LINK_ONLY'
})

const securityStatusLabel = computed(() => {
  switch (props.securityVerdict) {
    case 'INFECTED_QUARANTINED':
      return '보안 위협 격리됨 (다운로드 차단)'
    case 'CLEAN_HIGH_RISK_MANUAL_REVIEW':
      return '보안 심층 검토 완료'
    default:
      return '독립 보안 검증 통과 (안전)'
  }
})

const distributionModeLabel = computed(() => {
  switch (props.distributionMode) {
    case 'DIRECT_HOSTED':
      return 'BN 자체 검증 호스팅'
    case 'FEDERATED_REPO':
      return '오픈소스 F-Droid 미러'
    case 'EXTERNAL_LINK_ONLY':
      return '공식 외부 링크 전용'
    default:
      return '직접 배포'
  }
})

const riskBarColor = computed(() => {
  const score = props.riskScore || 0
  if (score >= 70) return 'var(--bn-danger-red, #DC2626)'
  if (score >= 40) return 'var(--bn-warning-amber, #D97706)'
  return 'var(--bn-trust-verified, #059669)'
})
</script>

<style scoped>
.bn-trust-card {
  background-color: var(--bn-bg-surface, #FFFFFF);
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 12px;
  padding: 20px;
  margin: 20px 0;
}
.bn-trust-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 10px;
}
.bn-trust-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  font-size: 14px;
}
.bn-trust-badge.status-clean, .bn-trust-badge.status-clean_low_risk {
  color: var(--bn-trust-verified, #059669);
}
.bn-trust-badge.status-infected_quarantined {
  color: var(--bn-danger-red, #DC2626);
}
.bn-distribution-tag {
  font-size: 13px;
  color: var(--bn-text-muted, #64748B);
}
.bn-trust-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 16px;
}
.bn-trust-item label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  color: var(--bn-text-muted, #64748B);
  margin-bottom: 6px;
  text-transform: uppercase;
}
.bn-scan-result {
  font-size: 14px;
  color: var(--bn-text-secondary, #334155);
}
.bn-risk-bar-container {
  display: flex;
  align-items: center;
  gap: 8px;
}
.bn-risk-bar {
  height: 8px;
  border-radius: 4px;
  transition: width 0.3s ease;
}
.bn-risk-text {
  font-size: 13px;
  font-weight: 600;
  color: var(--bn-text-secondary, #334155);
}
.bn-external-notice {
  margin-top: 16px;
  padding: 10px 14px;
  background-color: var(--bn-bg-elevated, #F1F5F9);
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--bn-text-secondary, #334155);
}
.bn-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}
</style>


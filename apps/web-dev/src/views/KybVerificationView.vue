<template>
  <div class="bn-dev-page">
    <header class="bn-page-header">
      <h2>개발자 조직 신원 검증 (KYB Verification)</h2>
      <p class="bn-text-muted">프로덕션 앱 출시 및 상거래 정산을 위해 조직 법인 신원 검증이 필요합니다.</p>
    </header>

    <div class="bn-status-banner" :class="'status-' + status.toLowerCase()">
      <span class="bn-status-title">현재 상태: <strong>{{ status }}</strong></span>
      <p v-if="status === 'VERIFIED'" class="bn-status-desc">인증이 성공적으로 완료되었습니다. 프로덕션 릴리스가 가능합니다.</p>
      <p v-else-if="status === 'PENDING_REVIEW'" class="bn-status-desc">심사자가 제출된 서류를 검토 중입니다 (SLA: 24~48시간).</p>
      <p v-else class="bn-status-desc">서류 및 도메인 소유권 검증 정보를 제출해 주십시오.</p>
    </div>

    <form class="bn-form-card" @submit.prevent="submitVerification">
      <div class="bn-form-group">
        <label>법인 사업자등록번호 / 법인식별기호</label>
        <input v-model="form.businessNumber" type="text" placeholder="예: 123-45-67890" required class="bn-input" />
      </div>

      <div class="bn-form-group">
        <label>공식 웹사이트 도메인 (Reverse DNS 검증용)</label>
        <input v-model="form.domain" type="text" placeholder="예: acme.com" required class="bn-input" />
        <p class="bn-help-text">도메인 DNS에 TXT 레코드 <code>_bn-verify.{{ form.domain || 'example.com' }} = {{ verificationToken }}</code>를 추가해 주십시오.</p>
      </div>

      <div class="bn-form-group">
        <label>사업자등록증 / 법인인감증명서 (PDF, 암호화 격리 저장)</label>
        <input type="file" accept=".pdf,.png,.jpg" class="bn-file-input" @change="handleFileChange" />
      </div>

      <button type="submit" class="bn-btn bn-btn-primary" :disabled="submitting || status === 'VERIFIED'">
        {{ submitting ? '제출 중...' : '검증 요청 제출' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const status = ref('UNVERIFIED')
const verificationToken = ref('bn_txt_org_' + Math.random().toString(36).substring(2, 10))
const submitting = ref(false)

onMounted(() => {
  if (typeof window !== 'undefined') {
    const saved = localStorage.getItem('bn_kyb_status')
    if (saved) status.value = saved
    const savedToken = localStorage.getItem('bn_dns_token')
    if (savedToken) {
      verificationToken.value = savedToken
    } else {
      localStorage.setItem('bn_dns_token', verificationToken.value)
    }
  }
})

const form = ref({
  businessNumber: '',
  domain: '',
  documentFile: null as File | null
})

function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  if (target.files && target.files[0]) {
    form.value.documentFile = target.files[0]
  }
}

async function submitVerification() {
  submitting.value = true
  try {
    await new Promise(r => setTimeout(r, 600))
    status.value = 'PENDING_REVIEW'
    if (typeof window !== 'undefined') {
      localStorage.setItem('bn_kyb_status', 'PENDING_REVIEW')
    }
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.bn-dev-page {
  padding: 24px;
  max-width: 800px;
}
.bn-page-header {
  margin-bottom: 24px;
}
.bn-status-banner {
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 24px;
}
.bn-status-banner.status-verified {
  background-color: #ECFDF5;
  border: 1px solid #10B981;
  color: #065F46;
}
.bn-status-banner.status-pending_review {
  background-color: #FEF3C7;
  border: 1px solid #F59E0B;
  color: #92400E;
}
.bn-status-banner.status-unverified {
  background-color: var(--bn-bg-elevated, #F1F5F9);
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  color: var(--bn-text-primary, #0F172A);
}
.bn-form-card {
  background-color: var(--bn-bg-surface, #FFFFFF);
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 12px;
  padding: 24px;
}
.bn-form-group {
  margin-bottom: 20px;
}
.bn-form-group label {
  display: block;
  font-weight: 600;
  margin-bottom: 6px;
  font-size: 14px;
}
.bn-input {
  width: 100%;
  padding: 10px 14px;
  border-radius: 6px;
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  font-size: 14px;
}
.bn-help-text {
  font-size: 12px;
  color: var(--bn-text-muted, #64748B);
  margin-top: 6px;
}
.bn-btn-primary {
  background-color: var(--bn-primary, #2563EB);
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}
.bn-btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>


<template>
  <div class="bn-dev-page">
    <header class="bn-page-header">
      <h2>개발자 조직 신원 검증 (KYB Verification)</h2>
      <p class="bn-text-muted">프로덕션 앱 출시 및 상거래 정산을 위해 조직 법인 신원 검증이 필요합니다.</p>
    </header>

    <div v-if="isLoading" class="bn-loading">불러오는 중…</div>
    <div v-if="loadError" class="bn-error">{{ loadError }}</div>

    <div v-if="!isLoading && !loadError">
      <div v-if="orgs.length > 0" class="bn-org-row">
        <label for="orgSelect">조직</label>
        <select id="orgSelect" v-model="activeOrgId" class="bn-input" @change="loadVerification">
          <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.org_name }} ({{ org.org_slug }})</option>
        </select>
      </div>

      <div class="bn-status-banner" :class="'status-' + status.toLowerCase()">
        <span class="bn-status-title">현재 상태: <strong>{{ status }}</strong></span>
        <p v-if="status === 'VERIFIED'" class="bn-status-desc">인증이 성공적으로 완료되었습니다. 프로덕션 릴리스가 가능합니다.</p>
        <p v-else-if="status === 'PENDING_REVIEW'" class="bn-status-desc">심사자가 제출된 서류를 검토 중입니다 (SLA: 24~48시간).</p>
        <p v-else-if="status === 'NO_ORG'" class="bn-status-desc">소속 조직이 없습니다. 아래에서 조직을 먼저 생성해 주십시오.</p>
        <p v-else class="bn-status-desc">서류 및 도메인 소유권 검증 정보를 제출해 주십시오.</p>
      </div>

      <div v-if="verification" class="bn-form-card bn-verify-detail">
        <h3>제출된 검증 정보</h3>
        <dl>
          <dt>문서 종류</dt><dd>{{ verification.doc_type }}</dd>
          <dt>문서 URL</dt><dd class="bn-mono">{{ verification.doc_url }}</dd>
          <dt>도메인</dt><dd>{{ verification.domain }}</dd>
          <dt>제출 시각</dt><dd>{{ verification.submitted_at }}</dd>
        </dl>
      </div>

      <form v-if="orgs.length === 0" class="bn-form-card" @submit.prevent="createOrg">
        <h3>조직 생성</h3>
        <div class="bn-form-group">
          <label>조직 슬러그 (고유 ID)</label>
          <input v-model="newOrg.slug" type="text" placeholder="예: acme" required class="bn-input" />
        </div>
        <div class="bn-form-group">
          <label>조직명</label>
          <input v-model="newOrg.name" type="text" placeholder="예: Acme Inc." required class="bn-input" />
        </div>
        <div v-if="createError" class="bn-error">{{ createError }}</div>
        <button type="submit" class="bn-btn bn-btn-primary" :disabled="creating">
          {{ creating ? '생성 중...' : '조직 생성' }}
        </button>
      </form>

      <form v-if="orgs.length > 0" class="bn-form-card" @submit.prevent="submitVerification">
        <div class="bn-form-group">
          <label>문서 종류</label>
          <select v-model="form.docType" class="bn-input">
            <option value="BUSINESS_REGISTRATION">사업자등록증</option>
            <option value="CERTIFICATE_OF_INCORPORATION">법인등기부등본</option>
            <option value="DUNS">D-U-N-S 번호 확인서</option>
          </select>
        </div>

        <div class="bn-form-group">
          <label>증빙 문서 URL</label>
          <input v-model="form.docUrl" type="url" placeholder="https://..." required class="bn-input" />
        </div>

        <div class="bn-form-group">
          <label>공식 웹사이트 도메인</label>
          <input v-model="form.domain" type="text" placeholder="예: acme.com" required class="bn-input" />
        </div>

        <div v-if="submitError" class="bn-error">{{ submitError }}</div>
        <div v-if="submitDone" class="bn-success">검증 요청이 접수되었습니다.</div>
        <button type="submit" class="bn-btn bn-btn-primary" :disabled="submitting || status === 'VERIFIED'">
          {{ submitting ? '제출 중...' : '검증 요청 제출' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getSharedClient, ApiClientError, type OrgSummary } from '@bookingnetwork/api-client'

const api = getSharedClient()

interface Verification {
  id: string
  org_id: string
  doc_type: string
  doc_url: string
  domain: string
  status: string
  submitted_at: string
}

const orgs = ref<OrgSummary[]>([])
const activeOrgId = ref('')
const status = ref('UNKNOWN')
const verification = ref<Verification | null>(null)
const submitting = ref(false)
const isLoading = ref(true)
const loadError = ref('')
const submitError = ref('')
const submitDone = ref(false)

const newOrg = ref({ slug: '', name: '' })
const creating = ref(false)
const createError = ref('')

const form = ref({
  docType: 'BUSINESS_REGISTRATION',
  docUrl: '',
  domain: ''
})

onMounted(async () => {
  await loadOrgs()
})

async function loadOrgs() {
  isLoading.value = true
  loadError.value = ''
  try {
    orgs.value = await api.listOrgs()
    if (orgs.value.length > 0) {
      activeOrgId.value = orgs.value[0].id
      await loadVerification()
    } else {
      status.value = 'NO_ORG'
    }
  } catch (err) {
    status.value = 'LOAD_FAILED'
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load organizations'
  } finally {
    isLoading.value = false
  }
}

async function loadVerification() {
  verification.value = null
  submitDone.value = false
  const org = orgs.value.find(o => o.id === activeOrgId.value)
  status.value = org?.verification_status || 'UNVERIFIED'
  if (!activeOrgId.value) return
  try {
    verification.value = await api.getVerification(activeOrgId.value) as Verification
    if (verification.value?.status) {
      status.value = verification.value.status
    }
  } catch (err) {
    if (err instanceof ApiClientError && err.status === 404) {
      return
    }
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load verification'
  }
}

async function createOrg() {
  if (!newOrg.value.slug.trim() || !newOrg.value.name.trim()) return
  creating.value = true
  createError.value = ''
  try {
    const org = await api.createOrganization(newOrg.value.slug.trim(), newOrg.value.name.trim())
    newOrg.value = { slug: '', name: '' }
    await loadOrgs()
    activeOrgId.value = org.id
    await loadVerification()
  } catch (err) {
    createError.value = err instanceof ApiClientError ? err.message : 'Failed to create organization'
  } finally {
    creating.value = false
  }
}

async function submitVerification() {
  if (!activeOrgId.value) return
  submitting.value = true
  submitError.value = ''
  try {
    verification.value = await api.submitVerification(activeOrgId.value, {
      doc_type: form.value.docType,
      doc_url: form.value.docUrl.trim(),
      domain: form.value.domain.trim()
    }) as Verification
    status.value = verification.value?.status || 'PENDING_REVIEW'
    submitDone.value = true
    await loadOrgs()
    activeOrgId.value = verification.value?.org_id || activeOrgId.value
  } catch (err) {
    submitError.value = err instanceof ApiClientError ? err.message : 'Failed to submit verification'
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.bn-dev-page {
  max-width: 820px;
  margin: 0 auto;
}
.bn-page-header {
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--bn-ink);
}
.bn-page-header h2 {
  font-family: var(--bn-font-display);
  font-size: 24px;
  font-weight: 900;
  letter-spacing: -0.02em;
  margin: 0 0 4px;
}
.bn-text-muted {
  color: var(--bn-text-secondary);
  font-size: 13px;
  margin: 0;
}
.bn-status-banner {
  padding: 16px 18px;
  border-radius: var(--bn-radius-md);
  margin-bottom: 24px;
  border: 1px solid var(--bn-border-subtle);
  background-color: var(--bn-bg-surface);
}
.bn-status-title {
  font-size: 13px;
  color: var(--bn-text-secondary);
}
.bn-status-title strong {
  font-family: var(--bn-font-mono);
  font-size: 13px;
  letter-spacing: 0.04em;
}
.bn-status-desc {
  font-size: 13px;
  color: var(--bn-text-secondary);
  margin: 6px 0 0;
  line-height: 1.6;
}
.bn-status-banner.status-verified {
  background-color: var(--bn-trust-wash);
  border-color: var(--bn-trust-verified);
}
.bn-status-banner.status-verified .bn-status-title strong {
  color: var(--bn-trust-verified);
}
.bn-status-banner.status-pending_review {
  background-color: var(--bn-warning-wash);
  border-color: var(--bn-warning-amber);
}
.bn-status-banner.status-pending_review .bn-status-title strong {
  color: var(--bn-warning-amber);
}
.bn-form-card {
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
  margin-bottom: 24px;
}
.bn-form-card h3 {
  font-family: var(--bn-font-display);
  font-size: 15px;
  font-weight: 800;
  margin: 0 0 16px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--bn-border-subtle);
}
.bn-form-group {
  margin-bottom: 18px;
}
.bn-form-group label {
  display: block;
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  margin-bottom: 8px;
}
.bn-input {
  width: 100%;
  box-sizing: border-box;
  padding: 10px 12px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-subtle);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
  font-size: 14px;
}
.bn-input:focus {
  outline: none;
  border-color: var(--bn-ink);
}
.bn-help-text {
  font-size: 12px;
  color: var(--bn-text-muted);
  margin-top: 6px;
}
.bn-btn {
  font-family: var(--bn-font-sans);
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
}
.bn-btn-primary {
  background-color: var(--bn-action-bg);
  color: var(--bn-action-text);
  border: none;
  padding: 11px 24px;
  border-radius: var(--bn-radius-sm);
}
.bn-btn-primary:hover:not(:disabled) {
  background-color: var(--bn-action-bg-hover);
}
.bn-btn-primary:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.bn-org-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  font-size: 13px;
  color: var(--bn-text-secondary);
}
.bn-org-row label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  flex-shrink: 0;
}
.bn-org-row .bn-input {
  width: auto;
  flex: 1;
}
.bn-loading {
  padding: 24px;
  color: var(--bn-text-muted);
  font-size: 14px;
  text-align: center;
}
.bn-error {
  padding: 12px 16px;
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  color: var(--bn-danger-red);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  margin-bottom: 16px;
}
.bn-success {
  padding: 12px 16px;
  background: var(--bn-trust-wash);
  border: 1px solid var(--bn-trust-verified);
  color: var(--bn-trust-verified);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  margin-bottom: 16px;
}
.bn-verify-detail dl {
  display: grid;
  grid-template-columns: 130px 1fr;
  gap: 10px 16px;
  font-size: 14px;
  margin: 4px 0 0;
}
.bn-verify-detail dt {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-text-muted);
  padding-top: 2px;
}
.bn-verify-detail dd {
  margin: 0;
}
.bn-mono {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  word-break: break-all;
}
</style>


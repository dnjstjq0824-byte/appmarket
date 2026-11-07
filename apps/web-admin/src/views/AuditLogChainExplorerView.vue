<template>
  <div class="bn-admin-page">
    <header class="bn-page-header">
      <div class="bn-header-flex">
        <div>
          <h2>불변 해시 체인 감사 로그 탐색기 (Audit Log Chain Explorer)</h2>
          <p class="bn-text-muted">모든 민감 관리자 변경 및 권한 집행 이력이 SHA-256 암호학적 해시 체인으로 보호됩니다.</p>
        </div>
        <button class="bn-btn bn-btn-verify" :disabled="verifying" @click="verifyChain">
          {{ verifying ? '체인 검증 중...' : '체인 전체 무결성 검증' }}
        </button>
      </div>
    </header>

    <!-- Integrity Alert Banner -->
    <div v-if="integrityStatus === 'TAMPERED'" class="bn-alert bn-alert-danger">
      <svg class="bn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <polygon points="7.86 2 16.14 2 22 7.86 22 16.14 16.14 22 7.86 22 2 16.14 2 7.86 7.86 2" />
        <line x1="12" y1="8" x2="12" y2="12" />
        <line x1="12" y1="16" x2="12.01" y2="16" />
      </svg>
      <strong>보안 경보: 감사 로그 체인 변조가 감지되었습니다! (인덱스 #{{ tamperedIndex }}) 즉시 보안팀에 보고하십시오.</strong>
    </div>

    <div v-else-if="integrityStatus === 'VERIFIED_OK'" class="bn-alert bn-alert-success">
      <svg class="bn-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
        <polyline points="22 4 12 14.01 9 11.01" />
      </svg>
      <strong>감사 체인 검증 완료: 전체 {{ auditLogs.length }}개 레코드의 암호학적 링크가 100% 무결합니다.</strong>
    </div>

    <div v-if="serverChainValid === false" class="bn-alert bn-alert-danger">
      <strong>서버 체인 파손 보고 (Block #{{ serverChainFailedAt }}). 즉시 보안팀에 보고하고 시스템을 동결하십시오.</strong>
    </div>

    <div v-if="isLoading" class="bn-loading">불러오는 중…</div>
    <div v-if="loadError" class="bn-error">{{ loadError }}</div>
    <div v-if="!isLoading && !loadError && auditLogs.length === 0" class="bn-empty">감사 로그가 없습니다.</div>

    <div v-if="!isLoading && !loadError" class="bn-form-card">
      <table class="bn-table">
        <thead>
          <tr>
            <th>시간</th>
            <th>행위자 (Actor)</th>
            <th>작업 (Action)</th>
            <th>대상</th>
            <th>이전 해시 (Prev Hash)</th>
            <th>레코드 해시 (Record Hash)</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in auditLogs" :key="log.id">
            <td class="bn-text-muted">{{ log.created_at }}</td>
            <td><strong>{{ log.actor_id }}</strong></td>
            <td><span class="bn-action-pill">{{ log.action }}</span></td>
            <td>{{ log.target_type }}: {{ log.target_id }}</td>
            <td><BnHashFingerprintBadge :hash="log.prev_hash" /></td>
            <td><BnHashFingerprintBadge :hash="log.record_hash" /></td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { BnHashFingerprintBadge } from '@bookingnetwork/ui-kit'
import { getSharedClient, ApiClientError } from '@bookingnetwork/api-client'

const api = getSharedClient()

interface AuditRecord {
  id: string
  seq?: number
  actor_id: string
  action: string
  target_type: string
  target_id: string
  diff_json?: string
  ip_address?: string
  prev_hash: string
  record_hash: string
  created_at: string
  created_at_nano?: string
}

const verifying = ref(false)
const integrityStatus = ref<'UNCHECKED' | 'VERIFIED_OK' | 'TAMPERED'>('UNCHECKED')
const tamperedIndex = ref<number | null>(null)
const serverChainValid = ref<boolean | null>(null)
const serverChainFailedAt = ref(-1)

const auditLogs = ref<AuditRecord[]>([])
const isLoading = ref(true)
const loadError = ref('')

onMounted(async () => {
  await loadLogs()
  isLoading.value = false
})

async function loadLogs() {
  loadError.value = ''
  try {
    const res = await api.listAuditLogs(200)
    serverChainValid.value = res.chainValid
    serverChainFailedAt.value = res.chainFailedAt ?? -1
    auditLogs.value = res.records || []
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load audit logs'
  }
}

async function sha256Hex(text: string): Promise<string> {
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(text))
  return [...new Uint8Array(digest)].map(b => b.toString(16).padStart(2, '0')).join('')
}

async function verifyChain() {
  verifying.value = true
  integrityStatus.value = 'UNCHECKED'
  tamperedIndex.value = null
  try {
    await loadLogs()
    let prev = ''
    for (let i = 0; i < auditLogs.value.length; i++) {
      const log = auditLogs.value[i]
      if (i > 0 && log.prev_hash !== prev) {
        integrityStatus.value = 'TAMPERED'
        tamperedIndex.value = i
        return
      }
      if (log.prev_hash && log.record_hash && log.created_at_nano) {
        const payload = [log.prev_hash, log.actor_id, log.action, log.target_type, log.target_id, log.diff_json || '', String(log.created_at_nano)].join('|')
        const recomputed = await sha256Hex(payload)
        if (recomputed !== log.record_hash.toLowerCase()) {
          integrityStatus.value = 'TAMPERED'
          tamperedIndex.value = i
          return
        }
      }
      prev = log.record_hash
    }
    if (serverChainValid.value === false) {
      integrityStatus.value = 'TAMPERED'
      tamperedIndex.value = serverChainFailedAt.value
      return
    }
    integrityStatus.value = 'VERIFIED_OK'
  } finally {
    verifying.value = false
  }
}
</script>

<style scoped>
.bn-admin-page {
  max-width: 1180px;
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
td.bn-text-muted {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  color: var(--bn-text-muted);
}
.bn-header-flex {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}
.bn-btn {
  font-family: var(--bn-font-sans);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
}
.bn-btn-verify {
  background-color: var(--bn-action-bg);
  color: var(--bn-action-text);
  border: none;
  padding: 10px 20px;
  border-radius: var(--bn-radius-sm);
}
.bn-btn-verify:hover:not(:disabled) {
  background-color: var(--bn-action-bg-hover);
}
.bn-btn-verify:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.bn-alert {
  padding: 14px 18px;
  border-radius: var(--bn-radius-md);
  margin: 20px 0;
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  line-height: 1.6;
}
.bn-alert-success {
  background-color: var(--bn-trust-wash);
  border: 1px solid var(--bn-trust-verified);
  color: var(--bn-trust-verified);
}
.bn-alert-danger {
  background-color: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  color: var(--bn-danger-red);
}
.bn-loading {
  padding: 24px;
  text-align: center;
  color: var(--bn-text-muted);
  font-size: 14px;
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
.bn-empty {
  padding: 28px 24px;
  text-align: center;
  color: var(--bn-text-muted);
  font-size: 14px;
  border: 1px dashed var(--bn-border-strong);
  border-radius: var(--bn-radius-md);
  margin-bottom: 16px;
}
.bn-form-card {
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
  overflow-x: auto;
}
.bn-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 760px;
}
.bn-table th, .bn-table td {
  padding: 12px 14px;
  text-align: left;
  border-bottom: 1px solid var(--bn-border-subtle);
  font-size: 13px;
  vertical-align: middle;
}
.bn-table tr:last-child td {
  border-bottom: none;
}
.bn-table th {
  font-family: var(--bn-font-mono);
  font-weight: 400;
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  white-space: nowrap;
}
.bn-action-pill {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  padding: 3px 9px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-primary);
  white-space: nowrap;
}
.bn-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}
</style>


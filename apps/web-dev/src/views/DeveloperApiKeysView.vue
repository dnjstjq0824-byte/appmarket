<template>
  <div class="bn-dev-page">
    <header class="bn-page-header">
      <h2>CI/CD 배포 자동화 API 키 관리</h2>
      <p class="bn-text-muted">GitHub Actions, GitLab CI 등에서 APK 업로드 및 릴리스 배포 자동화에 사용할 키를 발급합니다.</p>
    </header>

    <div v-if="orgs.length > 0" class="bn-org-row">
      <label for="orgSelect">조직</label>
      <select id="orgSelect" v-model="activeOrgId" class="bn-input" @change="loadKeys">
        <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.org_name }}</option>
      </select>
    </div>
    <div v-if="isLoading" class="bn-loading">불러오는 중…</div>
    <div v-if="loadError" class="bn-error">{{ loadError }}</div>
    <div v-if="actionError" class="bn-error">{{ actionError }}</div>

    <div v-if="!isLoading && !loadError" class="bn-form-card">
      <div v-if="newlyIssuedKey" class="bn-key-alert">
        <strong>새로 발급된 API Key (보안상 지금만 표시됩니다):</strong>
        <code class="bn-key-code">{{ newlyIssuedKey }}</code>
        <p class="bn-key-hint">이 키를 안전한 장소에 복사하십시오. 창을 닫으면 다시 조회할 수 없습니다.</p>
      </div>

      <div class="bn-create-bar">
        <input v-model="keyName" type="text" placeholder="키 이름 (예: GitHub Actions Production CI)" class="bn-input" />
        <input v-model="keyScopes" type="text" placeholder="스코프 (쉼표 구분)" class="bn-input bn-input-sm" />
        <input v-model.number="keyTtlDays" type="number" min="1" max="365" title="유효일" class="bn-input bn-input-xs" />
        <button class="bn-btn bn-btn-primary" :disabled="isActing" @click="generateKey">새 API 키 발급</button>
      </div>

      <div v-if="apiKeys.length === 0" class="bn-empty">발급된 API 키가 없습니다.</div>
      <table v-else class="bn-table">
        <thead>
          <tr>
            <th>이름</th>
            <th>스코프</th>
            <th>만료일</th>
            <th>생성일</th>
            <th>상태</th>
            <th>작업</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="k in apiKeys" :key="k.id">
            <td><strong>{{ k.name }}</strong></td>
            <td><code>{{ (k.scopes || []).join(', ') }}</code></td>
            <td class="bn-text-muted">{{ k.expires_at }}</td>
            <td class="bn-text-muted">{{ k.created_at }}</td>
            <td>
              <span class="bn-status-badge" :class="!k.revoked ? 'active' : 'revoked'">
                {{ !k.revoked ? '활성 (Active)' : '폐기됨 (Revoked)' }}
              </span>
            </td>
            <td>
              <button v-if="!k.revoked" class="bn-btn-text-danger" :disabled="isActing" @click="revokeKey(k.id)">즉시 폐기</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getSharedClient, ApiClientError, type OrgSummary } from '@bookingnetwork/api-client'

const api = getSharedClient()

interface ApiKeyRecord {
  id: string
  org_id: string
  name: string
  scopes: string[]
  expires_at: string
  revoked: boolean
  created_at: string
}

const orgs = ref<OrgSummary[]>([])
const activeOrgId = ref('')
const apiKeys = ref<ApiKeyRecord[]>([])
const newlyIssuedKey = ref<string | null>(null)
const keyName = ref('')
const keyScopes = ref('release:upload')
const keyTtlDays = ref(90)
const isLoading = ref(true)
const loadError = ref('')
const actionError = ref('')
const isActing = ref(false)

onMounted(async () => {
  try {
    orgs.value = await api.listOrgs()
    if (orgs.value.length > 0) {
      activeOrgId.value = orgs.value[0].id
      await loadKeys()
    }
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load organizations'
  } finally {
    isLoading.value = false
  }
})

async function loadKeys() {
  if (!activeOrgId.value) return
  isLoading.value = true
  loadError.value = ''
  try {
    apiKeys.value = await api.listApiKeys(activeOrgId.value) as ApiKeyRecord[]
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load API keys'
  } finally {
    isLoading.value = false
  }
}

async function generateKey() {
  if (!keyName.value.trim() || !activeOrgId.value) return
  isActing.value = true
  actionError.value = ''
  newlyIssuedKey.value = null
  try {
    const scopes = keyScopes.value.split(',').map(s => s.trim()).filter(Boolean)
    const res = await api.createApiKey(activeOrgId.value, {
      name: keyName.value.trim(),
      scopes,
      ttl_days: keyTtlDays.value
    }) as { api_key: string }
    newlyIssuedKey.value = res.api_key
    keyName.value = ''
    await loadKeys()
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to issue API key'
  } finally {
    isActing.value = false
  }
}

async function revokeKey(id: string) {
  if (!confirm('Revoke this API key immediately? Active CI jobs using it will fail.')) return
  isActing.value = true
  actionError.value = ''
  try {
    await api.revokeApiKey(activeOrgId.value, id)
    await loadKeys()
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to revoke API key'
  } finally {
    isActing.value = false
  }
}
</script>

<style scoped>
.bn-dev-page {
  max-width: 960px;
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
.bn-form-card {
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
}
.bn-key-alert {
  background-color: var(--bn-warning-wash);
  border: 1px solid var(--bn-warning-amber);
  padding: 16px;
  border-radius: var(--bn-radius-sm);
  margin-bottom: 24px;
  font-size: 13px;
}
.bn-key-code {
  display: block;
  margin-top: 10px;
  font-family: var(--bn-font-mono);
  font-size: 13px;
  background-color: var(--bn-bg-surface);
  padding: 10px 12px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-subtle);
  color: var(--bn-text-primary);
  word-break: break-all;
}
.bn-key-hint {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--bn-text-secondary);
}
.bn-create-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}
.bn-input {
  flex: 1;
  min-width: 200px;
  padding: 10px 12px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-subtle);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
  font-size: 13px;
  box-sizing: border-box;
}
.bn-input:focus {
  outline: none;
  border-color: var(--bn-ink);
}
.bn-input-sm {
  flex: 0 1 220px;
  min-width: 160px;
}
.bn-input-xs {
  flex: 0 1 90px;
  min-width: 70px;
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
.bn-empty {
  padding: 28px 24px;
  text-align: center;
  color: var(--bn-text-muted);
  font-size: 14px;
  border: 1px dashed var(--bn-border-strong);
  border-radius: var(--bn-radius-md);
}
.bn-btn {
  font-family: var(--bn-font-sans);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
}
.bn-btn-primary {
  background-color: var(--bn-action-bg);
  color: var(--bn-action-text);
  border: none;
  padding: 10px 20px;
  border-radius: var(--bn-radius-sm);
}
.bn-btn-primary:hover:not(:disabled) {
  background-color: var(--bn-action-bg-hover);
}
.bn-btn-primary:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.bn-table {
  width: 100%;
  border-collapse: collapse;
}
.bn-table th, .bn-table td {
  padding: 12px;
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
}
.bn-table td code {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  color: var(--bn-text-secondary);
  word-break: break-all;
}
.bn-status-badge {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.04em;
  padding: 3px 9px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-secondary);
  white-space: nowrap;
}
.bn-status-badge.active {
  border-color: var(--bn-trust-verified);
  color: var(--bn-trust-verified);
}
.bn-status-badge.revoked {
  border-color: var(--bn-danger-red);
  color: var(--bn-danger-red);
}
.bn-btn-text-danger {
  color: var(--bn-danger-red);
  background: none;
  border: none;
  cursor: pointer;
  font-size: 13px;
  font-weight: 700;
  padding: 4px 0;
}
.bn-btn-text-danger:hover:not(:disabled) {
  text-decoration: underline;
  text-underline-offset: 3px;
}
.bn-btn-text-danger:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
</style>


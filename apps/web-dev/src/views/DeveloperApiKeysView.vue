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
        <p class="bn-text-muted" style="margin: 6px 0 0 0; font-size: 12px;">이 키를 안전한 장소에 복사하십시오. 창을 닫으면 다시 조회할 수 없습니다.</p>
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
  padding: 24px;
  max-width: 900px;
}
.bn-form-card {
  background-color: var(--bn-bg-surface, #FFFFFF);
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 12px;
  padding: 24px;
}
.bn-key-alert {
  background-color: #FEF3C7;
  border: 1px solid #F59E0B;
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 24px;
}
.bn-key-code {
  display: block;
  margin-top: 8px;
  font-family: "JetBrains Mono", monospace;
  font-size: 14px;
  background-color: #FFFFFF;
  padding: 8px 12px;
  border-radius: 4px;
  border: 1px solid #D97706;
  color: #92400E;
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
  padding: 10px 14px;
  border-radius: 6px;
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
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
}
.bn-loading {
  padding: 16px;
  color: var(--bn-text-muted, #64748B);
  font-size: 14px;
}
.bn-error {
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #991b1b;
  border-radius: 6px;
  font-size: 13px;
  margin-bottom: 16px;
}
.bn-empty {
  padding: 24px;
  text-align: center;
  color: var(--bn-text-muted, #64748B);
  font-size: 14px;
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
.bn-table {
  width: 100%;
  border-collapse: collapse;
}
.bn-table th, .bn-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid var(--bn-border-subtle, #E2E8F0);
  font-size: 14px;
}
.bn-status-badge {
  font-size: 12px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 4px;
}
.bn-status-badge.active {
  background-color: #ECFDF5;
  color: #065F46;
}
.bn-status-badge.revoked {
  background-color: #FEE2E2;
  color: #991B1B;
}
.bn-btn-text-danger {
  color: var(--bn-danger-red, #DC2626);
  background: none;
  border: none;
  cursor: pointer;
  font-weight: 600;
}
</style>


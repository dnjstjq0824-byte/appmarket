<template>
  <div class="bn-dev-page">
    <header class="bn-page-header">
      <h2>조직 팀원 및 역할 관리 (Team Members RBAC)</h2>
      <p class="bn-text-muted">조직에 팀원을 초대하고 직무 분리 원칙에 따른 권한을 할당합니다.</p>
    </header>

    <div v-if="orgs.length > 0" class="bn-org-row">
      <label for="orgSelect">조직</label>
      <select id="orgSelect" v-model="activeOrgId" class="bn-select" @change="loadMembers">
        <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.org_name }}</option>
      </select>
    </div>
    <div v-if="isLoading" class="bn-loading">불러오는 중…</div>
    <div v-if="loadError" class="bn-error">{{ loadError }}</div>
    <div v-if="actionError" class="bn-error">{{ actionError }}</div>

    <div v-if="!isLoading && !loadError" class="bn-form-card">
      <div class="bn-invite-bar">
        <input v-model="newUserId" type="text" placeholder="팀원 사용자 ID..." class="bn-input" />
        <select v-model="newRole" class="bn-select">
          <option value="DEV_RELEASE_MGR">릴리스 관리자 (DEV_RELEASE_MGR)</option>
          <option value="DEV_EDITOR">콘텐츠 편집자 (DEV_EDITOR)</option>
          <option value="DEV_FINANCE">재무/정산 담당 (DEV_FINANCE)</option>
          <option value="DEV_SUPPORT">고객 지원 (DEV_SUPPORT)</option>
        </select>
        <button class="bn-btn bn-btn-primary" :disabled="isActing" @click="inviteMember">초대</button>
      </div>

      <div v-if="members.length === 0" class="bn-empty">등록된 팀원이 없습니다. 위에서 초대해 주십시오.</div>
      <table v-else class="bn-table">
        <thead>
          <tr>
            <th>팀원 사용자 ID</th>
            <th>역할</th>
            <th>권한 요약</th>
            <th>관리</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in members" :key="m.user_id">
            <td><strong>{{ m.user_id }}</strong></td>
            <td><span class="bn-role-tag">{{ m.role }}</span></td>
            <td class="bn-text-muted">{{ roleDescription(m.role) }}</td>
            <td>
              <button
                class="bn-btn-text-danger"
                :disabled="isActing"
                @click="removeMember(m.user_id)"
              >
                삭제
              </button>
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

interface Member {
  org_id: string
  user_id: string
  role: string
}

const ROLE_DESCRIPTIONS: Record<string, string> = {
  DEV_OWNER: '조직 전권 및 소유권 관리',
  DEV_RELEASE_MGR: 'APK 업로드 및 출시 심사 제출',
  DEV_EDITOR: '스토어 리스팅 콘텐츠 편집',
  DEV_FINANCE: '매출 통계 및 정산 계좌 관리',
  DEV_SUPPORT: '리뷰 답글 및 고객 지원'
}

const orgs = ref<OrgSummary[]>([])
const activeOrgId = ref('')
const members = ref<Member[]>([])
const newUserId = ref('')
const newRole = ref('DEV_RELEASE_MGR')
const isLoading = ref(true)
const loadError = ref('')
const actionError = ref('')
const isActing = ref(false)

onMounted(async () => {
  try {
    orgs.value = await api.listOrgs()
    if (orgs.value.length > 0) {
      activeOrgId.value = orgs.value[0].id
      await loadMembers()
    }
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load organizations'
  } finally {
    isLoading.value = false
  }
})

async function loadMembers() {
  if (!activeOrgId.value) return
  isLoading.value = true
  loadError.value = ''
  try {
    members.value = await api.listMembers(activeOrgId.value) as Member[]
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load members'
  } finally {
    isLoading.value = false
  }
}

function roleDescription(role: string): string {
  return ROLE_DESCRIPTIONS[role] || '할당된 조직 역할'
}

async function inviteMember() {
  if (!newUserId.value.trim() || !activeOrgId.value) return
  isActing.value = true
  actionError.value = ''
  try {
    await api.addMember(activeOrgId.value, newUserId.value.trim(), newRole.value)
    newUserId.value = ''
    await loadMembers()
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to invite member'
  } finally {
    isActing.value = false
  }
}

async function removeMember(userId: string) {
  if (!confirm(`Remove member ${userId} from this organization?`)) return
  isActing.value = true
  actionError.value = ''
  try {
    await api.removeMember(activeOrgId.value, userId)
    await loadMembers()
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to remove member'
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
.bn-form-card {
  background-color: var(--bn-bg-surface, #FFFFFF);
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 12px;
  padding: 24px;
}
.bn-invite-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}
.bn-input {
  flex: 1;
  min-width: 240px;
  padding: 10px 14px;
  border-radius: 6px;
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
}
.bn-select {
  padding: 10px 14px;
  border-radius: 6px;
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
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
.bn-role-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  background-color: var(--bn-bg-elevated, #F1F5F9);
  color: var(--bn-primary, #2563EB);
}
.bn-btn-text-danger {
  color: var(--bn-danger-red, #DC2626);
  background: none;
  border: none;
  cursor: pointer;
  font-weight: 600;
}
</style>


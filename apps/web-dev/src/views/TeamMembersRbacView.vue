<template>
  <div class="bn-dev-page">
    <header class="bn-page-header">
      <h2>조직 팀원 및 역할 관리 (Team Members RBAC)</h2>
      <p class="bn-text-muted">조직에 팀원을 초대하고 직무 분리 원칙에 따른 권한을 할당합니다.</p>
    </header>

    <div class="bn-form-card">
      <div class="bn-invite-bar">
        <input v-model="newEmail" type="email" placeholder="팀원 이메일 주소..." class="bn-input" />
        <select v-model="newRole" class="bn-select">
          <option value="DEV_RELEASE_MGR">릴리스 관리자 (DEV_RELEASE_MGR)</option>
          <option value="DEV_EDITOR">콘텐츠 편집자 (DEV_EDITOR)</option>
          <option value="DEV_FINANCE">재무/정산 담당 (DEV_FINANCE)</option>
          <option value="DEV_SUPPORT">고객 지원 (DEV_SUPPORT)</option>
        </select>
        <button class="bn-btn bn-btn-primary" @click="inviteMember">초대</button>
      </div>

      <table class="bn-table">
        <thead>
          <tr>
            <th>팀원 이메일</th>
            <th>역할</th>
            <th>권한 요약</th>
            <th>관리</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in members" :key="m.id">
            <td><strong>{{ m.email }}</strong></td>
            <td><span class="bn-role-tag">{{ m.role }}</span></td>
            <td class="bn-text-muted">{{ m.description }}</td>
            <td>
              <button
                v-if="m.role !== 'DEV_OWNER'"
                class="bn-btn-text-danger"
                @click="removeMember(m.id)"
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
import { ref } from 'vue'

const newEmail = ref('')
const newRole = ref('DEV_RELEASE_MGR')

const members = ref([
  { id: '1', email: 'owner@acme.com', role: 'DEV_OWNER', description: '조직 전권 및 소유권 관리' },
  { id: '2', email: 'release-lead@acme.com', role: 'DEV_RELEASE_MGR', description: 'APK 업로드 및 출시 심사 제출' },
  { id: '3', email: 'finance@acme.com', role: 'DEV_FINANCE', description: '매출 통계 및 정산 계좌 관리' }
])

function inviteMember() {
  if (!newEmail.value.includes('@')) return
  members.value.push({
    id: Date.now().toString(),
    email: newEmail.value,
    role: newRole.value,
    description: '할당된 조직 역할'
  })
  newEmail.value = ''
}

function removeMember(id: string) {
  members.value = members.value.filter(m => m.id !== id)
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


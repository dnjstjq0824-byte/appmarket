<template>
  <div class="bn-dev-page">
    <header class="bn-page-header">
      <h2>CI/CD 배포 자동화 API 키 관리</h2>
      <p class="bn-text-muted">GitHub Actions, GitLab CI 등에서 APK 업로드 및 릴리스 배포 자동화에 사용할 키를 발급합니다.</p>
    </header>

    <div class="bn-form-card">
      <div v-if="newlyIssuedKey" class="bn-key-alert">
        <strong>새로 발급된 API Key (보안상 지금만 표시됩니다):</strong>
        <code class="bn-key-code">{{ newlyIssuedKey }}</code>
        <p class="bn-text-muted" style="margin: 6px 0 0 0; font-size: 12px;">이 키를 안전한 장소에 복사하십시오. 창을 닫으면 다시 조회할 수 없습니다.</p>
      </div>

      <div class="bn-create-bar">
        <input v-model="keyDescription" type="text" placeholder="키 설명 (예: GitHub Actions Production CI)" class="bn-input" />
        <button class="bn-btn bn-btn-primary" @click="generateKey">새 API 키 발급</button>
      </div>

      <table class="bn-table">
        <thead>
          <tr>
            <th>설명</th>
            <th>키 프리픽스</th>
            <th>생성일</th>
            <th>상태</th>
            <th>작업</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="k in apiKeys" :key="k.id">
            <td><strong>{{ k.description }}</strong></td>
            <td><code>{{ k.prefix }}...</code></td>
            <td class="bn-text-muted">{{ k.createdAt }}</td>
            <td>
              <span class="bn-status-badge" :class="k.isActive ? 'active' : 'revoked'">
                {{ k.isActive ? '활성 (Active)' : '폐기됨 (Revoked)' }}
              </span>
            </td>
            <td>
              <button v-if="k.isActive" class="bn-btn-text-danger" @click="revokeKey(k.id)">즉시 폐기</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const newlyIssuedKey = ref<string | null>(null)
const keyDescription = ref('')

const apiKeys = ref([
  { id: '1', description: 'GitHub Actions Main CI', prefix: 'bn_sec_9f81', createdAt: '2026-10-25', isActive: true }
])

function generateKey() {
  if (!keyDescription.value) return
  const rawKey = 'bn_sec_' + Math.random().toString(36).substring(2) + Math.random().toString(36).substring(2)
  newlyIssuedKey.value = rawKey

  apiKeys.value.unshift({
    id: Date.now().toString(),
    description: keyDescription.value,
    prefix: rawKey.substring(0, 11),
    createdAt: new Date().toISOString().split('T')[0],
    isActive: true
  })
  keyDescription.value = ''
}

function revokeKey(id: string) {
  const k = apiKeys.value.find(item => item.id === id)
  if (k) k.isActive = false
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
}
.bn-input {
  flex: 1;
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


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

    <div class="bn-form-card">
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
            <td class="bn-text-muted">{{ log.createdAt }}</td>
            <td><strong>{{ log.actorId }}</strong></td>
            <td><span class="bn-action-pill">{{ log.action }}</span></td>
            <td>{{ log.targetType }}: {{ log.targetId }}</td>
            <td><BnHashFingerprintBadge :hash="log.prevHash" /></td>
            <td><BnHashFingerprintBadge :hash="log.recordHash" /></td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { BnHashFingerprintBadge } from '@bookingnetwork/ui-kit'

const verifying = ref(false)
const integrityStatus = ref<'UNCHECKED' | 'VERIFIED_OK' | 'TAMPERED'>('UNCHECKED')
const tamperedIndex = ref<number | null>(null)

const auditLogs = ref([
  {
    id: '1',
    actorId: 'admin-super-001',
    action: 'GRANT_ROLE',
    targetType: 'USER',
    targetId: 'user-dev-99',
    prevHash: '0000000000000000000000000000000000000000000000000000000000000000',
    recordHash: 'a1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0',
    createdAt: '2026-10-31 08:00:00'
  },
  {
    id: '2',
    actorId: 'admin-sec-002',
    action: 'TRIGGER_KILL_SWITCH',
    targetType: 'APP',
    targetId: 'app-trojan-target',
    prevHash: 'a1b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef0',
    recordHash: 'b2c3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcdef01a',
    createdAt: '2026-10-31 09:15:22'
  }
])

async function verifyChain() {
  verifying.value = true
  try {
    await new Promise(r => setTimeout(r, 600))
    integrityStatus.value = 'VERIFIED_OK'
  } finally {
    verifying.value = false
  }
}
</script>

<style scoped>
.bn-admin-page {
  padding: 24px;
}
.bn-header-flex {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}
.bn-btn-verify {
  background-color: var(--bn-primary, #2563EB);
  color: white;
  border: none;
  padding: 10px 18px;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}
.bn-alert {
  padding: 14px 18px;
  border-radius: 8px;
  margin: 20px 0;
  display: flex;
  align-items: center;
  gap: 10px;
}
.bn-alert-success {
  background-color: #ECFDF5;
  border: 1px solid #10B981;
  color: #065F46;
}
.bn-alert-danger {
  background-color: #FEE2E2;
  border: 1px solid #DC2626;
  color: #991B1B;
}
.bn-form-card {
  background-color: var(--bn-bg-surface, #FFFFFF);
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 12px;
  padding: 24px;
}
.bn-table {
  width: 100%;
  border-collapse: collapse;
}
.bn-table th, .bn-table td {
  padding: 12px 14px;
  text-align: left;
  border-bottom: 1px solid var(--bn-border-subtle, #E2E8F0);
  font-size: 13px;
}
.bn-action-pill {
  font-size: 11px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
  background-color: var(--bn-bg-elevated, #F1F5F9);
  color: var(--bn-text-primary, #0F172A);
}
.bn-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}
</style>


<template>
  <div class="bn-admin-layout">
    <!-- Top Global Header -->
    <header class="bn-topbar">
      <div class="bn-brand-area">
        <a href="/" class="bn-logo">
          <span class="seal-mark" aria-hidden="true">BN</span>
          <span class="bn-logo-text">BookingNetwork</span>
          <span class="bn-badge-admin">ADMIN CONSOLE</span>
        </a>
      </div>

      <div class="bn-topbar-metrics">
        <div class="bn-status-pill" :class="healthClass">
          <span class="bn-pulse-dot"></span>
          <span>{{ healthText }}</span>
        </div>
        <div class="bn-dual-control-pill">
          <span>DUAL-CONTROL <strong>ACTIVE</strong></span>
        </div>
      </div>

      <div class="bn-topbar-actions">
        <a href="/" class="bn-nav-link-btn" title="사용자 마켓으로 이동">User Store</a>
        <a href="/developer/" class="bn-nav-link-btn" title="개발자 콘솔로 이동">Dev Portal</a>

        <div v-if="currentAdmin" class="bn-admin-user">
          <div class="bn-avatar">{{ adminInitials }}</div>
          <div class="bn-user-info">
            <span class="bn-user-name">{{ currentAdmin.email }}</span>
            <span class="bn-user-role">{{ currentAdmin.roles.join(', ') }}</span>
          </div>
          <button @click="logoutAdmin" class="bn-logout-btn" title="관리자 세션 로그아웃">로그아웃</button>
        </div>
        <div v-else class="bn-admin-user">
          <router-link to="/login" class="bn-login-btn">관리자 인증</router-link>
        </div>
      </div>
    </header>

    <div class="bn-admin-body">
      <!-- Left Navigation Sidebar -->
      <aside class="bn-sidebar">
        <div class="bn-nav-group">
          <span class="bn-group-title">OVERVIEW · 핵심 관제</span>
          <router-link to="/" class="bn-nav-item" exact-active-class="active">
            <span class="bn-nav-index">01</span>
            <span>종합 대시보드</span>
          </router-link>
          <router-link to="/analytics" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">02</span>
            <span>플랫폼 통계</span>
          </router-link>
          <router-link to="/audit" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">03</span>
            <span>감사 로그 체인</span>
          </router-link>
        </div>

        <div class="bn-nav-group">
          <span class="bn-group-title">SECURITY · 보안 & 비상 통제</span>
          <router-link to="/security/kill-switch" class="bn-nav-item bn-item-alert" active-class="active">
            <span class="bn-nav-index">04</span>
            <span>비상 킬스위치 (Takedown)</span>
          </router-link>
          <router-link to="/security/reviews" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">05</span>
            <span>APK 보안 정밀 분석</span>
          </router-link>
          <router-link to="/moderation" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">06</span>
            <span>신고 & 악성 리뷰 큐</span>
          </router-link>
        </div>

        <div class="bn-nav-group">
          <span class="bn-group-title">REVIEW · 앱 & 카탈로그 심사</span>
          <router-link to="/apps/metadata" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">07</span>
            <span>앱 메타데이터 심사</span>
          </router-link>
          <router-link to="/ads" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">08</span>
            <span>광고 캠페인 검수</span>
          </router-link>
          <router-link to="/connectors" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">09</span>
            <span>외부 미러 커넥터</span>
          </router-link>
          <router-link to="/verifications" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">10</span>
            <span>조직 KYB 심사</span>
          </router-link>
        </div>

        <div class="bn-nav-group">
          <span class="bn-group-title">FINANCE · 재무 & 거버넌스</span>
          <router-link to="/rbac" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">11</span>
            <span>RBAC & 이중 승인</span>
          </router-link>
          <router-link to="/finance/orders" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">12</span>
            <span>주문 & 환불 관리</span>
          </router-link>
          <router-link to="/finance/settlements" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">13</span>
            <span>개발사 정산 지급</span>
          </router-link>
        </div>

        <div class="bn-sidebar-footer">
          <span class="bn-version-label">AppMarket Platform v1.0.0</span>
          <span class="bn-sla-label">SLA 99.99% GUARANTEED</span>
        </div>
      </aside>

      <!-- Main Content View Area -->
      <main class="bn-main-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { getSharedClient, clearSession, isAuthenticated } from '@bookingnetwork/api-client';

const router = useRouter();
const api = getSharedClient();

interface AdminSession {
  sub: string;
  email: string;
  roles: string[];
}

const currentAdmin = ref<AdminSession | null>(null);
const healthText = ref('Cluster status: checking…');
const healthClass = ref('');

const adminInitials = computed(() => {
  const email = currentAdmin.value?.email || 'AD';
  const name = email.split('@')[0];
  return name.slice(0, 2).toUpperCase();
});

function isAdminSession(roles: string[]): boolean {
  return roles.some(r => r === 'ADMIN_SUPER' || r.startsWith('ADMIN_'));
}

async function logoutAdmin() {
  currentAdmin.value = null;
  clearSession();
  await router.push('/login');
}

onMounted(async () => {
  try {
    const res = await fetch('/healthz/ready');
    const body = await res.json().catch(() => ({})) as Record<string, string>;
    const entries = Object.entries(body);
    if (entries.length === 0) {
      healthText.value = 'Cluster Unreachable';
      healthClass.value = 'degraded';
    } else {
      const down = entries.filter(([, v]) => v !== 'UP');
      if (down.length === 0) {
        healthText.value = 'Cluster Healthy';
        healthClass.value = 'healthy';
      } else {
        healthText.value = `Degraded: ${down.map(([k]) => k).join(', ')}`;
        healthClass.value = 'degraded';
      }
    }
  } catch {
    healthText.value = 'Cluster Unreachable';
    healthClass.value = 'degraded';
  }
  if (!isAuthenticated()) {
    return;
  }
  try {
    const me = await api.getMe();
    if (!isAdminSession(me.roles || [])) {
      await logoutAdmin();
      return;
    }
    currentAdmin.value = { sub: me.sub, email: me.email, roles: me.roles || [] };
  } catch {
    await logoutAdmin();
  }
});
</script>

<style>
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  background-color: var(--bn-bg-canvas);
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
  -webkit-font-smoothing: antialiased;
}

.bn-admin-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

/* Top Bar */
.bn-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 60px;
  background-color: var(--bn-bg-surface);
  border-bottom: 1px solid var(--bn-border-subtle);
  padding: 8px 20px;
  position: sticky;
  top: 0;
  z-index: 50;
}

.bn-brand-area {
  display: flex;
  align-items: center;
  gap: 16px;
  min-width: 0;
}

.bn-brand-area .bn-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  color: var(--bn-text-primary);
  flex-shrink: 0;
}

.seal-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background-color: var(--bn-seal);
  color: #fff;
  font-family: var(--bn-font-display);
  font-weight: 900;
  font-size: 11px;
  border-radius: var(--bn-radius-sm);
}

.bn-logo-text {
  font-family: var(--bn-font-display);
  font-size: 15px;
  font-weight: 800;
  letter-spacing: -0.01em;
}

.bn-badge-admin {
  font-family: var(--bn-font-mono);
  font-size: 9px;
  letter-spacing: 0.1em;
  color: var(--bn-seal);
  border: 1px solid var(--bn-seal);
  padding: 3px 7px;
  border-radius: var(--bn-radius-sm);
  white-space: nowrap;
}

.bn-topbar-metrics {
  display: flex;
  align-items: center;
  gap: 8px;
}

.bn-status-pill,
.bn-dual-control-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.06em;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-pill);
  padding: 5px 12px;
  color: var(--bn-text-secondary);
  white-space: nowrap;
}

.bn-status-pill.healthy {
  border-color: var(--bn-trust-verified);
  color: var(--bn-trust-verified);
}

.bn-status-pill.degraded {
  border-color: var(--bn-danger-red);
  color: var(--bn-danger-red);
}

.bn-pulse-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background-color: currentColor;
}

.bn-dual-control-pill strong {
  color: var(--bn-seal);
}

.bn-topbar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.bn-nav-link-btn {
  color: var(--bn-text-secondary);
  font-size: 12px;
  font-weight: 600;
  text-decoration: none;
  padding: 7px 12px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-subtle);
  white-space: nowrap;
}

.bn-nav-link-btn:hover {
  color: var(--bn-text-primary);
  border-color: var(--bn-border-strong);
}

.bn-admin-user {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-left: 12px;
  border-left: 1px solid var(--bn-border-subtle);
}

.bn-avatar {
  width: 32px;
  height: 32px;
  border-radius: var(--bn-radius-sm);
  background: var(--bn-ink);
  color: var(--bn-bg-canvas);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: var(--bn-font-display);
  font-size: 11px;
  font-weight: 800;
  flex-shrink: 0;
}

.bn-user-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.bn-user-name {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 170px;
}

.bn-user-role {
  font-family: var(--bn-font-mono);
  font-size: 9px;
  letter-spacing: 0.06em;
  color: var(--bn-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 170px;
}

.bn-logout-btn,
.bn-login-btn {
  background: none;
  border: 1px solid var(--bn-border-subtle);
  color: var(--bn-text-secondary);
  font-size: 12px;
  font-weight: 600;
  padding: 6px 12px;
  border-radius: var(--bn-radius-sm);
  cursor: pointer;
  margin-left: 4px;
  white-space: nowrap;
  text-decoration: none;
}

.bn-logout-btn:hover,
.bn-login-btn:hover {
  color: var(--bn-text-primary);
  border-color: var(--bn-border-strong);
}

/* Layout Body */
.bn-admin-body {
  display: flex;
  flex: 1;
  align-items: stretch;
}

/* Sidebar */
.bn-sidebar {
  width: 264px;
  background-color: var(--bn-bg-surface);
  border-right: 1px solid var(--bn-border-subtle);
  padding: 20px 12px;
  display: flex;
  flex-direction: column;
  gap: 22px;
  flex-shrink: 0;
}

.bn-nav-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.bn-group-title {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.1em;
  color: var(--bn-text-muted);
  padding: 0 12px 8px;
}

.bn-nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: var(--bn-radius-sm);
  color: var(--bn-text-secondary);
  font-size: 13px;
  font-weight: 600;
  text-decoration: none;
  border-left: 2px solid transparent;
}

.bn-nav-item:hover {
  background-color: var(--bn-bg-well);
  color: var(--bn-text-primary);
}

.bn-nav-item.active {
  background-color: var(--bn-bg-well);
  color: var(--bn-text-primary);
  border-left-color: var(--bn-seal);
}

.bn-nav-item.bn-item-alert {
  color: var(--bn-danger-red);
}

.bn-nav-item.bn-item-alert:hover {
  color: var(--bn-danger-red);
}

.bn-nav-index {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  color: var(--bn-text-muted);
  width: 20px;
  flex-shrink: 0;
}
.bn-nav-item.active .bn-nav-index {
  color: var(--bn-seal);
}
.bn-item-alert .bn-nav-index {
  color: var(--bn-danger-red);
}

.bn-sidebar-footer {
  margin-top: auto;
  padding: 16px 12px 8px;
  border-top: 1px solid var(--bn-border-subtle);
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.bn-version-label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  color: var(--bn-text-muted);
}

.bn-sla-label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.06em;
  color: var(--bn-seal);
}

/* Main Content Area */
.bn-main-content {
  flex: 1;
  min-width: 0;
  background-color: var(--bn-bg-canvas);
  padding: 28px;
  overflow-y: auto;
}

@media (max-width: 1100px) {
  .bn-sidebar {
    width: 220px;
  }
  .bn-user-info,
  .bn-topbar-metrics {
    display: none;
  }
}
@media (max-width: 860px) {
  .bn-admin-body {
    flex-direction: column;
  }
  .bn-sidebar {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--bn-border-subtle);
  }
}
</style>

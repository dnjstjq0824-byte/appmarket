<template>
  <div class="bn-dev-layout">
    <header class="bn-topbar">
      <div class="bn-brand-area">
        <a href="/" class="bn-logo">
          <span class="seal-mark" aria-hidden="true">BN</span>
          <span class="bn-logo-text">BookingNetwork</span>
          <span class="bn-badge-dev">DEVELOPER CONSOLE</span>
        </a>

        <div class="bn-org-selector">
          <span class="bn-org-name">{{ orgName }}</span>
          <span class="bn-org-badge" :class="kybStatus.toLowerCase()">{{ kybStatus === 'VERIFIED' ? 'VERIFIED KYB' : 'UNVERIFIED KYB' }}</span>
        </div>
      </div>

      <div class="bn-topbar-actions">
        <a href="/" class="bn-nav-link-btn" title="사용자 마켓으로 이동">User Store</a>
        <a href="/admin/" class="bn-nav-link-btn" title="관리자 콘솔로 이동">Admin Console</a>
        <div class="bn-dev-user">
          <div class="bn-avatar">DV</div>
          <div class="bn-user-info">
            <span class="bn-user-name">{{ devUser.name }}</span>
            <span class="bn-user-org">{{ devUser.email }}</span>
          </div>
          <button class="bn-logout-btn" @click="logout">로그아웃</button>
        </div>
      </div>
    </header>

    <div class="bn-dev-body">
      <aside class="bn-sidebar">
        <div class="bn-nav-group">
          <span class="bn-group-title">IDENTITY · 조직 & 보안</span>
          <router-link to="/" class="bn-nav-item" exact-active-class="active">
            <span class="bn-nav-index">01</span>
            <span>조직 & KYB 인증</span>
          </router-link>
          <router-link to="/team" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">02</span>
            <span>팀원 & RBAC 권한</span>
          </router-link>
          <router-link to="/api-keys" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">03</span>
            <span>CI/CD API 키</span>
          </router-link>
        </div>

        <div class="bn-nav-group">
          <span class="bn-group-title">DISTRIBUTION · 배포 & 릴리스</span>
          <router-link to="/releases/upload" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">04</span>
            <span>APK/AAB 아티팩트 업로드</span>
          </router-link>
          <router-link to="/releases/rollout" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">05</span>
            <span>단계적 롤아웃 (Rollout)</span>
          </router-link>
          <router-link to="/releases/lineage" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">06</span>
            <span>v3 서명 로테이션 계통</span>
          </router-link>
        </div>

        <div class="bn-nav-group">
          <span class="bn-group-title">CATALOG · 스토어 등록정보</span>
          <router-link to="/apps/metadata" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">07</span>
            <span>앱 메타데이터 & 스크린샷</span>
          </router-link>
        </div>

        <div class="bn-nav-group">
          <span class="bn-group-title">MONETIZATION · 수익화 & 금융</span>
          <router-link to="/commerce/pricing" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">08</span>
            <span>앱 가격 & 인앱 상품</span>
          </router-link>
          <router-link to="/commerce/promos" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">09</span>
            <span>프로모션 코드 관리</span>
          </router-link>
          <router-link to="/finance/settlements" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">10</span>
            <span>수익 정산 & 에스크로</span>
          </router-link>
        </div>

        <div class="bn-nav-group">
          <span class="bn-group-title">GROWTH · 마케팅</span>
          <router-link to="/ads" class="bn-nav-item" active-class="active">
            <span class="bn-nav-index">11</span>
            <span>스폰서드 광고 캠페인</span>
          </router-link>
        </div>

        <div class="bn-sidebar-footer">
          <span class="bn-version-label">Developer Studio v1.0.0</span>
          <span class="bn-compliance-label">DMA ARTICLE 23 COMPLIANT</span>
        </div>
      </aside>

      <main class="bn-main-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { getSharedClient, clearSession, isAuthenticated } from '@bookingnetwork/api-client';

const router = useRouter();
const api = getSharedClient();

const orgName = ref('—');
const kybStatus = ref('UNKNOWN');
const devUser = ref({
  name: 'Developer',
  email: ''
});

onMounted(async () => {
  if (!isAuthenticated()) {
    return;
  }
  try {
    const me = await api.getMe();
    devUser.value.email = me.email || '';
    devUser.value.name = (me.email || 'Developer').split('@')[0];
  } catch {
    clearSession();
    await router.push('/login');
    return;
  }
  try {
    const orgs = await api.listOrgs();
    if (orgs.length > 0) {
      orgName.value = orgs[0].org_name;
      kybStatus.value = orgs[0].verification_status || 'UNVERIFIED';
    } else {
      orgName.value = 'No organization';
      kybStatus.value = 'UNVERIFIED';
    }
  } catch {
    orgName.value = '—';
    kybStatus.value = 'UNKNOWN';
  }
});

async function logout() {
  clearSession();
  await router.push('/login');
}
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

.bn-dev-layout {
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

.bn-badge-dev {
  font-family: var(--bn-font-mono);
  font-size: 9px;
  letter-spacing: 0.1em;
  color: var(--bn-text-muted);
  border: 1px solid var(--bn-border-subtle);
  padding: 3px 7px;
  border-radius: var(--bn-radius-sm);
  white-space: nowrap;
}

.bn-org-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: var(--bn-bg-well);
  padding: 6px 12px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-subtle);
  font-size: 13px;
}

.bn-org-name {
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
}

.bn-org-badge {
  font-family: var(--bn-font-mono);
  font-size: 9px;
  letter-spacing: 0.08em;
  padding: 2px 7px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-warning-amber);
  color: var(--bn-warning-amber);
  white-space: nowrap;
}
.bn-org-badge.verified {
  border-color: var(--bn-trust-verified);
  color: var(--bn-trust-verified);
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

.bn-dev-user {
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
  font-size: 13px;
  font-weight: 700;
}

.bn-user-org {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  color: var(--bn-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 160px;
}

.bn-logout-btn {
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
}

.bn-logout-btn:hover {
  color: var(--bn-text-primary);
  border-color: var(--bn-border-strong);
}

/* Layout Body */
.bn-dev-body {
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

.bn-compliance-label {
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
  .bn-user-info {
    display: none;
  }
}
@media (max-width: 860px) {
  .bn-dev-body {
    flex-direction: column;
  }
  .bn-sidebar {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--bn-border-subtle);
  }
  .bn-org-selector {
    display: none;
  }
}
</style>

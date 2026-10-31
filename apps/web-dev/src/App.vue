<template>
  <div class="bn-dev-layout">
    <!-- Top Global Header -->
    <header class="bn-topbar">
      <div class="bn-brand-area">
        <a href="/" class="bn-logo">
          <span class="bn-logo-icon">▲</span>
          <span class="bn-logo-text">BookingNetwork</span>
          <span class="bn-badge-dev">DEVELOPER CONSOLE</span>
        </a>

        <div class="bn-org-selector">
          <span class="bn-org-icon">🏢</span>
          <span class="bn-org-name">OmniWork Technologies</span>
          <span class="bn-org-badge">VERIFIED KYB</span>
        </div>
      </div>

      <div class="bn-topbar-actions">
        <a href="/" class="bn-nav-link-btn" title="사용자 마켓으로 이동">🛒 User Store</a>
        <a href="/admin/" class="bn-nav-link-btn" title="관리자 콘솔로 이동">⚙️ Admin Console</a>
        <div class="bn-dev-user">
          <div class="bn-avatar">DV</div>
          <div class="bn-user-info">
            <span class="bn-user-name">Lead Developer</span>
            <span class="bn-user-org">dev@omniwork.com</span>
          </div>
        </div>
      </div>
    </header>

    <div class="bn-dev-body">
      <!-- Left Navigation Sidebar -->
      <aside class="bn-sidebar">
        <div class="bn-nav-group">
          <span class="bn-group-title">조직 & 보안 (Identity)</span>
          <router-link to="/" class="bn-nav-item" exact-active-class="active">
            <span class="bn-nav-icon">🏢</span>
            <span>조직 & KYB 인증</span>
          </router-link>
          <router-link to="/team" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">👥</span>
            <span>팀원 & RBAC 권한</span>
          </router-link>
          <router-link to="/api-keys" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">🔑</span>
            <span>CI/CD API 키</span>
          </router-link>
        </div>

        <div class="bn-nav-group">
          <span class="bn-group-title">배포 & 릴리스 (Distribution)</span>
          <router-link to="/releases/upload" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">🚀</span>
            <span>APK/AAB 아티팩트 업로드</span>
          </router-link>
          <router-link to="/releases/rollout" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">📊</span>
            <span>단계적 롤아웃 (Rollout)</span>
          </router-link>
          <router-link to="/releases/lineage" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">🛡️</span>
            <span>v3 서명 로테이션 계통</span>
          </router-link>
        </div>

        <div class="bn-nav-group">
          <span class="bn-group-title">스토어 등록정보 (Catalog)</span>
          <router-link to="/apps/metadata" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">📝</span>
            <span>앱 메타데이터 & 스크린샷</span>
          </router-link>
        </div>

        <div class="bn-nav-group">
          <span class="bn-group-title">수익화 & 금융 (Monetization)</span>
          <router-link to="/commerce/pricing" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">🏷️</span>
            <span>앱 가격 & 인앱 상품</span>
          </router-link>
          <router-link to="/commerce/promos" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">🎟️</span>
            <span>프로모션 코드 관리</span>
          </router-link>
          <router-link to="/finance/settlements" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">💰</span>
            <span>수익 정산 & 에스크로</span>
          </router-link>
        </div>

        <div class="bn-nav-group">
          <span class="bn-group-title">마케팅 (Growth)</span>
          <router-link to="/ads" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">📢</span>
            <span>스폰서드 광고 캠페인</span>
          </router-link>
        </div>

        <div class="bn-sidebar-footer">
          <span class="bn-version-label">Developer Studio v1.0.0</span>
          <span class="bn-compliance-label">Article 23 Compliant</span>
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
// Developer global layout
</script>

<style>
/* Global Reset and Dark Studio Palette */
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  background-color: #0b1120;
  color: #f1f5f9;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
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
  height: 64px;
  background-color: #0f172a;
  border-bottom: 1px solid #1e293b;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: 50;
}

.bn-brand-area {
  display: flex;
  align-items: center;
  gap: 20px;
}

.bn-brand-area .bn-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  color: #f8fafc;
}

.bn-logo-icon {
  font-size: 20px;
  color: #10b981;
}

.bn-logo-text {
  font-size: 17px;
  font-weight: 700;
  letter-spacing: -0.3px;
}

.bn-badge-dev {
  background-color: #0284c7;
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 4px;
  letter-spacing: 0.5px;
}

.bn-org-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: #1e293b;
  padding: 6px 12px;
  border-radius: 8px;
  border: 1px solid #334155;
  font-size: 13px;
}

.bn-org-icon {
  font-size: 14px;
}

.bn-org-name {
  font-weight: 600;
  color: #f8fafc;
}

.bn-org-badge {
  background-color: #065f46;
  color: #34d399;
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}

.bn-topbar-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.bn-nav-link-btn {
  color: #94a3b8;
  font-size: 13px;
  font-weight: 500;
  text-decoration: none;
  padding: 6px 12px;
  border-radius: 6px;
  border: 1px solid #334155;
  transition: all 0.15s ease;
}

.bn-nav-link-btn:hover {
  background-color: #1e293b;
  color: #f8fafc;
}

.bn-dev-user {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-left: 12px;
  border-left: 1px solid #334155;
}

.bn-avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  background: linear-gradient(135deg, #059669, #0284c7);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
}

.bn-user-info {
  display: flex;
  flex-direction: column;
}

.bn-user-name {
  font-size: 13px;
  font-weight: 600;
  color: #f8fafc;
}

.bn-user-org {
  font-size: 11px;
  color: #64748b;
}

/* Layout Body */
.bn-dev-body {
  display: flex;
  flex: 1;
}

/* Left Sidebar */
.bn-sidebar {
  width: 260px;
  background-color: #0f172a;
  border-right: 1px solid #1e293b;
  padding: 20px 12px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  flex-shrink: 0;
}

.bn-nav-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.bn-group-title {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: #64748b;
  letter-spacing: 0.5px;
  padding: 0 12px 6px;
}

.bn-nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 8px;
  color: #94a3b8;
  font-size: 13px;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.15s ease;
}

.bn-nav-item:hover {
  background-color: #1e293b;
  color: #f8fafc;
}

.bn-nav-item.active {
  background-color: #0284c7;
  color: #ffffff;
  font-weight: 600;
}

.bn-sidebar-footer {
  margin-top: auto;
  padding: 16px 12px 8px;
  border-top: 1px solid #1e293b;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.bn-version-label {
  font-size: 11px;
  color: #64748b;
  font-family: monospace;
}

.bn-compliance-label {
  font-size: 11px;
  color: #38bdf8;
  font-weight: 600;
}

/* Main Content Area */
.bn-main-content {
  flex: 1;
  background-color: #0b1120;
  overflow-y: auto;
  min-height: calc(100vh - 64px);
}
</style>

<template>
  <div class="bn-admin-layout">
    <!-- Top Global Header -->
    <header class="bn-topbar">
      <div class="bn-brand-area">
        <a href="/" class="bn-logo">
          <span class="bn-logo-icon">▲</span>
          <span class="bn-logo-text">BookingNetwork</span>
          <span class="bn-badge-admin">ADMIN CONSOLE</span>
        </a>
      </div>

      <div class="bn-topbar-metrics">
        <div class="bn-status-pill">
          <span class="bn-pulse-dot"></span>
          <span>Cluster Healthy (QPS: 1,482)</span>
        </div>
        <div class="bn-dual-control-pill">
          <span>Dual-Control: <strong>ACTIVE</strong></span>
        </div>
      </div>

      <div class="bn-topbar-actions">
        <a href="/" class="bn-nav-link-btn" title="사용자 마켓으로 이동">🛒 User Store</a>
        <a href="/developer/" class="bn-nav-link-btn" title="개발자 콘솔로 이동">💻 Dev Portal</a>
        <div class="bn-admin-user">
          <div class="bn-avatar">SA</div>
          <div class="bn-user-info">
            <span class="bn-user-name">Super Admin</span>
            <span class="bn-user-role">Security Officer</span>
          </div>
        </div>
      </div>
    </header>

    <div class="bn-admin-body">
      <!-- Left Navigation Sidebar -->
      <aside class="bn-sidebar">
        <div class="bn-nav-group">
          <span class="bn-group-title">핵심 관제 (Overview)</span>
          <router-link to="/" class="bn-nav-item" exact-active-class="active">
            <span class="bn-nav-icon">📊</span>
            <span>종합 대시보드</span>
          </router-link>
          <router-link to="/analytics" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">📈</span>
            <span>플랫폼 통계</span>
          </router-link>
          <router-link to="/audit" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">📜</span>
            <span>감사 로그 체인</span>
          </router-link>
        </div>

        <div class="bn-nav-group">
          <span class="bn-group-title">보안 & 비상 통제 (Security)</span>
          <router-link to="/security/kill-switch" class="bn-nav-item bn-item-alert" active-class="active">
            <span class="bn-nav-icon">🚨</span>
            <span>비상 킬스위치 (Takedown)</span>
          </router-link>
          <router-link to="/security/reviews" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">🛡️</span>
            <span>APK 보안 정밀 분석</span>
          </router-link>
          <router-link to="/moderation" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">⚖️</span>
            <span>신고 & 악성 리뷰 큐</span>
          </router-link>
        </div>

        <div class="bn-nav-group">
          <span class="bn-group-title">앱 & 카탈로그 심사 (Review)</span>
          <router-link to="/apps/metadata" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">📝</span>
            <span>앱 메타데이터 심사</span>
          </router-link>
          <router-link to="/ads" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">📢</span>
            <span>광고 캠페인 검수</span>
          </router-link>
          <router-link to="/connectors" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">🔌</span>
            <span>외부 미러 커넥터</span>
          </router-link>
        </div>

        <div class="bn-nav-group">
          <span class="bn-group-title">재무 & 거버넌스 (Finance)</span>
          <router-link to="/rbac" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">👥</span>
            <span>RBAC & 이중 승인</span>
          </router-link>
          <router-link to="/finance/orders" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">💳</span>
            <span>주문 & 환불 관리</span>
          </router-link>
          <router-link to="/finance/settlements" class="bn-nav-item" active-class="active">
            <span class="bn-nav-icon">🏦</span>
            <span>개발사 정산 지급</span>
          </router-link>
        </div>

        <div class="bn-sidebar-footer">
          <span class="bn-version-label">AppMarket Platform v1.0.0</span>
          <span class="bn-sla-label">SLA 99.99% Guaranteed</span>
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
// Admin global layout
</script>

<style>
/* Global Reset and Dark Admin Palette */
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
  height: 64px;
  background-color: #0f172a;
  border-bottom: 1px solid #1e293b;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: 50;
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
  color: #38bdf8;
}

.bn-logo-text {
  font-size: 17px;
  font-weight: 700;
  letter-spacing: -0.3px;
}

.bn-badge-admin {
  background-color: #dc2626;
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 4px;
  letter-spacing: 0.5px;
}

.bn-topbar-metrics {
  display: flex;
  align-items: center;
  gap: 16px;
}

.bn-status-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: #064e3b;
  color: #34d399;
  font-size: 12px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 9999px;
  border: 1px solid #059669;
}

.bn-pulse-dot {
  width: 8px;
  height: 8px;
  background-color: #10b981;
  border-radius: 50%;
  box-shadow: 0 0 8px #10b981;
}

.bn-dual-control-pill {
  background-color: #1e293b;
  color: #94a3b8;
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 6px;
  border: 1px solid #334155;
}

.bn-dual-control-pill strong {
  color: #38bdf8;
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

.bn-admin-user {
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
  background: linear-gradient(135deg, #38bdf8, #2563eb);
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

.bn-user-role {
  font-size: 11px;
  color: #64748b;
}

/* Layout Body */
.bn-admin-body {
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
  background-color: #1d4ed8;
  color: #ffffff;
  font-weight: 600;
}

.bn-item-alert:hover {
  background-color: #7f1d1d;
  color: #fca5a5;
}

.bn-item-alert.active {
  background-color: #b91c1c;
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

.bn-sla-label {
  font-size: 11px;
  color: #10b981;
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

<template>
  <div class="bn-user-store-root" :data-theme="currentTheme">
    <!-- Global Header Navbar -->
    <header class="bn-navbar">
      <div class="bn-navbar-inner">
        <!-- Logo -->
        <NuxtLink to="/" class="bn-logo">
          <div class="bn-logo-badge">▲</div>
          <div class="bn-logo-title-group">
            <span class="bn-logo-title">BookingNetwork</span>
            <span class="bn-logo-sub">APP MARKET</span>
          </div>
        </NuxtLink>

        <!-- Global Search Bar -->
        <div class="bn-search-container">
          <span class="bn-search-icon">🔍</span>
          <input
            v-model="searchQuery"
            type="search"
            placeholder="앱, 게임, 개발사 패키지 검색... (Enter)"
            class="bn-search-input"
            @keyup.enter="handleSearch"
          />
        </div>

        <!-- Navigation Links & Portals -->
        <nav class="bn-nav-links">
          <NuxtLink to="/" class="bn-nav-link" active-class="active">스토어 홈</NuxtLink>
          <NuxtLink to="/account/library" class="bn-nav-link" active-class="active">내 보관함</NuxtLink>
          <NuxtLink to="/support/report" class="bn-nav-link" active-class="active">신고 센터</NuxtLink>
        </nav>

        <!-- Right Action Portals & Theme -->
        <div class="bn-actions">
          <a href="/developer/" class="bn-portal-btn bn-dev-btn" title="개발자 콘솔로 이동">
            <span>💻</span> 개발자 콘솔
          </a>
          <a href="/admin/" class="bn-portal-btn bn-admin-btn" title="관리자 콘솔로 이동">
            <span>⚙️</span> 관리자 콘솔
          </a>
          <button class="bn-theme-toggle" @click="toggleTheme" title="다크/라이트 모드 전환">
            {{ currentTheme === 'dark' ? '☀️' : '🌙' }}
          </button>
        </div>
      </div>
    </header>

    <!-- Page Body -->
    <main class="bn-body">
      <NuxtPage />
    </main>

    <!-- Global Footer -->
    <footer class="bn-footer">
      <div class="bn-footer-inner">
        <div class="bn-footer-col">
          <div class="bn-footer-brand">
            <span class="bn-footer-logo">▲ BookingNetwork App Market</span>
            <span class="bn-compliance-pill">EU DMA Article 23 Compliant</span>
          </div>
          <p class="bn-footer-desc">
            글로벌 독립 앱 유통 마켓플레이스입니다. 구글 플레이 스크래핑을 원천 금지하며,
            V2/V3 서명 무결성 검증과 투명한 직접 배포(Direct-Hosted) 생태계를 제공합니다.
          </p>
        </div>

        <div class="bn-footer-col">
          <h4>서비스 바로가기</h4>
          <ul>
            <li><NuxtLink to="/">인기 애플리케이션</NuxtLink></li>
            <li><NuxtLink to="/search">고급 패키지 검색</NuxtLink></li>
            <li><NuxtLink to="/account/library">구매 및 설치 내역</NuxtLink></li>
            <li><a href="/developer/">개발자 등록 및 APK 배포</a></li>
            <li><a href="/admin/">플랫폼 관리자 콘솔</a></li>
          </ul>
        </div>

        <div class="bn-footer-col">
          <h4>신뢰 및 규정 준수</h4>
          <ul>
            <li><NuxtLink to="/support/report">악성 소프트웨어 / 보안 신고</NuxtLink></li>
            <li><a href="#terms">이용약관 및 개발사 계약</a></li>
            <li><a href="#privacy">개인정보처리방침</a></li>
            <li><a href="#legal">전자상거래 및 소비자보호 규정</a></li>
          </ul>
        </div>
      </div>

      <div class="bn-footer-bottom">
        <p>© 2026 Booking Network Inc. All rights reserved. Universal Independent App Store.</p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const searchQuery = ref('')
const currentTheme = ref('dark')

function toggleTheme() {
  currentTheme.value = currentTheme.value === 'dark' ? 'light' : 'dark'
  if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('data-theme', currentTheme.value)
  }
}

function handleSearch() {
  if (searchQuery.value.trim()) {
    router.push({ path: '/search', query: { q: searchQuery.value.trim() } })
  }
}

onMounted(() => {
  if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('data-theme', currentTheme.value)
  }
})
</script>

<style>
/* Global Store Shell Styles */
.bn-user-store-root {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: var(--bn-bg-canvas);
  color: var(--bn-text-primary);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  transition: background-color 0.2s ease, color 0.2s ease;
}

/* Navbar */
.bn-navbar {
  position: sticky;
  top: 0;
  z-index: 100;
  background-color: var(--bn-bg-surface);
  border-bottom: 1px solid var(--bn-border-subtle);
  backdrop-filter: blur(12px);
}

.bn-navbar-inner {
  max-width: 1280px;
  margin: 0 auto;
  height: 70px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

/* Logo */
.bn-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  text-decoration: none;
  flex-shrink: 0;
}

.bn-logo-badge {
  width: 38px;
  height: 38px;
  border-radius: 10px;
  background: linear-gradient(135deg, #2563eb, #38bdf8);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 800;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
}

.bn-logo-title-group {
  display: flex;
  flex-direction: column;
}

.bn-logo-title {
  font-size: 17px;
  font-weight: 800;
  letter-spacing: -0.4px;
  color: var(--bn-text-primary);
}

.bn-logo-sub {
  font-size: 10px;
  font-weight: 700;
  color: var(--bn-primary);
  letter-spacing: 1px;
}

/* Search Bar */
.bn-search-container {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: var(--bn-bg-elevated);
  border: 1px solid var(--bn-border-subtle);
  border-radius: 10px;
  padding: 8px 14px;
  flex: 1;
  max-width: 440px;
  transition: all 0.2s ease;
}

.bn-search-container:focus-within {
  border-color: var(--bn-primary);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.15);
}

.bn-search-icon {
  font-size: 14px;
  opacity: 0.6;
}

.bn-search-input {
  border: none;
  background: transparent;
  width: 100%;
  color: var(--bn-text-primary);
  font-size: 14px;
  outline: none;
}

/* Nav Links */
.bn-nav-links {
  display: flex;
  align-items: center;
  gap: 20px;
}

.bn-nav-link {
  color: var(--bn-text-secondary);
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  padding: 6px 10px;
  border-radius: 6px;
  transition: all 0.15s ease;
}

.bn-nav-link:hover {
  color: var(--bn-primary);
  background-color: var(--bn-bg-elevated);
}

.bn-nav-link.active {
  color: var(--bn-primary);
  font-weight: 700;
}

/* Right Actions */
.bn-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.bn-portal-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  text-decoration: none;
  padding: 7px 12px;
  border-radius: 8px;
  transition: all 0.15s ease;
}

.bn-dev-btn {
  background-color: rgba(2, 132, 199, 0.12);
  color: #38bdf8;
  border: 1px solid rgba(2, 132, 199, 0.3);
}

.bn-dev-btn:hover {
  background-color: rgba(2, 132, 199, 0.25);
}

.bn-admin-btn {
  background-color: rgba(220, 38, 38, 0.12);
  color: #f87171;
  border: 1px solid rgba(220, 38, 38, 0.3);
}

.bn-admin-btn:hover {
  background-color: rgba(220, 38, 38, 0.25);
}

.bn-theme-toggle {
  background-color: var(--bn-bg-elevated);
  border: 1px solid var(--bn-border-subtle);
  border-radius: 8px;
  padding: 7px 10px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.15s ease;
}

.bn-theme-toggle:hover {
  background-color: var(--bn-border-subtle);
}

/* Body */
.bn-body {
  flex: 1;
}

/* Footer */
.bn-footer {
  background-color: var(--bn-bg-surface);
  border-top: 1px solid var(--bn-border-subtle);
  padding: 48px 24px 24px;
  margin-top: 60px;
}

.bn-footer-inner {
  max-width: 1280px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 2fr 1fr 1fr;
  gap: 40px;
}

.bn-footer-brand {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}

.bn-footer-logo {
  font-size: 16px;
  font-weight: 700;
}

.bn-compliance-pill {
  background-color: rgba(5, 150, 105, 0.15);
  color: var(--bn-trust-verified);
  font-size: 11px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 9999px;
  border: 1px solid var(--bn-trust-verified);
}

.bn-footer-desc {
  font-size: 13px;
  color: var(--bn-text-muted);
  line-height: 1.6;
  max-width: 480px;
}

.bn-footer-col h4 {
  font-size: 14px;
  font-weight: 700;
  margin-bottom: 16px;
  color: var(--bn-text-primary);
}

.bn-footer-col ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.bn-footer-col a {
  color: var(--bn-text-muted);
  font-size: 13px;
  text-decoration: none;
  transition: color 0.15s ease;
}

.bn-footer-col a:hover {
  color: var(--bn-primary);
}

.bn-footer-bottom {
  max-width: 1280px;
  margin: 32px auto 0;
  padding-top: 24px;
  border-top: 1px solid var(--bn-border-subtle);
  text-align: center;
  font-size: 12px;
  color: var(--bn-text-muted);
}
</style>

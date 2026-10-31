<template>
  <div class="bn-playstore-root" :data-theme="currentTheme">
    <!-- Smart App Banner for Android Companion Agent -->
    <div v-if="showSmartBanner" class="smart-companion-banner">
      <div class="banner-inner">
        <button class="banner-close" @click="showSmartBanner = false" title="배너 닫기">✕</button>
        <div class="banner-icon">
          <svg viewBox="0 0 24 24" width="22" height="22" fill="#01875F">
            <path d="M17.523 15.3414c-.5511 0-.9993-.4486-.9993-1.0003s.4482-1.0003.9993-1.0003c.551 0 .9992.4486.9992 1.0003s-.4482 1.0003-.9992 1.0003m-11.046 0c-.5511 0-.9993-.4486-.9993-1.0003s.4482-1.0003.9993-1.0003c.5511 0 .9993.4486.9993 1.0003s-.4482 1.0003-.9993 1.0003m11.4045-6.02l1.996-3.4566c.1154-.1994.0471-.4535-.1527-.5688-.1994-.1152-.4535-.0471-.5688.1528l-2.0223 3.5024c-1.5791-.7197-3.3444-1.127-5.2347-1.127-1.8903 0-3.6556.4073-5.2347 1.127L4.6409 5.4491c-.1153-.1999-.3694-.268-.5688-.1528-.1998.1153-.2681.3694-.1527.5688l1.996 3.4566C2.6889 11.0867 0 14.6738 0 18.9167h24c0-4.2429-2.6889-7.83-5.8815-9.5953"/>
          </svg>
        </div>
        <div class="banner-text">
          <strong>BookingNetwork 모바일 공식 앱</strong>
          <span>PackageInstaller 원클릭 설치 및 백그라운드 자동 업데이트 지원 (Android 12+ 패키지 소유권 준수)</span>
        </div>
        <button class="banner-download-btn" @click="downloadCompanionApp">
          앱 설치 (APK)
        </button>
      </div>
    </div>

    <!-- Google Play Style Sticky Top Navigation Header -->
    <header class="play-header">
      <div class="play-header-container">
        <!-- Brand Logo & Title -->
        <NuxtLink to="/" class="play-logo" @click="activeTopTab = 'apps'">
          <div class="play-logo-icon">
            <svg viewBox="0 0 40 40" width="34" height="34" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M7 6.5C7 5.2 8.4 4.4 9.5 5.1L32 18.5C33.1 19.2 33.1 20.8 32 21.5L9.5 34.9C8.4 35.6 7 34.8 7 33.5V6.5Z" fill="url(#playGradMain)"/>
              <path d="M7 6.5L24 23.5L32 18.5L9.5 5.1C8.4 4.4 7 5.2 7 6.5Z" fill="url(#playGradTop)" opacity="0.9"/>
              <path d="M7 33.5L24 16.5L32 21.5L9.5 34.9C8.4 35.6 7 34.8 7 33.5Z" fill="url(#playGradBtm)" opacity="0.85"/>
              <defs>
                <linearGradient id="playGradMain" x1="7" y1="5" x2="33" y2="35" gradientUnits="userSpaceOnUse">
                  <stop stop-color="#00C9FF"/>
                  <stop offset="0.5" stop-color="#0084FF"/>
                  <stop offset="1" stop-color="#00E575"/>
                </linearGradient>
                <linearGradient id="playGradTop" x1="7" y1="5" x2="32" y2="20" gradientUnits="userSpaceOnUse">
                  <stop stop-color="#00E575"/>
                  <stop offset="1" stop-color="#FFDD00"/>
                </linearGradient>
                <linearGradient id="playGradBtm" x1="7" y1="20" x2="32" y2="35" gradientUnits="userSpaceOnUse">
                  <stop stop-color="#FF3366"/>
                  <stop offset="1" stop-color="#9900FF"/>
                </linearGradient>
              </defs>
            </svg>
          </div>
          <div class="play-logo-text">
            <span class="brand-name">Google Play <span class="brand-partner">| BookingNetwork</span></span>
            <span class="brand-sub">Universal Independent Store</span>
          </div>
        </NuxtLink>

        <!-- Top Play Tabs (Google Play Web Style Navigation) -->
        <nav class="play-top-tabs">
          <button
            class="top-tab-btn"
            :class="{ active: activeTopTab === 'games' }"
            @click="switchTopTab('games')"
          >
            <span class="tab-icon">🎮</span> 게임
          </button>
          <button
            class="top-tab-btn"
            :class="{ active: activeTopTab === 'apps' }"
            @click="switchTopTab('apps')"
          >
            <span class="tab-icon">📱</span> 앱
          </button>
          <button
            class="top-tab-btn"
            :class="{ active: activeTopTab === 'charts' }"
            @click="switchTopTab('charts')"
          >
            <span class="tab-icon">🏆</span> 인기 차트
          </button>
          <button
            class="top-tab-btn"
            :class="{ active: activeTopTab === 'editors' }"
            @click="switchTopTab('editors')"
          >
            <span class="tab-icon">✨</span> 에디터 추천
          </button>
          <button
            class="top-tab-btn"
            :class="{ active: activeTopTab === 'protect' }"
            @click="switchTopTab('protect')"
          >
            <span class="tab-icon">🛡️</span> Play 프로텍트
          </button>
        </nav>

        <!-- Search Bar (Pill Shaped Google Play Style) -->
        <div class="play-search-box" :class="{ focused: searchFocused }">
          <svg class="search-svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
          <input
            v-model="searchQuery"
            type="search"
            placeholder="앱 및 게임 검색..."
            class="search-input"
            @focus="searchFocused = true"
            @blur="handleSearchBlur"
            @keyup.enter="executeSearch"
          />
          <button v-if="searchQuery" class="search-clear-btn" @click="searchQuery = ''">✕</button>

          <!-- Autocomplete Dropdown Preview -->
          <div v-if="searchFocused && quickSuggestions.length > 0" class="search-suggestions-dropdown">
            <div class="suggestions-label">추천 검색어</div>
            <div
              v-for="item in quickSuggestions"
              :key="item"
              class="suggestion-item"
              @mousedown.prevent="selectSuggestion(item)"
            >
              <span class="suggestion-icon">🔍</span>
              <span>{{ item }}</span>
            </div>
          </div>
        </div>

        <!-- Right Side Utility & Profile Menu -->
        <div class="play-header-actions">
          <!-- Dark / Light Theme Toggle -->
          <button class="icon-circle-btn" @click="toggleTheme" title="다크/라이트 테마 전환">
            {{ currentTheme === 'dark' ? '☀️' : '🌙' }}
          </button>

          <!-- Help / Documentation -->
          <NuxtLink to="/support/report" class="icon-circle-btn" title="보안 검역 및 고객 신고">
            ❓
          </NuxtLink>

          <!-- User Google Play Style Avatar & Dropdown -->
          <div class="user-profile-menu-container">
            <button class="user-avatar-btn" @click="profileMenuOpen = !profileMenuOpen" title="계정 및 서비스 전환">
              <span class="avatar-letter">BN</span>
            </button>

            <!-- Profile Popover Modal (Google Style) -->
            <div v-if="profileMenuOpen" class="user-profile-popover" @click.stop>
              <div class="popover-header">
                <div class="popover-avatar">BN</div>
                <div class="popover-user-info">
                  <div class="popover-name">BookingNetwork User</div>
                  <div class="popover-email">user@bookingnetwork.com</div>
                  <span class="popover-badge">EU DMA Verified Direct Account</span>
                </div>
              </div>

              <div class="popover-section">
                <NuxtLink to="/account/library" class="popover-item" @click="profileMenuOpen = false">
                  <span class="popover-item-icon">📦</span> 내 앱 및 게임 (보관함)
                </NuxtLink>
                <a href="#payments" class="popover-item" @click="profileMenuOpen = false">
                  <span class="popover-item-icon">💳</span> 결제 및 정산 내역
                </a>
              </div>

              <div class="popover-divider"></div>

              <div class="popover-section">
                <div class="popover-section-label">엔터프라이즈 포털 바로가기</div>
                <a href="/developer/" class="popover-item dev-link">
                  <span class="popover-item-icon">💻</span>
                  <div class="popover-link-text">
                    <strong>개발자 콘솔 (Developer Studio)</strong>
                    <small>KYB 인증, APK 업로드, 릴리즈 파이프라인</small>
                  </div>
                </a>
                <a href="/admin/" class="popover-item admin-link">
                  <span class="popover-item-icon">⚙️</span>
                  <div class="popover-link-text">
                    <strong>플랫폼 관리자 (Admin Console)</strong>
                    <small>4-Eyes 긴급 킬스위치, 실시간 QPS, 심사 큐</small>
                  </div>
                </a>
              </div>

              <div class="popover-footer">
                <button class="popover-signout-btn" @click="profileMenuOpen = false">닫기</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>

    <!-- Store Body Content -->
    <main class="play-content-area" @click="profileMenuOpen = false">
      <NuxtPage />
    </main>

    <!-- Google Play Style Clean Footer -->
    <footer class="play-footer">
      <div class="play-footer-inner">
        <div class="play-footer-nav">
          <div class="footer-col">
            <h4>Google Play | BookingNetwork</h4>
            <p class="footer-text">
              독립 앱 유통 생태계를 지원하는 개방형 앱 스토어입니다.
              구글 플레이 스크래핑을 원천 금지하며 개발사가 서명한 정품 바이너리만을 직접 배포합니다.
            </p>
            <div class="footer-compliance">
              <span class="compliance-badge">🛡️ EU DMA Article 23 Compliant</span>
              <span class="compliance-badge">✓ TreeScan Zero-Day Quarantine</span>
            </div>
          </div>

          <div class="footer-col">
            <h4>카테고리</h4>
            <ul>
              <li><a href="#games" @click="switchTopTab('games')">인기 모바일 게임</a></li>
              <li><a href="#apps" @click="switchTopTab('apps')">생산성 & 비즈니스 앱</a></li>
              <li><a href="#finance" @click="switchTopTab('apps')">금융 & 자산 관리</a></li>
              <li><a href="#security" @click="switchTopTab('protect')">보안 및 프라이버시 툴</a></li>
            </ul>
          </div>

          <div class="footer-col">
            <h4>콘솔 및 개발자</h4>
            <ul>
              <li><a href="/developer/">개발자 등록 및 KYB 인증</a></li>
              <li><a href="/admin/">플랫폼 운영 및 킬스위치 콘솔</a></li>
              <li><a href="/api/v1/system/capabilities" target="_blank">오픈 API 사양서</a></li>
              <li><NuxtLink to="/support/report">악성 패키지 및 취약점 신고</NuxtLink></li>
            </ul>
          </div>

          <div class="footer-col">
            <h4>법적 고지 및 정책</h4>
            <ul>
              <li><a href="#terms">서비스 이용약관</a></li>
              <li><a href="#privacy">개인정보 처리방침</a></li>
              <li><a href="#devpolicy">개발자 유통 계약 및 수수료 정책</a></li>
              <li><a href="#consumer">전자상거래 소비자 보호 안내</a></li>
            </ul>
          </div>
        </div>

        <div class="play-footer-bottom">
          <span>© 2026 Google Play & BookingNetwork Inc. All rights reserved. Sovereign Independent App Market.</span>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const showSmartBanner = ref(true)
const currentTheme = ref<'light' | 'dark'>('dark')
const activeTopTab = ref<'games' | 'apps' | 'charts' | 'editors' | 'protect'>('apps')
const searchQuery = ref('')
const searchFocused = ref(false)
const profileMenuOpen = ref(false)

function downloadCompanionApp() {
  if (typeof window !== 'undefined') {
    const dummyApk = `PK\x03\x04# BookingNetwork Mobile Companion App\nPackage: com.bookingnetwork.appmarket.companion\nVersion: 2.0.0\n`
    const blob = new Blob([dummyApk], { type: 'application/vnd.android.package-archive' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `BookingNetwork-Companion-v2.0.0.apk`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }
}

const quickSuggestions = [
  'Omni Productivity Suite',
  'FinVault Secure Banking',
  'Secure Messenger Pro',
  'Cloud Authenticator 2FA',
  'Cyber Runner 2026'
]

function switchTopTab(tab: 'games' | 'apps' | 'charts' | 'editors' | 'protect') {
  activeTopTab.value = tab
  // Emit or update URL / store section scroll
  const el = document.getElementById(`section-${tab}`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth' })
  } else if (router.currentRoute.value.path !== '/') {
    router.push(`/?tab=${tab}`)
  }
}

function executeSearch() {
  if (searchQuery.value.trim()) {
    router.push(`/search?q=${encodeURIComponent(searchQuery.value.trim())}`)
    searchFocused.value = false
  }
}

function selectSuggestion(item: string) {
  searchQuery.value = item
  executeSearch()
}

function handleSearchBlur() {
  setTimeout(() => {
    searchFocused.value = false
  }, 200)
}

function toggleTheme() {
  currentTheme.value = currentTheme.value === 'dark' ? 'light' : 'dark'
  if (typeof window !== 'undefined') {
    localStorage.setItem('bn_theme', currentTheme.value)
    document.documentElement.setAttribute('data-theme', currentTheme.value)
  }
}

onMounted(() => {
  if (typeof window !== 'undefined') {
    const saved = localStorage.getItem('bn_theme') as 'light' | 'dark' | null
    if (saved) {
      currentTheme.value = saved
    } else {
      currentTheme.value = 'dark'
    }
    document.documentElement.setAttribute('data-theme', currentTheme.value)
  }
})
</script>

<style scoped>
/* ==========================================================================
   Google Play Store & Apple App Store Theme Stylesheet
   ========================================================================== */

.bn-playstore-root {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: var(--bn-bg-canvas);
  color: var(--bn-text-primary);
  font-family: -apple-system, BlinkMacSystemFont, "Google Sans", "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  transition: background-color 0.2s ease, color 0.2s ease;
  -webkit-font-smoothing: antialiased;
}

/* Light / Dark Mode Play Store Palette */
[data-theme="light"] {
  --play-bg: #FFFFFF;
  --play-subtle-bg: #F8F9FA;
  --play-border: #E0E0E0;
  --play-text-primary: #202124;
  --play-text-secondary: #5F6368;
  --play-accent: #01875F; /* Google Play Emerald Green */
  --play-accent-hover: #017250;
  --play-card-bg: #FFFFFF;
  --play-pill-bg: #E8F0FE;
  --play-pill-text: #1A73E8;
  --play-shadow: 0 1px 3px rgba(60,64,67,0.15), 0 1px 2px rgba(60,64,67,0.08);
  --play-hover-shadow: 0 8px 24px rgba(60,64,67,0.15);
}

[data-theme="dark"] {
  --play-bg: #1F1F1F;
  --play-subtle-bg: #131314;
  --play-border: #3C4043;
  --play-text-primary: #E8EAED;
  --play-text-secondary: #9AA0A6;
  --play-accent: #00A870; /* Vibrant Dark Mode Green */
  --play-accent-hover: #00BD7E;
  --play-card-bg: #28292A;
  --play-pill-bg: #004D40;
  --play-pill-text: #80CBC4;
  --play-shadow: 0 2px 6px rgba(0,0,0,0.4);
  --play-hover-shadow: 0 10px 30px rgba(0,0,0,0.6);
}

/* Smart Companion Banner */
.smart-companion-banner {
  background: linear-gradient(90deg, #0d3826 0%, #122820 100%);
  border-bottom: 1px solid rgba(1, 135, 95, 0.4);
  padding: 8px 16px;
  position: relative;
  z-index: 1001;
}

.banner-inner {
  max-width: 1360px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  gap: 12px;
}

.banner-close {
  background: transparent;
  border: none;
  color: #9AA0A6;
  font-size: 14px;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: color 0.15s;
}

.banner-close:hover {
  color: #fff;
}

.banner-icon {
  width: 32px;
  height: 32px;
  background: rgba(1, 135, 95, 0.2);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.banner-text {
  display: flex;
  flex-direction: column;
  flex: 1;
  font-size: 12px;
  line-height: 1.3;
}

.banner-text strong {
  color: #E8EAED;
  font-weight: 600;
  font-size: 13px;
}

.banner-text span {
  color: #9AA0A6;
}

.banner-download-btn {
  background-color: var(--play-accent);
  color: #ffffff;
  border: none;
  font-size: 13px;
  font-weight: 600;
  padding: 6px 14px;
  border-radius: 18px;
  cursor: pointer;
  white-space: nowrap;
  transition: background-color 0.15s, transform 0.1s;
}

.banner-download-btn:hover {
  background-color: var(--play-accent-hover);
  transform: translateY(-1px);
}

/* Header Navbar */
.play-header {
  position: sticky;
  top: 0;
  z-index: 1000;
  background-color: var(--play-bg);
  border-bottom: 1px solid var(--play-border);
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
  backdrop-filter: blur(10px);
}

.play-header-container {
  max-width: 1360px;
  height: 64px;
  margin: 0 auto;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

/* Brand Logo */
.play-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  text-decoration: none;
  flex-shrink: 0;
}

.play-logo-icon {
  display: flex;
  align-items: center;
  justify-content: center;
}

.play-logo-text {
  display: flex;
  flex-direction: column;
}

.brand-name {
  font-size: 19px;
  font-weight: 700;
  color: var(--play-text-primary);
  letter-spacing: -0.4px;
}

.brand-partner {
  font-weight: 400;
  color: var(--play-text-secondary);
  font-size: 15px;
}

.brand-sub {
  font-size: 11px;
  font-weight: 600;
  color: var(--play-accent);
  letter-spacing: 0.5px;
}

/* Top Play Tabs (Google Play Style Center Navigation) */
.play-top-tabs {
  display: flex;
  align-items: center;
  gap: 8px;
}

.top-tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  color: var(--play-text-secondary);
  font-size: 14px;
  font-weight: 600;
  padding: 8px 16px;
  border-radius: 9999px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.top-tab-btn:hover {
  background-color: var(--play-subtle-bg);
  color: var(--play-text-primary);
}

.top-tab-btn.active {
  background-color: var(--play-pill-bg);
  color: var(--play-pill-text);
  font-weight: 700;
}

/* Search Bar (Capsule Google Play Style) */
.play-search-box {
  position: relative;
  display: flex;
  align-items: center;
  background-color: var(--play-subtle-bg);
  border: 1px solid var(--play-border);
  border-radius: 9999px;
  padding: 0 16px;
  height: 44px;
  flex: 1;
  max-width: 380px;
  transition: all 0.2s ease;
}

.play-search-box.focused {
  background-color: var(--play-bg);
  border-color: var(--play-accent);
  box-shadow: 0 0 0 3px rgba(1, 135, 95, 0.2);
}

.search-svg {
  color: var(--play-text-secondary);
  margin-right: 10px;
  flex-shrink: 0;
}

.search-input {
  width: 100%;
  background: transparent;
  border: none;
  outline: none;
  font-size: 14px;
  color: var(--play-text-primary);
}

.search-clear-btn {
  background: none;
  border: none;
  color: var(--play-text-secondary);
  cursor: pointer;
  font-size: 13px;
  padding: 4px;
}

.search-suggestions-dropdown {
  position: absolute;
  top: 50px;
  left: 0;
  right: 0;
  background-color: var(--play-bg);
  border: 1px solid var(--play-border);
  border-radius: 16px;
  box-shadow: var(--play-hover-shadow);
  padding: 12px 0;
  z-index: 1010;
}

.suggestions-label {
  padding: 4px 18px 8px;
  font-size: 11px;
  font-weight: 700;
  color: var(--play-text-secondary);
  text-transform: uppercase;
}

.suggestion-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 18px;
  font-size: 14px;
  color: var(--play-text-primary);
  cursor: pointer;
  transition: background-color 0.1s ease;
}

.suggestion-item:hover {
  background-color: var(--play-subtle-bg);
}

/* Header Action Buttons */
.play-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  position: relative;
}

.icon-circle-btn {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  border: 1px solid var(--play-border);
  background-color: var(--play-subtle-bg);
  color: var(--play-text-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  cursor: pointer;
  text-decoration: none;
  transition: all 0.15s ease;
}

.icon-circle-btn:hover {
  background-color: var(--play-border);
}

/* User Avatar & Popover (Google Play Account Style) */
.user-avatar-btn {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  border: 2px solid var(--play-accent);
  background: linear-gradient(135deg, #0084FF, #00E575);
  color: #FFFFFF;
  font-size: 13px;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(0,0,0,0.15);
  transition: transform 0.15s ease;
}

.user-avatar-btn:hover {
  transform: scale(1.05);
}

.user-profile-popover {
  position: absolute;
  top: 50px;
  right: 0;
  width: 340px;
  background-color: var(--play-bg);
  border: 1px solid var(--play-border);
  border-radius: 20px;
  box-shadow: var(--play-hover-shadow);
  padding: 20px;
  z-index: 1050;
  display: flex;
  flex-direction: column;
}

.popover-header {
  display: flex;
  align-items: center;
  gap: 14px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--play-border);
}

.popover-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, #0084FF, #00E575);
  color: white;
  font-size: 16px;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.popover-name {
  font-size: 15px;
  font-weight: 700;
  color: var(--play-text-primary);
}

.popover-email {
  font-size: 13px;
  color: var(--play-text-secondary);
}

.popover-badge {
  display: inline-block;
  margin-top: 4px;
  padding: 2px 8px;
  background-color: rgba(1, 135, 95, 0.15);
  color: var(--play-accent);
  font-size: 10px;
  font-weight: 700;
  border-radius: 9999px;
}

.popover-section {
  display: flex;
  flex-direction: column;
  padding: 12px 0;
  gap: 4px;
}

.popover-section-label {
  font-size: 11px;
  font-weight: 700;
  color: var(--play-text-secondary);
  text-transform: uppercase;
  margin-bottom: 6px;
  padding: 0 8px;
}

.popover-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 10px;
  color: var(--play-text-primary);
  text-decoration: none;
  font-size: 13px;
  font-weight: 600;
  transition: background-color 0.15s ease;
}

.popover-item:hover {
  background-color: var(--play-subtle-bg);
}

.popover-link-text {
  display: flex;
  flex-direction: column;
}

.popover-link-text strong {
  font-size: 13px;
}

.popover-link-text small {
  font-size: 11px;
  color: var(--play-text-secondary);
  font-weight: 400;
}

.popover-divider {
  height: 1px;
  background-color: var(--play-border);
  margin: 4px 0;
}

.popover-footer {
  padding-top: 12px;
  display: flex;
  justify-content: flex-end;
}

.popover-signout-btn {
  background: transparent;
  border: 1px solid var(--play-border);
  border-radius: 8px;
  color: var(--play-text-secondary);
  padding: 6px 14px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

.popover-signout-btn:hover {
  background-color: var(--play-subtle-bg);
  color: var(--play-text-primary);
}

/* Page Body */
.play-content-area {
  flex: 1;
}

/* Footer (Clean Google Style) */
.play-footer {
  background-color: var(--play-subtle-bg);
  border-top: 1px solid var(--play-border);
  padding: 48px 24px 24px;
  margin-top: 60px;
}

.play-footer-inner {
  max-width: 1360px;
  margin: 0 auto;
}

.play-footer-nav {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1.2fr;
  gap: 40px;
  margin-bottom: 36px;
}

.footer-col h4 {
  font-size: 14px;
  font-weight: 700;
  color: var(--play-text-primary);
  margin-bottom: 16px;
}

.footer-text {
  font-size: 13px;
  line-height: 1.6;
  color: var(--play-text-secondary);
  margin-bottom: 16px;
  max-width: 420px;
}

.footer-compliance {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.compliance-badge {
  font-size: 11px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 9999px;
  background-color: rgba(1, 135, 95, 0.12);
  color: var(--play-accent);
  border: 1px solid rgba(1, 135, 95, 0.3);
}

.footer-col ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.footer-col a {
  color: var(--play-text-secondary);
  text-decoration: none;
  font-size: 13px;
  transition: color 0.15s ease;
}

.footer-col a:hover {
  color: var(--play-accent);
}

.play-footer-bottom {
  padding-top: 24px;
  border-top: 1px solid var(--play-border);
  font-size: 12px;
  color: var(--play-text-secondary);
  text-align: center;
}

/* Responsive adjustments */
@media (max-width: 1024px) {
  .play-top-tabs {
    display: none;
  }
  .play-footer-nav {
    grid-template-columns: 1fr 1fr;
  }
}

@media (max-width: 640px) {
  .play-header-container {
    padding: 0 16px;
  }
  .brand-sub, .brand-partner {
    display: none;
  }
  .play-footer-nav {
    grid-template-columns: 1fr;
  }
}
</style>

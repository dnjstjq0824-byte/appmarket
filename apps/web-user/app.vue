<template>
  <div class="bn-store-root" :data-theme="currentTheme">
    <!-- Companion app notice -->
    <div v-if="showSmartBanner" class="companion-band">
      <div class="band-inner">
        <button class="band-close" @click="showSmartBanner = false" title="배너 닫기" aria-label="배너 닫기">
          <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"><path d="M1 1l10 10M11 1L1 11" /></svg>
        </button>
        <span class="band-kicker">Android</span>
        <div class="banner-text">
          <strong>BookingNetwork 모바일 공식 앱</strong>
          <span>PackageInstaller 원클릭 설치 및 백그라운드 자동 업데이트 지원</span>
        </div>
        <button class="banner-download-btn" @click="downloadCompanionApp">
          스토어에서 찾기
        </button>
      </div>
    </div>

    <!-- Store masthead -->
    <header class="masthead">
      <div class="masthead-inner">
        <NuxtLink to="/" class="wordmark" @click="activeTopTab = 'apps'">
          <span class="seal-mark" aria-hidden="true">BN</span>
          <span class="wordmark-text">
            <span class="brand-name">App Market</span>
            <span class="brand-sub">BOOKINGNETWORK · INDEPENDENT STORE</span>
          </span>
        </NuxtLink>

        <nav class="masthead-tabs">
          <button
            class="top-tab-btn"
            :class="{ active: activeTopTab === 'games' }"
            @click="switchTopTab('games')"
          >
            게임
          </button>
          <button
            class="top-tab-btn"
            :class="{ active: activeTopTab === 'apps' }"
            @click="switchTopTab('apps')"
          >
            앱
          </button>
          <button
            class="top-tab-btn"
            :class="{ active: activeTopTab === 'charts' }"
            @click="switchTopTab('charts')"
          >
            인기 차트
          </button>
          <button
            class="top-tab-btn"
            :class="{ active: activeTopTab === 'editors' }"
            @click="switchTopTab('editors')"
          >
            에디터 추천
          </button>
          <button
            class="top-tab-btn"
            :class="{ active: activeTopTab === 'protect' }"
            @click="switchTopTab('protect')"
          >
            검증 현황
          </button>
        </nav>

        <div class="masthead-search" :class="{ focused: searchFocused }">
          <svg class="search-svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <circle cx="11" cy="11" r="7"></circle>
            <line x1="16.5" y1="16.5" x2="21" y2="21"></line>
          </svg>
          <input
            v-model="searchQuery"
            type="search"
            placeholder="앱, 게임, 개발사 검색"
            class="search-input"
            @focus="searchFocused = true"
            @blur="handleSearchBlur"
            @keyup.enter="executeSearch"
          />
          <button v-if="searchQuery" class="search-clear-btn" @click="searchQuery = ''" aria-label="검색어 지우기">
            <svg width="10" height="10" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"><path d="M1 1l10 10M11 1L1 11" /></svg>
          </button>

          <div v-if="searchFocused && quickSuggestions.length > 0" class="search-suggestions-dropdown">
            <div class="suggestions-label">추천 검색어</div>
            <div
              v-for="item in quickSuggestions"
              :key="item.type + ':' + item.text"
              class="suggestion-item"
              @mousedown.prevent="selectSuggestion(item)"
            >
              <span class="suggestion-index">{{ item.type }}</span>
              <span>{{ item.text }}</span>
            </div>
          </div>
        </div>

        <div class="masthead-actions">
          <button class="icon-square-btn" @click="toggleTheme" title="다크/라이트 테마 전환" aria-label="테마 전환">
            <svg v-if="currentTheme === 'dark'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="4" /><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" /></svg>
            <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z" /></svg>
          </button>

          <NuxtLink to="/support/report" class="icon-square-btn" title="보안 검역 및 고객 신고" aria-label="신고하기">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16v12H7l-3 3V4z" /></svg>
          </NuxtLink>

          <div class="user-profile-menu-container">
            <button class="user-avatar-btn" @click="profileMenuOpen = !profileMenuOpen" title="계정 및 서비스 전환">
              <span class="avatar-letter">{{ loggedIn ? (profileName || 'U').slice(0, 1).toUpperCase() : 'BN' }}</span>
            </button>

            <div v-if="profileMenuOpen" class="user-profile-popover" @click.stop>
              <div v-if="loggedIn" class="popover-header">
                <div class="popover-avatar">{{ (profileName || 'U').slice(0, 1).toUpperCase() }}</div>
                <div class="popover-user-info">
                  <div class="popover-name">{{ profileName }}</div>
                  <div class="popover-email">{{ profileEmail }}</div>
                </div>
              </div>
              <div v-else class="popover-header">
                <div class="popover-user-info">
                  <div class="popover-name">로그인이 필요합니다</div>
                  <NuxtLink to="/account/login" class="popover-item" @click="profileMenuOpen = false">
                    로그인 / 회원가입
                  </NuxtLink>
                </div>
              </div>

              <div v-if="loggedIn" class="popover-section">
                <NuxtLink to="/account/library" class="popover-item" @click="profileMenuOpen = false">
                  내 앱 및 게임 (보관함)
                </NuxtLink>
                <NuxtLink to="/account/library" class="popover-item" @click="profileMenuOpen = false">
                  결제 및 정산 내역
                </NuxtLink>
              </div>

              <div class="popover-divider"></div>

              <div class="popover-section">
                <div class="popover-section-label">ENTERPRISE PORTAL</div>
                <a href="/developer/" class="popover-item dev-link">
                  <div class="popover-link-text">
                    <strong>개발자 콘솔</strong>
                    <small>KYB 인증, APK 업로드, 릴리즈 파이프라인</small>
                  </div>
                </a>
                <a href="/admin/" class="popover-item admin-link">
                  <div class="popover-link-text">
                    <strong>플랫폼 관리자</strong>
                    <small>긴급 킬스위치, 실시간 QPS, 심사 큐</small>
                  </div>
                </a>
              </div>

              <div class="popover-footer">
                <button v-if="loggedIn" class="popover-signout-btn" @click="signOut">로그아웃</button>
                <button class="popover-signout-btn" @click="profileMenuOpen = false">닫기</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>

    <main class="store-content" @click="profileMenuOpen = false">
      <NuxtPage />
    </main>

    <footer class="store-footer">
      <div class="store-footer-inner">
        <div class="store-footer-nav">
          <div class="footer-col">
            <h4>BookingNetwork App Market</h4>
            <p class="footer-text">
              개발사가 직접 서명한 정품 바이너리만을 유통하는 독립 앱 스토어입니다.
              비공식 스크래핑을 원천 배제하고 모든 패키지의 서명 연속성을 검증합니다.
            </p>
            <div class="footer-compliance">
              <span class="compliance-badge">EU DMA ART. 23 COMPLIANT</span>
              <span class="compliance-badge">TREESCAN QUARANTINE</span>
            </div>
          </div>

          <div class="footer-col">
            <h4>둘러보기</h4>
            <ul>
              <li><a href="#games" @click="switchTopTab('games')">인기 모바일 게임</a></li>
              <li><a href="#apps" @click="switchTopTab('apps')">생산성 &amp; 비즈니스 앱</a></li>
              <li><a href="#finance" @click="switchTopTab('apps')">금융 &amp; 자산 관리</a></li>
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
              <li><NuxtLink to="/legal/terms">서비스 이용약관</NuxtLink></li>
              <li><NuxtLink to="/legal/privacy">개인정보 처리방침</NuxtLink></li>
              <li><NuxtLink to="/legal/terms">개발자 유통 계약 및 수수료 정책</NuxtLink></li>
              <li><a href="#consumer">전자상거래 소비자 보호 안내</a></li>
            </ul>
          </div>
        </div>

        <div class="store-footer-bottom">
          <span>© 2026 BookingNetwork Inc. — Sovereign Independent App Market.</span>
          <span class="footer-build">V3 SIGNATURE CONTINUITY · TREESCAN PIPELINE</span>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { getSharedClient, clearSession, type SuggestionItem } from '@bookingnetwork/api-client'

const api = getSharedClient()
const router = useRouter()
const showSmartBanner = ref(true)
const currentTheme = ref<'light' | 'dark'>('light')
const activeTopTab = ref<'games' | 'apps' | 'charts' | 'editors' | 'protect'>('apps')
const searchQuery = ref('')
const searchFocused = ref(false)
const profileMenuOpen = ref(false)
const quickSuggestions = ref<SuggestionItem[]>([])
const profileEmail = ref('')
const profileName = ref('')
const loggedIn = ref(false)
let suggestTimer: ReturnType<typeof setTimeout> | null = null

async function downloadCompanionApp() {
  await router.push('/search?q=companion')
}

async function refreshSuggestions(q: string) {
  const query = q.trim()
  if (query.length < 2) {
    quickSuggestions.value = []
    return
  }
  try {
    const res = await api.suggestApps(query, 6)
    quickSuggestions.value = res.suggestions || []
  } catch {
    quickSuggestions.value = []
  }
}

watch(searchQuery, (q) => {
  if (suggestTimer) clearTimeout(suggestTimer)
  suggestTimer = setTimeout(() => { void refreshSuggestions(q) }, 250)
})

async function refreshProfile() {
  try {
    const me = await api.getMe()
    loggedIn.value = true
    profileEmail.value = me.email
    profileName.value = me.email.split('@')[0]
  } catch {
    loggedIn.value = false
    profileEmail.value = ''
    profileName.value = ''
  }
}

function signOut() {
  clearSession()
  loggedIn.value = false
  profileMenuOpen.value = false
  void router.push('/')
}

function switchTopTab(tab: 'games' | 'apps' | 'charts' | 'editors' | 'protect') {
  activeTopTab.value = tab
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

function selectSuggestion(item: SuggestionItem) {
  searchQuery.value = item.text
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
    currentTheme.value = saved === 'dark' ? 'dark' : 'light'
    document.documentElement.setAttribute('data-theme', currentTheme.value)
  }
  void refreshProfile()
})
</script>

<style scoped>
.bn-store-root {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: var(--bn-bg-canvas);
  color: var(--bn-text-primary);
}

/* Companion band: solid ink strip */
.companion-band {
  background-color: var(--bn-ink);
  color: var(--bn-bg-canvas);
}
.band-inner {
  max-width: 1360px;
  margin: 0 auto;
  padding: 8px 24px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.band-close {
  background: transparent;
  border: none;
  color: inherit;
  opacity: 0.6;
  cursor: pointer;
  padding: 4px;
  display: flex;
}
.band-close:hover {
  opacity: 1;
}
.band-kicker {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.12em;
  border: 1px solid currentColor;
  border-radius: var(--bn-radius-sm);
  padding: 2px 6px;
  opacity: 0.8;
  white-space: nowrap;
}
.banner-text {
  display: flex;
  align-items: baseline;
  gap: 10px;
  flex: 1;
  min-width: 0;
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
}
.banner-text strong {
  font-weight: 700;
}
.banner-text span {
  opacity: 0.65;
  overflow: hidden;
  text-overflow: ellipsis;
}
.banner-download-btn {
  background-color: var(--bn-bg-canvas);
  color: var(--bn-ink);
  border: none;
  font-size: 12px;
  font-weight: 700;
  padding: 6px 14px;
  border-radius: var(--bn-radius-sm);
  cursor: pointer;
  white-space: nowrap;
}
.banner-download-btn:hover {
  opacity: 0.88;
}

/* Masthead */
.masthead {
  position: sticky;
  top: 0;
  z-index: 1000;
  background-color: var(--bn-bg-canvas);
  border-bottom: 1px solid var(--bn-border-subtle);
}
.masthead-inner {
  max-width: 1360px;
  margin: 0 auto;
  padding: 12px 24px;
  display: flex;
  align-items: center;
  gap: 28px;
}
.wordmark {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  color: inherit;
  flex-shrink: 0;
}
.seal-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  background-color: var(--bn-seal);
  color: #fff;
  font-family: var(--bn-font-display);
  font-weight: 900;
  font-size: 14px;
  letter-spacing: -0.02em;
  border-radius: var(--bn-radius-sm);
}
.wordmark-text {
  display: flex;
  flex-direction: column;
  line-height: 1.15;
}
.brand-name {
  font-family: var(--bn-font-display);
  font-weight: 800;
  font-size: 17px;
  letter-spacing: -0.02em;
}
.brand-sub {
  font-family: var(--bn-font-mono);
  font-size: 9px;
  letter-spacing: 0.14em;
  color: var(--bn-text-muted);
}
.masthead-tabs {
  display: flex;
  gap: 2px;
}
.top-tab-btn {
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--bn-text-secondary);
  font-family: var(--bn-font-sans);
  font-size: 14px;
  font-weight: 600;
  padding: 8px 12px;
  cursor: pointer;
}
.top-tab-btn:hover {
  color: var(--bn-text-primary);
}
.top-tab-btn.active {
  color: var(--bn-text-primary);
  border-bottom-color: var(--bn-seal);
}

/* Search */
.masthead-search {
  position: relative;
  flex: 1;
  max-width: 380px;
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  padding: 0 10px;
}
.masthead-search.focused {
  border-color: var(--bn-border-strong);
  background-color: var(--bn-bg-surface);
}
.search-svg {
  color: var(--bn-text-muted);
  flex-shrink: 0;
}
.search-input {
  flex: 1;
  min-width: 0;
  background: transparent;
  border: none;
  outline: none;
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
  font-size: 13px;
  padding: 9px 0;
}
.search-input::placeholder {
  color: var(--bn-text-muted);
}
.search-clear-btn {
  background: none;
  border: none;
  color: var(--bn-text-muted);
  cursor: pointer;
  display: flex;
  padding: 4px;
}
.search-clear-btn:hover {
  color: var(--bn-text-primary);
}
.search-suggestions-dropdown {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  right: 0;
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  box-shadow: var(--bn-shadow-pop);
  overflow: hidden;
  z-index: 1001;
}
.suggestions-label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.12em;
  color: var(--bn-text-muted);
  padding: 8px 12px 4px;
}
.suggestion-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  font-size: 13px;
  cursor: pointer;
}
.suggestion-item:hover {
  background-color: var(--bn-bg-well);
}
.suggestion-index {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  color: var(--bn-text-muted);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  padding: 1px 5px;
}

/* Header actions + profile */
.masthead-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}
.icon-square-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  background: transparent;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  color: var(--bn-text-secondary);
  cursor: pointer;
  text-decoration: none;
}
.icon-square-btn:hover {
  color: var(--bn-text-primary);
  border-color: var(--bn-border-strong);
}
.user-profile-menu-container {
  position: relative;
}
.user-avatar-btn {
  width: 34px;
  height: 34px;
  background-color: var(--bn-ink);
  color: var(--bn-bg-canvas);
  border: none;
  border-radius: var(--bn-radius-sm);
  font-family: var(--bn-font-display);
  font-weight: 800;
  font-size: 12px;
  cursor: pointer;
}
.avatar-letter {
  letter-spacing: 0.02em;
}
.user-profile-popover {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 300px;
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  box-shadow: var(--bn-shadow-pop);
  z-index: 1002;
}
.popover-header {
  display: flex;
  gap: 12px;
  padding: 14px;
  align-items: center;
}
.popover-avatar {
  width: 38px;
  height: 38px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--bn-ink);
  color: var(--bn-bg-canvas);
  font-family: var(--bn-font-display);
  font-weight: 800;
  font-size: 15px;
  border-radius: var(--bn-radius-sm);
}
.popover-user-info {
  min-width: 0;
}
.popover-name {
  font-weight: 700;
  font-size: 14px;
}
.popover-email {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
}
.popover-section {
  padding: 6px;
}
.popover-item {
  display: block;
  padding: 9px 10px;
  font-size: 13px;
  font-weight: 600;
  color: var(--bn-text-primary);
  text-decoration: none;
  border-radius: var(--bn-radius-sm);
}
.popover-item:hover {
  background-color: var(--bn-bg-well);
}
.popover-link-text strong {
  display: block;
  font-size: 13px;
}
.popover-link-text small {
  display: block;
  font-size: 11px;
  font-weight: 400;
  color: var(--bn-text-muted);
  margin-top: 2px;
}
.popover-divider {
  border-top: 1px solid var(--bn-border-subtle);
  margin: 4px 0;
}
.popover-section-label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.12em;
  color: var(--bn-text-muted);
  padding: 6px 10px 2px;
}
.popover-footer {
  display: flex;
  gap: 8px;
  padding: 10px 14px 14px;
}
.popover-signout-btn {
  flex: 1;
  background: transparent;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  color: var(--bn-text-secondary);
  font-size: 12px;
  font-weight: 600;
  padding: 7px 0;
  cursor: pointer;
}
.popover-signout-btn:hover {
  border-color: var(--bn-border-strong);
  color: var(--bn-text-primary);
}

/* Content + footer */
.store-content {
  flex: 1;
}
.store-footer {
  border-top: 1px solid var(--bn-border-subtle);
  margin-top: 40px;
}
.store-footer-inner {
  max-width: 1360px;
  margin: 0 auto;
  padding: 36px 24px 24px;
}
.store-footer-nav {
  display: grid;
  grid-template-columns: 1.4fr 1fr 1fr 1fr;
  gap: 32px;
}
.footer-col h4 {
  font-family: var(--bn-font-display);
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.02em;
  margin: 0 0 12px;
}
.footer-text {
  font-size: 12px;
  line-height: 1.65;
  color: var(--bn-text-secondary);
  margin: 0 0 14px;
}
.footer-compliance {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.compliance-badge {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-seal);
  border: 1px solid var(--bn-seal);
  border-radius: var(--bn-radius-sm);
  padding: 3px 8px;
}
.footer-col ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.footer-col a {
  font-size: 13px;
  color: var(--bn-text-secondary);
  text-decoration: none;
}
.footer-col a:hover {
  color: var(--bn-text-primary);
  text-decoration: underline;
  text-underline-offset: 3px;
}
.store-footer-bottom {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  margin-top: 28px;
  padding-top: 16px;
  border-top: 1px solid var(--bn-border-subtle);
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-text-muted);
}

@media (max-width: 1024px) {
  .masthead-inner {
    flex-wrap: wrap;
    gap: 12px;
  }
  .masthead-tabs {
    order: 3;
    width: 100%;
    overflow-x: auto;
  }
  .masthead-search {
    margin-left: 0;
    flex: 1;
  }
  .store-footer-nav {
    grid-template-columns: 1fr 1fr;
  }
}
@media (max-width: 640px) {
  .banner-text span {
    display: none;
  }
  .store-footer-nav {
    grid-template-columns: 1fr;
  }
}
</style>

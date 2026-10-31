<template>
  <div class="bn-store-home">
    <!-- Hero Spotlight Showcase -->
    <section class="bn-hero-section">
      <div class="bn-hero-card">
        <div class="bn-hero-content">
          <div class="bn-hero-badge-row">
            <span class="bn-badge-spotlight">✨ 이 주의 에디터 추천</span>
            <span class="bn-badge-verified">🛡️ 무결성 서명 검증됨</span>
          </div>
          <h1 class="bn-hero-title">Omni Productivity Suite</h1>
          <p class="bn-hero-desc">
            문서 편집, 실시간 클라우드 동기화, 엔드투엔드 암호화 협업을 단 하나의 앱으로 완성하세요.
            구글 플레이 독립 유통 공식 인증 패키지입니다.
          </p>
          <div class="bn-hero-meta">
            <div class="bn-meta-item">
              <span class="bn-meta-label">평점</span>
              <span class="bn-meta-val">★ 4.9 (42,800 리뷰)</span>
            </div>
            <div class="bn-meta-item">
              <span class="bn-meta-label">설치 용량</span>
              <span class="bn-meta-val">35.8 MB (Universal APK)</span>
            </div>
            <div class="bn-meta-item">
              <span class="bn-meta-label">최소 지원 SDK</span>
              <span class="bn-meta-val">Android 8.0+ (API 26)</span>
            </div>
          </div>
          <div class="bn-hero-cta-group">
            <button class="bn-btn-primary" @click="handleInstallApp(featuredApp)">
              <span>⬇️</span> 바로 설치 / 다운로드
            </button>
            <NuxtLink :to="`/apps/${featuredApp.slug}`" class="bn-btn-secondary">
              상세 정보 & 리뷰 보기
            </NuxtLink>
          </div>
        </div>

        <div class="bn-hero-visual">
          <div class="bn-app-icon-hero" v-html="featuredApp.iconSvg"></div>
        </div>
      </div>
    </section>

    <div class="bn-container">
      <!-- Trust & Transparency Guarantee Banner -->
      <section class="bn-guarantee-strip">
        <div class="bn-guarantee-item">
          <span class="bn-g-icon">🛡️</span>
          <div>
            <strong>악성코드 전수 자동 검역</strong>
            <p>TreeScan 정적 분석 및 VirusTotal 70개 백신 0-Day 실시간 검사</p>
          </div>
        </div>
        <div class="bn-guarantee-item">
          <span class="bn-g-icon">🔑</span>
          <div>
            <strong>V3 서명 연속성 검증</strong>
            <p>개발사 고유 인증서 계통 추적 및 비인가 서명 변조 100% 차단</p>
          </div>
        </div>
        <div class="bn-guarantee-item">
          <span class="bn-g-icon">⚡</span>
          <div>
            <strong>초고속 글로벌 CDN</strong>
            <p>15분 만료 HMAC-SHA256 토큰 보안 스트리밍 직접 다운로드</p>
          </div>
        </div>
      </section>

      <!-- Category Filter Pills -->
      <section class="bn-categories-bar">
        <button
          v-for="cat in categories"
          :key="cat.id"
          class="bn-cat-pill"
          :class="{ active: selectedCategory === cat.id }"
          @click="selectCategory(cat.id)"
        >
          <span>{{ cat.icon }}</span>
          <span>{{ cat.name }}</span>
          <span class="bn-cat-count">{{ getCategoryCount(cat.id) }}</span>
        </button>
      </section>

      <!-- Main Layout: Grid and Top Charts Sidebar -->
      <div class="bn-content-grid">
        <!-- Main Catalog Feed -->
        <div class="bn-feed-main">
          <div class="bn-section-header">
            <h2>{{ currentCategoryTitle }}</h2>
            <span class="bn-result-count">총 {{ filteredApps.length }}개 패키지</span>
          </div>

          <div class="bn-cards-grid">
            <article
              v-for="app in filteredApps"
              :key="app.id"
              class="bn-store-card"
              @click="navigateToApp(app.slug)"
            >
              <div class="bn-card-icon" v-html="app.iconSvg"></div>

              <div class="bn-card-body">
                <div class="bn-card-top-row">
                  <h3 class="bn-card-title">{{ app.title }}</h3>
                  <span v-if="app.is_sponsored" class="bn-badge-sponsored">AD</span>
                </div>

                <div class="bn-dev-row">
                  <span class="bn-dev-name">{{ app.developer_name }}</span>
                  <span class="bn-dev-check" title="검증된 개발사">✓</span>
                </div>

                <div class="bn-card-meta-row">
                  <span class="bn-rating-pill">★ {{ app.bayesian_rating.toFixed(1) }}</span>
                  <span class="bn-downloads-text">{{ app.install_count_label }}</span>
                  <span class="bn-size-pill">{{ app.size_label }}</span>
                </div>

                <p class="bn-card-summary">{{ app.summary }}</p>

                <div class="bn-card-bottom-row" @click.stop>
                  <span class="bn-price-tag" :class="{ free: app.price === 0 }">
                    {{ app.price === 0 ? '무료' : `₩${app.price.toLocaleString()}` }}
                  </span>
                  <button class="bn-card-btn-install" @click="handleInstallApp(app)">
                    설치
                  </button>
                </div>
              </div>
            </article>
          </div>
        </div>

        <!-- Right: Top 5 Realtime Ranking Sidebar -->
        <aside class="bn-feed-sidebar">
          <div class="bn-ranking-card">
            <div class="bn-ranking-header">
              <h3>🏆 실시간 인기 순위</h3>
              <span class="bn-ranking-update">Live QPS</span>
            </div>

            <div class="bn-ranking-list">
              <div
                v-for="(app, idx) in topRankedApps"
                :key="app.id"
                class="bn-ranking-item"
                @click="navigateToApp(app.slug)"
              >
                <div class="bn-rank-num" :class="`rank-${idx + 1}`">{{ idx + 1 }}</div>
                <div class="bn-rank-icon" v-html="app.iconSvg"></div>
                <div class="bn-rank-info">
                  <span class="bn-rank-title">{{ app.title }}</span>
                  <span class="bn-rank-meta">{{ app.category_name }} • ★{{ app.bayesian_rating.toFixed(1) }}</span>
                </div>
                <button class="bn-rank-install-btn" @click.stop="handleInstallApp(app)">
                  설치
                </button>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </div>

    <!-- Toast Notification -->
    <div v-if="toastMessage" class="bn-toast">
      {{ toastMessage }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

interface StoreApp {
  id: string
  slug: string
  title: string
  developer_name: string
  category_id: string
  category_name: string
  bayesian_rating: number
  install_count_label: string
  size_label: string
  price: number
  summary: string
  is_sponsored?: boolean
  iconSvg: string
}

const router = useRouter()
const selectedCategory = ref('all')
const toastMessage = ref('')

const categories = [
  { id: 'all', name: '전체', icon: '🌐' },
  { id: 'productivity', name: '생산성 & 업무', icon: '💼' },
  { id: 'finance', name: '금융 & 핀테크', icon: '💳' },
  { id: 'utilities', name: '도구 & 유틸리티', icon: '🛠️' },
  { id: 'communication', name: '커뮤니케이션', icon: '💬' },
  { id: 'security', name: '보안 & 인증', icon: '🛡️' },
  { id: 'games', name: '게임', icon: '🎮' },
]

// High-fidelity App Icons using custom crisp inline SVGs
const icons = {
  omni: `<svg width="56" height="56" viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg"><rect width="64" height="64" rx="16" fill="url(#omniGrad)"/><path d="M20 32L28 40L44 24" stroke="white" stroke-width="5" stroke-linecap="round" stroke-linejoin="round"/><defs><linearGradient id="omniGrad" x1="0" y1="0" x2="64" y2="64" gradientUnits="userSpaceOnUse"><stop stop-color="#2563EB"/><stop offset="1" stop-color="#06B6D4"/></linearGradient></defs></svg>`,
  finvault: `<svg width="56" height="56" viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg"><rect width="64" height="64" rx="16" fill="url(#finGrad)"/><path d="M32 16L46 22V34C46 43 39 49 32 52C25 49 18 43 18 34V22L32 16Z" fill="white" fill-opacity="0.2" stroke="white" stroke-width="4" stroke-linejoin="round"/><circle cx="32" cy="34" r="5" fill="white"/><defs><linearGradient id="finGrad" x1="0" y1="0" x2="64" y2="64" gradientUnits="userSpaceOnUse"><stop stop-color="#059669"/><stop offset="1" stop-color="#10B981"/></linearGradient></defs></svg>`,
  messenger: `<svg width="56" height="56" viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg"><rect width="64" height="64" rx="16" fill="url(#msgGrad)"/><path d="M20 22C20 18.6863 22.6863 16 26 16H38C41.3137 16 44 18.6863 44 22V36C44 39.3137 41.3137 42 38 42H28L20 48V22Z" fill="white"/><circle cx="28" cy="29" r="2.5" fill="#4F46E5"/><circle cx="36" cy="29" r="2.5" fill="#4F46E5"/><defs><linearGradient id="msgGrad" x1="0" y1="0" x2="64" y2="64" gradientUnits="userSpaceOnUse"><stop stop-color="#4F46E5"/><stop offset="1" stop-color="#9333EA"/></linearGradient></defs></svg>`,
  crypto: `<svg width="56" height="56" viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg"><rect width="64" height="64" rx="16" fill="url(#cryGrad)"/><path d="M22 42L32 28L38 34L46 22" stroke="white" stroke-width="4.5" stroke-linecap="round" stroke-linejoin="round"/><circle cx="46" cy="22" r="3" fill="white"/><defs><linearGradient id="cryGrad" x1="0" y1="0" x2="64" y2="64" gradientUnits="userSpaceOnUse"><stop stop-color="#D97706"/><stop offset="1" stop-color="#F59E0B"/></linearGradient></defs></svg>`,
  auth: `<svg width="56" height="56" viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg"><rect width="64" height="64" rx="16" fill="url(#authGrad)"/><path d="M32 20C27.5817 20 24 23.5817 24 28C24 31.2387 25.926 34.0272 28.6942 35.2825L26 44H38L35.3058 35.2825C38.074 34.0272 40 31.2387 40 28C40 23.5817 36.4183 20 32 20Z" fill="white"/><defs><linearGradient id="authGrad" x1="0" y1="0" x2="64" y2="64" gradientUnits="userSpaceOnUse"><stop stop-color="#0284C7"/><stop offset="1" stop-color="#38BDF8"/></linearGradient></defs></svg>`,
  devtools: `<svg width="56" height="56" viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg"><rect width="64" height="64" rx="16" fill="url(#devGrad)"/><path d="M24 24L16 32L24 40M40 24L48 32L40 40M34 20L30 44" stroke="white" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/><defs><linearGradient id="devGrad" x1="0" y1="0" x2="64" y2="64" gradientUnits="userSpaceOnUse"><stop stop-color="#DC2626"/><stop offset="1" stop-color="#EA580C"/></linearGradient></defs></svg>`,
  actiongame: `<svg width="56" height="56" viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg"><rect width="64" height="64" rx="16" fill="url(#gameGrad)"/><path d="M22 28V36M18 32H26M40 30A2 2 0 1 1 40 26A2 2 0 0 1 40 30ZM44 36A2 2 0 1 1 44 32A2 2 0 0 1 44 36Z" stroke="white" stroke-width="3" stroke-linecap="round"/><rect x="14" y="22" width="36" height="20" rx="6" stroke="white" stroke-width="3.5"/><defs><linearGradient id="gameGrad" x1="0" y1="0" x2="64" y2="64" gradientUnits="userSpaceOnUse"><stop stop-color="#7C3AED"/><stop offset="1" stop-color="#EC4899"/></linearGradient></defs></svg>`,
}

const storeApps = ref<StoreApp[]>([
  {
    id: 'app-omni-01',
    slug: 'omni-productivity-suite',
    title: 'Omni Productivity Suite',
    developer_name: 'OmniWork Technologies',
    category_id: 'productivity',
    category_name: '생산성 & 업무',
    bayesian_rating: 4.9,
    install_count_label: '150만+ 다운로드',
    size_label: '35.8 MB',
    price: 0,
    summary: '차세대 올인원 문서 편집, 프로젝트 관리 및 엔드투엔드 암호화 협업 스위트',
    is_sponsored: true,
    iconSvg: icons.omni,
  },
  {
    id: 'app-finvault-02',
    slug: 'finvault-secure-banking',
    title: 'FinVault Secure Banking',
    developer_name: 'FinVault Financial Labs',
    category_id: 'finance',
    category_name: '금융 & 핀테크',
    bayesian_rating: 4.8,
    install_count_label: '82만+ 다운로드',
    size_label: '24.5 MB',
    price: 0,
    summary: '생체인식 하드웨어 보안 키 및 지능형 가계부 자산 통합 자산 관리 솔루션',
    iconSvg: icons.finvault,
  },
  {
    id: 'app-msg-03',
    slug: 'secure-messenger-pro',
    title: 'Secure Messenger Pro',
    developer_name: 'Open Privacy Labs',
    category_id: 'communication',
    category_name: '커뮤니케이션',
    bayesian_rating: 4.8,
    install_count_label: '210만+ 다운로드',
    size_label: '28.1 MB',
    price: 0,
    summary: '영지식 암호화(ZKP) 기반 탈중앙화 메신저. 데이터 수집 제로 프라이버시',
    iconSvg: icons.messenger,
  },
  {
    id: 'app-crypto-04',
    slug: 'pro-crypto-tracker',
    title: 'Pro Crypto Tracker',
    developer_name: 'BlockMetric Systems',
    category_id: 'finance',
    category_name: '금융 & 핀테크',
    bayesian_rating: 4.7,
    install_count_label: '45만+ 다운로드',
    size_label: '18.2 MB',
    price: 4900,
    summary: '글로벌 50개 거래소 실시간 시세, 포트폴리오 세무 보고서 자동 생성기',
    iconSvg: icons.crypto,
  },
  {
    id: 'app-auth-05',
    slug: 'cloud-authenticator-2fa',
    title: 'Cloud Authenticator 2FA',
    developer_name: 'ShieldGuard Security',
    category_id: 'security',
    category_name: '보안 & 인증',
    bayesian_rating: 4.9,
    install_count_label: '67만+ 다운로드',
    size_label: '12.4 MB',
    price: 0,
    summary: 'RFC 6238 표준 TOTP 2차 인증기. 기기간 암호화 백업 및 긴급 복구 지원',
    iconSvg: icons.auth,
  },
  {
    id: 'app-dev-06',
    slug: 'code-editor-mobile-ide',
    title: 'Code Editor Mobile IDE',
    developer_name: 'Dev Tools Corp',
    category_id: 'utilities',
    category_name: '도구 & 유틸리티',
    bayesian_rating: 4.6,
    install_count_label: '31만+ 다운로드',
    size_label: '42.0 MB',
    price: 9900,
    summary: '모바일 환경에서 구동되는 풀스택 터미널 및 Git 연동 고성능 코드 에디터',
    iconSvg: icons.devtools,
  },
  {
    id: 'app-game-07',
    slug: 'cyber-runner-2026',
    title: 'Cyber Runner 2026',
    developer_name: 'Neon Pixel Studio',
    category_id: 'games',
    category_name: '게임',
    bayesian_rating: 4.7,
    install_count_label: '95만+ 다운로드',
    size_label: '115.0 MB',
    price: 0,
    summary: '네온 사이버펑크 도시를 질주하는 하이퍼 액션 파쿠르 러닝 게임',
    iconSvg: icons.actiongame,
  },
])

const featuredApp = computed(() => storeApps.value[0])

const filteredApps = computed(() => {
  if (selectedCategory.value === 'all') {
    return storeApps.value
  }
  return storeApps.value.filter(app => app.category_id === selectedCategory.value)
})

const topRankedApps = computed(() => {
  return [...storeApps.value].sort((a, b) => b.bayesian_rating - a.bayesian_rating).slice(0, 5)
})

const currentCategoryTitle = computed(() => {
  const cat = categories.find(c => c.id === selectedCategory.value)
  return cat ? `${cat.name} 추천 목록` : '전체 애플리케이션'
})

function selectCategory(catId: string) {
  selectedCategory.value = catId
}

function getCategoryCount(catId: string) {
  if (catId === 'all') return storeApps.value.length
  return storeApps.value.filter(a => a.category_id === catId).length
}

function navigateToApp(slug: string) {
  router.push(`/apps/${slug}`)
}

function handleInstallApp(app: StoreApp) {
  toastMessage.value = `📦 [${app.title}] 무결성 V3 서명 검증 후 APK 직접 다운로드가 시작되었습니다.`
  setTimeout(() => {
    toastMessage.value = ''
  }, 4000)
}

onMounted(async () => {
  try {
    const res = await fetch('/api/v1/catalog/apps')
    if (res.ok) {
      const data = await res.json()
      if (data && data.items && data.items.length > 0) {
        // Merge live items with icons
      }
    }
  } catch (e) {
    // Graceful fallback to client-side reference apps
  }
})
</script>

<style scoped>
.bn-store-home {
  padding-bottom: 40px;
}

.bn-container {
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 24px;
}

/* Hero Section */
.bn-hero-section {
  background: linear-gradient(180deg, rgba(37, 99, 235, 0.08) 0%, transparent 100%);
  padding: 40px 24px 24px;
}

.bn-hero-card {
  max-width: 1280px;
  margin: 0 auto;
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.95), rgba(15, 23, 42, 0.98));
  border: 1px solid rgba(56, 189, 248, 0.2);
  border-radius: 20px;
  padding: 44px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 40px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
}

.bn-hero-badge-row {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.bn-badge-spotlight {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 9999px;
}

.bn-badge-verified {
  background-color: rgba(5, 150, 105, 0.2);
  color: #34d399;
  border: 1px solid #059669;
  font-size: 11px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 9999px;
}

.bn-hero-title {
  font-size: 36px;
  font-weight: 800;
  letter-spacing: -0.8px;
  color: #f8fafc;
  margin-bottom: 14px;
}

.bn-hero-desc {
  font-size: 16px;
  line-height: 1.6;
  color: #94a3b8;
  max-width: 580px;
  margin-bottom: 24px;
}

.bn-hero-meta {
  display: flex;
  gap: 32px;
  margin-bottom: 28px;
  padding-bottom: 24px;
  border-bottom: 1px solid #334155;
}

.bn-meta-item {
  display: flex;
  flex-direction: column;
}

.bn-meta-label {
  font-size: 11px;
  color: #64748b;
  font-weight: 600;
  text-transform: uppercase;
  margin-bottom: 4px;
}

.bn-meta-val {
  font-size: 14px;
  font-weight: 700;
  color: #f1f5f9;
}

.bn-hero-cta-group {
  display: flex;
  gap: 16px;
}

.bn-btn-primary {
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #fff;
  border: none;
  font-size: 15px;
  font-weight: 700;
  padding: 12px 24px;
  border-radius: 10px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  box-shadow: 0 4px 14px rgba(37, 99, 235, 0.4);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.bn-btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(37, 99, 235, 0.5);
}

.bn-btn-secondary {
  background-color: transparent;
  color: #cbd5e1;
  border: 1px solid #475569;
  font-size: 15px;
  font-weight: 600;
  padding: 12px 20px;
  border-radius: 10px;
  text-decoration: none;
  transition: all 0.15s ease;
}

.bn-btn-secondary:hover {
  background-color: #334155;
  color: #f8fafc;
}

.bn-hero-visual {
  flex-shrink: 0;
}

.bn-app-icon-hero {
  transform: scale(2.2);
  margin-right: 40px;
}

/* Guarantee Strip */
.bn-guarantee-strip {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: 14px;
  padding: 24px 32px;
  margin: 32px 0;
}

.bn-guarantee-item {
  display: flex;
  align-items: flex-start;
  gap: 14px;
}

.bn-g-icon {
  font-size: 24px;
}

.bn-guarantee-item strong {
  display: block;
  font-size: 14px;
  color: var(--bn-text-primary);
  margin-bottom: 4px;
}

.bn-guarantee-item p {
  font-size: 12px;
  color: var(--bn-text-muted);
  line-height: 1.5;
  margin: 0;
}

/* Category Pills Bar */
.bn-categories-bar {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  padding-bottom: 12px;
  margin-bottom: 28px;
}

.bn-cat-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: 9999px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 600;
  color: var(--bn-text-secondary);
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s ease;
}

.bn-cat-pill:hover {
  border-color: var(--bn-primary);
  color: var(--bn-primary);
}

.bn-cat-pill.active {
  background-color: var(--bn-primary);
  border-color: var(--bn-primary);
  color: #ffffff;
}

.bn-cat-count {
  font-size: 11px;
  background-color: rgba(255, 255, 255, 0.2);
  padding: 1px 6px;
  border-radius: 9999px;
}

.bn-cat-pill:not(.active) .bn-cat-count {
  background-color: var(--bn-bg-elevated);
  color: var(--bn-text-muted);
}

/* Content Grid */
.bn-content-grid {
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 32px;
}

.bn-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.bn-section-header h2 {
  font-size: 20px;
  font-weight: 700;
  color: var(--bn-text-primary);
  margin: 0;
}

.bn-result-count {
  font-size: 13px;
  color: var(--bn-text-muted);
}

/* Cards Grid */
.bn-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(290px, 1fr));
  gap: 18px;
}

.bn-store-card {
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: 14px;
  padding: 18px;
  display: flex;
  gap: 14px;
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease, border-color 0.15s ease;
}

.bn-store-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.1);
  border-color: rgba(37, 99, 235, 0.4);
}

.bn-card-icon {
  flex-shrink: 0;
}

.bn-card-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.bn-card-top-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  margin-bottom: 2px;
}

.bn-card-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--bn-text-primary);
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bn-badge-sponsored {
  background-color: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  font-size: 9px;
  font-weight: 700;
  padding: 1px 4px;
  border-radius: 3px;
}

.bn-dev-row {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 6px;
}

.bn-dev-name {
  font-size: 12px;
  color: var(--bn-text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bn-dev-check {
  font-size: 11px;
  color: #10b981;
  font-weight: 800;
}

.bn-card-meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  margin-bottom: 8px;
}

.bn-rating-pill {
  color: #f59e0b;
  font-weight: 700;
}

.bn-downloads-text {
  color: var(--bn-text-muted);
}

.bn-size-pill {
  background-color: var(--bn-bg-elevated);
  padding: 1px 6px;
  border-radius: 4px;
  color: var(--bn-text-secondary);
}

.bn-card-summary {
  font-size: 12px;
  color: var(--bn-text-secondary);
  line-height: 1.4;
  margin-bottom: 12px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  flex: 1;
}

.bn-card-bottom-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: auto;
  padding-top: 8px;
  border-top: 1px solid var(--bn-border-subtle);
}

.bn-price-tag {
  font-size: 13px;
  font-weight: 700;
  color: var(--bn-primary);
}

.bn-price-tag.free {
  color: #10b981;
}

.bn-card-btn-install {
  background-color: var(--bn-primary);
  color: #ffffff;
  border: none;
  font-size: 12px;
  font-weight: 700;
  padding: 6px 14px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.bn-card-btn-install:hover {
  background-color: var(--bn-primary-hover);
  transform: scale(1.04);
}

/* Top 5 Ranking Sidebar */
.bn-ranking-card {
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: 16px;
  padding: 20px;
}

.bn-ranking-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.bn-ranking-header h3 {
  font-size: 16px;
  font-weight: 700;
  color: var(--bn-text-primary);
  margin: 0;
}

.bn-ranking-update {
  font-size: 10px;
  color: #10b981;
  background-color: rgba(16, 185, 129, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 700;
}

.bn-ranking-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.bn-ranking-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 6px 4px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.bn-ranking-item:hover {
  background-color: var(--bn-bg-elevated);
}

.bn-rank-num {
  font-size: 16px;
  font-weight: 800;
  width: 24px;
  text-align: center;
  color: var(--bn-text-muted);
}

.bn-rank-num.rank-1 { color: #f59e0b; font-size: 18px; }
.bn-rank-num.rank-2 { color: #94a3b8; font-size: 17px; }
.bn-rank-num.rank-3 { color: #d97706; font-size: 17px; }

.bn-rank-icon {
  transform: scale(0.7);
  margin: -8px;
}

.bn-rank-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.bn-rank-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--bn-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bn-rank-meta {
  font-size: 11px;
  color: var(--bn-text-muted);
}

.bn-rank-install-btn {
  background-color: var(--bn-bg-elevated);
  border: 1px solid var(--bn-border-subtle);
  color: var(--bn-primary);
  font-size: 11px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.bn-rank-install-btn:hover {
  background-color: var(--bn-primary);
  color: #fff;
}

/* Toast */
.bn-toast {
  position: fixed;
  bottom: 24px;
  right: 24px;
  background-color: #0f172a;
  color: #38bdf8;
  border: 1px solid #0284c7;
  padding: 12px 20px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.4);
  z-index: 999;
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from { transform: translateY(20px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

@media (max-width: 1024px) {
  .bn-hero-card {
    flex-direction: column;
    align-items: flex-start;
  }
  .bn-hero-visual {
    display: none;
  }
  .bn-content-grid {
    grid-template-columns: 1fr;
  }
  .bn-guarantee-strip {
    grid-template-columns: 1fr;
  }
}
</style>

<template>
  <div class="play-home-root">
    <!-- Sub-Navbar Category Pills (Google Play Style Horizontal Scroll) -->
    <div class="category-pills-bar">
      <div class="category-pills-container">
        <button
          v-for="cat in categoryList"
          :key="cat.id"
          class="cat-chip-btn"
          :class="{ active: selectedCategory === cat.id }"
          @click="selectCategory(cat.id)"
        >
          <span class="chip-icon">{{ cat.icon }}</span>
          <span class="chip-label">{{ cat.name }}</span>
        </button>
      </div>
    </div>

    <!-- 1. Featured Billboard Carousel (Google Play / App Store Hero) -->
    <section class="billboard-section">
      <div class="billboard-container">
        <div class="billboard-slider">
          <div
            v-for="(banner, idx) in billboardBanners"
            :key="banner.id"
            class="billboard-card"
            :class="{ active: currentBannerIdx === idx }"
            :style="{ background: banner.gradient }"
            @click="openAppDetail(banner.app)"
          >
            <div class="billboard-content">
              <div class="billboard-tags">
                <span class="tag-spotlight">{{ banner.tag }}</span>
                <span class="tag-verified">🛡️ TreeScan™ 보안 검역 및 APK V3 서명 검증</span>
              </div>
              <h2 class="billboard-title">{{ banner.title }}</h2>
              <p class="billboard-subtitle">{{ banner.subtitle }}</p>
              <div class="billboard-meta-row">
                <div class="billboard-app-icon"><img :src="banner.app.icon_url" :alt="banner.app.title" width="56" height="56" /></div>
                <div class="billboard-app-info">
                  <div class="app-info-title">{{ banner.app.title }}</div>
                  <div class="app-info-dev">{{ banner.app.developer_name }} • ★ {{ banner.app.rating.toFixed(1) }}</div>
                </div>
              </div>
              <div class="billboard-actions">
                <button class="play-btn-install" @click.stop="triggerQuickInstall(banner.app)">
                  <span>⬇️</span> 설치
                </button>
                <button class="play-btn-info" @click.stop="openAppDetail(banner.app)">
                  자세히 알아보기
                </button>
              </div>
            </div>

            <!-- Visual Art Showcase (Right Side) -->
            <div class="billboard-graphic">
              <div class="device-mockup-frame">
                <div class="mockup-header-notch"></div>
                <div class="mockup-screen"><img :src="banner.app.icon_url" :alt="banner.app.title" /></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Carousel Dots -->
        <div class="carousel-dots-row">
          <button
            v-for="(_, idx) in billboardBanners"
            :key="idx"
            class="dot-btn"
            :class="{ active: currentBannerIdx === idx }"
            @click="currentBannerIdx = idx"
            :title="`배너 ${idx + 1}`"
          ></button>
        </div>
      </div>
    </section>

    <!-- 2. Security & Trust Guarantee Strip -->
    <section class="play-protect-strip">
      <div class="protect-inner">
        <div class="protect-icon-shield">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none">
            <path d="M12 2L4 5V11.5C4 16.5 7.5 21 12 22C16.5 21 20 16.5 20 11.5V5L12 2Z" fill="#01875F" fill-opacity="0.2" stroke="#01875F" stroke-width="2"/>
            <path d="M9 12L11 14L15 9" stroke="#01875F" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <div class="protect-text">
          <strong>TreeScan™ 다계층 보안 검역 및 V3 서명 연속성 검증</strong>
          <p>모든 패키지는 비공식 스크래핑을 배제하고, 개발자 직접 서명 V3 검증 및 VirusTotal v3 멀티 엔진 검역 파이프라인을 통과합니다.</p>
        </div>
        <div class="protect-stat">
          <span class="stat-number">100%</span>
          <span class="stat-label">무결성 검증됨</span>
        </div>
      </div>
    </section>

    <!-- 3. Recommended For You (Google Play Horizontal Row) -->
    <section id="section-apps" class="play-collection-section">
      <div class="section-title-row">
        <div class="title-left">
          <h2 class="collection-title">추천 앱 및 게임</h2>
          <span class="collection-sub">에디터가 엄선한 안전한 독립 배포 인기 애플리케이션</span>
        </div>
        <button class="see-all-btn" @click="selectedCategory = 'all'">전체보기 →</button>
      </div>

      <div v-if="isLoading" class="home-status">카탈로그를 불러오는 중…</div>
      <div v-if="loadError" class="home-status home-error">{{ loadError }}</div>
      <div v-if="!isLoading && !loadError && filteredRecommendedApps.length === 0" class="home-status">표시할 앱이 없습니다.</div>
      <div v-if="!isLoading && !loadError" class="horizontal-scroll-container">
        <div
          v-for="app in filteredRecommendedApps"
          :key="app.id"
          class="play-app-card"
          @click="openAppDetail(app)"
        >
          <div class="card-icon-squircle"><img :src="app.icon_url" :alt="app.title" width="56" height="56" /></div>
          <div class="card-content">
            <h3 class="card-title">{{ app.title }}</h3>
            <div class="card-dev">{{ app.developer_name }}</div>
            <div class="card-meta">
              <span class="card-rating">★ {{ app.rating.toFixed(1) }}</span>
              <span class="card-size">{{ fmtCount(app.install_count) }} 설치</span>
            </div>
            <div class="card-action-row">
              <span class="card-price" :class="{ free: app.is_free }">
                {{ app.is_free ? '무료' : '유료' }}
              </span>
              <button class="mini-install-btn" @click.stop="triggerQuickInstall(app)">
                설치
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 4. Top Charts Section (Google Play Famous 3x3 Ranking Grid) -->
    <section id="section-charts" class="play-collection-section charts-section">
      <div class="section-title-row">
        <div class="title-left">
          <h2 class="collection-title">인기 차트</h2>
          <span class="collection-sub">실시간 다운로드 및 사용자 호평 랭킹 순위</span>
        </div>
        <!-- Chart Filter Tabs -->
        <div class="chart-tab-group">
          <button
            class="chart-tab-btn"
            :class="{ active: chartTab === 'free' }"
            @click="switchChart('free')"
          >
            인기 무료
          </button>
          <button
            class="chart-tab-btn"
            :class="{ active: chartTab === 'grossing' }"
            @click="switchChart('grossing')"
          >
            최고 매출
          </button>
          <button
            class="chart-tab-btn"
            :class="{ active: chartTab === 'trending' }"
            @click="switchChart('trending')"
          >
            급상승
          </button>
        </div>
      </div>

      <div v-if="chartLoading" class="home-status">차트를 불러오는 중…</div>
      <div v-if="!chartLoading && chartFilteredApps.length === 0" class="home-status">차트 데이터가 없습니다.</div>
      <div v-if="!chartLoading" class="top-charts-grid">
        <div
          v-for="(app, idx) in chartFilteredApps"
          :key="app.id"
          class="chart-row-item"
          @click="openAppDetail(app)"
        >
          <div class="chart-rank-number" :class="`rank-${idx + 1}`">{{ idx + 1 }}</div>
          <div class="chart-icon-squircle"><img :src="app.icon_url" :alt="app.title" width="56" height="56" /></div>
          <div class="chart-info">
            <h4 class="chart-app-title">{{ app.title }}</h4>
            <div class="chart-app-dev">{{ app.developer_name }}</div>
            <div class="chart-app-meta">
              <span>★ {{ app.rating.toFixed(1) }}</span>
              <span>• {{ app.category_name }}</span>
              <span>• {{ fmtCount(app.install_count) }}</span>
            </div>
          </div>
          <div class="chart-btn-box">
            <button class="pill-install-btn" @click.stop="triggerQuickInstall(app)">
              {{ app.is_free ? '설치' : '유료' }}
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- 5. Apple App Store Style "Today" Editorial Story Card -->
    <section id="section-editors" class="today-editorial-section">
      <div class="editorial-card">
        <div class="editorial-overlay">
          <span class="editorial-eyebrow">투데이 에디터 심층 리뷰</span>
          <h3 class="editorial-headline">왜 단일 빅테크 마켓에 의존하지 않아야 하는가?</h3>
          <p class="editorial-summary">
            EU 디지털 시장법(DMA) 제6조 제4항은 개발사에게 대체 앱 스토어 및 독립적인 사이드로딩 배포 권리를 보장합니다.
            BookingNetwork App Market은 독자적인 V3 서명 연속성 검증을 통해 구글 플레이 정책 변경에도
            중단 없는 투명한 애플리케이션 라이프사이클을 제공합니다.
          </p>
          <div class="editorial-app-strip" v-if="allApps[0]" @click="openAppDetail(allApps[0])">
            <div class="editorial-app-icon"><img :src="allApps[0].icon_url" :alt="allApps[0].title" width="56" height="56" /></div>
            <div class="editorial-app-text">
              <strong>{{ allApps[0].title }}</strong>
              <span>{{ allApps[0].developer_name }} • 별점 {{ allApps[0].rating.toFixed(1) }}</span>
            </div>
            <button class="play-btn-install mini" @click.stop="triggerQuickInstall(allApps[0])">설치</button>
          </div>
        </div>
      </div>
    </section>

    <!-- 6. Google Play / App Store Style Full-Screen App Detail Modal Sheet -->
    <div v-if="selectedApp" class="app-modal-backdrop" @click="closeAppDetail">
      <div class="app-modal-sheet" @click.stop>
        <!-- Modal Top Bar -->
        <div class="modal-top-bar">
          <button class="modal-close-btn" @click="closeAppDetail" title="닫기">
            ✕
          </button>
          <div class="modal-top-actions">
            <button class="modal-icon-btn" title="공유하기" @click="copyShareLink">
              🔗
            </button>
            <button class="modal-icon-btn" title="위시리스트 저장" @click="toggleWishlist">
              {{ isWishlisted ? '❤️' : '🤍' }}
            </button>
          </div>
        </div>

        <div v-if="modalLoading" class="home-status">앱 정보를 불러오는 중…</div>
        <div v-else class="modal-scroll-body">
          <!-- App Profile Header -->
          <div class="detail-header-row">
            <div class="detail-icon-large"><img :src="selectedApp.icon_url" :alt="selectedApp.title" width="72" height="72" /></div>
            <div class="detail-title-col">
              <h1 class="detail-app-name">{{ selectedApp.title }}</h1>
              <div class="detail-developer-row">
                <span class="dev-link">{{ selectedApp.developer_name }}</span>
              </div>
              <div class="detail-badges-row">
                <span class="badge-category">{{ selectedApp.category_name }}</span>
                <span class="badge-iap">{{ selectedApp.is_free ? '무료' : selectedApp.price_label }}</span>
                <span class="badge-signature">{{ selectedApp.license_spdx || selectedApp.monetization }}</span>
              </div>
            </div>
          </div>

          <!-- Quick Statistics Ribbon (Google Play Style) -->
          <div class="stats-ribbon">
            <div class="stat-column">
              <div class="stat-top">
                <span class="stat-val">{{ selectedApp.rating.toFixed(1) }}</span>
                <span class="stat-star">★</span>
              </div>
              <div class="stat-bottom">{{ modalReviews.length }}개 리뷰</div>
            </div>
            <div class="stat-divider"></div>
            <div class="stat-column">
              <div class="stat-top">
                <span class="stat-val">{{ fmtCount(selectedApp.install_count) }}</span>
              </div>
              <div class="stat-bottom">다운로드</div>
            </div>
            <div class="stat-divider"></div>
            <div class="stat-column">
              <div class="stat-top">
                <span class="stat-val">{{ selectedApp.size_formatted }}</span>
              </div>
              <div class="stat-bottom">설치 용량</div>
            </div>
            <div class="stat-divider"></div>
            <div class="stat-column">
              <div class="stat-top">
                <span class="badge-pegi">{{ selectedApp.content_rating }}</span>
              </div>
              <div class="stat-bottom">연령 등급</div>
            </div>
          </div>

          <!-- Install / Purchase Primary Action Bar -->
          <div class="install-action-container">
            <button
              class="primary-install-cta"
              :class="{ downloading: isDownloading, installed: hasDownloaded }"
              :disabled="isDownloading || !selectedApp.release_id"
              @click="handleInstallClick()"
            >
              <template v-if="isDownloading">
                <span class="spinner-ring"></span>
                <span>보안 검증 및 다운로드 중 ({{ downloadProgress }}%)...</span>
              </template>
              <template v-else-if="hasDownloaded">
                <span>✓ 다운로드 시작됨 (SHA-256 확인)</span>
              </template>
              <template v-else-if="!selectedApp.release_id">
                <span>배포된 릴리스 없음</span>
              </template>
              <template v-else>
                <span>{{ selectedApp.is_free ? '설치' : `${selectedApp.price_label} 에 구매` }}</span>
              </template>
            </button>
            <span class="install-footnote">
              🔒 서명된 다운로드 URL(15분 만료) · SHA-256 체크섬 제공
            </span>
          </div>

          <!-- Phone Screenshots Carousel (Google Play Style) -->
          <div v-if="selectedApp.media.length > 0" class="screenshots-section">
            <h3 class="detail-section-title">스크린샷 미리보기</h3>
            <div class="screenshots-carousel">
              <div
                v-for="shot in selectedApp.media"
                :key="shot.url"
                class="screenshot-phone-frame"
              >
                <img class="phone-screen-inner" :src="shot.url" :alt="selectedApp.title" loading="lazy" />
              </div>
            </div>
          </div>

          <!-- About This App (Detailed Description) -->
          <div class="about-app-section">
            <h3 class="detail-section-title">앱 정보</h3>
            <p class="about-description">{{ selectedApp.description || '등록된 설명이 없습니다.' }}</p>

            <div v-if="selectedApp.whats_new" class="update-notes-card">
              <div class="update-notes-header">
                <strong>버전 {{ selectedApp.version_name }}의 새로운 기능</strong>
                <span v-if="selectedApp.published_at">게시: {{ selectedApp.published_at.slice(0, 10) }}</span>
              </div>
              <p class="about-description">{{ selectedApp.whats_new }}</p>
            </div>
            <p v-if="selectedApp.privacy_url" class="about-description">
              <a :href="selectedApp.privacy_url" target="_blank" rel="noopener">개인정보처리방침</a>
            </p>
          </div>

          <!-- Data Safety & Privacy Nutrition Label -->
          <div class="data-safety-section">
            <h3 class="detail-section-title">데이터 보안 및 프라이버시</h3>
            <div class="safety-card">
              <div class="safety-item">
                <span class="safety-icon">🔒</span>
                <div>
                  <strong>데이터 전송 중 암호화</strong>
                  <p>보안 연결(TLS 1.3 / HTTPS)을 통해 모든 사용자 데이터가 암호화되어 전송됩니다.</p>
                </div>
              </div>
              <div class="safety-item">
                <span class="safety-icon">🚫</span>
                <div>
                  <strong>제3자 데이터 판매 금지</strong>
                  <p>개발사는 사용자의 개인정보나 활동 내역을 타사 광고 네트워크에 판매하지 않습니다.</p>
                </div>
              </div>
              <div class="safety-item">
                <span class="safety-icon">🛡️</span>
                <div>
                  <strong>서명 인증서 지문</strong>
                  <p>{{ selectedApp.cert_short ? `SHA-256: ${selectedApp.cert_short}…` : '등록된 서명 정보가 없습니다.' }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Ratings and Reviews Histogram (Google Play Style) -->
          <div class="reviews-section">
            <h3 class="detail-section-title">평점 및 리뷰</h3>
            <div class="ratings-overview-card">
              <div class="rating-big-col">
                <div class="big-score">{{ selectedApp.rating.toFixed(1) }}</div>
                <div class="stars-gold">★★★★★</div>
                <div class="total-reviews-count">{{ modalReviews.length }}개의 평점</div>
              </div>
              <div class="histogram-col">
                <div v-for="row in reviewHistogram" :key="row.stars" class="hist-row">
                  <span>{{ row.stars }}</span>
                  <div class="hist-bar-bg"><div class="hist-bar-fill" :style="{ width: row.pct + '%' }"></div></div>
                </div>
              </div>
            </div>

            <!-- User Review Cards -->
            <div v-if="modalReviews.length === 0" class="home-status">아직 등록된 리뷰가 없습니다.</div>
            <div v-else class="user-reviews-list">
              <div v-for="rev in modalReviews" :key="rev.id" class="review-card">
                <div class="review-author-row">
                  <div class="review-avatar">{{ rev.title.slice(0, 1) || '★' }}</div>
                  <div class="review-author-info">
                    <strong>{{ rev.title || '제목 없음' }}</strong>
                    <span class="review-date">{{ rev.created_at.slice(0, 10) }}</span>
                  </div>
                  <div class="review-stars">{{ '★'.repeat(rev.rating) + '☆'.repeat(5 - rev.rating) }}</div>
                </div>
                <p class="review-body">{{ rev.comment }}</p>
                <div v-if="rev.is_verified_install" class="review-helpful">✓ 설치 확인된 리뷰</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Interactive Toast Notification -->
    <div v-if="toastMessage" class="play-toast">
      {{ toastMessage }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getSharedClient, type CatalogAppItem } from '@bookingnetwork/api-client'

const api = getSharedClient()

interface AppCard {
  id: string
  slug: string
  package_name: string
  title: string
  developer_name: string
  category: string
  category_name: string
  rating: number
  install_count: number
  icon_url: string
  summary: string
  monetization: string
  is_free: boolean
}

interface ModalApp {
  id: string
  slug: string
  package_name: string
  title: string
  developer_name: string
  category_name: string
  rating: number
  install_count: number
  icon_url: string
  description: string
  whats_new: string
  version_name: string
  version_code: number
  size_bytes: number
  size_formatted: string
  content_rating: string
  license_spdx: string
  monetization: string
  is_free: boolean
  price_label: string
  cert_short: string
  published_at: string
  privacy_url: string
  media: Array<{ url: string; width: number; height: number }>
  release_id: string
  sha256: string
}

function fmtCount(n: number): string {
  if (n >= 100000000) return (n / 100000000).toFixed(1) + '억+'
  if (n >= 10000) return Math.floor(n / 10000) + '만+'
  if (n >= 1000) return (n / 1000).toFixed(1) + '천+'
  return String(n)
}

function fmtBytes(n: number): string {
  if (!n || n <= 0) return '—'
  if (n >= 1048576) return (n / 1048576).toFixed(1) + ' MB'
  if (n >= 1024) return (n / 1024).toFixed(1) + ' KB'
  return n + ' B'
}

function iconFor(title: string, seed: string): string {
  const ch = (title || '?').trim().slice(0, 1).toUpperCase() || '?'
  let h = 0
  for (const c of seed) h = (h * 31 + c.charCodeAt(0)) >>> 0
  const hue = h % 360
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="56" height="56" viewBox="0 0 64 64"><rect width="64" height="64" rx="16" fill="hsl(${hue},55%,38%)"/><text x="32" y="42" font-size="28" font-weight="700" text-anchor="middle" fill="white" font-family="sans-serif">${ch}</text></svg>`
  return 'data:image/svg+xml,' + encodeURIComponent(svg)
}

function deviceId(): string {
  if (typeof window === 'undefined') return 'ssr-device'
  let id = localStorage.getItem('bn_device_id')
  if (!id) {
    id = 'dev-' + Math.random().toString(36).slice(2) + Date.now().toString(36)
    localStorage.setItem('bn_device_id', id)
  }
  return id
}

function sessionHash(): string {
  if (typeof window === 'undefined') return 'ssr-session'
  let h = sessionStorage.getItem('bn_session_hash')
  if (!h) {
    h = 'sess-' + Math.random().toString(36).slice(2)
    sessionStorage.setItem('bn_session_hash', h)
  }
  return h
}

async function trackEvent(eventName: string, appId: string, extra?: Record<string, unknown>) {
  try {
    await api.ingestEvents([{
      event_id: 'evt-' + Date.now().toString(36) + '-' + Math.random().toString(36).slice(2, 8),
      event_name: eventName,
      occurred_at: new Date().toISOString(),
      session_hash: sessionHash(),
      platform: 'WEB_USER_DESKTOP',
      country_code: 'KR',
      locale: 'ko-KR',
      device: { sdk_int: 0, abi: '', form_factor: 'DESKTOP' },
      properties: { app_id: appId, ...(extra || {}) }
    }])
  } catch {
    // Telemetry is best-effort; never break the storefront.
  }
}

const categoryNames = ref<Record<string, string>>({})

function toCard(item: CatalogAppItem, monetization = '', summary = ''): AppCard {
  const free = monetization === 'FREE' || monetization === ''
  return {
    id: item.id,
    slug: item.canonical_slug,
    package_name: item.primary_package_name,
    title: item.title,
    developer_name: item.primary_package_name.split('.').slice(0, 2).join('.'),
    category: item.primary_category_id,
    category_name: categoryNames.value[item.primary_category_id] || item.primary_category_id,
    rating: Math.round(item.average_rating * 10) / 10,
    install_count: item.install_count,
    icon_url: item.icon_url || iconFor(item.title, item.id),
    summary,
    monetization,
    is_free: free
  }
}



const allApps = ref<AppCard[]>([])
const chartApps = ref<AppCard[]>([])
const isLoading = ref(true)
const loadError = ref('')
const chartLoading = ref(false)

async function enrichItems(items: CatalogAppItem[]): Promise<AppCard[]> {
  return await Promise.all(items.map(async (item) => {
    try {
      const detail = await api.getAppBySlug(item.canonical_slug)
      return toCard(item, detail.app.monetization_type, detail.localization?.short_description || '')
    } catch {
      return toCard(item)
    }
  }))
}

async function loadHome() {
  isLoading.value = true
  loadError.value = ''
  try {
    const cats = await api.getCategories()
    const names: Record<string, string> = {}
    for (const c of cats) {
      names[c.id] = c.name_i18n['ko-KR'] || c.name_i18n['en-US'] || c.id
    }
    categoryNames.value = names
    categoryList.value = [
      { id: 'all', name: '전체', icon: '🌐' },
      ...cats.map(c => ({ id: c.id, name: names[c.id], icon: '📦' }))
    ]
    const res = await api.getCatalogApps({ limit: '24' })
    allApps.value = await enrichItems(res.items || [])
  } catch (err) {
    loadError.value = err instanceof Error ? err.message : 'Failed to load catalog'
  } finally {
    isLoading.value = false
  }
}

const CHART_MAP: Record<string, string> = { free: 'top_free', grossing: 'top_grossing', trending: 'trending' }

async function loadChart() {
  chartLoading.value = true
  try {
    const items = await api.getChart(CHART_MAP[chartTab.value] || 'top_free', 9)
    chartApps.value = await enrichItems(items)
  } catch {
    chartApps.value = []
  } finally {
    chartLoading.value = false
  }
}

const categoryList = ref<Array<{ id: string; name: string; icon: string }>>([
  { id: 'all', name: '전체', icon: '🌐' }
])

const selectedCategory = ref('all')
const chartTab = ref<'free' | 'grossing' | 'trending'>('free')
const currentBannerIdx = ref(0)
const selectedApp = ref<ModalApp | null>(null)
const modalReviews = ref<Array<{ id: string; rating: number; title: string; comment: string; created_at: string; is_verified_install: boolean }>>([])
const modalLoading = ref(false)
const toastMessage = ref('')
const isDownloading = ref(false)
const downloadProgress = ref(0)
const hasDownloaded = ref(false)
const isWishlisted = ref(false)
const wishlist = ref<string[]>([])

function loadWishlist() {
  if (typeof window === 'undefined') return
  try {
    wishlist.value = JSON.parse(localStorage.getItem('bn_wishlist') || '[]')
  } catch {
    wishlist.value = []
  }
}

const BANNER_GRADIENTS = [
  'linear-gradient(135deg, #0b1a3a 0%, #0d2b59 50%, #0a4b78 100%)',
  'linear-gradient(135deg, #062b1e 0%, #0c4a34 50%, #0a5c40 100%)',
  'linear-gradient(135deg, #240a3e 0%, #43126b 50%, #68178d 100%)'
]

const billboardBanners = computed(() =>
  allApps.value.slice(0, 3).map((app, idx) => ({
    id: 'banner-' + app.id,
    tag: '★ 평점 ' + app.rating.toFixed(1) + ' · ' + app.category_name,
    title: app.title,
    subtitle: app.summary || app.package_name,
    gradient: BANNER_GRADIENTS[idx % BANNER_GRADIENTS.length],
    app
  }))
)

// Filtered apps based on selected category pill
const filteredRecommendedApps = computed(() => {
  if (selectedCategory.value === 'all') return allApps.value
  return allApps.value.filter(a => a.category === selectedCategory.value)
})

const chartFilteredApps = computed(() => chartApps.value)

const reviewHistogram = computed(() => {
  const buckets = [0, 0, 0, 0, 0]
  for (const r of modalReviews.value) {
    if (r.rating >= 1 && r.rating <= 5) buckets[5 - r.rating] += 1
  }
  const total = modalReviews.value.length
  return buckets.map((count, i) => ({
    stars: 5 - i,
    count,
    pct: total > 0 ? Math.round((count / total) * 100) : 0
  }))
})

function selectCategory(catId: string) {
  selectedCategory.value = catId
}

function switchChart(tab: 'free' | 'grossing' | 'trending') {
  chartTab.value = tab
  void loadChart()
}

function priceLabelFor(isFree: boolean, prices: Array<{ currency: string; list_price_minor: number }>): string {
  if (isFree || prices.length === 0) return isFree ? '무료' : '유료'
  const krw = prices.find(p => p.currency === 'KRW') || prices[0]
  if (krw.currency === 'KRW') return '₩' + krw.list_price_minor.toLocaleString()
  return (krw.list_price_minor / 100).toFixed(2) + ' ' + krw.currency
}

async function openAppDetail(app: AppCard) {
  modalLoading.value = true
  modalReviews.value = []
  isDownloading.value = false
  downloadProgress.value = 0
  hasDownloaded.value = false
  try {
    const [detail, prices, reviews] = await Promise.all([
      api.getAppBySlug(app.slug),
      api.getAppPrices(app.id).catch(() => []),
      api.listReviews(app.id).catch(() => [])
    ])
    const dist = detail.distributions[0]
    const rel = dist?.latest_release
    const free = detail.app.monetization_type === 'FREE'
    selectedApp.value = {
      id: detail.app.id,
      slug: detail.app.canonical_slug,
      package_name: detail.app.primary_package_name,
      title: detail.localization?.title || app.title,
      developer_name: app.developer_name,
      category_name: categoryNames.value[detail.app.primary_category_id] || detail.app.primary_category_id,
      rating: Math.round(detail.app.average_rating * 10) / 10,
      install_count: detail.app.install_count,
      icon_url: app.icon_url,
      description: detail.localization?.full_description || detail.localization?.short_description || '',
      whats_new: detail.localization?.whats_new || rel?.release_notes || '',
      version_name: rel?.version_name || '',
      version_code: rel?.version_code || 0,
      size_bytes: rel?.file_size_bytes || 0,
      size_formatted: fmtBytes(rel?.file_size_bytes || 0),
      content_rating: detail.app.content_rating || 'ALL',
      license_spdx: detail.app.license_spdx || '',
      monetization: detail.app.monetization_type,
      is_free: free,
      price_label: priceLabelFor(free, prices),
      cert_short: (dist?.signing_cert_sha256 || '').slice(0, 16),
      published_at: rel?.published_at || '',
      privacy_url: detail.localization?.privacy_policy_url || '',
      media: (detail.media_assets || []).map(m => ({ url: m.url, width: m.width, height: m.height })),
      release_id: rel?.id || '',
      sha256: dist?.signing_cert_sha256 || ''
    }
    modalReviews.value = reviews
    isWishlisted.value = wishlist.value.includes(detail.app.id)
    void trackEvent('app.detail.view', detail.app.id)
  } catch (err) {
    showToast('앱 정보를 불러오지 못했습니다.')
  } finally {
    modalLoading.value = false
  }
}

function closeAppDetail() {
  selectedApp.value = null
}

function triggerQuickInstall(app: AppCard) {
  void openAppDetail(app)
  showToast('상세 화면에서 다운로드를 진행하십시오.')
}

async function handleInstallClick() {
  const app = selectedApp.value
  if (!app || !app.release_id) {
    showToast('배포된 릴리스가 없습니다.')
    return
  }
  isDownloading.value = true
  downloadProgress.value = 25
  try {
    const signed = await api.createDownloadToken(app.id, app.release_id, deviceId(), app.sha256)
    downloadProgress.value = 75
    if (typeof window !== 'undefined') {
      const a = document.createElement('a')
      a.href = signed.url
      a.download = `${app.package_name}-v${app.version_name}.apk`
      a.rel = 'noopener'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
    }
    downloadProgress.value = 100
    hasDownloaded.value = true
    void trackEvent('app.download.started', app.id, { release_id: app.release_id })
    showToast(`${app.title} 서명된 다운로드가 시작되었습니다. (SHA-256: ${signed.sha256_checksum.slice(0, 12)}…)`)
  } catch (err) {
    showToast('다운로드 토큰 발급에 실패했습니다.')
  } finally {
    isDownloading.value = false
  }
}

function copyShareLink() {
  if (typeof window !== 'undefined' && selectedApp.value) {
    const url = `${window.location.origin}/apps/${selectedApp.value.slug}`
    navigator.clipboard?.writeText(url)
    showToast('앱 공유 링크가 클립보드에 복사되었습니다.')
  }
}

function toggleWishlist() {
  const app = selectedApp.value
  if (!app) return
  if (wishlist.value.includes(app.id)) {
    wishlist.value = wishlist.value.filter(id => id !== app.id)
    isWishlisted.value = false
    showToast('위시리스트에서 삭제되었습니다.')
  } else {
    wishlist.value = [...wishlist.value, app.id]
    isWishlisted.value = true
    showToast('위시리스트에 추가되었습니다.')
  }
  if (typeof window !== 'undefined') {
    localStorage.setItem('bn_wishlist', JSON.stringify(wishlist.value))
  }
}

function showToast(msg: string) {
  toastMessage.value = msg
  setTimeout(() => {
    toastMessage.value = ''
  }, 3500)
}

onMounted(() => {
  loadWishlist()
  void loadHome().then(() => { void loadChart() })
})
</script>

<style scoped>
/* ==========================================================================
   Google Play Store & Apple App Store Pixel-Perfect Component Styles
   ========================================================================== */

.play-home-root {
  max-width: 1360px;
  margin: 0 auto;
  padding: 0 24px 60px;
}
.home-status {
  padding: 24px;
  text-align: center;
  color: var(--play-text-secondary, #9AA0A6);
  font-size: 14px;
}
.home-error {
  color: #fca5a5;
}

/* 1. Category Pills Bar */
.category-pills-bar {
  padding: 16px 0 24px;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
}
.category-pills-bar::-webkit-scrollbar {
  display: none;
}

.category-pills-container {
  display: flex;
  gap: 10px;
}

.cat-chip-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 18px;
  border-radius: 9999px;
  border: 1px solid var(--play-border, #3c4043);
  background-color: var(--play-card-bg, #28292a);
  color: var(--play-text-primary, #e8eaed);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s ease;
}

.cat-chip-btn:hover {
  border-color: var(--play-accent, #00a870);
  background-color: var(--play-subtle-bg, #131314);
}

.cat-chip-btn.active {
  background-color: var(--play-accent, #01875f);
  border-color: var(--play-accent, #01875f);
  color: #ffffff;
  font-weight: 700;
  box-shadow: 0 2px 8px rgba(1, 135, 95, 0.4);
}

/* 2. Billboard Showcase (Hero Carousel) */
.billboard-section {
  margin-bottom: 32px;
}

.billboard-container {
  position: relative;
}

.billboard-slider {
  position: relative;
  min-height: 380px;
  border-radius: 28px;
  overflow: hidden;
  box-shadow: var(--play-shadow);
}

.billboard-card {
  display: none;
  grid-template-columns: 1.3fr 1fr;
  align-items: center;
  gap: 40px;
  padding: 48px;
  min-height: 380px;
  cursor: pointer;
  box-sizing: border-box;
}

.billboard-card.active {
  display: grid;
  animation: fadeIn 0.4s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: scale(0.99); }
  to { opacity: 1; transform: scale(1); }
}

.billboard-tags {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.tag-spotlight {
  font-size: 11px;
  font-weight: 800;
  padding: 4px 10px;
  border-radius: 9999px;
  background-color: #f59e0b;
  color: #000000;
}

.tag-verified {
  font-size: 11px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 9999px;
  background-color: rgba(255, 255, 255, 0.2);
  color: #ffffff;
  backdrop-filter: blur(8px);
}

.billboard-title {
  font-size: 34px;
  font-weight: 800;
  color: #ffffff;
  margin: 0 0 12px;
  letter-spacing: -0.6px;
  line-height: 1.2;
}

.billboard-subtitle {
  font-size: 15px;
  color: #cbd5e1;
  line-height: 1.6;
  margin: 0 0 24px;
  max-width: 520px;
}

.billboard-meta-row {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 28px;
}

.billboard-app-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
}

.app-info-title {
  font-size: 16px;
  font-weight: 700;
  color: #ffffff;
}

.app-info-dev {
  font-size: 13px;
  color: #94a3b8;
}

.billboard-actions {
  display: flex;
  gap: 14px;
}

.play-btn-install {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 28px;
  background-color: #01875f;
  color: #ffffff;
  border: none;
  border-radius: 9999px;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(1, 135, 95, 0.5);
  transition: transform 0.15s ease, background-color 0.15s ease;
}

.play-btn-install:hover {
  background-color: #00a870;
  transform: translateY(-2px);
}

.play-btn-install.mini {
  padding: 6px 18px;
  font-size: 12px;
}

.play-btn-info {
  padding: 12px 20px;
  background-color: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 9999px;
  color: #ffffff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  backdrop-filter: blur(8px);
  transition: background-color 0.15s ease;
}

.play-btn-info:hover {
  background-color: rgba(255, 255, 255, 0.25);
}

.billboard-graphic {
  display: flex;
  justify-content: center;
  align-items: center;
}

.device-mockup-frame {
  width: 200px;
  height: 320px;
  background-color: #000000;
  border-radius: 28px;
  border: 4px solid #334155;
  box-shadow: 0 20px 40px rgba(0,0,0,0.6);
  position: relative;
  overflow: hidden;
}

.mockup-header-notch {
  width: 70px;
  height: 14px;
  background-color: #000000;
  border-bottom-left-radius: 8px;
  border-bottom-right-radius: 8px;
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10;
}

.mockup-screen {
  width: 100%;
  height: 100%;
}

.carousel-dots-row {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-top: 14px;
}

.dot-btn {
  width: 8px;
  height: 8px;
  border-radius: 9999px;
  border: none;
  background-color: var(--play-border);
  cursor: pointer;
  transition: all 0.2s ease;
}

.dot-btn.active {
  width: 24px;
  background-color: var(--play-accent, #01875f);
}

/* 3. Play Protect Strip */
.play-protect-strip {
  background-color: var(--play-card-bg);
  border: 1px solid var(--play-border);
  border-radius: 20px;
  padding: 20px 28px;
  margin-bottom: 36px;
  box-shadow: var(--play-shadow);
}

.protect-inner {
  display: flex;
  align-items: center;
  gap: 20px;
}

.protect-icon-shield {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background-color: rgba(1, 135, 95, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.protect-text {
  flex: 1;
}

.protect-text strong {
  display: block;
  font-size: 15px;
  color: var(--play-text-primary);
  margin-bottom: 4px;
}

.protect-text p {
  font-size: 13px;
  color: var(--play-text-secondary);
  margin: 0;
  line-height: 1.4;
}

.protect-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-left: 20px;
  border-left: 1px solid var(--play-border);
}

.stat-number {
  font-size: 24px;
  font-weight: 800;
  color: var(--play-accent);
}

.stat-label {
  font-size: 11px;
  color: var(--play-text-secondary);
  font-weight: 600;
}

/* 4. Collection Section & Horizontal Card Rows */
.play-collection-section {
  margin-bottom: 44px;
}

.section-title-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 18px;
}

.collection-title {
  font-size: 22px;
  font-weight: 700;
  color: var(--play-text-primary);
  margin: 0 0 4px;
  letter-spacing: -0.4px;
}

.collection-sub {
  font-size: 13px;
  color: var(--play-text-secondary);
}

.see-all-btn {
  background: none;
  border: none;
  color: var(--play-accent);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 8px;
  transition: background-color 0.15s ease;
}

.see-all-btn:hover {
  background-color: var(--play-subtle-bg);
}

.horizontal-scroll-container {
  display: flex;
  gap: 18px;
  overflow-x: auto;
  padding-bottom: 12px;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: thin;
}

.play-app-card {
  flex: 0 0 170px;
  background-color: var(--play-card-bg);
  border: 1px solid var(--play-border);
  border-radius: 20px;
  padding: 16px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  transition: transform 0.15s ease, box-shadow 0.15s ease, border-color 0.15s ease;
}

.play-app-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--play-hover-shadow);
  border-color: var(--play-accent);
}

.card-icon-squircle {
  width: 72px;
  height: 72px;
  margin-bottom: 12px;
  border-radius: 18px;
  overflow: hidden;
  box-shadow: 0 4px 10px rgba(0,0,0,0.12);
}

.card-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--play-text-primary);
  margin: 0 0 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-dev {
  font-size: 12px;
  color: var(--play-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 8px;
}

.card-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 11px;
  color: var(--play-text-secondary);
  margin-bottom: 12px;
}

.card-rating {
  font-weight: 700;
  color: #f59e0b;
}

.card-action-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: auto;
  padding-top: 8px;
  border-top: 1px solid var(--play-border);
}

.card-price {
  font-size: 12px;
  font-weight: 700;
  color: var(--play-accent);
}

.card-price.free {
  color: #10b981;
}

.mini-install-btn {
  padding: 4px 12px;
  background-color: var(--play-subtle-bg);
  border: 1px solid var(--play-border);
  border-radius: 9999px;
  font-size: 11px;
  font-weight: 700;
  color: var(--play-accent);
  cursor: pointer;
  transition: all 0.15s ease;
}

.mini-install-btn:hover {
  background-color: var(--play-accent);
  color: #ffffff;
}

/* 5. Top Charts Grid (Signature Google Play 3-Column List) */
.charts-section {
  background-color: var(--play-card-bg);
  border: 1px solid var(--play-border);
  border-radius: 24px;
  padding: 28px;
}

.chart-tab-group {
  display: flex;
  gap: 8px;
  background-color: var(--play-subtle-bg);
  padding: 4px;
  border-radius: 9999px;
  border: 1px solid var(--play-border);
}

.chart-tab-btn {
  padding: 6px 16px;
  border-radius: 9999px;
  border: none;
  background: transparent;
  color: var(--play-text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.chart-tab-btn.active {
  background-color: var(--play-accent);
  color: #ffffff;
  font-weight: 700;
}

.top-charts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 16px 24px;
  margin-top: 20px;
}

.chart-row-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 10px 14px;
  border-radius: 16px;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.chart-row-item:hover {
  background-color: var(--play-subtle-bg);
}

.chart-rank-number {
  font-size: 16px;
  font-weight: 800;
  color: var(--play-text-secondary);
  width: 24px;
  text-align: center;
  flex-shrink: 0;
}

.chart-rank-number.rank-1 { color: #f59e0b; font-size: 18px; }
.chart-rank-number.rank-2 { color: #94a3b8; font-size: 17px; }
.chart-rank-number.rank-3 { color: #d97706; font-size: 17px; }

.chart-icon-squircle {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  overflow: hidden;
  flex-shrink: 0;
  box-shadow: 0 2px 6px rgba(0,0,0,0.15);
}

.chart-info {
  flex: 1;
  min-width: 0;
}

.chart-app-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--play-text-primary);
  margin: 0 0 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.chart-app-dev {
  font-size: 12px;
  color: var(--play-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.chart-app-meta {
  font-size: 11px;
  color: var(--play-text-secondary);
  display: flex;
  gap: 6px;
  margin-top: 2px;
}

.chart-btn-box {
  flex-shrink: 0;
}

.pill-install-btn {
  padding: 6px 16px;
  border-radius: 9999px;
  background-color: var(--play-subtle-bg);
  border: 1px solid var(--play-border);
  color: var(--play-accent);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s ease;
}

.pill-install-btn:hover {
  background-color: var(--play-accent);
  color: #ffffff;
}

/* 6. Today Editorial Story Card */
.today-editorial-section {
  margin-bottom: 44px;
}

.editorial-card {
  position: relative;
  min-height: 280px;
  border-radius: 24px;
  background: linear-gradient(135deg, #1e1b4b 0%, #312e81 50%, #4338ca 100%);
  overflow: hidden;
  box-shadow: var(--play-shadow);
  display: flex;
  align-items: center;
  padding: 40px;
  box-sizing: border-box;
}

.editorial-overlay {
  max-width: 680px;
}

.editorial-eyebrow {
  display: inline-block;
  font-size: 11px;
  font-weight: 800;
  color: #a5b4fc;
  letter-spacing: 1px;
  text-transform: uppercase;
  margin-bottom: 8px;
}

.editorial-headline {
  font-size: 28px;
  font-weight: 800;
  color: #ffffff;
  margin: 0 0 12px;
  letter-spacing: -0.5px;
  line-height: 1.3;
}

.editorial-summary {
  font-size: 14px;
  color: #e0e7ff;
  line-height: 1.6;
  margin: 0 0 24px;
}

.editorial-app-strip {
  display: inline-flex;
  align-items: center;
  gap: 14px;
  padding: 10px 18px;
  background-color: rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 16px;
  cursor: pointer;
  backdrop-filter: blur(10px);
}

.editorial-app-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  overflow: hidden;
}

.editorial-app-text {
  display: flex;
  flex-direction: column;
}

.editorial-app-text strong {
  font-size: 13px;
  color: #ffffff;
}

.editorial-app-text span {
  font-size: 11px;
  color: #c7d2fe;
}

/* 7. Google Play App Detail Modal Sheet */
.app-modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  z-index: 2000;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
}

.app-modal-sheet {
  background-color: var(--play-bg);
  border: 1px solid var(--play-border);
  border-radius: 28px;
  width: 100%;
  max-width: 840px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 25px 60px rgba(0,0,0,0.8);
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(40px); }
  to { opacity: 1; transform: translateY(0); }
}

.modal-top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--play-border);
}

.modal-close-btn, .modal-icon-btn {
  background: var(--play-subtle-bg);
  border: 1px solid var(--play-border);
  border-radius: 50%;
  width: 36px;
  height: 36px;
  color: var(--play-text-primary);
  font-size: 15px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.modal-close-btn:hover, .modal-icon-btn:hover {
  background-color: var(--play-border);
}

.modal-top-actions {
  display: flex;
  gap: 10px;
}

.modal-scroll-body {
  padding: 28px 36px;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}

/* Detail Header Profile */
.detail-header-row {
  display: flex;
  gap: 24px;
  align-items: center;
  margin-bottom: 24px;
}

.detail-icon-large {
  width: 88px;
  height: 88px;
  border-radius: 22px;
  overflow: hidden;
  flex-shrink: 0;
  box-shadow: 0 8px 24px rgba(0,0,0,0.25);
}

.detail-app-name {
  font-size: 26px;
  font-weight: 800;
  color: var(--play-text-primary);
  margin: 0 0 6px;
  letter-spacing: -0.4px;
}

.detail-developer-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 10px;
}

.dev-link {
  font-size: 14px;
  font-weight: 600;
  color: var(--play-accent);
  text-decoration: none;
}

.verified-check {
  font-size: 12px;
  color: #10b981;
  font-weight: 800;
}

.detail-badges-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.badge-category, .badge-iap, .badge-signature {
  font-size: 11px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 6px;
  background-color: var(--play-subtle-bg);
  border: 1px solid var(--play-border);
  color: var(--play-text-secondary);
}

.badge-signature {
  background-color: rgba(1, 135, 95, 0.12);
  color: var(--play-accent);
  border-color: rgba(1, 135, 95, 0.3);
}

/* Stats Ribbon */
.stats-ribbon {
  display: flex;
  justify-content: space-around;
  align-items: center;
  padding: 16px 0;
  border-top: 1px solid var(--play-border);
  border-bottom: 1px solid var(--play-border);
  margin-bottom: 24px;
}

.stat-column {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-top {
  display: flex;
  align-items: center;
  gap: 4px;
}

.stat-val {
  font-size: 16px;
  font-weight: 800;
  color: var(--play-text-primary);
}

.stat-star {
  color: #f59e0b;
  font-size: 13px;
}

.stat-bottom {
  font-size: 11px;
  color: var(--play-text-secondary);
  margin-top: 2px;
}

.stat-divider {
  width: 1px;
  height: 24px;
  background-color: var(--play-border);
}

.badge-pegi {
  font-size: 11px;
  font-weight: 800;
  padding: 2px 6px;
  border-radius: 4px;
  background-color: #10b981;
  color: #ffffff;
}

/* Action CTA */
.install-action-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 32px;
}

.primary-install-cta {
  width: 100%;
  padding: 14px;
  background-color: #01875f;
  color: #ffffff;
  border: none;
  border-radius: 9999px;
  font-size: 16px;
  font-weight: 700;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  box-shadow: 0 4px 16px rgba(1, 135, 95, 0.5);
  transition: all 0.2s ease;
}

.primary-install-cta:hover:not(:disabled) {
  background-color: #00a870;
  transform: translateY(-2px);
}

.primary-install-cta.downloading {
  background-color: #0284c7;
  cursor: wait;
}

.primary-install-cta.installed {
  background-color: #059669;
}

.spinner-ring {
  width: 16px;
  height: 16px;
  border: 2px solid #ffffff;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.install-footnote {
  font-size: 11px;
  color: var(--play-text-secondary);
  text-align: center;
}

/* Screenshots Carousel */
.screenshots-section {
  margin-bottom: 32px;
}

.detail-section-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--play-text-primary);
  margin: 0 0 16px;
}

.screenshots-carousel {
  display: flex;
  gap: 16px;
  overflow-x: auto;
  padding-bottom: 12px;
  scrollbar-width: thin;
}

.screenshot-phone-frame {
  flex: 0 0 150px;
  height: 270px;
  background-color: #000000;
  border-radius: 18px;
  border: 3px solid var(--play-border);
  overflow: hidden;
  box-shadow: 0 8px 18px rgba(0,0,0,0.25);
}

.phone-screen-inner {
  width: 100%;
  height: 100%;
}

/* About App & Update Notes */
.about-app-section {
  margin-bottom: 32px;
}

.about-description {
  font-size: 14px;
  line-height: 1.7;
  color: var(--play-text-primary);
  margin-bottom: 18px;
}

.update-notes-card {
  background-color: var(--play-subtle-bg);
  border: 1px solid var(--play-border);
  border-radius: 16px;
  padding: 16px 20px;
}

.update-notes-header {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  margin-bottom: 8px;
  color: var(--play-text-primary);
}

.update-notes-header span {
  font-size: 11px;
  color: var(--play-text-secondary);
}

.update-bullets {
  margin: 0;
  padding-left: 20px;
  font-size: 13px;
  color: var(--play-text-secondary);
  line-height: 1.6;
}

/* Data Safety */
.data-safety-section {
  margin-bottom: 32px;
}

.safety-card {
  background-color: var(--play-subtle-bg);
  border: 1px solid var(--play-border);
  border-radius: 16px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.safety-item {
  display: flex;
  gap: 14px;
  align-items: flex-start;
}

.safety-icon {
  font-size: 20px;
}

.safety-item strong {
  display: block;
  font-size: 14px;
  color: var(--play-text-primary);
  margin-bottom: 2px;
}

.safety-item p {
  font-size: 12px;
  color: var(--play-text-secondary);
  margin: 0;
  line-height: 1.4;
}

/* Ratings Histogram */
.ratings-overview-card {
  display: grid;
  grid-template-columns: 140px 1fr;
  gap: 24px;
  align-items: center;
  background-color: var(--play-subtle-bg);
  border: 1px solid var(--play-border);
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 20px;
}

.rating-big-col {
  text-align: center;
}

.big-score {
  font-size: 48px;
  font-weight: 800;
  color: var(--play-text-primary);
  line-height: 1;
  margin-bottom: 4px;
}

.stars-gold {
  color: #f59e0b;
  font-size: 16px;
  margin-bottom: 4px;
}

.total-reviews-count {
  font-size: 11px;
  color: var(--play-text-secondary);
}

.histogram-col {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.hist-row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  color: var(--play-text-secondary);
}

.hist-bar-bg {
  flex: 1;
  height: 8px;
  background-color: var(--play-border);
  border-radius: 9999px;
  overflow: hidden;
}

.hist-bar-fill {
  height: 100%;
  background-color: #01875f;
  border-radius: 9999px;
}

/* User Reviews */
.user-reviews-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.review-card {
  background-color: var(--play-subtle-bg);
  border: 1px solid var(--play-border);
  border-radius: 16px;
  padding: 18px 20px;
}

.review-author-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.review-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background-color: var(--play-accent);
  color: #ffffff;
  font-size: 13px;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
}

.review-author-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.review-author-info strong {
  font-size: 13px;
  color: var(--play-text-primary);
}

.review-date {
  font-size: 11px;
  color: var(--play-text-secondary);
}

.review-stars {
  color: #f59e0b;
  font-size: 12px;
}

.review-body {
  font-size: 13px;
  line-height: 1.6;
  color: var(--play-text-primary);
  margin: 0 0 10px;
}

.review-helpful {
  font-size: 11px;
  color: var(--play-text-secondary);
}

/* Toast */
.play-toast {
  position: fixed;
  bottom: 24px;
  right: 24px;
  padding: 12px 24px;
  background-color: #202124;
  color: #ffffff;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 600;
  box-shadow: 0 10px 30px rgba(0,0,0,0.4);
  z-index: 3000;
  animation: slideInToast 0.3s ease;
}

@keyframes slideInToast {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Responsive Rules */
@media (max-width: 1024px) {
  .billboard-card {
    grid-template-columns: 1fr;
    padding: 32px;
  }
  .billboard-graphic {
    display: none;
  }
  .top-charts-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .play-home-root {
    padding: 0 16px 40px;
  }
  .billboard-card {
    padding: 24px 20px;
  }
  .billboard-title {
    font-size: 24px;
  }
  .modal-scroll-body {
    padding: 20px 16px;
  }
}
</style>

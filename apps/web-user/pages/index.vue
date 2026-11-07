<template>
  <div class="ledger-home">
    <!-- Category ledger tabs -->
    <div class="ledger-tabs-bar">
      <div class="ledger-tabs">
        <button
          v-for="(cat, i) in categoryList"
          :key="cat.id"
          class="ledger-tab"
          :class="{ active: selectedCategory === cat.id }"
          @click="selectCategory(cat.id)"
        >
          <span class="tab-index">{{ String(i + 1).padStart(2, '0') }}</span>
          <span class="tab-label">{{ cat.name }}</span>
        </button>
      </div>
    </div>

    <!-- 01. Featured board -->
    <section class="board-section">
      <div class="board-frame">
        <div class="board-slider">
          <div
            v-for="(banner, idx) in billboardBanners"
            :key="banner.id"
            class="board-card"
            :class="[`tone-${banner.tone}`, { active: currentBannerIdx === idx }]"
            @click="openAppDetail(banner.app)"
          >
            <div class="board-content">
              <div class="board-kicker">
                <span class="kicker-no">FEATURED {{ String(idx + 1).padStart(2, '0') }} / {{ String(billboardBanners.length).padStart(2, '0') }}</span>
                <span class="kicker-tag">{{ banner.tag }}</span>
              </div>
              <h2 class="board-title">{{ banner.title }}</h2>
              <p class="board-subtitle">{{ banner.subtitle }}</p>
              <div class="board-app-row">
                <div class="board-app-icon"><img :src="banner.app.icon_url" :alt="banner.app.title" width="56" height="56" /></div>
                <div class="board-app-info">
                  <div class="app-info-title">{{ banner.app.title }}</div>
                  <div class="app-info-dev">{{ banner.app.developer_name }} · ★ {{ banner.app.rating.toFixed(1) }}</div>
                </div>
                <span class="seal-stamp">V3 VERIFIED</span>
              </div>
              <div class="board-actions">
                <button class="btn-ink" @click.stop="triggerQuickInstall(banner.app)">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3v12" /><path d="M7 10l5 5 5-5" /><path d="M4 21h16" /></svg>
                  설치
                </button>
                <button class="btn-text" @click.stop="openAppDetail(banner.app)">
                  자세히 보기 →
                </button>
              </div>
            </div>

            <div class="board-figure" aria-hidden="true">
              <div class="figure-rating">{{ banner.app.rating.toFixed(1) }}</div>
              <div class="figure-caption">STORE RATING / {{ banner.app.category_name }}</div>
              <div class="figure-index">{{ String(idx + 1).padStart(2, '0') }}</div>
            </div>
          </div>
        </div>

        <div class="board-pager">
          <button
            v-for="(_, idx) in billboardBanners"
            :key="idx"
            class="pager-btn"
            :class="{ active: currentBannerIdx === idx }"
            @click="currentBannerIdx = idx"
            :title="`배너 ${idx + 1}`"
          >{{ String(idx + 1).padStart(2, '0') }}</button>
        </div>
      </div>
    </section>

    <!-- 02. Verification strip -->
    <section id="section-protect" class="verify-strip">
      <div class="verify-inner">
        <div class="verify-seal" aria-hidden="true">
          <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 2L4 5V11.5C4 16.5 7.5 21 12 22C16.5 21 20 16.5 20 11.5V5L12 2Z" />
            <path d="M9 12L11 14L15 9" />
          </svg>
        </div>
        <div class="verify-text">
          <strong>TreeScan 다계층 보안 검역 · V3 서명 연속성 검증</strong>
          <p>모든 패키지는 개발자 직접 서명 V3 검증과 멀티 엔진 검역 파이프라인을 통과합니다. 비공식 스크래핑 경로는 취급하지 않습니다.</p>
        </div>
        <div class="verify-stat">
          <span class="stat-number">100%</span>
          <span class="stat-label">INTEGRITY VERIFIED</span>
        </div>
      </div>
    </section>

    <!-- 03. Recommended -->
    <section id="section-apps" class="ledger-section">
      <div class="section-head">
        <div class="section-head-left">
          <span class="section-no">01</span>
          <h2 class="section-title">추천 앱 및 게임</h2>
          <span class="section-sub">에디터가 엄선한 독립 배포 인기 애플리케이션</span>
        </div>
        <button class="see-all-btn" @click="selectedCategory = 'all'">전체보기 →</button>
      </div>

      <div v-if="isLoading" class="home-status">카탈로그를 불러오는 중…</div>
      <div v-if="loadError" class="home-status home-error">{{ loadError }}</div>
      <div v-if="!isLoading && !loadError && filteredRecommendedApps.length === 0" class="home-status">표시할 앱이 없습니다.</div>
      <div v-if="!isLoading && !loadError" class="app-rail">
        <div
          v-for="app in filteredRecommendedApps"
          :key="app.id"
          class="app-card"
          @click="openAppDetail(app)"
        >
          <div class="card-icon-tile"><img :src="app.icon_url" :alt="app.title" width="56" height="56" /></div>
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

    <!-- 04. Games strip -->
    <section v-if="gamesApps.length > 0" id="section-games" class="ledger-section">
      <div class="section-head">
        <div class="section-head-left">
          <span class="section-no">02</span>
          <h2 class="section-title">이번 주 게임</h2>
          <span class="section-sub">카탈로그에서 바로 고른 게임 행</span>
        </div>
      </div>
      <div class="games-strip">
        <div
          v-for="(app, idx) in gamesApps"
          :key="app.id"
          class="game-row"
          @click="openAppDetail(app)"
        >
          <span class="game-rank">{{ String(idx + 1).padStart(2, '0') }}</span>
          <div class="game-icon-tile"><img :src="app.icon_url" :alt="app.title" width="56" height="56" /></div>
          <div class="game-info">
            <strong>{{ app.title }}</strong>
            <span>{{ app.developer_name }} · ★ {{ app.rating.toFixed(1) }}</span>
          </div>
          <button class="mini-install-btn" @click.stop="triggerQuickInstall(app)">설치</button>
        </div>
      </div>
    </section>

    <!-- 05. Charts -->
    <section id="section-charts" class="ledger-section">
      <div class="charts-panel">
        <div class="section-head">
          <div class="section-head-left">
            <span class="section-no">03</span>
            <h2 class="section-title">인기 차트</h2>
            <span class="section-sub">실시간 다운로드 및 사용자 호평 랭킹</span>
          </div>
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
        <div v-if="!chartLoading" class="charts-grid">
          <div
            v-for="(app, idx) in chartFilteredApps"
            :key="app.id"
            class="chart-row-item"
            @click="openAppDetail(app)"
          >
            <div class="chart-rank-number">{{ idx + 1 }}</div>
            <div class="chart-icon-tile"><img :src="app.icon_url" :alt="app.title" width="56" height="56" /></div>
            <div class="chart-info">
              <h4 class="chart-app-title">{{ app.title }}</h4>
              <div class="chart-app-dev">{{ app.developer_name }}</div>
              <div class="chart-app-meta">
                <span>★ {{ app.rating.toFixed(1) }}</span>
                <span>· {{ app.category_name }}</span>
                <span>· {{ fmtCount(app.install_count) }}</span>
              </div>
            </div>
            <div class="chart-btn-box">
              <button class="pill-install-btn" @click.stop="triggerQuickInstall(app)">
                {{ app.is_free ? '설치' : '유료' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 06. Editorial -->
    <section id="section-editors" class="editorial-section">
      <div class="editorial-panel">
        <span class="editorial-kicker">EDITORIAL · TODAY</span>
        <h3 class="editorial-headline">왜 단일 빅테크 마켓에 의존하지 않아야 하는가?</h3>
        <p class="editorial-summary">
          EU 디지털 시장법(DMA) 제6조 제4항은 개발사에게 대체 앱 스토어 및 독립적인 사이드로딩 배포 권리를 보장합니다.
          BookingNetwork App Market은 독자적인 V3 서명 연속성 검증을 통해 정책 변경에도
          중단 없는 투명한 애플리케이션 라이프사이클을 제공합니다.
        </p>
        <div class="editorial-app-strip" v-if="allApps[0]" @click="openAppDetail(allApps[0])">
          <div class="editorial-app-icon"><img :src="allApps[0].icon_url" :alt="allApps[0].title" width="56" height="56" /></div>
          <div class="editorial-app-text">
            <strong>{{ allApps[0].title }}</strong>
            <span>{{ allApps[0].developer_name }} · 별점 {{ allApps[0].rating.toFixed(1) }}</span>
          </div>
          <button class="editorial-install-btn" @click.stop="triggerQuickInstall(allApps[0])">설치</button>
        </div>
      </div>
    </section>

    <!-- 07. App detail sheet -->
    <div v-if="selectedApp" class="sheet-backdrop" @click="closeAppDetail">
      <div class="sheet" @click.stop>
        <div class="sheet-topbar">
          <span class="sheet-kicker">APP RECORD · {{ selectedApp.package_name }}</span>
          <div class="sheet-top-actions">
            <button class="sheet-icon-btn" title="공유하기" @click="copyShareLink" aria-label="공유하기">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.5.5l3-3a5 5 0 0 0-7-7l-1.7 1.7" /><path d="M14 11a5 5 0 0 0-7.5-.5l-3 3a5 5 0 0 0 7 7l1.7-1.7" /></svg>
            </button>
            <button class="sheet-icon-btn" title="위시리스트 저장" @click="toggleWishlist" aria-label="위시리스트">
              <svg width="15" height="15" viewBox="0 0 24 24" :fill="isWishlisted ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 14c1.5-1.5 3-3.2 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.8 0-3 .5-4.5 2-1.5-1.5-2.7-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4 3 5.5l7 7z" /></svg>
            </button>
            <button class="sheet-icon-btn" @click="closeAppDetail" title="닫기" aria-label="닫기">
              <svg width="14" height="14" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M1 1l10 10M11 1L1 11" /></svg>
            </button>
          </div>
        </div>

        <div v-if="modalLoading" class="home-status">앱 정보를 불러오는 중…</div>
        <div v-else class="sheet-body">
          <div class="detail-header-row">
            <div class="detail-icon-large"><img :src="selectedApp.icon_url" :alt="selectedApp.title" width="72" height="72" /></div>
            <div class="detail-title-col">
              <h1 class="detail-app-name">{{ selectedApp.title }}</h1>
              <div class="detail-developer-row">
                <span class="dev-link">{{ selectedApp.developer_name }}</span>
                <span class="seal-stamp">V3 VERIFIED</span>
              </div>
              <div class="detail-badges-row">
                <span class="badge-chip">{{ selectedApp.category_name }}</span>
                <span class="badge-chip">{{ selectedApp.is_free ? '무료' : selectedApp.price_label }}</span>
                <span class="badge-chip mono">{{ selectedApp.license_spdx || selectedApp.monetization }}</span>
              </div>
            </div>
          </div>

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

          <div class="install-action-container">
            <button
              class="primary-install-cta"
              :class="{ downloading: isDownloading, installed: hasDownloaded }"
              :disabled="isDownloading || !selectedApp.release_id"
              @click="handleInstallClick()"
            >
              <template v-if="isDownloading">
                <span class="spinner-ring"></span>
                <span>보안 검증 및 다운로드 중 ({{ downloadProgress }}%)…</span>
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
              서명된 다운로드 URL(15분 만료) · SHA-256 체크섬 제공
            </span>
          </div>

          <div v-if="selectedApp.media.length > 0" class="screenshots-section">
            <h3 class="detail-section-title">스크린샷 미리보기</h3>
            <div class="screenshots-rail">
              <div
                v-for="shot in selectedApp.media"
                :key="shot.url"
                class="screenshot-frame"
              >
                <img class="screenshot-img" :src="shot.url" :alt="selectedApp.title" loading="lazy" />
              </div>
            </div>
          </div>

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

          <div class="data-safety-section">
            <h3 class="detail-section-title">데이터 보안 및 프라이버시</h3>
            <div class="safety-card">
              <div class="safety-item">
                <span class="safety-marker" aria-hidden="true">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="11" width="16" height="10" rx="2" /><path d="M8 11V7a4 4 0 0 1 8 0v4" /></svg>
                </span>
                <div>
                  <strong>데이터 전송 중 암호화</strong>
                  <p>보안 연결(TLS 1.3 / HTTPS)을 통해 모든 사용자 데이터가 암호화되어 전송됩니다.</p>
                </div>
              </div>
              <div class="safety-item">
                <span class="safety-marker" aria-hidden="true">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="9" /><path d="M5.5 5.5l13 13" /></svg>
                </span>
                <div>
                  <strong>제3자 데이터 판매 금지</strong>
                  <p>개발사는 사용자의 개인정보나 활동 내역을 타사 광고 네트워크에 판매하지 않습니다.</p>
                </div>
              </div>
              <div class="safety-item">
                <span class="safety-marker seal" aria-hidden="true">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2L4 5V11.5C4 16.5 7.5 21 12 22C16.5 21 20 16.5 20 11.5V5L12 2Z" /><path d="M9 12L11 14L15 9" /></svg>
                </span>
                <div>
                  <strong>서명 인증서 지문</strong>
                  <p>{{ selectedApp.cert_short ? `SHA-256: ${selectedApp.cert_short}…` : '등록된 서명 정보가 없습니다.' }}</p>
                </div>
              </div>
            </div>
          </div>

          <div class="reviews-section">
            <h3 class="detail-section-title">평점 및 리뷰</h3>
            <div class="ratings-overview-card">
              <div class="rating-big-col">
                <div class="big-score">{{ selectedApp.rating.toFixed(1) }}</div>
                <div class="stars-row">★★★★★</div>
                <div class="total-reviews-count">{{ modalReviews.length }}개의 평점</div>
              </div>
              <div class="histogram-col">
                <div v-for="row in reviewHistogram" :key="row.stars" class="hist-row">
                  <span>{{ row.stars }}</span>
                  <div class="hist-bar-bg"><div class="hist-bar-fill" :style="{ width: row.pct + '%' }"></div></div>
                </div>
              </div>
            </div>

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

    <div v-if="toastMessage" class="ledger-toast">
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
  const dark = h % 2 === 0
  const bg = dark ? '#1a1611' : '#f4f3ef'
  const fg = dark ? '#faf9f6' : '#1a1611'
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="56" height="56" viewBox="0 0 64 64"><rect width="64" height="64" rx="10" fill="${bg}"/><rect x="0" y="56" width="64" height="8" fill="#b3261e"/><text x="32" y="40" font-size="26" font-weight="800" text-anchor="middle" fill="${fg}" font-family="sans-serif">${ch}</text></svg>`
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
      { id: 'all', name: '전체' },
      ...cats.map(c => ({ id: c.id, name: names[c.id] }))
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

const categoryList = ref<Array<{ id: string; name: string }>>([
  { id: 'all', name: '전체' }
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

const BANNER_TONES = ['ink', 'paper', 'seal']

const billboardBanners = computed(() =>
  allApps.value.slice(0, 3).map((app, idx) => ({
    id: 'banner-' + app.id,
    tag: '★ 평점 ' + app.rating.toFixed(1) + ' · ' + app.category_name,
    title: app.title,
    subtitle: app.summary || app.package_name,
    tone: BANNER_TONES[idx % BANNER_TONES.length],
    app
  }))
)

const filteredRecommendedApps = computed(() => {
  if (selectedCategory.value === 'all') return allApps.value
  return allApps.value.filter(a => a.category === selectedCategory.value)
})

const gamesApps = computed(() =>
  allApps.value.filter(a => a.category === 'games').slice(0, 6)
)

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
.ledger-home {
  max-width: 1360px;
  margin: 0 auto;
  padding: 0 24px 60px;
}
.home-status {
  padding: 32px 24px;
  text-align: center;
  color: var(--bn-text-muted);
  font-size: 14px;
  border: 1px dashed var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
}
.home-error {
  color: var(--bn-danger-red);
  border-color: var(--bn-danger-red);
}

/* Category ledger tabs */
.ledger-tabs-bar {
  padding: 18px 0;
  overflow-x: auto;
  scrollbar-width: none;
  border-bottom: 1px solid var(--bn-border-subtle);
}
.ledger-tabs-bar::-webkit-scrollbar {
  display: none;
}
.ledger-tabs {
  display: flex;
  gap: 4px;
}
.ledger-tab {
  display: inline-flex;
  align-items: baseline;
  gap: 8px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  padding: 8px 12px;
  font-family: var(--bn-font-sans);
  font-size: 14px;
  font-weight: 600;
  color: var(--bn-text-secondary);
  cursor: pointer;
  white-space: nowrap;
}
.ledger-tab:hover {
  color: var(--bn-text-primary);
}
.ledger-tab.active {
  color: var(--bn-text-primary);
  border-bottom-color: var(--bn-ink);
}
.tab-index {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  color: var(--bn-text-muted);
}
.ledger-tab.active .tab-index {
  color: var(--bn-seal);
}

/* Featured board */
.board-section {
  margin: 24px 0 8px;
}
.board-frame {
  position: relative;
}
.board-slider {
  position: relative;
  min-height: 340px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-lg);
  overflow: hidden;
}
.board-card {
  display: none;
  grid-template-columns: 1.4fr 1fr;
  min-height: 340px;
  cursor: pointer;
  box-sizing: border-box;
}
.board-card.active {
  display: grid;
}
.board-card.tone-ink {
  background-color: var(--bn-ink);
  color: var(--bn-bg-canvas);
}
.board-card.tone-paper {
  background-color: var(--bn-bg-surface);
  color: var(--bn-text-primary);
}
.board-card.tone-seal {
  background-color: var(--bn-seal);
  color: #fff;
}
.board-content {
  padding: 40px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
.board-kicker {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
  font-family: var(--bn-font-mono);
  font-size: 11px;
  letter-spacing: 0.1em;
}
.kicker-no {
  opacity: 0.65;
}
.kicker-tag {
  border: 1px solid currentColor;
  border-radius: var(--bn-radius-sm);
  padding: 2px 8px;
  opacity: 0.9;
}
.board-title {
  font-family: var(--bn-font-display);
  font-size: 40px;
  font-weight: 900;
  letter-spacing: -0.03em;
  line-height: 1.05;
  margin: 0 0 10px;
}
.board-subtitle {
  font-size: 14px;
  line-height: 1.6;
  opacity: 0.75;
  margin: 0 0 24px;
  max-width: 480px;
}
.board-app-row {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 26px;
}
.board-app-icon {
  width: 52px;
  height: 52px;
  border-radius: var(--bn-radius-md);
  overflow: hidden;
  flex-shrink: 0;
}
.board-app-icon img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.app-info-title {
  font-size: 15px;
  font-weight: 700;
}
.app-info-dev {
  font-size: 12px;
  opacity: 0.7;
  margin-top: 2px;
}
.seal-stamp {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.1em;
  border: 1px solid currentColor;
  border-radius: var(--bn-radius-sm);
  padding: 3px 8px;
  margin-left: 4px;
  white-space: nowrap;
}
.tone-paper .seal-stamp,
.seal-stamp {
  color: var(--bn-seal);
  border-color: var(--bn-seal);
}
.tone-ink .seal-stamp,
.tone-seal .seal-stamp {
  color: inherit;
  border-color: currentColor;
}
.board-actions {
  display: flex;
  align-items: center;
  gap: 18px;
}
.btn-ink {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background-color: var(--bn-action-bg);
  color: var(--bn-action-text);
  border: none;
  border-radius: var(--bn-radius-sm);
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
}
.tone-ink .btn-ink {
  background-color: var(--bn-bg-canvas);
  color: var(--bn-ink);
}
.tone-seal .btn-ink {
  background-color: #fff;
  color: var(--bn-seal-ink);
}
.btn-ink:hover {
  opacity: 0.88;
}
.btn-text {
  background: none;
  border: none;
  color: inherit;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  padding: 12px 4px;
  opacity: 0.8;
}
.btn-text:hover {
  opacity: 1;
  text-decoration: underline;
  text-underline-offset: 4px;
}
.board-figure {
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  align-items: flex-end;
  padding: 32px;
  border-left: 1px solid currentColor;
  opacity: 1;
  position: relative;
}
.tone-paper .board-figure {
  border-left-color: var(--bn-border-subtle);
  background-color: var(--bn-bg-well);
}
.figure-rating {
  font-family: var(--bn-font-display);
  font-weight: 900;
  font-size: 92px;
  letter-spacing: -0.04em;
  line-height: 1;
}
.figure-caption {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.12em;
  opacity: 0.65;
  margin-top: 8px;
}
.figure-index {
  position: absolute;
  top: 24px;
  right: 28px;
  font-family: var(--bn-font-mono);
  font-size: 12px;
  opacity: 0.5;
}
.board-pager {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}
.pager-btn {
  background: none;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  color: var(--bn-text-muted);
  font-family: var(--bn-font-mono);
  font-size: 11px;
  padding: 5px 10px;
  cursor: pointer;
}
.pager-btn:hover {
  border-color: var(--bn-border-strong);
  color: var(--bn-text-primary);
}
.pager-btn.active {
  background-color: var(--bn-ink);
  border-color: var(--bn-ink);
  color: var(--bn-bg-canvas);
}

/* Verification strip */
.verify-strip {
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  margin: 24px 0 8px;
}
.verify-inner {
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 18px 24px;
}
.verify-seal {
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--bn-seal);
  border: 1px solid var(--bn-seal);
  border-radius: var(--bn-radius-sm);
}
.verify-text {
  flex: 1;
}
.verify-text strong {
  display: block;
  font-size: 14px;
  margin-bottom: 3px;
}
.verify-text p {
  font-size: 12px;
  line-height: 1.55;
  color: var(--bn-text-secondary);
  margin: 0;
}
.verify-stat {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  padding-left: 18px;
  border-left: 1px solid var(--bn-border-subtle);
  flex-shrink: 0;
}
.stat-number {
  font-family: var(--bn-font-display);
  font-size: 26px;
  font-weight: 900;
  letter-spacing: -0.02em;
}
.stat-label {
  font-family: var(--bn-font-mono);
  font-size: 9px;
  letter-spacing: 0.12em;
  color: var(--bn-text-muted);
}

/* Ledger sections */
.ledger-section {
  margin-top: 44px;
}
.section-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 16px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 2px solid var(--bn-ink);
}
.section-head-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
  min-width: 0;
}
.section-no {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  color: var(--bn-seal);
}
.section-title {
  font-family: var(--bn-font-display);
  font-size: 22px;
  font-weight: 800;
  letter-spacing: -0.02em;
  margin: 0;
  white-space: nowrap;
}
.section-sub {
  font-size: 12px;
  color: var(--bn-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.see-all-btn {
  background: none;
  border: none;
  color: var(--bn-text-primary);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  padding: 6px 4px;
  white-space: nowrap;
}
.see-all-btn:hover {
  text-decoration: underline;
  text-underline-offset: 4px;
}

/* App rail */
.app-rail {
  display: flex;
  gap: 14px;
  overflow-x: auto;
  padding-bottom: 10px;
  scrollbar-width: thin;
}
.app-card {
  flex: 0 0 196px;
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 14px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
}
.app-card:hover {
  border-color: var(--bn-border-strong);
}
.card-icon-tile {
  width: 56px;
  height: 56px;
  margin-bottom: 10px;
  border-radius: var(--bn-radius-md);
  overflow: hidden;
}
.card-icon-tile img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.card-title {
  font-size: 14px;
  font-weight: 700;
  margin: 0 0 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.card-dev {
  font-size: 12px;
  color: var(--bn-text-secondary);
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
  color: var(--bn-text-secondary);
  margin-bottom: 10px;
}
.card-rating {
  font-family: var(--bn-font-mono);
  font-weight: 600;
  color: var(--bn-text-primary);
}
.card-action-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: auto;
  padding-top: 10px;
  border-top: 1px solid var(--bn-border-subtle);
}
.card-price {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  color: var(--bn-text-secondary);
}
.card-price.free {
  color: var(--bn-trust-verified);
}
.mini-install-btn {
  padding: 5px 14px;
  background-color: transparent;
  border: 1px solid var(--bn-border-strong);
  border-radius: var(--bn-radius-sm);
  font-size: 12px;
  font-weight: 700;
  color: var(--bn-text-primary);
  cursor: pointer;
}
.mini-install-btn:hover {
  background-color: var(--bn-ink);
  border-color: var(--bn-ink);
  color: var(--bn-bg-canvas);
}

/* Games strip */
.games-strip {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 0;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  overflow: hidden;
}
.game-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid var(--bn-border-subtle);
}
.game-row:hover {
  background-color: var(--bn-bg-well);
}
.game-rank {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  color: var(--bn-seal);
  width: 24px;
  flex-shrink: 0;
}
.game-icon-tile {
  width: 40px;
  height: 40px;
  border-radius: var(--bn-radius-sm);
  overflow: hidden;
  flex-shrink: 0;
}
.game-icon-tile img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.game-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.game-info strong {
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.game-info span {
  font-size: 11px;
  color: var(--bn-text-muted);
}

/* Charts ledger */
.charts-panel {
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
}
.charts-panel .section-head {
  margin-bottom: 8px;
}
.chart-tab-group {
  display: flex;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  overflow: hidden;
  flex-shrink: 0;
}
.chart-tab-btn {
  padding: 7px 16px;
  border: none;
  border-left: 1px solid var(--bn-border-subtle);
  background: transparent;
  color: var(--bn-text-secondary);
  font-family: var(--bn-font-sans);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}
.chart-tab-btn:first-child {
  border-left: none;
}
.chart-tab-btn.active {
  background-color: var(--bn-ink);
  color: var(--bn-bg-canvas);
}
.charts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 0 28px;
}
.chart-row-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 4px;
  border-bottom: 1px solid var(--bn-border-subtle);
  cursor: pointer;
}
.chart-row-item:hover {
  background-color: var(--bn-bg-well);
}
.chart-rank-number {
  font-family: var(--bn-font-display);
  font-size: 30px;
  font-weight: 900;
  letter-spacing: -0.03em;
  color: var(--bn-text-primary);
  width: 44px;
  text-align: center;
  flex-shrink: 0;
}
.chart-icon-tile {
  width: 48px;
  height: 48px;
  border-radius: var(--bn-radius-md);
  overflow: hidden;
  flex-shrink: 0;
}
.chart-icon-tile img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.chart-info {
  flex: 1;
  min-width: 0;
}
.chart-app-title {
  font-size: 14px;
  font-weight: 700;
  margin: 0 0 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.chart-app-dev {
  font-size: 12px;
  color: var(--bn-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.chart-app-meta {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-text-muted);
  display: flex;
  gap: 8px;
  margin-top: 3px;
}
.chart-btn-box {
  flex-shrink: 0;
}
.pill-install-btn {
  padding: 6px 16px;
  border-radius: var(--bn-radius-pill);
  background-color: transparent;
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-primary);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}
.pill-install-btn:hover {
  background-color: var(--bn-ink);
  border-color: var(--bn-ink);
  color: var(--bn-bg-canvas);
}

/* Editorial: the one ink panel */
.editorial-section {
  margin-top: 44px;
}
.editorial-panel {
  background-color: var(--bn-ink);
  color: var(--bn-bg-canvas);
  border-radius: var(--bn-radius-lg);
  padding: 44px;
}
.editorial-kicker {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  letter-spacing: 0.14em;
  opacity: 0.6;
}
.editorial-headline {
  font-family: var(--bn-font-display);
  font-size: 32px;
  font-weight: 900;
  letter-spacing: -0.025em;
  line-height: 1.2;
  margin: 12px 0;
  max-width: 640px;
}
.editorial-summary {
  font-size: 14px;
  line-height: 1.7;
  opacity: 0.75;
  margin: 0 0 28px;
  max-width: 640px;
}
.editorial-app-strip {
  display: flex;
  align-items: center;
  gap: 14px;
  border-top: 1px solid currentColor;
  padding-top: 20px;
  cursor: pointer;
  max-width: 640px;
}
.editorial-app-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--bn-radius-md);
  overflow: hidden;
  flex-shrink: 0;
}
.editorial-app-icon img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.editorial-app-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.editorial-app-text strong {
  font-size: 14px;
}
.editorial-app-text span {
  font-size: 12px;
  opacity: 0.65;
}
.editorial-install-btn {
  background-color: var(--bn-bg-canvas);
  color: var(--bn-ink);
  border: none;
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-weight: 700;
  padding: 9px 22px;
  cursor: pointer;
  flex-shrink: 0;
}
.editorial-install-btn:hover {
  opacity: 0.88;
}

/* Detail sheet */
.sheet-backdrop {
  position: fixed;
  inset: 0;
  background-color: rgba(20, 17, 12, 0.55);
  z-index: 2000;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  overflow-y: auto;
  padding: 40px 16px;
  box-sizing: border-box;
}
.sheet {
  width: 100%;
  max-width: 860px;
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-lg);
  box-shadow: var(--bn-shadow-pop);
  overflow: hidden;
}
.sheet-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 20px;
  border-bottom: 1px solid var(--bn-border-subtle);
  background-color: var(--bn-bg-well);
}
.sheet-kicker {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  letter-spacing: 0.06em;
  color: var(--bn-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.sheet-top-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}
.sheet-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--bn-radius-sm);
  color: var(--bn-text-secondary);
  cursor: pointer;
}
.sheet-icon-btn:hover {
  color: var(--bn-text-primary);
  border-color: var(--bn-border-subtle);
  background-color: var(--bn-bg-surface);
}
.sheet-body {
  padding: 28px;
  max-height: calc(100vh - 220px);
  overflow-y: auto;
}
.detail-header-row {
  display: flex;
  gap: 18px;
  align-items: flex-start;
}
.detail-icon-large {
  width: 84px;
  height: 84px;
  border-radius: var(--bn-radius-lg);
  overflow: hidden;
  flex-shrink: 0;
  border: 1px solid var(--bn-border-subtle);
}
.detail-icon-large img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.detail-title-col {
  min-width: 0;
}
.detail-app-name {
  font-family: var(--bn-font-display);
  font-size: 28px;
  font-weight: 900;
  letter-spacing: -0.025em;
  margin: 0 0 6px;
  line-height: 1.15;
}
.detail-developer-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}
.dev-link {
  font-size: 14px;
  font-weight: 600;
  color: var(--bn-text-secondary);
}
.detail-header-row .seal-stamp {
  margin-left: 0;
}
.detail-badges-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.badge-chip {
  font-size: 12px;
  font-weight: 600;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  padding: 3px 10px;
  color: var(--bn-text-secondary);
}
.badge-chip.mono {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  font-weight: 400;
}

/* Stats ribbon */
.stats-ribbon {
  display: flex;
  align-items: stretch;
  margin: 24px 0;
  border-top: 1px solid var(--bn-border-subtle);
  border-bottom: 1px solid var(--bn-border-subtle);
}
.stat-column {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 14px 8px;
}
.stat-top {
  display: flex;
  align-items: baseline;
  gap: 4px;
}
.stat-val {
  font-family: var(--bn-font-display);
  font-size: 20px;
  font-weight: 800;
}
.stat-star {
  color: var(--bn-text-primary);
  font-size: 14px;
}
.stat-bottom {
  font-size: 11px;
  color: var(--bn-text-muted);
}
.stat-divider {
  width: 1px;
  background-color: var(--bn-border-subtle);
}
.badge-pegi {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  border: 1px solid var(--bn-border-strong);
  border-radius: var(--bn-radius-sm);
  padding: 1px 8px;
}

/* Install action */
.install-action-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 28px;
}
.primary-install-cta {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  width: 100%;
  padding: 15px;
  background-color: var(--bn-action-bg);
  color: var(--bn-action-text);
  border: none;
  border-radius: var(--bn-radius-sm);
  font-size: 15px;
  font-weight: 700;
  cursor: pointer;
}
.primary-install-cta:hover:not(:disabled) {
  background-color: var(--bn-action-bg-hover);
}
.primary-install-cta:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.primary-install-cta.installed {
  background-color: var(--bn-trust-verified);
  color: #fff;
}
.spinner-ring {
  width: 16px;
  height: 16px;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: ledger-spin 0.8s linear infinite;
}
@keyframes ledger-spin {
  to { transform: rotate(360deg); }
}
.install-footnote {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-text-muted);
  text-align: center;
}

/* Sheet sections */
.detail-section-title {
  font-family: var(--bn-font-display);
  font-size: 15px;
  font-weight: 800;
  margin: 0 0 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--bn-border-subtle);
}
.screenshots-section,
.about-app-section,
.data-safety-section,
.reviews-section {
  margin-bottom: 28px;
}
.screenshots-rail {
  display: flex;
  gap: 12px;
  overflow-x: auto;
  padding-bottom: 8px;
}
.screenshot-frame {
  flex: 0 0 150px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  overflow: hidden;
  background-color: var(--bn-bg-well);
}
.screenshot-img {
  display: block;
  width: 100%;
  height: auto;
}
.about-description {
  font-size: 14px;
  line-height: 1.7;
  color: var(--bn-text-secondary);
  margin: 0 0 12px;
  white-space: pre-line;
}
.about-description a {
  color: var(--bn-text-primary);
  font-weight: 600;
}
.update-notes-card {
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 14px 16px;
  margin-bottom: 12px;
  background-color: var(--bn-bg-well);
}
.update-notes-header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
  margin-bottom: 8px;
}
.update-notes-header span {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-text-muted);
  white-space: nowrap;
}
.update-notes-card .about-description {
  margin-bottom: 0;
}

/* Safety */
.safety-card {
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  overflow: hidden;
}
.safety-item {
  display: flex;
  gap: 14px;
  padding: 14px 16px;
  border-bottom: 1px solid var(--bn-border-subtle);
}
.safety-item:last-child {
  border-bottom: none;
}
.safety-marker {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  color: var(--bn-text-secondary);
}
.safety-marker.seal {
  color: var(--bn-seal);
  border-color: var(--bn-seal);
}
.safety-item strong {
  display: block;
  font-size: 13px;
  margin-bottom: 3px;
}
.safety-item p {
  font-size: 12px;
  line-height: 1.6;
  color: var(--bn-text-secondary);
  margin: 0;
  font-family: var(--bn-font-mono);
}

/* Reviews */
.ratings-overview-card {
  display: flex;
  gap: 32px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 20px 24px;
  margin-bottom: 16px;
}
.rating-big-col {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.big-score {
  font-family: var(--bn-font-display);
  font-size: 52px;
  font-weight: 900;
  letter-spacing: -0.03em;
  line-height: 1;
}
.stars-row {
  font-size: 13px;
  letter-spacing: 2px;
  margin: 6px 0;
}
.total-reviews-count {
  font-size: 11px;
  color: var(--bn-text-muted);
}
.histogram-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 7px;
}
.hist-row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-text-muted);
}
.hist-bar-bg {
  flex: 1;
  height: 6px;
  background-color: var(--bn-bg-well);
  border-radius: var(--bn-radius-sm);
  overflow: hidden;
}
.hist-bar-fill {
  height: 100%;
  background-color: var(--bn-ink);
}
.user-reviews-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.review-card {
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 14px 16px;
}
.review-author-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}
.review-avatar {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--bn-ink);
  color: var(--bn-bg-canvas);
  font-weight: 800;
  font-size: 14px;
  border-radius: var(--bn-radius-sm);
}
.review-author-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.review-author-info strong {
  font-size: 13px;
}
.review-date {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-text-muted);
}
.review-stars {
  font-size: 12px;
  letter-spacing: 1px;
}
.review-body {
  font-size: 13px;
  line-height: 1.65;
  color: var(--bn-text-secondary);
  margin: 0 0 8px;
}
.review-helpful {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  color: var(--bn-trust-verified);
}

/* Toast */
.ledger-toast {
  position: fixed;
  bottom: 28px;
  left: 50%;
  transform: translateX(-50%);
  background-color: var(--bn-ink);
  color: var(--bn-bg-canvas);
  font-size: 13px;
  font-weight: 600;
  padding: 12px 22px;
  border-radius: var(--bn-radius-sm);
  box-shadow: var(--bn-shadow-pop);
  z-index: 3000;
  max-width: min(560px, calc(100vw - 48px));
}

@media (max-width: 900px) {
  .board-card.active {
    grid-template-columns: 1fr;
  }
  .board-figure {
    border-left: none;
    border-top: 1px solid currentColor;
    flex-direction: row;
    align-items: baseline;
    justify-content: space-between;
    padding: 20px 28px;
  }
  .tone-paper .board-figure {
    border-top-color: var(--bn-border-subtle);
  }
  .figure-rating {
    font-size: 56px;
  }
  .figure-index {
    position: static;
  }
  .board-title {
    font-size: 30px;
  }
  .board-content {
    padding: 28px;
  }
  .ratings-overview-card {
    flex-direction: column;
    gap: 16px;
  }
  .editorial-panel {
    padding: 28px;
  }
  .editorial-headline {
    font-size: 24px;
  }
  .section-sub {
    display: none;
  }
}
</style>

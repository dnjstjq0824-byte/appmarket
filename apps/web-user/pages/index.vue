<template>
  <div class="play-home-root">
    <!-- Sub-Navbar Category Pills (Google Play Style Horizontal Scroll) -->
    <div class="category-pills-bar">
      <div class="category-pills-container">
        <button
          v-for="cat in categories"
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
                <span class="tag-verified">🛡️ Google Play Protect & TreeScan™ 인증</span>
              </div>
              <h2 class="billboard-title">{{ banner.title }}</h2>
              <p class="billboard-subtitle">{{ banner.subtitle }}</p>
              <div class="billboard-meta-row">
                <div class="billboard-app-icon" v-html="banner.app.iconSvg"></div>
                <div class="billboard-app-info">
                  <div class="app-info-title">{{ banner.app.title }}</div>
                  <div class="app-info-dev">{{ banner.app.developer_name }} • ★ {{ banner.app.rating }}</div>
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
                <div class="mockup-screen" v-html="banner.graphicSvg"></div>
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

    <!-- 2. Google Play Protect Security & Trust Guarantee Strip -->
    <section class="play-protect-strip">
      <div class="protect-inner">
        <div class="protect-icon-shield">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none">
            <path d="M12 2L4 5V11.5C4 16.5 7.5 21 12 22C16.5 21 20 16.5 20 11.5V5L12 2Z" fill="#01875F" fill-opacity="0.2" stroke="#01875F" stroke-width="2"/>
            <path d="M9 12L11 14L15 9" stroke="#01875F" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <div class="protect-text">
          <strong>Google Play Protect & TreeScan™ 다계층 보안 검역</strong>
          <p>이 마켓의 모든 바이너리는 구글 플레이 스크래핑을 배제하고, 개발자 V3 서명 연속성 검증 및 70개 엔진 0-Day 악성코드 분석을 완료했습니다.</p>
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

      <div class="horizontal-scroll-container">
        <div
          v-for="app in filteredRecommendedApps"
          :key="app.id"
          class="play-app-card"
          @click="openAppDetail(app)"
        >
          <div class="card-icon-squircle" v-html="app.iconSvg"></div>
          <div class="card-content">
            <h3 class="card-title">{{ app.title }}</h3>
            <div class="card-dev">{{ app.developer_name }}</div>
            <div class="card-meta">
              <span class="card-rating">★ {{ app.rating }}</span>
              <span class="card-size">{{ app.size_formatted }}</span>
            </div>
            <div class="card-action-row">
              <span class="card-price" :class="{ free: app.is_free }">
                {{ app.is_free ? '무료' : `₩${app.price_krw.toLocaleString()}` }}
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
            @click="chartTab = 'free'"
          >
            인기 무료
          </button>
          <button
            class="chart-tab-btn"
            :class="{ active: chartTab === 'grossing' }"
            @click="chartTab = 'grossing'"
          >
            최고 매출
          </button>
          <button
            class="chart-tab-btn"
            :class="{ active: chartTab === 'trending' }"
            @click="chartTab = 'trending'"
          >
            급상승
          </button>
        </div>
      </div>

      <div class="top-charts-grid">
        <div
          v-for="(app, idx) in chartFilteredApps"
          :key="app.id"
          class="chart-row-item"
          @click="openAppDetail(app)"
        >
          <div class="chart-rank-number" :class="`rank-${idx + 1}`">{{ idx + 1 }}</div>
          <div class="chart-icon-squircle" v-html="app.iconSvg"></div>
          <div class="chart-info">
            <h4 class="chart-app-title">{{ app.title }}</h4>
            <div class="chart-app-dev">{{ app.developer_name }}</div>
            <div class="chart-app-meta">
              <span>★ {{ app.rating }}</span>
              <span>• {{ app.category_name }}</span>
              <span>• {{ app.downloads }}</span>
            </div>
          </div>
          <div class="chart-btn-box">
            <button class="pill-install-btn" @click.stop="triggerQuickInstall(app)">
              {{ app.is_free ? '설치' : `₩${app.price_krw.toLocaleString()}` }}
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
            EU 디지털 시장법(DMA) 23조는 개발사에게 독립적이고 자유로운 배포 권리를 보장합니다.
            BookingNetwork App Market은 독자적인 V3 서명 연속성 검증을 통해 구글 플레이 정책 변경에도
            중단 없는 투명한 애플리케이션 라이프사이클을 제공합니다.
          </p>
          <div class="editorial-app-strip" v-if="allApps[0]" @click="openAppDetail(allApps[0])">
            <div class="editorial-app-icon" v-html="allApps[0].iconSvg"></div>
            <div class="editorial-app-text">
              <strong>{{ allApps[0].title }}</strong>
              <span>{{ allApps[0].developer_name }} • 에디터 추천 별점 4.9</span>
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

        <div class="modal-scroll-body">
          <!-- App Profile Header -->
          <div class="detail-header-row">
            <div class="detail-icon-large" v-html="selectedApp.iconSvg"></div>
            <div class="detail-title-col">
              <h1 class="detail-app-name">{{ selectedApp.title }}</h1>
              <div class="detail-developer-row">
                <a href="#dev" class="dev-link">{{ selectedApp.developer_name }}</a>
                <span class="verified-check" title="인증된 합법 개발사">✓</span>
              </div>
              <div class="detail-badges-row">
                <span class="badge-category">{{ selectedApp.category_name }}</span>
                <span class="badge-iap">인앱 구매 제공</span>
                <span class="badge-signature">{{ selectedApp.signature_scheme }}</span>
              </div>
            </div>
          </div>

          <!-- Quick Statistics Ribbon (Google Play Style) -->
          <div class="stats-ribbon">
            <div class="stat-column">
              <div class="stat-top">
                <span class="stat-val">{{ selectedApp.rating }}</span>
                <span class="stat-star">★</span>
              </div>
              <div class="stat-bottom">{{ selectedApp.rating_count.toLocaleString() }}개 리뷰</div>
            </div>
            <div class="stat-divider"></div>
            <div class="stat-column">
              <div class="stat-top">
                <span class="stat-val">{{ selectedApp.downloads }}</span>
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
                <span class="badge-pegi">전체이용가</span>
              </div>
              <div class="stat-bottom">연령 등급</div>
            </div>
          </div>

          <!-- Install / Purchase Primary Action Bar -->
          <div class="install-action-container">
            <button
              class="primary-install-cta"
              :class="{ downloading: isDownloading, installed: isInstalled }"
              :disabled="isDownloading"
              @click="handleInstallClick(selectedApp)"
            >
              <template v-if="isDownloading">
                <span class="spinner-ring"></span>
                <span>보안 검증 및 다운로드 중 ({{ downloadProgress }}%)...</span>
              </template>
              <template v-else-if="isInstalled">
                <span>✓ 설치 완료 (열기)</span>
              </template>
              <template v-else>
                <span>{{ selectedApp.is_free ? '설치' : `₩${selectedApp.price_krw.toLocaleString()} 에 구매` }}</span>
              </template>
            </button>
            <span class="install-footnote">
              🔒 TreeScan™ 0-Day 악성코드 검역 완료 및 15분 만료 보안 스트리밍 토큰 적용
            </span>
          </div>

          <!-- Phone Screenshots Carousel (Google Play Style) -->
          <div class="screenshots-section">
            <h3 class="detail-section-title">스크린샷 미리보기</h3>
            <div class="screenshots-carousel">
              <div
                v-for="(screen, sIdx) in getAppScreenshots(selectedApp)"
                :key="sIdx"
                class="screenshot-phone-frame"
              >
                <div class="phone-screen-inner" v-html="screen"></div>
              </div>
            </div>
          </div>

          <!-- About This App (Detailed Description) -->
          <div class="about-app-section">
            <h3 class="detail-section-title">앱 정보</h3>
            <p class="about-description">{{ selectedApp.description }}</p>

            <div class="update-notes-card">
              <div class="update-notes-header">
                <strong>버전 {{ selectedApp.version_name }}의 새로운 기능</strong>
                <span>최근 업데이트: 2026년 10월 30일</span>
              </div>
              <ul class="update-bullets">
                <li>Android 15 (API 35) 및 16k Page Size 완전 대응 최적화</li>
                <li>TreeScan 정적 AST 바이트코드 무결성 서명 계통 갱신</li>
                <li>종단간 암호화(E2EE) 동기화 성능 40% 향상</li>
              </ul>
            </div>
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
                  <strong>V3 서명 연속성 보장</strong>
                  <p>인증서 계통이 변조되지 않았음을 확인하는 SHA-256 서명 지문 일치 검사를 완료했습니다.</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Ratings and Reviews Histogram (Google Play Style) -->
          <div class="reviews-section">
            <h3 class="detail-section-title">평점 및 리뷰</h3>
            <div class="ratings-overview-card">
              <div class="rating-big-col">
                <div class="big-score">{{ selectedApp.rating }}</div>
                <div class="stars-gold">★★★★★</div>
                <div class="total-reviews-count">{{ selectedApp.rating_count.toLocaleString() }}개의 평점</div>
              </div>
              <div class="histogram-col">
                <div class="hist-row">
                  <span>5</span>
                  <div class="hist-bar-bg"><div class="hist-bar-fill" style="width: 82%;"></div></div>
                </div>
                <div class="hist-row">
                  <span>4</span>
                  <div class="hist-bar-bg"><div class="hist-bar-fill" style="width: 12%;"></div></div>
                </div>
                <div class="hist-row">
                  <span>3</span>
                  <div class="hist-bar-bg"><div class="hist-bar-fill" style="width: 4%;"></div></div>
                </div>
                <div class="hist-row">
                  <span>2</span>
                  <div class="hist-bar-bg"><div class="hist-bar-fill" style="width: 1%;"></div></div>
                </div>
                <div class="hist-row">
                  <span>1</span>
                  <div class="hist-bar-bg"><div class="hist-bar-fill" style="width: 1%;"></div></div>
                </div>
              </div>
            </div>

            <!-- User Review Cards -->
            <div class="user-reviews-list">
              <div class="review-card">
                <div class="review-author-row">
                  <div class="review-avatar">김</div>
                  <div class="review-author-info">
                    <strong>김*호</strong>
                    <span class="review-date">2026년 10월 28일</span>
                  </div>
                  <div class="review-stars">★★★★★</div>
                </div>
                <p class="review-body">
                  구글 플레이에서 독립 배포 버전으로 갈아탔는데 광고 트래커도 전혀 없고 백그라운드 배터리 소모가 눈에 띄게 줄었습니다.
                  UI도 깔끔하고 무엇보다 무결성 서명 검증이 표시되어 안심하고 사용 중입니다.
                </p>
                <div class="review-helpful">이 리뷰가 유용했나요? • <strong>👍 342명</strong>이 유용하다고 평가함</div>
              </div>

              <div class="review-card">
                <div class="review-author-row">
                  <div class="review-avatar">P</div>
                  <div class="review-author-info">
                    <strong>Park J.</strong>
                    <span class="review-date">2026년 10월 24일</span>
                  </div>
                  <div class="review-stars">★★★★★</div>
                </div>
                <p class="review-body">
                  문서 편집 및 클라우드 동기화 속도가 타사 앱 대비 압도적입니다. 독립 마켓에서 이런 고품질 앱을 공식적으로 설치할 수 있어서 너무 좋습니다.
                </p>
                <div class="review-helpful">이 리뷰가 유용했나요? • <strong>👍 189명</strong>이 유용하다고 평가함</div>
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
import { ref, computed } from 'vue'

interface AppItem {
  id: string
  package_name: string
  title: string
  developer_name: string
  developer_verified: boolean
  category: string
  category_name: string
  rating: number
  rating_count: number
  downloads: string
  size_bytes: number
  size_formatted: string
  price_krw: number
  is_free: boolean
  min_sdk: number
  version_name: string
  version_code: number
  sha256: string
  summary: string
  description: string
  permissions: string[]
  signature_scheme: string
  verified_at: string
  is_sponsored?: boolean
  iconSvg: string
}

// Crisp Vector SVG Icons
const SVG_ICONS = {
  omni: `<svg width="56" height="56" viewBox="0 0 64 64" fill="none"><rect width="64" height="64" rx="16" fill="url(#omniG)"/><path d="M20 32L28 40L44 24" stroke="white" stroke-width="5" stroke-linecap="round" stroke-linejoin="round"/><defs><linearGradient id="omniG" x1="0" y1="0" x2="64" y2="64" gradientUnits="userSpaceOnUse"><stop stop-color="#0084FF"/><stop offset="1" stop-color="#00C9FF"/></linearGradient></defs></svg>`,
  finvault: `<svg width="56" height="56" viewBox="0 0 64 64" fill="none"><rect width="64" height="64" rx="16" fill="url(#finG)"/><path d="M32 16L46 22V34C46 43 39 49 32 52C25 49 18 43 18 34V22L32 16Z" fill="white" fill-opacity="0.2" stroke="white" stroke-width="4" stroke-linejoin="round"/><circle cx="32" cy="34" r="5" fill="white"/><defs><linearGradient id="finG" x1="0" y1="0" x2="64" y2="64" gradientUnits="userSpaceOnUse"><stop stop-color="#00A870"/><stop offset="1" stop-color="#00E575"/></linearGradient></defs></svg>`,
  messenger: `<svg width="56" height="56" viewBox="0 0 64 64" fill="none"><rect width="64" height="64" rx="16" fill="url(#msgG)"/><path d="M20 22C20 18.7 22.7 16 26 16H38C41.3 16 44 18.7 44 22V36C44 39.3 41.3 42 38 42H28L20 48V22Z" fill="white"/><circle cx="28" cy="29" r="2.5" fill="#4F46E5"/><circle cx="36" cy="29" r="2.5" fill="#4F46E5"/><defs><linearGradient id="msgG" x1="0" y1="0" x2="64" y2="64" gradientUnits="userSpaceOnUse"><stop stop-color="#6366F1"/><stop offset="1" stop-color="#A855F7"/></linearGradient></defs></svg>`,
  crypto: `<svg width="56" height="56" viewBox="0 0 64 64" fill="none"><rect width="64" height="64" rx="16" fill="url(#cryG)"/><path d="M22 42L32 28L38 34L46 22" stroke="white" stroke-width="4.5" stroke-linecap="round" stroke-linejoin="round"/><circle cx="46" cy="22" r="3" fill="white"/><defs><linearGradient id="cryG" x1="0" y1="0" x2="64" y2="64" gradientUnits="userSpaceOnUse"><stop stop-color="#F59E0B"/><stop offset="1" stop-color="#D97706"/></linearGradient></defs></svg>`,
  auth: `<svg width="56" height="56" viewBox="0 0 64 64" fill="none"><rect width="64" height="64" rx="16" fill="url(#authG)"/><path d="M32 20C27.6 20 24 23.6 24 28C24 31.2 25.9 34 28.7 35.3L26 44H38L35.3 35.3C38.1 34 40 31.2 40 28C40 23.6 36.4 20 32 20Z" fill="white"/><defs><linearGradient id="authG" x1="0" y1="0" x2="64" y2="64" gradientUnits="userSpaceOnUse"><stop stop-color="#0284C7"/><stop offset="1" stop-color="#38BDF8"/></linearGradient></defs></svg>`,
  editor: `<svg width="56" height="56" viewBox="0 0 64 64" fill="none"><rect width="64" height="64" rx="16" fill="url(#devG)"/><path d="M24 24L16 32L24 40M40 24L48 32L40 40M34 20L30 44" stroke="white" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/><defs><linearGradient id="devG" x1="0" y1="0" x2="64" y2="64" gradientUnits="userSpaceOnUse"><stop stop-color="#EF4444"/><stop offset="1" stop-color="#F97316"/></linearGradient></defs></svg>`,
  runner: `<svg width="56" height="56" viewBox="0 0 64 64" fill="none"><rect width="64" height="64" rx="16" fill="url(#gameG)"/><path d="M22 28V36M18 32H26M40 30A2 2 0 1 1 40 26A2 2 0 0 1 40 30ZM44 36A2 2 0 1 1 44 32A2 2 0 0 1 44 36Z" stroke="white" stroke-width="3" stroke-linecap="round"/><rect x="14" y="22" width="36" height="20" rx="6" stroke="white" stroke-width="3.5"/><defs><linearGradient id="gameG" x1="0" y1="0" x2="64" y2="64" gradientUnits="userSpaceOnUse"><stop stop-color="#8B5CF6"/><stop offset="1" stop-color="#EC4899"/></linearGradient></defs></svg>`
}

// 7 Pre-seeded Apps matching Google Play Store Quality
const allApps = ref<AppItem[]>([
  {
    id: 'app-omni-productivity-01',
    package_name: 'com.omniwork.office',
    title: 'Omni Productivity Suite',
    developer_name: 'OmniWork Technologies',
    developer_verified: true,
    category: 'productivity',
    category_name: '생산성 & 업무',
    rating: 4.9,
    rating_count: 42800,
    downloads: '150만+',
    size_bytes: 37539072,
    size_formatted: '35.8 MB',
    price_krw: 0,
    is_free: true,
    min_sdk: 26,
    version_name: '3.4.2',
    version_code: 30402,
    sha256: '9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08',
    summary: '차세대 올인원 문서 편집, 프로젝트 관리 및 엔드투엔드 암호화 협업 스위트',
    description: '문서 편집, 스프레드시트, 실시간 클라우드 동기화, 엔드투엔드 암호화 협업을 단 하나의 앱으로 완성하세요. 구글 플레이 독립 유통 공식 인증 패키지입니다.',
    permissions: ['INTERNET', 'READ_EXTERNAL_STORAGE', 'POST_NOTIFICATIONS'],
    signature_scheme: 'V3_CONTINUITY',
    verified_at: '2026-10-30T12:00:00Z',
    is_sponsored: true,
    iconSvg: SVG_ICONS.omni
  },
  {
    id: 'app-finvault-secure-02',
    package_name: 'com.finvault.banking',
    title: 'FinVault Secure Banking',
    developer_name: 'FinVault Financial Labs',
    developer_verified: true,
    category: 'finance',
    category_name: '금융 & 핀테크',
    rating: 4.8,
    rating_count: 28900,
    downloads: '82만+',
    size_bytes: 25690112,
    size_formatted: '24.5 MB',
    price_krw: 0,
    is_free: true,
    min_sdk: 28,
    version_name: '2.1.0',
    version_code: 20100,
    sha256: 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
    summary: '생체인식 하드웨어 보안 키 및 지능형 가계부 자산 통합 자산 관리 솔루션',
    description: '하드웨어 보안 모듈(HSM) 및 생체인증 FIDO2 기반 모바일 뱅킹 및 자산 관리 스위트.',
    permissions: ['INTERNET', 'USE_BIOMETRIC', 'CAMERA'],
    signature_scheme: 'V3_CONTINUITY',
    verified_at: '2026-10-29T08:30:00Z',
    is_sponsored: false,
    iconSvg: SVG_ICONS.finvault
  },
  {
    id: 'app-cloud-authenticator-05',
    package_name: 'com.shieldguard.auth2fa',
    title: 'Cloud Authenticator 2FA',
    developer_name: 'ShieldGuard Security',
    developer_verified: true,
    category: 'tools',
    category_name: '보안 & 인증',
    rating: 4.9,
    rating_count: 36700,
    downloads: '67만+',
    size_bytes: 13002342,
    size_formatted: '12.4 MB',
    price_krw: 0,
    is_free: true,
    min_sdk: 24,
    version_name: '4.0.1',
    version_code: 40001,
    sha256: '4b227777d4dd1fc61c6f884f48641d02b4d121d3fd328cb08b5531fcacdabf8a',
    summary: 'RFC 6238 표준 TOTP 2차 인증기. 기기간 암호화 백업 및 긴급 복구 지원',
    description: '표준 2단계 인증 코드 생성 및 종단간 암호화(E2EE) 클라우드 동기화 백업.',
    permissions: ['CAMERA', 'USE_BIOMETRIC'],
    signature_scheme: 'V3_CONTINUITY',
    verified_at: '2026-10-26T17:40:00Z',
    is_sponsored: false,
    iconSvg: SVG_ICONS.auth
  },
  {
    id: 'app-secure-messenger-03',
    package_name: 'org.openprivacy.messenger',
    title: 'Secure Messenger Pro',
    developer_name: 'Open Privacy Labs',
    developer_verified: true,
    category: 'communication',
    category_name: '커뮤니케이션',
    rating: 4.8,
    rating_count: 51200,
    downloads: '210만+',
    size_bytes: 29464985,
    size_formatted: '28.1 MB',
    price_krw: 0,
    is_free: true,
    min_sdk: 26,
    version_name: '5.8.0',
    version_code: 50800,
    sha256: 'ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb',
    summary: '영지식 암호화(ZKP) 기반 탈중앙화 메신저. 데이터 수집 제로 프라이버시',
    description: '서버에 어떠한 메시지나 메타데이터도 저장하지 않는 순수 P2P 영지식 암호화 메신저.',
    permissions: ['INTERNET', 'RECORD_AUDIO', 'CAMERA', 'VIBRATE'],
    signature_scheme: 'V3_CONTINUITY',
    verified_at: '2026-10-28T14:15:00Z',
    is_sponsored: false,
    iconSvg: SVG_ICONS.messenger
  },
  {
    id: 'app-crypto-tracker-04',
    package_name: 'com.blockmetric.crypto',
    title: 'Pro Crypto Tracker',
    developer_name: 'BlockMetric Systems',
    developer_verified: true,
    category: 'finance',
    category_name: '금융 & 핀테크',
    rating: 4.7,
    rating_count: 14200,
    downloads: '45만+',
    size_bytes: 19084083,
    size_formatted: '18.2 MB',
    price_krw: 4900,
    is_free: false,
    min_sdk: 26,
    version_name: '1.9.4',
    version_code: 10904,
    sha256: '4e07408562bedb8b60ce05c1decfe3ad16b72230967de01f640b7e4729b49fce',
    summary: '글로벌 50개 거래소 실시간 시세, 포트폴리오 세무 보고서 자동 생성기',
    description: '실시간 WebSockets 시세 스트리밍 및 국세청 가상자산 소득세 신고용 원장 자동 생성.',
    permissions: ['INTERNET', 'ACCESS_NETWORK_STATE'],
    signature_scheme: 'V3_CONTINUITY',
    verified_at: '2026-10-27T10:00:00Z',
    is_sponsored: false,
    iconSvg: SVG_ICONS.crypto
  },
  {
    id: 'app-mobile-ide-06',
    package_name: 'com.devtools.mobileide',
    title: 'Code Editor Mobile IDE',
    developer_name: 'Dev Tools Corp',
    developer_verified: true,
    category: 'productivity',
    category_name: '도구 & 유틸리티',
    rating: 4.6,
    rating_count: 9800,
    downloads: '31만+',
    size_bytes: 44040192,
    size_formatted: '42.0 MB',
    price_krw: 9900,
    is_free: false,
    min_sdk: 28,
    version_name: '2.5.0',
    version_code: 20500,
    sha256: 'ef2d127de37b942baad06145e54b0c619a1f22327b2ebbcfbec78f5564afe39d',
    summary: '모바일 환경에서 구동되는 풀스택 터미널 및 Git 연동 고성능 코드 에디터',
    description: '구문 강조, LSP 코드 자동완성, 통합 bash 터미널 및 Git 버전 관리 탑재.',
    permissions: ['INTERNET', 'READ_EXTERNAL_STORAGE', 'WRITE_EXTERNAL_STORAGE'],
    signature_scheme: 'V3_CONTINUITY',
    verified_at: '2026-10-25T11:20:00Z',
    is_sponsored: false,
    iconSvg: SVG_ICONS.editor
  },
  {
    id: 'app-cyber-runner-07',
    package_name: 'com.neonpixel.cyberrunner',
    title: 'Cyber Runner 2026',
    developer_name: 'Neon Pixel Studio',
    developer_verified: true,
    category: 'games',
    category_name: '게임',
    rating: 4.7,
    rating_count: 63100,
    downloads: '95만+',
    size_bytes: 120586240,
    size_formatted: '115.0 MB',
    price_krw: 0,
    is_free: true,
    min_sdk: 26,
    version_name: '1.2.0',
    version_code: 10200,
    sha256: 'e7f6c011776e8db7cd330b54174fd76f7d0216b612387a5ffcfb81e6f0919683',
    summary: '네온 사이버펑크 도시를 질주하는 하이퍼 액션 파쿠르 러닝 게임',
    description: 'Vulkan 고성능 그래픽 엔진 탑재 및 120FPS 모바일 최적화 사이버펑크 러닝 게임.',
    permissions: ['INTERNET', 'VIBRATE'],
    signature_scheme: 'V3_CONTINUITY',
    verified_at: '2026-10-24T09:10:00Z',
    is_sponsored: false,
    iconSvg: SVG_ICONS.runner
  }
])

const categories = [
  { id: 'all', name: '전체', icon: '🌐' },
  { id: 'productivity', name: '생산성 & 업무', icon: '💼' },
  { id: 'finance', name: '금융 & 핀테크', icon: '💳' },
  { id: 'communication', name: '커뮤니케이션', icon: '💬' },
  { id: 'tools', name: '도구 & 보안', icon: '🛠️' },
  { id: 'games', name: '게임', icon: '🎮' }
]

const selectedCategory = ref('all')
const chartTab = ref<'free' | 'grossing' | 'trending'>('free')
const currentBannerIdx = ref(0)
const selectedApp = ref<AppItem | null>(null)
const toastMessage = ref('')
const isDownloading = ref(false)
const downloadProgress = ref(0)
const isInstalled = ref(false)
const isWishlisted = ref(false)

// 3 Cinematic Hero Billboards
const billboardBanners = [
  {
    id: 'banner-omni',
    tag: '✨ 에디터 추천 필수 앱',
    title: 'Omni Productivity Suite',
    subtitle: '차세대 클라우드 협업, 문서 편집, 엔드투엔드 암호화가 집약된 올인원 비즈니스 오피스.',
    gradient: 'linear-gradient(135deg, #0b1a3a 0%, #0d2b59 50%, #0a4b78 100%)',
    app: allApps.value[0],
    graphicSvg: `<svg viewBox="0 0 200 320" width="100%" height="100%"><rect width="200" height="320" rx="16" fill="#0A192F"/><rect x="15" y="30" width="170" height="40" rx="8" fill="#1E3A8A"/><rect x="15" y="80" width="170" height="70" rx="8" fill="#172554"/><path d="M30 115L60 95L100 125L170 85" stroke="#38BDF8" stroke-width="4" stroke-linecap="round"/><rect x="15" y="160" width="80" height="90" rx="8" fill="#1E293B"/><rect x="105" y="160" width="80" height="90" rx="8" fill="#1E293B"/><circle cx="100" cy="285" r="14" fill="#2563EB"/></svg>`
  },
  {
    id: 'banner-finvault',
    tag: '🛡️ 최고 등급 금융 보안',
    title: 'FinVault Secure Banking',
    subtitle: '생체인식 하드웨어 보안 모듈(HSM) 탑재. 개인 자산 및 법인 금융 완벽 보호.',
    gradient: 'linear-gradient(135deg, #062b1e 0%, #0c4a34 50%, #0a5c40 100%)',
    app: allApps.value[1],
    graphicSvg: `<svg viewBox="0 0 200 320" width="100%" height="100%"><rect width="200" height="320" rx="16" fill="#061F17"/><circle cx="100" cy="90" r="45" fill="#059669" fill-opacity="0.3"/><circle cx="100" cy="90" r="28" fill="#10B981"/><path d="M90 90L98 98L112 82" stroke="white" stroke-width="4" stroke-linecap="round"/><rect x="20" y="160" width="160" height="50" rx="10" fill="#064E3B"/><rect x="20" y="220" width="160" height="70" rx="10" fill="#064E3B"/></svg>`
  },
  {
    id: 'banner-cyberrunner',
    tag: '🎮 120FPS 모바일 액션',
    title: 'Cyber Runner 2026',
    subtitle: '네온 사이버펑크 도시를 질주하는 하이퍼 파쿠르 러닝. Vulkan 모바일 그래픽 엔진 적용.',
    gradient: 'linear-gradient(135deg, #240a3e 0%, #43126b 50%, #68178d 100%)',
    app: allApps.value[6],
    graphicSvg: `<svg viewBox="0 0 200 320" width="100%" height="100%"><rect width="200" height="320" rx="16" fill="#1E0B36"/><path d="M10 200L100 80L190 200Z" fill="url(#neonG)" fill-opacity="0.5"/><circle cx="100" cy="140" r="24" fill="#F43F5E"/><path d="M0 260L200 240L100 320Z" fill="#7C3AED"/><defs><linearGradient id="neonG"><stop stop-color="#EC4899"/><stop offset="1" stop-color="#8B5CF6"/></linearGradient></defs></svg>`
  }
]

// Filtered apps based on selected category pill
const filteredRecommendedApps = computed(() => {
  if (selectedCategory.value === 'all') return allApps.value
  return allApps.value.filter(a => a.category === selectedCategory.value)
})

// Filtered apps for Top Charts
const chartFilteredApps = computed(() => {
  if (chartTab.value === 'free') {
    return allApps.value.filter(a => a.is_free)
  } else if (chartTab.value === 'grossing') {
    return [...allApps.value].sort((a, b) => b.price_krw - a.price_krw)
  } else {
    return [...allApps.value].sort((a, b) => b.rating - a.rating)
  }
})

function selectCategory(catId: string) {
  selectedCategory.value = catId
}

function openAppDetail(app: AppItem) {
  selectedApp.value = app
  isDownloading.value = false
  downloadProgress.value = 0
  isInstalled.value = false
}

function closeAppDetail() {
  selectedApp.value = null
}

function triggerQuickInstall(app: AppItem) {
  openAppDetail(app)
  handleInstallClick(app)
}

async function handleInstallClick(app: AppItem) {
  if (isInstalled.value) {
    showToast(`${app.title} 실행 중...`)
    return
  }

  isDownloading.value = true
  downloadProgress.value = 10

  // 1. Edge API token request simulation
  try {
    const res = await fetch('/api/v1/distribution/tokens', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        app_id: app.id,
        package_name: app.package_name
      })
    })
    if (res.ok) {
      downloadProgress.value = 45
    }
  } catch {
    downloadProgress.value = 40
  }

  // 2. Simulated progressive download progress
  const interval = setInterval(() => {
    downloadProgress.value += 20
    if (downloadProgress.value >= 100) {
      clearInterval(interval)
      isDownloading.value = false
      isInstalled.value = true
      showToast(`${app.title} 설치가 완료되었습니다. (TreeScan 무결성 통과)`)
    }
  }, 250)
}

function copyShareLink() {
  if (typeof window !== 'undefined' && selectedApp.value) {
    const url = `${window.location.origin}/apps/${selectedApp.value.id}`
    navigator.clipboard?.writeText(url)
    showToast('앱 공유 링크가 클립보드에 복사되었습니다.')
  }
}

function toggleWishlist() {
  isWishlisted.value = !isWishlisted.value
  showToast(isWishlisted.value ? '위시리스트에 추가되었습니다.' : '위시리스트에서 삭제되었습니다.')
}

function showToast(msg: string) {
  toastMessage.value = msg
  setTimeout(() => {
    toastMessage.value = ''
  }, 3500)
}

// Generate Realistic Phone Mockup Screenshots
function getAppScreenshots(app: AppItem): string[] {
  return [
    `<svg viewBox="0 0 180 320" width="100%" height="100%"><rect width="180" height="320" rx="16" fill="#131B2E"/><rect x="14" y="24" width="152" height="40" rx="8" fill="#1E293B"/><rect x="14" y="74" width="152" height="120" rx="8" fill="#0F172A"/><circle cx="50" cy="134" r="24" fill="#0084FF" fill-opacity="0.3"/><path d="M40 134L48 142L62 126" stroke="#0084FF" stroke-width="3" stroke-linecap="round"/><rect x="85" y="120" width="65" height="10" rx="4" fill="#38BDF8"/><rect x="85" y="138" width="45" height="8" rx="4" fill="#64748B"/><rect x="14" y="204" width="152" height="38" rx="8" fill="#1E293B"/><rect x="14" y="250" width="152" height="50" rx="8" fill="#1E293B"/></svg>`,
    `<svg viewBox="0 0 180 320" width="100%" height="100%"><rect width="180" height="320" rx="16" fill="#0A101D"/><rect x="14" y="24" width="152" height="28" rx="6" fill="#1E293B"/><path d="M20 160C40 120 70 200 100 130C130 80 150 140 165 110" stroke="#00E575" stroke-width="4" stroke-linecap="round"/><rect x="14" y="190" width="70" height="100" rx="8" fill="#1E293B"/><rect x="96" y="190" width="70" height="100" rx="8" fill="#1E293B"/></svg>`,
    `<svg viewBox="0 0 180 320" width="100%" height="100%"><rect width="180" height="320" rx="16" fill="#151226"/><rect x="14" y="24" width="152" height="60" rx="10" fill="#2D1B4E"/><rect x="14" y="94" width="152" height="60" rx="10" fill="#2D1B4E"/><rect x="14" y="164" width="152" height="60" rx="10" fill="#2D1B4E"/><rect x="14" y="234" width="152" height="60" rx="10" fill="#2D1B4E"/></svg>`,
    `<svg viewBox="0 0 180 320" width="100%" height="100%"><rect width="180" height="320" rx="16" fill="#1A1F2C"/><circle cx="90" cy="110" r="40" fill="#059669" fill-opacity="0.2"/><path d="M90 85L108 95V115C108 128 100 138 90 142C80 138 72 128 72 115V95L90 85Z" fill="#10B981"/><rect x="20" y="180" width="140" height="40" rx="8" fill="#2A324B"/><rect x="20" y="230" width="140" height="40" rx="8" fill="#2A324B"/></svg>`
  ]
}
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

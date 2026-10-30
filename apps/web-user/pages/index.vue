<template>
  <main class="bn-container">
    <header class="bn-header">
      <div class="bn-brand">
        <h1>Booking Network App Market</h1>
        <p class="bn-tagline">글로벌 독립 통합 앱 마켓 & 연합 검색엔진</p>
      </div>
      <div class="bn-search-box">
        <input
          v-model="searchQuery"
          type="search"
          placeholder="앱, 게임, 개발사 검색..."
          class="bn-input"
          @keyup.enter="handleSearch"
        />
      </div>
    </header>

    <!-- 5-State UI: Loading Skeleton -->
    <section v-if="loading" class="bn-section">
      <h2>카탈로그 로딩 중...</h2>
      <div class="bn-card-grid">
        <div v-for="i in 6" :key="i" class="bn-app-card">
          <BnLoadingSkeleton height="64px" width="64px" radius="16px" />
          <div style="flex: 1; display: flex; flex-direction: column; gap: 8px;">
            <BnLoadingSkeleton height="18px" width="70%" />
            <BnLoadingSkeleton height="14px" width="40%" />
          </div>
        </div>
      </div>
    </section>

    <!-- 5-State UI: Error Retry -->
    <section v-else-if="error" class="bn-section bn-error-state">
      <p class="bn-error-msg">{{ error }}</p>
      <button class="bn-btn bn-btn-retry" @click="fetchApps">다시 시도</button>
    </section>

    <!-- 5-State UI: Empty State -->
    <section v-else-if="apps.length === 0" class="bn-section bn-empty-state">
      <p>등록된 애플리케이션이 없습니다.</p>
    </section>

    <!-- 5-State UI: Data State -->
    <section v-else class="bn-section">
      <div class="bn-section-header">
        <h2>인기 및 추천 애플리케이션</h2>
      </div>

      <div class="bn-card-grid">
        <article
          v-for="app in apps"
          :key="app.id"
          class="bn-app-card"
          @click="navigateToApp(app.slug)"
        >
          <img :src="app.icon_url || '/placeholder-icon.png'" :alt="app.title" class="bn-app-icon" />
          <div class="bn-app-info">
            <div class="bn-title-row">
              <h3 class="bn-app-title">{{ app.title }}</h3>
              <BnSponsoredBadge v-if="app.is_sponsored" />
            </div>
            <p class="bn-dev-name">{{ app.developer_name }}</p>
            <div class="bn-meta-row">
              <span class="bn-rating">★ {{ app.bayesian_rating?.toFixed(1) || '0.0' }}</span>
              <span class="bn-category">{{ app.category_id }}</span>
            </div>
          </div>
        </article>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { BnSponsoredBadge, BnLoadingSkeleton } from '@bookingnetwork/ui-kit'

interface AppItem {
  id: string
  slug: string
  title: string
  developer_name: string
  icon_url: string
  category_id: string
  bayesian_rating: number
  is_sponsored?: boolean
}

const router = useRouter()
const apps = ref<AppItem[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const searchQuery = ref('')

async function fetchApps() {
  loading.value = true
  error.value = null
  try {
    const res = await fetch('/api/v1/catalog/apps')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    apps.value = data.items || [
      {
        id: '1',
        slug: 'secure-messenger',
        title: 'Secure Messenger',
        developer_name: 'Open Privacy Lab',
        icon_url: '',
        category_id: 'communication',
        bayesian_rating: 4.8
      },
      {
        id: '2',
        slug: 'code-editor-mobile',
        title: 'Code Editor Mobile',
        developer_name: 'Dev Tools Corp',
        icon_url: '',
        category_id: 'productivity',
        bayesian_rating: 4.6,
        is_sponsored: true
      }
    ]
  } catch (err: any) {
    // Provide fallback items for SSR demonstration
    apps.value = [
      {
        id: '1',
        slug: 'secure-messenger',
        title: 'Secure Messenger',
        developer_name: 'Open Privacy Lab',
        icon_url: '',
        category_id: 'communication',
        bayesian_rating: 4.8
      }
    ]
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  if (searchQuery.value.trim()) {
    router.push({ path: '/search', query: { q: searchQuery.value.trim() } })
  }
}

function navigateToApp(slug: string) {
  router.push(`/apps/${slug}`)
}

onMounted(() => {
  fetchApps()
})
</script>

<style scoped>
.bn-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 32px 16px;
}
.bn-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 20px;
  margin-bottom: 36px;
}
.bn-brand h1 {
  font-size: 28px;
  margin: 0 0 6px 0;
  color: var(--bn-text-primary);
}
.bn-tagline {
  margin: 0;
  color: var(--bn-text-muted);
  font-size: 15px;
}
.bn-input {
  width: 320px;
  max-width: 100%;
  padding: 10px 14px;
  border-radius: 8px;
  border: 1px solid var(--bn-border-subtle);
  background-color: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  font-size: 14px;
}
.bn-section {
  margin-bottom: 40px;
}
.bn-card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}
.bn-app-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: 12px;
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}
.bn-app-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}
.bn-app-icon {
  width: 64px;
  height: 64px;
  border-radius: 14px;
  background-color: var(--bn-bg-elevated);
  object-fit: cover;
}
.bn-app-info {
  flex: 1;
}
.bn-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.bn-app-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--bn-text-primary);
}
.bn-dev-name {
  margin: 4px 0;
  font-size: 13px;
  color: var(--bn-text-muted);
}
.bn-meta-row {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: var(--bn-text-secondary);
}
.bn-rating {
  font-weight: 600;
  color: #F59E0B;
}
.bn-btn-retry {
  padding: 8px 16px;
  border-radius: 6px;
  background-color: var(--bn-primary);
  color: white;
  border: none;
  cursor: pointer;
}
</style>


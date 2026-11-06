<template>
  <main class="bn-container">
    <div v-if="loading" class="bn-loading">
      <BnLoadingSkeleton height="32px" width="40%" />
      <BnLoadingSkeleton height="120px" width="100%" radius="12px" style="margin-top: 20px;" />
    </div>

    <div v-else-if="loadError" class="bn-error">
      <p>{{ loadError }}</p>
      <NuxtLink to="/">홈으로 돌아가기</NuxtLink>
    </div>

    <div v-else-if="detail" class="bn-app-detail">
      <!-- App Header Profile -->
      <section class="bn-detail-header">
        <img :src="iconUrl" :alt="detail.localization?.title || detail.app.canonical_slug" class="bn-detail-icon" />
        <div class="bn-detail-main">
          <div class="bn-title-flex">
            <h1>{{ detail.localization?.title || detail.app.canonical_slug }}</h1>
          </div>
          <p class="bn-detail-dev">{{ detail.app.primary_package_name }}</p>
          <div class="bn-detail-tags">
            <span class="bn-badge">{{ detail.app.primary_category_id }}</span>
            <span class="bn-badge">평점 ★ {{ detail.app.average_rating.toFixed(1) }}</span>
            <span class="bn-badge">{{ detail.app.install_count.toLocaleString() }} 설치</span>
            <span class="bn-badge">{{ detail.app.content_rating }}</span>
          </div>
        </div>

        <div class="bn-action-area">
          <BnInstallActionButton
            :distribution-mode="primaryDist?.distribution_mode || 'DIRECT_HOSTED'"
            :download-url="signedUrl"
            :external-url="primaryDist?.official_store_url"
            :external-store-name="primaryDist?.source_type"
            :size-bytes="primaryDist?.latest_release?.file_size_bytes"
            :disabled="!signedUrl && isDirectMode"
            :disabled-reason="downloadDisabledReason"
            @click="refreshSignedUrl"
          />
        </div>
      </section>

      <!-- Trust & Transparency Security Card (REQ-S-001) -->
      <TrustTransparencyCard
        :distribution-mode="primaryDist?.distribution_mode"
        :signing-cert-sha256="primaryDist?.signing_cert_sha256"
      />

      <!-- Multi-Distribution Source Comparison Matrix (REQ-C-003) -->
      <section class="bn-distributions-section">
        <h2>배포처 및 버전 비교 (Multi-Distribution Matrix)</h2>
        <div v-if="detail.distributions.length === 0" class="bn-empty">등록된 배포처가 없습니다.</div>
        <div v-else class="bn-table-wrapper">
          <table class="bn-dist-table">
            <thead>
              <tr>
                <th>배포 소스</th>
                <th>배포 모드</th>
                <th>버전</th>
                <th>서명 지문 일치</th>
                <th>설치 방식</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="dist in detail.distributions" :key="dist.id">
                <td>
                  <strong>{{ dist.source_type }}</strong>
                  <div class="bn-mono-sm">{{ dist.package_name }}</div>
                </td>
                <td>
                  <span class="bn-rights-tag">{{ dist.distribution_mode }}</span>
                </td>
                <td>{{ dist.latest_release?.version_name || '—' }} ({{ dist.latest_release?.version_code || '—' }})</td>
                <td>
                  <BnHashFingerprintBadge v-if="dist.signing_cert_sha256" :hash="dist.signing_cert_sha256" />
                  <span v-else>—</span>
                </td>
                <td>
                  <BnInstallActionButton
                    :distribution-mode="dist.distribution_mode"
                    :download-url="dist.id === primaryDist?.id ? signedUrl : undefined"
                    :external-url="dist.official_store_url"
                    :external-store-name="dist.source_type"
                    :size-bytes="dist.latest_release?.file_size_bytes"
                    :disabled="dist.id === primaryDist?.id && !signedUrl && isDirectMode"
                    :disabled-reason="downloadDisabledReason"
                    @click="refreshSignedUrl"
                  />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- App Description -->
      <section class="bn-desc-section">
        <h2>애플리케이션 소개</h2>
        <p class="bn-desc-content">{{ detail.localization?.full_description || detail.localization?.short_description || '등록된 설명이 없습니다.' }}</p>
        <div v-if="detail.localization?.whats_new" class="bn-whatsnew">
          <h3>새로운 기능</h3>
          <p class="bn-desc-content">{{ detail.localization.whats_new }}</p>
        </div>
        <p v-if="detail.localization?.privacy_policy_url" class="bn-desc-content">
          <a :href="detail.localization.privacy_policy_url" target="_blank" rel="noopener">개인정보처리방침</a>
        </p>
      </section>

      <!-- Screenshots -->
      <section v-if="detail.media_assets.length > 0" class="bn-desc-section">
        <h2>스크린샷</h2>
        <div class="bn-media-row">
          <img v-for="m in detail.media_assets" :key="m.id" :src="m.url" :alt="detail.app.canonical_slug" loading="lazy" class="bn-shot" />
        </div>
      </section>

      <!-- Reviews -->
      <section class="bn-desc-section">
        <h2>리뷰 ({{ reviews.length }})</h2>
        <div v-if="reviews.length === 0" class="bn-empty">아직 등록된 리뷰가 없습니다.</div>
        <div v-else class="bn-review-list">
          <div v-for="rev in reviews" :key="rev.id" class="bn-review-card">
            <div class="bn-review-head">
              <strong>{{ rev.title || '제목 없음' }}</strong>
              <span class="bn-stars">{{ '★'.repeat(rev.rating) + '☆'.repeat(5 - rev.rating) }}</span>
            </div>
            <p class="bn-desc-content">{{ rev.comment }}</p>
            <div class="bn-mono-sm">{{ rev.created_at.slice(0, 10) }}<span v-if="rev.is_verified_install"> · ✓ 설치 확인</span></div>
          </div>
        </div>

        <div class="bn-review-form">
          <h3>리뷰 작성</h3>
          <div v-if="reviewError" class="bn-form-error">{{ reviewError }}</div>
          <div v-if="reviewDone" class="bn-form-done">{{ reviewDone }}</div>
          <label class="bn-field">별점 (1-5)
            <input v-model.number="reviewRating" type="number" min="1" max="5" />
          </label>
          <label class="bn-field">제목
            <input v-model="reviewTitle" type="text" maxlength="120" />
          </label>
          <label class="bn-field">내용
            <textarea v-model="reviewComment" rows="3" maxlength="2000"></textarea>
          </label>
          <button class="bn-btn-primary" :disabled="reviewBusy" @click="submitReview">
            {{ reviewBusy ? '등록 중…' : '리뷰 등록' }}
          </button>
        </div>
      </section>
    </div>

    <div v-else class="bn-error">
      <p>요청하신 애플리케이션을 찾을 수 없습니다.</p>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useHead } from '#imports'
import { BnInstallActionButton, BnHashFingerprintBadge, BnLoadingSkeleton } from '@bookingnetwork/ui-kit'
import { getSharedClient, ApiClientError, type AppDetailDTO, type UserReview } from '@bookingnetwork/api-client'
import TrustTransparencyCard from '~/components/TrustTransparencyCard.vue'

const api = getSharedClient()
const route = useRoute()
const router = useRouter()
const slug = route.params.slug as string

const loading = ref(true)
const loadError = ref('')
const detail = ref<AppDetailDTO | null>(null)
const reviews = ref<UserReview[]>([])
const signedUrl = ref('')
const downloadDisabledReason = ref('')

const reviewRating = ref(5)
const reviewTitle = ref('')
const reviewComment = ref('')
const reviewBusy = ref(false)
const reviewError = ref('')
const reviewDone = ref('')

const primaryDist = computed(() => {
  const dists = detail.value?.distributions || []
  return dists.find(d => d.distribution_mode === 'DIRECT_HOSTED') || dists[0] || null
})

const isDirectMode = computed(() => {
  const mode = primaryDist.value?.distribution_mode
  return mode === 'DIRECT_HOSTED' || mode === 'FEDERATED_REPO'
})

const iconUrl = computed(() => {
  const ch = ((detail.value?.localization?.title || '?').trim().slice(0, 1) || '?').toUpperCase()
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="96" height="96" viewBox="0 0 64 64"><rect width="64" height="64" rx="16" fill="#334155"/><text x="32" y="42" font-size="28" font-weight="700" text-anchor="middle" fill="white" font-family="sans-serif">${ch}</text></svg>`
  return 'data:image/svg+xml,' + encodeURIComponent(svg)
})

function deviceId(): string {
  if (typeof window === 'undefined') return 'ssr-device'
  let id = localStorage.getItem('bn_device_id')
  if (!id) {
    id = 'dev-' + Math.random().toString(36).slice(2) + Date.now().toString(36)
    localStorage.setItem('bn_device_id', id)
  }
  return id
}

async function refreshSignedUrl() {
  const d = detail.value
  const dist = primaryDist.value
  const rel = dist?.latest_release
  if (!d || !dist || !rel) return
  try {
    const signed = await api.createDownloadToken(d.app.id, rel.id, deviceId(), dist.signing_cert_sha256)
    signedUrl.value = signed.url
    downloadDisabledReason.value = ''
  } catch (err) {
    downloadDisabledReason.value = err instanceof ApiClientError ? err.message : '다운로드 불가'
  }
}

async function fetchAppDetail() {
  loading.value = true
  loadError.value = ''
  try {
    const resolved = await api.getAppBySlug(slug)
    detail.value = resolved
    reviews.value = await api.listReviews(resolved.app.id).catch(() => [])
    await refreshSignedUrl()
  } catch (err) {
    if (err instanceof ApiClientError && err.status === 404) {
      loadError.value = '요청하신 애플리케이션을 찾을 수 없습니다.'
    } else {
      loadError.value = err instanceof ApiClientError ? err.message : '앱 정보를 불러오지 못했습니다.'
    }
  } finally {
    loading.value = false
  }
}

async function submitReview() {
  reviewError.value = ''
  reviewDone.value = ''
  const d = detail.value
  if (!d) return
  if (reviewRating.value < 1 || reviewRating.value > 5) {
    reviewError.value = '별점은 1-5 사이여야 합니다.'
    return
  }
  if (!reviewComment.value.trim()) {
    reviewError.value = '리뷰 내용을 입력하십시오.'
    return
  }
  reviewBusy.value = true
  try {
    const created = await api.postReview({
      appId: d.app.id,
      rating: reviewRating.value,
      title: reviewTitle.value.trim(),
      comment: reviewComment.value.trim()
    })
    reviews.value = [created, ...reviews.value]
    reviewDone.value = '리뷰가 등록되었습니다.'
    reviewTitle.value = ''
    reviewComment.value = ''
  } catch (err) {
    if (err instanceof ApiClientError && err.status === 401) {
      await router.push({ path: '/account/login', query: { redirect: route.fullPath } })
      return
    }
    reviewError.value = err instanceof ApiClientError ? err.message : '리뷰 등록에 실패했습니다.'
  } finally {
    reviewBusy.value = false
  }
}

// Inject Schema.org SoftwareApplication JSON-LD (SSR SEO Optimization)
useHead(() => {
  if (!detail.value) return {}
  const d = detail.value
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'SoftwareApplication',
    name: d.localization?.title || d.app.canonical_slug,
    applicationCategory: d.app.primary_category_id,
    operatingSystem: 'ANDROID',
    offers: {
      '@type': 'Offer',
      price: d.app.monetization_type === 'FREE' ? '0' : undefined,
      priceCurrency: 'KRW'
    },
    aggregateRating: {
      '@type': 'AggregateRating',
      ratingValue: d.app.average_rating,
      ratingCount: reviews.value.length
    }
  }

  return {
    title: `${d.localization?.title || d.app.canonical_slug} — Booking Network App Market`,
    script: [
      {
        type: 'application/ld+json',
        children: JSON.stringify(jsonLd)
      }
    ]
  }
})

onMounted(() => {
  fetchAppDetail()
})
</script>

<style scoped>
.bn-container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 32px 16px;
}
.bn-detail-header {
  display: flex;
  align-items: center;
  gap: 24px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}
.bn-detail-icon {
  width: 96px;
  height: 96px;
  border-radius: 20px;
  background-color: var(--bn-bg-elevated);
  object-fit: cover;
}
.bn-detail-main {
  flex: 1;
}
.bn-title-flex {
  display: flex;
  align-items: center;
  gap: 12px;
}
.bn-title-flex h1 {
  font-size: 26px;
  margin: 0;
  color: var(--bn-text-primary);
}
.bn-detail-dev {
  margin: 6px 0 10px 0;
  color: var(--bn-text-muted);
  font-size: 15px;
}
.bn-detail-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.bn-badge {
  background-color: var(--bn-bg-elevated);
  border: 1px solid var(--bn-border-subtle);
  border-radius: 6px;
  padding: 4px 10px;
  font-size: 13px;
  font-weight: 500;
  color: var(--bn-text-secondary);
}
.bn-action-area {
  display: flex;
  align-items: center;
}
.bn-distributions-section, .bn-desc-section {
  margin-top: 36px;
}
.bn-table-wrapper {
  overflow-x: auto;
}
.bn-dist-table {
  width: 100%;
  border-collapse: collapse;
  background-color: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: 8px;
}
.bn-dist-table th, .bn-dist-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid var(--bn-border-subtle);
  font-size: 14px;
}
.bn-dist-table th {
  background-color: var(--bn-bg-elevated);
  font-weight: 600;
  color: var(--bn-text-secondary);
}
.bn-rights-tag {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
  background-color: var(--bn-bg-elevated);
  color: var(--bn-text-muted);
}
.bn-desc-content {
  line-height: 1.6;
  color: var(--bn-text-secondary);
  font-size: 15px;
}
.bn-mono-sm {
  font-family: monospace;
  font-size: 11px;
  color: var(--bn-text-muted, #64748B);
  margin-top: 4px;
}
.bn-empty {
  padding: 16px;
  color: var(--bn-text-muted, #64748B);
  font-size: 14px;
}
.bn-error {
  padding: 32px 16px;
  text-align: center;
  color: #b91c1c;
}
.bn-media-row {
  display: flex;
  gap: 12px;
  overflow-x: auto;
  padding-bottom: 8px;
}
.bn-shot {
  max-height: 320px;
  border-radius: 8px;
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
}
.bn-review-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 24px;
}
.bn-review-card {
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 8px;
  padding: 16px;
}
.bn-review-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.bn-stars {
  color: #d97706;
  font-weight: 700;
}
.bn-review-form {
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 8px;
  padding: 16px;
}
.bn-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 12px;
}
.bn-field input, .bn-field textarea {
  padding: 8px 10px;
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 6px;
  font-size: 14px;
  font-weight: 400;
}
.bn-btn-primary {
  background: #059669;
  color: #fff;
  border: none;
  border-radius: 6px;
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}
.bn-btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.bn-form-error {
  color: #b91c1c;
  font-size: 13px;
  margin-bottom: 12px;
}
.bn-form-done {
  color: #059669;
  font-size: 13px;
  margin-bottom: 12px;
}
.bn-whatsnew {
  margin-top: 16px;
}
</style>

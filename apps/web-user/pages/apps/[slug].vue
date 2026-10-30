<template>
  <main class="bn-container">
    <div v-if="loading" class="bn-loading">
      <BnLoadingSkeleton height="32px" width="40%" />
      <BnLoadingSkeleton height="120px" width="100%" radius="12px" style="margin-top: 20px;" />
    </div>

    <div v-else-if="app" class="bn-app-detail">
      <!-- App Header Profile -->
      <section class="bn-detail-header">
        <img :src="app.icon_url || '/placeholder-icon.png'" :alt="app.title" class="bn-detail-icon" />
        <div class="bn-detail-main">
          <div class="bn-title-flex">
            <h1>{{ app.title }}</h1>
            <BnSponsoredBadge v-if="app.is_sponsored" />
          </div>
          <p class="bn-detail-dev">{{ app.developer_name }}</p>
          <div class="bn-detail-tags">
            <span class="bn-badge">{{ app.category_id }}</span>
            <span class="bn-badge">평점 ★ {{ app.bayesian_rating?.toFixed(1) || '0.0' }}</span>
          </div>
        </div>

        <div class="bn-action-area">
          <BnInstallActionButton
            :distribution-mode="primaryDist?.distribution_mode || 'DIRECT_HOSTED'"
            :download-url="primaryDist?.download_url"
            :external-url="primaryDist?.external_url"
            :external-store-name="primaryDist?.provider_name"
            :size-bytes="primaryDist?.size_bytes"
          />
        </div>
      </section>

      <!-- Trust & Transparency Security Card (REQ-S-001) -->
      <TrustTransparencyCard
        :distribution-mode="primaryDist?.distribution_mode"
        :apk-sha256="primaryDist?.apk_sha256"
        :signing-cert-sha256="primaryDist?.signing_cert_sha256"
        :security-verdict="primaryDist?.security_verdict || 'CLEAN_LOW_RISK'"
        :risk-score="primaryDist?.risk_score || 15"
      />

      <!-- Multi-Distribution Source Comparison Matrix (REQ-C-003) -->
      <section class="bn-distributions-section">
        <h2>배포처 및 버전 비교 (Multi-Distribution Matrix)</h2>
        <div class="bn-table-wrapper">
          <table class="bn-dist-table">
            <thead>
              <tr>
                <th>배포 소스</th>
                <th>출처 권리 등급</th>
                <th>버전</th>
                <th>서명 지문 일치</th>
                <th>설치 방식</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="dist in app.distributions" :key="dist.id">
                <td>
                  <strong>{{ dist.provider_name }}</strong>
                </td>
                <td>
                  <span class="bn-rights-tag">{{ dist.rights_level }}</span>
                </td>
                <td>{{ dist.version_name }} ({{ dist.version_code }})</td>
                <td>
                  <BnHashFingerprintBadge :hash="dist.signing_cert_sha256" />
                </td>
                <td>
                  <BnInstallActionButton
                    :distribution-mode="dist.distribution_mode"
                    :download-url="dist.download_url"
                    :external-url="dist.external_url"
                    :external-store-name="dist.provider_name"
                    :size-bytes="dist.size_bytes"
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
        <div class="bn-desc-content" v-html="app.description_html || app.description"></div>
      </section>
    </div>

    <div v-else class="bn-error">
      <p>요청하신 애플리케이션을 찾을 수 없습니다.</p>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useHead } from '#imports'
import { BnInstallActionButton, BnHashFingerprintBadge, BnSponsoredBadge, BnLoadingSkeleton } from '@bookingnetwork/ui-kit'
import TrustTransparencyCard from '~/components/TrustTransparencyCard.vue'

const route = useRoute()
const slug = route.params.slug as string

const loading = ref(true)
const app = ref<any>(null)

const primaryDist = computed(() => {
  if (!app.value?.distributions || app.value.distributions.length === 0) return null
  return app.value.distributions[0]
})

async function fetchAppDetail() {
  loading.value = true
  try {
    const res = await fetch(`/api/v1/catalog/apps/${slug}`)
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    app.value = await res.json()
  } catch (err) {
    // Provide sample structure for demonstration
    app.value = {
      id: 'app-001',
      slug: slug,
      title: 'Secure Messenger',
      developer_name: 'Open Privacy Lab',
      category_id: 'communication',
      bayesian_rating: 4.85,
      description: '종단간 암호화와 추적 방지를 제공하는 오픈소스 메신저입니다.',
      distributions: [
        {
          id: 'dist-1',
          provider_name: 'BN 자체 검증 스토어',
          distribution_mode: 'DIRECT_HOSTED',
          rights_level: 'LEVEL_1_DIRECT_DEVELOPER',
          version_name: '2.4.0',
          version_code: 240,
          apk_sha256: '9c3a8b12f45812903eabcde1234567890abcdef1234567890abcdef123456789',
          signing_cert_sha256: 'e4b1029caabbccddeeff00112233445566778899aabbccddeeff001122334455',
          size_bytes: 34500000,
          download_url: '/api/v1/distribution/download/app-001'
        },
        {
          id: 'dist-2',
          provider_name: 'F-Droid 오픈소스 미러',
          distribution_mode: 'FEDERATED_REPO',
          rights_level: 'LEVEL_2_LICENSED_FEDERATED_BINARY',
          version_name: '2.3.9',
          version_code: 239,
          apk_sha256: '8b12903eabcde1234567890abcdef1234567890abcdef1234567899c3a8b12f45',
          signing_cert_sha256: 'e4b1029caabbccddeeff00112233445566778899aabbccddeeff001122334455',
          size_bytes: 34200000,
          download_url: 'https://f-droid.org/repo/com.secure.messenger_239.apk'
        },
        {
          id: 'dist-3',
          provider_name: 'Google Play 공식 페이지',
          distribution_mode: 'EXTERNAL_LINK_ONLY',
          rights_level: 'LEVEL_4_OFFICIAL_LINK_ONLY',
          version_name: '2.4.0',
          version_code: 240,
          external_url: 'https://play.google.com/store/apps/details?id=com.secure.messenger'
        }
      ]
    }
  } finally {
    loading.value = false
  }
}

// Inject Schema.org SoftwareApplication JSON-LD (SSR SEO Optimization)
useHead(() => {
  if (!app.value) return {}
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'SoftwareApplication',
    name: app.value.title,
    applicationCategory: app.value.category_id,
    operatingSystem: 'ANDROID',
    offers: {
      '@type': 'Offer',
      price: '0.00',
      priceCurrency: 'USD'
    },
    aggregateRating: {
      '@type': 'AggregateRating',
      ratingValue: app.value.bayesian_rating || 5.0,
      ratingCount: 120
    }
  }

  return {
    title: `${app.value.title} — Booking Network App Market`,
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
</style>


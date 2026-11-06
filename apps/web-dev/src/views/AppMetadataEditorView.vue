<template>
  <div class="bn-dev-page">
    <header class="bn-page-header">
      <h2>앱 메타데이터 및 다국어 편집기</h2>
      <p class="bn-text-muted">스토어에 노출될 앱 제목, 설명, 스크린샷 및 배포 카테고리를 설정합니다.</p>
    </header>

    <div class="bn-select-row">
      <div class="bn-select-group">
        <label for="orgSelect">조직</label>
        <select id="orgSelect" v-model="activeOrgId" class="bn-input" @change="onOrgChange">
          <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.org_name }}</option>
        </select>
      </div>
      <div class="bn-select-group">
        <label for="appSelect">앱</label>
        <select id="appSelect" v-model="activeAppId" class="bn-input" @change="loadAll">
          <option v-for="app in apps" :key="app.id" :value="app.id">{{ app.slug }} ({{ app.packageName }})</option>
        </select>
      </div>
    </div>
    <div v-if="isLoading" class="bn-loading">불러오는 중…</div>
    <div v-if="loadError" class="bn-error">{{ loadError }}</div>

    <template v-if="!isLoading && !loadError && activeAppId">
      <form class="bn-form-card" @submit.prevent="saveMetadata">
        <div class="bn-lang-tabs">
          <button
            v-for="loc in supportedLocales"
            :key="loc"
            type="button"
            class="bn-tab"
            :class="{ active: currentLocale === loc }"
            @click="currentLocale = loc"
          >
            {{ loc }}
          </button>
        </div>

        <div class="bn-form-group">
          <label>앱 이름 ({{ currentLocale }})</label>
          <input v-model="i18nTitles[currentLocale]" type="text" class="bn-input" required />
        </div>

        <div class="bn-form-group">
          <label>간략 소개 (한 줄 요약)</label>
          <input v-model="i18nShortDescs[currentLocale]" type="text" class="bn-input" />
        </div>

        <div class="bn-form-group">
          <label>상세 설명 (HTML 또는 텍스트)</label>
          <textarea v-model="i18nDescs[currentLocale]" rows="6" class="bn-textarea"></textarea>
        </div>

        <div class="bn-form-group">
          <label>새로운 소식 (Whats New)</label>
          <textarea v-model="i18nWhatsNew[currentLocale]" rows="3" class="bn-textarea"></textarea>
        </div>

        <div v-if="saveError" class="bn-error">{{ saveError }}</div>
        <div v-if="saveDone" class="bn-success">메타데이터가 성공적으로 저장되었습니다.</div>
        <div class="bn-btn-row">
          <button type="submit" class="bn-btn bn-btn-primary" :disabled="saving">
            {{ saving ? '저장 중...' : '메타데이터 저장' }}
          </button>
          <button type="button" class="bn-btn bn-btn-secondary" :disabled="submitting" @click="submitAppForReview">
            {{ submitting ? '제출 중...' : '심사에 제출' }}
          </button>
        </div>
        <div v-if="submitError" class="bn-error">{{ submitError }}</div>
        <div v-if="submitDone" class="bn-success">심사에 제출되었습니다. 승인 시 스토어에 게시됩니다.</div>
      </form>

      <section class="bn-form-card bn-media-card">
        <h3>스크린샷 및 배너 ({{ currentLocale }})</h3>
        <p class="bn-help-text">PNG/JPEG, 장당 최대 20MB. 업로드 즉시 스토어 리스팅에 반영됩니다.</p>

        <div v-if="mediaError" class="bn-error">{{ mediaError }}</div>
        <div v-if="media.length === 0" class="bn-empty">등록된 이미지가 없습니다.</div>
        <div v-else class="bn-media-grid">
          <figure v-for="m in media" :key="m.id" class="bn-media-item">
            <img :src="previewUrl(m.id)" :alt="`${m.asset_type} ${m.width}x${m.height}`" loading="lazy" />
            <figcaption>
              <span class="bn-media-tag">{{ m.asset_type }}</span>
              <span class="bn-text-muted">{{ m.width }}x{{ m.height }} · {{ m.device_form_factor }}</span>
              <button type="button" class="bn-btn-text-danger" :disabled="mediaBusy" @click="removeMedia(m.id)">삭제</button>
            </figcaption>
          </figure>
        </div>

        <div class="bn-screenshot-dropzone">
          <input type="file" accept="image/png,image/jpeg" :disabled="mediaBusy" @change="uploadScreenshot" />
          <p class="bn-drop-text">권장 규격: 1080x1920 (세로) 또는 1920x1080 (가로)</p>
        </div>
      </section>
    </template>
    <div v-else-if="!isLoading && !loadError" class="bn-empty">조직에 등록된 앱이 없습니다.</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getSharedClient, getStoredAccessToken, ApiClientError, type OrgSummary } from '@bookingnetwork/api-client'

const api = getSharedClient()

interface OrgApp {
  id: string
  slug: string
  packageName: string
  status: string
}

interface MediaAsset {
  id: string
  app_id: string
  locale: string
  device_form_factor: string
  asset_type: string
  url: string
  width: number
  height: number
}

const supportedLocales = ['ko-KR', 'en-US', 'ja-JP', 'ar-SA']
const currentLocale = ref('ko-KR')
const saving = ref(false)
const saveError = ref('')
const saveDone = ref(false)

const orgs = ref<OrgSummary[]>([])
const apps = ref<OrgApp[]>([])
const activeOrgId = ref('')
const activeAppId = ref('')
const isLoading = ref(true)
const loadError = ref('')

const i18nTitles = ref<Record<string, string>>({})
const i18nShortDescs = ref<Record<string, string>>({})
const i18nDescs = ref<Record<string, string>>({})
const i18nWhatsNew = ref<Record<string, string>>({})

const media = ref<MediaAsset[]>([])
const mediaError = ref('')
const mediaBusy = ref(false)
const previewCache = ref<Record<string, string>>({})
const submitting = ref(false)
const submitError = ref('')
const submitDone = ref(false)

onMounted(async () => {
  try {
    orgs.value = await api.listOrgs()
    if (orgs.value.length > 0) {
      activeOrgId.value = orgs.value[0].id
      await loadApps()
      if (apps.value.length > 0) {
        activeAppId.value = apps.value[0].id
        await loadAll()
      }
    }
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load organizations'
  } finally {
    isLoading.value = false
  }
})

async function onOrgChange() {
  await loadApps()
  activeAppId.value = apps.value.length > 0 ? apps.value[0].id : ''
  if (activeAppId.value) {
    await loadAll()
  }
}

async function loadApps() {
  apps.value = await api.listOrgApps(activeOrgId.value) as OrgApp[]
}

async function loadAll() {
  if (!activeAppId.value) return
  isLoading.value = true
  loadError.value = ''
  try {
    const locs = await api.listLocalizations(activeAppId.value, activeOrgId.value) as Array<Record<string, string>>
    i18nTitles.value = {}
    i18nShortDescs.value = {}
    i18nDescs.value = {}
    i18nWhatsNew.value = {}
    for (const loc of locs) {
      const locale = String(loc.locale || '')
      if (!locale) continue
      i18nTitles.value[locale] = String(loc.title || '')
      i18nShortDescs.value[locale] = String(loc.shortDescription || loc.short_description || '')
      i18nDescs.value[locale] = String(loc.fullDescription || loc.full_description || '')
      i18nWhatsNew.value[locale] = String(loc.whatsNew || loc.whats_new || '')
    }
    await loadMedia()
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load metadata'
  } finally {
    isLoading.value = false
  }
}

async function loadMedia() {
  media.value = await api.listMedia(activeAppId.value, currentLocale.value, activeOrgId.value) as MediaAsset[]
  for (const m of media.value) {
    void ensurePreview(m.id)
  }
}

async function saveMetadata() {
  saving.value = true
  saveError.value = ''
  saveDone.value = false
  try {
    await api.saveLocalization(activeAppId.value, {
      orgId: activeOrgId.value,
      locale: currentLocale.value,
      title: i18nTitles.value[currentLocale.value] || '',
      shortDescription: i18nShortDescs.value[currentLocale.value] || '',
      fullDescription: i18nDescs.value[currentLocale.value] || '',
      whatsNew: i18nWhatsNew.value[currentLocale.value] || ''
    })
    saveDone.value = true
  } catch (err) {
    saveError.value = err instanceof ApiClientError ? err.message : 'Failed to save metadata'
  } finally {
    saving.value = false
  }
}

async function submitAppForReview() {
  if (!activeAppId.value) return
  submitting.value = true
  submitError.value = ''
  submitDone.value = false
  try {
    await api.submitForReview(activeAppId.value, activeOrgId.value)
    submitDone.value = true
  } catch (err) {
    submitError.value = err instanceof ApiClientError ? err.message : 'Failed to submit for review'
  } finally {
    submitting.value = false
  }
}

async function uploadScreenshot(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files && target.files[0]
  if (!file || !activeAppId.value) return
  mediaBusy.value = true
  mediaError.value = ''
  try {
    await api.uploadMedia(activeAppId.value, file, {
      org: activeOrgId.value,
      locale: currentLocale.value,
      formFactor: 'PHONE',
      assetType: 'SCREENSHOT'
    })
    await loadMedia()
  } catch (err) {
    mediaError.value = err instanceof ApiClientError ? err.message : 'Failed to upload image'
  } finally {
    mediaBusy.value = false
    target.value = ''
  }
}

async function removeMedia(assetId: string) {
  if (!confirm('이 이미지를 삭제하시겠습니까?')) return
  mediaBusy.value = true
  mediaError.value = ''
  try {
    await api.deleteMedia(activeAppId.value, assetId, activeOrgId.value)
    const cached = previewCache.value[assetId]
    if (cached) {
      URL.revokeObjectURL(cached)
      delete previewCache.value[assetId]
    }
    await loadMedia()
  } catch (err) {
    mediaError.value = err instanceof ApiClientError ? err.message : 'Failed to delete image'
  } finally {
    mediaBusy.value = false
  }
}

function previewUrl(assetId: string): string {
  return previewCache.value[assetId] || ''
}

async function ensurePreview(assetId: string) {
  if (previewCache.value[assetId]) return
  const token = getStoredAccessToken()
  const res = await fetch(api.mediaContentUrl(activeAppId.value, assetId, activeOrgId.value), {
    headers: token ? { Authorization: `Bearer ${token}` } : {}
  })
  if (!res.ok) return
  const blob = await res.blob()
  previewCache.value[assetId] = URL.createObjectURL(blob)
}
</script>

<style scoped>
.bn-dev-page {
  padding: 24px;
  max-width: 900px;
}
.bn-select-row {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.bn-select-group {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}
.bn-loading {
  padding: 16px;
  color: var(--bn-text-muted, #64748B);
  font-size: 14px;
}
.bn-error {
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #991b1b;
  border-radius: 6px;
  font-size: 13px;
  margin-bottom: 16px;
}
.bn-success {
  padding: 12px 16px;
  background: #ECFDF5;
  border: 1px solid #10B981;
  color: #065F46;
  border-radius: 6px;
  font-size: 13px;
  margin-bottom: 16px;
}
.bn-empty {
  padding: 24px;
  text-align: center;
  color: var(--bn-text-muted, #64748B);
  font-size: 14px;
}
.bn-page-header {
  margin-bottom: 24px;
}
.bn-lang-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
}
.bn-tab {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  background-color: var(--bn-bg-elevated, #F1F5F9);
  cursor: pointer;
  font-weight: 500;
}
.bn-tab.active {
  background-color: var(--bn-primary, #2563EB);
  color: white;
  border-color: var(--bn-primary, #2563EB);
}
.bn-form-card {
  background-color: var(--bn-bg-surface, #FFFFFF);
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 12px;
  padding: 24px;
}
.bn-media-card {
  margin-top: 24px;
}
.bn-media-card h3 {
  margin: 0 0 4px;
}
.bn-help-text {
  font-size: 12px;
  color: var(--bn-text-muted, #64748B);
  margin: 0 0 16px;
}
.bn-media-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}
.bn-media-item {
  margin: 0;
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  border-radius: 8px;
  overflow: hidden;
}
.bn-media-item img {
  width: 100%;
  height: 160px;
  object-fit: cover;
  display: block;
  background: var(--bn-bg-elevated, #F1F5F9);
}
.bn-media-item figcaption {
  padding: 8px 10px;
  font-size: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.bn-media-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  background-color: var(--bn-bg-elevated, #F1F5F9);
  color: var(--bn-primary, #2563EB);
}
.bn-btn-text-danger {
  color: var(--bn-danger-red, #DC2626);
  background: none;
  border: none;
  cursor: pointer;
  font-weight: 600;
  padding: 0;
  text-align: left;
}
.bn-form-group {
  margin-bottom: 20px;
}
.bn-form-group label {
  display: block;
  font-weight: 600;
  margin-bottom: 6px;
  font-size: 14px;
}
.bn-input, .bn-textarea {
  width: 100%;
  padding: 10px 14px;
  border-radius: 6px;
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  font-size: 14px;
}
.bn-select-group .bn-input {
  width: auto;
  min-width: 220px;
}
.bn-screenshot-dropzone {
  border: 2px dashed var(--bn-border-subtle, #E2E8F0);
  border-radius: 8px;
  padding: 24px;
  text-align: center;
}
.bn-drop-text {
  font-size: 12px;
  color: var(--bn-text-muted, #64748B);
  margin-top: 8px;
}
.bn-btn-row {
  display: flex;
  gap: 12px;
}
.bn-btn-primary {
  background-color: var(--bn-primary, #2563EB);
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}
.bn-btn-secondary {
  background-color: var(--bn-bg-elevated, #F1F5F9);
  color: var(--bn-text-primary, #0F172A);
  border: 1px solid var(--bn-border-subtle, #E2E8F0);
  padding: 10px 20px;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
}
.bn-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>

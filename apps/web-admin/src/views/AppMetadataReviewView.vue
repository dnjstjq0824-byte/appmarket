<template>
  <div class="metadata-review-view">
    <div class="view-header">
      <h2>App Metadata & Compliance Review (ADM-04)</h2>
      <p class="subtitle">Mandatory 4-item compliance verification before catalog publication</p>
    </div>

    <div v-if="isLoading" class="loading-row">Loading review queue…</div>
    <div v-if="loadError" class="load-error">{{ loadError }}</div>

    <template v-if="!isLoading && !loadError">
      <div class="queue-bar">
        <label for="queueSelect">Review queue ({{ queue.length }})</label>
        <select id="queueSelect" v-model="activeAppId" @change="loadDetail">
          <option v-for="app in queue" :key="app.id" :value="app.id">
            {{ app.canonical_slug }} ({{ app.primary_package_name }})
          </option>
        </select>
        <button class="btn btn-secondary btn-sm" @click="reloadAll">새로고침</button>
      </div>

      <div v-if="queue.length === 0" class="empty-row">No apps awaiting initial review.</div>

      <div v-else-if="detail" class="review-layout">
        <div class="app-details-panel">
          <div class="panel-header">
            <div class="app-identity">
              <div>
                <h3>{{ detailTitle }}</h3>
                <span class="font-mono text-sm text-muted">{{ detail.app.primary_package_name }}</span>
              </div>
            </div>
            <span class="badge badge-warning">{{ detail.app.catalog_status }}</span>
          </div>

          <div class="metadata-section">
            <label>Category & Content Rating</label>
            <div class="val">{{ detail.app.primary_category_id || '—' }} / {{ detail.app.global_age_rating || 'Unrated' }}</div>
          </div>

          <div class="metadata-section">
            <label>Short Description ({{ detail.resolved_locale }})</label>
            <p class="desc-text">{{ detail.localization?.short_description || '—' }}</p>
          </div>

          <div class="metadata-section">
            <label>Full Description</label>
            <p class="desc-text">{{ detail.localization?.full_description || '—' }}</p>
          </div>

          <div class="metadata-section">
            <label>Whats New</label>
            <p class="desc-text">{{ detail.localization?.whats_new || '—' }}</p>
          </div>

          <div class="metadata-section">
            <label>Privacy Policy URL</label>
            <div class="val">{{ detail.localization?.privacy_policy_url || '—' }}</div>
          </div>

          <div class="metadata-section">
            <label>Media Assets ({{ (detail.media_assets || []).length }})</label>
            <div class="asset-grid">
              <div class="asset-placeholder" v-for="m in (detail.media_assets || [])" :key="m.id">
                {{ m.asset_type }} {{ m.width }}x{{ m.height }}
              </div>
              <div v-if="(detail.media_assets || []).length === 0" class="text-muted text-sm">No screenshots submitted.</div>
            </div>
          </div>

          <div class="metadata-section">
            <label>Distributions ({{ (detail.distributions || []).length }})</label>
            <div v-for="d in (detail.distributions || [])" :key="d.id" class="dist-row">
              <span class="font-mono text-sm">{{ d.package_name }}</span>
              <span class="badge badge-secondary">{{ d.distribution_mode }}</span>
            </div>
          </div>
        </div>

        <div class="checklist-panel">
          <h3>Mandatory Review Checklist</h3>
          <p class="checklist-desc">Per policy compliance (TC-FUNC-013), all 4 items must be verified and checked to enable approval.</p>

          <div class="checklist-items">
            <label class="check-item" :class="{ 'checked': checklist.contentRatingValid }">
              <input type="checkbox" v-model="checklist.contentRatingValid" />
              <div class="check-text">
                <strong>1. Content Rating & Age Appropriateness</strong>
                <span>App age classification matches declared content, absence of unrated adult or gambling elements.</span>
              </div>
            </label>

            <label class="check-item" :class="{ 'checked': checklist.privacyPolicyCompliant }">
              <input type="checkbox" v-model="checklist.privacyPolicyCompliant" />
              <div class="check-text">
                <strong>2. Privacy Policy & Data Transparency</strong>
                <span>Accessible HTTPS privacy policy accurately detailing collected PII and third-party SDKs.</span>
              </div>
            </label>

            <label class="check-item" :class="{ 'checked': checklist.mediaAssetsAccurate }">
              <input type="checkbox" v-model="checklist.mediaAssetsAccurate" />
              <div class="check-text">
                <strong>3. Accurate Media & Description</strong>
                <span>Screenshots and description represent actual app gameplay/features without misleading claims.</span>
              </div>
            </label>

            <label class="check-item" :class="{ 'checked': checklist.noBrandImpersonation }">
              <input type="checkbox" v-model="checklist.noBrandImpersonation" />
              <div class="check-text">
                <strong>4. Brand Impersonation & IP Clearance</strong>
                <span>Title, icon, and assets do not infringe registered trademarks or impersonate protected brands.</span>
              </div>
            </label>
          </div>

          <div class="notes-box">
            <label>Review Notes (Required on rejection)</label>
            <textarea v-model="reviewerNotes" rows="3" placeholder="Provide feedback or rejection details..."></textarea>
          </div>

          <div v-if="actionError" class="load-error">{{ actionError }}</div>
          <div v-if="actionDone" class="action-done">{{ actionDone }}</div>
          <div class="action-buttons">
            <button @click="rejectSubmission" class="btn btn-danger" :disabled="!reviewerNotes.trim() || isProcessing">
              Reject Submission
            </button>
            <button
              @click="approveSubmission"
              class="btn btn-primary"
              :disabled="!isAllChecked || isProcessing"
              id="btnApproveRelease"
            >
              {{ isProcessing ? 'Submitting...' : 'Approve & Release (4/4 Verified)' }}
            </button>
          </div>

          <div v-if="!isAllChecked" class="checklist-warning">
            All 4 checklist items must be verified before approval is permitted.
          </div>
        </div>
      </div>

      <div v-if="decisions.length > 0" class="decisions-card">
        <h3>Recorded Decisions</h3>
        <table class="decisions-table">
          <thead>
            <tr><th>App</th><th>Decision</th><th>Decided By</th><th>At</th></tr>
          </thead>
          <tbody>
            <tr v-for="d in decisions" :key="d.appId">
              <td class="font-mono text-sm">{{ d.appId }}</td>
              <td><span class="badge" :class="d.decision === 'APPROVED' ? 'badge-success' : 'badge-danger'">{{ d.decision }}</span></td>
              <td class="text-sm">{{ d.decidedBy }}</td>
              <td class="font-mono text-sm">{{ d.decidedAt }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { getSharedClient, ApiClientError } from '@bookingnetwork/api-client';

const api = getSharedClient();

interface QueueApp {
  id: string;
  canonical_slug: string;
  primary_package_name: string;
  catalog_status: string;
}

interface MediaAsset {
  id: string;
  app_id: string;
  locale: string;
  device_form_factor: string;
  asset_type: string;
  url: string;
  width: number;
  height: number;
  display_order: number;
}

interface Distribution {
  id: string;
  app_id: string;
  source_type: string;
  distribution_mode: string;
  package_name: string;
}

interface AppDetail {
  app: Record<string, string>;
  localization: Record<string, string> | null;
  media_assets: MediaAsset[];
  distributions: Distribution[];
  tags: unknown[];
  resolved_locale: string;
}

interface Decision {
  appId: string;
  decision: string;
  notes: string;
  decidedBy: string;
  decidedAt: string;
}

const queue = ref<QueueApp[]>([]);
const activeAppId = ref('');
const detail = ref<AppDetail | null>(null);
const decisions = ref<Decision[]>([]);
const isLoading = ref(true);
const loadError = ref('');
const actionError = ref('');
const actionDone = ref('');

const checklist = ref({
  contentRatingValid: false,
  privacyPolicyCompliant: false,
  mediaAssetsAccurate: false,
  noBrandImpersonation: false
});

const reviewerNotes = ref('');
const isProcessing = ref(false);

const isAllChecked = computed(() => {
  return checklist.value.contentRatingValid &&
         checklist.value.privacyPolicyCompliant &&
         checklist.value.mediaAssetsAccurate &&
         checklist.value.noBrandImpersonation;
});

const detailTitle = computed(() => {
  return detail.value?.localization?.title || detail.value?.app?.canonical_slug || '';
});

onMounted(async () => {
  await reloadAll();
});

async function reloadAll() {
  isLoading.value = true;
  loadError.value = '';
  try {
    queue.value = await api.getReviewQueue() as QueueApp[];
    decisions.value = await api.listDecisions() as Decision[];
    if (queue.value.length > 0) {
      activeAppId.value = queue.value[0].id;
      await loadDetail();
    } else {
      detail.value = null;
    }
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load review queue';
  } finally {
    isLoading.value = false;
  }
}

async function loadDetail() {
  if (!activeAppId.value) return;
  actionError.value = '';
  actionDone.value = '';
  try {
    detail.value = await api.adminGetApp(activeAppId.value) as AppDetail;
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to load app detail';
  }
}

async function decide(verdict: 'APPROVED' | 'REJECTED') {
  isProcessing.value = true;
  actionError.value = '';
  actionDone.value = '';
  try {
    await api.decideReview(activeAppId.value, verdict, reviewerNotes.value, { ...checklist.value });
    actionDone.value = verdict === 'APPROVED'
      ? 'Release approved and published to the catalog.'
      : 'Submission rejected and returned to draft.';
    reviewerNotes.value = '';
    checklist.value = { contentRatingValid: false, privacyPolicyCompliant: false, mediaAssetsAccurate: false, noBrandImpersonation: false };
    await reloadAll();
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to record decision';
  } finally {
    isProcessing.value = false;
  }
}

async function approveSubmission() {
  if (!isAllChecked.value) return;
  await decide('APPROVED');
}

async function rejectSubmission() {
  if (!reviewerNotes.value.trim()) return;
  await decide('REJECTED');
}
</script>

<style scoped>
.metadata-review-view {
  max-width: 1100px;
  margin: 0 auto;
}
.view-header {
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--bn-ink);
}
.view-header h2 {
  font-family: var(--bn-font-display);
  font-size: 24px;
  font-weight: 900;
  letter-spacing: -0.02em;
  margin: 0 0 4px;
}
.subtitle {
  color: var(--bn-text-secondary);
  font-size: 13px;
  margin: 0;
}
.loading-row {
  padding: 24px;
  color: var(--bn-text-muted);
  font-size: 14px;
  text-align: center;
}
.load-error {
  padding: 12px 16px;
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  color: var(--bn-danger-red);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  margin-bottom: 16px;
}
.action-done {
  padding: 12px 16px;
  background: var(--bn-trust-wash);
  border: 1px solid var(--bn-trust-verified);
  color: var(--bn-trust-verified);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  margin-bottom: 16px;
}
.empty-row {
  padding: 28px 24px;
  color: var(--bn-text-muted);
  font-size: 14px;
  text-align: center;
  border: 1px dashed var(--bn-border-strong);
  border-radius: var(--bn-radius-md);
}
.queue-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  font-size: 13px;
  color: var(--bn-text-secondary);
  flex-wrap: wrap;
}
.queue-bar label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  flex-shrink: 0;
}
.queue-bar select {
  padding: 8px 12px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-family: var(--bn-font-mono);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  min-width: 320px;
  flex: 1;
}
.review-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}
.app-details-panel, .checklist-panel {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
}
.checklist-panel > h3 {
  font-family: var(--bn-font-display);
  font-size: 15px;
  font-weight: 800;
  margin: 0 0 6px;
}
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  border-bottom: 1px solid var(--bn-border-subtle);
  padding-bottom: 16px;
  gap: 12px;
}
.app-identity h3 {
  font-family: var(--bn-font-display);
  font-size: 17px;
  font-weight: 800;
  margin: 0 0 4px 0;
}
.metadata-section {
  margin-bottom: 16px;
}
.metadata-section label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  display: block;
  margin-bottom: 6px;
}
.metadata-section .val {
  font-size: 13px;
}
.desc-text {
  font-size: 13px;
  line-height: 1.65;
  color: var(--bn-text-secondary);
  margin: 0;
}
.asset-grid {
  display: flex;
  gap: 8px;
  margin-top: 6px;
  flex-wrap: wrap;
}
.asset-placeholder {
  min-width: 90px;
  font-family: var(--bn-font-mono);
  font-size: 11px;
  background: var(--bn-bg-well);
  border: 1px dashed var(--bn-border-strong);
  border-radius: var(--bn-radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--bn-text-muted);
  text-align: center;
  padding: 12px 8px;
}
.dist-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 6px;
}
.checklist-desc {
  font-size: 12px;
  line-height: 1.6;
  color: var(--bn-text-secondary);
  margin: 0 0 16px;
}
.checklist-items {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 20px;
}
.check-item {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  padding: 12px 14px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  cursor: pointer;
}
.check-item.checked {
  border-color: var(--bn-trust-verified);
  background: var(--bn-trust-wash);
}
.check-item input {
  margin-top: 3px;
  cursor: pointer;
  accent-color: var(--bn-trust-verified);
}
.check-text strong {
  display: block;
  font-size: 13px;
  margin-bottom: 2px;
}
.check-text span {
  font-size: 12px;
  color: var(--bn-text-secondary);
  line-height: 1.55;
}
.notes-box {
  margin-bottom: 20px;
}
.notes-box label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  display: block;
  margin-bottom: 8px;
}
.notes-box textarea {
  width: 100%;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  padding: 10px 12px;
  font-family: var(--bn-font-sans);
  font-size: 13px;
  line-height: 1.6;
  box-sizing: border-box;
  resize: vertical;
}
.notes-box textarea:focus {
  outline: none;
  border-color: var(--bn-ink);
}
.action-buttons {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}
.btn {
  padding: 10px 20px;
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}
.btn-primary {
  background: var(--bn-action-bg);
  color: var(--bn-action-text);
  border: none;
}
.btn-primary:hover:not(:disabled) {
  background: var(--bn-action-bg-hover);
}
.btn-primary:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-danger {
  background: var(--bn-danger-red);
  color: #fff;
  border: none;
}
.btn-danger:hover:not(:disabled) {
  filter: brightness(0.92);
}
.btn-danger:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-secondary {
  background: transparent;
  border: 1px solid var(--bn-border-subtle);
  color: var(--bn-text-secondary);
}
.btn-secondary:hover:not(:disabled) {
  border-color: var(--bn-border-strong);
  color: var(--bn-text-primary);
}
.btn-secondary:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-sm { font-size: 12px; padding: 7px 14px; }
.checklist-warning {
  margin-top: 12px;
  font-size: 12px;
  color: var(--bn-warning-amber);
}
.badge {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.04em;
  padding: 3px 9px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-secondary);
  white-space: nowrap;
}
.badge-warning { border-color: var(--bn-warning-amber); color: var(--bn-warning-amber); }
.badge-success { border-color: var(--bn-trust-verified); color: var(--bn-trust-verified); }
.badge-danger { border-color: var(--bn-danger-red); color: var(--bn-danger-red); }
.badge-secondary { border-color: var(--bn-border-strong); color: var(--bn-text-secondary); }
.font-mono { font-family: var(--bn-font-mono); }
.font-xs { font-size: 10px; }
.text-sm { font-size: 12px; }
.text-muted { color: var(--bn-text-muted); }
.decisions-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
  margin-top: 24px;
}
.decisions-card h3 {
  font-family: var(--bn-font-display);
  font-size: 15px;
  font-weight: 800;
  margin: 0 0 16px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--bn-border-subtle);
}
.decisions-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.decisions-table th {
  text-align: left;
  padding: 8px 12px;
  border-bottom: 2px solid var(--bn-ink);
  font-family: var(--bn-font-mono);
  font-weight: 400;
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
}
.decisions-table td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--bn-border-subtle);
  vertical-align: middle;
}
.decisions-table tr:last-child td {
  border-bottom: none;
}
@media (max-width: 900px) {
  .review-layout {
    grid-template-columns: 1fr;
  }
}
</style>

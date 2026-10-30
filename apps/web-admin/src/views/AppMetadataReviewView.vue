<template>
  <div class="metadata-review-view">
    <div class="view-header">
      <h2>App Metadata & Compliance Review (ADM-04)</h2>
      <p class="subtitle">Mandatory 4-item compliance verification before catalog publication</p>
    </div>

    <div class="review-layout">
      <!-- App Information Panel -->
      <div class="app-details-panel">
        <div class="panel-header">
          <div class="app-identity">
            <div class="app-icon-preview">📱</div>
            <div>
              <h3>{{ submission.appName }}</h3>
              <span class="font-mono text-sm text-muted">{{ submission.packageName }} (v{{ submission.versionName }})</span>
            </div>
          </div>
          <span class="badge badge-warning">Awaiting Review</span>
        </div>

        <div class="metadata-section">
          <label>Developer Organization</label>
          <div class="val">{{ submission.developerName }} <span class="badge badge-success font-xs">KYB Verified</span></div>
        </div>

        <div class="metadata-section">
          <label>Category & Target Rating</label>
          <div class="val">{{ submission.category }} / Rated: {{ submission.contentRating }}</div>
        </div>

        <div class="metadata-section">
          <label>Short Description</label>
          <p class="desc-text">{{ submission.shortDescription }}</p>
        </div>

        <div class="metadata-section">
          <label>Privacy Policy URL</label>
          <a :href="submission.privacyPolicyUrl" target="_blank" class="link">{{ submission.privacyPolicyUrl }} ↗</a>
        </div>

        <div class="metadata-section">
          <label>Submitted Assets</label>
          <div class="asset-grid">
            <div class="asset-placeholder" v-for="n in 3" :key="n">Screenshot {{ n }}</div>
          </div>
        </div>
      </div>

      <!-- Mandatory 4-Checklist Review Panel -->
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

        <!-- Reviewer Notes -->
        <div class="notes-box">
          <label>Review Notes (Required on rejection)</label>
          <textarea v-model="reviewerNotes" rows="3" placeholder="Provide feedback or rejection details..."></textarea>
        </div>

        <!-- Action Controls -->
        <div class="action-buttons">
          <button @click="rejectSubmission" class="btn btn-danger" :disabled="!reviewerNotes.trim()">
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
          ⚠️ All 4 checklist items must be verified before approval is permitted.
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const submission = ref({
  appId: 'app-98213812',
  appName: 'Secure Pay Wallet',
  packageName: 'com.securepay.wallet',
  versionName: '1.2.0',
  developerName: 'Fintech Solutions Ltd.',
  category: 'FINANCE',
  contentRating: 'PEGI 3 / Everyone',
  shortDescription: 'Fast, secure contactless NFC payment solution with biometric authorization.',
  privacyPolicyUrl: 'https://securepay.example.com/privacy'
});

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

async function approveSubmission() {
  if (!isAllChecked.value) return;
  isProcessing.value = true;
  try {
    const res = await fetch(`/api/v1/admin/reviews/${submission.value.appId}/decision`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        decision: 'APPROVED',
        notes: reviewerNotes.value,
        checklist: checklist.value
      })
    });
    if (res.ok) {
      alert('Release successfully approved!');
    }
  } catch (err) {
    console.error(err);
  } finally {
    isProcessing.value = false;
  }
}

async function rejectSubmission() {
  if (!reviewerNotes.value.trim()) return;
  isProcessing.value = true;
  try {
    const res = await fetch(`/api/v1/admin/reviews/${submission.value.appId}/decision`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        decision: 'REJECTED',
        notes: reviewerNotes.value,
        checklist: checklist.value
      })
    });
    if (res.ok) {
      alert('Submission rejected and notification sent to developer.');
    }
  } catch (err) {
    console.error(err);
  } finally {
    isProcessing.value = false;
  }
}
</script>

<style scoped>
.metadata-review-view {
  padding: 24px;
  max-width: 1100px;
  margin: 0 auto;
}
.view-header {
  margin-bottom: 24px;
}
.subtitle {
  color: #64748b;
  font-size: 14px;
}
.review-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}
.app-details-panel, .checklist-panel {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 24px;
}
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  border-bottom: 1px solid #e2e8f0;
  padding-bottom: 16px;
}
.app-identity {
  display: flex;
  gap: 12px;
  align-items: center;
}
.app-icon-preview {
  font-size: 32px;
  background: #f1f5f9;
  padding: 8px;
  border-radius: 8px;
}
.app-identity h3 {
  margin: 0 0 4px 0;
}
.metadata-section {
  margin-bottom: 16px;
}
.metadata-section label {
  font-size: 11px;
  text-transform: uppercase;
  color: #64748b;
  font-weight: 600;
  display: block;
  margin-bottom: 4px;
}
.desc-text {
  font-size: 13px;
  line-height: 1.5;
  color: #334155;
}
.link {
  color: #2563eb;
  text-decoration: none;
  font-size: 13px;
}
.asset-grid {
  display: flex;
  gap: 8px;
  margin-top: 6px;
}
.asset-placeholder {
  width: 90px;
  height: 150px;
  background: #f1f5f9;
  border: 1px dashed #cbd5e1;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  color: #64748b;
  text-align: center;
  padding: 4px;
}
.checklist-desc {
  font-size: 12px;
  color: #64748b;
  margin-bottom: 16px;
}
.checklist-items {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 20px;
}
.check-item {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  padding: 12px;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  cursor: pointer;
  background: #f8fafc;
  transition: all 0.15s;
}
.check-item.checked {
  border-color: #22c55e;
  background: #f0fdf4;
}
.check-item input {
  margin-top: 3px;
  cursor: pointer;
}
.check-text strong {
  display: block;
  font-size: 13px;
  margin-bottom: 2px;
}
.check-text span {
  font-size: 12px;
  color: #64748b;
  line-height: 1.4;
}
.notes-box {
  margin-bottom: 20px;
}
.notes-box label {
  font-size: 11px;
  text-transform: uppercase;
  color: #64748b;
  font-weight: 600;
  display: block;
  margin-bottom: 4px;
}
.notes-box textarea {
  width: 100%;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  padding: 8px;
  font-family: inherit;
  font-size: 13px;
  box-sizing: border-box;
}
.action-buttons {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}
.btn {
  padding: 10px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border: none;
}
.btn-primary { background: #2563eb; color: white; }
.btn-primary:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-danger { background: #dc2626; color: white; }
.btn-danger:disabled { opacity: 0.4; cursor: not-allowed; }
.checklist-warning {
  margin-top: 12px;
  font-size: 12px;
  color: #b45309;
}
.badge {
  padding: 3px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
}
.badge-warning { background: #fef3c7; color: #92400e; }
.badge-success { background: #dcfce7; color: #166534; }
.font-xs { font-size: 10px; }
.text-muted { color: #64748b; }
</style>

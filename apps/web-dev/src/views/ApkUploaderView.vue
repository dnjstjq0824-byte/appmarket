<template>
  <div class="apk-uploader-view">
    <div class="view-header">
      <h2>Release Binary Uploader (APK / AAB)</h2>
      <p class="subtitle">Secure multipart upload, client-side SHA-256 integrity verification, and binary validation</p>
    </div>

    <!-- Org selector -->
    <div v-if="orgs.length > 0" class="org-row">
      <label for="orgSelect">Organization</label>
      <select id="orgSelect" :value="activeOrgId" @change="selectOrg(($event.target as HTMLSelectElement).value)">
        <option v-for="org in orgs" :key="org.id" :value="org.id">{{ org.org_name }} ({{ org.verification_status }})</option>
      </select>
    </div>
    <div v-if="orgLoadError" class="alert alert-error">
      <div class="alert-content">
        <h4>Failed to load organizations</h4>
        <p>{{ orgLoadError }}</p>
      </div>
    </div>

    <!-- KYB Gate Alert if not verified -->
    <div v-if="kybStatus !== 'VERIFIED'" class="kyb-gate-banner">
      <strong>조직 KYB 미검증 상태 (업로드 제한)</strong>
      <p>보안 규정에 따라 법인 신원 검증(KYB)이 완료되지 않은 조직은 릴리스 바이너리를 업로드할 수 없습니다. <router-link to="/">[조직 & KYB 인증]</router-link> 페이지에서 서류를 제출해 주십시오.</p>
    </div>

    <!-- Drop Zone -->
    <div 
      class="upload-dropzone" 
      :class="{ 'dragging': isDragging, 'has-file': !!selectedFile, 'disabled-zone': kybStatus !== 'VERIFIED' }"
      @dragover.prevent="kybStatus === 'VERIFIED' && (isDragging = true)"
      @dragleave.prevent="isDragging = false"
      @drop.prevent="kybStatus === 'VERIFIED' && handleFileDrop($event)"
    >
      <input 
        type="file" 
        id="fileInput" 
        accept=".apk,.aab" 
        @change="handleFileInput" 
        class="hidden-file-input"
        :disabled="uploadState === 'UPLOADING' || uploadState === 'HASHING' || kybStatus !== 'VERIFIED'"
      />
      <label for="fileInput" class="dropzone-label">
        <div class="drop-icon" aria-hidden="true"><svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3v12" /><path d="M7 10l5 5 5-5" /><path d="M4 21h16" /></svg></div>
        <div v-if="!selectedFile" class="prompt-text">
          <strong>Drag & Drop APK or AAB bundle here</strong>
          <span>or click to browse from local computer (Max: 4GB)</span>
        </div>
        <div v-else class="selected-file-info">
          <span class="file-name">{{ selectedFile.name }}</span>
          <span class="file-size">({{ formatFileSize(selectedFile.size) }})</span>
        </div>
      </label>
    </div>

    <!-- Hashing & Upload Progress Bar -->
    <div v-if="uploadState === 'HASHING' || uploadState === 'UPLOADING'" class="progress-section">
      <div class="progress-labels">
        <span>{{ uploadState === 'HASHING' ? 'Calculating client SHA-256 fingerprint...' : 'Uploading multipart chunks...' }}</span>
        <span>{{ currentProgress }}%</span>
      </div>
      <div class="progress-track">
        <div class="progress-bar" :style="{ width: currentProgress + '%' }"></div>
      </div>
      <div class="upload-stats">
        <span v-if="computedSha256">SHA-256: <code>{{ computedSha256 }}</code></span>
        <button v-if="uploadState === 'UPLOADING'" @click="abortUpload" class="btn btn-secondary btn-sm">Cancel</button>
      </div>
    </div>

    <!-- Validation Error Alert -->
    <div v-if="errorMessage" class="alert alert-error">
      <div class="alert-icon" aria-hidden="true"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.3 3.9 1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0z"/><path d="M12 9v4"/><path d="M12 17h.01"/></svg></div>
      <div class="alert-content">
        <h4>{{ errorCode || 'Validation Error' }}</h4>
        <p>{{ errorMessage }}</p>
        <div v-if="resolutionGuide" class="guide-box">
          <strong>How to resolve:</strong> {{ resolutionGuide }}
        </div>
      </div>
      <button @click="resetUploader" class="btn btn-sm btn-outline">Try Again</button>
    </div>

    <!-- Metadata-only completion receipt (edge fallback path) -->
    <div v-if="uploadReceipt && uploadState === 'COMPLETED'" class="binary-details-card">
      <div class="card-header">
        <h3>Binary Received</h3>
        <span class="badge badge-success">Upload {{ uploadReceipt.uploadId }} · {{ uploadReceipt.status }}</span>
      </div>
      <p class="perm-notice">Manifest parsing runs on the connected backend. Reconnect to the full API to inspect extracted metadata.</p>
      <div class="card-actions">
        <button @click="resetUploader" class="btn btn-secondary">Upload Another</button>
      </div>
    </div>

    <!-- Extracted Manifest & Binary Details -->
    <div v-if="binaryDetails && uploadState === 'COMPLETED'" class="binary-details-card">
      <div class="card-header">
        <h3>Extracted Binary Metadata</h3>
        <span class="badge badge-success">✓ Verified v2/v3 Signed</span>
      </div>

      <div class="grid-details">
        <div class="detail-item">
          <label>Package Name</label>
          <div class="value font-mono">{{ binaryDetails.packageName }}</div>
        </div>
        <div class="detail-item">
          <label>Version Code</label>
          <div class="value font-mono">{{ binaryDetails.versionCode }}</div>
        </div>
        <div class="detail-item">
          <label>Version Name</label>
          <div class="value">{{ binaryDetails.versionName }}</div>
        </div>
        <div class="detail-item">
          <label>SDK Target</label>
          <div class="value">Min SDK: {{ binaryDetails.minSdk }} / Target SDK: {{ binaryDetails.targetSdk }}</div>
        </div>
        <div class="detail-item col-span-2">
          <label>SHA-256 Digest</label>
          <div class="value font-mono hash-box">{{ binaryDetails.sha256 }}</div>
        </div>
        <div class="detail-item col-span-2">
          <label>Certificate SHA-256</label>
          <div class="value font-mono hash-box">{{ binaryDetails.certSha256 }}</div>
        </div>
      </div>

      <!-- Dangerous Permissions Warning -->
      <div v-if="binaryDetails.dangerousPermissions?.length" class="permissions-section">
        <h4>Declared Dangerous Permissions ({{ binaryDetails.dangerousPermissions.length }})</h4>
        <div class="perm-tags">
          <span 
            v-for="perm in binaryDetails.dangerousPermissions" 
            :key="perm" 
            class="tag tag-warning"
          >
            {{ perm }}
          </span>
        </div>
        <p class="perm-notice">These permissions will be flagged for compliance review during release submission.</p>
      </div>

      <!-- Action Buttons -->
      <div class="card-actions">
        <button @click="resetUploader" class="btn btn-secondary">Upload Another</button>
        <button @click="proceedToRelease" class="btn btn-primary">Proceed to Release Track</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { getSharedClient, ApiClientError, type OrgSummary } from '@bookingnetwork/api-client';

type UploadState = 'IDLE' | 'HASHING' | 'UPLOADING' | 'COMPLETED' | 'ERROR';

const router = useRouter();
const api = getSharedClient();

const orgs = ref<OrgSummary[]>([]);
const activeOrgId = ref('');
const kybStatus = ref('UNKNOWN');
const orgLoadError = ref('');

onMounted(async () => {
  try {
    orgs.value = await api.listOrgs();
    if (orgs.value.length > 0) {
      activeOrgId.value = orgs.value[0].id;
      kybStatus.value = orgs.value[0].verification_status || 'UNVERIFIED';
    } else {
      kybStatus.value = 'NO_ORG';
    }
  } catch (err) {
    kybStatus.value = 'LOAD_FAILED';
    orgLoadError.value = err instanceof ApiClientError ? err.message : 'Failed to load organizations';
  }
});

function selectOrg(id: string) {
  activeOrgId.value = id;
  const org = orgs.value.find(o => o.id === id);
  kybStatus.value = org?.verification_status || 'UNVERIFIED';
}

const isDragging = ref(false);
const selectedFile = ref<File | null>(null);
const uploadState = ref<UploadState>('IDLE');
const currentProgress = ref(0);
const computedSha256 = ref('');
const errorMessage = ref('');
const errorCode = ref('');
const resolutionGuide = ref('');
const uploadReceipt = ref<{ uploadId: string; status: string } | null>(null);

const binaryDetails = ref<{
  packageName: string;
  versionCode: number;
  versionName: string;
  minSdk: number;
  targetSdk: number;
  sha256: string;
  certSha256: string;
  dangerousPermissions: string[];
} | null>(null);

let activeWorker: Worker | null = null;

function handleFileInput(e: Event) {
  const target = e.target as HTMLInputElement;
  if (target.files && target.files.length > 0) {
    processFile(target.files[0]);
  }
}

function handleFileDrop(e: DragEvent) {
  isDragging.value = false;
  if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
    processFile(e.dataTransfer.files[0]);
  }
}

async function processFile(file: File) {
  selectedFile.value = file;
  uploadState.value = 'HASHING';
  currentProgress.value = 0;
  errorMessage.value = '';
  errorCode.value = '';

  try {
    // Run chunked SHA-256 in Web Worker
    activeWorker = new Worker(new URL('../workers/sha256.worker.ts', import.meta.url), { type: 'module' });
    activeWorker.postMessage({ file });

    activeWorker.onmessage = async (e) => {
      const { type, progress, sha256, error } = e.data;
      if (type === 'PROGRESS') {
        currentProgress.value = Math.round(progress / 2); // 0%~50% for hashing
      } else if (type === 'COMPLETE') {
        computedSha256.value = sha256;
        activeWorker?.terminate();
        activeWorker = null;
        await startMultipartUpload(file, sha256);
      } else if (type === 'ERROR') {
        uploadState.value = 'ERROR';
        errorMessage.value = error || 'Failed to compute SHA-256 fingerprint';
        activeWorker?.terminate();
        activeWorker = null;
      }
    };
  } catch (err: any) {
    uploadState.value = 'ERROR';
    errorMessage.value = err?.message || 'Worker initialization failed';
  }
}

interface CompleteUploadResponse {
  uploadId: string;
  status: string;
  details?: {
    packageName: string;
    versionCode: number;
    versionName: string;
    minSdk: number;
    targetSdk: number;
    sha256: string;
    certSha256: string;
    dangerousPermissions: string[];
  };
}

async function startMultipartUpload(file: File, sha256: string) {
  uploadState.value = 'UPLOADING';

  try {
    // Step 1: Initiate upload session with backend
    const session = await api.initUpload('current', {
      orgId: activeOrgId.value,
      fileName: file.name,
      fileSize: file.size,
      expectedSha256: sha256
    });

    // Step 2: Stream real bytes through the quarantine pipeline
    currentProgress.value = 50;
    const result = await api.completeUpload('current', session.uploadId, sha256, file) as CompleteUploadResponse;
    currentProgress.value = 100;

    if (!result.details) {
      uploadReceipt.value = { uploadId: result.uploadId, status: result.status };
      uploadState.value = 'COMPLETED';
      return;
    }
    uploadReceipt.value = null;
    binaryDetails.value = {
      packageName: result.details.packageName,
      versionCode: result.details.versionCode,
      versionName: result.details.versionName,
      minSdk: result.details.minSdk,
      targetSdk: result.details.targetSdk,
      sha256: result.details.sha256 || sha256,
      certSha256: result.details.certSha256,
      dangerousPermissions: result.details.dangerousPermissions || []
    };
    uploadState.value = 'COMPLETED';
  } catch (err: any) {
    const code = err instanceof ApiClientError ? (err.problem?.code || '') : (err.code || '');
    if (code === 'ERR_APK_V2_SIGNATURE_REQUIRED') {
      errorCode.value = 'ERR_APK_V2_SIGNATURE_REQUIRED';
      errorMessage.value = 'This APK is signed with APK Signature Scheme v1 only. APK Signature Scheme v2 or v3 is strictly required.';
      resolutionGuide.value = 'Sign the APK using apksigner from Android SDK Build-Tools 24.0.3+ with flags --v2-signing-enabled true.';
      uploadState.value = 'ERROR';
      return;
    }
    uploadState.value = 'ERROR';
    errorCode.value = code || 'ERR_UPLOAD_FAILURE';
    errorMessage.value = err.message || 'An unexpected error occurred during upload.';
  }
}

function abortUpload() {
  if (activeWorker) {
    activeWorker.terminate();
    activeWorker = null;
  }
  resetUploader();
}

function resetUploader() {
  uploadState.value = 'IDLE';
  selectedFile.value = null;
  currentProgress.value = 0;
  computedSha256.value = '';
  errorMessage.value = '';
  errorCode.value = '';
  binaryDetails.value = null;
  uploadReceipt.value = null;
}

function proceedToRelease() {
  router.push('/releases/rollout');
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
</script>

<style scoped>
.apk-uploader-view {
  max-width: 920px;
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
.org-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  font-size: 13px;
  color: var(--bn-text-secondary);
}
.org-row label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
}
.org-row select {
  padding: 8px 12px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
}
.kyb-gate-banner {
  background: var(--bn-warning-wash);
  border: 1px solid var(--bn-warning-amber);
  border-radius: var(--bn-radius-md);
  padding: 14px 18px;
  margin-bottom: 20px;
  font-size: 13px;
}
.kyb-gate-banner strong {
  color: var(--bn-warning-amber);
}
.kyb-gate-banner p {
  margin: 6px 0 0;
  color: var(--bn-text-secondary);
  line-height: 1.6;
}
.kyb-gate-banner a {
  color: var(--bn-text-primary);
  font-weight: 700;
}
.upload-dropzone {
  border: 1px dashed var(--bn-border-strong);
  border-radius: var(--bn-radius-md);
  padding: 44px 24px;
  text-align: center;
  background: var(--bn-bg-surface);
  cursor: pointer;
}
.upload-dropzone.dragging {
  border-color: var(--bn-seal);
  border-style: solid;
}
.upload-dropzone.has-file {
  border-style: solid;
  border-color: var(--bn-trust-verified);
}
.upload-dropzone.disabled-zone {
  opacity: 0.5;
  cursor: not-allowed;
}
.hidden-file-input {
  display: none;
}
.dropzone-label {
  display: block;
  cursor: pointer;
}
.disabled-zone .dropzone-label {
  cursor: not-allowed;
}
.drop-icon {
  color: var(--bn-text-muted);
  margin-bottom: 12px;
}
.dragging .drop-icon {
  color: var(--bn-seal);
}
.has-file .drop-icon {
  display: none;
}
.prompt-text strong {
  display: block;
  font-size: 15px;
  margin-bottom: 4px;
}
.prompt-text span {
  font-size: 13px;
  color: var(--bn-text-muted);
}
.selected-file-info {
  font-weight: 600;
  font-size: 14px;
}
.file-size {
  font-family: var(--bn-font-mono);
  font-size: 12px;
  font-weight: 400;
  color: var(--bn-text-muted);
}
.progress-section {
  margin-top: 20px;
  background: var(--bn-bg-surface);
  padding: 18px;
  border-radius: var(--bn-radius-md);
  border: 1px solid var(--bn-border-subtle);
}
.progress-labels {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
  font-size: 13px;
}
.progress-labels span:last-child {
  font-family: var(--bn-font-mono);
  color: var(--bn-text-secondary);
}
.progress-track {
  height: 6px;
  background: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  overflow: hidden;
}
.progress-bar {
  height: 100%;
  background: var(--bn-ink);
}
.upload-stats {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-top: 12px;
  font-size: 12px;
  color: var(--bn-text-secondary);
}
.upload-stats code {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  word-break: break-all;
}
.alert {
  border-radius: var(--bn-radius-md);
}
.alert-error {
  margin-top: 20px;
  padding: 18px;
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  display: flex;
  gap: 16px;
  align-items: flex-start;
}
.alert-icon {
  color: var(--bn-danger-red);
  flex-shrink: 0;
}
.alert-content {
  flex: 1;
}
.alert-content h4 {
  margin: 0 0 6px 0;
  font-size: 14px;
  color: var(--bn-danger-red);
}
.alert-content p {
  margin: 0;
  color: var(--bn-text-secondary);
  font-size: 13px;
  line-height: 1.6;
}
.guide-box {
  margin-top: 10px;
  padding: 10px 12px;
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  font-size: 13px;
  font-family: var(--bn-font-mono);
}
.binary-details-card {
  margin-top: 24px;
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--bn-border-subtle);
  flex-wrap: wrap;
}
.card-header h3 {
  font-family: var(--bn-font-display);
  font-size: 15px;
  font-weight: 800;
  margin: 0;
}
.badge {
  display: inline-block;
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.04em;
  padding: 4px 10px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-secondary);
}
.badge-success {
  border-color: var(--bn-trust-verified);
  color: var(--bn-trust-verified);
}
.grid-details {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}
.col-span-2 {
  grid-column: span 2;
}
.detail-item label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  margin-bottom: 6px;
  display: block;
}
.detail-item .value {
  font-size: 14px;
}
.font-mono {
  font-family: var(--bn-font-mono);
}
.hash-box {
  background: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  padding: 10px 12px;
  border-radius: var(--bn-radius-sm);
  word-break: break-all;
  font-size: 12px;
}
.permissions-section {
  border-top: 1px solid var(--bn-border-subtle);
  padding-top: 16px;
  margin-bottom: 20px;
}
.permissions-section h4 {
  font-size: 13px;
  margin: 0;
}
.perm-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 10px 0;
}
.tag-warning {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  border: 1px solid var(--bn-warning-amber);
  color: var(--bn-warning-amber);
  padding: 3px 9px;
  border-radius: var(--bn-radius-sm);
}
.perm-notice {
  font-size: 12px;
  color: var(--bn-text-muted);
  margin: 4px 0 0;
}
.card-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  border-top: 1px solid var(--bn-border-subtle);
  padding-top: 16px;
}
.btn {
  padding: 9px 18px;
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
.btn-primary:hover {
  background: var(--bn-action-bg-hover);
}
.btn-secondary {
  background: transparent;
  border: 1px solid var(--bn-border-subtle);
  color: var(--bn-text-secondary);
}
.btn-secondary:hover {
  border-color: var(--bn-border-strong);
  color: var(--bn-text-primary);
}
.btn-outline {
  border: 1px solid var(--bn-danger-red);
  background: transparent;
  color: var(--bn-danger-red);
  flex-shrink: 0;
}
.btn-sm {
  padding: 5px 12px;
  font-size: 12px;
}
@media (max-width: 640px) {
  .grid-details {
    grid-template-columns: 1fr;
  }
  .col-span-2 {
    grid-column: span 1;
  }
}
</style>


<template>
  <div class="staged-rollout-view">
    <div class="view-header">
      <h2>Staged Rollout & Release Control</h2>
      <p class="subtitle">Gradual rollout management, bucket evaluation, and instant emergency halt</p>
    </div>

    <!-- Active Release Status Card -->
    <div class="release-info-card">
      <div class="info-header">
        <div>
          <h3>Release v{{ release.versionName }} (Build {{ release.versionCode }})</h3>
          <span class="track-tag">{{ release.track }}</span>
        </div>
        <div class="status-indicator">
          <span class="badge" :class="statusBadgeClass">{{ release.status }}</span>
        </div>
      </div>

      <div v-if="isLoading" class="loading-row">Loading current release…</div>
      <div v-if="loadError" class="alert alert-danger">{{ loadError }}</div>
      <div class="rollout-stats-grid">
        <div class="stat-box">
          <label>Target Percentage</label>
          <div class="stat-value">{{ currentPercentage }}%</div>
        </div>
        <div class="stat-box">
          <label>Version</label>
          <div class="stat-value font-mono">{{ release.versionName }} ({{ release.versionCode }})</div>
        </div>
        <div class="stat-box">
          <label>Track</label>
          <div class="stat-value">{{ release.track }}</div>
        </div>
        <div class="stat-box">
          <label>Distribution State</label>
          <div class="stat-value">{{ isHalted ? 'HALTED' : 'SERVING' }}</div>
        </div>
      </div>
    </div>

    <!-- Rollout Adjustment Section -->
    <div class="rollout-control-card">
      <h3>Adjust Rollout Percentage</h3>
      <p class="section-desc">
        Rollout percentages are monotonically increasing. To stop rollout due to critical regressions, use Emergency Halt.
      </p>

      <div class="stages-selector">
        <div 
          v-for="stage in availableStages" 
          :key="stage"
          class="stage-chip"
          :class="{
            'active': selectedPercentage === stage,
            'disabled': stage < currentPercentage || isHalted
          }"
          @click="selectStage(stage)"
        >
          {{ stage }}%
        </div>
      </div>

      <div class="slider-container">
        <input 
          type="range" 
          min="1" 
          max="100" 
          step="1"
          v-model.number="selectedPercentage"
          :disabled="isHalted"
          class="rollout-range"
        />
        <div class="range-marks">
          <span>1% (Canary)</span>
          <span>10% (Early)</span>
          <span>50% (Wide)</span>
          <span>100% (Full)</span>
        </div>
      </div>

      <!-- Action Confirmation -->
      <div class="control-actions">
        <button 
          @click="applyRolloutUpdate" 
          class="btn btn-primary"
          :disabled="selectedPercentage <= currentPercentage || isUpdating || isHalted"
        >
          {{ isUpdating ? 'Updating...' : `Update Rollout to ${selectedPercentage}%` }}
        </button>

        <button
          v-if="!isHalted"
          @click="triggerEmergencyHalt"
          class="btn btn-danger"
          :disabled="isUpdating"
        >
          Emergency Halt Rollout
        </button>
        <button
          v-else
          @click="resumeRollout"
          class="btn btn-warning"
          :disabled="isUpdating"
        >
          Resume Rollout
        </button>
      </div>
      <div v-if="actionError" class="alert alert-danger">{{ actionError }}</div>

      <!-- Emergency Halt Notice -->
      <div v-if="isHalted" class="alert alert-danger">
        <strong>Rollout is currently HALTED.</strong> No new devices will receive this update until resumed or patched.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { getSharedClient, ApiClientError, type ReleaseDTO } from '@bookingnetwork/api-client';

const api = getSharedClient();

const release = ref({
  versionName: '-',
  versionCode: 0,
  track: '-',
  status: 'LOADING'
});

const currentPercentage = ref(0);
const selectedPercentage = ref(0);
const isHalted = ref(false);
const isUpdating = ref(false);
const isLoading = ref(true);
const loadError = ref('');

onMounted(async () => {
  try {
    const current: ReleaseDTO = await api.getCurrentRelease('current');
    applyRelease(current);
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Failed to load current release';
    release.value.status = 'LOAD_FAILED';
  } finally {
    isLoading.value = false;
  }
});

function applyRelease(current: ReleaseDTO) {
  release.value = {
    versionName: current.versionName,
    versionCode: current.versionCode,
    track: current.track,
    status: current.status
  };
  currentPercentage.value = current.rolloutPercentage;
  selectedPercentage.value = current.rolloutPercentage;
  isHalted.value = current.status === 'HALTED' || !!current.isHalted;
}

const availableStages = [1, 5, 10, 20, 50, 100];

const statusBadgeClass = computed(() => {
  if (isHalted.value) return 'badge-danger';
  if (currentPercentage.value === 100) return 'badge-success';
  return 'badge-primary';
});

function selectStage(stage: number) {
  if (stage >= currentPercentage.value && !isHalted.value) {
    selectedPercentage.value = stage;
  }
}

const actionError = ref('');

async function applyRolloutUpdate() {
  if (selectedPercentage.value <= currentPercentage.value) return;
  isUpdating.value = true;
  actionError.value = '';
  try {
    const updated = await api.updateRollout('current', { rolloutPercentage: selectedPercentage.value });
    applyRelease(updated);
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to update rollout';
  } finally {
    isUpdating.value = false;
  }
}

async function triggerEmergencyHalt() {
  if (!confirm('Are you sure you want to trigger EMERGENCY HALT? This will immediately suspend all distribution.')) {
    return;
  }
  isUpdating.value = true;
  actionError.value = '';
  try {
    const updated = await api.updateRollout('current', { action: 'HALT' });
    applyRelease(updated);
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to halt rollout';
  } finally {
    isUpdating.value = false;
  }
}

async function resumeRollout() {
  isUpdating.value = true;
  actionError.value = '';
  try {
    const updated = await api.updateRollout('current', { action: 'RESUME' });
    applyRelease(updated);
  } catch (err) {
    actionError.value = err instanceof ApiClientError ? err.message : 'Failed to resume rollout';
  } finally {
    isUpdating.value = false;
  }
}
</script>

<style scoped>
.staged-rollout-view {
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
.loading-row {
  padding: 24px;
  color: var(--bn-text-muted);
  font-size: 14px;
  text-align: center;
}
.release-info-card, .rollout-control-card {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 24px;
  margin-bottom: 24px;
}
.rollout-control-card > h3 {
  font-family: var(--bn-font-display);
  font-size: 15px;
  font-weight: 800;
  margin: 0 0 6px;
}
.info-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--bn-border-subtle);
  flex-wrap: wrap;
}
.info-header h3 {
  font-family: var(--bn-font-display);
  font-size: 17px;
  font-weight: 800;
  margin: 0 0 6px 0;
}
.track-tag {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.06em;
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-secondary);
  padding: 3px 9px;
  border-radius: var(--bn-radius-sm);
}
.badge {
  font-family: var(--bn-font-mono);
  font-size: 11px;
  letter-spacing: 0.04em;
  padding: 4px 12px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-secondary);
  white-space: nowrap;
}
.badge-primary { border-color: var(--bn-seal); color: var(--bn-seal); }
.badge-success { border-color: var(--bn-trust-verified); color: var(--bn-trust-verified); }
.badge-danger { border-color: var(--bn-danger-red); color: var(--bn-danger-red); }
.rollout-stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}
.stat-box {
  background: var(--bn-bg-well);
  border: 1px solid var(--bn-border-subtle);
  padding: 14px;
  border-radius: var(--bn-radius-sm);
}
.stat-box label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.08em;
  color: var(--bn-text-muted);
  display: block;
  margin-bottom: 6px;
}
.stat-value {
  font-family: var(--bn-font-display);
  font-size: 22px;
  font-weight: 900;
  letter-spacing: -0.02em;
}
.stat-value.font-mono {
  font-family: var(--bn-font-mono);
  font-size: 15px;
  font-weight: 600;
}
.section-desc {
  font-size: 13px;
  line-height: 1.6;
  color: var(--bn-text-secondary);
  margin: 0 0 20px;
}
.stages-selector {
  display: flex;
  gap: 0;
  margin-bottom: 24px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  overflow: hidden;
  width: fit-content;
  max-width: 100%;
  flex-wrap: wrap;
}
.stage-chip {
  padding: 9px 18px;
  border: none;
  border-left: 1px solid var(--bn-border-subtle);
  cursor: pointer;
  font-family: var(--bn-font-mono);
  font-size: 13px;
  color: var(--bn-text-secondary);
  background: transparent;
}
.stage-chip:first-child {
  border-left: none;
}
.stage-chip.active {
  background: var(--bn-ink);
  color: var(--bn-bg-canvas);
  font-weight: 700;
}
.stage-chip.disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
.slider-container {
  margin-bottom: 24px;
}
.rollout-range {
  width: 100%;
  accent-color: var(--bn-seal);
  cursor: pointer;
}
.rollout-range:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}
.range-marks {
  display: flex;
  justify-content: space-between;
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.04em;
  color: var(--bn-text-muted);
  margin-top: 8px;
}
.control-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}
.btn {
  padding: 11px 22px;
  border-radius: var(--bn-radius-sm);
  font-size: 14px;
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
.btn-primary:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.btn-danger {
  background: var(--bn-danger-red);
  color: #fff;
  border: none;
}
.btn-danger:hover:not(:disabled) {
  filter: brightness(0.92);
}
.btn-danger:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.btn-warning {
  background: transparent;
  border: 1px solid var(--bn-warning-amber);
  color: var(--bn-warning-amber);
}
.btn-warning:hover:not(:disabled) {
  background: var(--bn-warning-wash);
}
.btn-warning:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.alert {
  border-radius: var(--bn-radius-sm);
}
.alert-danger {
  margin-top: 16px;
  padding: 12px 14px;
  background: var(--bn-danger-wash);
  border: 1px solid var(--bn-danger-red);
  color: var(--bn-danger-red);
  font-size: 13px;
  line-height: 1.6;
}
@media (max-width: 720px) {
  .rollout-stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>


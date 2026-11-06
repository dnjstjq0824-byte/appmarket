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
  padding: 24px;
  max-width: 900px;
  margin: 0 auto;
}
.view-header {
  margin-bottom: 24px;
}
.subtitle {
  color: #64748b;
  font-size: 14px;
}
.loading-row {
  padding: 16px;
  color: #64748b;
  font-size: 14px;
}
.release-info-card, .rollout-control-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 24px;
}
.info-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.info-header h3 {
  margin: 0 0 4px 0;
}
.track-tag {
  background: #f1f5f9;
  color: #475569;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}
.badge {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}
.badge-primary { background: #eff6ff; color: #1d4ed8; }
.badge-success { background: #f0fdf4; color: #166534; }
.badge-danger { background: #fef2f2; color: #991b1b; }
.rollout-stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.stat-box {
  background: #f8fafc;
  padding: 12px;
  border-radius: 6px;
}
.stat-box label {
  font-size: 11px;
  color: #64748b;
  display: block;
  margin-bottom: 4px;
  text-transform: uppercase;
}
.stat-value {
  font-size: 18px;
  font-weight: 700;
  color: #0f172a;
}
.text-success { color: #16a34a; }
.section-desc {
  font-size: 13px;
  color: #64748b;
  margin-bottom: 20px;
}
.stages-selector {
  display: flex;
  gap: 10px;
  margin-bottom: 24px;
}
.stage-chip {
  padding: 8px 16px;
  border: 1px solid #cbd5e1;
  border-radius: 20px;
  cursor: pointer;
  font-weight: 600;
  font-size: 13px;
  background: white;
  transition: all 0.15s;
}
.stage-chip.active {
  background: #2563eb;
  color: white;
  border-color: #2563eb;
}
.stage-chip.disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.slider-container {
  margin-bottom: 24px;
}
.rollout-range {
  width: 100%;
  height: 6px;
  accent-color: #2563eb;
  cursor: pointer;
}
.range-marks {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #64748b;
  margin-top: 6px;
}
.control-actions {
  display: flex;
  gap: 12px;
}
.btn {
  padding: 10px 20px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  border: none;
}
.btn-primary { background: #2563eb; color: white; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-danger { background: #dc2626; color: white; }
.btn-warning { background: #f59e0b; color: white; }
.alert-danger {
  margin-top: 16px;
  padding: 12px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #991b1b;
  border-radius: 6px;
  font-size: 13px;
}
</style>


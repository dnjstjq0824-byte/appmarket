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

      <div class="rollout-stats-grid">
        <div class="stat-box">
          <label>Target Percentage</label>
          <div class="stat-value">{{ currentPercentage }}%</div>
        </div>
        <div class="stat-box">
          <label>Estimated Reached Users</label>
          <div class="stat-value font-mono">{{ estimatedUsers.toLocaleString() }}</div>
        </div>
        <div class="stat-box">
          <label>Crash-Free Sessions</label>
          <div class="stat-value text-success font-mono">99.82%</div>
        </div>
        <div class="stat-box">
          <label>Hash Bucketing Algorithm</label>
          <div class="stat-value font-mono text-sm">CRC32(app_id + device_id)</div>
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
          🚨 Emergency Halt Rollout
        </button>
        <button 
          v-else 
          @click="resumeRollout" 
          class="btn btn-warning"
          :disabled="isUpdating"
        >
          ▶️ Resume Rollout
        </button>
      </div>

      <!-- Emergency Halt Notice -->
      <div v-if="isHalted" class="alert alert-danger">
        <strong>Rollout is currently HALTED.</strong> No new devices will receive this update until resumed or patched.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const release = ref({
  versionName: '2.5.0',
  versionCode: 20500,
  track: 'PRODUCTION',
  status: 'ROLLOUT_IN_PROGRESS'
});

const currentPercentage = ref(10);
const selectedPercentage = ref(10);
const isHalted = ref(false);
const isUpdating = ref(false);

const availableStages = [1, 5, 10, 20, 50, 100];

const statusBadgeClass = computed(() => {
  if (isHalted.value) return 'badge-danger';
  if (currentPercentage.value === 100) return 'badge-success';
  return 'badge-primary';
});

const estimatedUsers = computed(() => {
  const totalBase = 250000;
  if (isHalted.value) return 0;
  return Math.round((totalBase * currentPercentage.value) / 100);
});

function selectStage(stage: number) {
  if (stage >= currentPercentage.value && !isHalted.value) {
    selectedPercentage.value = stage;
  }
}

async function applyRolloutUpdate() {
  if (selectedPercentage.value <= currentPercentage.value) return;
  isUpdating.value = true;
  try {
    const res = await fetch('/api/v1/developer/releases/current/rollout', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ rolloutPercentage: selectedPercentage.value })
    });
    if (res.ok) {
      currentPercentage.value = selectedPercentage.value;
      if (currentPercentage.value === 100) {
        release.value.status = 'COMPLETED';
      }
    }
  } catch (err) {
    console.error('Failed to update rollout', err);
  } finally {
    isUpdating.value = false;
  }
}

async function triggerEmergencyHalt() {
  if (!confirm('Are you sure you want to trigger EMERGENCY HALT? This will immediately suspend all distribution.')) {
    return;
  }
  isUpdating.value = true;
  try {
    const res = await fetch('/api/v1/developer/releases/current/rollout', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ action: 'HALT' })
    });
    if (res.ok) {
      isHalted.value = true;
      release.value.status = 'HALTED';
    }
  } catch (err) {
    console.error('Failed to halt rollout', err);
  } finally {
    isUpdating.value = false;
  }
}

async function resumeRollout() {
  isUpdating.value = true;
  try {
    const res = await fetch('/api/v1/developer/releases/current/rollout', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ action: 'RESUME', rolloutPercentage: currentPercentage.value })
    });
    if (res.ok) {
      isHalted.value = false;
      release.value.status = 'ROLLOUT_IN_PROGRESS';
    }
  } catch (err) {
    console.error('Failed to resume rollout', err);
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


<template>
  <div class="account-library-container">
    <div class="library-header">
      <h2>My App Library & Entitlements (USR-18)</h2>
      <p class="subtitle">Permanent purchase records, licensed applications, and secure re-downloads</p>
    </div>

    <div class="library-stats">
      <div class="stat-pill">
        <label>Total Entitlements</label>
        <span class="val font-mono">{{ libraryItems.length }}</span>
      </div>
      <div class="stat-pill">
        <label>Installed on Device</label>
        <span class="val font-mono">2</span>
      </div>
      <div class="stat-pill">
        <label>Pending Updates</label>
        <span class="val font-mono text-warning">1 Update</span>
      </div>
    </div>

    <!-- Library Table -->
    <div class="library-table-card">
      <table class="library-table">
        <thead>
          <tr>
            <th>Application</th>
            <th>License / Purchase Date</th>
            <th>Installed Build</th>
            <th>Store Latest Build</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in libraryItems" :key="item.appId">
            <td class="app-info-cell">
              <div class="app-icon">📦</div>
              <div>
                <strong>{{ item.title }}</strong>
                <div class="font-mono text-xs text-muted">{{ item.packageName }}</div>
              </div>
            </td>
            <td>
              <span class="badge" :class="item.isPaid ? 'badge-paid' : 'badge-free'">
                {{ item.isPaid ? 'Purchased (₩' + item.price.toLocaleString() + ')' : 'Free License' }}
              </span>
              <div class="date-text text-xs">{{ item.acquiredDate }}</div>
            </td>
            <td class="font-mono text-sm">v{{ item.installedVersionName }}</td>
            <td class="font-mono text-sm">
              v{{ item.latestVersionName }}
              <span v-if="item.updateAvailable" class="badge badge-warning text-xs">Update</span>
            </td>
            <td>
              <div class="action-buttons">
                <button @click="downloadFreshApk(item)" class="btn btn-primary btn-sm">
                  📥 Download APK
                </button>
                <NuxtLink :to="'/apps/' + item.slug" class="btn btn-outline btn-sm">
                  Details
                </NuxtLink>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const libraryItems = ref([
  {
    appId: 'app-1',
    slug: 'secure-authenticator-2fa',
    title: 'Secure Authenticator 2FA',
    packageName: 'com.secure.authenticator',
    isPaid: false,
    price: 0,
    acquiredDate: '2026-09-15',
    installedVersionName: '1.2.0',
    latestVersionName: '1.2.0',
    updateAvailable: false
  },
  {
    appId: 'app-2',
    slug: 'pro-crypto-tracker',
    title: 'Pro Crypto Tracker & Portfolio',
    packageName: 'com.crypto.protracker',
    isPaid: true,
    price: 4900,
    acquiredDate: '2026-10-01',
    installedVersionName: '2.0.1',
    latestVersionName: '2.1.0',
    updateAvailable: true
  }
]);

async function downloadFreshApk(item: any) {
  try {
    const res = await fetch(`/api/v1/catalog/apps/${item.appId}/releases/latest/download-token`, {
      method: 'POST'
    });
    if (res.ok) {
      const { url } = await res.json();
      window.location.href = url;
    } else {
      alert('Failed to generate download token. Please verify network access.');
    }
  } catch (err) {
    console.error(err);
  }
}
</script>

<style scoped>
.account-library-container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 32px 16px;
}
.library-header {
  margin-bottom: 24px;
}
.subtitle {
  color: #64748b;
  font-size: 14px;
}
.library-stats {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
}
.stat-pill {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 12px 20px;
  flex: 1;
}
.stat-pill label {
  font-size: 11px;
  text-transform: uppercase;
  color: #64748b;
  display: block;
  margin-bottom: 4px;
  font-weight: 600;
}
.stat-pill .val {
  font-size: 20px;
  font-weight: 700;
  color: #0f172a;
}
.library-table-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
}
.library-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.library-table th {
  text-align: left;
  padding: 10px 12px;
  border-bottom: 2px solid #e2e8f0;
  color: #475569;
}
.library-table td {
  padding: 14px 12px;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: middle;
}
.app-info-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}
.app-icon {
  font-size: 28px;
  background: #f8fafc;
  padding: 6px;
  border-radius: 6px;
}
.badge {
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
}
.badge-free { background: #f1f5f9; color: #475569; }
.badge-paid { background: #eff6ff; color: #1d4ed8; }
.badge-warning { background: #fef3c7; color: #92400e; margin-left: 6px; }
.date-text { color: #94a3b8; margin-top: 4px; }
.action-buttons {
  display: flex;
  gap: 8px;
}
.btn {
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
}
.btn-primary { background: #2563eb; color: white; border: none; }
.btn-outline { background: transparent; border: 1px solid #cbd5e1; color: #334155; }
.text-warning { color: #d97706; }
.text-muted { color: #94a3b8; }
.text-xs { font-size: 11px; }
.font-mono { font-family: monospace; }
</style>


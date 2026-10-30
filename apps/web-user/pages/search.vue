<template>
  <div class="search-page-container">
    <div class="search-header-section">
      <h1>Catalog & Multi-Store Search</h1>
      <p class="subtitle">Unified search across native hosted releases and verified external open-source repositories</p>

      <!-- Search Input & Autocomplete Dropdown (USR-03) -->
      <div class="search-box-wrapper">
        <div class="input-with-icon">
          <span class="search-icon">🔍</span>
          <input 
            type="text" 
            v-model="searchQuery" 
            @input="handleInput"
            @keydown.enter="triggerSearch"
            placeholder="Search by app title, developer, or package name..."
            class="main-search-input"
            id="mainSearchInput"
          />
          <button v-if="searchQuery" @click="clearSearch" class="clear-btn">✕</button>
        </div>

        <!-- Autocomplete Dropdown List -->
        <div v-if="showSuggestions && suggestions.length > 0" class="suggestions-dropdown">
          <div 
            v-for="item in suggestions" 
            :key="item.text"
            class="suggestion-row"
            @click="selectSuggestion(item)"
          >
            <span class="suggestion-type-badge">{{ item.type }}</span>
            <span class="suggestion-text">{{ item.text }}</span>
            <span v-if="item.package" class="suggestion-pkg font-mono">{{ item.package }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Content: Facets & Results (USR-02) -->
    <div class="search-layout">
      <!-- Faceted Filters Sidebar -->
      <aside class="facets-sidebar">
        <h3>Filter Results</h3>

        <div class="facet-group">
          <label>Distribution Type</label>
          <div class="filter-options">
            <label class="radio-label">
              <input type="radio" value="ALL" v-model="selectedRightsFilter" @change="triggerSearch" />
              <span>All Sources (Unified)</span>
            </label>
            <label class="radio-label">
              <input type="radio" value="DIRECT_ONLY" v-model="selectedRightsFilter" @change="triggerSearch" />
              <span>Direct APK Download Only</span>
            </label>
          </div>
        </div>

        <div class="facet-group">
          <label>Category</label>
          <select v-model="selectedCategory" @change="triggerSearch" class="facet-select">
            <option value="">All Categories</option>
            <option value="TOOLS">Tools & Utilities</option>
            <option value="FINANCE">Finance & Banking</option>
            <option value="COMMUNICATION">Communication & Messaging</option>
            <option value="PRODUCTIVITY">Productivity</option>
            <option value="GAMES">Games</option>
          </select>
        </div>

        <div class="facet-group">
          <label>Price</label>
          <div class="filter-options">
            <label class="radio-label">
              <input type="radio" value="ALL" v-model="selectedPriceFilter" @change="triggerSearch" />
              <span>All Prices</span>
            </label>
            <label class="radio-label">
              <input type="radio" value="FREE" v-model="selectedPriceFilter" @change="triggerSearch" />
              <span>Free Only</span>
            </label>
          </div>
        </div>
      </aside>

      <!-- Search Results Area -->
      <main class="results-main">
        <div class="results-summary">
          <span v-if="!isLoading">
            Found <strong>{{ totalResults }}</strong> results for "<em>{{ activeQuery }}</em>" ({{ tookMs }}ms)
          </span>
          <span v-else>Searching catalog...</span>
        </div>

        <div v-if="results.length > 0" class="results-grid">
          <div v-for="app in results" :key="app.appId" class="app-search-card">
            <div class="app-card-icon">📱</div>
            <div class="app-card-body">
              <div class="card-title-row">
                <NuxtLink :to="'/apps/' + app.slug" class="app-title-link">
                  <h4>{{ app.title }}</h4>
                </NuxtLink>
                <span class="badge" :class="app.directDownloadAvailable ? 'badge-direct' : 'badge-external'">
                  {{ app.directDownloadAvailable ? 'Direct Install' : 'External Store' }}
                </span>
              </div>

              <div class="app-developer-text">{{ app.developerName }} • {{ app.category }}</div>
              <p class="app-desc">{{ app.shortDescription }}</p>

              <div class="card-footer-meta">
                <span class="rating font-mono">★ {{ app.bayesianRating.toFixed(1) }} ({{ app.ratingCount }})</span>
                <span class="font-mono text-muted text-xs">{{ app.packageName }}</span>
              </div>
            </div>
          </div>
        </div>

        <div v-else-if="!isLoading" class="no-results-box">
          <div class="no-results-icon">🔎</div>
          <h3>No applications found</h3>
          <p>Try searching for a different keyword or relaxing your filter criteria.</p>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';

const searchQuery = ref('Authenticator');
const activeQuery = ref('Authenticator');
const showSuggestions = ref(false);
const isLoading = ref(false);
const totalResults = ref(3);
const tookMs = ref(12);

const selectedRightsFilter = ref('ALL');
const selectedCategory = ref('');
const selectedPriceFilter = ref('ALL');

const suggestions = ref([
  { text: 'Secure Authenticator 2FA', type: 'APP', package: 'com.secure.authenticator' },
  { text: 'OTP Generator', type: 'KEYWORD' }
]);

const results = ref([
  {
    appId: 'app-1',
    slug: 'secure-authenticator-2fa',
    title: 'Secure Authenticator 2FA',
    packageName: 'com.secure.authenticator',
    developerName: 'Security Labs',
    category: 'TOOLS',
    shortDescription: 'High-security two-factor authentication generator with local encrypted vault.',
    bayesianRating: 4.8,
    ratingCount: 12000,
    directDownloadAvailable: true
  },
  {
    appId: 'app-3',
    slug: 'cloud-authenticator-pro',
    title: 'Cloud Authenticator Pro',
    packageName: 'com.cloud.authenticator',
    developerName: 'CloudCorp',
    category: 'TOOLS',
    shortDescription: 'Cloud-synced authenticator with encrypted backup.',
    bayesianRating: 4.7,
    ratingCount: 20000,
    directDownloadAvailable: false
  }
]);

function handleInput() {
  if (searchQuery.value.length >= 2) {
    showSuggestions.value = true;
  } else {
    showSuggestions.value = false;
  }
}

function selectSuggestion(item: any) {
  searchQuery.value = item.text;
  showSuggestions.value = false;
  triggerSearch();
}

function triggerSearch() {
  activeQuery.value = searchQuery.value;
  showSuggestions.value = false;
  isLoading.value = true;
  setTimeout(() => {
    isLoading.value = false;
  }, 100);
}

function clearSearch() {
  searchQuery.value = '';
  showSuggestions.value = false;
}
</script>

<style scoped>
.search-page-container {
  max-width: 1100px;
  margin: 0 auto;
  padding: 32px 16px;
  font-family: inherit;
}
.search-header-section {
  text-align: center;
  margin-bottom: 32px;
}
.search-header-section h1 {
  font-size: 28px;
  margin: 0 0 8px 0;
}
.subtitle {
  color: #64748b;
  font-size: 14px;
  margin: 0 0 24px 0;
}
.search-box-wrapper {
  position: relative;
  max-width: 640px;
  margin: 0 auto;
}
.input-with-icon {
  display: flex;
  align-items: center;
  border: 2px solid #cbd5e1;
  border-radius: 30px;
  padding: 4px 16px;
  background: white;
  box-shadow: 0 2px 6px rgba(0,0,0,0.05);
}
.search-icon {
  font-size: 18px;
  color: #94a3b8;
  margin-right: 8px;
}
.main-search-input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 16px;
  padding: 10px 4px;
}
.clear-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 16px;
  cursor: pointer;
}
.suggestions-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: white;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  margin-top: 6px;
  box-shadow: 0 8px 16px rgba(0,0,0,0.1);
  z-index: 100;
  text-align: left;
}
.suggestion-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  cursor: pointer;
  border-bottom: 1px solid #f1f5f9;
}
.suggestion-row:hover {
  background: #f8fafc;
}
.suggestion-type-badge {
  font-size: 10px;
  background: #f1f5f9;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 700;
}
.suggestion-text {
  flex: 1;
  font-size: 14px;
  color: #0f172a;
}
.suggestion-pkg {
  font-size: 11px;
  color: #94a3b8;
}
.search-layout {
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: 28px;
}
.facets-sidebar {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 20px;
  height: fit-content;
}
.facets-sidebar h3 {
  margin: 0 0 16px 0;
  font-size: 15px;
}
.facet-group {
  margin-bottom: 20px;
}
.facet-group label {
  font-size: 11px;
  text-transform: uppercase;
  color: #64748b;
  font-weight: 600;
  display: block;
  margin-bottom: 8px;
}
.filter-options {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.radio-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  cursor: pointer;
  text-transform: none;
  font-weight: normal;
}
.facet-select {
  width: 100%;
  padding: 8px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 13px;
}
.results-main {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.results-summary {
  font-size: 13px;
  color: #64748b;
}
.results-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.app-search-card {
  display: flex;
  gap: 16px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
  transition: all 0.15s;
}
.app-search-card:hover {
  border-color: #3b82f6;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}
.app-card-icon {
  font-size: 40px;
  background: #f8fafc;
  padding: 12px;
  border-radius: 8px;
  height: fit-content;
}
.app-card-body {
  flex: 1;
}
.card-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}
.app-title-link {
  text-decoration: none;
  color: #0f172a;
}
.app-title-link h4 {
  margin: 0;
  font-size: 16px;
}
.app-developer-text {
  font-size: 12px;
  color: #64748b;
  margin-bottom: 6px;
}
.app-desc {
  font-size: 13px;
  color: #334155;
  margin: 0 0 10px 0;
  line-height: 1.4;
}
.card-footer-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
}
.rating {
  color: #d97706;
  font-weight: 600;
}
.badge {
  padding: 3px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
}
.badge-direct { background: #dcfce7; color: #166534; }
.badge-external { background: #f1f5f9; color: #475569; }
.no-results-box {
  text-align: center;
  padding: 60px 20px;
  background: white;
  border: 1px dashed #cbd5e1;
  border-radius: 8px;
}
.no-results-icon { font-size: 40px; margin-bottom: 12px; }
.text-muted { color: #94a3b8; }
.text-xs { font-size: 11px; }
.font-mono { font-family: monospace; }
</style>

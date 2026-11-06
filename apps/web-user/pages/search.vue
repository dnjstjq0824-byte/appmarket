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
            <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
      </aside>

      <!-- Search Results Area -->
      <main class="results-main">
        <div class="results-summary">
          <span v-if="!isLoading && !loadError">
            Found <strong>{{ totalResults }}</strong> results for "<em>{{ activeQuery }}</em>" ({{ tookMs }}ms)
            <span v-if="fallbackUsed" class="fallback-note">· fallback index</span>
          </span>
          <span v-else-if="loadError" class="search-error">{{ loadError }}</span>
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
import { ref, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { getSharedClient, ApiClientError, type SearchDoc, type SuggestionItem } from '@bookingnetwork/api-client';

const api = getSharedClient();
const route = useRoute();
const router = useRouter();

const searchQuery = ref('');
const activeQuery = ref('');
const showSuggestions = ref(false);
const isLoading = ref(false);
const totalResults = ref(0);
const tookMs = ref(0);
const loadError = ref('');
const fallbackUsed = ref(false);

const selectedRightsFilter = ref('ALL');
const selectedCategory = ref('');

interface AppResultItem {
  appId: string;
  slug: string;
  title: string;
  packageName: string;
  developerName: string;
  category: string;
  shortDescription: string;
  bayesianRating: number;
  ratingCount: number;
  directDownloadAvailable: boolean;
}

const suggestions = ref<SuggestionItem[]>([]);
const results = ref<AppResultItem[]>([]);
const categories = ref<Array<{ id: string; name: string }>>([]);
let suggestTimer: ReturnType<typeof setTimeout> | null = null;

function titleOf(doc: SearchDoc): string {
  return doc.title_i18n['ko-KR'] || doc.title_i18n['en-US'] || Object.values(doc.title_i18n)[0] || doc.package_name;
}

function descOf(doc: SearchDoc): string {
  return doc.description_i18n['ko-KR'] || doc.description_i18n['en-US'] || Object.values(doc.description_i18n)[0] || '';
}

async function performRealSearch() {
  const query = searchQuery.value.trim();
  activeQuery.value = query;
  showSuggestions.value = false;
  isLoading.value = true;
  loadError.value = '';

  try {
    const extra: Record<string, string> = {};
    if (selectedCategory.value) extra.category = selectedCategory.value;
    if (selectedRightsFilter.value === 'DIRECT_ONLY') extra.direct_only = 'true';
    const res = await api.searchApps(query, 20, extra);
    results.value = (res.items || []).map((d) => ({
      appId: d.app_id,
      slug: d.app_id,
      title: titleOf(d),
      packageName: d.package_name,
      developerName: d.developer_name,
      category: d.category,
      shortDescription: descOf(d),
      bayesianRating: d.bayesian_rating,
      ratingCount: d.rating_count,
      directDownloadAvailable: d.direct_download_available
    }));
    totalResults.value = res.total_count;
    tookMs.value = res.took_ms;
    fallbackUsed.value = res.fallback_used;
    void router.replace({ query: { ...route.query, q: query } }).catch(() => {});
  } catch (err) {
    loadError.value = err instanceof ApiClientError ? err.message : 'Search failed';
    results.value = [];
    totalResults.value = 0;
  } finally {
    isLoading.value = false;
  }
}

function handleInput() {
  if (suggestTimer) clearTimeout(suggestTimer);
  if (searchQuery.value.trim().length < 2) {
    showSuggestions.value = false;
    suggestions.value = [];
    return;
  }
  suggestTimer = setTimeout(async () => {
    try {
      const res = await api.suggestApps(searchQuery.value.trim(), 8);
      suggestions.value = res.suggestions || [];
      showSuggestions.value = suggestions.value.length > 0;
    } catch {
      suggestions.value = [];
      showSuggestions.value = false;
    }
  }, 250);
}

function selectSuggestion(item: SuggestionItem) {
  searchQuery.value = item.text;
  showSuggestions.value = false;
  triggerSearch();
}

function triggerSearch() {
  performRealSearch();
}

function clearSearch() {
  searchQuery.value = '';
  showSuggestions.value = false;
  performRealSearch();
}

onMounted(async () => {
  const initialQ = (route.query.q as string) || '';
  if (initialQ) {
    searchQuery.value = initialQ;
  }
  try {
    const cats = await api.getCategories();
    categories.value = cats.map(c => ({ id: c.id, name: c.name_i18n['ko-KR'] || c.name_i18n['en-US'] || c.id }));
  } catch {
    categories.value = [];
  }
  performRealSearch();
});

watch(() => route.query.q, (newQ) => {
  if (typeof newQ === 'string' && newQ !== searchQuery.value) {
    searchQuery.value = newQ;
    performRealSearch();
  }
});
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
.fallback-note { color: #b45309; font-size: 12px; }
.search-error { color: #dc2626; }
</style>


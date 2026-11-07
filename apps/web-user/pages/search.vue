<template>
  <div class="search-page-container">
    <div class="search-header-section">
      <h1>Catalog & Multi-Store Search</h1>
      <p class="subtitle">Unified search across native hosted releases and verified external open-source repositories</p>

      <!-- Search Input & Autocomplete Dropdown (USR-03) -->
      <div class="search-box-wrapper">
        <div class="input-with-icon">
          <svg class="search-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="11" cy="11" r="7"></circle><line x1="16.5" y1="16.5" x2="21" y2="21"></line></svg>
          <input 
            type="text" 
            v-model="searchQuery" 
            @input="handleInput"
            @keydown.enter="triggerSearch"
            placeholder="Search by app title, developer, or package name..."
            class="main-search-input"
            id="mainSearchInput"
          />
          <button v-if="searchQuery" @click="clearSearch" class="clear-btn" aria-label="검색어 지우기"><svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"><path d="M1 1l10 10M11 1L1 11" /></svg></button>
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
            <div class="app-card-icon">{{ app.title.slice(0, 1).toUpperCase() }}</div>
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
          <div class="no-results-icon">0 RESULTS</div>
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
  padding: 32px 24px 60px;
}
.search-header-section {
  margin-bottom: 28px;
  padding-bottom: 20px;
  border-bottom: 2px solid var(--bn-ink);
}
.search-header-section h1 {
  font-family: var(--bn-font-display);
  font-size: 30px;
  font-weight: 900;
  letter-spacing: -0.025em;
  margin: 0 0 6px 0;
}
.subtitle {
  color: var(--bn-text-secondary);
  font-size: 13px;
  margin: 0 0 20px 0;
}
.search-box-wrapper {
  position: relative;
  max-width: 640px;
}
.input-with-icon {
  display: flex;
  align-items: center;
  gap: 10px;
  border: 1px solid var(--bn-border-strong);
  border-radius: var(--bn-radius-sm);
  padding: 4px 14px;
  background: var(--bn-bg-surface);
}
.input-with-icon:focus-within {
  border-color: var(--bn-ink);
}
.search-icon {
  color: var(--bn-text-muted);
  flex-shrink: 0;
}
.main-search-input {
  flex: 1;
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
  font-size: 15px;
  padding: 10px 4px;
}
.clear-btn {
  background: transparent;
  border: none;
  color: var(--bn-text-muted);
  cursor: pointer;
  display: flex;
  padding: 6px;
}
.clear-btn:hover {
  color: var(--bn-text-primary);
}
.suggestions-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  margin-top: 6px;
  box-shadow: var(--bn-shadow-pop);
  z-index: 100;
  overflow: hidden;
}
.suggestion-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  cursor: pointer;
  border-bottom: 1px solid var(--bn-border-subtle);
}
.suggestion-row:last-child {
  border-bottom: none;
}
.suggestion-row:hover {
  background: var(--bn-bg-well);
}
.suggestion-type-badge {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  border: 1px solid var(--bn-border-subtle);
  padding: 2px 6px;
  border-radius: var(--bn-radius-sm);
  color: var(--bn-text-muted);
}
.suggestion-text {
  flex: 1;
  font-size: 14px;
  color: var(--bn-text-primary);
}
.suggestion-pkg {
  font-size: 11px;
  color: var(--bn-text-muted);
}
.search-layout {
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: 28px;
}
.facets-sidebar {
  background: var(--bn-bg-surface);
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  padding: 20px;
  height: fit-content;
}
.facets-sidebar h3 {
  font-family: var(--bn-font-display);
  margin: 0 0 16px 0;
  font-size: 14px;
  font-weight: 800;
}
.facet-group {
  margin-bottom: 20px;
}
.facet-group > label {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.1em;
  color: var(--bn-text-muted);
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
  color: var(--bn-text-secondary);
}
.radio-label input {
  accent-color: var(--bn-seal);
}
.facet-select {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-sm);
  background: var(--bn-bg-surface);
  color: var(--bn-text-primary);
  font-family: var(--bn-font-sans);
  font-size: 13px;
}
.results-main {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}
.results-summary {
  font-size: 13px;
  color: var(--bn-text-secondary);
}
.results-summary strong {
  font-family: var(--bn-font-display);
  color: var(--bn-text-primary);
}
.results-grid {
  display: flex;
  flex-direction: column;
  gap: 0;
  border: 1px solid var(--bn-border-subtle);
  border-radius: var(--bn-radius-md);
  overflow: hidden;
}
.app-search-card {
  display: flex;
  gap: 16px;
  background: var(--bn-bg-surface);
  border-bottom: 1px solid var(--bn-border-subtle);
  padding: 16px;
}
.app-search-card:last-child {
  border-bottom: none;
}
.app-search-card:hover {
  background: var(--bn-bg-well);
}
.app-card-icon {
  width: 52px;
  height: 52px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bn-ink);
  color: var(--bn-bg-canvas);
  font-family: var(--bn-font-display);
  font-weight: 900;
  font-size: 22px;
  border-radius: var(--bn-radius-md);
}
.app-card-body {
  flex: 1;
  min-width: 0;
}
.card-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 4px;
}
.app-title-link {
  text-decoration: none;
  color: var(--bn-text-primary);
  min-width: 0;
}
.app-title-link h4 {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
}
.app-title-link:hover h4 {
  text-decoration: underline;
  text-underline-offset: 3px;
}
.app-developer-text {
  font-size: 12px;
  color: var(--bn-text-secondary);
  margin-bottom: 6px;
}
.app-desc {
  font-size: 13px;
  color: var(--bn-text-secondary);
  margin: 0 0 10px 0;
  line-height: 1.55;
}
.card-footer-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  font-size: 12px;
}
.rating {
  font-family: var(--bn-font-mono);
  color: var(--bn-text-primary);
  font-weight: 600;
}
.badge {
  font-family: var(--bn-font-mono);
  font-size: 10px;
  letter-spacing: 0.06em;
  padding: 3px 9px;
  border-radius: var(--bn-radius-sm);
  border: 1px solid var(--bn-border-strong);
  color: var(--bn-text-secondary);
  white-space: nowrap;
}
.badge-direct {
  border-color: var(--bn-trust-verified);
  color: var(--bn-trust-verified);
}
.badge-external {
  border-color: var(--bn-border-strong);
  color: var(--bn-text-secondary);
}
.no-results-box {
  text-align: center;
  padding: 60px 20px;
  background: var(--bn-bg-surface);
  border: 1px dashed var(--bn-border-strong);
  border-radius: var(--bn-radius-md);
}
.no-results-box h3 {
  font-family: var(--bn-font-display);
  font-size: 17px;
  margin: 0 0 6px;
}
.no-results-box p {
  font-size: 13px;
  color: var(--bn-text-secondary);
  margin: 0;
}
.no-results-icon {
  font-family: var(--bn-font-mono);
  font-size: 13px;
  letter-spacing: 0.14em;
  color: var(--bn-text-muted);
  margin-bottom: 14px;
}
.text-muted { color: var(--bn-text-muted); }
.text-xs { font-size: 11px; }
.font-mono { font-family: var(--bn-font-mono); }
.fallback-note { color: var(--bn-warning-amber); font-size: 12px; }
.search-error { color: var(--bn-danger-red); }

@media (max-width: 860px) {
  .search-layout {
    grid-template-columns: 1fr;
  }
}
</style>


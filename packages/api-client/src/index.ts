export interface ProblemDetails {
  type: string
  title: string
  status: number
  code: string
  detail: string
  trace_id: string
  timestamp: string
}

export class ApiClientError extends Error {
  public problem?: ProblemDetails
  public status: number

  constructor(status: number, message: string, problem?: ProblemDetails) {
    super(message)
    this.name = 'ApiClientError'
    this.status = status
    this.problem = problem
  }
}

export interface ClientConfig {
  baseUrl: string
  getAccessToken?: () => string | null
  onTokenExpired?: () => Promise<string | null>
}

// ---------------------------------------------------------------------------
// Token store (localStorage-backed, SSR-safe).
// ---------------------------------------------------------------------------

const ACCESS_TOKEN_KEY = 'bn_access_token'
const REFRESH_TOKEN_KEY = 'bn_refresh_token'

function storage(): Storage | null {
  if (typeof window === 'undefined' || !window.localStorage) {
    return null
  }
  return window.localStorage
}

export function getStoredAccessToken(): string | null {
  return storage()?.getItem(ACCESS_TOKEN_KEY) ?? null
}

export function getStoredRefreshToken(): string | null {
  return storage()?.getItem(REFRESH_TOKEN_KEY) ?? null
}

export function storeSession(accessToken: string, refreshToken?: string): void {
  const s = storage()
  if (!s) {
    return
  }
  s.setItem(ACCESS_TOKEN_KEY, accessToken)
  if (refreshToken) {
    s.setItem(REFRESH_TOKEN_KEY, refreshToken)
  }
}

export function clearSession(): void {
  const s = storage()
  if (!s) {
    return
  }
  s.removeItem(ACCESS_TOKEN_KEY)
  s.removeItem(REFRESH_TOKEN_KEY)
}

export function isAuthenticated(): boolean {
  return !!getStoredAccessToken()
}

// ---------------------------------------------------------------------------
// Shared DTOs (mirror backend JSON shapes).
// ---------------------------------------------------------------------------

export interface AuthResult {
  access_token?: string
  refresh_token?: string
  expires_in?: number
  token_type?: string
  user_id?: string
  mfa_required?: boolean
  mfa_enrollment_required?: boolean
  user?: { id: string; email: string }
}

export interface MFAEnrollment {
  user_id: string
  secret: string
  provisioning_uri: string
  recovery_codes: string[]
}

export interface CatalogAppItem {
  id: string
  canonical_slug: string
  title: string
  primary_package_name: string
  primary_category_id: string
  average_rating: number
  bayesian_rating: number
  install_count: number
  icon_url?: string
}

export interface CatalogListResponse {
  items: CatalogAppItem[]
  cursor?: string
}

export interface CatalogApp {
  id: string
  canonical_slug: string
  primary_package_name: string
  developer_org_id?: string
  primary_category_id: string
  content_rating: string
  default_locale: string
  catalog_status: string
  average_rating: number
  bayesian_rating: number
  install_count: number
  source_code_url?: string
  license_spdx: string
  monetization_type: string
  created_at: string
  updated_at: string
}

export interface AppLocalization {
  id: string
  app_id: string
  locale: string
  title: string
  short_description: string
  full_description: string
  whats_new: string
  privacy_policy_url: string
}

export interface AppMediaAsset {
  id: string
  app_id: string
  locale: string
  device_form_factor: string
  asset_type: string
  url: string
  width: number
  height: number
  display_order: number
}

export interface AppDistribution {
  id: string
  app_id: string
  source_type: string
  distribution_mode: string
  package_name: string
  signing_cert_sha256: string
  signature_compatibility_group: string
  official_store_url: string
  last_synced_at?: string
  latest_release?: LatestRelease
}

export interface AppTag {
  id: string
  name_i18n: Record<string, string>
}

export interface AppDetailDTO {
  app: CatalogApp
  localization: AppLocalization | null
  media_assets: AppMediaAsset[]
  distributions: AppDistribution[]
  tags: AppTag[]
  resolved_locale: string
}

export interface CategoryDTO {
  id: string
  parent_id?: string
  name_i18n: Record<string, string>
  app_count: number
}

export interface LatestRelease {
  id: string
  distribution_id: string
  app_id: string
  release_track: string
  version_code: number
  version_name: string
  min_sdk: number
  target_sdk: number
  supported_abis: string[]
  file_size_bytes: number
  release_notes: string
  published_at?: string
}

export interface OrgSummary {
  id: string
  org_slug: string
  org_name: string
  verification_status: string
  mfa_required?: boolean
  created_at?: string
}

export interface ReleaseDTO {
  id: string
  versionName: string
  versionCode: number
  track: string
  status: string
  rolloutPercentage: number
  isHalted?: boolean
}

export interface UploadSessionDTO {
  uploadId: string
  uploadUrl?: string
  expiresAt?: string
}

export interface OrderDTO {
  id: string
  user_id: string
  app_id: string
  org_id?: string
  currency: string
  gross_amount: number
  tax_amount: number
  commission_amount: number
  dev_payout_amount: number
  status: string
  created_at: string
}

export interface DualApprovalDTO {
  id: string
  operation_type: string
  initiator_id: string
  approver_id?: string
  target_resource_id: string
  payload_json: string
  status: string
  reason: string
  created_at: string
  decided_at?: string
}

export interface OverviewEntry {
  key: string
  count: number
}

export interface RevenueEntry {
  key: string
  count: number
  revenueCents: number
}

export interface AnalyticsOverview {
  totalEvents: number
  last24h: number
  totalRevenueCents: number
  byEvent: OverviewEntry[]
  byPlatform: OverviewEntry[]
  byCountry: OverviewEntry[]
  byApp: OverviewEntry[]
  revenueByCountry: RevenueEntry[]
}

export interface SearchDoc {
  app_id: string
  package_name: string
  title_i18n: Record<string, string>
  description_i18n: Record<string, string>
  developer_name: string
  category: string
  bayesian_rating: number
  rating_count: number
  download_count: number
  direct_download_available: boolean
  rights_tier: string
  published_at: string
  keywords: string[]
  score?: number
}

export interface SearchResult {
  items: SearchDoc[]
  total_count: number
  took_ms: number
  fallback_used: boolean
}

export interface SuggestionItem {
  text: string
  type: string
  package?: string
}

export interface SuggestResult {
  suggestions: SuggestionItem[]
  took_ms: number
}

export interface UserReview {
  id: string
  app_id: string
  user_id: string
  rating: number
  title: string
  comment: string
  version_code: number
  is_verified_install: boolean
  fraud_score: number
  status: string
  created_at: string
}

export interface UserEntitlement {
  id: string
  user_id: string
  app_id: string
  order_id: string
  status: string
  granted_at: string
}

export interface PromoRedemption {
  promotion: {
    id: string
    app_id: string
    campaign_name: string
    discount_type: string
    discount_value: number
    starts_at: string
    ends_at: string
    max_uses: number
    used_count: number
  }
  code: string
}

export interface PriceTier {
  app_id: string
  country_code: string
  currency: string
  list_price_minor: number
  tax_inclusive: boolean
  updated_at: string
}

export interface SignedDownloadURL {
  url: string
  expires_at: string
  sha256_checksum: string
}

export interface AuditLogRecordDTO {
  id: string
  actor_id: string
  action: string
  target_type: string
  target_id: string
  diff_json?: string
  prev_hash: string
  record_hash: string
  created_at: string
  created_at_nano?: string
}

export interface CheckoutSession {
  provider: string
  order_id: string
  amount_krw: number
  checkout_url: string
  expires_at: string
}

// ---------------------------------------------------------------------------
// Client.
// ---------------------------------------------------------------------------

export class ApiClient {
  private baseUrl: string
  private getAccessToken?: () => string | null
  private onTokenExpired?: () => Promise<string | null>

  constructor(config: ClientConfig) {
    this.baseUrl = config.baseUrl.replace(/\/$/, '')
    this.getAccessToken = config.getAccessToken
    this.onTokenExpired = config.onTokenExpired
  }

  async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseUrl}${path.startsWith('/') ? path : '/' + path}`
    const headers = new Headers(options.headers || {})

    if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
      headers.set('Content-Type', 'application/json')
    }

    const token = this.getAccessToken ? this.getAccessToken() : null
    if (token && !headers.has('Authorization')) {
      headers.set('Authorization', `Bearer ${token}`)
    }

    let response = await fetch(url, { ...options, headers })

    if (response.status === 401 && this.onTokenExpired) {
      const newToken = await this.onTokenExpired()
      if (newToken) {
        headers.set('Authorization', `Bearer ${newToken}`)
        response = await fetch(url, { ...options, headers })
      }
    }

    if (!response.ok) {
      let problem: ProblemDetails | undefined
      try {
        problem = await response.json()
      } catch {
        // Not a JSON problem response
      }
      throw new ApiClientError(
        response.status,
        problem?.detail || `API Request failed with status ${response.status}`,
        problem
      )
    }

    if (response.status === 204) {
      return {} as T
    }

    return response.json() as Promise<T>
  }

  private post<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>(path, {
      method: 'POST',
      body: body === undefined ? undefined : JSON.stringify(body)
    })
  }

  private patch<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>(path, { method: 'PATCH', body: JSON.stringify(body) })
  }

  private put<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>(path, { method: 'PUT', body: JSON.stringify(body) })
  }

  private del<T = Record<string, string>>(path: string): Promise<T> {
    return this.request<T>(path, { method: 'DELETE' })
  }

  // -- Auth ---------------------------------------------------------------
  async register(email: string, password: string, displayName: string): Promise<AuthResult> {
    return this.post<AuthResult>('/api/v1/auth/register', {
      email,
      password,
      display_name: displayName
    })
  }

  async login(email: string, password: string): Promise<AuthResult> {
    return this.post<AuthResult>('/api/v1/auth/login', { email, password })
  }

  async refreshToken(refreshToken: string): Promise<AuthResult> {
    return this.post<AuthResult>('/api/v1/auth/token/refresh', { refresh_token: refreshToken })
  }

  async verifyMfa(userId: string, code: string): Promise<AuthResult> {
    return this.post<AuthResult>('/api/v1/auth/mfa/verify', { user_id: userId, mfa_code: code })
  }

  async enrollMfa(email: string, password: string): Promise<MFAEnrollment> {
    return this.post<MFAEnrollment>('/api/v1/auth/mfa/enroll', { email, password })
  }

  async getMe(): Promise<{ sub: string; email: string; roles: string[]; org_roles: Record<string, string>; mfa_verified: boolean }> {
    return this.request('/api/v1/auth/users/me')
  }

  async listOAuthProviders(): Promise<Array<{ id: string; name: string }>> {
    return this.request('/api/v1/auth/oauth/providers')
  }

  oauthLoginUrl(provider: string): string {
    return `${this.baseUrl}/api/v1/auth/oauth/${provider}/login`
  }

  async listApprovals(): Promise<DualApprovalDTO[]> {
    return this.request<DualApprovalDTO[]>('/api/v1/auth/admin/approvals/pending')
  }

  async createApproval(payload: { operation_type: string; target_resource_id?: string; payload_json?: string; reason: string }): Promise<DualApprovalDTO> {
    return this.post<DualApprovalDTO>('/api/v1/auth/admin/approvals', payload)
  }

  async approveDualControl(requestId: string) {
    return this.post(`/api/v1/auth/admin/approvals/${requestId}/approve`)
  }

  async rejectDualControl(requestId: string, reason?: string) {
    return this.post(`/api/v1/auth/admin/approvals/${requestId}/reject`, { reason: reason || '' })
  }

  // -- Catalog ------------------------------------------------------------
  async getCatalogApps(params?: Record<string, string>): Promise<CatalogListResponse> {
    const qs = params ? '?' + new URLSearchParams(params).toString() : ''
    return this.request<CatalogListResponse>(`/api/v1/catalog/apps${qs}`)
  }

  async getAppBySlug(slug: string, locale?: string): Promise<AppDetailDTO> {
    const qs = locale ? `?locale=${encodeURIComponent(locale)}` : ''
    return this.request<AppDetailDTO>(`/api/v1/catalog/apps/${slug}${qs}`)
  }

  async getLatestRelease(slug: string): Promise<LatestRelease> {
    return this.request<LatestRelease>(`/api/v1/catalog/apps/${slug}/releases/latest`)
  }

  async getCategories(): Promise<CategoryDTO[]> {
    const res = await this.request<{ items: CategoryDTO[] }>('/api/v1/catalog/categories')
    return res.items || []
  }

  async getChart(chartType: string, limit = 20): Promise<CatalogAppItem[]> {
    const res = await this.request<{ chart_type: string; items: CatalogAppItem[] }>(`/api/v1/catalog/charts/${chartType}?limit=${limit}`)
    return res.items || []
  }

  // -- Distribution -------------------------------------------------------
  async createDownloadToken(appId: string, releaseId: string, deviceId: string, sha256?: string): Promise<SignedDownloadURL> {
    return this.post<SignedDownloadURL>(`/api/v1/distribution/apps/${appId}/releases/${releaseId}/download-token`, {
      device_or_user_id: deviceId,
      sha256_checksum: sha256 || ''
    })
  }

  async checkUpdates(device: Record<string, unknown>, packages: Array<{ package_name: string; version_code: number; cert_sha256?: string; installer_package_name?: string }>): Promise<unknown[]> {
    return this.post('/api/v1/distribution/device/check-updates', { device, packages })
  }

  // -- Developer console --------------------------------------------------
  async listOrgs(): Promise<OrgSummary[]> {
    return this.request('/api/v1/developer/orgs')
  }

  async createOrganization(slug: string, name: string): Promise<OrgSummary> {
    return this.post<OrgSummary>('/api/v1/developer/orgs', { slug, name })
  }

  async getVerification(orgId: string): Promise<unknown> {
    return this.request(`/api/v1/developer/orgs/${orgId}/verification`)
  }

  async submitVerification(orgId: string, payload: { doc_type: string; doc_url: string; domain: string }): Promise<unknown> {
    return this.post(`/api/v1/developer/orgs/${orgId}/verification`, payload)
  }

  async listMembers(orgId: string): Promise<unknown[]> {
    return this.request(`/api/v1/developer/orgs/${orgId}/members`)
  }

  async addMember(orgId: string, userId: string, role: string): Promise<unknown> {
    return this.post(`/api/v1/developer/orgs/${orgId}/members`, { user_id: userId, role })
  }

  async removeMember(orgId: string, userId: string): Promise<unknown> {
    return this.del(`/api/v1/developer/orgs/${orgId}/members/${encodeURIComponent(userId)}`)
  }

  async listApiKeys(orgId: string): Promise<unknown[]> {
    return this.request(`/api/v1/developer/orgs/${orgId}/api-keys`)
  }

  async createApiKey(orgId: string, payload: { name: string; scopes: string[]; ttl_days: number }): Promise<unknown> {
    return this.post(`/api/v1/developer/orgs/${orgId}/api-keys`, payload)
  }

  async revokeApiKey(orgId: string, keyId: string): Promise<unknown> {
    return this.del(`/api/v1/developer/orgs/${orgId}/api-keys/${keyId}`)
  }

  async listOrgApps(orgId: string): Promise<unknown[]> {
    return this.request(`/api/v1/developer/orgs/${orgId}/apps`)
  }

  async getOrgSettlements(orgId: string): Promise<unknown> {
    return this.request(`/api/v1/developer/orgs/${orgId}/settlements`)
  }

  async getPayoutAccount(orgId: string): Promise<unknown> {
    return this.request(`/api/v1/developer/orgs/${orgId}/payout-account`)
  }

  async savePayoutAccount(orgId: string, payload: { currency: string; bank_name: string; holder_name: string; account_number: string; routing_code?: string }): Promise<unknown> {
    return this.put(`/api/v1/developer/orgs/${orgId}/payout-account`, payload)
  }

  async listLocalizations(appId: string, orgId?: string): Promise<unknown[]> {
    const qs = orgId ? `?org=${encodeURIComponent(orgId)}` : ''
    return this.request(`/api/v1/developer/apps/${appId}/metadata${qs}`)
  }

  async submitForReview(appId: string, orgId?: string): Promise<unknown> {
    const qs = orgId ? `?org=${encodeURIComponent(orgId)}` : ''
    return this.post(`/api/v1/developer/apps/${appId}/submit${qs}`)
  }

  async saveLocalization(appId: string, payload: Record<string, string>): Promise<unknown> {
    return this.put(`/api/v1/developer/apps/${appId}/metadata`, payload)
  }

  async listMedia(appId: string, locale?: string, orgId?: string): Promise<unknown[]> {
    const params = new URLSearchParams()
    if (locale) {
      params.set('locale', locale)
    }
    if (orgId) {
      params.set('org', orgId)
    }
    const qs = params.toString() ? `?${params.toString()}` : ''
    return this.request(`/api/v1/developer/apps/${appId}/media${qs}`)
  }

  async uploadMedia(appId: string, image: File, fields: { org?: string; locale: string; formFactor: string; assetType: string }): Promise<unknown> {
    const form = new FormData()
    form.set('image', image, image.name)
    if (fields.org) {
      form.set('org', fields.org)
    }
    form.set('locale', fields.locale)
    form.set('formFactor', fields.formFactor)
    form.set('assetType', fields.assetType)
    return this.request(`/api/v1/developer/apps/${appId}/media`, { method: 'POST', body: form })
  }

  mediaContentUrl(appId: string, assetId: string, orgId?: string): string {
    const qs = orgId ? `?org=${encodeURIComponent(orgId)}` : ''
    return `/api/v1/developer/apps/${appId}/media/${assetId}/content${qs}`
  }

  async deleteMedia(appId: string, assetId: string, orgId?: string): Promise<unknown> {
    const qs = orgId ? `?org=${encodeURIComponent(orgId)}` : ''
    return this.del(`/api/v1/developer/apps/${appId}/media/${assetId}${qs}`)
  }

  async initUpload(appId: string, payload: { orgId: string; fileName: string; fileSize: number; expectedSha256: string }): Promise<UploadSessionDTO> {
    return this.post<UploadSessionDTO>(`/api/v1/developer/apps/${appId}/uploads`, payload)
  }

  async completeUpload(appId: string, uploadId: string, sha256: string, apk?: File): Promise<unknown> {
    if (apk) {
      const form = new FormData()
      form.set('sha256', sha256)
      form.set('apk', apk, apk.name)
      return this.request(`/api/v1/developer/apps/${appId}/uploads/${uploadId}/complete`, {
        method: 'POST',
        body: form
      })
    }
    return this.post(`/api/v1/developer/apps/${appId}/uploads/${uploadId}/complete`, { sha256 })
  }

  async listReleases(appId: string, orgId?: string): Promise<unknown[]> {
    const qs = orgId ? `?org=${encodeURIComponent(orgId)}` : ''
    return this.request(`/api/v1/developer/apps/${appId}/releases${qs}`)
  }

  async getCurrentRelease(appId: string, orgId?: string): Promise<ReleaseDTO> {
    const qs = orgId ? `?org=${encodeURIComponent(orgId)}` : ''
    return this.request<ReleaseDTO>(`/api/v1/developer/apps/${appId}/releases/current${qs}`)
  }

  async updateRollout(appId: string, payload: { rolloutPercentage?: number; action?: string; reason?: string }, orgId?: string): Promise<ReleaseDTO> {
    const qs = orgId ? `?org=${encodeURIComponent(orgId)}` : ''
    return this.patch<ReleaseDTO>(`/api/v1/developer/apps/${appId}/releases/current/rollout${qs}`, payload)
  }

  async getLineage(appId: string, orgId?: string): Promise<unknown> {
    const qs = orgId ? `?org=${encodeURIComponent(orgId)}` : ''
    return this.request(`/api/v1/developer/apps/${appId}/lineage${qs}`)
  }

  async listCampaigns(appId: string, orgId?: string): Promise<unknown[]> {
    const qs = orgId ? `?org=${encodeURIComponent(orgId)}` : ''
    return this.request(`/api/v1/developer/apps/${appId}/ads/campaigns${qs}`)
  }

  async createCampaign(appId: string, payload: { orgId: string; appName: string; targetKeywords: string[]; maxBidCpc: number; dailyBudget: number }): Promise<unknown> {
    return this.post(`/api/v1/developer/apps/${appId}/ads/campaigns`, payload)
  }

  async setCampaignPaused(appId: string, campaignId: string, paused: boolean, orgId?: string): Promise<unknown> {
    const qs = orgId ? `?org=${encodeURIComponent(orgId)}` : ''
    return this.post(`/api/v1/developer/apps/${appId}/ads/campaigns/${campaignId}/pause${qs}`, { paused })
  }

  async listPromotions(appId: string, orgId?: string): Promise<unknown[]> {
    const qs = orgId ? `?org=${encodeURIComponent(orgId)}` : ''
    return this.request(`/api/v1/developer/apps/${appId}/promotions${qs}`)
  }

  async createPromotion(appId: string, payload: Record<string, unknown>): Promise<{ promotion: unknown; codes: string[] }> {
    return this.post(`/api/v1/developer/apps/${appId}/promotions`, payload)
  }

  async listPriceTiers(appId: string, orgId?: string): Promise<unknown[]> {
    const qs = orgId ? `?org=${encodeURIComponent(orgId)}` : ''
    return this.request(`/api/v1/developer/apps/${appId}/pricing${qs}`)
  }

  async setPriceTier(appId: string, payload: { orgId: string; countryCode: string; currency: string; listPriceMinor: number; taxInclusive: boolean }): Promise<unknown> {
    return this.post(`/api/v1/developer/apps/${appId}/pricing`, payload)
  }

  // -- Commerce -----------------------------------------------------------
  async getAppPrices(appId: string): Promise<PriceTier[]> {
    const res = await this.request<{ items: PriceTier[] }>(`/api/v1/commerce/apps/${appId}/prices`)
    return res.items || []
  }

  async createOrder(payload: { app_id: string; org_id?: string; gross_amount: number; currency?: string; idempotency_key?: string; promo_code?: string }): Promise<OrderDTO> {
    return this.post<OrderDTO>('/api/v1/commerce/orders', payload)
  }

  async listMyOrders(): Promise<OrderDTO[]> {
    return this.request('/api/v1/commerce/orders')
  }

  async checkoutOrder(orderId: string): Promise<CheckoutSession> {
    return this.post<CheckoutSession>(`/api/v1/commerce/orders/${orderId}/checkout`)
  }

  async refundOrder(orderId: string, reason?: string): Promise<OrderDTO> {
    return this.post<OrderDTO>(`/api/v1/commerce/orders/${orderId}/refund`, { reason: reason || '' })
  }

  async listEntitlements(): Promise<UserEntitlement[]> {
    return this.request<UserEntitlement[]>('/api/v1/commerce/entitlements')
  }

  async listInvoices(): Promise<unknown[]> {
    return this.request('/api/v1/commerce/invoices')
  }

  async listSubscriptions(): Promise<unknown[]> {
    return this.request('/api/v1/commerce/subscriptions')
  }

  async createSubscription(payload: { app_id: string; org_id?: string; billing_interval: string; amount_minor: number; currency?: string }): Promise<unknown> {
    return this.post('/api/v1/commerce/subscriptions', payload)
  }

  async redeemPromo(code: string, appId: string): Promise<PromoRedemption> {
    return this.post<PromoRedemption>('/api/v1/commerce/promos/redeem', { code, appId })
  }

  async validatePromo(code: string, appId: string): Promise<PromoRedemption> {
    return this.post<PromoRedemption>('/api/v1/commerce/promos/validate', { code, appId })
  }

  // -- Moderation ---------------------------------------------------------
  async fileReport(payload: { appIdentifier: string; category: string; reporterEmail: string; description: string; confirmAccuracy: boolean }): Promise<{ ticketId: string; priority: string; status: string }> {
    return this.post('/api/v1/moderation/reports', payload)
  }

  async listReviews(appId: string): Promise<UserReview[]> {
    return this.request<UserReview[]>(`/api/v1/moderation/apps/${appId}/reviews`)
  }

  async postReview(payload: { appId: string; rating: number; title: string; comment: string; versionCode?: number }): Promise<UserReview> {
    return this.post<UserReview>('/api/v1/moderation/reviews', payload)
  }

  async submitAppeal(payload: { orgId: string; releaseId: string; reason: string; originalReviewerId?: string }): Promise<unknown> {
    return this.post('/api/v1/moderation/appeals', payload)
  }

  // -- Search & analytics -------------------------------------------------
  async searchApps(query: string, limit = 20, extra?: Record<string, string>): Promise<SearchResult> {
    const qs = new URLSearchParams({ q: query, limit: String(limit), ...(extra || {}) }).toString()
    return this.request<SearchResult>(`/api/v1/search?${qs}`)
  }

  async suggestApps(query: string, limit = 10): Promise<SuggestResult> {
    return this.request<SuggestResult>(`/api/v1/search/suggest?q=${encodeURIComponent(query)}&limit=${limit}`)
  }

  async ingestEvents(events: Array<Record<string, unknown>>): Promise<{ accepted: number }> {
    return this.post('/api/v1/analytics/events', events)
  }

  // -- Admin --------------------------------------------------------------
  async decideReview(appId: string, decision: string, notes: string, checklist?: Record<string, boolean>): Promise<unknown> {
    return this.post(`/api/v1/admin/reviews/${appId}/decision`, { decision, notes, checklist: checklist || {} })
  }

  async listDecisions(): Promise<unknown[]> {
    return this.request('/api/v1/admin/reviews/decisions')
  }

  async getReviewQueue(): Promise<unknown[]> {
    return this.request('/api/v1/admin/apps/review-queue')
  }

  async adminGetApp(appId: string): Promise<unknown> {
    return this.request(`/api/v1/admin/apps/${appId}`)
  }

  async listVerifications(status = 'PENDING_REVIEW'): Promise<unknown[]> {
    return this.request(`/api/v1/admin/verifications?status=${encodeURIComponent(status)}`)
  }

  async decideVerification(orgId: string, approved: boolean, notes: string): Promise<unknown> {
    return this.post(`/api/v1/admin/verifications/${orgId}/decision`, { approved, notes })
  }

  async getKillSwitch(packageName: string): Promise<unknown> {
    return this.request(`/api/v1/admin/security/kill-switch/${encodeURIComponent(packageName)}`)
  }

  async listKillSwitches(): Promise<unknown[]> {
    return this.request('/api/v1/admin/security/kill-switches')
  }

  async triggerKillSwitch(packageName: string, payload: { scope?: string; reasonCode: string; evidence?: string }): Promise<unknown> {
    return this.post(`/api/v1/admin/security/kill-switch/${encodeURIComponent(packageName)}`, payload)
  }

  async clearKillSwitch(packageName: string): Promise<unknown> {
    return this.del(`/api/v1/admin/security/kill-switch/${encodeURIComponent(packageName)}`)
  }

  async listScans(): Promise<unknown[]> {
    return this.request('/api/v1/admin/security/scans')
  }

  async getScanReport(artifactId: string): Promise<unknown> {
    return this.request(`/api/v1/admin/security/scans?artifactId=${encodeURIComponent(artifactId)}`)
  }

  async listConnectors(): Promise<unknown[]> {
    return this.request('/api/v1/admin/connectors')
  }

  async registerConnector(payload: { id: string; name: string; kind: string; feedUrl?: string; fingerprint?: string; packages?: string[] }): Promise<unknown> {
    return this.post('/api/v1/admin/connectors', payload)
  }

  async triggerConnectorSync(connectorId: string): Promise<unknown> {
    return this.post(`/api/v1/admin/connectors/${connectorId}/trigger-sync`)
  }

  async listSettlements(): Promise<unknown[]> {
    return this.request('/api/v1/admin/settlements')
  }

  async createSettlement(payload: { orgId: string; amount: number; currency?: string }): Promise<unknown> {
    return this.post('/api/v1/admin/settlements', payload)
  }

  async approveSettlement(settlementId: string, approverId: string, initiatorId?: string): Promise<unknown> {
    return this.post(`/api/v1/admin/settlements/${settlementId}/approve`, {
      approver_id: approverId,
      initiator_id: initiatorId || ''
    })
  }

  async listAuditLogs(limit = 100): Promise<{ chainValid: boolean; chainFailedAt: number; records: AuditLogRecordDTO[] }> {
    return this.request(`/api/v1/admin/audit-logs?limit=${limit}`)
  }

  async listAdminOrders(params?: Record<string, string>): Promise<OrderDTO[]> {
    const qs = params ? '?' + new URLSearchParams(params).toString() : ''
    return this.request(`/api/v1/admin/orders${qs}`)
  }

  async adminRefundOrder(orderId: string, reason: string): Promise<OrderDTO> {
    return this.post(`/api/v1/admin/orders/${orderId}/refund`, { reason })
  }

  async listAdminAdCampaigns(status?: string): Promise<unknown[]> {
    const qs = status ? `?status=${encodeURIComponent(status)}` : ''
    return this.request(`/api/v1/admin/ads/campaigns${qs}`)
  }

  async reviewAdCampaign(campaignId: string, approve: boolean): Promise<unknown> {
    return this.post(`/api/v1/admin/ads/campaigns/${campaignId}/review`, { approve })
  }

  async getAnalyticsOverview(): Promise<AnalyticsOverview> {
    return this.request<AnalyticsOverview>('/api/v1/admin/analytics/overview')
  }

  async listReports(status?: string): Promise<unknown[]> {
    const qs = status ? `?status=${encodeURIComponent(status)}` : ''
    return this.request(`/api/v1/admin/reports${qs}`)
  }

  async claimReport(reportId: string): Promise<unknown> {
    return this.post(`/api/v1/admin/reports/${reportId}/claim`)
  }

  async resolveReport(reportId: string, actionTaken: boolean, notes: string): Promise<unknown> {
    return this.post(`/api/v1/admin/reports/${reportId}/resolve`, { action_taken: actionTaken, notes })
  }

  async listAppeals(): Promise<unknown[]> {
    return this.request('/api/v1/admin/appeals')
  }

  async decideAppeal(appealId: string, decision: string, notes: string): Promise<unknown> {
    return this.post(`/api/v1/admin/appeals/${appealId}/decide`, { decision, notes })
  }
}

// ---------------------------------------------------------------------------
// Default shared client (same-origin, token-aware, refresh-on-401).
// ---------------------------------------------------------------------------

let sharedClient: ApiClient | null = null

export function createDefaultClient(baseUrl = ''): ApiClient {
  return new ApiClient({
    baseUrl,
    getAccessToken: () => getStoredAccessToken(),
    onTokenExpired: async () => {
      const refresh = getStoredRefreshToken()
      if (!refresh) {
        return null
      }
      try {
        const bare = new ApiClient({ baseUrl })
        const result = await bare.refreshToken(refresh)
        if (result.access_token) {
          storeSession(result.access_token, result.refresh_token)
          return result.access_token
        }
      } catch {
        clearSession()
      }
      return null
    }
  })
}

export function getSharedClient(): ApiClient {
  if (!sharedClient) {
    sharedClient = createDefaultClient()
  }
  return sharedClient
}

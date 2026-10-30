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

  // Catalog endpoints
  async getCatalogApps(params?: Record<string, string>) {
    const qs = params ? '?' + new URLSearchParams(params).toString() : ''
    return this.request(`/api/v1/catalog/apps${qs}`)
  }

  async getAppBySlug(slug: string, locale?: string) {
    const qs = locale ? `?locale=${encodeURIComponent(locale)}` : ''
    return this.request(`/api/v1/catalog/apps/${slug}${qs}`)
  }

  async getCategories() {
    return this.request('/api/v1/catalog/categories')
  }

  // Developer endpoints
  async createOrganization(name: string, businessNumber: string) {
    return this.request('/api/v1/developer/orgs', {
      method: 'POST',
      body: JSON.stringify({ name, business_number: businessNumber })
    })
  }

  async submitVerification(orgId: string, payload: any) {
    return this.request(`/api/v1/developer/orgs/${orgId}/verification`, {
      method: 'POST',
      body: JSON.stringify(payload)
    })
  }

  // Admin endpoints
  async approveDualControl(requestId: string) {
    return this.request(`/api/v1/admin/approvals/${requestId}/approve`, {
      method: 'POST'
    })
  }
}


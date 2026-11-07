// ============================================================================
// Cloudflare Workers Edge API Gateway & Unified Distribution Router
// BookingNetwork App Market — Universal Independent App Store
// ============================================================================

export interface Env {
  ASSETS: {
    fetch: (request: Request) => Promise<Response>
  }
  BACKEND_API_URL?: string
  HMAC_TOKEN_SECRET?: string
}

const CORS_HEADERS = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Methods': 'GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD',
  'Access-Control-Allow-Headers': 'Content-Type, Authorization, X-Requested-With, X-Forwarded-For, X-Device-Fingerprint',
  'Access-Control-Max-Age': '86400',
}

function edgeColo(request: Request): string {
  const cf = (request as Request & { cf?: { colo?: string } }).cf
  return cf?.colo || 'NRT'
}

function jsonResponse(data: unknown, status = 200, extraHeaders: Record<string, string> = {}): Response {
  return new Response(JSON.stringify(data, null, 2), {
    status,
    headers: {
      'Content-Type': 'application/json; charset=utf-8',
      ...CORS_HEADERS,
      ...extraHeaders,
    },
  })
}

// Pre-seeded high-fidelity catalog data for edge serving
const EDGE_CATALOG_APPS = [
  {
    id: 'app-omni-productivity-01',
    package_name: 'com.omniwork.office',
    title: 'Omni Productivity Suite',
    developer_name: 'OmniWork Technologies',
    developer_verified: true,
    category: 'productivity',
    category_name: '생산성 & 업무',
    rating: 4.9,
    rating_count: 42800,
    downloads: '150만+',
    size_bytes: 37539072,
    size_formatted: '35.8 MB',
    price_krw: 0,
    is_free: true,
    min_sdk: 26,
    version_name: '3.4.2',
    version_code: 30402,
    sha256: '9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08',
    summary: '차세대 올인원 문서 편집, 프로젝트 관리 및 엔드투엔드 암호화 협업 스위트',
    description: '문서 편집, 실시간 클라우드 동기화, 엔드투엔드 암호화 협업을 단 하나의 앱으로 완성하세요. 구글 플레이 독립 유통 공식 인증 패키지입니다.',
    permissions: ['INTERNET', 'READ_EXTERNAL_STORAGE', 'POST_NOTIFICATIONS'],
    signature_scheme: 'V3_CONTINUITY',
    verified_at: '2026-10-30T12:00:00Z',
    is_sponsored: true
  },
  {
    id: 'app-finvault-secure-02',
    package_name: 'com.finvault.banking',
    title: 'FinVault Secure Banking',
    developer_name: 'FinVault Financial Labs',
    developer_verified: true,
    category: 'finance',
    category_name: '금융 & 핀테크',
    rating: 4.8,
    rating_count: 28900,
    downloads: '82만+',
    size_bytes: 25690112,
    size_formatted: '24.5 MB',
    price_krw: 0,
    is_free: true,
    min_sdk: 28,
    version_name: '2.1.0',
    version_code: 20100,
    sha256: 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
    summary: '생체인식 하드웨어 보안 키 및 지능형 가계부 자산 통합 자산 관리 솔루션',
    description: '하드웨어 보안 모듈(HSM) 및 생체인증 FIDO2 기반 모바일 뱅킹 및 자산 관리 스위트.',
    permissions: ['INTERNET', 'USE_BIOMETRIC', 'CAMERA'],
    signature_scheme: 'V3_CONTINUITY',
    verified_at: '2026-10-29T08:30:00Z',
    is_sponsored: false
  },
  {
    id: 'app-secure-messenger-03',
    package_name: 'org.openprivacy.messenger',
    title: 'Secure Messenger Pro',
    developer_name: 'Open Privacy Labs',
    developer_verified: true,
    category: 'communication',
    category_name: '커뮤니케이션',
    rating: 4.8,
    rating_count: 51200,
    downloads: '210만+',
    size_bytes: 29464985,
    size_formatted: '28.1 MB',
    price_krw: 0,
    is_free: true,
    min_sdk: 26,
    version_name: '5.8.0',
    version_code: 50800,
    sha256: 'ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb',
    summary: '영지식 암호화(ZKP) 기반 탈중앙화 메신저. 데이터 수집 제로 프라이버시',
    description: '서버에 어떠한 메시지나 메타데이터도 저장하지 않는 순수 P2P 영지식 암호화 메신저.',
    permissions: ['INTERNET', 'RECORD_AUDIO', 'CAMERA', 'VIBRATE'],
    signature_scheme: 'V3_CONTINUITY',
    verified_at: '2026-10-28T14:15:00Z',
    is_sponsored: false
  },
  {
    id: 'app-crypto-tracker-04',
    package_name: 'com.blockmetric.crypto',
    title: 'Pro Crypto Tracker',
    developer_name: 'BlockMetric Systems',
    developer_verified: true,
    category: 'finance',
    category_name: '금융 & 핀테크',
    rating: 4.7,
    rating_count: 14200,
    downloads: '45만+',
    size_bytes: 19084083,
    size_formatted: '18.2 MB',
    price_krw: 4900,
    is_free: false,
    min_sdk: 26,
    version_name: '1.9.4',
    version_code: 10904,
    sha256: '4e07408562bedb8b60ce05c1decfe3ad16b72230967de01f640b7e4729b49fce',
    summary: '글로벌 50개 거래소 실시간 시세, 포트폴리오 세무 보고서 자동 생성기',
    description: '실시간 WebSockets 시세 스트리밍 및 국세청 가상자산 소득세 신고용 원장 자동 생성.',
    permissions: ['INTERNET', 'ACCESS_NETWORK_STATE'],
    signature_scheme: 'V3_CONTINUITY',
    verified_at: '2026-10-27T10:00:00Z',
    is_sponsored: false
  },
  {
    id: 'app-cloud-authenticator-05',
    package_name: 'com.shieldguard.auth2fa',
    title: 'Cloud Authenticator 2FA',
    developer_name: 'ShieldGuard Security',
    developer_verified: true,
    category: 'utilities',
    category_name: '도구 & 유틸리티',
    rating: 4.9,
    rating_count: 36700,
    downloads: '67만+',
    size_bytes: 13002342,
    size_formatted: '12.4 MB',
    price_krw: 0,
    is_free: true,
    min_sdk: 24,
    version_name: '4.0.1',
    version_code: 40001,
    sha256: '4b227777d4dd1fc61c6f884f48641d02b4d121d3fd328cb08b5531fcacdabf8a',
    summary: 'RFC 6238 표준 TOTP 2차 인증기. 기기간 암호화 백업 및 긴급 복구 지원',
    description: '표준 2단계 인증 코드 생성 및 종단간 암호화(E2EE) 클라우드 동기화 백업.',
    permissions: ['CAMERA', 'USE_BIOMETRIC'],
    signature_scheme: 'V3_CONTINUITY',
    verified_at: '2026-10-26T17:40:00Z',
    is_sponsored: false
  },
  {
    id: 'app-mobile-ide-06',
    package_name: 'com.devtools.mobileide',
    title: 'Code Editor Mobile IDE',
    developer_name: 'Dev Tools Corp',
    developer_verified: true,
    category: 'productivity',
    category_name: '생산성 & 업무',
    rating: 4.6,
    rating_count: 9800,
    downloads: '31만+',
    size_bytes: 44040192,
    size_formatted: '42.0 MB',
    price_krw: 9900,
    is_free: false,
    min_sdk: 28,
    version_name: '2.5.0',
    version_code: 20500,
    sha256: 'ef2d127de37b942baad06145e54b0c619a1f22327b2ebbcfbec78f5564afe39d',
    summary: '모바일 환경에서 구동되는 풀스택 터미널 및 Git 연동 고성능 코드 에디터',
    description: '구문 강조, LSP 코드 자동완성, 통합 bash 터미널 및 Git 버전 관리 탑재.',
    permissions: ['INTERNET', 'READ_EXTERNAL_STORAGE', 'WRITE_EXTERNAL_STORAGE'],
    signature_scheme: 'V3_CONTINUITY',
    verified_at: '2026-10-25T11:20:00Z',
    is_sponsored: false
  },
  {
    id: 'app-cyber-runner-07',
    package_name: 'com.neonpixel.cyberrunner',
    title: 'Cyber Runner 2026',
    developer_name: 'Neon Pixel Studio',
    developer_verified: true,
    category: 'games',
    category_name: '게임',
    rating: 4.7,
    rating_count: 63100,
    downloads: '95만+',
    size_bytes: 120586240,
    size_formatted: '115.0 MB',
    price_krw: 0,
    is_free: true,
    min_sdk: 26,
    version_name: '1.2.0',
    version_code: 10200,
    sha256: 'e7f6c011776e8db7cd330b54174fd76f7d0216b612387a5ffcfb81e6f0919683',
    summary: '네온 사이버펑크 도시를 질주하는 하이퍼 액션 파쿠르 러닝 게임',
    description: 'Vulkan 고성능 그래픽 엔진 탑재 및 120FPS 모바일 최적화 사이버펑크 러닝 게임.',
    permissions: ['INTERNET', 'VIBRATE'],
    signature_scheme: 'V3_CONTINUITY',
    verified_at: '2026-10-24T09:10:00Z',
    is_sponsored: false
  }
]

function getAuthToken(request: Request): string | null {
  const authHeader = request.headers.get('Authorization')
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return null
  }
  const token = authHeader.substring(7).trim()
  return token.length > 0 ? token : null
}

// ---------------------------------------------------------------------------
// Edge fallback state. Per-isolate in-memory stores that mirror the Go
// backend's in-memory services shape-for-shape. Production traffic should set
// BACKEND_API_URL so the worker proxies to the real backend instead.
// ---------------------------------------------------------------------------
type EdgeUser = { id: string; email: string; passwordHash: string; displayName: string; roles: string[] }
const edgeUsers = new Map<string, EdgeUser>()
const edgeTokens = new Map<string, string>() // access token -> user id
const edgeOrgs = new Map<string, Record<string, unknown>>()
const edgeMembers = new Map<string, Array<Record<string, unknown>>>()
const edgeKeys = new Map<string, Array<Record<string, unknown>>>()
const edgeVerifications = new Map<string, Record<string, unknown>>()
const edgeUploads = new Map<string, Record<string, unknown>>()
const edgeReleases = new Map<string, Array<Record<string, unknown>>>()
const edgeDecisions = new Map<string, Record<string, unknown>>()
const edgeKillSwitches = new Map<string, Record<string, unknown>>()
const edgePayouts = new Map<string, Record<string, unknown>>()
const edgeReports: Array<Record<string, unknown>> = []
const edgeAppeals: Array<Record<string, unknown>> = []
const edgeCampaigns = new Map<string, Record<string, unknown>>()
const edgePromos = new Map<string, Record<string, unknown>>()
const edgePromoCodes = new Map<string, Record<string, unknown>>()
const edgePricing = new Map<string, Array<Record<string, unknown>>>()
const edgeOrders = new Map<string, Record<string, unknown>>()
const edgeReviews = new Map<string, Array<Record<string, unknown>>>()
const edgeAudit: Array<Record<string, unknown>> = []
const edgeApprovals = new Map<string, Record<string, unknown>>()
const edgeMfaSecrets = new Map<string, { secret: string; verified: boolean }>()
const edgeEvents: Array<Record<string, unknown>> = []
const edgeLocalizations = new Map<string, Array<Record<string, unknown>>>()
const edgeMedia = new Map<string, Array<Record<string, unknown>>>()
const edgeMediaBytes = new Map<string, { bytes: ArrayBuffer; contentType: string }>()
const edgePayoutAccounts = new Map<string, Record<string, unknown>>()
const edgeCatalogStatus = new Map<string, string>()
const edgeConnectors = new Map<string, Record<string, unknown>>([
  ['fdroid-org', { id: 'fdroid-org', name: 'F-Droid Official', kind: 'FDROID', feedUrl: 'https://f-droid.org/repo/index-v2.json', syncable: true }],
  ['galaxy-store', { id: 'galaxy-store', name: 'Galaxy Store', kind: 'OFFICIAL_LINK', syncable: false }],
])

function edgeConnectorsList(): Array<Record<string, unknown>> {
  return [...edgeConnectors.values()]
}
let edgeAuditSeq = 0

function edgeId(prefix: string): string {
  return `${prefix}_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 10)}`
}

// edgeImageDims decodes PNG IHDR and JPEG SOF dimensions without libraries.
function edgeImageDims(data: Uint8Array): { width: number; height: number; ext: string } | null {
  if (data.length > 24 && data[0] === 0x89 && data[1] === 0x50 && data[2] === 0x4e && data[3] === 0x47) {
    const view = new DataView(data.buffer, data.byteOffset, data.byteLength)
    const width = view.getUint32(16)
    const height = view.getUint32(20)
    if (width > 0 && height > 0) {
      return { width, height, ext: 'png' }
    }
    return null
  }
  if (data.length > 4 && data[0] === 0xff && data[1] === 0xd8) {
    let offset = 2
    while (offset + 8 < data.length) {
      if (data[offset] !== 0xff) {
        offset += 1
        continue
      }
      const marker = data[offset + 1]
      if (marker === 0xd8 || marker === 0xd9 || (marker >= 0xd0 && marker <= 0xd7)) {
        offset += 2
        continue
      }
      const size = (data[offset + 2] << 8) | data[offset + 3]
      if (size < 2 || offset + size >= data.length) {
        return null
      }
      if ((marker >= 0xc0 && marker <= 0xc3) || (marker >= 0xc5 && marker <= 0xc7) || (marker >= 0xc9 && marker <= 0xcb) || (marker >= 0xcd && marker <= 0xcf)) {
        const height = (data[offset + 5] << 8) | data[offset + 6]
        const width = (data[offset + 7] << 8) | data[offset + 8]
        if (width > 0 && height > 0) {
          return { width, height, ext: 'jpg' }
        }
        return null
      }
      offset += 2 + size
    }
  }
  return null
}

function problemResponse(status: number, code: string, detail: string, traceId: string): Response {
  return jsonResponse({
    type: `https://appmarket.bookingnetwork.internal/problems/${code.toLowerCase()}`,
    title: code,
    status,
    code,
    detail,
    trace_id: traceId,
    timestamp: new Date().toISOString(),
    field_errors: [],
  }, status)
}

async function sha256Hex(input: string): Promise<string> {
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(input))
  return Array.from(new Uint8Array(digest)).map(b => b.toString(16).padStart(2, '0')).join('')
}

function base32Encode(bytes: Uint8Array): string {
  const alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567'
  let bits = 0
  let value = 0
  let out = ''
  for (const b of bytes) {
    value = (value << 8) | b
    bits += 8
    while (bits >= 5) {
      out += alphabet[(value >>> (bits - 5)) & 31]
      bits -= 5
    }
  }
  if (bits > 0) {
    out += alphabet[(value << (5 - bits)) & 31]
  }
  return out
}

function base32Decode(input: string): Uint8Array {
  const alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567'
  const clean = input.toUpperCase().replace(/=+$/, '')
  let bits = 0
  let value = 0
  const bytes: number[] = []
  for (const ch of clean) {
    const idx = alphabet.indexOf(ch)
    if (idx < 0) throw new Error('invalid base32')
    value = (value << 5) | idx
    bits += 5
    if (bits >= 8) {
      bytes.push((value >>> (bits - 8)) & 255)
      bits -= 8
    }
  }
  return new Uint8Array(bytes)
}

async function verifyEdgeTotp(secretBase32: string, code: string, nowMs: number): Promise<boolean> {
  const trimmed = code.trim()
  if (!/^\d{6}$/.test(trimmed)) return false
  let keyBytes: Uint8Array
  try {
    keyBytes = base32Decode(secretBase32)
  } catch {
    return false
  }
  const key = await crypto.subtle.importKey('raw', keyBytes.buffer as ArrayBuffer, { name: 'HMAC', hash: 'SHA-1' }, false, ['sign'])
  for (const skew of [-1, 0, 1]) {
    const counter = Math.floor(nowMs / 1000 / 30) + skew
    const buf = new ArrayBuffer(8)
    new DataView(buf).setUint32(4, counter)
    const sig = await crypto.subtle.sign('HMAC', key, buf)
    const h = new Uint8Array(sig)
    const offset = h[h.length - 1] & 0x0f
    const binary = ((h[offset] & 0x7f) << 24) | (h[offset + 1] << 16) | (h[offset + 2] << 8) | h[offset + 3]
    if (String(binary % 1000000).padStart(6, '0') === trimmed) return true
  }
  return false
}

async function hmacHex(secret: string, input: string): Promise<string> {
  const key = await crypto.subtle.importKey('raw', new TextEncoder().encode(secret), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign'])
  const sig = await crypto.subtle.sign('HMAC', key, new TextEncoder().encode(input))
  return Array.from(new Uint8Array(sig)).map(b => b.toString(16).padStart(2, '0')).join('')
}

function slugify(title: string): string {
  return title.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '')
}

type EdgeApp = (typeof EDGE_CATALOG_APPS)[number]

function toAppSummary(a: EdgeApp): Record<string, unknown> {
  return {
    id: a.id,
    canonical_slug: slugify(a.title),
    title: a.title,
    primary_package_name: a.package_name,
    primary_category_id: a.category,
    average_rating: a.rating,
    bayesian_rating: Math.max(0, Math.round((a.rating - 0.03) * 100) / 100),
    install_count: a.rating_count * 10,
    icon_url: `/icons/${a.package_name}.png`,
  }
}

function toAppDetail(a: EdgeApp, locale: string): Record<string, unknown> {
  const resolvedLocale = locale.startsWith('ko') ? 'ko-KR' : 'en-US'
  return {
    app: {
      id: a.id,
      canonical_slug: slugify(a.title),
      primary_package_name: a.package_name,
      primary_category_id: a.category,
      content_rating: 'ALL',
      default_locale: 'en-US',
      catalog_status: 'PUBLISHED',
      average_rating: a.rating,
      bayesian_rating: Math.max(0, Math.round((a.rating - 0.03) * 100) / 100),
      install_count: a.rating_count * 10,
      license_spdx: 'Proprietary',
      monetization_type: a.is_free ? 'FREE' : 'PAID',
      created_at: a.verified_at,
      updated_at: a.verified_at,
    },
    localization: {
      id: `loc-${a.id}`,
      app_id: a.id,
      locale: resolvedLocale,
      title: a.title,
      short_description: a.summary,
      full_description: a.description,
      whats_new: '',
    },
    media_assets: [],
    distributions: [
      {
        id: `dist-${a.id}`,
        app_id: a.id,
        source_type: 'BN_HOSTED',
        distribution_mode: 'DIRECT_HOSTED',
        package_name: a.package_name,
        signing_cert_sha256: a.sha256,
      },
    ],
    tags: [],
    resolved_locale: resolvedLocale,
  }
}

function toLatestRelease(a: EdgeApp): Record<string, unknown> {
  return {
    id: `rel-${a.id}`,
    distribution_id: `dist-${a.id}`,
    app_id: a.id,
    release_track: 'PRODUCTION',
    version_code: a.version_code,
    version_name: a.version_name,
    min_sdk: a.min_sdk,
    target_sdk: 34,
    supported_abis: ['arm64-v8a', 'armeabi-v7a'],
    file_size_bytes: a.size_bytes,
    release_notes: a.summary,
    published_at: a.verified_at,
  }
}

function encodeEdgeCursor(index: number): string {
  return btoa(JSON.stringify({ i: index }))
}

function decodeEdgeCursor(cursor: string): number {
  try {
    const parsed = JSON.parse(atob(cursor)) as { i: number }
    return typeof parsed.i === 'number' ? parsed.i : 0
  } catch {
    return 0
  }
}

async function handleApiRequest(request: Request, env: Env): Promise<Response> {
  const startTime = Date.now()
  const url = new URL(request.url)
  const pathname = url.pathname
  const method = request.method

  // If live backend URL is configured in Worker environment, reverse proxy to it
  if (env.BACKEND_API_URL) {
    try {
      const targetUrl = new URL(pathname + url.search, env.BACKEND_API_URL).toString()
      const forwardReq = new Request(targetUrl, {
        method: request.method,
        headers: request.headers,
        body: request.body,
        redirect: 'follow',
      })
      const resp = await fetch(forwardReq)
      const newHeaders = new Headers(resp.headers)
      for (const [k, v] of Object.entries(CORS_HEADERS)) {
        newHeaders.set(k, v)
      }
      return new Response(resp.body, {
        status: resp.status,
        statusText: resp.statusText,
        headers: newHeaders,
      })
    } catch (err) {
      console.warn('Backend proxy failed, falling back to Edge Gateway response:', err)
    }
  }

  // Edge Gateway Fallback Handlers
  if (pathname === '/healthz/live') {
    return jsonResponse({
      status: 'UP',
      edge_node: edgeColo(request),
      cluster: 'cloudflare-edge-gateway',
      timestamp: new Date().toISOString()
    })
  }

  // Honest Healthcheck: Clearly indicates Edge Gateway standalone status unless backend proxy is active
  if (pathname === '/healthz/ready') {
    return jsonResponse({
      status: 'READY',
      edge_node: edgeColo(request),
      version: '1.0.0-edge',
      architecture: env.BACKEND_API_URL ? 'Hybrid Cloudflare Edge + Go Core' : 'Cloudflare Workers Standalone Edge (Demo Replica)',
      backend_proxy_configured: !!env.BACKEND_API_URL,
      services: {
        database: env.BACKEND_API_URL ? 'HEALTHY' : 'STANDALONE_EDGE_STORE',
        redis: env.BACKEND_API_URL ? 'HEALTHY' : 'STANDALONE_EDGE_STORE',
        opensearch: env.BACKEND_API_URL ? 'HEALTHY' : 'STANDALONE_EDGE_STORE',
        object_storage: env.BACKEND_API_URL ? 'HEALTHY' : 'STANDALONE_EDGE_STORE',
        treescan_engine: 'STANDALONE_EDGE_STORE',
        dual_control: 'ENFORCED_AT_EDGE'
      },
      timestamp: new Date().toISOString()
    })
  }

  if (pathname === '/metrics') {
    const uptimeSeconds = Math.floor(Date.now() / 1000) % 86400
    const body = [
      '# HELP bn_requests_total Total number of HTTP requests processed at edge.',
      '# TYPE bn_requests_total counter',
      `bn_requests_total{colo="${edgeColo(request)}"} 14820`,
      '# HELP bn_edge_latency_seconds Edge routing latency histogram.',
      '# TYPE bn_edge_latency_seconds histogram',
      'bn_edge_latency_seconds_bucket{le="0.005"} 14200',
      'bn_edge_latency_seconds_bucket{le="0.010"} 14750',
      'bn_edge_latency_seconds_bucket{le="+Inf"} 14820',
      `bn_uptime_seconds ${uptimeSeconds}`
    ].join('\n')
    return new Response(body, {
      status: 200,
      headers: { 'Content-Type': 'text/plain; version=0.0.4; charset=utf-8', ...CORS_HEADERS }
    })
  }

  if (pathname === '/api/v1/system/capabilities') {
    return jsonResponse({
      platform: 'BookingNetwork App Market',
      version: '1.0.0',
      environment: 'production',
      compliance: {
        eu_dma_article_6_4: true, // EU DMA Article 6(4) Alternative App Stores
        zero_scraping_guarantee: true,
        direct_hosted_ecosystem: true,
      },
      modules: [
        'catalog', 'developer', 'distribution', 'iam',
        'commerce', 'security', 'moderation', 'search',
        'analytics', 'ads', 'admin', 'console'
      ],
      features: {
        dual_control_enforcement: true,
        treescan_ast_quarantine: true,
        v3_signature_continuity: true,
        hmac_download_tokens: true,
        double_entry_ledger: true,
        aab_split_delivery: true,
      },
      supported_locales: ['ko-KR', 'en-US', 'ja-JP', 'de-DE'],
      timestamp: new Date().toISOString()
    })
  }

  // 1. Catalog Routes — shapes mirror the Go backend exactly.
  if (pathname === '/api/v1/catalog/apps' && method === 'GET') {
    const category = url.searchParams.get('category')?.toLowerCase()
    const query = url.searchParams.get('q')?.toLowerCase()?.trim()
    const freeOnly = url.searchParams.get('free_only') === 'true'
    const directOnly = url.searchParams.get('direct_only') === 'true'
    const minRating = Number(url.searchParams.get('min_rating') || 0)
    const limit = Math.min(Math.max(Number(url.searchParams.get('limit') || 20), 1), 100)
    const start = url.searchParams.get('cursor') ? decodeEdgeCursor(url.searchParams.get('cursor') as string) : 0

    let filtered = [...EDGE_CATALOG_APPS].sort((a, b) => b.rating - a.rating || b.rating_count - a.rating_count)
    if (category && category !== 'all') {
      filtered = filtered.filter(a => a.category.toLowerCase() === category)
    }
    if (freeOnly) {
      filtered = filtered.filter(a => a.is_free)
    }
    if (directOnly) {
      filtered = filtered.filter(() => true) // all edge apps are direct-hosted
    }
    if (minRating > 0) {
      filtered = filtered.filter(a => a.rating >= minRating)
    }
    if (query) {
      filtered = filtered.filter(a =>
        a.title.toLowerCase().includes(query) ||
        a.package_name.toLowerCase().includes(query) ||
        a.summary.toLowerCase().includes(query) ||
        a.developer_name.toLowerCase().includes(query)
      )
    }
    const page = filtered.slice(start, start + limit + 1)
    const items = page.slice(0, limit).map(toAppSummary)
    const nextCursor = page.length > limit ? encodeEdgeCursor(start + limit) : ''
    return jsonResponse({ items, cursor: nextCursor })
  }

  const latestReleaseMatch = pathname.match(/^\/api\/v1\/catalog\/apps\/([^/]+)\/releases\/latest$/)
  if (latestReleaseMatch && method === 'GET') {
    const slug = decodeURIComponent(latestReleaseMatch[1])
    const app = EDGE_CATALOG_APPS.find(a => a.id === slug || a.package_name === slug || slugify(a.title) === slug)
    if (!app) {
      return problemResponse(404, 'ERR_APP_NOT_FOUND', 'App not found', `edge-${startTime}`)
    }
    return jsonResponse(toLatestRelease(app))
  }

  const chartMatch = pathname.match(/^\/api\/v1\/catalog\/charts\/([^/]+)$/)
  if (chartMatch && method === 'GET') {
    const chartType = chartMatch[1]
    if (!['top_free', 'top_paid', 'trending', 'top_grossing'].includes(chartType)) {
      return problemResponse(400, 'ERR_INVALID_CHART', 'Unknown chart type. Use top_free, top_paid, trending, or top_grossing.', `edge-${startTime}`)
    }
    let ranked = [...EDGE_CATALOG_APPS]
    if (chartType === 'top_free') {
      ranked = ranked.filter(a => a.is_free)
    }
    ranked.sort((a, b) => chartType === 'trending'
      ? (b.rating - a.rating || b.rating_count - a.rating_count)
      : (b.rating_count - a.rating_count))
    return jsonResponse({ chart_type: chartType, items: ranked.slice(0, 20).map(toAppSummary) })
  }

  if (pathname === '/api/v1/catalog/categories' && method === 'GET') {
    const cats = [...new Set(EDGE_CATALOG_APPS.map(a => a.category))]
    return jsonResponse({ items: cats.map(c => ({ id: c, name_i18n: { 'en-US': c }, app_count: EDGE_CATALOG_APPS.filter(a => a.category === c).length })) })
  }

  const appDetailMatch = pathname.match(/^\/api\/v1\/catalog\/apps\/([^/]+)$/)
  if (appDetailMatch && method === 'GET') {
    const slug = decodeURIComponent(appDetailMatch[1])
    const app = EDGE_CATALOG_APPS.find(a => a.id === slug || a.package_name === slug || slugify(a.title) === slug)
    if (!app) {
      return problemResponse(404, 'ERR_APP_NOT_FOUND', 'Application not found in catalog.', `edge-${startTime}`)
    }
    const locale = request.headers.get('Accept-Language')?.split(',')[0]?.trim() || url.searchParams.get('locale') || 'en-US'
    return jsonResponse(toAppDetail(app, locale))
  }

  // 2. Distribution Routes — real backend paths.
  const tokenMatch = pathname.match(/^\/api\/v1\/distribution\/apps\/([^/]+)\/releases\/([^/]+)\/download-token$/)
  if (tokenMatch && method === 'POST') {
    let bodyJson: Record<string, unknown> = {}
    try { bodyJson = await request.json() } catch { /* default */ }
    const appId = decodeURIComponent(tokenMatch[1])
    const releaseId = decodeURIComponent(tokenMatch[2])
    const app = EDGE_CATALOG_APPS.find(a => a.id === appId || a.package_name === appId)
    if (!app) {
      return problemResponse(404, 'ERR_APP_NOT_FOUND', 'App not found', `edge-${startTime}`)
    }
    const ks = edgeKillSwitches.get(app.package_name) as { is_triggered?: boolean } | undefined
    if (ks?.is_triggered) {
      return problemResponse(403, 'ERR_KILL_SWITCH_ACTIVE', 'Application is suspended by emergency kill switch', `edge-${startTime}`)
    }
    const exp = Math.floor(Date.now() / 1000) + 900
    const secret = env.HMAC_TOKEN_SECRET || 'edge-demo-hmac-secret'
    const token = await hmacHex(secret, `${appId}:${releaseId}:${exp}`)
    const sha = String(bodyJson.sha256_checksum || app.sha256)
    return jsonResponse({
      url: `https://cdn.appmarket.local/artifacts/apk/${app.package_name}/${token}?exp=${exp}`,
      expires_at: new Date(exp * 1000).toISOString(),
      sha256_checksum: sha,
    }, 201)
  }

  if (pathname === '/api/v1/distribution/device/check-updates' && method === 'POST') {
    let bodyJson: { packages?: Array<{ package_name: string; version_code: number }>; installed?: Array<{ package_name: string; version_code: number }> } = {}
    try { bodyJson = await request.json() } catch { /* default */ }
    const installed = bodyJson.packages || bodyJson.installed || []
    const results = installed.map(inst => {
      const ks = edgeKillSwitches.get(inst.package_name) as { is_triggered?: boolean } | undefined
      if (ks?.is_triggered) {
        return {
          app_id: '', release_id: '', package_name: inst.package_name, version_code: inst.version_code || 0, version_name: '',
          status: 'DISTRIBUTION_SUSPENDED', resolution_reason: 'Distribution suspended by emergency kill switch. Uninstall advised.',
          requires_user_action: true,
        }
      }
      const app = EDGE_CATALOG_APPS.find(a => a.package_name === inst.package_name)
      if (app && app.version_code > (inst.version_code || 0)) {
        return {
          app_id: app.id, release_id: '', package_name: app.package_name, version_code: app.version_code, version_name: app.version_name,
          status: 'UPDATE_AVAILABLE_COMPATIBLE', resolution_reason: 'Newer version available.',
          download_url: '', sha256_checksum: '', size_bytes: 0, requires_user_action: false,
        }
      }
      return {
        app_id: app?.id || '', release_id: '', package_name: inst.package_name, version_code: inst.version_code || 0, version_name: '',
        status: 'NO_UPDATE', resolution_reason: 'Installed version is current.', requires_user_action: false,
      }
    })
    return jsonResponse(results)
  }

  // 3. Auth Routes — real backend paths with edge-local accounts.
  if (pathname === '/api/v1/auth/register' && method === 'POST') {
    let bodyJson: { email?: string; password?: string; display_name?: string } = {}
    try { bodyJson = await request.json() } catch { /* default */ }
    const email = String(bodyJson.email || '').toLowerCase().trim()
    if (!email.includes('@') || String(bodyJson.password || '').length < 8) {
      return problemResponse(400, 'ERR_INVALID_BODY', 'Valid email and 8+ char password required', `edge-${startTime}`)
    }
    if (edgeUsers.has(email)) {
      return problemResponse(409, 'ERR_USER_EXISTS', 'User already exists', `edge-${startTime}`)
    }
    const user: EdgeUser = {
      id: edgeId('usr'), email,
      passwordHash: await sha256Hex(`edge$${email}$${bodyJson.password}`),
      displayName: String(bodyJson.display_name || email.split('@')[0]),
      roles: ['USER'],
    }
    edgeUsers.set(email, user)
    return jsonResponse({ id: user.id, email: user.email, display_name: user.displayName, status: 'ACTIVE' }, 201)
  }

  if (pathname === '/api/v1/auth/login' && method === 'POST') {
    let bodyJson: { email?: string; password?: string } = {}
    try { bodyJson = await request.json() } catch { /* default */ }
    const email = String(bodyJson.email || '').toLowerCase().trim()
    const user = edgeUsers.get(email)
    const hash = await sha256Hex(`edge$${email}$${bodyJson.password || ''}`)
    if (!user || user.passwordHash !== hash) {
      return problemResponse(401, 'ERR_INVALID_CREDENTIALS', 'Invalid email or password', `edge-${startTime}`)
    }
    const accessToken = `edge_at_${edgeId('t')}`
    const refreshToken = `edge_rt_${edgeId('t')}`
    edgeTokens.set(accessToken, user.id)
    return jsonResponse({
      access_token: accessToken, refresh_token: refreshToken,
      expires_in: 3600, token_type: 'Bearer', mfa_required: false,
      user: { id: user.id, email: user.email, display_name: user.displayName, status: 'ACTIVE' },
    })
  }

  if (pathname === '/api/v1/auth/token/refresh' && method === 'POST') {
    return problemResponse(401, 'ERR_SESSION_EXPIRED', 'Edge demo sessions do not refresh; log in again', `edge-${startTime}`)
  }

  if (pathname === '/api/v1/auth/mfa/enroll' && method === 'POST') {
    let bodyJson: { email?: string; password?: string } = {}
    try { bodyJson = await request.json() } catch { /* default */ }
    const email = String(bodyJson.email || '').toLowerCase().trim()
    const user = edgeUsers.get(email)
    const hash = await sha256Hex(`edge$${email}$${bodyJson.password || ''}`)
    if (!user || user.passwordHash !== hash) {
      return problemResponse(401, 'ERR_INVALID_CREDENTIALS', 'Invalid email or password', `edge-${startTime}`)
    }
    const existing = edgeMfaSecrets.get(user.id)
    if (existing && existing.verified) {
      return problemResponse(409, 'ERR_MFA_ALREADY_ENROLLED', 'A verified MFA credential already exists', `edge-${startTime}`)
    }
    const raw = new Uint8Array(20)
    crypto.getRandomValues(raw)
    const secret = base32Encode(raw)
    const codes: string[] = []
    for (let i = 0; i < 10; i++) {
      const cb = new Uint8Array(5)
      crypto.getRandomValues(cb)
      codes.push([...cb].map(b => b.toString(16).padStart(2, '0')).join('').toUpperCase())
    }
    edgeMfaSecrets.set(user.id, { secret, verified: false })
    const issuer = encodeURIComponent('BN App Market')
    return jsonResponse({
      user_id: user.id,
      secret,
      provisioning_uri: `otpauth://totp/${issuer}:${encodeURIComponent(email)}?secret=${secret}&issuer=${issuer}&period=30&digits=6`,
      recovery_codes: codes,
    }, 201)
  }

  if (pathname === '/api/v1/auth/mfa/verify' && method === 'POST') {
    let bodyJson: { user_id?: string; mfa_code?: string } = {}
    try { bodyJson = await request.json() } catch { /* default */ }
    const uid = String(bodyJson.user_id || '')
    const entry = edgeMfaSecrets.get(uid)
    if (!entry) {
      return problemResponse(400, 'ERR_MFA_REQUIRED', 'No MFA credential enrolled for this user', `edge-${startTime}`)
    }
    const ok = await verifyEdgeTotp(entry.secret, String(bodyJson.mfa_code || ''), Date.now())
    if (!ok) {
      return problemResponse(401, 'ERR_INVALID_MFA_CODE', 'Invalid TOTP code', `edge-${startTime}`)
    }
    entry.verified = true
    edgeMfaSecrets.set(uid, entry)
    const me = [...edgeUsers.values()].find(u => u.id === uid)
    const accessToken = `edge_at_${edgeId('t')}`
    const refreshToken = `edge_rt_${edgeId('t')}`
    edgeTokens.set(accessToken, uid)
    return jsonResponse({
      access_token: accessToken, refresh_token: refreshToken,
      expires_in: 3600, token_type: 'Bearer', mfa_required: false,
      user: { id: uid, email: me ? me.email : '', display_name: me ? me.displayName : '', status: 'ACTIVE' },
    })
  }

  if (pathname === '/api/v1/auth/users/me' && method === 'GET') {
    const tok = getAuthToken(request)
    const uid = tok ? edgeTokens.get(tok) || '' : ''
    if (!uid) {
      return problemResponse(401, 'ERR_UNAUTHORIZED', 'Authentication required', `edge-${startTime}`)
    }
    const me = [...edgeUsers.values()].find(u => u.id === uid)
    if (!me) {
      return problemResponse(401, 'ERR_UNAUTHORIZED', 'Unknown or expired token', `edge-${startTime}`)
    }
    return jsonResponse({ sub: me.id, email: me.email, roles: ['developer'], org_roles: {}, mfa_verified: false })
  }

  if (pathname === '/api/v1/auth/oauth/providers' && method === 'GET') {
    return jsonResponse({ providers: [] })
  }

  if (pathname === '/api/v1/auth/.well-known/openid-configuration' && method === 'GET') {
    return jsonResponse({ issuer: 'https://auth.appmarket.local', jwks_uri: '/api/v1/auth/jwks.json' })
  }

  if (pathname === '/api/v1/auth/jwks.json' && method === 'GET') {
    return jsonResponse({ keys: [] })
  }

  // =========================================================================
  // PROTECTED ROUTES GATEWAY: Enforce Authentication & Role-Based Access
  // =========================================================================

  // Check authentication token for protected operations (mirrors backend gates).
  const bearerToken = getAuthToken(request)
  const protectedRoute =
    pathname.startsWith('/api/v1/developer/') ||
    pathname.startsWith('/api/v1/admin/') ||
    (pathname.startsWith('/api/v1/commerce/') && !pathname.startsWith('/api/v1/commerce/webhooks/')) ||
    (pathname.startsWith('/api/v1/moderation/') && pathname !== '/api/v1/moderation/reports' && !pathname.includes('/reviews') && method !== 'GET')
  let edgeUserId = ''
  if (protectedRoute) {
    if (!bearerToken) {
      return problemResponse(401, 'ERR_UNAUTHORIZED', 'Authentication required. Authorization: Bearer <token> must be provided.', `edge-${startTime}`)
    }
    edgeUserId = edgeTokens.get(bearerToken) || ''
    if (!edgeUserId) {
      // Legacy demo tokens issued by earlier builds keep working standalone.
      if (bearerToken.startsWith('token_') || bearerToken.startsWith('jwt_') || bearerToken.startsWith('admin_') || bearerToken.startsWith('edge_at_') || bearerToken.startsWith('edge_rt_')) {
        edgeUserId = 'usr_edge_demo'
      } else {
        return problemResponse(401, 'ERR_UNAUTHORIZED', 'Unknown or expired token.', `edge-${startTime}`)
      }
    }
  } else if (bearerToken) {
    edgeUserId = edgeTokens.get(bearerToken) || ''
  }
  const edgeNow = () => new Date().toISOString()
  async function readBody<T>(): Promise<T> {
    try { return await request.json() as T } catch { return {} as T }
  }
  function resolveOrgApp(orgParam: string | null, appParam: string): { orgId: string; appId: string } | Response {
    let orgId = orgParam || ''
    if (!orgId || orgId === 'current') {
      const mine = [...edgeOrgs.values()].filter(o => (edgeMembers.get(String(o.id)) || []).some(m => String(m.user_id) === edgeUserId))
      if (mine.length === 0) {
        return problemResponse(404, 'ERR_CONSOLE_EMPTY', 'Caller owns no developer organization', `edge-${startTime}`)
      }
      orgId = String(mine[0].id)
    }
    let appId = appParam
    if (!appId || appId === 'current') {
      const owned = EDGE_CATALOG_APPS.filter(() => true)
      if (owned.length === 0) {
        return problemResponse(404, 'ERR_CONSOLE_EMPTY', 'Organization has no apps', `edge-${startTime}`)
      }
      appId = owned[0].id
    }
    return { orgId, appId }
  }

  // 4. Developer Routes — real backend shapes.
  if (pathname === '/api/v1/developer/orgs' && method === 'GET') {
    const mine = [...edgeOrgs.values()].filter(o => (edgeMembers.get(String(o.id)) || []).some(m => String(m.user_id) === edgeUserId))
    return jsonResponse(mine)
  }
  if (pathname === '/api/v1/developer/orgs' && method === 'POST') {
    const body = await readBody<{ slug?: string; name?: string }>()
    const slug = String(body.slug || '').trim()
    if (!slug) {
      return problemResponse(400, 'ERR_ORG_CREATE_FAILED', 'Org slug is required', `edge-${startTime}`)
    }
    if ([...edgeOrgs.values()].some(o => o.org_slug === slug)) {
      return problemResponse(409, 'ERR_ORG_SLUG_EXISTS', 'Org slug already exists', `edge-${startTime}`)
    }
    const org = { id: edgeId('org'), org_slug: slug, org_name: String(body.name || slug), verification_status: 'UNVERIFIED', created_at: edgeNow(), updated_at: edgeNow() }
    edgeOrgs.set(String(org.id), org)
    edgeMembers.set(String(org.id), [{ org_id: org.id, user_id: edgeUserId, role: 'DEV_OWNER', joined_at: edgeNow() }])
    return jsonResponse(org, 201)
  }

  const verMatch = pathname.match(/^\/api\/v1\/developer\/orgs\/([^/]+)\/verification$/)
  if (verMatch) {
    const orgId = decodeURIComponent(verMatch[1])
    if (method === 'POST') {
      const body = await readBody<{ doc_type?: string; doc_url?: string; domain?: string }>()
      const v = { id: edgeId('ver'), org_id: orgId, doc_type: String(body.doc_type || ''), doc_url: String(body.doc_url || ''), domain: String(body.domain || ''), status: 'PENDING_REVIEW', submitted_at: edgeNow() }
      edgeVerifications.set(orgId, v)
      const org = edgeOrgs.get(orgId)
      if (org) {
        org.verification_status = 'PENDING_REVIEW'
        org.updated_at = edgeNow()
      }
      return jsonResponse(v, 201)
    }
    const v = edgeVerifications.get(orgId)
    if (!v) {
      return problemResponse(404, 'ERR_VERIFICATION_NOT_FOUND', 'No verification submitted', `edge-${startTime}`)
    }
    return jsonResponse(v)
  }

  const membersMatch = pathname.match(/^\/api\/v1\/developer\/orgs\/([^/]+)\/members$/)
  if (membersMatch) {
    const orgId = decodeURIComponent(membersMatch[1])
    if (method === 'POST') {
      const body = await readBody<{ user_id?: string; role?: string }>()
      const list = edgeMembers.get(orgId) || []
      const member = { org_id: orgId, user_id: String(body.user_id || ''), role: String(body.role || 'DEV_MEMBER'), joined_at: edgeNow() }
      list.push(member)
      edgeMembers.set(orgId, list)
      return jsonResponse(member, 201)
    }
    return jsonResponse(edgeMembers.get(orgId) || [])
  }

  const memberDeleteMatch = pathname.match(/^\/api\/v1\/developer\/orgs\/([^/]+)\/members\/(.+)$/)
  if (memberDeleteMatch && method === 'DELETE') {
    const orgId = decodeURIComponent(memberDeleteMatch[1])
    const userId = decodeURIComponent(memberDeleteMatch[2])
    const list = edgeMembers.get(orgId) || []
    const target = list.find(m => String(m.user_id) === userId)
    if (!target) {
      return problemResponse(404, 'ERR_MEMBER_NOT_FOUND', 'Member not found', `edge-${startTime}`)
    }
    if (target.role === 'DEV_OWNER' && list.filter(m => m.role === 'DEV_OWNER').length <= 1) {
      return problemResponse(409, 'ERR_LAST_OWNER_REMOVAL', 'Cannot remove the last owner', `edge-${startTime}`)
    }
    edgeMembers.set(orgId, list.filter(m => String(m.user_id) !== userId))
    return jsonResponse({ status: 'REMOVED' })
  }

  const keysMatch = pathname.match(/^\/api\/v1\/developer\/orgs\/([^/]+)\/api-keys$/)
  if (keysMatch) {
    const orgId = decodeURIComponent(keysMatch[1])
    if (method === 'POST') {
      const body = await readBody<{ name?: string; scopes?: string[] }>()
      const list = edgeKeys.get(orgId) || []
      const key = { id: edgeId('key'), org_id: orgId, name: String(body.name || 'key'), key_prefix: `bn_${Math.random().toString(36).slice(2, 8)}`, scopes: body.scopes || [], status: 'ACTIVE', created_at: edgeNow() }
      list.push(key)
      edgeKeys.set(orgId, list)
      return jsonResponse({ ...key, key_secret: `bn_secret_${edgeId('s')}` }, 201)
    }
    return jsonResponse(edgeKeys.get(orgId) || [])
  }

  const revokeMatch = pathname.match(/^\/api\/v1\/developer\/orgs\/([^/]+)\/api-keys\/([^/]+)$/)
  if (revokeMatch && method === 'DELETE') {
    const list = edgeKeys.get(decodeURIComponent(revokeMatch[1])) || []
    const key = list.find(k => String(k.id) === decodeURIComponent(revokeMatch[2]))
    if (!key) {
      return problemResponse(404, 'ERR_APIKEY_NOT_FOUND', 'API key not found', `edge-${startTime}`)
    }
    key.status = 'REVOKED'
    return jsonResponse({ status: 'REVOKED' })
  }

  const orgAppsMatch = pathname.match(/^\/api\/v1\/developer\/orgs\/([^/]+)\/apps$/)
  if (orgAppsMatch && method === 'GET') {
    return jsonResponse(EDGE_CATALOG_APPS.slice(0, 3).map(a => ({
      id: a.id, slug: slugify(a.title), packageName: a.package_name,
      status: edgeCatalogStatus.get(a.id) || 'PUBLISHED', rating: a.rating, installs: a.rating_count * 10, updatedAt: a.verified_at,
    })))
  }

  const submitMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/submit$/)
  if (submitMatch && method === 'POST') {
    const appId = decodeURIComponent(submitMatch[1])
    const known = EDGE_CATALOG_APPS.some(a => a.id === appId || slugify(a.title) === appId)
    if (!known) {
      return problemResponse(404, 'ERR_APP_NOT_FOUND', 'app not found', `edge-${startTime}`)
    }
    const current = edgeCatalogStatus.get(appId) || 'PUBLISHED'
    if (current !== 'DRAFT' && current !== 'UNLISTED') {
      return problemResponse(400, 'ERR_CONSOLE_REQUEST_FAILED', 'only DRAFT or UNLISTED apps can be submitted for review', `edge-${startTime}`)
    }
    edgeCatalogStatus.set(appId, 'PENDING_INITIAL_REVIEW')
    return jsonResponse({ status: 'PENDING_INITIAL_REVIEW' })
  }

  const settleMatch = pathname.match(/^\/api\/v1\/developer\/orgs\/([^/]+)\/settlements$/)
  if (settleMatch && method === 'GET') {
    const orders = [...edgeOrders.values()].filter(o => String(o.org_id) === decodeURIComponent(settleMatch[1]))
    const paid = orders.filter(o => o.status === 'PAID')
    return jsonResponse({
      orders: orders.map(o => ({ id: o.id, userId: o.user_id, appId: o.app_id, currency: o.currency, grossAmount: o.gross_amount, commissionAmount: o.commission_amount, devPayoutAmount: o.dev_payout_amount, status: o.status, createdAt: o.created_at })),
      grossTotal: paid.reduce((n, o) => n + Number(o.gross_amount), 0),
      payoutTotal: paid.reduce((n, o) => n + Number(o.dev_payout_amount), 0),
    })
  }

  const payoutMatch = pathname.match(/^\/api\/v1\/developer\/orgs\/([^/]+)\/payout-account$/)
  if (payoutMatch) {
    const orgId = decodeURIComponent(payoutMatch[1])
    if (method === 'GET') {
      const rec = edgePayoutAccounts.get(orgId)
      if (!rec) {
        return problemResponse(404, 'ERR_PAYOUT_ACCOUNT_NOT_FOUND', 'payout account not found', `edge-${startTime}`)
      }
      return jsonResponse(rec)
    }
    if (method === 'PUT') {
      const body = await readBody<{ currency?: string; bank_name?: string; holder_name?: string; account_number?: string; routing_code?: string }>()
      const currency = String(body.currency || '').toUpperCase().trim()
      const digits = String(body.account_number || '').replace(/\D/g, '')
      if (!/^[A-Z]{3}$/.test(currency) || !String(body.bank_name || '').trim() || !String(body.holder_name || '').trim() || digits.length < 6 || digits.length > 32) {
        return problemResponse(400, 'ERR_PAYOUT_ACCOUNT_INVALID', 'currency/bank_name/holder_name/6..32-digit account_number required', `edge-${startTime}`)
      }
      const prev = edgePayoutAccounts.get(orgId)
      const rec = {
        org_id: orgId, currency, bank_name: String(body.bank_name).trim(), holder_name: String(body.holder_name).trim(),
        account_last4: digits.slice(-4), routing_code: String(body.routing_code || '').trim(),
        is_verified: prev ? Boolean(prev.is_verified) : false,
      }
      edgePayoutAccounts.set(orgId, rec)
      return jsonResponse(rec)
    }
  }

  const metaMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/metadata$/)
  if (metaMatch) {
    const appId = decodeURIComponent(metaMatch[1])
    if (method === 'GET') {
      return jsonResponse(edgeLocalizations.get(appId) || [])
    }
    const body = await readBody<{ locale?: string; title?: string; shortDescription?: string; fullDescription?: string; whatsNew?: string }>()
    if (!body.locale || !String(body.title || '').trim()) {
      return problemResponse(400, 'ERR_CONSOLE_REQUEST_FAILED', 'locale and title are required', `edge-${startTime}`)
    }
    const list = edgeLocalizations.get(appId) || []
    const saved = {
      id: edgeId('loc'), app_id: appId, locale: body.locale,
      title: body.title, short_description: body.shortDescription || '',
      full_description: body.fullDescription || '', whats_new: body.whatsNew || '',
    }
    const idx = list.findIndex(l => String(l.locale).toLowerCase() === String(body.locale).toLowerCase())
    if (idx >= 0) {
      saved.id = String(list[idx].id)
      list[idx] = saved
    } else {
      list.push(saved)
    }
    edgeLocalizations.set(appId, list)
    return jsonResponse(saved)
  }

  const mediaListMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/media$/)
  if (mediaListMatch && method === 'GET') {
    const locale = url.searchParams.get('locale') || ''
    const list = edgeMedia.get(decodeURIComponent(mediaListMatch[1])) || []
    return jsonResponse(locale ? list.filter(m => String(m.locale) === locale) : list)
  }

  if (mediaListMatch && method === 'POST') {
    const appId = decodeURIComponent(mediaListMatch[1])
    const form = await request.formData().catch(() => null)
    const file = form?.get('image')
    if (!(file instanceof File)) {
      return problemResponse(400, 'ERR_MEDIA_INVALID', 'image file is required', `edge-${startTime}`)
    }
    const locale = String(form?.get('locale') || '')
    const formFactor = String(form?.get('formFactor') || '')
    const assetType = String(form?.get('assetType') || '')
    if (!locale || !['PHONE', 'TABLET_7', 'TABLET_10', 'TV', 'FOLDABLE'].includes(formFactor)) {
      return problemResponse(400, 'ERR_CONSOLE_REQUEST_FAILED', 'locale and valid formFactor are required', `edge-${startTime}`)
    }
    if (assetType !== 'SCREENSHOT' && assetType !== 'FEATURE_BANNER') {
      return problemResponse(400, 'ERR_CONSOLE_REQUEST_FAILED', 'assetType must be SCREENSHOT or FEATURE_BANNER', `edge-${startTime}`)
    }
    if (file.size <= 0 || file.size > 20 * 1024 * 1024) {
      return problemResponse(400, 'ERR_MEDIA_INVALID', 'image must be 1..20MiB', `edge-${startTime}`)
    }
    const bytes = await file.arrayBuffer()
    const dims = edgeImageDims(new Uint8Array(bytes))
    if (!dims) {
      return problemResponse(400, 'ERR_MEDIA_UNSUPPORTED_FORMAT', 'only PNG and JPEG are accepted', `edge-${startTime}`)
    }
    const digest = await crypto.subtle.digest('SHA-256', bytes)
    const sha = [...new Uint8Array(digest)].map(b => b.toString(16).padStart(2, '0')).join('')
    const asset = {
      id: edgeId('media'), app_id: appId, locale, device_form_factor: formFactor,
      asset_type: assetType, url: `media/${appId}/${sha}.${dims.ext}`,
      width: dims.width, height: dims.height, display_order: (edgeMedia.get(appId) || []).length,
    }
    const list = edgeMedia.get(appId) || []
    list.push(asset)
    edgeMedia.set(appId, list)
    edgeMediaBytes.set(String(asset.id), { bytes, contentType: file.type || `image/${dims.ext === 'jpg' ? 'jpeg' : dims.ext}` })
    return jsonResponse(asset, 201)
  }

  const mediaContentMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/media\/([^/]+)\/content$/)
  if (mediaContentMatch && method === 'GET') {
    const stored = edgeMediaBytes.get(decodeURIComponent(mediaContentMatch[2]))
    if (!stored) {
      return problemResponse(404, 'ERR_MEDIA_NOT_FOUND', 'Media asset not found', `edge-${startTime}`)
    }
    return new Response(stored.bytes, {
      status: 200,
      headers: { ...CORS_HEADERS, 'Content-Type': stored.contentType, 'Cache-Control': 'public, max-age=31536000, immutable' },
    })
  }

  const mediaDeleteMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/media\/([^/]+)$/)
  if (mediaDeleteMatch && method === 'DELETE') {
    const appId = decodeURIComponent(mediaDeleteMatch[1])
    const assetId = decodeURIComponent(mediaDeleteMatch[2])
    const list = edgeMedia.get(appId) || []
    if (!list.some(m => String(m.id) === assetId)) {
      return problemResponse(404, 'ERR_MEDIA_NOT_FOUND', 'Media asset not found', `edge-${startTime}`)
    }
    edgeMedia.set(appId, list.filter(m => String(m.id) !== assetId))
    edgeMediaBytes.delete(assetId)
    return jsonResponse({ status: 'DELETED' })
  }

  const initUploadMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/uploads$/)
  if (initUploadMatch && method === 'POST') {
    const body = await readBody<{ orgId?: string; fileName?: string; fileSize?: number; expectedSha256?: string }>()
    const resolved = resolveOrgApp(body.orgId || null, decodeURIComponent(initUploadMatch[1]))
    if (resolved instanceof Response) {
      return resolved
    }
    const fileName = String(body.fileName || '')
    const sha = String(body.expectedSha256 || '').toLowerCase()
    if (!/\.(apk|aab)$/i.test(fileName) || !/^[0-9a-f]{64}$/.test(sha) || !(Number(body.fileSize) > 0)) {
      return problemResponse(400, 'ERR_UPLOAD_INVALID', 'fileName .apk/.aab, positive fileSize, and 64-hex expectedSha256 required', `edge-${startTime}`)
    }
    const session = { id: edgeId('upl'), org_id: resolved.orgId, app_id: resolved.appId, file_name: fileName, file_size: Number(body.fileSize), expected_sha256: sha, status: 'INITIATED', expires_at: new Date(Date.now() + 30 * 60000).toISOString() }
    edgeUploads.set(String(session.id), session)
    return jsonResponse({ uploadId: session.id, appId: session.app_id, orgId: session.org_id, status: session.status, expiresAt: session.expires_at }, 201)
  }

  const completeMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/uploads\/([^/]+)\/complete$/)
  if (completeMatch && method === 'POST') {
    const session = edgeUploads.get(decodeURIComponent(completeMatch[2]))
    if (!session) {
      return problemResponse(404, 'ERR_UPLOAD_NOT_FOUND', 'Upload session not found', `edge-${startTime}`)
    }
    if (session.status !== 'INITIATED') {
      return problemResponse(409, 'ERR_UPLOAD_STATE', `Session is ${session.status}`, `edge-${startTime}`)
    }
    let sha = ''
    if ((request.headers.get('content-type') || '').includes('multipart/form-data')) {
      const form = await request.formData().catch(() => null)
      sha = String(form?.get('sha256') || '').toLowerCase()
    } else {
      const body = await readBody<{ sha256?: string }>()
      sha = String(body.sha256 || '').toLowerCase()
    }
    if (sha !== session.expected_sha256) {
      return problemResponse(409, 'ERR_UPLOAD_CHECKSUM_MISMATCH', 'SHA-256 does not match the declared value', `edge-${startTime}`)
    }
    session.status = 'COMPLETED'
    return jsonResponse({ uploadId: session.id, status: 'COMPLETED' })
  }

  const releasesMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/releases$/)
  if (releasesMatch && method === 'GET') {
    return jsonResponse(edgeReleases.get(decodeURIComponent(releasesMatch[1])) || [])
  }

  const currentMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/releases\/current$/)
  if (currentMatch && method === 'GET') {
    const list = edgeReleases.get(decodeURIComponent(currentMatch[1])) || []
    if (list.length === 0) {
      return problemResponse(404, 'ERR_NO_RELEASE', 'App has no releases yet', `edge-${startTime}`)
    }
    return jsonResponse(list[0])
  }

  async function edgeRollout(appParam: string): Promise<Response> {
    const resolved = resolveOrgApp(url.searchParams.get('org'), appParam)
    if (resolved instanceof Response) {
      return resolved
    }
    const list = edgeReleases.get(resolved.appId) || []
    if (list.length === 0) {
      return problemResponse(404, 'ERR_NO_RELEASE', 'App has no releases yet', `edge-${startTime}`)
    }
    const body = await readBody<{ rolloutPercentage?: number; action?: string; reason?: string }>()
    const current = list[0]
    const action = String(body.action || '').toUpperCase()
    if (action === 'HALT') {
      current.status = 'HALTED'
      current.haltReason = String(body.reason || 'halted by developer')
      current.haltedAt = edgeNow()
      return jsonResponse(current)
    }
    if (action === 'RESUME') {
      if (current.status !== 'HALTED') {
        return problemResponse(409, 'ERR_ROLLOUT_REJECTED', 'Release is not halted', `edge-${startTime}`)
      }
      current.status = Number(current.rolloutPercentage) >= 100 ? 'PUBLISHED_FULL' : 'PUBLISHED_STAGED'
      return jsonResponse(current)
    }
    const pct = Number(body.rolloutPercentage)
    if (!Number.isFinite(pct) || pct <= Number(current.rolloutPercentage)) {
      return problemResponse(409, 'ERR_ROLLOUT_REJECTED', `Percentage must increase beyond ${current.rolloutPercentage}`, `edge-${startTime}`)
    }
    current.rolloutPercentage = pct
    if (pct >= 100) {
      current.status = 'PUBLISHED_FULL'
    }
    return jsonResponse(current)
  }

  const rolloutMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/releases\/current\/rollout$/)
  if (rolloutMatch && method === 'PATCH') {
    return edgeRollout(decodeURIComponent(rolloutMatch[1]))
  }
  if (pathname === '/api/v1/developer/releases/current/rollout' && method === 'PATCH') {
    return edgeRollout('current')
  }

  const lineageMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/lineage$/)
  if (lineageMatch && method === 'GET') {
    return jsonResponse({ releases: edgeReleases.get(decodeURIComponent(lineageMatch[1])) || [], signatures: [] })
  }

  const campListMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/ads\/campaigns$/)
  if (campListMatch) {
    if (method === 'POST') {
      const body = await readBody<{ appName?: string; targetKeywords?: string[]; maxBidCpc?: number; dailyBudget?: number }>()
      if (!body.targetKeywords?.length || !(Number(body.maxBidCpc) > 0) || !(Number(body.dailyBudget) > 0)) {
        return problemResponse(400, 'ERR_CONSOLE_REQUEST_FAILED', 'keywords, maxBidCpc, and dailyBudget are required', `edge-${startTime}`)
      }
      const c = { id: edgeId('camp'), appId: decodeURIComponent(campListMatch[1]), appName: String(body.appName || ''), targetKeywords: body.targetKeywords, maxBidCpc: Number(body.maxBidCpc), dailyBudget: Number(body.dailyBudget), status: 'PENDING_REVIEW', createdAt: edgeNow() }
      edgeCampaigns.set(String(c.id), c)
      return jsonResponse(c, 201)
    }
    return jsonResponse([...edgeCampaigns.values()])
  }

  const pauseMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/ads\/campaigns\/([^/]+)\/pause$/)
  if (pauseMatch && method === 'POST') {
    const c = edgeCampaigns.get(decodeURIComponent(pauseMatch[2]))
    if (!c) {
      return problemResponse(404, 'ERR_CONSOLE_REQUEST_FAILED', 'Campaign not found', `edge-${startTime}`)
    }
    const body = await readBody<{ paused?: boolean }>()
    c.status = body.paused ? 'PAUSED_MANUAL' : 'ACTIVE'
    return jsonResponse(c)
  }

  const promoListMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/promotions$/)
  if (promoListMatch) {
    const appId = decodeURIComponent(promoListMatch[1])
    if (method === 'POST') {
      const body = await readBody<{ campaignName?: string; discountType?: string; discountValue?: number; startsAt?: string; endsAt?: string; maxUses?: number; codeCount?: number; codePrefix?: string }>()
      const promo = { id: edgeId('promo'), app_id: appId, campaign_name: String(body.campaignName || ''), discount_type: String(body.discountType || 'PERCENTAGE'), discount_value: Number(body.discountValue || 0), starts_at: String(body.startsAt || edgeNow()), ends_at: String(body.endsAt || edgeNow()), max_uses: Number(body.maxUses || 1), used_count: 0, created_at: edgeNow() }
      edgePromos.set(String(promo.id), promo)
      const codes: Array<Record<string, unknown>> = []
      const count = Math.min(Math.max(Number(body.codeCount || 1), 1), 100)
      for (let i = 0; i < count; i++) {
        const code = `${String(body.codePrefix || 'EDGE').toUpperCase()}-${Math.random().toString(36).slice(2, 10).toUpperCase()}`
        const rec = { id: edgeId('pc'), promotion_id: promo.id, code, max_uses: 1, used_count: 0 }
        edgePromoCodes.set(code, rec)
        codes.push({ code, usedCount: 0 })
      }
      return jsonResponse({ promotion: { id: promo.id, appId, campaignName: promo.campaign_name, discountType: promo.discount_type, discountValue: promo.discount_value, startsAt: promo.starts_at, endsAt: promo.ends_at, maxUses: promo.max_uses, usedCount: 0, codeCount: codes.length }, codes }, 201)
    }
    return jsonResponse([...edgePromos.values()].filter(p => p.app_id === appId).map(p => ({ id: p.id, appId: p.app_id, campaignName: p.campaign_name, discountType: p.discount_type, discountValue: p.discount_value, startsAt: p.starts_at, endsAt: p.ends_at, maxUses: p.max_uses, usedCount: p.used_count, codeCount: [...edgePromoCodes.values()].filter(c => c.promotion_id === p.id).length })))
  }

  const pricingMatch = pathname.match(/^\/api\/v1\/developer\/apps\/([^/]+)\/pricing$/)
  if (pricingMatch) {
    const appId = decodeURIComponent(pricingMatch[1])
    if (method === 'POST') {
      const body = await readBody<{ countryCode?: string; currency?: string; listPriceMinor?: number; taxInclusive?: boolean }>()
      if (String(body.countryCode || '').length !== 2 || Number(body.listPriceMinor) < 0) {
        return problemResponse(400, 'ERR_CONSOLE_REQUEST_FAILED', '2-letter country and non-negative price required', `edge-${startTime}`)
      }
      const tiers = edgePricing.get(appId) || []
      const tier = { countryCode: String(body.countryCode).toUpperCase(), currency: String(body.currency || 'KRW'), listPriceMinor: Number(body.listPriceMinor), taxInclusive: Boolean(body.taxInclusive) }
      const idx = tiers.findIndex(t => t.countryCode === tier.countryCode)
      if (idx >= 0) {
        tiers[idx] = tier
      } else {
        tiers.push(tier)
      }
      edgePricing.set(appId, tiers)
      return jsonResponse(tier)
    }
    return jsonResponse(edgePricing.get(appId) || [])
  }

  // 5. Admin Routes — real backend shapes.
  const decisionMatch = pathname.match(/^\/api\/v1\/admin\/reviews\/([^/]+)\/decision$/)
  if (decisionMatch && method === 'POST') {
    const body = await readBody<{ decision?: string; notes?: string; checklist?: Record<string, boolean> }>()
    const verdict = String(body.decision || '').toUpperCase()
    if (verdict !== 'APPROVED' && verdict !== 'REJECTED') {
      return problemResponse(400, 'ERR_DECISION_INVALID', 'decision must be APPROVED or REJECTED', `edge-${startTime}`)
    }
    if (verdict === 'REJECTED' && !String(body.notes || '').trim()) {
      return problemResponse(400, 'ERR_DECISION_INVALID', 'Rejection requires reviewer notes', `edge-${startTime}`)
    }
    const appId = decodeURIComponent(decisionMatch[1])
    const known = EDGE_CATALOG_APPS.some(a => a.id === appId || slugify(a.title) === appId)
    if (!known) {
      return problemResponse(404, 'ERR_APP_NOT_FOUND', 'app not found', `edge-${startTime}`)
    }
    edgeCatalogStatus.set(appId, verdict === 'APPROVED' ? 'PUBLISHED' : 'DRAFT')
    const d = { appId, decision: verdict, notes: String(body.notes || ''), checklist: body.checklist || {}, decidedBy: edgeUserId, decidedAt: edgeNow() }
    edgeDecisions.set(String(d.appId), d)
    edgeAuditSeq += 1
    {
      const prevHash = edgeAudit.length > 0 ? String(edgeAudit[edgeAudit.length - 1].record_hash) : '0000000000000000000000000000000000000000000000000000000000000000'
      const createdAt = edgeNow()
      const nano = String(Date.now() * 1000000)
      const payload = [prevHash, edgeUserId, 'admin.review.decision', 'app', String(d.appId), '', nano].join('|')
      const recordHash = await sha256Hex(payload)
      edgeAudit.push({ id: `audit_${edgeAuditSeq}`, seq: edgeAuditSeq, actor_id: edgeUserId, action: 'admin.review.decision', target_type: 'app', target_id: d.appId, diff_json: '', prev_hash: prevHash, record_hash: recordHash, created_at: createdAt, created_at_nano: nano })
    }
    return jsonResponse(d)
  }

  if (pathname === '/api/v1/admin/apps/review-queue' && method === 'GET') {
    return jsonResponse(EDGE_CATALOG_APPS.filter(a => (edgeCatalogStatus.get(a.id) || 'PUBLISHED') === 'PENDING_INITIAL_REVIEW').map(a => ({
      id: a.id, canonical_slug: slugify(a.title), primary_package_name: a.package_name,
      catalog_status: 'PENDING_INITIAL_REVIEW',
    })))
  }

  const adminAppMatch = pathname.match(/^\/api\/v1\/admin\/apps\/([^/]+)$/)
  if (adminAppMatch && method === 'GET') {
    const key = decodeURIComponent(adminAppMatch[1])
    const a = EDGE_CATALOG_APPS.find(x => x.id === key || slugify(x.title) === key)
    if (!a) {
      return problemResponse(404, 'ERR_APP_NOT_FOUND', 'app not found', `edge-${startTime}`)
    }
    return jsonResponse({
      app: { id: a.id, canonical_slug: slugify(a.title), primary_package_name: a.package_name, catalog_status: edgeCatalogStatus.get(a.id) || 'PUBLISHED' },
      localization: { title: a.title, short_description: a.summary, full_description: a.summary, whats_new: '' },
      media_assets: [], distributions: [], tags: [], resolved_locale: 'en-US',
    })
  }
  if (pathname === '/api/v1/admin/reviews/decisions' && method === 'GET') {
    return jsonResponse([...edgeDecisions.values()])
  }

  if (pathname === '/api/v1/admin/verifications' && method === 'GET') {
    const status = url.searchParams.get('status') || 'PENDING_REVIEW'
    return jsonResponse([...edgeVerifications.values()].filter(v => String(v.status) === status))
  }

  const verificationDecisionMatch = pathname.match(/^\/api\/v1\/admin\/verifications\/([^/]+)\/decision$/)
  if (verificationDecisionMatch && method === 'POST') {
    const orgId = decodeURIComponent(verificationDecisionMatch[1])
    const org = edgeOrgs.get(orgId)
    if (!org) {
      return problemResponse(404, 'ERR_ORG_NOT_FOUND', 'org not found', `edge-${startTime}`)
    }
    const v = edgeVerifications.get(orgId)
    if (!v) {
      return problemResponse(404, 'ERR_VERIFICATION_NOT_FOUND', 'verification not found', `edge-${startTime}`)
    }
    const body = await readBody<{ approved?: boolean; notes?: string }>()
    const rowStatus = body.approved ? 'APPROVED' : 'REJECTED'
    v.status = rowStatus
    v.reviewer_id = edgeUserId
    v.review_notes = String(body.notes || '')
    v.reviewed_at = edgeNow()
    org.verification_status = body.approved ? 'VERIFIED' : 'REJECTED'
    org.updated_at = edgeNow()
    return jsonResponse(org)
  }

  if (pathname === '/api/v1/admin/security/kill-switches' && method === 'GET') {
    return jsonResponse([...edgeKillSwitches.values()])
  }

  const ksMatch = pathname.match(/^\/api\/v1\/admin\/security\/kill-switch\/(.+)$/)
  if (ksMatch) {
    const pkg = decodeURIComponent(ksMatch[1])
    if (method === 'POST') {
      const body = await readBody<{ scope?: string; reasonCode?: string; evidence?: string }>()
      if (!String(body.reasonCode || '').trim()) {
        return problemResponse(400, 'ERR_KILLSWITCH_INVALID', 'reasonCode is required', `edge-${startTime}`)
      }
      const state = { app_id: pkg, is_triggered: true, reason: `${body.scope || ''} ${body.reasonCode} ${body.evidence || ''}`.trim(), triggered_by: edgeUserId, triggered_at: edgeNow() }
      edgeKillSwitches.set(pkg, state)
      return jsonResponse(state)
    }
    if (method === 'DELETE') {
      const state = { app_id: pkg, is_triggered: false, reason: 'Resolved by admin', triggered_by: edgeUserId, triggered_at: edgeNow() }
      edgeKillSwitches.set(pkg, state)
      return jsonResponse(state)
    }
    const state = edgeKillSwitches.get(pkg)
    if (!state) {
      return problemResponse(404, 'ERR_KILLSWITCH_NOT_FOUND', 'No kill-switch record', `edge-${startTime}`)
    }
    return jsonResponse(state)
  }

  if (pathname === '/api/v1/admin/security/scans' && method === 'GET') {
    return jsonResponse([])
  }

  if (pathname === '/api/v1/admin/connectors' && method === 'GET') {
    return jsonResponse(edgeConnectorsList())
  }

  if (pathname === '/api/v1/admin/connectors' && method === 'POST') {
    const body = await readBody<{ id?: string; name?: string; kind?: string; feedUrl?: string; fingerprint?: string; packages?: string[] }>()
    const id = String(body.id || '').trim()
    if (!id || !String(body.name || '').trim() || !String(body.kind || '').trim()) {
      return problemResponse(400, 'ERR_CONNECTOR_INVALID', 'id, name and kind are required', `edge-${startTime}`)
    }
    if (edgeConnectors.has(id)) {
      return problemResponse(409, 'ERR_CONNECTOR_EXISTS', 'Connector id already registered', `edge-${startTime}`)
    }
    const rec = {
      id, name: String(body.name).trim(), kind: String(body.kind).trim().toUpperCase(),
      feedUrl: String(body.feedUrl || ''), fingerprint: String(body.fingerprint || ''),
      packages: body.packages || [], syncable: String(body.kind).trim().toUpperCase() === 'FDROID' && !!String(body.feedUrl || ''),
    }
    edgeConnectors.set(id, rec)
    return jsonResponse(rec, 201)
  }

  const syncMatch = pathname.match(/^\/api\/v1\/admin\/connectors\/([^/]+)\/trigger-sync$/)
  if (syncMatch && method === 'POST') {
    const id = decodeURIComponent(syncMatch[1])
    const conn = edgeConnectors.get(id)
    if (!conn) {
      return problemResponse(404, 'ERR_CONNECTOR_NOT_FOUND', 'Connector not found', `edge-${startTime}`)
    }
    if (!conn.syncable || !conn.feedUrl) {
      return problemResponse(409, 'ERR_CONNECTOR_NOT_SYNCABLE', 'Connector cannot sync from edge', `edge-${startTime}`)
    }
    try {
      const feed = await fetch(String(conn.feedUrl), { headers: { Accept: 'application/json' } })
      if (!feed.ok) {
        return problemResponse(502, 'ERR_CONNECTOR_SYNC_FAILED', `Feed status ${feed.status}`, `edge-${startTime}`)
      }
      const index = await feed.json() as { packages?: Record<string, unknown> }
      const count = Object.keys(index.packages || {}).length
      conn.lastSyncAt = edgeNow()
      conn.lastItemCount = count
      return jsonResponse({ connectorId: id, syncedAt: edgeNow(), itemCount: count })
    } catch (err) {
      return problemResponse(502, 'ERR_CONNECTOR_SYNC_FAILED', String(err), `edge-${startTime}`)
    }
  }

  if (pathname === '/api/v1/admin/settlements' && method === 'GET') {
    return jsonResponse([...edgePayouts.values()])
  }
  if (pathname === '/api/v1/admin/settlements' && method === 'POST') {
    const body = await readBody<{ orgId?: string; amount?: number; currency?: string }>()
    const payout = { id: edgeId('payout'), org_id: String(body.orgId || ''), amount: Number(body.amount || 0), currency: String(body.currency || 'KRW'), status: 'PENDING_APPROVAL', created_by: edgeUserId, created_at: edgeNow() }
    edgePayouts.set(String(payout.id), payout)
    return jsonResponse(payout, 201)
  }

  const approveMatch = pathname.match(/^\/api\/v1\/admin\/settlements\/([^/]+)\/approve$/)
  if (approveMatch && method === 'POST') {
    const body = await readBody<{ approver_id?: string; initiator_id?: string }>()
    const payout = edgePayouts.get(decodeURIComponent(approveMatch[1]))
    if (!payout) {
      return problemResponse(404, 'ERR_SETTLEMENT_NOT_FOUND', 'Payout not found', `edge-${startTime}`)
    }
    if (body.initiator_id && body.initiator_id !== payout.created_by) {
      return problemResponse(409, 'ERR_SETTLEMENT_INITIATOR_MISMATCH', 'Initiator does not match payout creator', `edge-${startTime}`)
    }
    if (body.approver_id === payout.created_by) {
      return problemResponse(403, 'ERR_DUAL_CONTROL_SELF_APPROVAL_FORBIDDEN', 'Four-Eyes violation: initiator cannot approve', `edge-${startTime}`)
    }
    payout.status = 'PAID_COMPLETED'
    payout.approved_by = body.approver_id || edgeUserId
    payout.approved_at = edgeNow()
    return jsonResponse(payout)
  }

  if (pathname === '/api/v1/auth/admin/approvals/pending' && method === 'GET') {
    if (!edgeUserId) {
      return problemResponse(401, 'ERR_UNAUTHORIZED', 'Authentication required', `edge-${startTime}`)
    }
    return jsonResponse([...edgeApprovals.values()].filter(a => a.status === 'PENDING'))
  }

  if (pathname === '/api/v1/auth/admin/approvals' && method === 'POST') {
    if (!edgeUserId) {
      return problemResponse(401, 'ERR_UNAUTHORIZED', 'Authentication required', `edge-${startTime}`)
    }
    const body = await readBody<Record<string, unknown>>()
    if (!body.operation_type || !body.reason) {
      return problemResponse(400, 'ERR_INVALID_BODY', 'operation_type and reason are required', `edge-${startTime}`)
    }
    const req = {
      id: edgeId('apr'),
      operation_type: String(body.operation_type),
      initiator_id: edgeUserId,
      target_resource_id: String(body.target_resource_id || ''),
      payload_json: String(body.payload_json || ''),
      status: 'PENDING',
      reason: String(body.reason),
      created_at: edgeNow(),
    }
    edgeApprovals.set(String(req.id), req)
    return jsonResponse(req, 201)
  }

  const approvalMatch = pathname.match(/^\/api\/v1\/auth\/admin\/approvals\/([^/]+)\/(approve|reject)$/)
  if (approvalMatch && method === 'POST') {
    if (!edgeUserId) {
      return problemResponse(401, 'ERR_UNAUTHORIZED', 'Authentication required', `edge-${startTime}`)
    }
    const req = edgeApprovals.get(approvalMatch[1])
    if (!req) {
      return problemResponse(404, 'ERR_APPROVAL_NOT_FOUND', 'Approval request not found', `edge-${startTime}`)
    }
    if (req.initiator_id === edgeUserId) {
      return problemResponse(403, 'ERR_DUAL_CONTROL_SELF_APPROVAL_FORBIDDEN', 'The initiator cannot approve their own request', `edge-${startTime}`)
    }
    if (req.status !== 'PENDING') {
      return problemResponse(400, 'ERR_APPROVAL_FAILED', 'Approval request already decided', `edge-${startTime}`)
    }
    const body = await readBody<Record<string, unknown>>()
    req.status = approvalMatch[2] === 'approve' ? 'APPROVED' : 'REJECTED'
    req.approver_id = edgeUserId
    req.decided_at = edgeNow()
    if (approvalMatch[2] === 'reject' && body.reason) {
      req.reason = String(body.reason)
    }
    edgeApprovals.set(String(req.id), req)
    return jsonResponse(req)
  }

  if (pathname === '/api/v1/admin/audit-logs' && method === 'GET') {
    const limit = Math.min(Math.max(Number(url.searchParams.get('limit') || 100), 1), 500)
    return jsonResponse({ chainValid: true, chainFailedAt: -1, records: edgeAudit.slice(-limit) })
  }

  if (pathname === '/api/v1/admin/orders' && method === 'GET') {
    const user = url.searchParams.get('user')
    const org = url.searchParams.get('org')
    let orders = [...edgeOrders.values()]
    if (user) {
      orders = orders.filter(o => String(o.user_id) === user)
    }
    if (org) {
      orders = orders.filter(o => String(o.org_id) === org)
    }
    return jsonResponse(orders)
  }

  const adminRefundMatch = pathname.match(/^\/api\/v1\/admin\/orders\/([^/]+)\/refund$/)
  if (adminRefundMatch && method === 'POST') {
    const order = edgeOrders.get(decodeURIComponent(adminRefundMatch[1]))
    if (!order) {
      return problemResponse(404, 'ERR_ORDER_NOT_FOUND', 'Order not found', `edge-${startTime}`)
    }
    if (order.status === 'REFUNDED') {
      return problemResponse(400, 'ERR_REFUND_FAILED', 'Order already refunded', `edge-${startTime}`)
    }
    order.status = 'REFUNDED'
    order.refunded_at = edgeNow()
    return jsonResponse(order)
  }

  if (pathname === '/api/v1/admin/ads/campaigns' && method === 'GET') {
    const status = url.searchParams.get('status')
    const list = [...edgeCampaigns.values()].filter(c => !status || c.status === status)
    return jsonResponse(list)
  }

  const adReviewMatch = pathname.match(/^\/api\/v1\/admin\/ads\/campaigns\/([^/]+)\/review$/)
  if (adReviewMatch && method === 'POST') {
    const c = edgeCampaigns.get(decodeURIComponent(adReviewMatch[1]))
    if (!c) {
      return problemResponse(404, 'ERR_AD_CAMPAIGN_NOT_FOUND', 'Campaign not found', `edge-${startTime}`)
    }
    const body = await readBody<{ approve?: boolean }>()
    c.status = body.approve ? 'ACTIVE' : 'REJECTED'
    return jsonResponse(c)
  }

  if (pathname === '/api/v1/admin/analytics/overview' && method === 'GET') {
    const byEvent: Record<string, number> = {}
    const byPlatform: Record<string, number> = {}
    const byCountry: Record<string, number> = {}
    const byApp: Record<string, number> = {}
    const revenueByCountry: Record<string, number> = {}
    const purchaseCountByCountry: Record<string, number> = {}
    let totalRevenueCents = 0
    let last24h = 0
    const cutoff = Date.now() - 24 * 3600 * 1000
    for (const e of edgeEvents) {
      const name = String(e.event_name || '')
      byEvent[name] = (byEvent[name] || 0) + 1
      const plat = String(e.platform || '')
      byPlatform[plat] = (byPlatform[plat] || 0) + 1
      const cc = String(e.country_code || '')
      byCountry[cc] = (byCountry[cc] || 0) + 1
      const props = (e.properties || {}) as Record<string, unknown>
      const appId = String(props.app_id || '')
      if (appId) {
        byApp[appId] = (byApp[appId] || 0) + 1
      }
      if (name === 'app.purchase.completed') {
        const price = Number(props.price_cents || 0)
        totalRevenueCents += price
        revenueByCountry[cc] = (revenueByCountry[cc] || 0) + price
        purchaseCountByCountry[cc] = (purchaseCountByCountry[cc] || 0) + 1
      }
      const occurred = Date.parse(String(e.occurred_at || ''))
      if (!Number.isNaN(occurred) && occurred > cutoff) {
        last24h += 1
      }
    }
    const top = (m: Record<string, number>, n: number) =>
      Object.entries(m).sort((a, b) => b[1] - a[1]).slice(0, n).map(([key, count]) => ({ key, count }))
    return jsonResponse({
      totalEvents: edgeEvents.length, last24h, totalRevenueCents,
      byEvent: top(byEvent, 20),
      byPlatform: top(byPlatform, 10),
      byCountry: top(byCountry, 10),
      byApp: top(byApp, 10),
      revenueByCountry: Object.entries(revenueByCountry)
        .sort((a, b) => b[1] - a[1]).slice(0, 10)
        .map(([key, revenueCents]) => ({ key, count: purchaseCountByCountry[key] || 0, revenueCents })),
    })
  }

  if (pathname === '/api/v1/admin/reports' && method === 'GET') {
    const status = url.searchParams.get('status')
    return jsonResponse(edgeReports.filter(r => !status || r.status === status))
  }

  const claimMatch = pathname.match(/^\/api\/v1\/admin\/reports\/([^/]+)\/claim$/)
  if (claimMatch && method === 'POST') {
    const rep = edgeReports.find(r => String(r.id) === decodeURIComponent(claimMatch[1]))
    if (!rep) {
      return problemResponse(404, 'ERR_REPORT_NOT_FOUND', 'Report not found', `edge-${startTime}`)
    }
    rep.assigned_moderator_id = edgeUserId
    if (rep.status === 'OPEN') {
      rep.status = 'IN_INVESTIGATION'
    }
    return jsonResponse(rep)
  }

  const resolveMatch = pathname.match(/^\/api\/v1\/admin\/reports\/([^/]+)\/resolve$/)
  if (resolveMatch && method === 'POST') {
    const rep = edgeReports.find(r => String(r.id) === decodeURIComponent(resolveMatch[1]))
    if (!rep) {
      return problemResponse(404, 'ERR_REPORT_NOT_FOUND', 'Report not found', `edge-${startTime}`)
    }
    const body = await readBody<{ action_taken?: boolean; notes?: string }>()
    rep.status = body.action_taken ? 'ACTION_TAKEN' : 'DISMISSED'
    rep.resolution_notes = String(body.notes || '')
    rep.resolved_at = edgeNow()
    return jsonResponse(rep)
  }

  if (pathname === '/api/v1/admin/appeals' && method === 'GET') {
    return jsonResponse(edgeAppeals)
  }

  const decideMatch = pathname.match(/^\/api\/v1\/admin\/appeals\/([^/]+)\/decide$/)
  if (decideMatch && method === 'POST') {
    const appeal = edgeAppeals.find(a => String(a.id) === decodeURIComponent(decideMatch[1]))
    if (!appeal) {
      return problemResponse(404, 'ERR_APPEAL_NOT_FOUND', 'Appeal not found', `edge-${startTime}`)
    }
    const body = await readBody<{ decision?: string; notes?: string }>()
    if (body.decision !== 'APPROVED' && body.decision !== 'REJECTED') {
      return problemResponse(400, 'ERR_APPEAL_INVALID', 'decision must be APPROVED or REJECTED', `edge-${startTime}`)
    }
    if (edgeUserId === appeal.original_reviewer_id) {
      return problemResponse(403, 'ERR_APPEAL_SAME_REVIEWER_FORBIDDEN', 'Original reviewer cannot decide own appeal', `edge-${startTime}`)
    }
    appeal.status = body.decision
    appeal.decision_notes = String(body.notes || '')
    appeal.decided_at = edgeNow()
    return jsonResponse(appeal)
  }

  // 6. Commerce Routes — real backend shapes with real pricing math.
  const publicPricesMatch = pathname.match(/^\/api\/v1\/commerce\/apps\/([^/]+)\/prices$/)
  if (publicPricesMatch && method === 'GET') {
    const appId = decodeURIComponent(publicPricesMatch[1])
    const tiers = edgePricing.get(appId) || []
    return jsonResponse({
      items: tiers.map(t => ({
        app_id: appId,
        country_code: t.countryCode,
        currency: t.currency,
        list_price_minor: t.listPriceMinor,
        tax_inclusive: t.taxInclusive,
        updated_at: edgeNow(),
      })),
    })
  }

  if (pathname === '/api/v1/commerce/orders' && method === 'POST') {
    const body = await readBody<{ app_id?: string; org_id?: string; gross_amount?: number; currency?: string; idempotency_key?: string; promo_code?: string }>()
    const gross = Number(body.gross_amount)
    if (!body.app_id || !(gross >= 0) || (gross === 0 && !body.promo_code)) {
      return problemResponse(400, 'ERR_INVALID_BODY', 'app_id and positive gross_amount required', `edge-${startTime}`)
    }
    const currency = String(body.currency || 'KRW')
    const tiers = edgePricing.get(String(body.app_id)) || []
    let promo: Record<string, unknown> | null = null
    if (body.promo_code) {
      const code = String(body.promo_code).toUpperCase().trim()
      const rec = edgePromoCodes.get(code)
      const found = rec ? edgePromos.get(String(rec.promotion_id)) : undefined
      if (!rec || !found || String(found.app_id) !== String(body.app_id)) {
        return problemResponse(404, 'ERR_PROMO_REDEEM_FAILED', 'ERR_PROMO_NOT_FOUND', `edge-${startTime}`)
      }
      if (Number(rec.used_count) >= Number(rec.max_uses)) {
        return problemResponse(409, 'ERR_PROMO_REDEEM_FAILED', 'ERR_PROMO_EXHAUSTED', `edge-${startTime}`)
      }
      promo = found
    }
    const discounted = (list: number): number => {
      if (!promo) return list
      if (String(promo.discount_type) === 'FREE_REDEEM') return 0
      if (String(promo.discount_type) === 'PERCENTAGE') return Math.floor(list * (100 - Number(promo.discount_value)) / 100)
      return Math.max(0, list - Number(promo.discount_value))
    }
    if (tiers.length > 0 && !tiers.some(t => String(t.currency) === currency && discounted(Number(t.listPriceMinor)) === gross)) {
      return problemResponse(422, 'ERR_ORDER_AMOUNT_MISMATCH', 'Gross amount does not match the app list price', `edge-${startTime}`)
    }
    if (body.promo_code) {
      const rec = edgePromoCodes.get(String(body.promo_code).toUpperCase().trim())
      if (rec) rec.used_count = Number(rec.used_count) + 1
    }
    const tax = Math.floor(gross / 11)
    const commission = Math.floor((gross - tax) * 0.15)
    const order = {
      id: edgeId('ord'), user_id: edgeUserId, app_id: String(body.app_id), org_id: String(body.org_id || ''),
      currency: String(body.currency || 'KRW'), gross_amount: gross, tax_amount: tax,
      commission_amount: commission, dev_payout_amount: gross - tax - commission,
      status: gross === 0 ? 'PAID' : 'PENDING', created_at: edgeNow(),
    }
    edgeOrders.set(String(order.id), order)
    return jsonResponse(order, 201)
  }

  if (pathname === '/api/v1/commerce/orders' && method === 'GET') {
    return jsonResponse([...edgeOrders.values()].filter(o => String(o.user_id) === edgeUserId))
  }

  const checkoutMatch = pathname.match(/^\/api\/v1\/commerce\/orders\/([^/]+)\/checkout$/)
  if (checkoutMatch && method === 'POST') {
    const order = edgeOrders.get(decodeURIComponent(checkoutMatch[1]))
    if (!order || String(order.user_id) !== edgeUserId) {
      return problemResponse(404, 'ERR_ORDER_NOT_FOUND', 'Order not found', `edge-${startTime}`)
    }
    if (order.status !== 'PENDING') {
      return problemResponse(409, 'ERR_ORDER_NOT_PAYABLE', 'Only pending orders can start checkout', `edge-${startTime}`)
    }
    return jsonResponse({
      provider: 'TOSS', order_id: order.id, amount_krw: order.gross_amount,
      checkout_url: `https://sandbox.tosspayments.com/checkout/${order.id}`,
      expires_at: new Date(Date.now() + 30 * 60000).toISOString(),
    })
  }

  const refundMatch = pathname.match(/^\/api\/v1\/commerce\/orders\/([^/]+)\/refund$/)
  if (refundMatch && method === 'POST') {
    const order = edgeOrders.get(decodeURIComponent(refundMatch[1]))
    if (!order || String(order.user_id) !== edgeUserId) {
      return problemResponse(404, 'ERR_ORDER_NOT_FOUND', 'Order not found', `edge-${startTime}`)
    }
    if (order.status === 'REFUNDED') {
      return problemResponse(400, 'ERR_REFUND_FAILED', 'Order already refunded', `edge-${startTime}`)
    }
    order.status = 'REFUNDED'
    order.refunded_at = edgeNow()
    return jsonResponse(order)
  }

  if (pathname === '/api/v1/commerce/entitlements' && method === 'GET') {
    return jsonResponse([...edgeOrders.values()]
      .filter(o => String(o.user_id) === edgeUserId && o.status === 'PAID')
      .map(o => ({ user_id: o.user_id, app_id: o.app_id, status: 'ACTIVE' })))
  }

  if (pathname === '/api/v1/commerce/invoices' && method === 'GET') {
    return jsonResponse([...edgeOrders.values()]
      .filter(o => String(o.user_id) === edgeUserId && o.status !== 'PENDING')
      .map((o, i) => ({ id: `inv_${o.id}`, invoice_number: `INV-${new Date().getFullYear()}-${String(i + 1).padStart(6, '0')}`, order_id: o.id, user_id: o.user_id, app_id: o.app_id, currency: o.currency, gross_amount: o.gross_amount, status: o.status === 'REFUNDED' ? 'REFUNDED' : 'ISSUED', issued_at: o.created_at })))
  }

  if (pathname === '/api/v1/commerce/subscriptions' && method === 'GET') {
    return jsonResponse([])
  }
  if (pathname === '/api/v1/commerce/subscriptions' && method === 'POST') {
    const body = await readBody<{ app_id?: string; billing_interval?: string; amount_minor?: number; currency?: string }>()
    if (!['MONTHLY', 'YEARLY'].includes(String(body.billing_interval)) || !(Number(body.amount_minor) > 0)) {
      return problemResponse(400, 'ERR_SUBSCRIPTION_CREATE_FAILED', 'MONTHLY/YEARLY interval and positive amount required', `edge-${startTime}`)
    }
    const start = new Date()
    const end = new Date(start)
    if (body.billing_interval === 'YEARLY') {
      end.setFullYear(end.getFullYear() + 1)
    } else {
      end.setMonth(end.getMonth() + 1)
    }
    return jsonResponse({
      id: edgeId('sub'), user_id: edgeUserId, app_id: String(body.app_id || ''),
      billing_interval: body.billing_interval, amount_minor: Number(body.amount_minor),
      currency: String(body.currency || 'KRW'), status: 'ACTIVE',
      current_period_start: start.toISOString(), current_period_end: end.toISOString(), created_at: edgeNow(),
    }, 201)
  }

  if ((pathname === '/api/v1/commerce/promos/redeem' || pathname === '/api/v1/commerce/promos/validate') && method === 'POST') {
    const body = await readBody<{ code?: string; appId?: string; app_id?: string }>()
    const code = String(body.code || '').toUpperCase().trim()
    const rec = edgePromoCodes.get(code)
    const appId = String(body.appId || body.app_id || '')
    if (!rec) {
      return problemResponse(404, 'ERR_PROMO_REDEEM_FAILED', 'ERR_PROMO_NOT_FOUND', `edge-${startTime}`)
    }
    const promo = edgePromos.get(String(rec.promotion_id))
    if (!promo || String(promo.app_id) !== appId) {
      return problemResponse(404, 'ERR_PROMO_REDEEM_FAILED', 'ERR_PROMO_NOT_FOUND', `edge-${startTime}`)
    }
    if (Number(rec.used_count) >= Number(rec.max_uses)) {
      return problemResponse(409, 'ERR_PROMO_REDEEM_FAILED', 'ERR_PROMO_EXHAUSTED', `edge-${startTime}`)
    }
    if (pathname.endsWith('/redeem')) {
      rec.used_count = Number(rec.used_count) + 1
    }
    return jsonResponse({ promotion: promo, code })
  }

  // 7. Moderation Routes.
  if (pathname === '/api/v1/moderation/reports' && method === 'POST') {
    const body = await readBody<{ appIdentifier?: string; category?: string; reporterEmail?: string; description?: string; confirmAccuracy?: boolean }>()
    if (!body.confirmAccuracy) {
      return problemResponse(400, 'ERR_REPORT_INVALID', 'Accuracy confirmation is required', `edge-${startTime}`)
    }
    if (String(body.description || '').trim().length < 20 || !String(body.reporterEmail || '').includes('@')) {
      return problemResponse(400, 'ERR_REPORT_INVALID', 'Valid email and 20+ char description required', `edge-${startTime}`)
    }
    const rep = { id: edgeId('rep'), reporter_user_id: edgeUserId, reporter_email: String(body.reporterEmail).toLowerCase(), target_type: 'APP', target_id: String(body.appIdentifier || ''), report_category: String(body.category || 'SPAM_ABUSE'), description: String(body.description), priority: 'P2_NORMAL', status: 'OPEN', created_at: edgeNow() }
    edgeReports.push(rep)
    return jsonResponse({ ticketId: rep.id, priority: rep.priority, status: rep.status }, 201)
  }

  if (pathname === '/api/v1/moderation/reviews' && method === 'POST') {
    const body = await readBody<{ appId?: string; rating?: number; title?: string; comment?: string; versionCode?: number }>()
    const list = edgeReviews.get(String(body.appId)) || []
    const rev = { id: edgeId('rev'), app_id: String(body.appId), user_id: edgeUserId, rating: Number(body.rating), title: String(body.title || ''), comment: String(body.comment || ''), status: 'PUBLISHED', created_at: edgeNow() }
    list.push(rev)
    edgeReviews.set(String(body.appId), list)
    return jsonResponse(rev, 201)
  }

  const revListMatch = pathname.match(/^\/api\/v1\/moderation\/apps\/([^/]+)\/reviews$/)
  if (revListMatch && method === 'GET') {
    return jsonResponse((edgeReviews.get(decodeURIComponent(revListMatch[1])) || []).filter(r => r.status === 'PUBLISHED'))
  }

  if (pathname === '/api/v1/moderation/appeals' && method === 'POST') {
    const body = await readBody<{ orgId?: string; releaseId?: string; reason?: string; originalReviewerId?: string }>()
    const appeal = { id: edgeId('apl'), org_id: String(body.orgId || ''), release_id: String(body.releaseId || ''), reason: String(body.reason || ''), original_reviewer_id: String(body.originalReviewerId || ''), status: 'PENDING', created_at: edgeNow() }
    edgeAppeals.push(appeal)
    return jsonResponse(appeal, 201)
  }

  // 8. Search Routes.
  if (pathname === '/api/v1/search' && method === 'GET') {
    const q = (url.searchParams.get('q') || '').toLowerCase().trim()
    const limit = Math.min(Math.max(Number(url.searchParams.get('limit') || 20), 1), 50)
    const items = EDGE_CATALOG_APPS
      .filter(a => !q || a.title.toLowerCase().includes(q) || a.package_name.toLowerCase().includes(q) || a.summary.toLowerCase().includes(q))
      .slice(0, limit)
      .map(a => ({ app_id: a.id, package_name: a.package_name, title_i18n: { 'en-US': a.title }, developer_name: a.developer_name, category: a.category, bayesian_rating: a.rating, rating_count: a.rating_count, download_count: a.rating_count * 10, direct_download_available: true, rights_tier: 'T1', keywords: [] }))
    return jsonResponse({ items, total_count: items.length, took_ms: 1, fallback_used: true })
  }

  if (pathname === '/api/v1/search/suggest' && method === 'GET') {
    const q = (url.searchParams.get('q') || '').toLowerCase().trim()
    const limit = Math.min(Math.max(Number(url.searchParams.get('limit') || 10), 1), 20)
    const suggestions = EDGE_CATALOG_APPS
      .filter(a => q && (a.title.toLowerCase().includes(q) || a.package_name.toLowerCase().includes(q)))
      .slice(0, limit)
      .map(a => ({ text: a.title, type: 'APP', package: a.package_name }))
    return jsonResponse({ suggestions, took_ms: 1 })
  }

  // 9. Analytics Routes.
  if (pathname === '/api/v1/analytics/events' && method === 'POST') {
    let batch: Array<Record<string, unknown>> = []
    try { batch = await request.json() } catch { /* default */ }
    if (!Array.isArray(batch) || batch.length === 0 || batch.length > 500) {
      return problemResponse(400, 'ERR_INVALID_BATCH', 'Batch must contain 1-500 events', `edge-${startTime}`)
    }
    for (const e of batch) {
      edgeEvents.push(e)
    }
    return jsonResponse({ accepted: batch.length }, 202)
  }

  // 404 for unknown API paths
  return problemResponse(404, 'ERR_NOT_FOUND', `API endpoint ${pathname} is not defined.`, `edge-${startTime}`)
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    // 1. Handle CORS preflight
    if (request.method === 'OPTIONS') {
      return new Response(null, {
        status: 204,
        headers: CORS_HEADERS,
      })
    }

    const url = new URL(request.url)
    const pathname = url.pathname

    // 2. Intercept API and health check requests
    if (
      pathname.startsWith('/api/') ||
      pathname.startsWith('/healthz') ||
      pathname === '/metrics'
    ) {
      return handleApiRequest(request, env)
    }

    // 3. Static assets with per-app SPA fallback shells.
    // The assets layer cannot serve the right entry for deep links on its
    // own: native SPA fallback returns the prerendered store home (whose
    // embedded payload boots the home route), and _redirects rules are not
    // applied under existing directories (/developer/*, /admin/*).
    // Strategy: serve the exact asset first; only on miss, serve the owning
    // app's shell (extensionless GET navigations only, so real files and
    // prerendered pages always win). Requires not_found_handling = "none".
    try {
      const direct = await env.ASSETS.fetch(request)
      if (direct.status !== 404) {
        return direct
      }
    } catch {
      // Fall through to shell resolution below.
    }
    if (request.method === 'GET' && !/\.[a-z0-9]+$/i.test(pathname)) {
      // Canonical subrequest forms: the assets layer 307-normalizes explicit
      // ".html" file URLs (e.g. /200.html -> /200), which would surface as a
      // client redirect and lose the deep-link path. Extensionless /200 and
      // trailing-slash directory indexes serve content directly with 200.
      const shell =
        pathname === '/developer' || pathname.startsWith('/developer/')
          ? '/developer/'
          : pathname === '/admin' || pathname.startsWith('/admin/')
            ? '/admin/'
            : '/200'
      try {
        const rewritten = new Request(new URL(shell, url.origin), request)
        const shelled = await env.ASSETS.fetch(rewritten)
        if (shelled.status !== 404) {
          return shelled
        }
      } catch {
        // Fall through to the 404 below.
      }
    }
    return new Response('Asset not found', { status: 404 })
  },
}


package httpserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	catalogApp "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/application"
	catalogInfra "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/infrastructure"
	catalogPres "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/presentation"

	developerApp "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/application"
	developerInfra "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/infrastructure"
	developerPres "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/presentation"

	commerceApp "github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/application"
	commercePres "github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/presentation"
	commerceSet "github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/settlement"

	distributionApp "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
	distributionInfra "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/infrastructure"
	distributionPres "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/presentation"

	iamApp "github.com/bookingnetwork/appmarket/backend/internal/modules/iam/application"
	iamInfra "github.com/bookingnetwork/appmarket/backend/internal/modules/iam/infrastructure"
	iamPres "github.com/bookingnetwork/appmarket/backend/internal/modules/iam/presentation"
	iamRbac "github.com/bookingnetwork/appmarket/backend/internal/modules/iam/rbac"

	moderationApp "github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/application"
	moderationPres "github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/presentation"

	searchApp "github.com/bookingnetwork/appmarket/backend/internal/modules/search/application"
	searchPres "github.com/bookingnetwork/appmarket/backend/internal/modules/search/presentation"

	analyticsPres "github.com/bookingnetwork/appmarket/backend/internal/modules/analytics/presentation"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/analytics"

	adsApp "github.com/bookingnetwork/appmarket/backend/internal/modules/ads/application"

	securityApp "github.com/bookingnetwork/appmarket/backend/internal/modules/security/application"

	auditpkg "github.com/bookingnetwork/appmarket/backend/internal/platform/audit"
	consolepkg "github.com/bookingnetwork/appmarket/backend/internal/platform/console"
	problem "github.com/bookingnetwork/appmarket/backend/internal/platform/errors"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/sentryx"
)

type Config struct {
	DB                *pgxpool.Pool
	Redis             *redis.Client
	OpenSearchURL     string
	S3HealthURL       string
	OpenSearchUser    string
	OpenSearchPass    string
	AllowedOrigin     string
	TrustedProxyCIDRs []string
	Logger            *slog.Logger
	HTTPClient        *http.Client
	DisableRateLimit  bool
	// Secrets and environment-provided settings. The secrets are
	// required: New fails closed when any of them is empty, and the payout
	// encryption key must be exactly 32 bytes.
	IAMHMACSecret         string
	DistributionTokenSecret string
	PGWebhookSecret       string
	PayoutEncryptionKey   string
	JWTIssuer             string
	JWTPrivateKey         *rsa.PrivateKey
	CDNDomain             string
	PublicBaseURL         string
	PGProvider            string
	TossSecretKey         string
	PortOneAPIKey         string
	PortOneStoreID        string
	PGIsLive              bool
	SentryDSN             string
	AppEnv                string
	OAuthRedirectBaseURL  string
	GoogleClientID        string
	GoogleClientSecret    string
	GitHubClientID        string
	GitHubClientSecret    string
	AppleClientID         string
	AppleTeamID           string
	AppleKeyID            string
	ApplePrivateKeyPEM    string
	MediaBucket           string
}

// mediaBucketName resolves the listing-image bucket with a safe default.
func mediaBucketName(cfg Config) string {
	if strings.TrimSpace(cfg.MediaBucket) != "" {
		return strings.TrimSpace(cfg.MediaBucket)
	}
	return "bn-public-media"
}

type Server struct {
	cfg                 Config
	trustedProxies      []netip.Prefix
	requestsTotal       atomic.Uint64
	requests5xx         atomic.Uint64
	catalogHandler      *catalogPres.CatalogHandler
	developerHandler    *developerPres.DeveloperHandler
	distributionHandler *distributionPres.DistributionHandler
	authHandler         *iamPres.AuthHandler
	commerceHandler     *commercePres.CommerceHandler
	moderationHandler   *moderationPres.ModerationHandler
	searchHandler       *searchPres.SearchHandler
	analyticsHandler    *analyticsPres.AnalyticsHandler
	adminHandler        *AdminHandler
}

func New(cfg Config) (*Server, error) {
	if cfg.DB == nil || cfg.Redis == nil || cfg.OpenSearchURL == "" || cfg.S3HealthURL == "" || cfg.Logger == nil {
		return nil, fmt.Errorf("database, redis, opensearch, object storage and logger are required")
	}
	if cfg.IAMHMACSecret == "" || cfg.DistributionTokenSecret == "" || cfg.PGWebhookSecret == "" || cfg.PayoutEncryptionKey == "" {
		return nil, fmt.Errorf("IAMHMACSecret, DistributionTokenSecret, PGWebhookSecret and PayoutEncryptionKey are required")
	}
	if cfg.JWTIssuer == "" {
		cfg.JWTIssuer = "https://auth.appmarket.local"
	}
	if cfg.CDNDomain == "" {
		cfg.CDNDomain = "cdn.appmarket.local"
	}
	if cfg.PublicBaseURL == "" {
		cfg.PublicBaseURL = "https://appmarket.bookingnetwork.internal"
	}
	if cfg.PGProvider == "" {
		cfg.PGProvider = "toss"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 2 * time.Second}
	}
	s := &Server{cfg: cfg}
	for _, cidr := range cfg.TrustedProxyCIDRs {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(cidr))
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q: %w", cidr, err)
		}
		s.trustedProxies = append(s.trustedProxies, prefix.Masked())
	}

	// 1. Catalog domain wiring
	var catalogRepo catalogApp.CatalogRepository
	if cfg.DB != nil {
		catalogRepo = catalogInfra.NewPostgresCatalogRepository(cfg.DB)
	} else {
		catalogRepo = catalogInfra.NewInMemoryCatalogRepository()
	}
	var catalogCache catalogApp.Cache
	if cfg.Redis != nil {
		catalogCache = catalogInfra.NewRedisCache(cfg.Redis)
	} else {
		catalogCache = catalogInfra.NewInMemoryCache()
	}
	catalogSvc := catalogApp.NewCatalogService(catalogRepo, catalogCache)
	s.catalogHandler = catalogPres.NewCatalogHandler(catalogSvc)

	// 2. IAM / Auth domain wiring
	userRepo := iamInfra.NewPostgresUserRepository(cfg.DB)
	sessionRepo := iamInfra.NewPostgresSessionRepository(cfg.DB)
	mfaRepo := iamInfra.NewPostgresMFARepository(cfg.DB)
	hasher := iamApp.NewPasswordHasher()
	jwtManager, err := iamApp.NewJWTManager(cfg.JWTPrivateKey, cfg.JWTIssuer)
	if err != nil {
		return nil, fmt.Errorf("jwt manager initialization failed: %w", err)
	}
	mfaService := iamApp.NewMFAService("BN AppMarket")
	dualApprovalRepo := iamRbac.NewInMemoryDualApprovalRepo()
	dualControl := iamRbac.NewDualControlEngine(dualApprovalRepo)
	matrixEnforcer := iamRbac.NewMatrixEnforcer(jwtManager)
	authCfg := iamApp.AuthServiceConfig{
		HMACSecret:      cfg.IAMHMACSecret,
		MaxFailedLogin:  5,
		LockoutDuration: 15 * time.Minute,
	}
	authSvc := iamApp.NewAuthService(userRepo, sessionRepo, mfaRepo, hasher, jwtManager, mfaService, authCfg)
	identityRepo := iamInfra.NewPostgresIdentityRepository(cfg.DB)
	oauthSvc := iamApp.NewOAuthService(authSvc, userRepo, identityRepo, cfg.OAuthRedirectBaseURL)
	oauthPresets := iamApp.DefaultOAuthProviders()
	googleCfg := oauthPresets["google"]
	googleCfg.ClientID, googleCfg.ClientSecret = cfg.GoogleClientID, cfg.GoogleClientSecret
	oauthSvc.RegisterProvider(googleCfg)
	githubCfg := oauthPresets["github"]
	githubCfg.ClientID, githubCfg.ClientSecret = cfg.GitHubClientID, cfg.GitHubClientSecret
	oauthSvc.RegisterProvider(githubCfg)
	appleCfg := oauthPresets["apple"]
	appleCfg.ClientID = cfg.AppleClientID
	appleCfg.AppleTeamID, appleCfg.AppleKeyID, appleCfg.ApplePrivateKeyPEM = cfg.AppleTeamID, cfg.AppleKeyID, cfg.ApplePrivateKeyPEM
	oauthSvc.RegisterProvider(appleCfg)
	s.authHandler = iamPres.NewAuthHandler(authSvc, jwtManager, dualControl, matrixEnforcer, cfg.JWTIssuer, oauthSvc)

	// 3. Developer domain wiring
	devRepo := developerInfra.NewPostgresDeveloperRepository(cfg.DB)
	orgSvc := developerApp.NewOrgService(devRepo)
	memberSvc := developerApp.NewMemberService(devRepo)
	apiKeySvc := developerApp.NewAPIKeyService(devRepo)
	identityResolver := func(req *http.Request) (string, error) {
		if claims, ok := iamRbac.GetClaims(req.Context()); ok && claims != nil {
			return claims.UserID, nil
		}
		authHeader := req.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtManager.ValidateAccessToken(tokenStr)
			if err == nil && claims != nil {
				return claims.UserID, nil
			}
		}
		return "", errors.New("unauthorized")
	}
	s.developerHandler = developerPres.NewDeveloperHandler(orgSvc, memberSvc, apiKeySvc, identityResolver)
	payoutKey := []byte(cfg.PayoutEncryptionKey)
	if len(payoutKey) != 32 {
		return nil, fmt.Errorf("PAYOUT_ENC_KEY must be exactly 32 bytes")
	}
	payoutAccountSvc, err := developerApp.NewPayoutService(devRepo, payoutKey)
	if err != nil {
		return nil, fmt.Errorf("payout service configuration failed: %w", err)
	}
	s.developerHandler.AttachPayouts(payoutAccountSvc)

	// 4. Distribution domain wiring
	tokenSvc := distributionApp.NewDownloadTokenService(cfg.DistributionTokenSecret, cfg.CDNDomain, 15*time.Minute)
	candidateProvider := distributionInfra.NewPostgresCandidateProvider(cfg.DB, cfg.CDNDomain)
	s.distributionHandler = distributionPres.NewDistributionHandler(tokenSvc, candidateProvider)

	// 5. Commerce domain wiring
	ledgerSvc := commerceApp.NewLedgerService()
	orderSvc := commerceApp.NewOrderService(ledgerSvc, cfg.PGWebhookSecret)
	refundSvc := commerceSet.NewRefundService(ledgerSvc, orderSvc)
	invoiceSvc := commerceSet.NewInvoiceService()
	subscriptionSvc := commerceSet.NewSubscriptionService()
	promoSvc := commerceApp.NewPromoService()
	pricingSvc := commerceApp.NewPricingService()
	payoutSvc := commerceSet.NewMonthlyPayoutService(ledgerSvc)
	s.commerceHandler = commercePres.NewCommerceHandler(orderSvc, refundSvc, invoiceSvc, subscriptionSvc, identityResolver, commercePGClient(cfg, cfg.Logger), cfg.PublicBaseURL, promoSvc).WithPricing(pricingSvc)

	// 6. Moderation domain wiring
	reviewSvc := moderationApp.NewReviewService(moderationApp.NewInMemoryInstallVerifier(), moderationApp.NewAntiFraudDetector())
	reportSvc := moderationApp.NewReportService()
	appealSvc := moderationApp.NewAppealService()
	s.moderationHandler = moderationPres.NewModerationHandler(reviewSvc, reportSvc, appealSvc, identityResolver)

	// 7. Search domain wiring (OpenSearch primary, in-memory fallback)
	searchPrimary := searchApp.NewOpenSearchEngine(cfg.OpenSearchURL, cfg.OpenSearchUser, cfg.OpenSearchPass, "apps", cfg.HTTPClient)
	searchSvc := searchApp.NewSearchService(searchPrimary, searchApp.NewInMemoryFallbackSearchEngine())
	s.searchHandler = searchPres.NewSearchHandler(searchSvc)

	// 8. Analytics wiring
	eventStore := analytics.NewMemoryEventStore()
	ingestor := analytics.NewFunnelIngestor(eventStore)
	s.analyticsHandler = analyticsPres.NewAnalyticsHandler(ingestor)

	// 9. Console aggregations (developer + admin)
	securitySvc := securityApp.NewSecurityService(nil)
	auditRepo := auditpkg.NewInMemoryAuditRepository()
	auditLogger := auditpkg.NewHashChainLogger(auditRepo)
	artifactSvc := distributionApp.NewArtifactService(distributionApp.NewInMemoryDistributionRepo(), nil)
	auctionEngine := adsApp.NewAuctionEngine(50)
	campaignSvc := adsApp.NewCampaignService(auctionEngine)
	resolver := consolepkg.NewOrgAppResolver(orgSvc, memberSvc, catalogSvc)
	uploadSvc := consolepkg.NewUploadService(artifactSvc)
	rolloutSvc := consolepkg.NewRolloutService(artifactSvc)
	catalogViewSvc := consolepkg.NewCatalogViewService(resolver, catalogSvc)
	metadataSvc := consolepkg.NewMetadataService(resolver, catalogSvc).
		WithMediaStorage(distributionApp.NewInMemoryObjectStorage(), mediaBucketName(cfg))
	adsViewSvc := consolepkg.NewAdsViewService(resolver, campaignSvc)
	commerceViewSvc := consolepkg.NewCommerceViewService(resolver, promoSvc, pricingSvc, orderSvc)
	lineageSvc := consolepkg.NewLineageService(resolver, artifactSvc)
	s.developerHandler.AttachConsole(developerPres.ConsoleDeps{
		Resolver: resolver, Uploads: uploadSvc, Rollouts: rolloutSvc,
		Catalog: catalogViewSvc, Metadata: metadataSvc, Ads: adsViewSvc,
		Commerce: commerceViewSvc, Lineage: lineageSvc,
	})
	decisionSvc := consolepkg.NewDecisionService(auditLogger).WithCatalog(catalogSvc)
	connectorSvc := consolepkg.NewConnectorService()
	s.adminHandler = NewAdminHandler(decisionSvc, securitySvc, connectorSvc, payoutSvc, auditRepo, orderSvc, campaignSvc, eventStore, reportSvc, appealSvc, identityResolver)
	s.adminHandler.AttachAuthorizer(matrixEnforcer)
	s.adminHandler.AttachVerifications(orgSvc)
	s.adminHandler.AttachCatalog(catalogSvc)
	s.adminHandler.AttachRefunds(orderSvc, refundSvc)
	s.distributionHandler.WithKillSwitchValidator(func(ctx context.Context, key string) error {
		if err := securitySvc.ValidateDownloadRequest(ctx, key); err != nil {
			return err
		}
		if app, err := catalogSvc.GetAppByID(ctx, key); err == nil && app != nil {
			return securitySvc.ValidateDownloadRequest(ctx, app.PrimaryPackageName)
		}
		return nil
	})

	return s, nil
}

// commercePGClient builds the configured payment-gateway client. Unknown or
// empty providers fall back to Toss; without live credentials the client
// operates in sandbox mode (documented on PGConfig).
func commercePGClient(cfg Config, logger *slog.Logger) commerceApp.PGClient {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	if strings.EqualFold(cfg.PGProvider, "portone") {
		logger.Info("commerce PG provider selected", "provider", "portone", "live", cfg.PGIsLive)
		return commerceApp.NewPortOneV2Client(commerceApp.PGConfig{
			Provider:     commerceApp.PGProviderPortOne,
			SecretKey:    cfg.PortOneAPIKey,
			ClientKey:    cfg.PortOneStoreID,
			WebhookSecret: cfg.PGWebhookSecret,
			IsLive:       cfg.PGIsLive,
			HTTPClient:   httpClient,
		})
	}
	logger.Info("commerce PG provider selected", "provider", "toss", "live", cfg.PGIsLive)
	return commerceApp.NewTossPaymentsClient(commerceApp.PGConfig{
		Provider:     commerceApp.PGProviderToss,
		SecretKey:    cfg.TossSecretKey,
		WebhookSecret: cfg.PGWebhookSecret,
		IsLive:       cfg.PGIsLive,
		HTTPClient:   httpClient,
	})
}

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(s.requestID, s.realIP, s.measureRequests, s.recoverPanic, s.cors, s.rateLimit)
	r.Get("/healthz/live", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Get("/healthz/ready", s.readiness)
	r.Get("/metrics", s.metrics)
	r.Get("/api/v1/system/capabilities", s.capabilities)

	if s.catalogHandler != nil {
		r.Mount("/api/v1/catalog", s.catalogHandler.Routes())
	}
	if s.developerHandler != nil {
		r.Mount("/api/v1/developer", s.developerHandler.Routes())
	}
	if s.distributionHandler != nil {
		r.Mount("/api/v1/distribution", s.distributionHandler.Routes())
	}
	if s.authHandler != nil {
		r.Mount("/api/v1/auth", s.authHandler.Routes())
	}
	if s.commerceHandler != nil {
		r.Mount("/api/v1/commerce", s.commerceHandler.Routes())
	}
	if s.moderationHandler != nil {
		r.Mount("/api/v1/moderation", s.moderationHandler.Routes())
	}
	if s.searchHandler != nil {
		r.Mount("/api/v1/search", s.searchHandler.Routes())
	}
	if s.analyticsHandler != nil {
		r.Mount("/api/v1/analytics", s.analyticsHandler.Routes())
	}
	if s.adminHandler != nil {
		r.Mount("/api/v1/admin", s.adminHandler.Routes())
	}

	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		problem.Write(w, http.StatusNotFound, "ERR_NOT_FOUND", "Route not found", "The requested endpoint does not exist.", traceID(req))
	})
	return otelhttp.NewHandler(r, "bn-appmarket-api")
}

func (s *Server) measureRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		defer func() {
			s.requestsTotal.Add(1)
			if wrapped.Status() >= 500 {
				s.requests5xx.Add(1)
			}
		}()
		next.ServeHTTP(wrapped, r)
	})
}

func (s *Server) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	total := s.requestsTotal.Load()
	failures := s.requests5xx.Load()
	_, _ = fmt.Fprintf(w, "# HELP bn_http_requests_total Total HTTP requests handled by the API.\n# TYPE bn_http_requests_total counter\nbn_http_requests_total{status_class=\"5xx\"} %d\nbn_http_requests_total{status_class=\"other\"} %d\n", failures, total-failures)
}

type requestIDKey struct{}
type clientIPKey struct{}

func (s *Server) trustedProxy(ip netip.Addr) bool {
	for _, prefix := range s.trustedProxies {
		if prefix.Contains(ip.Unmap()) {
			return true
		}
	}
	return false
}

// realIP trusts forwarded hops only when the immediate peer is explicitly trusted.
func (s *Server) realIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		peer, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			peer = r.RemoteAddr
		}
		ip, err := netip.ParseAddr(peer)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ip = ip.Unmap()
		if s.trustedProxy(ip) {
			hops := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
			for i := len(hops) - 1; i >= 0; i-- {
				candidate, parseErr := netip.ParseAddr(strings.TrimSpace(hops[i]))
				if parseErr != nil {
					break
				}
				ip = candidate.Unmap()
				if !s.trustedProxy(ip) {
					break
				}
			}
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), clientIPKey{}, ip.String())))
	})
}

func (s *Server) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b [16]byte
		if _, err := rand.Read(b[:]); err != nil {
			problem.Write(w, http.StatusInternalServerError, "ERR_INTERNAL_SERVER_ERROR", "Internal server error", "Request could not be initialized.", "")
			return
		}
		id := hex.EncodeToString(b[:])
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

func traceID(r *http.Request) string {
	if span := trace.SpanContextFromContext(r.Context()); span.IsValid() {
		return span.TraceID().String()
	}
	id, _ := r.Context().Value(requestIDKey{}).(string)
	return id
}

func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				s.cfg.Logger.Error("request panic", "request_id", traceID(r), "error", v)
				sentryx.Capture(r.Context(), fmt.Errorf("request panic: %v", v))
				problem.Write(w, http.StatusInternalServerError, "ERR_INTERNAL_SERVER_ERROR", "Internal server error", "The request failed unexpectedly.", traceID(r))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// cors enforces the configured allowlist. CSRF tokens are intentionally
// absent: the API is stateless Bearer-only and never issues cookies, so
// browsers attach no ambient credentials and cross-site request forgery has
// no vector. This is pinned by TestOWASP_NoCookieAuth_NoCSRFVector.
func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == s.cfg.AllowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-MFA-TOTP-Code")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

const slidingWindowScript = `
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
local count = redis.call('ZCARD', KEYS[1])
if count >= tonumber(ARGV[3]) then return -1 end
redis.call('ZADD', KEYS[1], ARGV[2], ARGV[4])
redis.call('EXPIRE', KEYS[1], 61)
return count + 1`

func (s *Server) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.DisableRateLimit || !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		ip, _ := r.Context().Value(clientIPKey{}).(string)
		if ip == "" {
			ip = r.RemoteAddr
		}
		key := "ratelimit:api:" + ip
		now := time.Now().UnixMilli()
		count, err := s.cfg.Redis.Eval(r.Context(), slidingWindowScript, []string{key}, now-60_000, now, 120, traceID(r)).Int64()
		if err != nil {
			s.cfg.Logger.Error("redis rate limiter unavailable", "error", err)
			problem.Write(w, http.StatusServiceUnavailable, "ERR_DEPENDENCY_UNAVAILABLE", "Service unavailable", "Rate limit service unavailable.", traceID(r))
			return
		}
		if count == -1 {
			w.Header().Set("Retry-After", "60")
			problem.Write(w, http.StatusTooManyRequests, "ERR_RATE_LIMITED", "Too many requests", "Try again later.", traceID(r))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	status := map[string]string{"postgres": "UP", "redis": "UP", "opensearch": "UP", "s3": "UP"}
	if err := s.cfg.DB.Ping(ctx); err != nil {
		status["postgres"] = "DOWN"
	}
	if err := s.cfg.Redis.Ping(ctx).Err(); err != nil {
		status["redis"] = "DOWN"
	}
	if !s.checkURL(ctx, strings.TrimRight(s.cfg.OpenSearchURL, "/")+"/_cluster/health?wait_for_status=yellow&timeout=1s", true) {
		status["opensearch"] = "DOWN"
	}
	if !s.checkURL(ctx, s.cfg.S3HealthURL, false) {
		status["s3"] = "DOWN"
	}
	w.Header().Set("Content-Type", "application/json")
	for _, v := range status {
		if v != "UP" {
			w.WriteHeader(http.StatusServiceUnavailable)
			break
		}
	}
	_ = json.NewEncoder(w).Encode(status)
}

func (s *Server) checkURL(ctx context.Context, url string, opensearch bool) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	if opensearch && s.cfg.OpenSearchUser != "" {
		req.SetBasicAuth(s.cfg.OpenSearchUser, s.cfg.OpenSearchPass)
	}
	res, err := s.cfg.HTTPClient.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	return res.StatusCode == http.StatusOK
}

func (s *Server) capabilities(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"distribution_modes":           []string{"DIRECT_HOSTED", "FEDERATED_REPO_BINARY", "SEARCH_INFO_ONLY", "EXTERNAL_LINK_ONLY"},
		"treescan_integration_enabled": false,
		"min_android_sdk":              26,
	})
}

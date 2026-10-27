package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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

	problem "github.com/bookingnetwork/appmarket/backend/internal/platform/errors"
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
}

type Server struct {
	cfg            Config
	trustedProxies []netip.Prefix
	requestsTotal  atomic.Uint64
	requests5xx    atomic.Uint64
}

func New(cfg Config) (*Server, error) {
	if cfg.DB == nil || cfg.Redis == nil || cfg.OpenSearchURL == "" || cfg.S3HealthURL == "" || cfg.Logger == nil {
		return nil, fmt.Errorf("database, redis, opensearch, object storage and logger are required")
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
	return s, nil
}

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(s.requestID, s.realIP, s.measureRequests, s.recoverPanic, s.cors, s.rateLimit)
	r.Get("/healthz/live", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Get("/healthz/ready", s.readiness)
	r.Get("/metrics", s.metrics)
	r.Get("/api/v1/system/capabilities", s.capabilities)
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
				problem.Write(w, http.StatusInternalServerError, "ERR_INTERNAL_SERVER_ERROR", "Internal server error", "The request failed unexpectedly.", traceID(r))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

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
		if !strings.HasPrefix(r.URL.Path, "/api/") {
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

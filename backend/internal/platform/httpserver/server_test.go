package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func TestMetricsCountActualHTTPStatus(t *testing.T) {
	s := &Server{}
	h := s.measureRequests(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/fail" {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	for _, path := range []string{"/ok", "/fail"} {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}
	metrics := httptest.NewRecorder()
	s.metrics(metrics, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	for _, want := range []string{
		`bn_http_requests_total{status_class="5xx"} 1`,
		`bn_http_requests_total{status_class="other"} 1`,
	} {
		if !strings.Contains(metrics.Body.String(), want) {
			t.Fatalf("metrics missing %q: %s", want, metrics.Body.String())
		}
	}
}

func TestRealIPOnlyTrustsConfiguredProxyChain(t *testing.T) {
	s := &Server{trustedProxies: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}}
	cases := []struct {
		name, peer, forwarded, want string
	}{
		{"untrusted peer cannot spoof", "203.0.113.7:1234", "198.51.100.8", "203.0.113.7"},
		{"trusted chain", "10.0.0.2:1234", "198.51.100.8, 10.0.0.3", "198.51.100.8"},
		{"malformed hop stops traversal", "10.0.0.2:1234", "invalid, 10.0.0.3", "10.0.0.3"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			h := s.realIP(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				got, _ = r.Context().Value(clientIPKey{}).(string)
			}))
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tc.peer
			r.Header.Set("X-Forwarded-For", tc.forwarded)
			h.ServeHTTP(httptest.NewRecorder(), r)
			if got != tc.want {
				t.Fatalf("client IP = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestReadinessFailsClosedWhileLivenessRemainsAvailable(t *testing.T) {
	probe := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer probe.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, err := pgxpool.New(ctx, "postgres://unused:unused@127.0.0.1:1/unused?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cache := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	defer cache.Close()
	s, err := New(Config{
		DB: db, Redis: cache, OpenSearchURL: probe.URL, S3HealthURL: probe.URL,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		DisableRateLimit: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	h := s.Handler()
	live := httptest.NewRecorder()
	h.ServeHTTP(live, httptest.NewRequest(http.MethodGet, "/healthz/live", nil))
	if live.Code != http.StatusOK {
		t.Fatalf("liveness = %d", live.Code)
	}
	ready := httptest.NewRecorder()
	h.ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/healthz/ready", nil))
	if ready.Code != http.StatusServiceUnavailable {
		t.Fatalf("readiness = %d: %s", ready.Code, ready.Body.String())
	}
	var state map[string]string
	if err := json.Unmarshal(ready.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state["postgres"] != "DOWN" || state["redis"] != "DOWN" || state["opensearch"] != "UP" || state["s3"] != "UP" {
		t.Fatalf("unexpected dependency state: %#v", state)
	}
}

func TestMountedDomainRoutes(t *testing.T) {
	probe := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer probe.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db, err := pgxpool.New(ctx, "postgres://unused:unused@127.0.0.1:1/unused?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cache := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	defer cache.Close()

	s, err := New(Config{
		DB: db, Redis: cache, OpenSearchURL: probe.URL, S3HealthURL: probe.URL,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		DisableRateLimit: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	h := s.Handler()

	// 1. Catalog apps route returns 200 OK
	wCat := httptest.NewRecorder()
	h.ServeHTTP(wCat, httptest.NewRequest(http.MethodGet, "/api/v1/catalog/apps", nil))
	if wCat.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/v1/catalog/apps, got %d: %s", wCat.Code, wCat.Body.String())
	}

	// 2. Auth discovery route returns 200 OK
	wAuth := httptest.NewRecorder()
	h.ServeHTTP(wAuth, httptest.NewRequest(http.MethodGet, "/api/v1/auth/.well-known/openid-configuration", nil))
	if wAuth.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/v1/auth/.well-known/openid-configuration, got %d: %s", wAuth.Code, wAuth.Body.String())
	}

	// 3. Distribution check-updates route returns 200 OK
	checkBody := `{"device":{"sdk_int":34,"supported_abis":["arm64-v8a"],"density_dpi":440,"available_features":[],"device_hash":"dev-1"},"packages":[{"package_name":"com.omniwork.pro","version_code":100,"cert_sha256":"A1B2C3D4E5F67890123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0","installer_package_name":"com.bookingnetwork.appmarket"}]}`
	reqDist := httptest.NewRequest(http.MethodPost, "/api/v1/distribution/device/check-updates", strings.NewReader(checkBody))
	reqDist.Header.Set("Content-Type", "application/json")
	wDist := httptest.NewRecorder()
	h.ServeHTTP(wDist, reqDist)
	if wDist.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/v1/distribution/device/check-updates, got %d: %s", wDist.Code, wDist.Body.String())
	}

	// 4. Developer protected route returns 401 Unauthorized without auth
	wDev := httptest.NewRecorder()
	h.ServeHTTP(wDev, httptest.NewRequest(http.MethodPost, "/api/v1/developer/orgs", strings.NewReader(`{"slug":"test-org","name":"Test Org"}`)))
	if wDev.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized from unauthenticated /api/v1/developer/orgs, got %d: %s", wDev.Code, wDev.Body.String())
	}
}

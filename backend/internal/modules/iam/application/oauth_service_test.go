package application_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
)

type mockOAuthUserRepo struct {
	mu    sync.RWMutex
	byID  map[string]*domain.User
	byIdx map[string]string
}

func newMockOAuthUserRepo() *mockOAuthUserRepo {
	return &mockOAuthUserRepo{byID: make(map[string]*domain.User), byIdx: make(map[string]string)}
}

func (r *mockOAuthUserRepo) CreateUser(_ context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *u
	r.byID[u.ID] = &cp
	r.byIdx[u.EmailBlindIndex] = u.ID
	return nil
}

func (r *mockOAuthUserRepo) GetUserByID(_ context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	cp := *u
	return &cp, nil
}

func (r *mockOAuthUserRepo) GetUserByEmailBlindIndex(_ context.Context, blindIndex string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byIdx[blindIndex]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	cp := *r.byID[id]
	return &cp, nil
}

func (r *mockOAuthUserRepo) UpdateUserStatus(_ context.Context, _ string, _ domain.UserStatus, _ int, _ *time.Time) error {
	return nil
}

func (r *mockOAuthUserRepo) GetUserRoles(_ context.Context, _ string) ([]string, map[string]string, error) {
	return []string{"USER"}, map[string]string{}, nil
}

type mockOAuthSessionRepo struct {
	mu       sync.RWMutex
	sessions map[string]*domain.AuthSession
}

func (r *mockOAuthSessionRepo) CreateSession(_ context.Context, s *domain.AuthSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sessions == nil {
		r.sessions = make(map[string]*domain.AuthSession)
	}
	r.sessions[s.RefreshTokenHash] = s
	return nil
}

func (r *mockOAuthSessionRepo) GetSessionByTokenHash(_ context.Context, h string) (*domain.AuthSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[h]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	return s, nil
}

func (r *mockOAuthSessionRepo) RotateSession(_ context.Context, _ string, _ *domain.AuthSession) error {
	return nil
}

func (r *mockOAuthSessionRepo) RevokeFamily(_ context.Context, _ string) error { return nil }

type mockOAuthMFARepo struct{}

func (mockOAuthMFARepo) GetMFACredential(_ context.Context, _ string) (*domain.MFACredential, error) {
	return nil, nil
}

func (mockOAuthMFARepo) SaveMFACredential(_ context.Context, _ *domain.MFACredential) error {
	return nil
}

type mockIdentityRepo struct {
	mu   sync.RWMutex
	link map[string]string
}

func (r *mockIdentityRepo) FindUserIDByProvider(_ context.Context, provider, subject string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.link[provider+":"+subject]
	if !ok {
		return "", domain.ErrUserNotFound
	}
	return id, nil
}

func (r *mockIdentityRepo) LinkIdentity(_ context.Context, identity *domain.UserIdentity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.link == nil {
		r.link = make(map[string]string)
	}
	r.link[identity.Provider+":"+identity.ProviderSubject] = identity.UserID
	return nil
}

func oauthTestSetup(t *testing.T) (*application.AuthService, *mockOAuthUserRepo, *mockIdentityRepo) {
	t.Helper()
	users := newMockOAuthUserRepo()
	identities := &mockIdentityRepo{}
	jwtMgr, err := application.NewJWTManager(nil, "https://auth.test")
	if err != nil {
		t.Fatalf("jwt manager failed: %v", err)
	}
	auth := application.NewAuthService(users, &mockOAuthSessionRepo{}, mockOAuthMFARepo{},
		application.NewPasswordHasher(), jwtMgr, application.NewMFAService("test"),
		application.AuthServiceConfig{HMACSecret: "oauth-test-hmac-secret-32-chars!", MaxFailedLogin: 5})
	return auth, users, identities
}

type mockOIDC struct {
	server   *httptest.Server
	key      *rsa.PrivateKey
	kid      string
	idToken  func(subject, email, name string) string
	serveBad func()
}

func newMockOIDC(t *testing.T, issuer, audience string) *mockOIDC {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}
	m := &mockOIDC{key: key, kid: "mock-key-1"}
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("code") != "good-code" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
			return
		}
		email := r.Form.Get("email_hint")
		if email == "" {
			email = "social.user@example.com"
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"access_token": "mock-access-token",
			"id_token":     m.idToken("google-sub-123", email, "Social User"),
			"token_type":   "Bearer",
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		n := base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString(bigIntToBytes(big.NewInt(int64(key.PublicKey.E))))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{
			{"kid": m.kid, "kty": "RSA", "n": n, "e": e},
		}})
	})
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":424242,"email":null,"name":"GH User","login":"ghuser"}`))
	})
	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"email":"gh.user@example.com","primary":true,"verified":true}]`))
	})
	m.server = httptest.NewServer(mux)
	t.Cleanup(m.server.Close)
	m.idToken = func(subject, email, name string) string {
		verified := true
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"iss": issuer, "aud": audience, "sub": subject,
			"email": email, "email_verified": verified, "name": name,
			"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
		})
		token.Header["kid"] = m.kid
		signed, err := token.SignedString(key)
		if err != nil {
			t.Fatalf("failed to sign mock id token: %v", err)
		}
		return signed
	}
	return m
}

func bigIntToBytes(v *big.Int) []byte {
	b := v.Bytes()
	if len(b) == 0 {
		return []byte{0}
	}
	return b
}

func registerMockGoogle(t *testing.T, svc *application.OAuthService, mock *mockOIDC) {
	t.Helper()
	presets := application.DefaultOAuthProviders()
	cfg := presets["google"]
	cfg.ClientID = "mock-google-client-id"
	cfg.ClientSecret = "mock-google-secret"
	cfg.TokenURL = mock.server.URL + "/token"
	cfg.JWKURL = mock.server.URL + "/jwks"
	cfg.Issuer = "https://mock-issuer.test"
	svc.RegisterProvider(cfg)
}

func TestOAuth_LoginURLUnknownProvider(t *testing.T) {
	auth, users, identities := oauthTestSetup(t)
	svc := application.NewOAuthService(auth, users, identities, "https://app.test/oauth")
	if _, err := svc.LoginURL("google", "state"); err != domain.ErrOAuthProviderUnknown {
		t.Fatalf("expected provider unknown, got %v", err)
	}
}

func TestOAuth_FullOIDCFlowProvisionsAndRelogins(t *testing.T) {
	auth, users, identities := oauthTestSetup(t)
	svc := application.NewOAuthService(auth, users, identities, "https://app.test/oauth")
	mock := newMockOIDC(t, "https://mock-issuer.test", "mock-google-client-id")
	registerMockGoogle(t, svc, mock)

	state, err := svc.GenerateState()
	if err != nil {
		t.Fatalf("state failed: %v", err)
	}
	loginURL, err := svc.LoginURL("google", state)
	if err != nil {
		t.Fatalf("login url failed: %v", err)
	}
	if !strings.Contains(loginURL, "mock-google-client-id") || !strings.Contains(loginURL, state) {
		t.Fatalf("login url missing params: %s", loginURL)
	}

	res, err := svc.HandleCallback(context.Background(), "google", "good-code", state, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("callback failed: %v", err)
	}
	if res.AccessToken == "" || res.RefreshToken == "" || res.User == nil {
		t.Fatalf("expected session tokens and user: %+v", res)
	}
	if res.User.Email != "social.user@example.com" {
		t.Fatalf("unexpected user email: %s", res.User.Email)
	}
	firstUserID := res.User.ID

	state2, _ := svc.GenerateState()
	res2, err := svc.HandleCallback(context.Background(), "google", "good-code", state2, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("second callback failed: %v", err)
	}
	if res2.User.ID != firstUserID {
		t.Fatalf("identity login returned different user: %s vs %s", res2.User.ID, firstUserID)
	}
	_ = users
	_ = identities
}

func TestOAuth_LinksPreexistingEmailUser(t *testing.T) {
	auth, users, identities := oauthTestSetup(t)
	svc := application.NewOAuthService(auth, users, identities, "https://app.test/oauth")
	mock := newMockOIDC(t, "https://mock-issuer.test", "mock-google-client-id")
	registerMockGoogle(t, svc, mock)

	existing, err := auth.Register(context.Background(), application.RegisterInput{
		Email: "social.user@example.com", Password: "Password123!", DisplayName: "Existing",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	state, _ := svc.GenerateState()
	res, err := svc.HandleCallback(context.Background(), "google", "good-code", state, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("callback failed: %v", err)
	}
	if res.User.ID != existing.ID {
		t.Fatalf("expected link to existing user %s, got %s", existing.ID, res.User.ID)
	}
}

func TestOAuth_RejectsBadStateAndCode(t *testing.T) {
	auth, users, identities := oauthTestSetup(t)
	svc := application.NewOAuthService(auth, users, identities, "https://app.test/oauth")
	mock := newMockOIDC(t, "https://mock-issuer.test", "mock-google-client-id")
	registerMockGoogle(t, svc, mock)

	if _, err := svc.HandleCallback(context.Background(), "google", "good-code", "forged-state", "", ""); err != domain.ErrOAuthStateInvalid {
		t.Fatalf("expected state invalid, got %v", err)
	}
	state, _ := svc.GenerateState()
	if _, err := svc.HandleCallback(context.Background(), "google", "bad-code", state, "", ""); err == nil {
		t.Fatalf("expected exchange failure for bad code")
	}
	// state is single-use
	if _, err := svc.HandleCallback(context.Background(), "google", "good-code", state, "", ""); err != domain.ErrOAuthStateInvalid {
		t.Fatalf("expected consumed state to be invalid, got %v", err)
	}
}

func TestOAuth_RejectsWrongKeyIDToken(t *testing.T) {
	auth, users, identities := oauthTestSetup(t)
	svc := application.NewOAuthService(auth, users, identities, "https://app.test/oauth")
	signer := newMockOIDC(t, "https://mock-issuer.test", "mock-google-client-id")
	keyServer := newMockOIDC(t, "https://mock-issuer.test", "mock-google-client-id")
	// Cross-wire: tokens signed by one key, verified against another key set.
	presets := application.DefaultOAuthProviders()
	cfg := presets["google"]
	cfg.ClientID = "mock-google-client-id"
	cfg.ClientSecret = "mock-google-secret"
	cfg.TokenURL = signer.server.URL + "/token"
	cfg.JWKURL = keyServer.server.URL + "/jwks"
	cfg.Issuer = "https://mock-issuer.test"
	svc.RegisterProvider(cfg)
	state, _ := svc.GenerateState()
	_, err := svc.HandleCallback(context.Background(), "google", "good-code", state, "", "")
	if err == nil || !strings.Contains(err.Error(), "ERR_OAUTH_ID_TOKEN_INVALID") {
		t.Fatalf("expected id-token rejection, got %v", err)
	}
}

func TestOAuth_RejectsUnreachableProvider(t *testing.T) {
	auth, users, identities := oauthTestSetup(t)
	svc := application.NewOAuthService(auth, users, identities, "https://app.test/oauth")
	presets := application.DefaultOAuthProviders()
	cfg := presets["google"]
	cfg.ClientID = "mock-google-client-id"
	cfg.ClientSecret = "mock-google-secret"
	cfg.TokenURL = "https://127.0.0.1:1/token"
	cfg.JWKURL = "https://127.0.0.1:1/jwks"
	cfg.Issuer = "https://mock-issuer.test"
	svc.RegisterProvider(cfg)
	state, _ := svc.GenerateState()
	if _, err := svc.HandleCallback(context.Background(), "google", "good-code", state, "", ""); err == nil {
		t.Fatalf("expected failure when provider is unreachable")
	}
}

func TestOAuth_GitHubFlowUsesEmailsAPI(t *testing.T) {
	auth, users, identities := oauthTestSetup(t)
	svc := application.NewOAuthService(auth, users, identities, "https://app.test/oauth")
	mock := newMockOIDC(t, "https://mock-issuer.test", "mock-github-client-id")
	presets := application.DefaultOAuthProviders()
	cfg := presets["github"]
	cfg.ClientID = "mock-github-client-id"
	cfg.ClientSecret = "mock-github-secret"
	cfg.TokenURL = mock.server.URL + "/token"
	cfg.UserInfoURL = mock.server.URL + "/user"
	cfg.UserEmailsURL = mock.server.URL + "/user/emails"
	svc.RegisterProvider(cfg)

	state, _ := svc.GenerateState()
	res, err := svc.HandleCallback(context.Background(), "github", "good-code", state, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("github callback failed: %v", err)
	}
	if res.User.Email != "gh.user@example.com" {
		t.Fatalf("expected primary verified email, got %s", res.User.Email)
	}
	_ = users
	_ = identities
}

func TestOAuth_ProvidersListsActive(t *testing.T) {
	auth, users, identities := oauthTestSetup(t)
	svc := application.NewOAuthService(auth, users, identities, "https://app.test/oauth")
	if len(svc.Providers()) != 0 {
		t.Fatalf("expected no providers before registration")
	}
	mock := newMockOIDC(t, "https://mock-issuer.test", "mock-google-client-id")
	registerMockGoogle(t, svc, mock)
	got := svc.Providers()
	if len(got) != 1 || got[0] != "google" {
		t.Fatalf("unexpected providers: %v", got)
	}
}

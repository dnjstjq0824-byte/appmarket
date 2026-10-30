package presentation

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
)

// Unexpected persistence calls panic so rejected requests cannot mutate state.
type permissionRepository struct {
	application.DeveloperRepository
	role domain.DevRole
	err  error
}

func (r permissionRepository) GetMember(_ context.Context, orgID, userID string) (*domain.OrgMember, error) {
	if r.err != nil {
		return nil, r.err
	}
	if orgID != "org-a" || userID != "user-a" {
		return nil, nil
	}
	return &domain.OrgMember{OrgID: orgID, UserID: userID, Role: r.role}, nil
}

func TestRoutesRequireAuthenticatedIdentity(t *testing.T) {
	for _, resolver := range []IdentityResolver{nil,
		func(*http.Request) (string, error) { return "", nil },
		func(*http.Request) (string, error) { return "user-a", errors.New("invalid token") },
	} {
		for _, path := range []string{"/orgs", "/orgs/org-a/verification", "/orgs/org-a/members", "/orgs/org-a/api-keys"} {
			h := NewDeveloperHandler(nil, nil, nil, resolver)
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader("{}"))
			req.Header.Set("X-User-ID", "user-a")
			w := httptest.NewRecorder()
			h.Routes().ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized || !strings.Contains(w.Body.String(), "ERR_UNAUTHORIZED") {
				t.Fatalf("%s: got %d %s", path, w.Code, w.Body.String())
			}
		}
	}
}

func TestTCSEC002HTTPOrganizationPermissions(t *testing.T) {
	for _, role := range []domain.DevRole{domain.RoleDevOwner, domain.RoleReleaseManager, domain.RoleListingEditor, domain.RoleFinanceManager, domain.RoleSupportAgent} {
		for _, org := range []string{"org-a", "org-b"} {
			for _, endpoint := range []string{"verification", "members", "api-keys"} {
				t.Run(string(role)+"/"+org+"/"+endpoint, func(t *testing.T) {
					members := application.NewMemberService(permissionRepository{role: role})
					h := NewDeveloperHandler(nil, members, nil, func(*http.Request) (string, error) { return "user-a", nil })
					w := httptest.NewRecorder()
					// Malformed JSON proves authorization runs before decoding or mutation.
					h.Routes().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/orgs/"+org+"/"+endpoint, strings.NewReader("{")))
					want := http.StatusForbidden
					if org == "org-a" && role == domain.RoleDevOwner {
						want = http.StatusBadRequest
					}
					if w.Code != want {
						t.Fatalf("got %d, want %d: %s", w.Code, want, w.Body.String())
					}
					if want == http.StatusForbidden && !strings.Contains(w.Body.String(), "ERR_FORBIDDEN_TENANT_ACCESS") {
						t.Fatal(w.Body.String())
					}
				})
			}
		}
	}
}

func TestMembershipFailureDeniesAccess(t *testing.T) {
	members := application.NewMemberService(permissionRepository{err: errors.New("database unavailable")})
	h := NewDeveloperHandler(nil, members, nil, func(*http.Request) (string, error) { return "user-a", nil })
	w := httptest.NewRecorder()
	h.Routes().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/orgs/org-a/api-keys", strings.NewReader("{}")))
	if w.Code != http.StatusForbidden {
		t.Fatalf("got %d", w.Code)
	}
}

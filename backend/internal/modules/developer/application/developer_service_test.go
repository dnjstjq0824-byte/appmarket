package application_test

import (
	"context"
	"sync"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
)

type mockDevRepo struct {
	mu         sync.RWMutex
	orgs       map[string]*domain.DeveloperOrg
	orgsSlug   map[string]string
	verifs     map[string][]domain.DeveloperVerification
	members    map[string]map[string]*domain.OrgMember
	keys       map[string]*domain.APIKey
	keysByHash map[string]string
	keysByOrg  map[string][]string
}

func newMockDevRepo() *mockDevRepo {
	return &mockDevRepo{
		orgs:       make(map[string]*domain.DeveloperOrg),
		orgsSlug:   make(map[string]string),
		verifs:     make(map[string][]domain.DeveloperVerification),
		members:    make(map[string]map[string]*domain.OrgMember),
		keys:       make(map[string]*domain.APIKey),
		keysByHash: make(map[string]string),
		keysByOrg:  make(map[string][]string),
	}
}

func (r *mockDevRepo) CreateOrg(ctx context.Context, org *domain.DeveloperOrg) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orgs[org.ID] = org
	r.orgsSlug[org.OrgSlug] = org.ID
	return nil
}
func (r *mockDevRepo) GetOrgByID(ctx context.Context, orgID string) (*domain.DeveloperOrg, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.orgs[orgID]
	if !ok {
		return nil, nil
	}
	cp := *o
	return &cp, nil
}
func (r *mockDevRepo) GetOrgBySlug(ctx context.Context, slug string) (*domain.DeveloperOrg, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.orgsSlug[slug]
	if !ok {
		return nil, nil
	}
	o := r.orgs[id]
	cp := *o
	return &cp, nil
}
func (r *mockDevRepo) UpdateOrgStatus(ctx context.Context, orgID string, status domain.OrgVerificationStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	o, ok := r.orgs[orgID]
	if !ok {
		return domain.ErrOrgNotFound
	}
	o.VerificationStatus = status
	return nil
}
func (r *mockDevRepo) CreateVerification(ctx context.Context, v *domain.DeveloperVerification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.verifs[v.OrgID] = append(r.verifs[v.OrgID], *v)
	return nil
}
func (r *mockDevRepo) GetLatestVerification(ctx context.Context, orgID string) (*domain.DeveloperVerification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	vs := r.verifs[orgID]
	if len(vs) == 0 {
		return nil, nil
	}
	cp := vs[len(vs)-1]
	return &cp, nil
}
func (r *mockDevRepo) AddMember(ctx context.Context, m *domain.OrgMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.members[m.OrgID] == nil {
		r.members[m.OrgID] = make(map[string]*domain.OrgMember)
	}
	cp := *m
	r.members[m.OrgID][m.UserID] = &cp
	return nil
}
func (r *mockDevRepo) GetMember(ctx context.Context, orgID, userID string) (*domain.OrgMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.members[orgID] == nil {
		return nil, nil
	}
	m, ok := r.members[orgID][userID]
	if !ok {
		return nil, nil
	}
	cp := *m
	return &cp, nil
}
func (r *mockDevRepo) ListMembers(ctx context.Context, orgID string) ([]domain.OrgMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.OrgMember
	for _, m := range r.members[orgID] {
		out = append(out, *m)
	}
	return out, nil
}
func (r *mockDevRepo) CreateAPIKey(ctx context.Context, k *domain.APIKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *k
	r.keys[k.ID] = &cp
	r.keysByHash[k.KeyHash] = k.ID
	r.keysByOrg[k.OrgID] = append(r.keysByOrg[k.OrgID], k.ID)
	return nil
}
func (r *mockDevRepo) GetAPIKeyByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.keysByHash[keyHash]
	if !ok {
		return nil, nil
	}
	k := r.keys[id]
	cp := *k
	return &cp, nil
}
func (r *mockDevRepo) RevokeAPIKey(ctx context.Context, keyID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k, ok := r.keys[keyID]
	if !ok {
		return domain.ErrAPIKeyNotFound
	}
	k.Revoked = true
	return nil
}
func (r *mockDevRepo) ListAPIKeys(ctx context.Context, orgID string) ([]domain.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.APIKey
	for _, id := range r.keysByOrg[orgID] {
		if k, ok := r.keys[id]; ok {
			out = append(out, *k)
		}
	}
	return out, nil
}

func setupDev(t *testing.T) (*application.OrgService, *application.MemberService, *application.APIKeyService, *mockDevRepo) {
	t.Helper()
	repo := newMockDevRepo()
	return application.NewOrgService(repo), application.NewMemberService(repo), application.NewAPIKeyService(repo), repo
}

func TestDevOrgUnverifiedBlocksProduction(t *testing.T) {
	orgs, _, _, _ := setupDev(t)
	ctx := context.Background()
	org, err := orgs.CreateOrg(ctx, "acme", "Acme Inc")
	if err != nil {
		t.Fatalf("create org failed: %v", err)
	}
	if org.VerificationStatus != domain.OrgUnverified {
		t.Fatalf("expected UNVERIFIED, got %s", org.VerificationStatus)
	}
	if err := orgs.EnsureVerified(ctx, org.ID); err != domain.ErrOrgNotVerified {
		t.Fatalf("expected ErrOrgNotVerified, got %v", err)
	}
	if _, err := orgs.SubmitVerification(ctx, org.ID, "KYB", "https://example.com/doc", "acme.com"); err != nil {
		t.Fatalf("submit verification failed: %v", err)
	}
	o, _ := orgs.GetOrg(ctx, org.ID)
	if o.VerificationStatus != domain.OrgPendingReview {
		t.Fatalf("expected PENDING_REVIEW, got %s", o.VerificationStatus)
	}
	if err := orgs.EnsureVerified(ctx, org.ID); err != domain.ErrOrgNotVerified {
		t.Fatalf("expected ErrOrgNotVerified after pending, got %v", err)
	}
}

package infrastructure

import (
	"context"
	"sync"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
)

// InMemoryDeveloperRepository provides a thread-safe in-memory store for developer organizations,
// verifications, members, and API keys.
type InMemoryDeveloperRepository struct {
	mu            sync.RWMutex
	orgs          map[string]*domain.DeveloperOrg
	orgsBySlug    map[string]*domain.DeveloperOrg
	verifications map[string][]*domain.DeveloperVerification
	members       map[string]map[string]*domain.OrgMember // orgID -> userID -> OrgMember
	apiKeys       map[string]*domain.APIKey               // keyID -> APIKey
	apiKeysByHash map[string]*domain.APIKey               // keyHash -> APIKey
	payouts       map[string]*domain.PayoutRecord          // orgID -> payout record
}

func NewInMemoryDeveloperRepository() *InMemoryDeveloperRepository {
	return &InMemoryDeveloperRepository{
		orgs:          make(map[string]*domain.DeveloperOrg),
		orgsBySlug:    make(map[string]*domain.DeveloperOrg),
		verifications: make(map[string][]*domain.DeveloperVerification),
		members:       make(map[string]map[string]*domain.OrgMember),
		apiKeys:       make(map[string]*domain.APIKey),
		apiKeysByHash: make(map[string]*domain.APIKey),
		payouts:       make(map[string]*domain.PayoutRecord),
	}
}

func (r *InMemoryDeveloperRepository) CreateOrg(ctx context.Context, org *domain.DeveloperOrg) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *org
	r.orgs[org.ID] = &cp
	r.orgsBySlug[org.OrgSlug] = &cp
	return nil
}

func (r *InMemoryDeveloperRepository) GetOrgByID(ctx context.Context, orgID string) (*domain.DeveloperOrg, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	org, ok := r.orgs[orgID]
	if !ok {
		return nil, domain.ErrOrgNotFound
	}
	cp := *org
	return &cp, nil
}

func (r *InMemoryDeveloperRepository) GetOrgBySlug(ctx context.Context, slug string) (*domain.DeveloperOrg, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	org, ok := r.orgsBySlug[slug]
	if !ok {
		return nil, domain.ErrOrgNotFound
	}
	cp := *org
	return &cp, nil
}

func (r *InMemoryDeveloperRepository) ListOrgsForUser(ctx context.Context, userID string) ([]domain.DeveloperOrg, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.DeveloperOrg
	for orgID, users := range r.members {
		if _, ok := users[userID]; ok {
			if o, exists := r.orgs[orgID]; exists {
				out = append(out, *o)
			}
		}
	}
	return out, nil
}

func (r *InMemoryDeveloperRepository) UpdateOrgStatus(ctx context.Context, orgID string, status domain.OrgVerificationStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	org, ok := r.orgs[orgID]
	if !ok {
		return domain.ErrOrgNotFound
	}
	org.VerificationStatus = status
	return nil
}

func (r *InMemoryDeveloperRepository) CreateVerification(ctx context.Context, v *domain.DeveloperVerification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *v
	r.verifications[v.OrgID] = append(r.verifications[v.OrgID], &cp)
	return nil
}

func (r *InMemoryDeveloperRepository) GetLatestVerification(ctx context.Context, orgID string) (*domain.DeveloperVerification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := r.verifications[orgID]
	if len(list) == 0 {
		return nil, nil
	}
	cp := *list[len(list)-1]
	return &cp, nil
}

func (r *InMemoryDeveloperRepository) DecideVerificationRecord(ctx context.Context, verificationID, status, reviewerID, notes string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, list := range r.verifications {
		for _, v := range list {
			if v.ID == verificationID {
				v.Status = status
				v.ReviewerID = reviewerID
				v.ReviewNotes = notes
				return nil
			}
		}
	}
	return domain.ErrVerificationNotFound
}

func (r *InMemoryDeveloperRepository) ListVerificationsByStatus(ctx context.Context, status string) ([]domain.DeveloperVerification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []domain.DeveloperVerification
	for _, list := range r.verifications {
		for _, v := range list {
			if v.Status == status {
				out = append(out, *v)
			}
		}
	}
	return out, nil
}

func (r *InMemoryDeveloperRepository) AddMember(ctx context.Context, m *domain.OrgMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.members[m.OrgID]; !ok {
		r.members[m.OrgID] = make(map[string]*domain.OrgMember)
	}
	cp := *m
	r.members[m.OrgID][m.UserID] = &cp
	return nil
}

func (r *InMemoryDeveloperRepository) GetMember(ctx context.Context, orgID, userID string) (*domain.OrgMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userMap, ok := r.members[orgID]
	if !ok {
		return nil, domain.ErrMemberNotFound
	}
	m, ok := userMap[userID]
	if !ok {
		return nil, domain.ErrMemberNotFound
	}
	cp := *m
	return &cp, nil
}

func (r *InMemoryDeveloperRepository) ListMembers(ctx context.Context, orgID string) ([]domain.OrgMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userMap, ok := r.members[orgID]
	if !ok {
		return nil, nil
	}
	res := make([]domain.OrgMember, 0, len(userMap))
	for _, m := range userMap {
		res = append(res, *m)
	}
	return res, nil
}

func (r *InMemoryDeveloperRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userMap, ok := r.members[orgID]
	if !ok {
		return domain.ErrMemberNotFound
	}
	if _, ok := userMap[userID]; !ok {
		return domain.ErrMemberNotFound
	}
	delete(userMap, userID)
	return nil
}

func (r *InMemoryDeveloperRepository) CreateAPIKey(ctx context.Context, k *domain.APIKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *k
	r.apiKeys[k.ID] = &cp
	r.apiKeysByHash[k.KeyHash] = &cp
	return nil
}

func (r *InMemoryDeveloperRepository) GetAPIKeyByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	k, ok := r.apiKeysByHash[keyHash]
	if !ok {
		return nil, domain.ErrAPIKeyNotFound
	}
	cp := *k
	return &cp, nil
}

func (r *InMemoryDeveloperRepository) RevokeAPIKey(ctx context.Context, keyID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	k, ok := r.apiKeys[keyID]
	if !ok {
		return domain.ErrAPIKeyNotFound
	}
	k.Revoked = true
	return nil
}

func (r *InMemoryDeveloperRepository) ListAPIKeys(ctx context.Context, orgID string) ([]domain.APIKey, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []domain.APIKey
	for _, k := range r.apiKeys {
		if k.OrgID == orgID {
			res = append(res, *k)
		}
	}
	return res, nil
}

func (r *InMemoryDeveloperRepository) SavePayoutRecord(ctx context.Context, rec *domain.PayoutRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *rec
	r.payouts[rec.OrgID] = &cp
	return nil
}

func (r *InMemoryDeveloperRepository) GetPayoutRecord(ctx context.Context, orgID string) (*domain.PayoutRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rec, ok := r.payouts[orgID]
	if !ok {
		return nil, nil
	}
	cp := *rec
	return &cp, nil
}


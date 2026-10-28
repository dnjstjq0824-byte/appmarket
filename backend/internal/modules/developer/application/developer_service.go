package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
)

// DeveloperRepository abstracts persistence.
type DeveloperRepository interface {
	CreateOrg(ctx context.Context, org *domain.DeveloperOrg) error
	GetOrgByID(ctx context.Context, orgID string) (*domain.DeveloperOrg, error)
	GetOrgBySlug(ctx context.Context, slug string) (*domain.DeveloperOrg, error)
	UpdateOrgStatus(ctx context.Context, orgID string, status domain.OrgVerificationStatus) error

	CreateVerification(ctx context.Context, v *domain.DeveloperVerification) error
	GetLatestVerification(ctx context.Context, orgID string) (*domain.DeveloperVerification, error)

	AddMember(ctx context.Context, m *domain.OrgMember) error
	GetMember(ctx context.Context, orgID, userID string) (*domain.OrgMember, error)
	ListMembers(ctx context.Context, orgID string) ([]domain.OrgMember, error)

	CreateAPIKey(ctx context.Context, k *domain.APIKey) error
	GetAPIKeyByHash(ctx context.Context, keyHash string) (*domain.APIKey, error)
	RevokeAPIKey(ctx context.Context, keyID string) error
	ListAPIKeys(ctx context.Context, orgID string) ([]domain.APIKey, error)
}

// OrgService handles org onboarding.
type OrgService struct {
	repo DeveloperRepository
}

func NewOrgService(repo DeveloperRepository) *OrgService {
	return &OrgService{repo: repo}
}

func (s *OrgService) CreateOrg(ctx context.Context, slug, name string) (*domain.DeveloperOrg, error) {
	if slug == "" {
		return nil, domain.ErrInvalidOrgSlug
	}
	existing, _ := s.repo.GetOrgBySlug(ctx, slug)
	if existing != nil {
		return nil, domain.ErrDuplicateOrgSlug
	}
	now := time.Now().UTC()
	org := &domain.DeveloperOrg{
		ID:                 uuid.New().String(),
		OrgSlug:            slug,
		OrgName:            name,
		VerificationStatus: domain.OrgUnverified,
		MFARequired:        false,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := s.repo.CreateOrg(ctx, org); err != nil {
		return nil, err
	}
	return org, nil
}

func (s *OrgService) GetOrg(ctx context.Context, orgID string) (*domain.DeveloperOrg, error) {
	org, err := s.repo.GetOrgByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, domain.ErrOrgNotFound
	}
	return org, nil
}

func (s *OrgService) SubmitVerification(ctx context.Context, orgID, docType, docURL, domainName string) (*domain.DeveloperVerification, error) {
	org, err := s.GetOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	_ = org
	now := time.Now().UTC()
	v := &domain.DeveloperVerification{
		ID:          uuid.New().String(),
		OrgID:       orgID,
		DocType:     docType,
		DocURL:      docURL,
		Domain:      domainName,
		Status:      "PENDING",
		SubmittedAt: now,
	}
	if err := s.repo.CreateVerification(ctx, v); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateOrgStatus(ctx, orgID, domain.OrgPendingReview); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *OrgService) EnsureVerified(ctx context.Context, orgID string) error {
	org, err := s.GetOrg(ctx, orgID)
	if err != nil {
		return err
	}
	if org.VerificationStatus != domain.OrgVerified {
		return domain.ErrOrgNotVerified
	}
	return nil
}

// MemberService handles team membership.
type MemberService struct {
	repo DeveloperRepository
}

func NewMemberService(repo DeveloperRepository) *MemberService {
	return &MemberService{repo: repo}
}

func (s *MemberService) AddMember(ctx context.Context, orgID, userID string, role domain.DevRole) (*domain.OrgMember, error) {
	if !isValidRole(role) {
		return nil, errors.New("invalid role")
	}
	org, err := s.repo.GetOrgByID(ctx, orgID)
	if err != nil || org == nil {
		return nil, domain.ErrOrgNotFound
	}
	m := &domain.OrgMember{OrgID: orgID, UserID: userID, Role: role}
	if err := s.repo.AddMember(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func isValidRole(r domain.DevRole) bool {
	switch r {
	case domain.RoleDevOwner, domain.RoleReleaseManager, domain.RoleListingEditor, domain.RoleFinanceManager, domain.RoleSupportAgent:
		return true
	}
	return false
}

func (s *MemberService) GetMember(ctx context.Context, orgID, userID string) (*domain.OrgMember, error) {
	m, err := s.repo.GetMember(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, domain.ErrMemberNotFound
	}
	return m, nil
}

func (s *MemberService) CheckPermission(ctx context.Context, orgID, userID, perm string) error {
	m, err := s.GetMember(ctx, orgID, userID)
	if err != nil {
		return domain.ErrForbiddenTenantAccess
	}
	if !domain.RoleHasPermission(m.Role, perm) {
		return domain.ErrForbiddenTenantAccess
	}
	return nil
}

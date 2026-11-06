package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/sanitize"
)

// DeveloperRepository abstracts persistence.
type DeveloperRepository interface {
	CreateOrg(ctx context.Context, org *domain.DeveloperOrg) error
	GetOrgByID(ctx context.Context, orgID string) (*domain.DeveloperOrg, error)
	GetOrgBySlug(ctx context.Context, slug string) (*domain.DeveloperOrg, error)
	ListOrgsForUser(ctx context.Context, userID string) ([]domain.DeveloperOrg, error)
	UpdateOrgStatus(ctx context.Context, orgID string, status domain.OrgVerificationStatus) error

	CreateVerification(ctx context.Context, v *domain.DeveloperVerification) error
	GetLatestVerification(ctx context.Context, orgID string) (*domain.DeveloperVerification, error)
	DecideVerificationRecord(ctx context.Context, verificationID, status, reviewerID, notes string) error
	ListVerificationsByStatus(ctx context.Context, status string) ([]domain.DeveloperVerification, error)

	AddMember(ctx context.Context, m *domain.OrgMember) error
	GetMember(ctx context.Context, orgID, userID string) (*domain.OrgMember, error)
	ListMembers(ctx context.Context, orgID string) ([]domain.OrgMember, error)
	RemoveMember(ctx context.Context, orgID, userID string) error
	SavePayoutRecord(ctx context.Context, rec *domain.PayoutRecord) error
	GetPayoutRecord(ctx context.Context, orgID string) (*domain.PayoutRecord, error)

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
		OrgSlug:            sanitize.Strict(slug),
		OrgName:            sanitize.Strict(name),
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

func (s *OrgService) ListOrgsForUser(ctx context.Context, userID string) ([]domain.DeveloperOrg, error) {
	return s.repo.ListOrgsForUser(ctx, userID)
}

func (s *OrgService) GetVerification(ctx context.Context, orgID string) (*domain.DeveloperVerification, error) {
	return s.repo.GetLatestVerification(ctx, orgID)
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
		DocType:     sanitize.Strict(docType),
		DocURL:      sanitize.Strict(docURL),
		Domain:      sanitize.Strict(domainName),
		Status:      string(domain.OrgPendingReview),
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

func (s *OrgService) DecideVerification(ctx context.Context, orgID string, approved bool, notes, reviewerID string) (*domain.DeveloperOrg, error) {
	org, err := s.GetOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	latest, err := s.repo.GetLatestVerification(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if latest == nil {
		return nil, domain.ErrVerificationNotFound
	}
	status := domain.OrgRejected
	rowStatus := "REJECTED"
	if approved {
		status = domain.OrgVerified
		rowStatus = "APPROVED"
	}
	if err := s.repo.DecideVerificationRecord(ctx, latest.ID, rowStatus, reviewerID, notes); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateOrgStatus(ctx, orgID, status); err != nil {
		return nil, err
	}
	org.VerificationStatus = status
	return org, nil
}

func (s *OrgService) ListVerificationsByStatus(ctx context.Context, status string) ([]domain.DeveloperVerification, error) {
	return s.repo.ListVerificationsByStatus(ctx, status)
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

func (s *MemberService) ListMembers(ctx context.Context, orgID string) ([]domain.OrgMember, error) {
	return s.repo.ListMembers(ctx, orgID)
}

func (s *MemberService) RemoveMember(ctx context.Context, orgID, userID string) error {
	m, err := s.GetMember(ctx, orgID, userID)
	if err != nil {
		return err
	}
	if m.Role == domain.RoleDevOwner {
		members, err := s.repo.ListMembers(ctx, orgID)
		if err != nil {
			return err
		}
		owners := 0
		for _, other := range members {
			if other.Role == domain.RoleDevOwner {
				owners++
			}
		}
		if owners <= 1 {
			return domain.ErrLastOwnerRemoval
		}
	}
	return s.repo.RemoveMember(ctx, orgID, userID)
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

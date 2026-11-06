package console

import (
	"context"

	distApp "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
)

// SignatureDTO is the signing lineage entry for a distribution.
type SignatureDTO struct {
	DistributionID string `json:"distributionId"`
	ArtifactID     string `json:"artifactId"`
	CertSHA256     string `json:"certSha256"`
	SchemeVersion  int    `json:"schemeVersion"`
	IsPrimary      bool   `json:"isPrimary"`
}

// LineageDTO pairs releases with their signing facts.
type LineageDTO struct {
	Releases   []ReleaseDTO   `json:"releases"`
	Signatures []SignatureDTO `json:"signatures"`
}

// LineageService exposes release + signing lineage for org-owned apps.
type LineageService struct {
	resolver  *OrgAppResolver
	artifacts *distApp.ArtifactService
}

func NewLineageService(resolver *OrgAppResolver, artifacts *distApp.ArtifactService) *LineageService {
	return &LineageService{resolver: resolver, artifacts: artifacts}
}

func (s *LineageService) AppLineage(ctx context.Context, userID, orgID, appID string) (*LineageDTO, error) {
	_, appID, err := s.resolver.ResolveOrgApp(ctx, userID, orgID, appID)
	if err != nil {
		return nil, err
	}
	releases, err := s.artifacts.ListReleasesByApp(ctx, appID)
	if err != nil {
		return nil, err
	}
	out := &LineageDTO{Releases: make([]ReleaseDTO, 0, len(releases))}
	seenDist := make(map[string]bool)
	for _, r := range releases {
		out.Releases = append(out.Releases, *ToReleaseDTO(r))
		if seenDist[r.DistributionID] {
			continue
		}
		seenDist[r.DistributionID] = true
		if sig, err := s.artifacts.LatestSignature(ctx, r.DistributionID); err == nil && sig != nil {
			out.Signatures = append(out.Signatures, SignatureDTO{
				DistributionID: r.DistributionID, ArtifactID: sig.ArtifactID,
				CertSHA256: sig.CertSHA256Fingerprint, SchemeVersion: sig.SignatureSchemeVersion,
				IsPrimary: sig.IsPrimaryLineage,
			})
		}
	}
	if out.Signatures == nil {
		out.Signatures = []SignatureDTO{}
	}
	return out, nil
}

// Package console aggregates cross-module operations for the developer and
// admin consoles. Modules never import each other (see architecture_test.go);
// this platform package is the sanctioned composition point, wired by
// httpserver and consumed through narrow interfaces defined by each
// presentation package.
package console

import (
	"context"
	"errors"
	"fmt"

	catalogApp "github.com/bookingnetwork/appmarket/backend/internal/modules/catalog/application"
	developerApp "github.com/bookingnetwork/appmarket/backend/internal/modules/developer/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
)

var (
	ErrNoOrg          = errors.New("ERR_NO_ORG: caller owns no developer organization")
	ErrNoApp          = errors.New("ERR_NO_APP: organization has no apps")
	ErrOrgAccessDenied = errors.New("ERR_FORBIDDEN_TENANT_ACCESS")
	ErrAppAccessDenied = errors.New("ERR_FORBIDDEN_APP_ACCESS")
)

// OrgAppResolver maps a caller plus optional "current" handles to concrete
// org and app ids. Empty orgID selects the caller's first org; empty appID
// (or "current") selects the org's most recently updated app.
type OrgAppResolver struct {
	orgs    *developerApp.OrgService
	members *developerApp.MemberService
	catalog *catalogApp.CatalogService
}

func NewOrgAppResolver(orgs *developerApp.OrgService, members *developerApp.MemberService, catalog *catalogApp.CatalogService) *OrgAppResolver {
	return &OrgAppResolver{orgs: orgs, members: members, catalog: catalog}
}

func (r *OrgAppResolver) ResolveOrgApp(ctx context.Context, userID, orgID, appID string) (string, string, error) {
	if orgID == "" || orgID == "current" {
		orgs, err := r.orgs.ListOrgsForUser(ctx, userID)
		if err != nil {
			return "", "", err
		}
		if len(orgs) == 0 {
			return "", "", ErrNoOrg
		}
		orgID = orgs[0].ID
	} else if err := r.checkMembership(ctx, orgID, userID); err != nil {
		return "", "", err
	}
	if appID == "" || appID == "current" {
		apps, err := r.catalog.ListAppsByOrg(ctx, orgID)
		if err != nil {
			return "", "", err
		}
		if len(apps) == 0 {
			return "", "", ErrNoApp
		}
		appID = apps[0].ID
	} else if err := r.checkAppOwnership(ctx, orgID, appID); err != nil {
		return "", "", err
	}
	return orgID, appID, nil
}

func (r *OrgAppResolver) checkMembership(ctx context.Context, orgID, userID string) error {
	_, err := r.members.GetMember(ctx, orgID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			return ErrOrgAccessDenied
		}
		return err
	}
	return nil
}

func (r *OrgAppResolver) checkAppOwnership(ctx context.Context, orgID, appID string) error {
	apps, err := r.catalog.ListAppsByOrg(ctx, orgID)
	if err != nil {
		return err
	}
	for _, a := range apps {
		if a.ID == appID {
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrAppAccessDenied, appID)
}

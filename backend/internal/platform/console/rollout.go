package console

import (
	"context"
	"errors"
	"fmt"

	distApp "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
	distDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
)

var (
	ErrNoRelease       = errors.New("ERR_NO_RELEASE: app has no releases yet")
	ErrRolloutRejected = errors.New("ERR_ROLLOUT_REJECTED")
)

// RolloutService controls staged rollouts for the developer console.
type RolloutService struct {
	artifacts *distApp.ArtifactService
}

func NewRolloutService(artifacts *distApp.ArtifactService) *RolloutService {
	return &RolloutService{artifacts: artifacts}
}

// CurrentRelease returns the newest release for an app.
func (s *RolloutService) CurrentRelease(ctx context.Context, appID string) (*distDomain.AppRelease, error) {
	releases, err := s.artifacts.ListReleasesByApp(ctx, appID)
	if err != nil {
		return nil, err
	}
	if len(releases) == 0 {
		return nil, ErrNoRelease
	}
	return releases[0], nil
}

// UpdateRollout raises the staged percentage of the current release.
func (s *RolloutService) UpdateRollout(ctx context.Context, appID string, percentage int) (*distDomain.AppRelease, error) {
	current, err := s.CurrentRelease(ctx, appID)
	if err != nil {
		return nil, err
	}
	if current.ReleaseStatus == distDomain.ReleaseHalted {
		return nil, fmt.Errorf("%w: resume the halted release first", ErrRolloutRejected)
	}
	if percentage <= current.RolloutPercentage {
		return nil, fmt.Errorf("%w: percentage must increase beyond %d", ErrRolloutRejected, current.RolloutPercentage)
	}
	return s.artifacts.UpdateStagedRollout(ctx, current.ID, percentage)
}

// Halt suspends the current release immediately.
func (s *RolloutService) Halt(ctx context.Context, appID, reason string) (*distDomain.AppRelease, error) {
	current, err := s.CurrentRelease(ctx, appID)
	if err != nil {
		return nil, err
	}
	return s.artifacts.HaltRelease(ctx, current.ID, reason)
}

// Resume returns a halted release to publication.
func (s *RolloutService) Resume(ctx context.Context, appID string) (*distDomain.AppRelease, error) {
	current, err := s.CurrentRelease(ctx, appID)
	if err != nil {
		return nil, err
	}
	if current.ReleaseStatus != distDomain.ReleaseHalted {
		return nil, fmt.Errorf("%w: release is not halted", ErrRolloutRejected)
	}
	return s.artifacts.ResumeRelease(ctx, current.ID)
}

package console

import (
	"context"
	"errors"
	"fmt"

	distApp "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
	distDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
)

var (
	ErrNoRelease          = errors.New("ERR_NO_RELEASE: app has no releases yet")
	ErrRolloutRejected    = errors.New("ERR_ROLLOUT_REJECTED")
	ErrReleaseNotReady    = errors.New("ERR_RELEASE_NOT_READY")
	ErrReleasePublishFail = errors.New("ERR_RELEASE_PUBLISH_FAILED")
)

// RolloutService controls staged rollouts for the developer console.
type RolloutService struct {
	artifacts *distApp.ArtifactService
	publisher *distApp.CDNPublisher
}

func NewRolloutService(artifacts *distApp.ArtifactService) *RolloutService {
	return &RolloutService{artifacts: artifacts}
}

// WithCDNPublisher enables artifact promotion on publish. Without it Publish
// fails closed instead of marking releases live with no bytes behind them.
func (s *RolloutService) WithCDNPublisher(publisher *distApp.CDNPublisher) *RolloutService {
	s.publisher = publisher
	return s
}

// SubmitForReview sends the current release to the review queue (Play Console
// "send for review").
func (s *RolloutService) SubmitForReview(ctx context.Context, appID string) (*distDomain.AppRelease, error) {
	current, err := s.CurrentRelease(ctx, appID)
	if err != nil {
		return nil, err
	}
	updated, err := s.artifacts.SubmitReleaseForReview(ctx, current.ID)
	if err != nil {
		if errors.Is(err, distDomain.ErrInvalidStateTransition) {
			return nil, ErrReleaseNotReady
		}
		return nil, err
	}
	return updated, nil
}

// PublishResult carries the published release plus its public CDN URLs.
type PublishResult struct {
	Release  *distDomain.AppRelease `json:"release"`
	CDNURLs  []string               `json:"cdnUrls"`
	Rollout  int                    `json:"rolloutPercentage"`
	FullRoll bool                   `json:"full"`
}

// Publish rolls out the current approved release (Play Console rollout):
// staged (1-99%, default 20) or full. Artifacts promote to the public CDN
// before the release goes live; any promotion failure aborts the publish.
func (s *RolloutService) Publish(ctx context.Context, appID string, full bool, percentage int) (*PublishResult, error) {
	current, err := s.CurrentRelease(ctx, appID)
	if err != nil {
		return nil, err
	}
	updated, urls, err := s.artifacts.PublishRelease(ctx, current.ID, full, percentage, s.publisher)
	if err != nil {
		if errors.Is(err, distDomain.ErrInvalidStateTransition) || errors.Is(err, distDomain.ErrRolloutPercentageInvalid) {
			return nil, ErrReleaseNotReady
		}
		return nil, ErrReleasePublishFail
	}
	return &PublishResult{Release: updated, CDNURLs: urls, Rollout: updated.RolloutPercentage, FullRoll: full}, nil
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

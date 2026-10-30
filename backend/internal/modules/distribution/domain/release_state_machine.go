package domain

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

var validTransitions = map[ReleaseStatus][]ReleaseStatus{
	ReleaseDraft: {
		ReleaseArtifactVerifying,
	},
	ReleaseArtifactVerifying: {
		ReleaseSecurityScanning,
		ReleaseRejected,
	},
	ReleaseSecurityScanning: {
		ReleaseReadyForSubmission,
		ReleaseScanFailedQuarantined,
	},
	ReleaseScanFailedQuarantined: {
		ReleaseDraft, // developer can fix/re-upload
	},
	ReleaseReadyForSubmission: {
		ReleaseSubmittedForReview,
	},
	ReleaseSubmittedForReview: {
		ReleaseInReview,
	},
	ReleaseInReview: {
		ReleaseApprovedPendingPub,
		ReleaseChangesRequested,
		ReleaseRejected,
	},
	ReleaseChangesRequested: {
		ReleaseSubmittedForReview,
		ReleaseDraft,
	},
	ReleaseApprovedPendingPub: {
		ReleasePublishedStaged,
		ReleasePublishedFull,
	},
	ReleasePublishedStaged: {
		ReleasePublishedFull,
		ReleaseHalted,
		ReleaseRevokedByAdmin,
	},
	ReleaseHalted: {
		ReleasePublishedStaged,
		ReleasePublishedFull,
		ReleaseRevokedByAdmin,
	},
	ReleasePublishedFull: {
		ReleaseSuperseded,
		ReleaseRevokedByAdmin,
	},
	ReleaseRejected:       {},
	ReleaseSuperseded:     {},
	ReleaseRevokedByAdmin: {},
}

// CanTransition checks whether fromStatus -> toStatus is legally permitted in the 12-state release state machine.
func CanTransition(from, to ReleaseStatus) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, target := range allowed {
		if target == to {
			return true
		}
	}
	return false
}

// TransitionRelease transitions the release status or returns ErrInvalidStateTransition.
func TransitionRelease(rel *AppRelease, newStatus ReleaseStatus) error {
	if !CanTransition(rel.ReleaseStatus, newStatus) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidStateTransition, rel.ReleaseStatus, newStatus)
	}
	rel.ReleaseStatus = newStatus
	return nil
}

// ComputeStagedRolloutBucket computes a deterministic bucket in [0, 99] for a device and release.
// Monotonicity property: If a device qualifies for percentage P1, it will qualify for any P2 > P1.
func ComputeStagedRolloutBucket(releaseID, deviceID string) int {
	input := fmt.Sprintf("%s:%s", releaseID, deviceID)
	sum := sha256.Sum256([]byte(input))
	val := binary.BigEndian.Uint64(sum[:8])
	return int(val % 100)
}

// IsDeviceEligibleForRollout checks if a device is included in a staged rollout percentage.
func IsDeviceEligibleForRollout(releaseID, deviceID string, rolloutPercentage int) bool {
	if rolloutPercentage >= 100 {
		return true
	}
	if rolloutPercentage <= 0 {
		return false
	}
	bucket := ComputeStagedRolloutBucket(releaseID, deviceID)
	return bucket < rolloutPercentage
}


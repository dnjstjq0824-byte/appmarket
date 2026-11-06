package console

import (
	"time"

	distDomain "github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
)

// UploadSessionDTO is the quarantine session shape served to consoles.
type UploadSessionDTO struct {
	UploadID  string    `json:"uploadId"`
	AppID     string    `json:"appId"`
	OrgID     string    `json:"orgId"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func ToUploadSessionDTO(s *distDomain.UploadSession) *UploadSessionDTO {
	return &UploadSessionDTO{UploadID: s.ID, AppID: s.AppID, OrgID: s.OrgID, Status: s.Status, ExpiresAt: s.ExpiresAt}
}

// ReleaseDTO is the release shape served to consoles.
type ReleaseDTO struct {
	ID                string     `json:"id"`
	AppID             string     `json:"appId"`
	VersionCode       int64      `json:"versionCode"`
	VersionName       string     `json:"versionName"`
	Status            string     `json:"status"`
	RolloutPercentage int        `json:"rolloutPercentage"`
	HaltReason        string     `json:"haltReason,omitempty"`
	HaltedAt          *time.Time `json:"haltedAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

func ToReleaseDTO(r *distDomain.AppRelease) *ReleaseDTO {
	out := &ReleaseDTO{
		ID: r.ID, AppID: r.AppID, VersionCode: r.VersionCode, VersionName: r.VersionName,
		Status: string(r.ReleaseStatus), RolloutPercentage: r.RolloutPercentage,
		HaltedAt: r.HaltedAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
	if r.HaltReason != nil {
		out.HaltReason = *r.HaltReason
	}
	return out
}

// CompleteUploadDTO is the upload-completion receipt.
type CompleteUploadDTO struct {
	UploadID string         `json:"uploadId"`
	Status   string         `json:"status"`
	Release  *ReleaseDTO    `json:"release,omitempty"`
	Details  *BinaryDetails `json:"details,omitempty"`
}

// ToCompleteUploadDTO maps a completion receipt to its console shape.
func ToCompleteUploadDTO(res *CompleteResult) *CompleteUploadDTO {
	out := &CompleteUploadDTO{UploadID: res.UploadID, Status: res.Status, Details: res.Details}
	if res.Release != nil {
		out.Release = ToReleaseDTO(res.Release)
	}
	return out
}

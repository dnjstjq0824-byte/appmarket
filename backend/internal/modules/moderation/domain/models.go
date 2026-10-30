package domain

import (
	"errors"
	"time"
)

var (
	ErrReviewRequiresInstallation   = errors.New("ERR_REVIEW_REQUIRES_INSTALLATION")
	ErrAppealSameReviewerForbidden  = errors.New("ERR_APPEAL_SAME_REVIEWER_FORBIDDEN")
	ErrReviewNotFound               = errors.New("ERR_REVIEW_NOT_FOUND")
	ErrAppealNotFound               = errors.New("ERR_APPEAL_NOT_FOUND")
	ErrInvalidStarRating            = errors.New("ERR_INVALID_STAR_RATING")
)

type ReviewStatus string

const (
	ReviewPublished        ReviewStatus = "PUBLISHED"
	ReviewFlaggedSuspicious ReviewStatus = "FLAGGED_SUSPICIOUS"
	ReviewQuarantined      ReviewStatus = "QUARANTINED"
)

type DeveloperReply struct {
	AuthorID  string    `json:"author_id"`
	Content   string    `json:"content"`
	RepliedAt time.Time `json:"replied_at"`
}

type UserReview struct {
	ID                string          `json:"id"`
	AppID             string          `json:"app_id"`
	UserID            string          `json:"user_id"`
	Rating            int             `json:"rating"` // 1 to 5
	Title             string          `json:"title"`
	Comment           string          `json:"comment"`
	VersionCode       int64           `json:"version_code"`
	IsVerifiedInstall bool            `json:"is_verified_install"`
	FraudScore        float64         `json:"fraud_score"` // 0.0 to 1.0
	Status            ReviewStatus    `json:"status"`
	DeveloperReply    *DeveloperReply `json:"developer_reply,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
}

type AppealStatus string

const (
	AppealPending  AppealStatus = "PENDING"
	AppealApproved AppealStatus = "APPROVED"
	AppealRejected AppealStatus = "REJECTED"
)

type DeveloperAppeal struct {
	ID                 string       `json:"id"`
	OrgID              string       `json:"org_id"`
	ReleaseID          string       `json:"release_id"`
	Reason             string       `json:"reason"`
	OriginalReviewerID string       `json:"original_reviewer_id"`
	AssignedReviewerID string       `json:"assigned_reviewer_id,omitempty"`
	Status             AppealStatus `json:"status"`
	DecisionNotes      string       `json:"decision_notes,omitempty"`
	CreatedAt          time.Time    `json:"created_at"`
	DecidedAt          *time.Time   `json:"decided_at,omitempty"`
}

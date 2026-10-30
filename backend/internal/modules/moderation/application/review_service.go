package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/domain"
)

type InstallVerifier interface {
	IsUserVerifiedInstaller(ctx context.Context, userID, appID string) bool
}

type InMemoryInstallVerifier struct {
	mu           sync.RWMutex
	installations map[string]bool // key: userID + ":" + appID
}

func NewInMemoryInstallVerifier() *InMemoryInstallVerifier {
	return &InMemoryInstallVerifier{
		installations: make(map[string]bool),
	}
}

func (v *InMemoryInstallVerifier) RecordInstallation(userID, appID string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.installations[userID+":"+appID] = true
}

func (v *InMemoryInstallVerifier) IsUserVerifiedInstaller(ctx context.Context, userID, appID string) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.installations[userID+":"+appID]
}

type ReviewService struct {
	verifier   InstallVerifier
	detector   *AntiFraudDetector
	mu         sync.RWMutex
	reviews    map[string]*domain.UserReview
	appReviews map[string][]*domain.UserReview
}

func NewReviewService(verifier InstallVerifier, detector *AntiFraudDetector) *ReviewService {
	return &ReviewService{
		verifier:   verifier,
		detector:   detector,
		reviews:    make(map[string]*domain.UserReview),
		appReviews: make(map[string][]*domain.UserReview),
	}
}

type CreateReviewInput struct {
	AppID           string
	UserID          string
	Rating          int
	Title           string
	Comment         string
	VersionCode     int64
	ClientIP        string
	AccountAgeHours float64
}

// PostReview mandates verified user installation before accepting user review (TC-SEC-008)
func (s *ReviewService) PostReview(ctx context.Context, in CreateReviewInput) (*domain.UserReview, error) {
	if in.Rating < 1 || in.Rating > 5 {
		return nil, domain.ErrInvalidStarRating
	}

	// 1. Mandatory Verified Installation Verification
	if !s.verifier.IsUserVerifiedInstaller(ctx, in.UserID, in.AppID) {
		return nil, domain.ErrReviewRequiresInstallation
	}

	// 2. Anti-Fraud Analysis
	fraudScore, status := s.detector.EvaluateReview(FraudEvaluationParams{
		UserID:          in.UserID,
		ClientIP:        in.ClientIP,
		AccountAgeHours: in.AccountAgeHours,
		Rating:          in.Rating,
		Comment:         in.Comment,
	})

	now := time.Now().UTC()
	revID := fmt.Sprintf("rev-%d", now.UnixNano())
	review := &domain.UserReview{
		ID:                revID,
		AppID:             in.AppID,
		UserID:            in.UserID,
		Rating:            in.Rating,
		Title:             in.Title,
		Comment:           in.Comment,
		VersionCode:       in.VersionCode,
		IsVerifiedInstall: true,
		FraudScore:        fraudScore,
		Status:            status,
		CreatedAt:         now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.reviews[revID] = review
	s.appReviews[in.AppID] = append(s.appReviews[in.AppID], review)

	return review, nil
}

func (s *ReviewService) GetAppReviews(appID string) []*domain.UserReview {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.appReviews[appID]
}


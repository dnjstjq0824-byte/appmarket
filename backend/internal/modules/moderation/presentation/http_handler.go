package presentation

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/moderation/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/platform/errors"
)

// IdentityResolver must return a user ID from verified authentication only.
type IdentityResolver func(*http.Request) (string, error)

// ModerationHandler serves review, report, and appeal endpoints.
type ModerationHandler struct {
	reviews *application.ReviewService
	reports *application.ReportService
	appeals *application.AppealService
	identity IdentityResolver
}

func NewModerationHandler(reviews *application.ReviewService, reports *application.ReportService, appeals *application.AppealService, identity IdentityResolver) *ModerationHandler {
	return &ModerationHandler{reviews: reviews, reports: reports, appeals: appeals, identity: identity}
}

func (h *ModerationHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/reports", h.handleFileReport)
	r.Get("/apps/{appId}/reviews", h.handleListReviews)
	r.Group(func(protected chi.Router) {
		protected.Use(h.requireAuth)
		protected.Post("/reviews", h.handlePostReview)
		protected.Post("/appeals", h.handleSubmitAppeal)
	})
	return r
}

func (h *ModerationHandler) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.identity == nil {
			errors.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Authentication required", r.URL.Path)
			return
		}
		if userID, err := h.identity(r); err != nil || userID == "" {
			errors.WriteProblem(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Authentication required", r.URL.Path)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type fileReportReq struct {
	AppIdentifier  string `json:"appIdentifier"`
	TargetType     string `json:"targetType"`
	Category       string `json:"category"`
	ReporterEmail  string `json:"reporterEmail"`
	Description    string `json:"description"`
	ConfirmAccuracy bool  `json:"confirmAccuracy"`
}

func (h *ModerationHandler) handleFileReport(w http.ResponseWriter, r *http.Request) {
	var req fileReportReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	if !req.ConfirmAccuracy {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_REPORT_INVALID", "Accuracy confirmation is required", r.URL.Path)
		return
	}
	reporterUserID := ""
	if h.identity != nil {
		reporterUserID, _ = h.identity(r)
	}
	rep, err := h.reports.FileReport(r.Context(), application.FileReportInput{
		ReporterUserID: reporterUserID,
		ReporterEmail:  req.ReporterEmail,
		TargetType:     req.TargetType,
		TargetID:       req.AppIdentifier,
		Category:       req.Category,
		Description:    req.Description,
	})
	if err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_REPORT_INVALID", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"ticketId": rep.ID, "priority": rep.Priority, "status": rep.Status})
}

func (h *ModerationHandler) handleListReviews(w http.ResponseWriter, r *http.Request) {
	list := h.reviews.GetAppReviews(chi.URLParam(r, "appId"))
	out := make([]domain.UserReview, 0, len(list))
	for _, rev := range list {
		if rev != nil && rev.Status == domain.ReviewPublished {
			out = append(out, *rev)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

type postReviewReq struct {
	AppID       string `json:"appId"`
	Rating      int    `json:"rating"`
	Title       string `json:"title"`
	Comment     string `json:"comment"`
	VersionCode int64  `json:"versionCode"`
}

func (h *ModerationHandler) handlePostReview(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.identity(r)
	var req postReviewReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	rev, err := h.reviews.PostReview(r.Context(), application.CreateReviewInput{
		AppID:       req.AppID,
		UserID:      userID,
		Rating:      req.Rating,
		Title:       req.Title,
		Comment:     req.Comment,
		VersionCode: req.VersionCode,
		ClientIP:    r.RemoteAddr,
	})
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "ERR_REVIEW_REQUIRES_INSTALLATION") {
			status = http.StatusForbidden
		}
		errors.WriteProblem(w, status, "ERR_REVIEW_REJECTED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(rev)
}

type submitAppealReq struct {
	OrgID              string `json:"orgId"`
	ReleaseID          string `json:"releaseId"`
	Reason             string `json:"reason"`
	OriginalReviewerID string `json:"originalReviewerId"`
}

func (h *ModerationHandler) handleSubmitAppeal(w http.ResponseWriter, r *http.Request) {
	var req submitAppealReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_INVALID_BODY", err.Error(), r.URL.Path)
		return
	}
	appeal, err := h.appeals.SubmitAppeal(r.Context(), req.OrgID, req.ReleaseID, req.Reason, req.OriginalReviewerID)
	if err != nil {
		errors.WriteProblem(w, http.StatusBadRequest, "ERR_APPEAL_FAILED", err.Error(), r.URL.Path)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(appeal)
}

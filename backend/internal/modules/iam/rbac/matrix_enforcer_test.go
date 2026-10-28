package rbac_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/rbac"
)

// TC-SEC-001: Fine-grained RBAC permission matrix enforcement
func TestRBACPermissionMatrixEnforcement(t *testing.T) {
	jwtMgr, err := application.NewJWTManager(nil, "https://api.appmarket.bookingnetwork.internal")
	if err != nil {
		t.Fatalf("failed to create jwt manager: %v", err)
	}

	enforcer := rbac.NewMatrixEnforcer(jwtMgr)

	// Protected handler requiring "admin:killswitch:execute"
	killSwitchHandler := enforcer.AuthenticateMiddleware(
		enforcer.RequirePermissionMiddleware("admin:killswitch:execute")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("executed killswitch"))
			}),
		),
	)

	// 1. Regular USER token -> must be rejected with 403 Forbidden
	userToken, err := jwtMgr.GenerateAccessToken("user-1", "user@test.com", []string{"USER"}, nil, false)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	reqUser := httptest.NewRequest(http.MethodPost, "/api/v1/admin/killswitch", nil)
	reqUser.Header.Set("Authorization", "Bearer "+userToken)
	recUser := httptest.NewRecorder()
	killSwitchHandler.ServeHTTP(recUser, reqUser)

	if recUser.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for regular user, got %d", recUser.Code)
	}

	// 2. ADMIN_SECURITY token -> must succeed with 200 OK
	secToken, err := jwtMgr.GenerateAccessToken("sec-admin-1", "sec@test.com", []string{"ADMIN_SECURITY"}, nil, true)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	reqSec := httptest.NewRequest(http.MethodPost, "/api/v1/admin/killswitch", nil)
	reqSec.Header.Set("Authorization", "Bearer "+secToken)
	recSec := httptest.NewRecorder()
	killSwitchHandler.ServeHTTP(recSec, reqSec)

	if recSec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for ADMIN_SECURITY, got %d", recSec.Code)
	}

	// 3. ADMIN_SUPER token (wildcard "*") -> must succeed with 200 OK
	superToken, err := jwtMgr.GenerateAccessToken("super-admin-1", "super@test.com", []string{"ADMIN_SUPER"}, nil, true)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	reqSuper := httptest.NewRequest(http.MethodPost, "/api/v1/admin/killswitch", nil)
	reqSuper.Header.Set("Authorization", "Bearer "+superToken)
	recSuper := httptest.NewRecorder()
	killSwitchHandler.ServeHTTP(recSuper, reqSuper)

	if recSuper.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for ADMIN_SUPER, got %d", recSuper.Code)
	}
}

// TC-SEC-006: Dual-Control (Four-Eyes Principle) Self-Approval Prevention
func TestDualControlSelfApprovalForbidden(t *testing.T) {
	ctx := context.Background()
	engine := rbac.NewDualControlEngine(nil)

	adminAliceID := "admin-alice-1111"
	adminBobID := "admin-bob-2222"

	// 1. Admin Alice initiates a high-value settlement payout request
	req, err := engine.CreateRequest(
		ctx,
		"SETTLEMENT_PAYOUT",
		adminAliceID,
		"settlement-txn-9999",
		`{"amount_cents": 5000000, "currency": "USD"}`,
		"Monthly developer revenue share payout over threshold",
	)
	if err != nil {
		t.Fatalf("failed to create dual approval request: %v", err)
	}
	if req.Status != domain.ApprovalPending {
		t.Fatalf("expected pending status, got %s", req.Status)
	}

	// 2. SELF-APPROVAL ATTEMPT: Admin Alice tries to approve her own request!
	_, err = engine.Approve(ctx, req.ID, adminAliceID)
	if err == nil {
		t.Fatalf("expected self-approval to be forbidden, but it succeeded")
	}
	if err != domain.ErrDualControlSelfApproval {
		t.Fatalf("expected ErrDualControlSelfApproval, got %v", err)
	}

	// 3. LEGITIMATE SECOND ADMIN APPROVAL: Admin Bob approves Alice's request
	approvedReq, err := engine.Approve(ctx, req.ID, adminBobID)
	if err != nil {
		t.Fatalf("expected approval by distinct admin Bob to succeed, got %v", err)
	}
	if approvedReq.Status != domain.ApprovalApproved {
		t.Fatalf("expected APPROVED status, got %s", approvedReq.Status)
	}
	if approvedReq.ApproverID == nil || *approvedReq.ApproverID != adminBobID {
		t.Fatalf("expected approver ID to be %s", adminBobID)
	}

	// 4. Double decision attempt on closed request must fail
	_, err = engine.Approve(ctx, req.ID, "admin-charlie-3333")
	if err != domain.ErrDualApprovalAlreadyClosed {
		t.Fatalf("expected ErrDualApprovalAlreadyClosed, got %v", err)
	}
}


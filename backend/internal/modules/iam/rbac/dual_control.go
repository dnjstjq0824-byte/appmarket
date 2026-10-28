package rbac

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/iam/domain"
	"github.com/google/uuid"
)

type DualApprovalRepository interface {
	CreateRequest(ctx context.Context, req *domain.DualApprovalRequest) error
	GetRequestByID(ctx context.Context, id string) (*domain.DualApprovalRequest, error)
	UpdateRequest(ctx context.Context, req *domain.DualApprovalRequest) error
}

type InMemoryDualApprovalRepo struct {
	mu       sync.RWMutex
	requests map[string]*domain.DualApprovalRequest
}

func NewInMemoryDualApprovalRepo() *InMemoryDualApprovalRepo {
	return &InMemoryDualApprovalRepo{
		requests: make(map[string]*domain.DualApprovalRequest),
	}
}

func (r *InMemoryDualApprovalRepo) CreateRequest(ctx context.Context, req *domain.DualApprovalRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests[req.ID] = req
	return nil
}

func (r *InMemoryDualApprovalRepo) GetRequestByID(ctx context.Context, id string) (*domain.DualApprovalRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	req, ok := r.requests[id]
	if !ok {
		return nil, fmt.Errorf("approval request not found")
	}
	// return copy
	c := *req
	return &c, nil
}

func (r *InMemoryDualApprovalRepo) UpdateRequest(ctx context.Context, req *domain.DualApprovalRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests[req.ID] = req
	return nil
}

type DualControlEngine struct {
	repo DualApprovalRepository
	mu   sync.Mutex
}

func NewDualControlEngine(repo DualApprovalRepository) *DualControlEngine {
	if repo == nil {
		repo = NewInMemoryDualApprovalRepo()
	}
	return &DualControlEngine{repo: repo}
}

// CreateRequest submits a new sensitive operation requiring second admin approval.
func (e *DualControlEngine) CreateRequest(
	ctx context.Context,
	operationType string,
	initiatorID string,
	targetResourceID string,
	payloadJSON string,
	reason string,
) (*domain.DualApprovalRequest, error) {
	if initiatorID == "" {
		return nil, fmt.Errorf("initiator ID cannot be empty")
	}

	req := &domain.DualApprovalRequest{
		ID:               uuid.New().String(),
		OperationType:    operationType,
		InitiatorID:      initiatorID,
		TargetResourceID: targetResourceID,
		PayloadJSON:      payloadJSON,
		Status:           domain.ApprovalPending,
		Reason:           reason,
		CreatedAt:        time.Now().UTC(),
	}

	if err := e.repo.CreateRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to create approval request: %w", err)
	}

	return req, nil
}

// Approve processes approval from a second administrator (TC-SEC-006).
func (e *DualControlEngine) Approve(ctx context.Context, requestID, approverID string) (*domain.DualApprovalRequest, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	req, err := e.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	// CRITICAL SECURITY RULE: The initiator CANNOT approve their own request!
	if req.InitiatorID == approverID {
		return nil, domain.ErrDualControlSelfApproval
	}

	if req.Status != domain.ApprovalPending {
		return nil, domain.ErrDualApprovalAlreadyClosed
	}

	now := time.Now().UTC()
	req.Status = domain.ApprovalApproved
	req.ApproverID = &approverID
	req.DecidedAt = &now

	if err := e.repo.UpdateRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to update approval request: %w", err)
	}

	return req, nil
}

// Reject rejects the pending dual-control request.
func (e *DualControlEngine) Reject(ctx context.Context, requestID, approverID, reason string) (*domain.DualApprovalRequest, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	req, err := e.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if req.InitiatorID == approverID {
		return nil, domain.ErrDualControlSelfApproval
	}

	if req.Status != domain.ApprovalPending {
		return nil, domain.ErrDualApprovalAlreadyClosed
	}

	now := time.Now().UTC()
	req.Status = domain.ApprovalRejected
	req.ApproverID = &approverID
	req.DecidedAt = &now
	req.Reason = reason

	if err := e.repo.UpdateRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to update approval request: %w", err)
	}

	return req, nil
}


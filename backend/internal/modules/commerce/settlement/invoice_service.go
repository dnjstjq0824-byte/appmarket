package settlement

import (
	"fmt"
	"sync"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/commerce/domain"
)

type InvoiceService struct {
	mu      sync.RWMutex
	byID    map[string]*domain.Invoice
	byOrder map[string]*domain.Invoice
	seq     int64
}

func NewInvoiceService() *InvoiceService {
	return &InvoiceService{
		byID:    make(map[string]*domain.Invoice),
		byOrder: make(map[string]*domain.Invoice),
	}
}

// IssueFromOrder creates the buyer invoice for a paid order. It is idempotent
// per order: repeated calls return the original invoice.
func (s *InvoiceService) IssueFromOrder(order *domain.Order) *domain.Invoice {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.byOrder[order.ID]; ok {
		return existing
	}
	now := time.Now().UTC()
	s.seq++
	inv := &domain.Invoice{
		ID:               fmt.Sprintf("inv-%d", now.UnixNano()),
		InvoiceNumber:    fmt.Sprintf("INV-%04d-%06d", now.Year(), s.seq),
		OrderID:          order.ID,
		UserID:           order.UserID,
		AppID:            order.AppID,
		OrgID:            order.OrgID,
		Currency:         order.Currency,
		GrossAmount:      order.GrossAmount,
		TaxAmount:        order.TaxAmount,
		CommissionAmount: order.CommissionAmount,
		DevPayoutAmount:  order.DevPayoutAmount,
		PGPaymentID:      order.PGPaymentID,
		Status:           domain.InvoiceIssued,
		IssuedAt:         now,
	}
	s.byID[inv.ID] = inv
	s.byOrder[order.ID] = inv
	return inv
}

// MarkRefunded flips the order invoice to REFUNDED. Missing invoices are a
// no-op so refunds issued before invoicing existed stay valid.
func (s *InvoiceService) MarkRefunded(orderID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if inv, ok := s.byOrder[orderID]; ok {
		inv.Status = domain.InvoiceRefunded
	}
}

func (s *InvoiceService) Get(invoiceID string) *domain.Invoice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.byID[invoiceID]
}

func (s *InvoiceService) ListByUser(userID string) []domain.Invoice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Invoice
	for _, inv := range s.byID {
		if inv.UserID == userID {
			out = append(out, *inv)
		}
	}
	return out
}

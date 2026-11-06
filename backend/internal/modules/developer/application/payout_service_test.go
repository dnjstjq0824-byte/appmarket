package application_test

import (
	"context"
	"testing"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/developer/domain"
)

func testPayoutKey(t *testing.T) []byte {
	t.Helper()
	return []byte("0123456789abcdef0123456789abcdef")
}

func TestPayoutAccountRoundtrip(t *testing.T) {
	orgs, _, _, repo := setupDev(t)
	ctx := context.Background()
	org, err := orgs.CreateOrg(ctx, "payoutorg", "Payout Org")
	if err != nil {
		t.Fatalf("create org failed: %v", err)
	}
	svc, err := application.NewPayoutService(repo, testPayoutKey(t))
	if err != nil {
		t.Fatalf("service build failed: %v", err)
	}
	saved, err := svc.SavePayoutAccount(ctx, org.ID, application.PayoutAccountInput{
		Currency: "KRW", BankName: "Shinhan Bank", HolderName: "Acme Inc",
		AccountNumber: "110-123-892401", RoutingCode: "088",
	})
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if saved.AccountLast4 != "2401" {
		t.Fatalf("expected masked last4 2401, got %q", saved.AccountLast4)
	}
	if saved.AccountNumber != "" {
		t.Fatal("plaintext account number must never be returned")
	}
	got, err := svc.GetPayoutAccount(ctx, org.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.BankName != "Shinhan Bank" || got.AccountLast4 != "2401" {
		t.Fatalf("unexpected account: %+v", got)
	}
	raw, err := repo.GetPayoutCiphertext(ctx, org.ID)
	if err != nil {
		t.Fatalf("ciphertext read failed: %v", err)
	}
	for _, needle := range []string{"110-123-892401", "892401"} {
		for i := 0; i+len(needle) <= len(raw); i++ {
			if string(raw[i:i+len(needle)]) == needle {
				t.Fatalf("plaintext %q leaked in stored ciphertext", needle)
			}
		}
	}
	if _, err := svc.GetPayoutAccount(ctx, "missing-org"); err != domain.ErrPayoutAccountNotFound {
		t.Fatalf("expected not-found, got %v", err)
	}
}

func TestPayoutAccountValidation(t *testing.T) {
	_, _, _, repo := setupDev(t)
	svc, err := application.NewPayoutService(repo, testPayoutKey(t))
	if err != nil {
		t.Fatalf("service build failed: %v", err)
	}
	if _, err := application.NewPayoutService(repo, []byte("short")); err == nil {
		t.Fatal("expected short key rejection, got nil")
	}
	ctx := context.Background()
	for _, tc := range []struct {
		name  string
		input application.PayoutAccountInput
	}{
		{"bad currency", application.PayoutAccountInput{Currency: "W0N", BankName: "B", HolderName: "H", AccountNumber: "123456"}},
		{"empty bank", application.PayoutAccountInput{Currency: "KRW", HolderName: "H", AccountNumber: "123456"}},
		{"empty holder", application.PayoutAccountInput{Currency: "KRW", BankName: "B", AccountNumber: "123456"}},
		{"short account", application.PayoutAccountInput{Currency: "KRW", BankName: "B", HolderName: "H", AccountNumber: "123"}},
	} {
		if _, err := svc.SavePayoutAccount(ctx, "org-x", tc.input); err == nil {
			t.Fatalf("%s: expected validation error, got nil", tc.name)
		}
	}
}

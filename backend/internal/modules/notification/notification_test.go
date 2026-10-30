package notification_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/notification"
)

type mockExternalClient struct {
	failEmail bool
	failPush  bool
	sentEmails []string
	sentPushes []string
}

func (m *mockExternalClient) SendEmail(ctx context.Context, recipient, subject, body string) error {
	if m.failEmail {
		return errors.New("external email provider 500 error")
	}
	m.sentEmails = append(m.sentEmails, recipient)
	return nil
}

func (m *mockExternalClient) SendPush(ctx context.Context, deviceToken, title, body string) error {
	if m.failPush {
		return errors.New("FCM push network timeout")
	}
	m.sentPushes = append(m.sentPushes, deviceToken)
	return nil
}

func TestNotification_TC_OPS_006_ResilienceAndCircuitBreaker(t *testing.T) {
	ctx := context.Background()
	repo := notification.NewMemoryNotificationRepository()
	extClient := &mockExternalClient{}
	cb := notification.NewCircuitBreaker(3, 100*time.Millisecond)
	svc := notification.NewNotificationService(repo, extClient, cb)

	userID := "user-001"

	// 1. Successful sending
	n1 := notification.Notification{
		ID:       "notif-1",
		UserID:   userID,
		Category: notification.CategorySecurity,
		Title:    "Password Changed",
		Body:     "Your account password was updated.",
	}
	res1, err := svc.Send(ctx, n1, "user@example.com", "fcm-token-1")
	if err != nil {
		t.Fatalf("unexpected error on valid send: %v", err)
	}
	if !res1.InAppSaved || !res1.ExternalSent {
		t.Errorf("expected both in-app and external sent to succeed")
	}

	// 2. Simulate external messaging outage
	extClient.failEmail = true
	extClient.failPush = true

	// 3 consecutive failures trigger circuit breaker OPEN
	for i := 2; i <= 4; i++ {
		n := notification.Notification{
			ID:       "notif-" + string(rune('0'+i)),
			UserID:   userID,
			Category: notification.CategoryBilling,
			Title:    "Payment Receipt",
			Body:     "Receipt for order.",
		}
		res, err := svc.Send(ctx, n, "user@example.com", "")
		if err != nil {
			t.Fatalf("Send must not return hard fatal error, should gracefully save in-app: %v", err)
		}
		if !res.InAppSaved {
			t.Errorf("in-app notification MUST be saved even during external failure")
		}
		if res.ExternalSent {
			t.Errorf("external delivery should have failed")
		}
	}

	// Verify Circuit is now OPEN
	if cb.GetState() != notification.StateOpen {
		t.Fatalf("expected circuit breaker to be OPEN, got: %s", cb.GetState())
	}

	// 3. 5th request while OPEN should fail fast externally, but IN-APP notification must STILL be saved!
	n5 := notification.Notification{
		ID:       "notif-5",
		UserID:   userID,
		Category: notification.CategorySystem,
		Title:    "Emergency Security Alert",
		Body:     "Critical update available.",
	}
	res5, err := svc.Send(ctx, n5, "user@example.com", "")
	if err != nil {
		t.Fatalf("unexpected error on send with open circuit: %v", err)
	}
	if !res5.InAppSaved {
		t.Fatalf("in-app notification must be saved even when circuit breaker is OPEN")
	}
	if res5.ExternalSent {
		t.Errorf("expected external send to be skipped/failed under open circuit")
	}
	if res5.ExternalError != notification.ErrCircuitBreakerOpen.Error() {
		t.Errorf("expected ErrCircuitBreakerOpen, got: %s", res5.ExternalError)
	}

	// 4. Verify all 5 notifications exist in user's in-app notification center!
	notifs, err := svc.GetUserNotifications(ctx, userID, false)
	if err != nil {
		t.Fatalf("failed to query user notifications: %v", err)
	}
	if len(notifs) != 5 {
		t.Fatalf("expected exactly 5 in-app notifications stored, got: %d", len(notifs))
	}

	// 5. Test Mark As Read
	if err := svc.MarkAsRead(ctx, "notif-1", userID); err != nil {
		t.Fatalf("failed to mark notification as read: %v", err)
	}
	unread, err := svc.GetUserNotifications(ctx, userID, true)
	if err != nil {
		t.Fatalf("failed to query unread: %v", err)
	}
	if len(unread) != 4 {
		t.Errorf("expected 4 unread notifications, got %d", len(unread))
	}

	// 6. Test circuit recovery after timeout
	time.Sleep(120 * time.Millisecond)
	extClient.failEmail = false
	extClient.failPush = false

	retried := svc.RetryPendingQueue(ctx, "user@example.com", "")
	if retried == 0 {
		t.Errorf("expected pending notifications in retry queue to be retried successfully")
	}
}

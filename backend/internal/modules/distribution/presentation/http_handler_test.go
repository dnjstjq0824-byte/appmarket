package presentation_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/application"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/domain"
	"github.com/bookingnetwork/appmarket/backend/internal/modules/distribution/presentation"
)

type mockCandidateProvider struct {
	candidates map[string]domain.CandidateRelease
}

func (m *mockCandidateProvider) GetCandidatesForPackages(ctx context.Context, packageNames []string) (map[string]domain.CandidateRelease, error) {
	res := make(map[string]domain.CandidateRelease)
	for _, pkg := range packageNames {
		if cand, ok := m.candidates[pkg]; ok {
			res[pkg] = cand
		}
	}
	return res, nil
}

func TestDistributionHandler_GenerateTokenAndCheckUpdates(t *testing.T) {
	tokenSvc := application.NewDownloadTokenService("test-secret-key-1234", "cdn.appmarket.local", 15*time.Minute)
	provider := &mockCandidateProvider{
		candidates: map[string]domain.CandidateRelease{
			"com.test.app": {
				AppID:             "app-1",
				ReleaseID:         "rel-101",
				PackageName:       "com.test.app",
				VersionCode:       101,
				VersionName:       "1.0.1",
				MinSDK:            26,
				TargetSDK:         34,
				NativeABIs:        []string{"arm64-v8a"},
				CertSHA256:        "AABBCCDD11223344",
				RolloutPercentage: 100,
				ReleaseStatus:     domain.ReleasePublishedFull,
				SizeBytes:         1024000,
				SHA256Checksum:    "checksum123",
				DownloadURL:       "https://cdn.appmarket.local/packages/app-1/rel-101/app.apk",
			},
		},
	}

	handler := presentation.NewDistributionHandler(tokenSvc, provider)
	r := handler.Routes()

	// 1. Valid Download Token
	body, _ := json.Marshal(map[string]any{
		"sha256_checksum":   "checksum123",
		"device_or_user_id": "user-1",
		"country_code":      "KR",
		"rights_level":      "LEVEL_1_DIRECT_HOSTED",
	})
	req := httptest.NewRequest(http.MethodPost, "/apps/app-1/releases/rel-101/download-token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var tokenResp application.SignedDownloadURL
	if err := json.NewDecoder(w.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("failed to decode token response: %v", err)
	}
	if tokenResp.URL == "" {
		t.Fatalf("expected non-empty URL: %+v", tokenResp)
	}

	// 2. Block External-Link-Only
	bodyExt, _ := json.Marshal(map[string]any{
		"sha256_checksum":   "checksum123",
		"device_or_user_id": "user-1",
		"country_code":      "KR",
		"rights_level":      "TIER_5_EXTERNAL_LINK_ONLY",
	})
	reqExt := httptest.NewRequest(http.MethodPost, "/apps/app-1/releases/rel-101/download-token", bytes.NewReader(bodyExt))
	reqExt.Header.Set("Content-Type", "application/json")
	wExt := httptest.NewRecorder()
	r.ServeHTTP(wExt, reqExt)

	if wExt.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for EXTERNAL_LINK_ONLY, got %d: %s", wExt.Code, wExt.Body.String())
	}

	// 3. Batch Check Updates
	checkBody, _ := json.Marshal(map[string]any{
		"device": domain.DeviceSpec{
			SDKInt:            34,
			SupportedABIs:     []string{"arm64-v8a"},
			DeviceHash:        "dev-abc",
			AvailableFeatures: []string{},
		},
		"packages": []domain.InstalledAppSpec{
			{
				PackageName:          "com.test.app",
				VersionCode:          100,
				CertSHA256:           "AABBCCDD11223344",
				InstallerPackageName: "com.bookingnetwork.appmarket",
			},
		},
	})
	reqCheck := httptest.NewRequest(http.MethodPost, "/device/check-updates", bytes.NewReader(checkBody))
	reqCheck.Header.Set("Content-Type", "application/json")
	wCheck := httptest.NewRecorder()
	r.ServeHTTP(wCheck, reqCheck)

	if wCheck.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for check-updates, got %d: %s", wCheck.Code, wCheck.Body.String())
	}

	var results []domain.UpdateCandidateResult
	if err := json.NewDecoder(wCheck.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode check updates response: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != domain.StatusUpdateAvailableCompatible {
		t.Fatalf("expected UPDATE_AVAILABLE_COMPATIBLE, got %s", results[0].Status)
	}
}

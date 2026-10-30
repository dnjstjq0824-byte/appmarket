package geo

import (
	"errors"
	"testing"
	"time"
)

// TC-SEC-004: Geo-Fencing, Sanctions Compliance (HTTP 451) & Age Verification
func TestTCSEC004GeoPolicyAndCompliance(t *testing.T) {
	evaluator := NewGeoPolicyEvaluator()

	t.Run("Sub-test 1: Legal Sanctions block takes precedence even if Developer allowed (HTTP 451)", func(t *testing.T) {
		userSanctioned := UserContext{
			CountryCode: "KP", // North Korea (Sanctioned)
		}

		devSettingAllowed := DeveloperCountrySetting{
			AppID:          "app-crypto-wallet",
			CountryCode:    "KP",
			IsAllowedByDev: true, // Developer mistakenly allowed
		}

		err := evaluator.EvaluateAccess(userSanctioned, devSettingAllowed, RatingEveryone)
		if !errors.Is(err, ErrUnavailableForLegalReasons) {
			t.Fatalf("expected ErrUnavailableForLegalReasons (HTTP 451), got %v", err)
		}
	})

	t.Run("Sub-test 2: Developer Geo-block prevents access in unselected countries", func(t *testing.T) {
		userJapan := UserContext{
			CountryCode: "JP",
		}

		devSettingBlocked := DeveloperCountrySetting{
			AppID:          "app-korea-service",
			CountryCode:    "JP",
			IsAllowedByDev: false, // Developer excluded Japan
		}

		err := evaluator.EvaluateAccess(userJapan, devSettingBlocked, RatingEveryone)
		if !errors.Is(err, ErrGeoBlockedByDeveloper) {
			t.Fatalf("expected ErrGeoBlockedByDeveloper, got %v", err)
		}
	})

	t.Run("Sub-test 3: South Korea mandates adult certification for 18+ apps", func(t *testing.T) {
		now := time.Now().UTC()
		devSetting := DeveloperCountrySetting{
			AppID:          "app-adult-game",
			CountryCode:    "KR",
			IsAllowedByDev: true,
		}

		// Case A: Unverified Korean user
		userUnverified := UserContext{
			CountryCode:     "KR",
			IsAdultVerified: false,
		}
		err := evaluator.EvaluateAccess(userUnverified, devSetting, RatingAdult18)
		if !errors.Is(err, ErrAdultVerificationRequired) {
			t.Fatalf("expected ErrAdultVerificationRequired for unverified adult app access, got %v", err)
		}

		// Case B: Verified Korean user
		userVerified := UserContext{
			CountryCode:     "KR",
			IsAdultVerified: true,
			AdultVerifiedAt: &now,
		}
		err = evaluator.EvaluateAccess(userVerified, devSetting, RatingAdult18)
		if err != nil {
			t.Fatalf("expected access permitted for verified adult user, got %v", err)
		}
	})
}

package geo

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrUnavailableForLegalReasons = errors.New("HTTP_451_UNAVAILABLE_FOR_LEGAL_REASONS")
	ErrGeoBlockedByDeveloper      = errors.New("ERR_GEO_BLOCKED_BY_DEVELOPER")
	ErrAdultVerificationRequired  = errors.New("ERR_ADULT_VERIFICATION_REQUIRED")
)

type AgeRating string

const (
	RatingEveryone AgeRating = "EVERYONE"
	RatingTeen     AgeRating = "TEEN"
	RatingAdult18  AgeRating = "AGE_18_ADULT"
)

type CountryPolicy struct {
	CountryCode        string
	IsBlockedByLaw     bool   // Sanctioned, export-controlled, or legally prohibited (e.g. North Korea, Iran, or gambling bans)
	RequiresAdultCert  bool   // e.g. South Korea K-KMC adult verification
	LegalNotice        string
}

type DeveloperCountrySetting struct {
	AppID              string
	CountryCode        string
	IsAllowedByDev     bool
}

type UserContext struct {
	CountryCode        string
	IsAdultVerified    bool
	AdultVerifiedAt    *time.Time
}

type GeoPolicyEvaluator struct {
	legalPolicies map[string]CountryPolicy // CountryCode -> Policy
}

func NewGeoPolicyEvaluator() *GeoPolicyEvaluator {
	return &GeoPolicyEvaluator{
		legalPolicies: map[string]CountryPolicy{
			"KP": {CountryCode: "KP", IsBlockedByLaw: true, LegalNotice: "OFAC Comprehensive Sanctions Block"},
			"IR": {CountryCode: "IR", IsBlockedByLaw: true, LegalNotice: "OFAC Comprehensive Sanctions Block"},
			"CU": {CountryCode: "CU", IsBlockedByLaw: true, LegalNotice: "OFAC Comprehensive Sanctions Block"},
			"SY": {CountryCode: "SY", IsBlockedByLaw: true, LegalNotice: "OFAC Comprehensive Sanctions Block"},
			"KR": {CountryCode: "KR", IsBlockedByLaw: false, RequiresAdultCert: true}, // Korea mandates adult cert for 18+
		},
	}
}

func (e *GeoPolicyEvaluator) AddLegalPolicy(policy CountryPolicy) {
	e.legalPolicies[strings.ToUpper(policy.CountryCode)] = policy
}

// EvaluateAccess evaluates whether a user from a specific country can access an app (TC-SEC-004)
func (e *GeoPolicyEvaluator) EvaluateAccess(user UserContext, devSetting DeveloperCountrySetting, appAgeRating AgeRating) error {
	country := strings.ToUpper(strings.TrimSpace(user.CountryCode))
	if country == "" {
		country = "US" // Default fallback
	}

	// 1. Legal & Regulatory Sanction Check (HIGHEST PRIORITY: Law overrides developer choice)
	if policy, exists := e.legalPolicies[country]; exists {
		if policy.IsBlockedByLaw {
			return ErrUnavailableForLegalReasons
		}
	}

	// 2. Developer Geo-Availability Setting
	if !devSetting.IsAllowedByDev {
		return ErrGeoBlockedByDeveloper
	}

	// 3. Age Rating & Adult Verification Mandate
	if appAgeRating == RatingAdult18 {
		policy, exists := e.legalPolicies[country]
		if exists && policy.RequiresAdultCert {
			if !user.IsAdultVerified || user.AdultVerifiedAt == nil {
				return ErrAdultVerificationRequired
			}
			// Verification valid for 1 year
			if time.Since(*user.AdultVerifiedAt) > 365*24*time.Hour {
				return ErrAdultVerificationRequired
			}
		}
	}

	return nil
}

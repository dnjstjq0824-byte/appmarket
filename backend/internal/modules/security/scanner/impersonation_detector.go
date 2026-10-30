package scanner

import (
	"strings"
	"unicode"
)

type ProtectedBrand struct {
	BrandName          string
	OfficialPackage    string
	CanonicalKeywords  []string
	AuthorizedOrgID    string
}

var DefaultProtectedBrands = []ProtectedBrand{
	{BrandName: "WhatsApp", OfficialPackage: "com.whatsapp", CanonicalKeywords: []string{"whatsapp", "whats app"}, AuthorizedOrgID: "org-meta"},
	{BrandName: "KakaoTalk", OfficialPackage: "com.kakao.talk", CanonicalKeywords: []string{"kakaotalk", "kakao talk"}, AuthorizedOrgID: "org-kakao"},
	{BrandName: "YouTube", OfficialPackage: "com.google.android.youtube", CanonicalKeywords: []string{"youtube", "you tube"}, AuthorizedOrgID: "org-google"},
	{BrandName: "Netflix", OfficialPackage: "com.netflix.mediaclient", CanonicalKeywords: []string{"netflix"}, AuthorizedOrgID: "org-netflix"},
	{BrandName: "Toss", OfficialPackage: "viva.republica.toss", CanonicalKeywords: []string{"toss", "toss bank"}, AuthorizedOrgID: "org-toss"},
	{BrandName: "Instagram", OfficialPackage: "com.instagram.android", CanonicalKeywords: []string{"instagram"}, AuthorizedOrgID: "org-meta"},
}

type ImpersonationResult struct {
	IsImpersonating bool
	MatchedBrand    string
	SimilarityScore float64
	Reason          string
}

// DetectBrandImpersonation evaluates if an app title or package name mimics a protected brand without authorization.
func DetectBrandImpersonation(appTitle, packageName, orgID string) ImpersonationResult {
	titleNorm := normalizeString(appTitle)
	pkgNorm := normalizeString(packageName)

	for _, brand := range DefaultProtectedBrands {
		if brand.AuthorizedOrgID == orgID {
			continue // authorized
		}

		for _, kw := range brand.CanonicalKeywords {
			kwNorm := normalizeString(kw)
			sim := stringSimilarity(titleNorm, kwNorm)

			if strings.Contains(pkgNorm, kwNorm) && !strings.HasPrefix(pkgNorm, brand.OfficialPackage) {
				return ImpersonationResult{
					IsImpersonating: true,
					MatchedBrand:    brand.BrandName,
					SimilarityScore: 0.95,
					Reason:          "Package name mimics official protected brand package",
				}
			}

			if sim >= 0.85 {
				return ImpersonationResult{
					IsImpersonating: true,
					MatchedBrand:    brand.BrandName,
					SimilarityScore: sim,
					Reason:          "App title similarity exceeds brand protection threshold (85%)",
				}
			}
		}
	}

	return ImpersonationResult{
		IsImpersonating: false,
		SimilarityScore: 0.0,
	}
}

func normalizeString(s string) string {
	var sb strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func stringSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Simple Levenshtein-based similarity
	dist := levenshteinDistance(a, b)
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	return 1.0 - (float64(dist) / float64(maxLen))
}

func levenshteinDistance(s, t string) int {
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	for i := 1; i <= len(s); i++ {
		for j := 1; j <= len(t); j++ {
			cost := 0
			if s[i-1] != t[j-1] {
				cost = 1
			}
			d[i][j] = min(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)
		}
	}
	return d[len(s)][len(t)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}


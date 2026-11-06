package domain

import (
	"encoding/base64"
	"encoding/json"
)

// List ordering contract shared by every CatalogRepository implementation:
// bayesian_rating DESC, total_installs DESC, id ASC. The id tiebreak keeps
// keyset pagination deterministic when ratings repeat.
type listCursor struct {
	BayesianRating float64 `json:"r"`
	TotalInstalls  int64   `json:"t"`
	ID             string  `json:"i"`
}

// EncodeListCursor builds an opaque pagination cursor positioned after the
// given row of the list ordering contract.
func EncodeListCursor(bayesianRating float64, totalInstalls int64, id string) string {
	raw, err := json.Marshal(listCursor{BayesianRating: bayesianRating, TotalInstalls: totalInstalls, ID: id})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeListCursor parses a cursor produced by EncodeListCursor.
// ok is false for empty or malformed input; callers treat that as "first page".
func DecodeListCursor(cursor string) (bayesianRating float64, totalInstalls int64, id string, ok bool) {
	if cursor == "" {
		return 0, 0, "", false
	}
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, 0, "", false
	}
	var c listCursor
	if err := json.Unmarshal(raw, &c); err != nil || c.ID == "" {
		return 0, 0, "", false
	}
	return c.BayesianRating, c.TotalInstalls, c.ID, true
}

// AfterCursor reports whether row (r, t, i) sorts strictly after the decoded
// cursor position under the list ordering contract.
func AfterCursor(r float64, t int64, id string, cr float64, ct int64, ci string) bool {
	if r != cr {
		return r < cr
	}
	if t != ct {
		return t < ct
	}
	return id > ci
}

package scanner

import (
	"strings"
)

var DangerousPermissionWeights = map[string]int{
	"android.permission.BIND_ACCESSIBILITY_SERVICE": 35,
	"android.permission.BIND_DEVICE_ADMIN":          30,
	"android.permission.RECEIVE_SMS":                25,
	"android.permission.SEND_SMS":                   25,
	"android.permission.READ_SMS":                   25,
	"android.permission.REQUEST_INSTALL_PACKAGES":   20,
	"android.permission.INSTALL_PACKAGES":           25,
	"android.permission.READ_CALL_LOG":              20,
	"android.permission.PROCESS_OUTGOING_CALLS":     20,
	"android.permission.SYSTEM_ALERT_WINDOW":        15,
	"android.permission.ACCESS_BACKGROUND_LOCATION": 15,
	"android.permission.ACCESS_FINE_LOCATION":       10,
	"android.permission.RECORD_AUDIO":               10,
	"android.permission.CAMERA":                     10,
	"android.permission.READ_CONTACTS":              10,
}

type PermissionAnalysisResult struct {
	DangerousPermissions []string
	TotalRiskScore       int
	RequiresManualReview bool
}

// CalculatePermissionRiskScore evaluates the requested permissions and computes a risk score in [0, 100].
func CalculatePermissionRiskScore(permissions []string) PermissionAnalysisResult {
	score := 0
	var dangerous []string

	for _, perm := range permissions {
		cleanPerm := strings.TrimSpace(perm)
		if weight, exists := DangerousPermissionWeights[cleanPerm]; exists {
			score += weight
			dangerous = append(dangerous, cleanPerm)
		}
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return PermissionAnalysisResult{
		DangerousPermissions: dangerous,
		TotalRiskScore:       score,
		RequiresManualReview: score >= 70,
	}
}


package domain

import (
	"strings"
)

type UpdateResolutionStatus string

const (
	StatusNoUpdate                     UpdateResolutionStatus = "NO_UPDATE"
	StatusUpdateAvailableCompatible   UpdateResolutionStatus = "UPDATE_AVAILABLE_COMPATIBLE"
	StatusUpdateAvailableKeyRotation  UpdateResolutionStatus = "COMPATIBLE_KEY_ROTATION"
	StatusUpdateAvailableSignerConflict UpdateResolutionStatus = "UPDATE_AVAILABLE_SIGNER_CONFLICT"
	StatusIncompatibleDeviceSdkAbi    UpdateResolutionStatus = "INCOMPATIBLE_DEVICE_SDK_ABI"
	StatusRolloutExcluded             UpdateResolutionStatus = "ROLLOUT_EXCLUDED"
)

type DeviceSpec struct {
	SDKInt            int      `json:"sdk_int"`
	SupportedABIs     []string `json:"supported_abis"`
	DensityDPI        int      `json:"density_dpi"`
	AvailableFeatures []string `json:"available_features"`
	DeviceHash        string   `json:"device_hash"`
}

type InstalledAppSpec struct {
	PackageName          string `json:"package_name"`
	VersionCode          int64  `json:"version_code"`
	CertSHA256           string `json:"cert_sha256"`
	InstallerPackageName string `json:"installer_package_name"`
}

type CandidateRelease struct {
	AppID              string        `json:"app_id"`
	ReleaseID          string        `json:"release_id"`
	PackageName        string        `json:"package_name"`
	VersionCode        int64         `json:"version_code"`
	VersionName        string        `json:"version_name"`
	MinSDK             int           `json:"min_sdk"`
	TargetSDK          int           `json:"target_sdk"`
	NativeABIs         []string      `json:"native_abis"`
	RequiredFeatures   []string      `json:"required_features"`
	CertSHA256         string        `json:"cert_sha256"`
	RotatedCertLineage []string      `json:"rotated_cert_lineage"`
	RolloutPercentage  int           `json:"rollout_percentage"`
	ReleaseStatus      ReleaseStatus `json:"release_status"`
	DownloadURL        string        `json:"download_url"`
	SHA256Checksum     string        `json:"sha256_checksum"`
	SizeBytes          int64         `json:"size_bytes"`
}

type UpdateCandidateResult struct {
	AppID                  string                 `json:"app_id"`
	ReleaseID              string                 `json:"release_id"`
	PackageName            string                 `json:"package_name"`
	VersionCode            int64                  `json:"version_code"`
	VersionName            string                 `json:"version_name"`
	Status                 UpdateResolutionStatus `json:"status"`
	ResolutionReason       string                 `json:"resolution_reason"`
	DownloadURL            string                 `json:"download_url,omitempty"`
	SHA256Checksum         string                 `json:"sha256_checksum,omitempty"`
	SizeBytes              int64                  `json:"size_bytes,omitempty"`
	RequiresUserAction     bool                   `json:"requires_user_action"`
	DangerousPermissionsNew []string              `json:"dangerous_permissions_new,omitempty"`
}

// ResolveUpdate evaluates 4D device compatibility, staged rollout bucketing, and cryptographic signature continuity
func ResolveUpdate(device DeviceSpec, installed InstalledAppSpec, candidate CandidateRelease) (UpdateResolutionStatus, *UpdateCandidateResult) {
	result := &UpdateCandidateResult{
		AppID:       candidate.AppID,
		ReleaseID:   candidate.ReleaseID,
		PackageName: candidate.PackageName,
		VersionCode: candidate.VersionCode,
		VersionName: candidate.VersionName,
	}

	// 1. VersionCode Monotonicity check
	if candidate.VersionCode <= installed.VersionCode {
		result.Status = StatusNoUpdate
		result.ResolutionReason = "Installed version is equal to or greater than available candidate."
		return StatusNoUpdate, result
	}

	// 2. Release Status check
	if candidate.ReleaseStatus != ReleasePublishedFull && candidate.ReleaseStatus != ReleasePublishedStaged {
		result.Status = StatusNoUpdate
		result.ResolutionReason = "Candidate release is not currently published."
		return StatusNoUpdate, result
	}

	// 3. 4-Dimensional Device Compatibility Check
	// 3a. Min SDK
	if device.SDKInt < candidate.MinSDK {
		result.Status = StatusIncompatibleDeviceSdkAbi
		result.ResolutionReason = "Device Android SDK is below required minimum SDK."
		return StatusIncompatibleDeviceSdkAbi, result
	}

	// 3b. Native ABIs (if app has native libs, device must support at least one)
	if len(candidate.NativeABIs) > 0 {
		abiMatched := false
		for _, candAbi := range candidate.NativeABIs {
			for _, devAbi := range device.SupportedABIs {
				if strings.EqualFold(candAbi, devAbi) {
					abiMatched = true
					break
				}
			}
			if abiMatched {
				break
			}
		}
		if !abiMatched {
			result.Status = StatusIncompatibleDeviceSdkAbi
			result.ResolutionReason = "Device architecture does not match candidate native ABIs."
			return StatusIncompatibleDeviceSdkAbi, result
		}
	}

	// 3c. Required Hardware Features
	if len(candidate.RequiredFeatures) > 0 {
		for _, reqFeat := range candidate.RequiredFeatures {
			featureFound := false
			for _, devFeat := range device.AvailableFeatures {
				if strings.EqualFold(reqFeat, devFeat) {
					featureFound = true
					break
				}
			}
			if !featureFound {
				result.Status = StatusIncompatibleDeviceSdkAbi
				result.ResolutionReason = "Device lacks required hardware feature: " + reqFeat
				return StatusIncompatibleDeviceSdkAbi, result
			}
		}
	}

	// 4. Staged Rollout Bucket Check
	if candidate.ReleaseStatus == ReleasePublishedStaged {
		if !IsDeviceEligibleForRollout(candidate.AppID, device.DeviceHash, candidate.RolloutPercentage) {
			result.Status = StatusRolloutExcluded
			result.ResolutionReason = "Device is outside active staged rollout percentage threshold."
			return StatusRolloutExcluded, result
		}
	}

	// 5. Signature Continuity & Lineage Check
	normInstalledCert := strings.ToUpper(strings.TrimSpace(installed.CertSHA256))
	normCandidateCert := strings.ToUpper(strings.TrimSpace(candidate.CertSHA256))

	if normInstalledCert == normCandidateCert {
		result.Status = StatusUpdateAvailableCompatible
		result.ResolutionReason = "Signatures match exactly. Seamless direct update permitted."
		result.DownloadURL = candidate.DownloadURL
		result.SHA256Checksum = candidate.SHA256Checksum
		result.SizeBytes = candidate.SizeBytes
		result.RequiresUserAction = false
		return StatusUpdateAvailableCompatible, result
	}

	// Check v3 Proof-of-Rotation lineage
	for _, historicalCert := range candidate.RotatedCertLineage {
		if strings.EqualFold(normInstalledCert, strings.TrimSpace(historicalCert)) {
			result.Status = StatusUpdateAvailableKeyRotation
			result.ResolutionReason = "Valid v3 signature key rotation lineage confirmed."
			result.DownloadURL = candidate.DownloadURL
			result.SHA256Checksum = candidate.SHA256Checksum
			result.SizeBytes = candidate.SizeBytes
			result.RequiresUserAction = false
			return StatusUpdateAvailableKeyRotation, result
		}
	}

	// Signature Mismatch without lineage -> Signer Conflict!
	result.Status = StatusUpdateAvailableSignerConflict
	result.ResolutionReason = "Signature certificates differ with no proof-of-rotation chain. Update requires manual re-installation."
	result.DownloadURL = candidate.DownloadURL
	result.SHA256Checksum = candidate.SHA256Checksum
	result.SizeBytes = candidate.SizeBytes
	result.RequiresUserAction = true
	return StatusUpdateAvailableSignerConflict, result
}

// BatchResolveUpdates evaluates multiple installed applications for a single device
func BatchResolveUpdates(device DeviceSpec, installedList []InstalledAppSpec, candidates map[string]CandidateRelease) []UpdateCandidateResult {
	results := make([]UpdateCandidateResult, 0, len(installedList))

	for _, installed := range installedList {
		candidate, exists := candidates[installed.PackageName]
		if !exists {
			results = append(results, UpdateCandidateResult{
				PackageName:      installed.PackageName,
				VersionCode:      installed.VersionCode,
				Status:           StatusNoUpdate,
				ResolutionReason: "No newer release found in store catalog.",
			})
			continue
		}

		_, res := ResolveUpdate(device, installed, candidate)
		results = append(results, *res)
	}

	return results
}


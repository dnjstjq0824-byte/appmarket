package domain

import (
	"errors"
	"time"
)

var (
	ErrDownloadForbiddenRightsLevel = errors.New("ERR_DIRECT_DOWNLOAD_FORBIDDEN_FOR_RIGHTS_LEVEL")
	ErrProvenanceMissing            = errors.New("ERR_PROVENANCE_REQUIRED_FIELDS_MISSING")
	ErrAbortedSignatureInvalid      = errors.New("ABORTED_SIGNATURE_INVALID")
)

type RightsLevel string

const (
	Level1DirectHosted          RightsLevel = "LEVEL_1_DIRECT_HOSTED"
	Level2LicensedFederated     RightsLevel = "LEVEL_2_LICENSED_FEDERATED_BINARY"
	Level3SearchAndInfoOnly     RightsLevel = "LEVEL_3_SEARCH_AND_INFO_ONLY"
	Level4OfficialLinkOnly      RightsLevel = "LEVEL_4_OFFICIAL_LINK_ONLY"
	Level5RestrictedProhibited  RightsLevel = "LEVEL_5_RESTRICTED_PROHIBITED"
)

type ProvenanceMetadata struct {
	ProvenanceLabel  string      `json:"provenance_label"`
	LastSyncedAt     time.Time   `json:"last_synced_at"`
	OfficialStoreURL string      `json:"official_store_url"`
	CanDirectInstall bool        `json:"can_direct_install"`
	RightsLevel      RightsLevel `json:"rights_level"`
	ActionText       string      `json:"action_text"`
}

type RightsGuard struct{}

func NewRightsGuard() *RightsGuard {
	return &RightsGuard{}
}

// ValidateDownloadAllowed enforces strict prevention of direct APK downloads on Level 3, 4, 5
func (g *RightsGuard) ValidateDownloadAllowed(level RightsLevel) error {
	switch level {
	case Level1DirectHosted, Level2LicensedFederated:
		return nil
	case Level3SearchAndInfoOnly, Level4OfficialLinkOnly, Level5RestrictedProhibited:
		return ErrDownloadForbiddenRightsLevel
	default:
		return ErrDownloadForbiddenRightsLevel
	}
}

// FormatProvenance builds the mandatory 4 provenance fields according to DOC-SRCH-004
func (g *RightsGuard) FormatProvenance(label string, officialURL string, lastSynced time.Time, level RightsLevel) (*ProvenanceMetadata, error) {
	if label == "" || officialURL == "" {
		return nil, ErrProvenanceMissing
	}

	canDirect := false
	actionText := ""

	switch level {
	case Level1DirectHosted:
		canDirect = true
		actionText = "직접 APK 다운로드"
	case Level2LicensedFederated:
		canDirect = true
		actionText = "오픈소스 패키지 설치 (출처: " + label + ")"
	case Level3SearchAndInfoOnly:
		canDirect = false
		actionText = label + "에서 확인 및 설치 ↗"
	case Level4OfficialLinkOnly:
		canDirect = false
		actionText = "공식 배포처로 이동 ↗"
	case Level5RestrictedProhibited:
		canDirect = false
		actionText = "배포 중단"
	}

	return &ProvenanceMetadata{
		ProvenanceLabel:  label,
		LastSyncedAt:     lastSynced.UTC(),
		OfficialStoreURL: officialURL,
		CanDirectInstall: canDirect,
		RightsLevel:      level,
		ActionText:       actionText,
	}, nil
}


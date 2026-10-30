package domain

import (
	"errors"
	"time"
)

type ScanVerdict string

const (
	VerdictCleanLowRisk            ScanVerdict = "CLEAN_LOW_RISK"
	VerdictCleanHighRiskManual     ScanVerdict = "CLEAN_HIGH_RISK_MANUAL_REVIEW"
	VerdictInfectedQuarantined     ScanVerdict = "INFECTED_QUARANTINED"
	VerdictScanError               ScanVerdict = "SCAN_ERROR"
)

type ClamAVStatus string

const (
	ClamAVClean    ClamAVStatus = "CLEAN"
	ClamAVInfected ClamAVStatus = "INFECTED"
	ClamAVError    ClamAVStatus = "ERROR"
)

var (
	ErrInfectedMalwareDetected = errors.New("INFECTED_QUARANTINED: malware detected in artifact")
	ErrKillSwitchActive        = errors.New("ERR_KILL_SWITCH_ACTIVE: application is suspended by emergency kill switch")
	ErrScanFailed              = errors.New("SCAN_ERROR: security scanner encountered an error")
)

type SecurityScanReport struct {
	ID                         string       `json:"id"`
	ArtifactID                 string       `json:"artifact_id"`
	ScannerEngineVersion       string       `json:"scanner_engine_version"`
	ClamAVStatus               ClamAVStatus `json:"clamav_status"`
	YARAMatchedRules           []string     `json:"yara_matched_rules"`
	DangerousPermissions       []string     `json:"dangerous_permissions"`
	ImpersonationScore         float64      `json:"impersonation_similarity_score"`
	RiskScore                  int          `json:"risk_score"`
	Verdict                    ScanVerdict  `json:"scan_verdict"`
	RawReport                  string       `json:"raw_report"`
	ScannedAt                  time.Time    `json:"scanned_at"`
}

type KillSwitchState struct {
	AppID       string    `json:"app_id"`
	IsTriggered bool      `json:"is_triggered"`
	Reason      string    `json:"reason"`
	TriggeredBy string    `json:"triggered_by"`
	TriggeredAt time.Time `json:"triggered_at"`
}


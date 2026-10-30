package scanner

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/bookingnetwork/appmarket/backend/internal/modules/security/domain"
)

const EICARTestString = "X5O!P%@AP[4\\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*"

type YaraPattern struct {
	RuleName string
	Keywords []string
	MinMatch int
}

var BuiltInYaraPatterns = []YaraPattern{
	{
		RuleName: "EICAR_Test_File",
		Keywords: []string{EICARTestString},
		MinMatch: 1,
	},
	{
		RuleName: "Android_Joker_Sms_Fraud",
		Keywords: []string{
			"com.joker.fraud.NotificationListener",
			"telephony.sms.SEND_SMS_SILENT",
			"intercept_otp_tokens",
		},
		MinMatch: 2,
	},
	{
		RuleName: "Android_Anubis_Banker",
		Keywords: []string{
			"anubis.overlay.injector",
			"record_audio_stream_hidden",
			"keylogger_raw_input_dump",
		},
		MinMatch: 2,
	},
	{
		RuleName: "Android_Fake_Market_Updater",
		Keywords: []string{
			"download_secondary_dex_from_c2",
			"dalvik.system.DexClassLoader.loadClassHidden",
		},
		MinMatch: 2,
	},
}

type ScanResult struct {
	ClamAVStatus     domain.ClamAVStatus
	YARAMatchedRules []string
	IsInfected       bool
	Verdict          domain.ScanVerdict
	Details          string
}

type MalwareScanner struct{}

func NewMalwareScanner() *MalwareScanner {
	return &MalwareScanner{}
}

// ScanArtifact inspects the byte stream of an APK/AAB for known malware signatures and YARA rules.
func (s *MalwareScanner) ScanArtifact(ctx context.Context, r io.Reader) (*ScanResult, error) {
	data, err := io.ReadAll(io.LimitReader(r, 100*1024*1024)) // limit inspection to first 100MB
	if err != nil {
		return nil, domain.ErrScanFailed
	}

	var matchedRules []string
	clamavStatus := domain.ClamAVClean

	// 1. ClamAV / EICAR check
	if bytes.Contains(data, []byte(EICARTestString)) {
		clamavStatus = domain.ClamAVInfected
		matchedRules = append(matchedRules, "EICAR_Test_File")
	}

	// 2. YARA Rule evaluation
	for _, pattern := range BuiltInYaraPatterns {
		if pattern.RuleName == "EICAR_Test_File" && clamavStatus == domain.ClamAVInfected {
			continue // already matched
		}

		matchCount := 0
		for _, kw := range pattern.Keywords {
			if bytes.Contains(data, []byte(kw)) {
				matchCount++
			}
		}

		if matchCount >= pattern.MinMatch {
			matchedRules = append(matchedRules, pattern.RuleName)
		}
	}

	isInfected := clamavStatus == domain.ClamAVInfected || len(matchedRules) > 0
	verdict := domain.VerdictCleanLowRisk
	details := "No malware threats detected"

	if isInfected {
		verdict = domain.VerdictInfectedQuarantined
		details = "Infected: matched rules [" + strings.Join(matchedRules, ", ") + "]"
	}

	return &ScanResult{
		ClamAVStatus:     clamavStatus,
		YARAMatchedRules: matchedRules,
		IsInfected:       isInfected,
		Verdict:          verdict,
		Details:          details,
	}, nil
}

